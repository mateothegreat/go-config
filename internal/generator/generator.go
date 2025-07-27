package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mateothegreat/go-config/internal/scanner"
)

// Generator is the main orchestrator for code generation
type Generator struct {
	config    GeneratorConfig
	scanner   *scanner.ASTScanner
	codegen   *CodeGenerator
}

// NewGenerator creates a new generator with the given configuration
func NewGenerator(options ...GeneratorOption) *Generator {
	config := GeneratorConfig{
		InputDir:    ".",
		OutputDir:   ".",
		PackageName: "",
		DryRun:      false,
		Multi:       false,
		Cache:       false,
		Verbose:     false,
	}

	// Apply options
	for _, option := range options {
		option(&config)
	}

	return &Generator{
		config:  config,
		scanner: scanner.NewASTScanner(config.Verbose),
		codegen: NewCodeGenerator(config),
	}
}

// GeneratorOption is a function that configures the generator
type GeneratorOption func(*GeneratorConfig)

// WithInputDir sets the input directory
func WithInputDir(dir string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.InputDir = dir
	}
}

// WithOutputDir sets the output directory
func WithOutputDir(dir string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.OutputDir = dir
	}
}

// WithPackage sets the package name
func WithPackage(pkg string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.PackageName = pkg
	}
}

// WithStructs sets specific struct names to generate for
func WithStructs(structs ...string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.Structs = structs
	}
}

// WithDryRun enables dry run mode
func WithDryRun(dryRun bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.DryRun = dryRun
	}
}

// WithMulti enables multi-file generation
func WithMulti(multi bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.Multi = multi
	}
}

// WithCache enables AST caching
func WithCache(cache bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.Cache = cache
	}
}

// WithVerbose enables verbose output
func WithVerbose(verbose bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.Verbose = verbose
	}
}

// Generate scans for structs and generates validation code
func (g *Generator) Generate() error {
	if g.config.Verbose {
		fmt.Printf("🔧 go-validate - Zero Reflection Validation Code Generator\n")
		fmt.Printf("   Input directory: %s\n", g.config.InputDir)
		fmt.Printf("   Output directory: %s\n", g.config.OutputDir)
		fmt.Printf("   Package: %s\n", g.config.PackageName)
		fmt.Printf("   Dry run: %v\n", g.config.DryRun)
		fmt.Printf("   Multi-file: %v\n", g.config.Multi)
	}

	// Scan for structs
	structs, err := g.scanner.ScanDirectory(g.config.InputDir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(structs) == 0 {
		if g.config.Verbose {
			fmt.Println("⚠️  No structs with validation tags found")
		}
		return nil
	}

	// Filter structs if specific names are provided
	if len(g.config.Structs) > 0 {
		structs = g.filterStructs(structs)
	}

	if len(structs) == 0 {
		if g.config.Verbose {
			fmt.Println("⚠️  No matching structs found after filtering")
		}
		return nil
	}

	// Determine package name if not set
	packageName := g.config.PackageName
	if packageName == "" && len(structs) > 0 {
		packageName = structs[0].PackageName
	}

	if g.config.Verbose {
		fmt.Printf("🏗️  Found %d structs with validation tags\n", len(structs))
		for _, s := range structs {
			fmt.Printf("   - %s\n", s.Name)
		}
	}

	// Generate code
	if g.config.Multi {
		return g.generateMultipleFiles(structs, packageName)
	} else {
		return g.generateSingleFile(structs, packageName)
	}
}

// generateSingleFile generates a single validation file with all structs
func (g *Generator) generateSingleFile(structs []scanner.StructInfo, packageName string) error {
	content, err := g.codegen.GenerateFileContent(structs, packageName)
	if err != nil {
		return fmt.Errorf("failed to generate file content: %w", err)
	}

	if g.config.DryRun {
		fmt.Println("// Generated code (dry run):")
		fmt.Println(content)
		return nil
	}

	outputPath := filepath.Join(g.config.OutputDir, "validation_generated.go")
	
	if err := g.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if g.config.Verbose {
		fmt.Printf("✅ Generated validation code written to %s\n", outputPath)
		fmt.Printf("🎯 %d validation methods created\n", len(structs))
	}

	return nil
}

// generateMultipleFiles generates separate validation files for each struct
func (g *Generator) generateMultipleFiles(structs []scanner.StructInfo, packageName string) error {
	if err := g.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, structInfo := range structs {
		content, err := g.codegen.GenerateFileContent([]scanner.StructInfo{structInfo}, packageName)
		if err != nil {
			if g.config.Verbose {
				fmt.Printf("❌ Error generating validation for %s: %v\n", structInfo.Name, err)
			}
			continue
		}

		fileName := fmt.Sprintf("%s_validation_generated.go", strings.ToLower(structInfo.Name))
		outputPath := filepath.Join(g.config.OutputDir, fileName)

		if g.config.DryRun {
			fmt.Printf("// Generated code for %s (dry run):\n", structInfo.Name)
			fmt.Println(content)
			continue
		}

		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			if g.config.Verbose {
				fmt.Printf("❌ Error writing file for %s: %v\n", structInfo.Name, err)
			}
			continue
		}

		if g.config.Verbose {
			fmt.Printf("✅ Generated validation for %s written to %s\n", structInfo.Name, outputPath)
		}
	}

	return nil
}

// GenerateBenchmarks generates benchmark tests for validation methods
func (g *Generator) GenerateBenchmarks(structs []scanner.StructInfo, packageName string) (string, error) {
	// Implementation for benchmark generation would go here
	// This is a placeholder for the benchmarking functionality mentioned in CLAUDE.md
	return "", fmt.Errorf("benchmark generation not yet implemented")
}

// GenerateTests generates test files for validation methods
func (g *Generator) GenerateTests(structs []scanner.StructInfo, packageName string) (string, error) {
	// Implementation for test generation would go here
	// This is a placeholder for test generation functionality
	return "", fmt.Errorf("test generation not yet implemented")
}

// filterStructs filters structs based on the configured struct names
func (g *Generator) filterStructs(structs []scanner.StructInfo) []scanner.StructInfo {
	structMap := make(map[string]bool)
	for _, name := range g.config.Structs {
		structMap[name] = true
	}

	var filtered []scanner.StructInfo
	for _, structInfo := range structs {
		if structMap[structInfo.Name] {
			filtered = append(filtered, structInfo)
		}
	}

	return filtered
}

// ensureOutputDir creates the output directory if it doesn't exist
func (g *Generator) ensureOutputDir() error {
	if g.config.OutputDir == "" {
		g.config.OutputDir = "."
	}

	info, err := os.Stat(g.config.OutputDir)
	if os.IsNotExist(err) {
		return os.MkdirAll(g.config.OutputDir, 0755)
	}

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("output path %s exists but is not a directory", g.config.OutputDir)
	}

	return nil
}

// ClearCache clears the AST cache
func (g *Generator) ClearCache() {
	if !g.config.Cache {
		g.scanner.ClearCache()
	}
}