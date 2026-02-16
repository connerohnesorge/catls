package catls

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/connerohnesorge/catls/internal/scanner"
)

// FileProcessor handles file content processing.
type FileProcessor struct {
	typeDetector TypeDetector
}

// ProcessedFile represents a file after processing.
type ProcessedFile struct {
	Info        scanner.FileInfo
	FileType    string
	Lines       []FilteredLine
	TotalLines  int
	IsTruncated bool
	Error       error
}

// TypeDetector defines interface for detecting file types.
// TypeDetector is used to identify the type of a file.
type TypeDetector interface {
	// DetectType returns the file type for the given file path.
	// DetectType returns the file type for the given file path.
	DetectType(filePath string) string
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		typeDetector: &ExtensionTypeDetector{},
	}
}

// ProcessFile processes a single file and returns its content.
func (p *FileProcessor) ProcessFile(file scanner.FileInfo, filter *FileFilter) ProcessedFile {
	result := ProcessedFile{
		Info: file,
	}

	if file.IsBinary {
		return result
	}

	// Detect file type
	result.FileType = p.typeDetector.DetectType(file.Path)

	// Read file content
	lines, err := p.readFileLines(file.Path)
	if err != nil {
		result.Error = err

		return result
	}

	result.TotalLines = len(lines)

	// Apply content filtering
	filteredLines := filter.FilterContent(lines)

	// Check if we need to truncate for display
	const maxDisplayLines = 1000
	const truncateToLines = 100

	if len(filteredLines) > maxDisplayLines && filter.contentPattern == nil {
		result.Lines = filteredLines[:truncateToLines]
		result.IsTruncated = true
	} else {
		result.Lines = filteredLines
	}

	return result
}

// readFileLines reads all lines from a file.
func (*FileProcessor) readFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log close error - in a real app you'd use a proper logger
			fmt.Fprintf(os.Stderr, "Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	var lines []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// ExtensionTypeDetector detects file types based on extensions.
type ExtensionTypeDetector struct{}

// DetectType implements TypeDetector.
func (*ExtensionTypeDetector) DetectType(filePath string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	typeMap := map[string]string{
		"sh":           langBash,
		langBash:       langBash,
		"rb":           langRuby,
		"py":           langPython,
		"js":           langJavaScript,
		"ts":           langTypeScript,
		"jsx":          langJavaScript,
		"tsx":          langTypeScript,
		langHTML:       langHTML,
		"htm":          langHTML,
		langNix:        langNix,
		langCSS:        langCSS,
		"scss":         langSCSS,
		"sass":         langSCSS,
		langJSON:       langJSON,
		"md":           langMarkdown,
		langMarkdown:   langMarkdown,
		langXML:        langXML,
		langC:          langC,
		langCPP:        langCPP,
		"cxx":          langCPP,
		"cc":           langCPP,
		"h":            langC,
		"hpp":          langCPP,
		"hxx":          langCPP,
		langTOML:       langTOML,
		langJava:       langJava,
		"rs":           langRust,
		langGo:         langGo,
		langPHP:        langPHP,
		"pl":           langPerl,
		langSQL:        langSQL,
		"templ":        langGo,
		"yml":          langYAML,
		langYAML:       langYAML,
		langDockerfile: langDockerfile,
		langMakefile:   langMakefile,
	}

	if fileType, exists := typeMap[ext]; exists {
		return fileType
	}

	return ""
}
