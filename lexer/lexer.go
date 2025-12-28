// Package lexer implements a Ruby lexer.
package lexer

import (
	"strings"

	"github.com/alexisbouchez/rubylexer/token"
)

// stringMode represents the type of string being lexed
type stringMode int

const (
	modeNone stringMode = iota
	modeSingleQuote
	modeDoubleQuote
	modeBacktick
	modeRegexp
	modePercentQ       // %q
	modePercentQUpper  // %Q
	modePercentW       // %w
	modePercentWUpper  // %W
	modePercentI       // %i
	modePercentIUpper  // %I
	modePercentR       // %r
	modePercentS       // %s
	modePercentX       // %x
	modePercentDefault // %( )
	modeHeredoc
	modeSymbolDoubleQuote
	modeSymbolSingleQuote
)

// stringState tracks the state of string lexing
type stringState struct {
	mode            stringMode
	terminator      byte
	openDelimiter   byte
	nestingLevel    int
	interpolating   bool
	heredocIdent    string
	heredocIndented bool
	heredocSquiggle bool
	heredocQuoted   bool
	savedBraceDepth int // Saved brace depth when entering string during interpolation
}

// Lexer represents a lexer for Ruby source code.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int
	prevColumn   int // column before advancing (for newline handling)

	// String lexing state
	stringStack  []stringState
	currentState *stringState

	// Interpolation tracking
	braceDepth int

	// Context for disambiguation
	afterOperator      bool
	afterKeyword       bool
	afterIdent         bool
	afterRightParen    bool
	afterRightBracket  bool
	inLabelContext     bool
	sawNewline         bool
	startOfLine        bool
	afterDot           bool

	// Heredoc queue for deferred processing
	heredocQueue []stringState
	heredocPos   int
}

