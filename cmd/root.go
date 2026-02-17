// Package cmd contains the root command for the catls program.
// The root command is the entry point for the CLI application and handles setting up all available flags.
// It also provides command execution and configuration building.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/connerohnesorge/catls/internal/catls"
	"github.com/spf13/cobra"
)

// rootCmd is the main cobra command for catls.
// It defines the usage, short description, and long description of the application.
// The command supports various options for file listing, filtering, and output formatting.
var rootCmd = &cobra.Command{
	Use:   "catls [directory] [files...]",
	Short: "List files and their contents",
	Long: `catls recursively lists files and displays their contents in XML format.
It supports filtering by glob patterns, ignoring directories, and various output options.`,
	RunE: runCatls,
}

// Execute runs the root command and returns any error that occurs.
// The caller is responsible for handling errors and exit codes.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	setupFlags()
}

func setupFlags() {
	flags := rootCmd.Flags()

	flags.BoolP(
		"all",
		"a",
		false,
		"Include hidden files",
	)
	flags.BoolP(
		"recursive",
		"r",
		false,
		"Recursively list files in subdirectories",
	)
	flags.StringSlice(
		"ignore-dir",
		defaultIgnoreDirs(),
		"Ignore directory DIR (can be used multiple times)",
	)
	flags.StringSlice(
		"globs",
		nil,
		"Only include files matching glob pattern (can be used multiple times)",
	)
	flags.StringSlice(
		"ignore-globs",
		nil,
		"Ignore files matching glob pattern (can be used multiple times)",
	)
	flags.String(
		"pattern",
		"",
		"Only show lines matching glob PATTERN",
	)
	flags.BoolP(
		"line-numbers",
		"n",
		false,
		"Show line numbers",
	)
	flags.Bool(
		"debug",
		false,
		"Enable debug output",
	)
	flags.BoolP(
		"interactive",
		"I",
		false,
		"Interactive file selection mode",
	)
	flags.Bool(
		"omit-bins",
		false,
		"Skip binary files in output",
	)
	flags.StringP(
		"format",
		"f",
		"xml",
		"Output format: xml, json, markdown",
	)
	flags.String(
		"relative-to",
		"",
		"Display paths relative to this directory (default: scan directory)",
	)
}

func defaultIgnoreDirs() []string {
	return []string{
		"node_modules",
		".direnv",
		"build",
		"dist",
		"target",
		"venv",
		"env",
		".env",
		"vendor",
		".bundle",
		"coverage",
		"static",
	}
}

func runCatls(cmd *cobra.Command, args []string) error {
	cfg, err := buildConfig(cmd, args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	app := catls.New(cfg)

	return app.Run(ctx)
}

func buildConfig(cmd *cobra.Command, args []string) (*catls.Config, error) {
	flags := cmd.Flags()

	cfg := &catls.Config{
		Directory: ".",
	}

	if len(args) > 0 {
		cfg.Directory = args[0]
		cfg.Files = args[1:]
	}

	cfg.ShowAll, _ = flags.GetBool("all")
	cfg.Recursive, _ = flags.GetBool("recursive")
	cfg.Debug, _ = flags.GetBool("debug")
	cfg.Interactive, _ = flags.GetBool("interactive")
	cfg.ShowLineNumbers, _ = flags.GetBool("line-numbers")
	cfg.OmitBins, _ = flags.GetBool("omit-bins")
	cfg.ContentPattern, _ = flags.GetString("pattern")
	cfg.RelativeTo, _ = flags.GetString("relative-to")
	cfg.IgnoreDir, _ = flags.GetStringSlice("ignore-dir")
	cfg.Globs, _ = flags.GetStringSlice("globs")
	cfg.IgnoreGlobs, _ = flags.GetStringSlice("ignore-globs")

	// Handle output format
	formatStr, _ := flags.GetString("format")
	cfg.OutputFormat = catls.OutputFormat(formatStr)
	if !cfg.OutputFormat.IsValid() {
		return nil, fmt.Errorf("unsupported output format: %s (supported: %s)",
			formatStr, strings.Join(catls.GetSupportedFormats(), ", "))
	}

	return cfg, nil
}
