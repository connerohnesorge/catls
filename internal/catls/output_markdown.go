// Package catls implements the core functionality for concatenating and formatting file listings.
package catls

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// MarkdownOutput handles Markdown output formatting. It implements the OutputFormatter interface to generate
// Markdown-formatted output with syntax-highlighted code blocks. The formatter intelligently detects programming
// languages for proper syntax highlighting based on file types and extensions.
type MarkdownOutput struct {
	// firstFile tracks whether this is the first file being written to avoid extra spacing.
	firstFile bool
}

// NewMarkdownOutput creates a new Markdown output formatter for generating syntax-highlighted file listings.
// The formatter tracks the first file to avoid unnecessary spacing at the beginning of output.
func NewMarkdownOutput() *MarkdownOutput {
	return &MarkdownOutput{
		firstFile: true,
	}
}

// WriteHeader writes the opening Markdown structure (no-op for Markdown). Markdown doesn't require headers.
func (*MarkdownOutput) WriteHeader(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// No header needed for Markdown
	return nil
}

// WriteFile writes a single processed file to Markdown output.
func (o *MarkdownOutput) WriteFile(ctx context.Context, file *ProcessedFile, cfg *Config) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Add spacing between files (except for the first file)
	if !o.firstFile {
		fmt.Println()
	}
	o.firstFile = false

	// Write file header
	fmt.Printf("## %s\n\n", file.Info.RelPath)

	// Handle errors
	if file.Error != nil {
		fmt.Printf("**Error:** %s\n\n", file.Error.Error())

		return nil
	}

	// Handle binary files
	if file.Info.IsBinary {
		fmt.Println("*Binary file - contents not displayed*")

		return nil
	}

	// Determine language for syntax highlighting
	language := o.getLanguageForSyntaxHighlighting(file.FileType, file.Info.RelPath)

	// Write code block with content
	fmt.Printf("```%s name=\"%s\"\n", language, filepath.Base(file.Info.RelPath))

	// Write content lines
	for _, line := range file.Lines {
		if cfg.ShowLineNumbers {
			fmt.Printf("%4d| %s\n", line.LineNumber, line.Content)
		} else {
			fmt.Println(line.Content)
		}
	}

	// Handle truncation
	if file.IsTruncated {
		remainingLines := file.TotalLines - len(file.Lines)
		if remainingLines > 0 {
			fmt.Printf("... (%d more lines)\n", remainingLines)
		}
	}

	fmt.Println("```")

	return nil
}

// WriteFooter writes the closing Markdown structure (no-op for Markdown).
func (*MarkdownOutput) WriteFooter(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// No footer needed for Markdown
	return nil
}

// Language constants for syntax highlighting.
const (
	langBash       = "bash"
	langRuby       = "ruby"
	langPython     = "python"
	langJavaScript = "javascript"
	langTypeScript = "typescript"
	langHTML       = "html"
	langNix        = "nix"
	langCSS        = "css"
	langSCSS       = "scss"
	langJSON       = "json"
	langMarkdown   = "markdown"
	langXML        = "xml"
	langC          = "c"
	langCPP        = "cpp"
	langTOML       = "toml"
	langJava       = "java"
	langRust       = "rust"
	langGo         = "go"
	langPHP        = "php"
	langPerl       = "perl"
	langSQL        = "sql"
	langYAML       = "yaml"
	langDockerfile = "dockerfile"
	langMakefile   = "makefile"
	langText       = "text"
)

// getLanguageForSyntaxHighlighting maps file types to syntax highlighting languages.
func (o *MarkdownOutput) getLanguageForSyntaxHighlighting(fileType, filePath string) string {
	// Use the detected file type first
	if fileType != "" {
		lang := o.languageFromFileType(fileType)
		if lang != "" {
			return lang
		}
	}

	// Fallback to extension-based detection
	return o.languageFromExtension(filePath)
}

// languageFromFileType returns the language for a detected file type.
func (*MarkdownOutput) languageFromFileType(fileType string) string {
	langMap := map[string]string{
		langBash:       langBash,
		langRuby:       langRuby,
		langPython:     langPython,
		langJavaScript: langJavaScript,
		langTypeScript: langTypeScript,
		langHTML:       langHTML,
		langNix:        langNix,
		langCSS:        langCSS,
		"scss":         langSCSS,
		"sass":         langSCSS,
		langJSON:       langJSON,
		langMarkdown:   langMarkdown,
		langXML:        langXML,
		langC:          langC,
		langCPP:        langCPP,
		langTOML:       langTOML,
		langJava:       langJava,
		langRust:       langRust,
		langGo:         langGo,
		langPHP:        langPHP,
		langPerl:       langPerl,
		langSQL:        langSQL,
		"templ":        langGo,
		langYAML:       langYAML,
		langDockerfile: langDockerfile,
		langMakefile:   langMakefile,
	}

	if lang, ok := langMap[fileType]; ok {
		return lang
	}

	return ""
}

// languageFromExtension returns the language based on file extension.
func (*MarkdownOutput) languageFromExtension(filePath string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	extMap := map[string]string{
		"sh":             langBash,
		langBash:         langBash,
		"zsh":            langBash,
		"rb":             langRuby,
		"py":             langPython,
		"js":             langJavaScript,
		"jsx":            langJavaScript,
		"ts":             langTypeScript,
		"tsx":            langTypeScript,
		langHTML:         langHTML,
		"htm":            langHTML,
		langNix:          langNix,
		langCSS:          langCSS,
		"scss":           langSCSS,
		"sass":           langSCSS,
		langJSON:         langJSON,
		"md":             langMarkdown,
		langMarkdown:     langMarkdown,
		langXML:          langXML,
		langC:            langC,
		"h":              langC,
		langCPP:          langCPP,
		"cxx":            langCPP,
		"cc":             langCPP,
		"hpp":            langCPP,
		"hxx":            langCPP,
		langTOML:         langTOML,
		langJava:         langJava,
		"rs":             langRust,
		langGo:           langGo,
		langPHP:          langPHP,
		"pl":             langPerl,
		langSQL:          langSQL,
		"templ":          langGo,
		"yml":            langYAML,
		langYAML:         langYAML,
		langDockerfile:   langDockerfile,
		langMakefile:     langMakefile,
		"txt":            langText,
		"":               langText,
	}

	if lang, ok := extMap[ext]; ok {
		return lang
	}

	return langText
}
