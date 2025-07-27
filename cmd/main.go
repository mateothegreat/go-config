package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mateothegreat/go-config/internal/generator"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go-validate",
		Short: "Fast, fluent, and flexible validation for Go configs",
		Long: `go-validate is a powerful code generation tool that generates zero-reflection,
allocation-conscious validation code directly from your Go structs.

Built on three foundational pillars:
- ⚡ Performance: Generates zero-reflection, allocation-conscious code
- ✨ Ergonomics: Offers intuitive interfaces and seamless config integration  
- 🔗 Fluency: Enables chainable APIs for expressive validation and error handling`,
		SilenceUsage: true,
	}

	cmd.AddCommand(
		newGenerateCmd(),
		newWatchCmd(),
		newBenchCmd(),
		newVersionCmd(),
	)

	return cmd
}

func newGenerateCmd() *cobra.Command {
	var (
		inputDir    string
		outputDir   string
		packageName string
		structs     []string
		dryRun      bool
		multi       bool
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Run one-time validator generation",
		Long: `Generate zero-reflection validation code for structs with validation tags.

The generate command scans Go source files for structs with validation tags
and generates optimized validation methods that avoid reflection entirely.`,
		Example: `  # Generate validation for all structs in current directory
  go-validate generate

  # Generate for specific structs only
  go-validate generate --structs MyConfig,ServerConfig

  # Preview generated code without writing files
  go-validate generate --dry-run

  # Generate separate files per struct
  go-validate generate --multi

  # Generate with custom output directory
  go-validate generate --output-dir ./generated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := generator.NewGenerator(
				generator.WithInputDir(inputDir),
				generator.WithOutputDir(outputDir),
				generator.WithPackage(packageName),
				generator.WithStructs(structs...),
				generator.WithDryRun(dryRun),
				generator.WithMulti(multi),
				generator.WithVerbose(verbose),
			)

			return gen.Generate()
		},
	}

	cmd.Flags().StringVar(&inputDir, "input-dir", ".", "Input directory to scan for structs")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "Custom base directory for output files")
	cmd.Flags().StringVar(&packageName, "package", "", "Package name for generated code (auto-detected if empty)")
	cmd.Flags().StringSliceVar(&structs, "structs", nil, "Target specific structs (e.g., --structs Foo,Bar)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Output to stdout without writing files")
	cmd.Flags().BoolVar(&multi, "multi", false, "Generate separate files per struct/package")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func newWatchCmd() *cobra.Command {
	var (
		inputDir    string
		outputDir   string
		packageName string
		structs     []string
		multi       bool
		cache       bool
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Auto-regenerate on file changes",
		Long: `Watch for file changes and automatically regenerate validation code.

The watch command monitors Go source files for changes and regenerates
validation code when structs with validation tags are modified.`,
		Example: `  # Watch current directory for changes
  go-validate watch

  # Watch with caching enabled for better performance
  go-validate watch --cache

  # Watch specific structs only
  go-validate watch --structs MyConfig,ServerConfig`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := generator.NewGenerator(
				generator.WithInputDir(inputDir),
				generator.WithOutputDir(outputDir),
				generator.WithPackage(packageName),
				generator.WithStructs(structs...),
				generator.WithMulti(multi),
				generator.WithCache(cache),
				generator.WithVerbose(verbose),
			)

			return watchFiles(gen, inputDir, verbose)
		},
	}

	cmd.Flags().StringVar(&inputDir, "input-dir", ".", "Input directory to scan for structs")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "Custom base directory for output files")
	cmd.Flags().StringVar(&packageName, "package", "", "Package name for generated code (auto-detected if empty)")
	cmd.Flags().StringSliceVar(&structs, "structs", nil, "Target specific structs (e.g., --structs Foo,Bar)")
	cmd.Flags().BoolVar(&multi, "multi", false, "Generate separate files per struct/package")
	cmd.Flags().BoolVar(&cache, "cache", false, "Cache AST scan results during watch mode")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func newBenchCmd() *cobra.Command {
	var (
		inputDir    string
		outputDir   string
		packageName string
		structs     []string
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "bench",
		Short: "Generate and run benchmark tests",
		Long: `Generate benchmark tests comparing reflection vs. generated validation paths.

The bench command generates benchmark tests that compare the performance
of reflection-based validation against the generated zero-reflection code.`,
		Example: `  # Generate and run benchmarks for all structs
  go-validate bench

  # Generate benchmarks for specific structs
  go-validate bench --structs MyConfig,ServerConfig

  # Generate benchmarks with custom output
  go-validate bench --output-dir ./benchmarks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := generator.NewGenerator(
				generator.WithInputDir(inputDir),
				generator.WithOutputDir(outputDir),
				generator.WithPackage(packageName),
				generator.WithStructs(structs...),
				generator.WithVerbose(verbose),
			)

			return generateBenchmarks(gen)
		},
	}

	cmd.Flags().StringVar(&inputDir, "input-dir", ".", "Input directory to scan for structs")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "Custom base directory for output files")
	cmd.Flags().StringVar(&packageName, "package", "", "Package name for generated code (auto-detected if empty)")
	cmd.Flags().StringSliceVar(&structs, "structs", nil, "Target specific structs (e.g., --structs Foo,Bar)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("go-validate %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	}
}

// watchFiles implements file watching functionality
func watchFiles(gen *generator.Generator, inputDir string, verbose bool) error {
	// This would implement file watching using fsnotify or similar
	// For now, return a placeholder error
	return fmt.Errorf("watch mode not yet implemented - use generate command for now")
}

// generateBenchmarks implements benchmark generation
func generateBenchmarks(gen *generator.Generator) error {
	// This would implement benchmark generation
	// For now, return a placeholder error
	return fmt.Errorf("benchmark generation not yet implemented - use generate command for now")
}