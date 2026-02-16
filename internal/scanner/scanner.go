// Package scanner handles file discovery and analysis.
package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileInfo represents information about a discovered file.
type FileInfo struct {
	Path     string // Path to the file
	RelPath  string // Relative path to the file
	IsBinary bool   // Whether the file is a binary file.
}

// Config holds scanner configuration.
type Config struct {
	Directory   string   // Directory to scan
	ShowAll     bool     // ShowAll option
	Recursive   bool     // Recursive option
	IgnoreDir   []string // IgnoreDir option
	IgnoreGlobs []string // IgnoreGlobs option
	Debug       bool     // Debug logging
	RelativeTo  string   // Base directory for relative paths (empty means use Directory)
}

// Scanner handles file discovery and filtering.
type Scanner struct {
	binaryDetector BinaryDetector
}

// New creates a new scanner.
func New() *Scanner {
	return &Scanner{
		binaryDetector: &FileBinaryDetector{},
	}
}

// Scan discovers files according to configuration.
func (s *Scanner) Scan(ctx context.Context, cfg *Config) ([]FileInfo, error) {
	var files []FileInfo
	maxDepth := 1
	if cfg.Recursive {
		maxDepth = -1
	}

	stack := []dirEntry{{cfg.Directory, 0}}

	scanCtx := &scanContext{
		cfg:   cfg,
		stack: &stack,
		files: &files,
	}

	for len(stack) > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if maxDepth != -1 && current.depth >= maxDepth {
			continue
		}

		s.scanDirectory(current.path, current.depth, scanCtx)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	return files, nil
}

type dirEntry struct {
	path  string
	depth int
}

type scanContext struct {
	cfg   *Config
	stack *[]dirEntry
	files *[]FileInfo
}

func (s *Scanner) scanDirectory(path string, depth int, ctx *scanContext) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if ctx.cfg.Debug {
			fmt.Fprintf(os.Stderr, "Error accessing directory %s: %v\n", path, err)
		}

		return
	}

	// Sort entries for consistent output
	entryNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryNames = append(entryNames, entry.Name())
	}
	sort.Strings(entryNames)

	for _, entryName := range entryNames {
		if entryName == "." || entryName == ".." {
			continue
		}

		if !ctx.cfg.ShowAll && strings.HasPrefix(entryName, ".") {
			continue
		}

		fullPath := filepath.Join(path, entryName)
		s.processEntry(fullPath, depth, ctx)
	}
}

func (s *Scanner) processEntry(fullPath string, currentDepth int, ctx *scanContext) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return
	}

	if info.IsDir() {
		if !s.shouldIgnoreDir(fullPath, ctx.cfg) {
			*ctx.stack = append(*ctx.stack, dirEntry{fullPath, currentDepth + 1})
		} else if ctx.cfg.Debug {
			fmt.Fprintf(os.Stderr, "Debug: Ignoring directory: %s\n", fullPath)
		}
	} else if info.Mode().IsRegular() {
		relPath, err := s.getRelativePath(fullPath, ctx.cfg)
		if err != nil {
			return
		}

		isBinary := s.binaryDetector.IsBinary(fullPath)

		*ctx.files = append(*ctx.files, FileInfo{
			Path:     fullPath,
			RelPath:  relPath,
			IsBinary: isBinary,
		})
	}
}

// getRelativePath returns the relative path from base directory.
func (*Scanner) getRelativePath(fullPath string, cfg *Config) (string, error) {
	baseDir := cfg.Directory

	// If RelativeTo is set, use it instead of the scan directory
	if cfg.RelativeTo != "" {
		baseDir = cfg.RelativeTo
	}

	if baseDir == "." {
		return fullPath, nil
	}

	return filepath.Rel(baseDir, fullPath)
}