// New creates a new Lexer instance.
func New(input string) *Lexer {
	l := &Lexer{
		input:       input,
		line:        1,
		column:      0,
		stringStack: make([]stringState, 0),
		startOfLine: true,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	l.prevColumn = l.column
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// If we're inside a string, handle string content
	if l.currentState != nil {
		return l.lexStringContent()
	}

	// Handle heredoc body after line continuation
	if len(l.heredocQueue) > 0 && l.sawNewline {
		return l.lexHeredocBody()
	}

	l.skipWhitespace()

	startLine := l.line
	startColumn := l.column
	startOffset := l.position

	switch l.ch {
	case '\n':
		tok = l.newToken(token.NEWLINE, "\n")
		l.readChar()
		l.sawNewline = true
		l.startOfLine = true
		l.afterOperator = false
		l.afterKeyword = false
		l.afterIdent = false
		l.afterRightParen = false
		l.afterRightBracket = false
		l.afterDot = false
		// If we have pending heredocs, start processing them
		if len(l.heredocQueue) > 0 {
			// Return the newline first, heredoc will be processed on next call
		}
		return l.setTokenPosition(tok, startLine, startColumn, startOffset)
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
	case '+':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.PLUS_EQUAL, "+=")
			l.afterOperator = true // After +=, we start an expression
		} else {
			tok = l.newToken(token.PLUS, "+")
			l.afterOperator = false // After +, / is division not regexp
		}
		l.afterIdent = false
		l.readChar()
	case '-':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.MINUS_EQUAL, "-=")
			l.afterOperator = true // After -=, we start an expression
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.MINUS_GREATER, "->")
			l.afterOperator = true // After ->, we can have regexp in block
		} else {
			tok = l.newToken(token.MINUS, "-")
			l.afterOperator = false // After -, / is division not regexp
		}
		l.afterIdent = false
		l.readChar()
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = l.newToken(token.STAR_STAR_EQUAL, "**=")
				l.afterOperator = true // After **=, we start an expression
			} else {
				tok = l.newToken(token.STAR_STAR, "**")
				l.afterOperator = false // After **, / is division
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.STAR_EQUAL, "*=")
			l.afterOperator = true // After *=, we start an expression
		} else {
			tok = l.newToken(token.STAR, "*")
			l.afterOperator = false // After *, / is division
		}
		l.afterIdent = false
		l.readChar()
	case '/':
		// Check for /= first (compound assignment)
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.SLASH_EQUAL, "/=")
			l.afterOperator = true
			l.readChar()
		} else if l.shouldLexRegexp() {
			// Check if this is a regexp
			tok = l.lexRegexp()
		} else {
			tok = l.newToken(token.SLASH, "/")
			l.afterOperator = false // After /, / would be division
			l.readChar()
		}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.PERCENT_EQUAL, "%=")
			l.afterOperator = true
			l.readChar()
		} else if l.isPercentLiteral() {
			tok = l.lexPercentLiteral()
			l.afterOperator = false
			l.afterIdent = true
		} else {
			tok = l.newToken(token.PERCENT, "%")
			l.afterOperator = false // After %, / should be division, not regexp
			l.afterIdent = true
			l.readChar()
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = l.newToken(token.AMPERSAND_AMPERSAND_EQUAL, "&&=")
			} else {
				tok = l.newToken(token.AMPERSAND_AMPERSAND, "&&")
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.AMPERSAND_EQUAL, "&=")
		} else if l.peekChar() == '.' {
			l.readChar()
			tok = l.newToken(token.AMPERSAND_DOT, "&.")
		} else {
			tok = l.newToken(token.AMPERSAND, "&")
		}
		l.afterOperator = true
		l.readChar()
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = l.newToken(token.PIPE_PIPE_EQUAL, "||=")
			} else {
				tok = l.newToken(token.PIPE_PIPE, "||")
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.PIPE_EQUAL, "|=")
		} else {
			tok = l.newToken(token.PIPE, "|")
		}
		l.afterOperator = true
		l.readChar()
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.CARET_EQUAL, "^=")
		} else {
			tok = l.newToken(token.CARET, "^")
		}
		l.afterOperator = true
		l.readChar()
	case '~':
		tok = l.newToken(token.TILDE, "~")
		l.afterOperator = true
		l.readChar()
	case '<':
		tok = l.lexLessThan()
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.GREATER_EQUAL, ">=")
		} else if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = l.newToken(token.GREATER_GREATER_EQUAL, ">>=")
			} else {
				tok = l.newToken(token.GREATER_GREATER, ">>")
			}
		} else {
			tok = l.newToken(token.GREATER, ">")
		}
		l.afterOperator = true
		l.readChar()
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = l.newToken(token.EQUAL_EQUAL_EQUAL, "===")
			} else {
				tok = l.newToken(token.EQUAL_EQUAL, "==")
			}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = l.newToken(token.EQUAL_TILDE, "=~")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.EQUAL_GREATER, "=>")
		} else if l.startOfLine && l.peekChar() == 'b' {
			// Check for =begin
			if l.readPosition+4 <= len(l.input) && l.input[l.readPosition:l.readPosition+5] == "begin" {
				return l.lexEmbeddedDoc()
			}
			tok = l.newToken(token.EQUAL, "=")
		} else {
			tok = l.newToken(token.EQUAL, "=")
		}
		l.afterOperator = true
		l.afterIdent = false // Reset to allow regex after =, =~, ==, etc.
		l.readChar()
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.BANG_EQUAL, "!=")
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = l.newToken(token.BANG_TILDE, "!~")
		} else {
			tok = l.newToken(token.BANG, "!")
		}
		l.afterOperator = true
		l.afterIdent = false // Reset to allow regex after !, !=, !~
		l.readChar()
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = l.newToken(token.DOT_DOT_DOT, "...")
			} else {
				tok = l.newToken(token.DOT_DOT, "..")
			}
			l.afterOperator = true
		} else {
			tok = l.newToken(token.DOT, ".")
			l.afterDot = true
		}
		l.readChar()
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok = l.newToken(token.COLON_COLON, "::")
			l.afterOperator = true
		} else if l.peekChar() == '"' {
			// Symbol with double-quoted string
			l.readChar()
			tok = l.newToken(token.SYMBOL_BEGIN, ":\"")
			l.pushStringState(modeSymbolDoubleQuote, '"', 0, true)
		} else if l.peekChar() == '\'' {
			// Symbol with single-quoted string
			l.readChar()
			tok = l.newToken(token.SYMBOL_BEGIN, ":'")
			l.pushStringState(modeSymbolSingleQuote, '\'', 0, false)
		} else if isLetter(l.peekChar()) || l.peekChar() == '_' {
			tok = l.newToken(token.SYMBOL_BEGIN, ":")
		} else {
			tok = l.newToken(token.COLON, ":")
		}
		l.readChar()
	case ',':
		tok = l.newToken(token.COMMA, ",")
		l.afterOperator = true
		l.afterIdent = false // Reset to allow regex after ,
		l.readChar()
	case ';':
		tok = l.newToken(token.SEMICOLON, ";")
		l.afterOperator = true
		l.startOfLine = true
		l.readChar()
	case '(':
		tok = l.newToken(token.LPAREN, "(")
		l.afterOperator = true
		l.afterIdent = false // Reset to allow regex after (
		l.readChar()
	case ')':
		tok = l.newToken(token.RPAREN, ")")
		l.afterRightParen = true
		l.afterOperator = false
		l.readChar()
	case '[':
		tok = l.newToken(token.LBRACKET, "[")
		l.afterOperator = true
		l.afterIdent = false // Reset to allow regex after [
		l.readChar()
	case ']':
		tok = l.newToken(token.RBRACKET, "]")
		l.afterRightBracket = true
		l.afterOperator = false
		l.readChar()
	case '{':
		// Track brace depth for interpolation
		if l.braceDepth > 0 {
			l.braceDepth++
		}
		tok = l.newToken(token.LBRACE, "{")
		l.inLabelContext = true
		l.afterOperator = true
		l.readChar()
	case '}':
		// Check if we're closing an interpolation
		if l.braceDepth > 0 {
			l.braceDepth--
			if l.braceDepth == 0 && len(l.stringStack) > 0 {
				tok = l.newToken(token.EMBEXPR_END, "}")
				l.readChar()
				// Pop the interpolation state and restore string state
				l.currentState = &l.stringStack[len(l.stringStack)-1]
				return l.setTokenPosition(tok, startLine, startColumn, startOffset)
			}
		}
		tok = l.newToken(token.RBRACE, "}")
		l.afterOperator = false
		l.inLabelContext = false
		l.readChar()
	case '\'':
		tok = l.newToken(token.STRING_BEGIN, "'")
		l.pushStringStateWithBraceDepth(modeSingleQuote, '\'', 0, false)
		l.readChar()
	case '"':
		tok = l.newToken(token.STRING_BEGIN, "\"")
		l.pushStringStateWithBraceDepth(modeDoubleQuote, '"', 0, true)
		l.readChar()
	case '`':
		tok = l.newToken(token.XSTRING_BEGIN, "`")
		l.pushStringStateWithBraceDepth(modeBacktick, '`', 0, true)
		l.readChar()
	case '#':
		tok = l.lexComment()
	case '?':
		if l.shouldLexCharLiteral() {
			tok = l.lexCharLiteral()
		} else {
			tok = l.newToken(token.QUESTION, "?")
			l.afterOperator = true
			l.readChar()
		}
	case '\\':
		if l.peekChar() == '\n' {
			// Line continuation
			l.readChar() // consume backslash
			l.readChar() // consume newline
			return l.NextToken()
		}
		tok = l.newToken(token.BACKSLASH, "\\")
		l.readChar()
	case '@':
		tok = l.lexInstanceOrClassVariable()
		// If we were in variable interpolation mode, restore string state
		if len(l.stringStack) > 0 && l.currentState == nil {
			l.currentState = &l.stringStack[len(l.stringStack)-1]
		}
	case '$':
		tok = l.lexGlobalVariable()
		// If we were in variable interpolation mode, restore string state
		if len(l.stringStack) > 0 && l.currentState == nil {
			l.currentState = &l.stringStack[len(l.stringStack)-1]
		}
	default:
		if isLetter(l.ch) || l.ch == '_' {
			tok = l.lexIdentifier()
			return l.setTokenPosition(tok, startLine, startColumn, startOffset)
		} else if isDigit(l.ch) {
			tok = l.lexNumber()
			return l.setTokenPosition(tok, startLine, startColumn, startOffset)
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch))
			l.readChar()
		}
	}

	// Reset startOfLine for all tokens except newline (handled above)
	if tok.Type != token.NEWLINE && tok.Type != token.EOF {
		l.startOfLine = false
	}
	return l.setTokenPosition(tok, startLine, startColumn, startOffset)
}

