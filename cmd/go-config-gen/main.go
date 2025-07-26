package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	goconfig "github.com/mateothegreat/go-config"
)

var (
	inputDir    = flag.String("input", ".", "Input directory to scan for structs")
	outputFile  = flag.String("output", "validation_generated.go", "Output file for generated validation code")
	packageName = flag.String("package", "", "Package name for generated code (auto-detected if empty)")
	structName  = flag.String("struct", "", "Specific struct name to generate for (generates for all if empty)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	if *verbose {
		log.Println("🔧 go-config-gen - Zero Reflection Validation Code Generator")
		log.Printf("   Input directory: %s", *inputDir)
		log.Printf("   Output file: %s", *outputFile)
		log.Printf("   Package: %s", *packageName)
		log.Printf("   Struct filter: %s", *structName)
	}

	// Find all Go files in the input directory
	goFiles, err := findGoFiles(*inputDir)
	if err != nil {
		log.Fatalf("Error finding Go files: %v", err)
	}

	if *verbose {
		log.Printf("📁 Found %d Go files", len(goFiles))
	}

	// Parse files and find structs with validation tags
	structs, detectedPackage, err := findValidatedStructs(goFiles)
	if err != nil {
		log.Fatalf("Error parsing Go files: %v", err)
	}

	if len(structs) == 0 {
		log.Println("⚠️  No structs with validation tags found")
		return
	}

	// Use detected package name if not specified
	targetPackage := *packageName
	if targetPackage == "" {
		targetPackage = detectedPackage
	}

	if *verbose {
		log.Printf("🏗️  Found %d structs with validation tags", len(structs))
		for _, s := range structs {
			log.Printf("   - %s", s.Name)
		}
	}

	// Generate validation code
	generator := goconfig.NewValidatorGenerator(targetPackage)
	var generatedMethods []string

	for _, structInfo := range structs {
		// Filter by struct name if specified
		if *structName != "" && structInfo.Name != *structName {
			continue
		}

		if *verbose {
			log.Printf("⚙️  Generating validation for %s", structInfo.Name)
		}

		validationCode, err := generator.GenerateValidatorCode(structInfo.Code)
		if err != nil {
			log.Printf("❌ Error generating validation for %s: %v", structInfo.Name, err)
			continue
		}

		generatedMethods = append(generatedMethods, validationCode)
	}

	if len(generatedMethods) == 0 {
		log.Println("⚠️  No validation methods generated")
		return
	}

	// Write the generated code to output file
	err = writeGeneratedCode(*outputFile, targetPackage, generatedMethods)
	if err != nil {
		log.Fatalf("Error writing generated code: %v", err)
	}

	if *verbose {
		log.Printf("✅ Generated validation code written to %s", *outputFile)
		log.Printf("🎯 %d validation methods created", len(generatedMethods))
	} else {
		fmt.Printf("Generated %d validation methods in %s\n", len(generatedMethods), *outputFile)
	}
}

type StructInfo struct {
	Name string
	Code string
}

func findGoFiles(dir string) ([]string, error) {
	var goFiles []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Only include .go files, excluding test files and generated files
		if strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go") &&
			!strings.Contains(path, "generated") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

func findValidatedStructs(goFiles []string) ([]StructInfo, string, error) {
	var structs []StructInfo
	var packageName string

	for _, file := range goFiles {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Get package name from first file
		if packageName == "" {
			packageName = node.Name.Name
		}

		// Find structs with validation tags
		ast.Inspect(node, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				if st, ok := ts.Type.(*ast.StructType); ok {
					if hasValidationTags(st) {
						structCode := generateStructCode(ts, st)
						structs = append(structs, StructInfo{
							Name: ts.Name.Name,
							Code: structCode,
						})
					}
				}
			}
			return true
		})
	}

	return structs, packageName, nil
}

func hasValidationTags(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			if strings.Contains(tag, "validate:") {
				return true
			}
		}
	}
	return false
}

func generateStructCode(ts *ast.TypeSpec, st *ast.StructType) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("type %s struct {\n", ts.Name.Name))

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue // Skip embedded fields
		}

		fieldName := field.Names[0].Name
		fieldType := getFieldTypeName(field.Type)

		builder.WriteString(fmt.Sprintf("\t%s %s", fieldName, fieldType))

		if field.Tag != nil {
			builder.WriteString(" " + field.Tag.Value)
		}

		builder.WriteString("\n")
	}

	builder.WriteString("}")
	return builder.String()
}

func getFieldTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getFieldTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getFieldTypeName(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getFieldTypeName(t.Key), getFieldTypeName(t.Value))
	default:
		return "interface{}"
	}
}

func writeGeneratedCode(outputFile, packageName string, methods []string) error {
	var builder strings.Builder

	// Write file header
	builder.WriteString("// Code generated by go-config-gen. DO NOT EDIT.\n")
	builder.WriteString("// This file contains zero reflection validation methods.\n\n")
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Write imports
	builder.WriteString("import (\n")
	builder.WriteString("\t\"regexp\"\n")
	builder.WriteString("\tgoconfig \"github.com/mateothegreat/go-config\"\n")
	builder.WriteString(")\n\n")

	// Write regex patterns
	builder.WriteString("// Pre-compiled regex patterns for validation performance\n")
	builder.WriteString("var (\n")
	builder.WriteString("\temailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`)\n")
	builder.WriteString("\turlRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\\s]*$`)\n")
	builder.WriteString("\talphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)\n")
	builder.WriteString("\talphaNumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)\n")
	builder.WriteString("\tnumericRegex = regexp.MustCompile(`^[0-9]+$`)\n")
	builder.WriteString(")\n\n")

	// Write generated methods
	for i, method := range methods {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(method)
		builder.WriteString("\n")
	}

	return os.WriteFile(outputFile, []byte(builder.String()), 0o644)
}
