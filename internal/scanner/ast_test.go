package scanner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testGoFile = `package test

import (
	"time"
)

// TestConfig represents a test configuration
//go:generate go-validate -struct TestConfig
type TestConfig struct {
	// Name is the service name
	Name string ` + "`" + `json:"name" validate:"required,minlen=3"` + "`" + `
	
	// Port is the service port
	Port int ` + "`" + `json:"port" validate:"min=1,max=65535"` + "`" + `
	
	// Enabled indicates if the service is enabled
	Enabled bool ` + "`" + `json:"enabled"` + "`" + `
	
	// Timeout for operations
	Timeout time.Duration ` + "`" + `json:"timeout" validate:"required"` + "`" + `
}

// AnotherConfig without validation tags
type AnotherConfig struct {
	Value string
}

// ConfigWithValidation has validation tags
type ConfigWithValidation struct {
	Email string ` + "`" + `validate:"required,email"` + "`" + `
	Age   int    ` + "`" + `validate:"min=18,max=100"` + "`" + `
}
`

func TestNewASTScanner(t *testing.T) {
	scanner := NewASTScanner(true)
	assert.NotNil(t, scanner)
	assert.NotNil(t, scanner.fileSet)
	assert.NotNil(t, scanner.cache)
	assert.True(t, scanner.verbose)
}

func TestScanFile(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(testGoFile), 0644)
	require.NoError(t, err)

	scanner := NewASTScanner(false)
	structs, err := scanner.scanFile(filePath)
	require.NoError(t, err)

	// Should find 2 structs with validation tags: TestConfig and ConfigWithValidation
	assert.Len(t, structs, 2)

	// Check TestConfig
	testConfig := findStructByName(structs, "TestConfig")
	require.NotNil(t, testConfig)
	assert.Equal(t, "TestConfig", testConfig.Name)
	assert.Equal(t, "test", testConfig.PackageName)
	assert.Len(t, testConfig.Fields, 4)

	// Check Name field
	nameField := findFieldByName(testConfig.Fields, "Name")
	require.NotNil(t, nameField)
	assert.Equal(t, "string", nameField.Type)
	assert.Equal(t, "name", nameField.JSONTag)
	assert.True(t, nameField.Required)
	assert.Equal(t, "3", nameField.ValidateRules["minlen"])

	// Check Port field
	portField := findFieldByName(testConfig.Fields, "Port")
	require.NotNil(t, portField)
	assert.Equal(t, "int", portField.Type)
	assert.Equal(t, "1", portField.ValidateRules["min"])
	assert.Equal(t, "65535", portField.ValidateRules["max"])

	// Check ConfigWithValidation
	configWithValidation := findStructByName(structs, "ConfigWithValidation")
	require.NotNil(t, configWithValidation)
	assert.Len(t, configWithValidation.Fields, 2)
}

func TestScanDirectory(t *testing.T) {
	// Create temporary directory with Go files
	tempDir := t.TempDir()
	
	// Create first file
	file1Path := filepath.Join(tempDir, "config1.go")
	err := os.WriteFile(file1Path, []byte(testGoFile), 0644)
	require.NoError(t, err)

	// Create second file without validation
	file2Content := `package test

type SimpleConfig struct {
	Value string
}
`
	file2Path := filepath.Join(tempDir, "config2.go")
	err = os.WriteFile(file2Path, []byte(file2Content), 0644)
	require.NoError(t, err)

	scanner := NewASTScanner(false)
	structs, err := scanner.ScanDirectory(tempDir)
	require.NoError(t, err)

	// Should find only structs with validation tags from file1
	assert.Len(t, structs, 2)
	assert.NotNil(t, findStructByName(structs, "TestConfig"))
	assert.NotNil(t, findStructByName(structs, "ConfigWithValidation"))
}

func TestHasValidationAnnotations(t *testing.T) {
	scanner := NewASTScanner(false)

	tests := []struct {
		name     string
		source   string
		expected bool
	}{
		{
			name: "has validation tags",
			source: `type Config struct {
				Name string ` + "`" + `validate:"required"` + "`" + `
			}`,
			expected: true,
		},
		{
			name: "no validation tags",
			source: `type Config struct {
				Name string ` + "`" + `json:"name"` + "`" + `
			}`,
			expected: false,
		},
		{
			name: "empty struct",
			source: `type Config struct {}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			src := "package test\n" + tt.source
			file, err := parser.ParseFile(fset, "", src, 0)
			require.NoError(t, err)

			var structType *ast.StructType
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					if st, ok := ts.Type.(*ast.StructType); ok {
						structType = st
						return false
					}
				}
				return true
			})

			require.NotNil(t, structType)
			result := scanner.hasValidationAnnotations(structType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseValidationRules(t *testing.T) {
	scanner := NewASTScanner(false)

	tests := []struct {
		name     string
		tag      string
		expected map[string]string
	}{
		{
			name: "simple rules",
			tag:  "required,min=1,max=100",
			expected: map[string]string{
				"required": "",
				"min":      "1",
				"max":      "100",
			},
		},
		{
			name: "rules with spaces",
			tag:  "required, minlen=3 , maxlen=20",
			expected: map[string]string{
				"required": "",
				"minlen":   "3",
				"maxlen":   "20",
			},
		},
		{
			name: "complex rule values",
			tag:  "oneof=red|green|blue,regex=^[a-zA-Z]+$",
			expected: map[string]string{
				"oneof": "red|green|blue",
				"regex": "^[a-zA-Z]+$",
			},
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.parseValidationRules(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractType(t *testing.T) {
	scanner := NewASTScanner(false)

	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "simple type",
			source:   "string",
			expected: "string",
		},
		{
			name:     "pointer type",
			source:   "*string",
			expected: "*string",
		},
		{
			name:     "slice type",
			source:   "[]string",
			expected: "[]string",
		},
		{
			name:     "map type",
			source:   "map[string]int",
			expected: "map[string]int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			src := "package test\ntype T " + tt.source
			file, err := parser.ParseFile(fset, "", src, 0)
			require.NoError(t, err)

			var expr ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					expr = ts.Type
					return false
				}
				return true
			})

			require.NotNil(t, expr)
			result := scanner.extractType(expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClearCache(t *testing.T) {
	scanner := NewASTScanner(false)
	
	// Add something to cache
	scanner.cache["test"] = &ast.File{}
	assert.Len(t, scanner.cache, 1)
	
	// Clear cache
	scanner.ClearCache()
	assert.Len(t, scanner.cache, 0)
}

// Helper functions for tests

func findStructByName(structs []StructInfo, name string) *StructInfo {
	for i := range structs {
		if structs[i].Name == name {
			return &structs[i]
		}
	}
	return nil
}

func findFieldByName(fields []FieldInfo, name string) *FieldInfo {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func BenchmarkScanFile(b *testing.B) {
	// Create temporary file
	tempDir := b.TempDir()
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(testGoFile), 0644)
	require.NoError(b, err)

	scanner := NewASTScanner(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.scanFile(filePath)
	}
}

func BenchmarkScanFileWithCache(b *testing.B) {
	// Create temporary file
	tempDir := b.TempDir()
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(testGoFile), 0644)
	require.NoError(b, err)

	scanner := NewASTScanner(false)
	
	// Prime the cache
	_, _ = scanner.scanFile(filePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.scanFile(filePath)
	}
}