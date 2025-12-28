package evaluator

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexisbouchez/rubylexer/lexer"
	"github.com/alexisbouchez/rubylexer/object"
	"github.com/alexisbouchez/rubylexer/parser"
)

var (
	loadedFiles      = make(map[string]bool)
	loadedFilesMutex sync.Mutex
	loadPath         = []string{"."}
	currentFile      = ""
)

// SetLoadPath sets the load path for require
func SetLoadPath(paths []string) {
	loadPath = paths
}

// AddToLoadPath adds a path to the load path
func AddToLoadPath(path string) {
	loadPath = append(loadPath, path)
}

// SetCurrentFile sets the current file being executed
func SetCurrentFile(path string) {
	currentFile = path
}

// GetCurrentFile returns the current file being executed
func GetCurrentFile() string {
	return currentFile
}

// RequireFile loads and evaluates a Ruby file
func RequireFile(filename string, env *object.Environment) object.Object {
	loadedFilesMutex.Lock()
	defer loadedFilesMutex.Unlock()

	// Add .rb extension if not present
	if !strings.HasSuffix(filename, ".rb") {
		filename = filename + ".rb"
	}

	// Find the file in load path
	fullPath, err := findFile(filename)
	if err != nil {
		return newError("cannot load such file -- %s", filename)
	}

	// Check if already loaded
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		absPath = fullPath
	}
	if loadedFiles[absPath] {
		return object.FALSE
	}

	// Load and evaluate
	result := loadAndEval(fullPath, env)
	if isError(result) {
		return result
	}

	loadedFiles[absPath] = true
	return object.TRUE
}

// RequireRelativeFile loads a file relative to the current file
func RequireRelativeFile(filename string, env *object.Environment) object.Object {
	loadedFilesMutex.Lock()
	defer loadedFilesMutex.Unlock()

	// Add .rb extension if not present
	if !strings.HasSuffix(filename, ".rb") {
		filename = filename + ".rb"
	}

	// Resolve relative to current file
	var fullPath string
	if currentFile != "" {
		dir := filepath.Dir(currentFile)
		fullPath = filepath.Join(dir, filename)
	} else {
		fullPath = filename
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return newError("cannot load such file -- %s", filename)
	}

	// Check if already loaded
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		absPath = fullPath
	}
	if loadedFiles[absPath] {
		return object.FALSE
	}

	// Load and evaluate
	result := loadAndEval(fullPath, env)
	if isError(result) {
		return result
	}

	loadedFiles[absPath] = true
	return object.TRUE
}

// LoadFile loads and evaluates a file (always reloads, unlike require)
func LoadFile(filename string, env *object.Environment) object.Object {
	// Add .rb extension if not present
	if !strings.HasSuffix(filename, ".rb") {
		filename = filename + ".rb"
	}

	// Find the file
	fullPath, err := findFile(filename)
	if err != nil {
		return newError("cannot load such file -- %s", filename)
	}

	return loadAndEval(fullPath, env)
}

func findFile(filename string) (string, error) {
	// Try absolute path first
	if filepath.IsAbs(filename) {
		if _, err := os.Stat(filename); err == nil {
			return filename, nil
		}
	}

	// Search in load path
	for _, path := range loadPath {
		fullPath := filepath.Join(path, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// Try current directory
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	return "", os.ErrNotExist
}

func loadAndEval(filename string, env *object.Environment) object.Object {
	content, err := os.ReadFile(filename)
	if err != nil {
		return newError("cannot read file: %s", err)
	}

	// Save and restore current file
	oldFile := currentFile
	absPath, _ := filepath.Abs(filename)
	currentFile = absPath
	defer func() { currentFile = oldFile }()

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return newError("parse error in %s: %s", filename, p.Errors()[0])
	}

	return Eval(program, env)
}

func init() {
	// Update Kernel builtins with real require implementation
	// This is done by modifying the getKernelBuiltins to use our RequireFile
}
