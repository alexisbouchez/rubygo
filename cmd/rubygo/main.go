// Package main provides the entry point for the Ruby interpreter.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alexisbouchez/rubylexer/evaluator"
	"github.com/alexisbouchez/rubylexer/lexer"
	"github.com/alexisbouchez/rubylexer/object"
	"github.com/alexisbouchez/rubylexer/parser"
	"github.com/alexisbouchez/rubylexer/repl"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// Start REPL
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	// Execute file
	filename := args[0]
	if err := runFile(filename); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func runFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "SyntaxError: %s\n", msg)
		}
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Set the current file for require_relative
	if absFilePath, pathErr := filepath.Abs(filename); pathErr == nil {
		evaluator.SetCurrentFile(absFilePath)
	} else {
		evaluator.SetCurrentFile(filename)
	}

	env := object.NewEnvironment()
	env.SetSelf(object.ObjectClass)

	result := evaluator.Eval(program, env)
	if err, ok := result.(*object.Error); ok {
		return fmt.Errorf("%s", err.Message)
	}

	return nil
}