func (l *Lexer) setTokenPosition(tok token.Token, line, column, offset int) token.Token {
	tok.Line = line
	tok.Column = column
	tok.Offset = offset
	return tok
}

func (l *Lexer) newToken(tokenType token.Type, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) lexIdentifier() token.Token {
	startPos := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	literal := l.input[startPos:l.position]

	// Check for method names ending in ? or !
	if l.ch == '?' || l.ch == '!' {
		literal += string(l.ch)
		l.readChar()

		// Check for defined? specifically
		if literal == "defined?" {
			l.afterKeyword = true
			return l.newToken(token.KEYWORD_DEFINED, literal)
		}

		l.afterIdent = true
		return l.newToken(token.METHOD_NAME, literal)
	}

	// Check for setter methods (name=)
	if l.ch == '=' && l.peekChar() != '=' && l.peekChar() != '~' && l.peekChar() != '>' {
		// Only treat as setter if not followed by another = or ~ or >
		// Check if we're in a context where = is an operator
		if !l.afterDot {
			// Check for label
			if l.inLabelContext && l.peekChar() != '=' {
				// This might be a label, not a setter
			} else if l.peekChar() != '=' {
				literal += "="
				l.readChar()
				l.afterIdent = true
				return l.newToken(token.METHOD_NAME, literal)
			}
		}
	}

	// Check for labels (identifier followed by : but not ::)
	// Labels are allowed at start of line, after operators, after {, after (, after [, after ,
	if l.ch == ':' && l.peekChar() != ':' {
		// Check if this could be a label context
		if l.inLabelContext || l.afterOperator || l.startOfLine {
			literal += ":"
			l.readChar()
			l.afterIdent = false
			return l.newToken(token.LABEL, literal)
		}
	}

	// Check if it's a keyword
	tokType := token.LookupIdent(literal)

	// Handle special __END__ case
	if literal == "__END__" && l.startOfLine {
		l.afterKeyword = false
		l.afterIdent = false
		// Skip all remaining content
		l.position = len(l.input)
		l.readPosition = len(l.input)
		l.ch = 0
		return l.newToken(token.END_MARKER, literal)
	}

	if tokType.IsKeyword() {
		l.afterKeyword = true
		l.afterIdent = false
	} else {
		l.afterKeyword = false
		l.afterIdent = true
	}

	l.startOfLine = false
	return l.newToken(tokType, literal)
}

