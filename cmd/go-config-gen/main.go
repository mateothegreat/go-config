package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mateothegreat/go-config/generator"
)

var (
	inputDir    = flag.String("input", ".", "Input directory to scan for structs")
	outputDir   = flag.String("output", ".", "Output directory for generated files")
	packageName = flag.String("package", "", "Package name for generated code (auto-detected if empty)")
	structNames = flag.String("structs", "", "Comma-separated list of struct names to generate for (all if empty)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	useAST      = flag.Bool("ast", true, "Use AST-based code generation (recommended)")
	multi       = flag.Bool("multi", false, "Generate separate files for each struct")
	dryRun      = flag.Bool("dry-run", false, "Print generated code without writing files")
	cache       = flag.Bool("cache", false, "Enable AST caching for better performance")
)

func main() {
	flag.Parse()

	// Parse struct names if provided
	var structs []string
	if *structNames != "" {
		structs = strings.Split(*structNames, ",")
		for i := range structs {
			structs[i] = strings.TrimSpace(structs[i])
		}
	}

	// Create generator options
	opts := []generator.GeneratorOption{
		generator.WithInputDir(*inputDir),
		generator.WithOutputDir(*outputDir),
		generator.WithVerbose(*verbose),
		generator.WithDryRun(*dryRun),
		generator.WithMulti(*multi),
		generator.WithCache(*cache),
	}

	if *packageName != "" {
		opts = append(opts, generator.WithPackage(*packageName))
	}

	if len(structs) > 0 {
		opts = append(opts, generator.WithStructs(structs...))
	}

	// Create and run generator
	gen := generator.NewGenerator(opts...)

	// Log configuration if verbose
	if *verbose {
		fmt.Println("🔧 go-config-gen - Zero Reflection Validation Code Generator")
		fmt.Printf("   Mode: %s\n", func() string {
			if *useAST {
				return "AST-based (high-performance)"
			}
			return "Template-based (deprecated)"
		}())
		fmt.Printf("   Input directory: %s\n", *inputDir)
		fmt.Printf("   Output directory: %s\n", *outputDir)
		if *packageName != "" {
			fmt.Printf("   Package: %s\n", *packageName)
		}
		if len(structs) > 0 {
			fmt.Printf("   Structs: %s\n", strings.Join(structs, ", "))
		}
		fmt.Printf("   Multi-file: %v\n", *multi)
		fmt.Printf("   Dry run: %v\n", *dryRun)
		fmt.Printf("   Cache: %v\n", *cache)
	}

	// Generate validation code
	if err := gen.Generate(); err != nil {
		log.Fatalf("Error generating validation code: %v", err)
	}

	// Clean up cache if not enabled
	if !*cache {
		gen.ClearCache()
	}

	if !*dryRun && !*verbose {
		fmt.Println("✅ Validation code generated successfully")
	}
}

// init ensures the generator uses AST mode by default
func init() {
	// Override the default value after parsing
	oldUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-config-gen - Generate zero-reflection validation code\n\n")
		fmt.Fprintf(os.Stderr, "This tool generates high-performance validation code that integrates with\n")
		fmt.Fprintf(os.Stderr, "the go-validation library without using runtime reflection.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go-config-gen [flags]\n")
		fmt.Fprintf(os.Stderr, "  go generate (when using //go:generate directive)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate for all structs in current directory\n")
		fmt.Fprintf(os.Stderr, "  go-config-gen\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate for specific structs\n")
		fmt.Fprintf(os.Stderr, "  go-config-gen -structs ServerConfig,DatabaseConfig\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate separate files for each struct\n")
		fmt.Fprintf(os.Stderr, "  go-config-gen -multi\n\n")
		fmt.Fprintf(os.Stderr, "  # Preview generated code without writing files\n")
		fmt.Fprintf(os.Stderr, "  go-config-gen -dry-run -verbose\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		oldUsage()
	}
}
