package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
)

// ASTScanner handles parsing Go source files and extracting struct information
type ASTScanner struct {
	fileSet *token.FileSet
	cache   map[string]*ast.File
	verbose bool
}

// NewASTScanner creates a new AST scanner
func NewASTScanner(verbose bool) *ASTScanner {
	return &ASTScanner{
		fileSet: token.NewFileSet(),
		cache:   make(map[string]*ast.File),
		verbose: verbose,
	}
}

// ScanDirectory scans a directory for Go files and extracts struct information
func (s *ASTScanner) ScanDirectory(dir string) ([]StructInfo, error) {
	goFiles, err := s.findGoFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	var allStructs []StructInfo

	if s.verbose {
		fmt.Printf("🔍 Found %d Go files to scan:\n", len(goFiles))
		for _, file := range goFiles {
			fmt.Printf("   - %s\n", file)
		}
	}

	for _, file := range goFiles {
		structs, err := s.scanFile(file)
		if err != nil {
			if s.verbose {
				fmt.Printf("Warning: failed to scan file %s: %v\n", file, err)
			}
			continue
		}
		allStructs = append(allStructs, structs...)
	}

	return allStructs, nil
}

// ScanFile scans a single Go file for struct information
func (s *ASTScanner) scanFile(filePath string) ([]StructInfo, error) {
	// Check cache first
	if cached, exists := s.cache[filePath]; exists {
		return s.extractStructsFromAST(cached, filePath)
	}

	// Parse the file
	node, err := parser.ParseFile(s.fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Cache the parsed AST
	s.cache[filePath] = node

	return s.extractStructsFromAST(node, filePath)
}

// extractStructsFromAST extracts struct information from parsed AST
func (s *ASTScanner) extractStructsFromAST(node *ast.File, filePath string) ([]StructInfo, error) {
	var structs []StructInfo
	packageName := node.Name.Name

	// Extract import paths
	imports := s.extractImports(node)

	// Find structs with validation annotations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				if s.hasValidationAnnotations(structType) || s.hasGoGenerateDirective(node, x.Name.Name) {
					structInfo := s.buildStructInfo(x, structType, packageName, filePath, imports)
					structs = append(structs, structInfo)
				}
			}
		}
		return true
	})

	return structs, nil
}

// hasValidationAnnotations checks if a struct has validation tags
func (s *ASTScanner) hasValidationAnnotations(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			if strings.Contains(tag, "validate:") {
				return true
			}
		}
	}
	return false
}

// hasGoGenerateDirective checks if there's a go:generate directive for the struct
func (s *ASTScanner) hasGoGenerateDirective(file *ast.File, structName string) bool {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if strings.Contains(comment.Text, "//go:generate") &&
				strings.Contains(comment.Text, "go-validate") &&
				strings.Contains(comment.Text, structName) {
				return true
			}
		}
	}
	return false
}

// buildStructInfo creates a StructInfo from AST nodes
func (s *ASTScanner) buildStructInfo(typeSpec *ast.TypeSpec, structType *ast.StructType, packageName, filePath string, imports []string) StructInfo {
	structInfo := StructInfo{
		Name:        typeSpec.Name.Name,
		PackageName: packageName,
		FilePath:    filePath,
		ImportPaths: imports,
		Fields:      make([]FieldInfo, 0),
	}

	// Extract struct comments
	if typeSpec.Doc != nil {
		structInfo.Comments = s.extractComments(typeSpec.Doc)
	}

	// Process each field
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue // Skip embedded fields for now
		}

		fieldInfo := s.buildFieldInfo(field)
		if fieldInfo.Name != "" {
			structInfo.Fields = append(structInfo.Fields, fieldInfo)
		}
	}

	return structInfo
}

// buildFieldInfo creates a FieldInfo from an AST field
func (s *ASTScanner) buildFieldInfo(field *ast.Field) FieldInfo {
	if len(field.Names) == 0 {
		return FieldInfo{} // Skip embedded fields
	}

	fieldInfo := FieldInfo{
		Name: field.Names[0].Name,
		Type: s.extractType(field.Type),
	}

	// Extract field comments
	if field.Doc != nil {
		fieldInfo.Comments = s.extractComments(field.Doc)
	}

	// Parse struct tags
	if field.Tag != nil {
		fieldInfo.Tag = strings.Trim(field.Tag.Value, "`")
		s.parseStructTags(&fieldInfo)
	}

	return fieldInfo
}

// parseStructTags parses struct tags and extracts validation rules
func (s *ASTScanner) parseStructTags(fieldInfo *FieldInfo) {
	if fieldInfo.Tag == "" {
		return
	}

	tag := reflect.StructTag(fieldInfo.Tag)

	// Extract common tags
	fieldInfo.JSONTag = tag.Get("json")
	fieldInfo.YAMLTag = tag.Get("yaml")
	fieldInfo.ConfigTag = tag.Get("config")

	// Parse validation rules
	validateTag := tag.Get("validate")
	if validateTag != "" {
		fieldInfo.ValidateRules = s.parseValidationRules(validateTag)
		fieldInfo.Required = s.hasRequiredRule(fieldInfo.ValidateRules)
	}
}

// parseValidationRules parses validation tag string into rules map
func (s *ASTScanner) parseValidationRules(tag string) map[string]string {
	rules := make(map[string]string)
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			rules[part] = ""
		}
	}

	return rules
}

// hasRequiredRule checks if the required rule is present
func (s *ASTScanner) hasRequiredRule(rules map[string]string) bool {
	_, exists := rules["required"]
	return exists
}

// extractType extracts the type string from an AST expression
func (s *ASTScanner) extractType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + s.extractType(t.X)
	case *ast.ArrayType:
		return "[]" + s.extractType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", s.extractType(t.Key), s.extractType(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", s.extractType(t.X), t.Sel.Name)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "interface{}"
	}
}

// extractImports extracts import paths from the AST
func (s *ASTScanner) extractImports(node *ast.File) []string {
	var imports []string

	for _, imp := range node.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, "\"")
			imports = append(imports, importPath)
		}
	}

	return imports
}

// extractComments extracts comment text from comment groups
func (s *ASTScanner) extractComments(commentGroup *ast.CommentGroup) string {
	if commentGroup == nil {
		return ""
	}

	var comments []string
	for _, comment := range commentGroup.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			comments = append(comments, text)
		}
	}

	return strings.Join(comments, " ")
}

// findGoFiles recursively finds all Go files in a directory
func (s *ASTScanner) findGoFiles(dir string) ([]string, error) {
	var goFiles []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, hidden directories, and common build directories (but never skip the root directory)
		if d.IsDir() && path != dir {
			name := d.Name()
			if strings.HasPrefix(name, ".") ||
				name == "vendor" ||
				name == "node_modules" ||
				name == "target" ||
				name == "dist" {
				if s.verbose {
					fmt.Printf("🚫 Skipping directory: %s\n", path)
				}
				return filepath.SkipDir
			}
		}

		// Include .go files, excluding test files and generated files
		if strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go") &&
			!strings.HasSuffix(path, "_generated.go") &&
			!strings.Contains(path, ".gen.") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

// ClearCache clears the AST cache
func (s *ASTScanner) ClearCache() {
	s.cache = make(map[string]*ast.File)
}