func (l *Lexer) lexNumber() token.Token {
	startPos := l.position
	isFloat := false
	isRational := false
	isImaginary := false

	if l.ch == '0' {
		l.readChar()
		switch l.ch {
		case 'x', 'X':
			// Hexadecimal
			l.readChar()
			for isHexDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			l.afterIdent = true
			l.startOfLine = false
			return l.newToken(token.INTEGER, l.input[startPos:l.position])
		case 'o', 'O':
			// Octal
			l.readChar()
			for isOctalDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			l.afterIdent = true
			l.startOfLine = false
			return l.newToken(token.INTEGER, l.input[startPos:l.position])
		case 'b', 'B':
			// Binary
			l.readChar()
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				l.readChar()
			}
			l.afterIdent = true
			l.startOfLine = false
			return l.newToken(token.INTEGER, l.input[startPos:l.position])
		case 'd', 'D':
			// Explicit decimal
			l.readChar()
			for isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			l.afterIdent = true
			l.startOfLine = false
			return l.newToken(token.INTEGER, l.input[startPos:l.position])
		case '.':
			if isDigit(l.peekChar()) {
				isFloat = true
				l.readChar()
				for isDigit(l.ch) || l.ch == '_' {
					l.readChar()
				}
			}
		case 'e', 'E':
			isFloat = true
			l.readChar()
			if l.ch == '+' || l.ch == '-' {
				l.readChar()
			}
			for isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
		}
	}

	// Read integer part
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Check for rational suffix
	if l.ch == 'r' {
		isRational = true
		l.readChar()
	}

	// Check for imaginary suffix
	if l.ch == 'i' {
		isImaginary = true
		l.readChar()
	}

	literal := l.input[startPos:l.position]

	// After a number, / should be division, not regexp
	l.afterIdent = true
	l.startOfLine = false

	if isImaginary {
		return l.newToken(token.IMAGINARY, literal)
	}
	if isRational {
		return l.newToken(token.RATIONAL, literal)
	}
	if isFloat {
		return l.newToken(token.FLOAT, literal)
	}
	return l.newToken(token.INTEGER, literal)
}

func (l *Lexer) lexInstanceOrClassVariable() token.Token {
	startPos := l.position
	l.readChar() // consume first @

	if l.ch == '@' {
		// Class variable
		l.readChar()
		for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
		l.afterIdent = true
		return l.newToken(token.CVAR, l.input[startPos:l.position])
	}

	// Instance variable
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	l.afterIdent = true
	return l.newToken(token.IVAR, l.input[startPos:l.position])
}

func (l *Lexer) lexGlobalVariable() token.Token {
	startPos := l.position
	l.readChar() // consume $

	// Check for nth reference ($1, $2, etc.)
	if isDigit(l.ch) && l.ch != '0' {
		for isDigit(l.ch) {
			l.readChar()
		}
		return l.newToken(token.NTH_REF, l.input[startPos:l.position])
	}

	// Check for back reference ($&, $`, $', $+)
	if l.ch == '&' || l.ch == '`' || l.ch == '\'' || l.ch == '+' {
		l.readChar()
		return l.newToken(token.BACK_REF, l.input[startPos:l.position])
	}

	// Check for special global variables with dash ($-w, etc.)
	if l.ch == '-' && isLetter(l.peekChar()) {
		l.readChar()
		l.readChar()
		return l.newToken(token.GVAR, l.input[startPos:l.position])
	}

	// Check for punctuation globals ($:, $;, $/, etc.)
	if isPunctuation(l.ch) {
		l.readChar()
		return l.newToken(token.GVAR, l.input[startPos:l.position])
	}

	// Regular global variable
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	l.afterIdent = true
	return l.newToken(token.GVAR, l.input[startPos:l.position])
}

