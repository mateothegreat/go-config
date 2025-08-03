package generator

import (
	"go/ast"
	"go/token"

	"github.com/mateothegreat/go-config/scanner"
)

// GeneratorConfig holds configuration for the code generator
type GeneratorConfig struct {
	InputDir    string
	OutputDir   string
	PackageName string
	Structs     []string
	DryRun      bool
	Multi       bool
	Cache       bool
	Verbose     bool
	UseAST      bool // Use AST-based code generation instead of templates
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Name       string
	Value      string
	Parameters []string
}

// CodeGenEngine is the main code generation engine
type CodeGenEngine struct {
	config  GeneratorConfig
	fileSet *token.FileSet
	structs []scanner.StructInfo
	imports map[string]bool
	cache   map[string]*ast.File
}

// TemplateData holds data for code generation templates
type TemplateData struct {
	PackageName string
	Imports     []string
	Structs     []scanner.StructInfo
	Timestamp   string
}

// ValidatorTemplate defines the structure for custom validator plugins
type ValidatorTemplate struct {
	Name        string
	Type        string
	Function    string
	Imports     []string
	Description string
}
