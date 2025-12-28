// Package repl implements a Read-Eval-Print Loop for Ruby.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/alexisbouchez/rubylexer/evaluator"
	"github.com/alexisbouchez/rubylexer/lexer"
	"github.com/alexisbouchez/rubylexer/object"
	"github.com/alexisbouchez/rubylexer/parser"
)

const PROMPT = "irb> "

// Start starts the REPL.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	env.SetSelf(object.ObjectClass)

	fmt.Fprintln(out, "Ruby interpreter (rubygo)")
	fmt.Fprintln(out, "Type 'exit' to quit")
	fmt.Fprintln(out)

	var multilineBuffer strings.Builder
	inMultiline := false

	for {
		if inMultiline {
			fmt.Fprint(out, "...  ")
		} else {
			fmt.Fprint(out, PROMPT)
		}

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		// Handle exit
		if strings.TrimSpace(line) == "exit" || strings.TrimSpace(line) == "quit" {
			fmt.Fprintln(out, "Goodbye!")
			return
		}

		// Check for multiline input
		if inMultiline {
			multilineBuffer.WriteString("\n")
			multilineBuffer.WriteString(line)

			// Check if we should end multiline mode
			if isCompleteInput(multilineBuffer.String()) {
				line = multilineBuffer.String()
				multilineBuffer.Reset()
				inMultiline = false
			} else {
				continue
			}
		} else {
			if !isCompleteInput(line) {
				multilineBuffer.WriteString(line)
				inMultiline = true
				continue
			}
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			if evaluated.Type() != object.NIL_OBJ {
				fmt.Fprintln(out, "=> "+evaluated.Inspect())
			} else {
				fmt.Fprintln(out, "=> nil")
			}
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintln(out, "SyntaxError: "+msg)
	}
}

// isCompleteInput checks if the input is a complete Ruby expression.
func isCompleteInput(input string) bool {
	// Count block delimiters
	openBlocks := 0
	openParens := 0
	openBrackets := 0
	openBraces := 0
	inString := false
	stringDelim := byte(0)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if inString {
			if ch == stringDelim && (i == 0 || input[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			inString = true
			stringDelim = ch
		case '(':
			openParens++
		case ')':
			openParens--
		case '[':
			openBrackets++
		case ']':
			openBrackets--
		case '{':
			openBraces++
		case '}':
			openBraces--
		}
	}

	// Check for block keywords
	words := strings.Fields(input)
	for _, word := range words {
		switch word {
		case "def", "class", "module", "if", "unless", "case", "while", "until", "for", "begin", "do":
			openBlocks++
		case "end":
			openBlocks--
		}
	}

	// Complete if all delimiters are balanced
	return openParens == 0 && openBrackets == 0 && openBraces == 0 && openBlocks <= 0 && !inString
}

// EvalString evaluates a Ruby program string and returns the result.
func EvalString(input string) (object.Object, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	env := object.NewEnvironment()
	env.SetSelf(object.ObjectClass)

	result := evaluator.Eval(program, env)
	if err, ok := result.(*object.Error); ok {
		return nil, fmt.Errorf("%s", err.Message)
	}

	return result, nil
}