func (l *Lexer) lexComment() token.Token {
	startPos := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.newToken(token.COMMENT, l.input[startPos:l.position])
}

func (l *Lexer) lexEmbeddedDoc() token.Token {
	// We're at '=' and have verified "=begin" follows
	startPos := l.position
	// Read "=begin"
	for i := 0; i < 6; i++ {
		l.readChar()
	}
	// Skip to end of line
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.readChar()
	}

	// Store the =begin token position
	tok := l.newToken(token.EMBDOC_BEGIN, "=begin")
	tok.Line = l.line - 1
	tok.Column = 1
	tok.Offset = startPos

	// Push a special state to track we're in embdoc
	l.pushStringState(modeNone, 0, 0, false)
	l.currentState.mode = modeNone // Mark as embdoc mode

	return tok
}

func (l *Lexer) lexCharLiteral() token.Token {
	startPos := l.position
	l.readChar() // consume ?

	if l.ch == '\\' {
		l.readChar() // consume \
		l.readChar() // consume escaped char
	} else {
		l.readChar() // consume char
	}

	return l.newToken(token.CHAR, l.input[startPos:l.position])
}

func (l *Lexer) shouldLexCharLiteral() bool {
	// ? followed by a character or escape sequence
	if l.peekChar() == 0 {
		return false
	}
	// Check if we're in a context where ? would be the ternary operator
	if l.afterIdent || l.afterRightParen || l.afterRightBracket {
		return false
	}
	return true
}

func (l *Lexer) lexLessThan() token.Token {
	if l.peekChar() == '=' {
		l.readChar()
		if l.peekChar() == '>' {
			l.readChar()
			l.afterOperator = true
			l.readChar()
			return l.newToken(token.LESS_EQUAL_GREATER, "<=>")
		}
		l.afterOperator = true
		l.readChar()
		return l.newToken(token.LESS_EQUAL, "<=")
	} else if l.peekChar() == '<' {
		l.readChar()
		// Check for heredoc
		if l.isHeredocStart() {
			return l.lexHeredocBegin()
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.afterOperator = true
			l.readChar()
			return l.newToken(token.LESS_LESS_EQUAL, "<<=")
		}
		l.afterOperator = true
		l.readChar()
		return l.newToken(token.LESS_LESS, "<<")
	}
	l.afterOperator = true
	l.readChar()
	return l.newToken(token.LESS, "<")
}

func (l *Lexer) isHeredocStart() bool {
	pos := l.readPosition
	// Skip optional - or ~
	if pos < len(l.input) && (l.input[pos] == '-' || l.input[pos] == '~') {
		pos++
	}
	// Check for identifier, quoted identifier, or backtick
	if pos < len(l.input) {
		ch := l.input[pos]
		return isLetter(ch) || ch == '_' || ch == '"' || ch == '\'' || ch == '`'
	}
	return false
}

func (l *Lexer) lexHeredocBegin() token.Token {
	startPos := l.position - 1 // Include the first <
	indented := false
	squiggle := false
	quoted := false
	var quoteChar byte

	// Check for - or ~
	if l.peekChar() == '-' {
		indented = true
		l.readChar()
	} else if l.peekChar() == '~' {
		squiggle = true
		indented = true
		l.readChar()
	}

	// Check for quoted identifier
	if l.peekChar() == '"' || l.peekChar() == '\'' || l.peekChar() == '`' {
		quoted = true
		quoteChar = l.peekChar()
		l.readChar()
	}

	// Read identifier
	l.readChar()
	identStart := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	ident := l.input[identStart:l.position]

	// Read closing quote if present
	if quoted && l.ch == quoteChar {
		l.readChar()
	}

	literal := l.input[startPos:l.position]

	// Consume the newline that follows the heredoc declaration
	if l.ch == '\n' {
		l.readChar()
	}

	// Push heredoc state for immediate processing
	state := stringState{
		mode:            modeHeredoc,
		heredocIdent:    ident,
		heredocIndented: indented,
		heredocSquiggle: squiggle,
		heredocQuoted:   quoted,
		interpolating:   !quoted || quoteChar != '\'',
	}
	l.stringStack = append(l.stringStack, state)
	l.currentState = &l.stringStack[len(l.stringStack)-1]

	return l.newToken(token.HEREDOC_BEGIN, literal)
}

