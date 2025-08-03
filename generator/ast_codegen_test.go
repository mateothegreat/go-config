package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/mateothegreat/go-config/scanner"
)

func TestASTCodeGenerator_GenerateFile(t *testing.T) {
	tests := []struct {
		name         string
		structs      []scanner.StructInfo
		packageName  string
		wantContains []string
		wantImports  []string
	}{
		{
			name: "basic struct with required validation",
			structs: []scanner.StructInfo{
				{
					Name:        "TestConfig",
					PackageName: "testpkg",
					Fields: []scanner.FieldInfo{
						{
							Name:     "Name",
							Type:     "string",
							Required: true,
							ValidateRules: map[string]string{
								"required": "",
								"minlen":   "3",
							},
						},
					},
				},
			},
			packageName: "testpkg",
			wantContains: []string{
				"package testpkg",
				"func init()",
				"validation.RegisterValidation",
				"func (s *TestConfig) Validate() error",
				"errors.NewError()",
				"rules.Required",
				"rules.MinLength(3)",
			},
			wantImports: []string{
				"github.com/mateothegreat/go-validation",
				"github.com/mateothegreat/go-validation/rules",
				"github.com/mateothegreat/go-config/errors",
			},
		},
		{
			name: "struct with multiple validation types",
			structs: []scanner.StructInfo{
				{
					Name:        "ServerConfig",
					PackageName: "config",
					Fields: []scanner.FieldInfo{
						{
							Name:     "Email",
							Type:     "string",
							Required: true,
							ValidateRules: map[string]string{
								"required": "",
								"email":    "",
							},
						},
						{
							Name: "Port",
							Type: "int",
							ValidateRules: map[string]string{
								"min": "1",
								"max": "65535",
							},
						},
						{
							Name: "URL",
							Type: "string",
							ValidateRules: map[string]string{
								"url": "",
							},
						},
					},
				},
			},
			packageName: "config",
			wantContains: []string{
				"rules.Email",
				"rules.URL",
				"rules.Min(1)",
				"rules.Max(65535)",
				"errs.Field(\"Email\").Required()",
				"errs.Field(\"Email\").Email()",
				"errs.Field(\"Port\").Min(1)",
				"errs.Field(\"Port\").Max(65535)",
				"errs.Field(\"URL\").URL()",
			},
		},
		{
			name: "multiple structs in single file",
			structs: []scanner.StructInfo{
				{
					Name:        "ConfigA",
					PackageName: "test",
					Fields: []scanner.FieldInfo{
						{
							Name: "FieldA",
							Type: "string",
							ValidateRules: map[string]string{
								"alpha": "",
							},
						},
					},
				},
				{
					Name:        "ConfigB",
					PackageName: "test",
					Fields: []scanner.FieldInfo{
						{
							Name: "FieldB",
							Type: "string",
							ValidateRules: map[string]string{
								"numeric": "",
							},
						},
					},
				},
			},
			packageName: "test",
			wantContains: []string{
				"func (s *ConfigA) Validate() error",
				"func (s *ConfigB) Validate() error",
				"rules.Alpha",
				"rules.Numeric",
				`validation.RegisterValidation("configa"`,
				`validation.RegisterValidation("configb"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create generator
			gen := NewASTCodeGenerator(GeneratorConfig{})

			// Generate code
			got, err := gen.GenerateFile(tt.structs, tt.packageName)
			if err != nil {
				t.Fatalf("GenerateFile() error = %v", err)
			}

			// Check that generated code contains expected strings
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("Generated code missing expected content: %q", want)
				}
			}

			// Parse and validate the generated code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "generated.go", got, parser.ParseComments)
			if err != nil {
				t.Fatalf("Generated code has syntax errors: %v", err)
			}

			// Check package name
			if file.Name.Name != tt.packageName {
				t.Errorf("Package name = %v, want %v", file.Name.Name, tt.packageName)
			}

			// Check imports
			importMap := make(map[string]bool)
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				importMap[path] = true
			}

			for _, wantImport := range tt.wantImports {
				if !importMap[wantImport] {
					t.Errorf("Missing expected import: %s", wantImport)
				}
			}

			// Verify init function exists
			hasInit := false
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "init" {
					hasInit = true
					break
				}
			}
			if !hasInit {
				t.Error("Generated code missing init function")
			}

			// Verify validation methods exist
			for _, structInfo := range tt.structs {
				hasValidate := false
				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok {
						if fn.Name.Name == "Validate" && fn.Recv != nil {
							// Check receiver type
							if len(fn.Recv.List) > 0 {
								if starExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
									if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == structInfo.Name {
										hasValidate = true
										break
									}
								}
							}
						}
					}
				}
				if !hasValidate {
					t.Errorf("Missing Validate method for struct %s", structInfo.Name)
				}
			}
		})
	}
}

func TestValidatorRegistry_GetValidatorExpression(t *testing.T) {
	registry := NewValidatorRegistry()

	tests := []struct {
		name      string
		rule      string
		value     string
		fieldType string
		wantType  string // Expected AST node type
	}{
		{
			name:      "simple required validator",
			rule:      "required",
			value:     "",
			fieldType: "string",
			wantType:  "*ast.SelectorExpr",
		},
		{
			name:      "parameterized min validator for int",
			rule:      "min",
			value:     "10",
			fieldType: "int",
			wantType:  "*ast.CallExpr",
		},
		{
			name:      "parameterized min validator for float",
			rule:      "min",
			value:     "10.5",
			fieldType: "float64",
			wantType:  "*ast.CallExpr",
		},
		{
			name:      "unknown validator",
			rule:      "custom_validator",
			value:     "",
			fieldType: "string",
			wantType:  "*ast.CallExpr", // Should return GetValidator call
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := registry.GetValidatorExpression(tt.rule, tt.value, tt.fieldType)

			// Get actual type name
			gotType := getTypeName(expr)

			if gotType != tt.wantType {
				t.Errorf("GetValidatorExpression() returned %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestValidatorRegistry_GetRequiredImports(t *testing.T) {
	registry := NewValidatorRegistry()

	tests := []struct {
		name        string
		rules       []string
		wantImports []string
	}{
		{
			name:  "basic validators",
			rules: []string{"required", "email", "min"},
			wantImports: []string{
				"github.com/mateothegreat/go-validation",
				"github.com/mateothegreat/go-validation/rules",
			},
		},
		{
			name:  "no rules",
			rules: []string{},
			wantImports: []string{
				"github.com/mateothegreat/go-validation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := registry.GetRequiredImports(tt.rules)

			for _, want := range tt.wantImports {
				found := false
				for _, imp := range imports {
					if imp == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Missing expected import: %s", want)
				}
			}
		})
	}
}

// Helper function to get type name
func getTypeName(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.SelectorExpr:
		return "*ast.SelectorExpr"
	case *ast.CallExpr:
		return "*ast.CallExpr"
	case *ast.Ident:
		return "*ast.Ident"
	default:
		return "unknown"
	}
}