func (l *Lexer) lexHeredocContent() token.Token {
	state := l.currentState
	ident := state.heredocIdent

	// If we've already read the content, this is the terminator call
	if state.terminator == 1 { // Using terminator as a flag
		l.popStringState()
		return l.newToken(token.HEREDOC_END, ident)
	}

	// Read heredoc content
	var content strings.Builder

	for {
		lineStart := l.position

		// Read the line
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}

		line := l.input[lineStart:l.position]
		trimmedLine := strings.TrimLeft(line, " \t")

		// Check if this is the terminator
		if (state.heredocIndented && trimmedLine == ident) || (!state.heredocIndented && line == ident) {
			// Found terminator - set flag and return content
			state.terminator = 1 // Flag that content has been read
			return l.newToken(token.STRING_CONTENT, content.String())
		}

		content.WriteString(line)
		if l.ch == '\n' {
			content.WriteByte('\n')
			l.readChar()
		}

		if l.ch == 0 {
			break
		}
	}

	l.popStringState()
	return l.newToken(token.STRING_CONTENT, content.String())
}

func (l *Lexer) lexHeredocBody() token.Token {
	if len(l.heredocQueue) == 0 {
		return l.NextToken()
	}

	state := l.heredocQueue[0]
	l.heredocQueue = l.heredocQueue[1:]
	l.sawNewline = false

	// Read heredoc content
	var content strings.Builder
	ident := state.heredocIdent

	for {
		lineStart := l.position

		// Read the line
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}

		line := l.input[lineStart:l.position]
		trimmedLine := strings.TrimLeft(line, " \t")

		// Check if this is the terminator
		if (state.heredocIndented && trimmedLine == ident) || (!state.heredocIndented && line == ident) {
			// Found terminator
			contentTok := l.newToken(token.STRING_CONTENT, content.String())

			// Store the terminator for next call
			l.pushStringState(modeHeredoc, 0, 0, false)
			l.currentState.heredocIdent = ident

			return contentTok
		}

		content.WriteString(line)
		if l.ch == '\n' {
			content.WriteByte('\n')
			l.readChar()
		}

		if l.ch == 0 {
			break
		}
	}

	return l.newToken(token.STRING_CONTENT, content.String())
}

func (l *Lexer) shouldLexRegexp() bool {
	// Regexp is expected after specific operators that start expressions, keywords, or at start of expression
	// NOT after binary operators like *, +, -, etc. where / would be division
	// We use afterKeyword for keywords that expect expressions (if, while, etc.)
	// We use startOfLine for beginning of statements
	// afterOperator is set for operators that could precede a regexp like = ( [ { ,
	// but binary arithmetic operators should set afterIdent to prevent this
	if l.afterIdent || l.afterRightParen || l.afterRightBracket {
		return false
	}
	return l.afterOperator || l.afterKeyword || l.startOfLine || l.sawNewline
}

func (l *Lexer) lexRegexp() token.Token {
	tok := l.newToken(token.REGEXP_BEGIN, "/")
	l.pushStringState(modeRegexp, '/', 0, true)
	l.readChar()
	return tok
}

func (l *Lexer) isPercentLiteral() bool {
	next := l.peekChar()
	// %q, %Q, %w, %W, %i, %I, %r, %s, %x followed by delimiter
	// or % followed directly by delimiter
	if next == 'q' || next == 'Q' || next == 'w' || next == 'W' ||
		next == 'i' || next == 'I' || next == 'r' || next == 's' || next == 'x' {
		return isPercentDelimiter(l.peekCharN(2))
	}
	return isPercentDelimiter(next)
}

func (l *Lexer) lexPercentLiteral() token.Token {
	startPos := l.position
	l.readChar() // consume %

	var mode stringMode
	var tokenType token.Type
	interpolating := true

	switch l.ch {
	case 'q':
		mode = modePercentQ
		tokenType = token.STRING_BEGIN
		interpolating = false
		l.readChar()
	case 'Q':
		mode = modePercentQUpper
		tokenType = token.STRING_BEGIN
		l.readChar()
	case 'w':
		mode = modePercentW
		tokenType = token.WORDS_BEGIN
		interpolating = false
		l.readChar()
	case 'W':
		mode = modePercentWUpper
		tokenType = token.WORDS_BEGIN
		l.readChar()
	case 'i':
		mode = modePercentI
		tokenType = token.SYMBOLS_BEGIN
		interpolating = false
		l.readChar()
	case 'I':
		mode = modePercentIUpper
		tokenType = token.SYMBOLS_BEGIN
		l.readChar()
	case 'r':
		mode = modePercentR
		tokenType = token.REGEXP_BEGIN
		l.readChar()
	case 's':
		mode = modePercentS
		tokenType = token.SYMBOL_BEGIN
		interpolating = false
		l.readChar()
	case 'x':
		mode = modePercentX
		tokenType = token.XSTRING_BEGIN
		l.readChar()
	default:
		mode = modePercentDefault
		tokenType = token.STRING_BEGIN
	}

	// Read delimiter
	openDelim := l.ch
	closeDelim := matchingDelimiter(openDelim)
	l.readChar()

	literal := l.input[startPos:l.position]
	l.pushStringState(mode, closeDelim, openDelim, interpolating)

	return l.newToken(tokenType, literal)
}

func (l *Lexer) pushStringState(mode stringMode, terminator byte, openDelim byte, interpolating bool) {
	state := stringState{
		mode:          mode,
		terminator:    terminator,
		openDelimiter: openDelim,
		nestingLevel:  1,
		interpolating: interpolating,
	}
	l.stringStack = append(l.stringStack, state)
	l.currentState = &l.stringStack[len(l.stringStack)-1]
}

func (l *Lexer) pushStringStateWithBraceDepth(mode stringMode, terminator byte, openDelim byte, interpolating bool) {
	state := stringState{
		mode:            mode,
		terminator:      terminator,
		openDelimiter:   openDelim,
		nestingLevel:    1,
		interpolating:   interpolating,
		savedBraceDepth: l.braceDepth, // Save current brace depth
	}
	l.braceDepth = 0 // Reset for the new string context
	l.stringStack = append(l.stringStack, state)
	l.currentState = &l.stringStack[len(l.stringStack)-1]
}

func (l *Lexer) popStringState() {
	if len(l.stringStack) > 0 {
		// Restore brace depth from the popped state
		l.braceDepth = l.stringStack[len(l.stringStack)-1].savedBraceDepth
		l.stringStack = l.stringStack[:len(l.stringStack)-1]
	}
	// If we're still inside an interpolation (braceDepth > 0), don't set currentState
	// We need to continue normal tokenization to find the closing }
	if l.braceDepth > 0 {
		l.currentState = nil
	} else if len(l.stringStack) > 0 {
		l.currentState = &l.stringStack[len(l.stringStack)-1]
	} else {
		l.currentState = nil
	}
}

func (l *Lexer) lexStringContent() token.Token {
	if l.currentState == nil {
		return l.NextToken()
	}

	// Handle heredoc mode
	if l.currentState.mode == modeHeredoc {
		return l.lexHeredocContent()
	}

	// Handle embdoc
	if l.currentState.mode == modeNone {
		return l.lexEmbdocContent()
	}

	state := l.currentState
	mode := state.mode

	// Handle word arrays differently
	if mode == modePercentW || mode == modePercentWUpper ||
		mode == modePercentI || mode == modePercentIUpper {
		return l.lexWordArrayContent()
	}

	var content strings.Builder
	startLine := l.line
	startColumn := l.column
	startOffset := l.position

	for {
		if l.ch == 0 {
			break
		}

		// Check for terminator
		if l.ch == state.terminator {
			if state.nestingLevel == 1 || state.openDelimiter == 0 {
				// End of string
				if content.Len() > 0 {
					tok := l.newToken(token.STRING_CONTENT, content.String())
					return l.setTokenPosition(tok, startLine, startColumn, startOffset)
				}

				// Return end token
				var endTok token.Token
				if mode == modeRegexp || mode == modePercentR {
					// Read regexp flags
					l.readChar()
					flags := ""
					for l.ch == 'i' || l.ch == 'm' || l.ch == 'x' || l.ch == 'o' ||
						l.ch == 'e' || l.ch == 's' || l.ch == 'u' || l.ch == 'n' {
						flags += string(l.ch)
						l.readChar()
					}
					endTok = l.newToken(token.REGEXP_END, "/"+flags)
					if mode == modePercentR {
						endTok.Literal = string(state.terminator) + flags
					}
				} else {
					endTok = l.newToken(token.STRING_END, string(state.terminator))
					l.readChar()
				}
				l.popStringState()
				return endTok
			}
			state.nestingLevel--
			content.WriteByte(l.ch)
			l.readChar()
			continue
		}

		// Check for nested delimiter
		if state.openDelimiter != 0 && l.ch == state.openDelimiter {
			state.nestingLevel++
			content.WriteByte(l.ch)
			l.readChar()
			continue
		}

		// Check for escape sequences
		if l.ch == '\\' {
			content.WriteByte(l.ch)
			l.readChar()
			if l.ch != 0 {
				content.WriteByte(l.ch)
				l.readChar()
			}
			continue
		}

		// Check for interpolation in interpolating strings
		if state.interpolating && l.ch == '#' {
			next := l.peekChar()
			if next == '{' {
				// Expression interpolation
				if content.Len() > 0 {
					tok := l.newToken(token.STRING_CONTENT, content.String())
					return l.setTokenPosition(tok, startLine, startColumn, startOffset)
				}
				l.readChar() // consume #
				l.readChar() // consume {
				l.braceDepth = 1
				// Don't pop string state, just pause it
				l.currentState = nil
				return l.newToken(token.EMBEXPR_BEGIN, "#{")
			} else if next == '@' || next == '$' {
				// Variable interpolation
				if content.Len() > 0 {
					tok := l.newToken(token.STRING_CONTENT, content.String())
					return l.setTokenPosition(tok, startLine, startColumn, startOffset)
				}
				l.readChar() // consume #
				l.currentState = nil
				return l.newToken(token.EMBVAR, "#")
			}
		}

		content.WriteByte(l.ch)
		l.readChar()
	}

	if content.Len() > 0 {
		tok := l.newToken(token.STRING_CONTENT, content.String())
		return l.setTokenPosition(tok, startLine, startColumn, startOffset)
	}

	l.popStringState()
	return l.newToken(token.EOF, "")
}

func (l *Lexer) lexWordArrayContent() token.Token {
	state := l.currentState

	// Check for terminator first
	if l.ch == state.terminator {
		if state.nestingLevel == 1 || state.openDelimiter == 0 {
			l.readChar()
			l.popStringState()
			return l.newToken(token.STRING_END, string(state.terminator))
		}
		state.nestingLevel--
	}

	// Skip leading whitespace and return separator if we're between words
	if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		// Skip all whitespace
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			if l.ch == state.terminator {
				break
			}
			l.readChar()
		}
		// Check for terminator after whitespace
		if l.ch == state.terminator {
			if state.nestingLevel == 1 || state.openDelimiter == 0 {
				l.readChar()
				l.popStringState()
				return l.newToken(token.STRING_END, string(state.terminator))
			}
		}
		// Return separator if more content follows
		if l.ch != 0 && l.ch != state.terminator {
			return l.newToken(token.WORDS_SEP, " ")
		}
	}

	// Check for nested delimiter
	if state.openDelimiter != 0 && l.ch == state.openDelimiter {
		state.nestingLevel++
	}

	// Read word content
	var content strings.Builder
	for {
		if l.ch == 0 {
			break
		}
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			// Word separator - return the word we've built
			if content.Len() > 0 {
				return l.newToken(token.STRING_CONTENT, content.String())
			}
			break
		}
		if l.ch == state.terminator {
			if state.nestingLevel == 1 || state.openDelimiter == 0 {
				break
			}
			state.nestingLevel--
		}
		if state.openDelimiter != 0 && l.ch == state.openDelimiter {
			state.nestingLevel++
		}
		if l.ch == '\\' {
			l.readChar()
			if l.ch != 0 {
				content.WriteByte(l.ch)
				l.readChar()
			}
			continue
		}
		content.WriteByte(l.ch)
		l.readChar()
	}

	if content.Len() > 0 {
		return l.newToken(token.STRING_CONTENT, content.String())
	}

	// Check for terminator again
	if l.ch == state.terminator {
		l.readChar()
		l.popStringState()
		return l.newToken(token.STRING_END, string(state.terminator))
	}

	l.popStringState()
	return l.newToken(token.EOF, "")
}

func (l *Lexer) lexEmbdocContent() token.Token {
	startPos := l.position
	startLine := l.line

	// Check for =end
	if l.ch == '=' && l.readPosition+2 < len(l.input) && l.input[l.readPosition:l.readPosition+3] == "end" {
		for i := 0; i < 4; i++ {
			l.readChar()
		}
		// Skip to end of line
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		l.popStringState()
		tok := l.newToken(token.EMBDOC_END, "=end")
		tok.Line = startLine
		return tok
	}

	// Read line
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	line := l.input[startPos:l.position]
	if l.ch == '\n' {
		line += "\n"
		l.readChar()
	}

	return l.newToken(token.EMBDOC_LINE, line)
}

func matchingDelimiter(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	default:
		return open
	}
}

func isPercentDelimiter(ch byte) bool {
	return ch != 0 && !isLetter(ch) && !isDigit(ch) && ch != ' ' && ch != '\t' && ch != '\n'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isPunctuation(ch byte) bool {
	return ch == ':' || ch == ';' || ch == '/' || ch == '\\' || ch == '!' ||
		ch == '?' || ch == '"' || ch == '\'' || ch == '<' || ch == '>' ||
		ch == '.' || ch == ',' || ch == '=' || ch == '~' || ch == '*' ||
		ch == '$' || ch == '@' || ch == '&' || ch == '`' || ch == '+' ||
		ch == '-' || ch == '0'
}
