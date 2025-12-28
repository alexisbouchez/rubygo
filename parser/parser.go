// Package parser implements a Ruby parser using Pratt parsing.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alexisbouchez/rubylexer/ast"
	"github.com/alexisbouchez/rubylexer/lexer"
	"github.com/alexisbouchez/rubylexer/token"
)

// Precedence levels for Ruby operators
const (
	_ int = iota
	LOWEST
	MODIFIER     // if, unless, while, until (modifier form)
	RESCUE_MOD   // rescue (modifier form)
	ASSIGNMENT   // =, +=, -=, etc.
	TERNARY      // ? :
	RANGE        // .., ...
	OR           // or
	AND          // and
	NOT          // not
	EQUALS       // ==, !=, ===, <=>
	COMPARE      // <, >, <=, >=
	BITOR        // |, ^
	BITAND       // &
	SHIFT        // <<, >>
	SUM          // +, -
	PRODUCT      // *, /, %
	UNARY        // !, ~, unary +, unary -
	POWER        // **
	INDEX        // [], .(), ::
	CALL         // method call
)

// precedences maps token types to their precedence levels
var precedences = map[token.Type]int{
	// Modifier keywords (lowest)
	token.KEYWORD_IF:              MODIFIER,
	token.KEYWORD_IF_MODIFIER:     MODIFIER,
	token.KEYWORD_UNLESS:          MODIFIER,
	token.KEYWORD_UNLESS_MODIFIER: MODIFIER,
	token.KEYWORD_WHILE:           MODIFIER,
	token.KEYWORD_WHILE_MODIFIER:  MODIFIER,
	token.KEYWORD_UNTIL:           MODIFIER,
	token.KEYWORD_UNTIL_MODIFIER:  MODIFIER,
	token.KEYWORD_RESCUE:          RESCUE_MOD,
	token.KEYWORD_RESCUE_MODIFIER: RESCUE_MOD,

	// Assignment
	token.EQUAL:                     ASSIGNMENT,
	token.PLUS_EQUAL:                ASSIGNMENT,
	token.MINUS_EQUAL:               ASSIGNMENT,
	token.STAR_EQUAL:                ASSIGNMENT,
	token.SLASH_EQUAL:               ASSIGNMENT,
	token.PERCENT_EQUAL:             ASSIGNMENT,
	token.STAR_STAR_EQUAL:           ASSIGNMENT,
	token.AMPERSAND_EQUAL:           ASSIGNMENT,
	token.PIPE_EQUAL:                ASSIGNMENT,
	token.CARET_EQUAL:               ASSIGNMENT,
	token.LESS_LESS_EQUAL:           ASSIGNMENT,
	token.GREATER_GREATER_EQUAL:     ASSIGNMENT,
	token.PIPE_PIPE_EQUAL:           ASSIGNMENT,
	token.AMPERSAND_AMPERSAND_EQUAL: ASSIGNMENT,

	// Ternary
	token.QUESTION: TERNARY,

	// Range
	token.DOT_DOT:     RANGE,
	token.DOT_DOT_DOT: RANGE,

	// Logical
	token.KEYWORD_OR:          OR,
	token.KEYWORD_AND:         AND,
	token.PIPE_PIPE:           OR + 5, // || has higher precedence than 'or'
	token.AMPERSAND_AMPERSAND: AND + 5,

	// Comparison
	token.EQUAL_EQUAL:       EQUALS,
	token.BANG_EQUAL:        EQUALS,
	token.EQUAL_EQUAL_EQUAL: EQUALS,
	token.LESS_EQUAL_GREATER: EQUALS,
	token.EQUAL_TILDE:       EQUALS,
	token.BANG_TILDE:        EQUALS,

	token.LESS:          COMPARE,
	token.GREATER:       COMPARE,
	token.LESS_EQUAL:    COMPARE,
	token.GREATER_EQUAL: COMPARE,

	// Bitwise
	token.PIPE:  BITOR,
	token.CARET: BITOR,
	token.AMPERSAND: BITAND,

	// Shift
	token.LESS_LESS:      SHIFT,
	token.GREATER_GREATER: SHIFT,

	// Arithmetic
	token.PLUS:  SUM,
	token.MINUS: SUM,

	token.STAR:    PRODUCT,
	token.SLASH:   PRODUCT,
	token.PERCENT: PRODUCT,

	// Power (right associative)
	token.STAR_STAR: POWER,

	// Index and call
	token.LBRACKET:      INDEX,
	token.DOT:           CALL,
	token.AMPERSAND_DOT: CALL,
	token.COLON_COLON:   CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser holds the state of the parser
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	// sawNewline indicates that we skipped a newline while getting to peekToken
	// This is used to properly terminate statements at newlines
	sawNewline bool

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.infixParseFns = make(map[token.Type]infixParseFn)

	// Register prefix parse functions
	p.registerPrefix(token.INTEGER, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING_BEGIN, p.parseStringLiteral)
	p.registerPrefix(token.STRING_CONTENT, p.parseSimpleStringLiteral)
	p.registerPrefix(token.SYMBOL_BEGIN, p.parseSymbolLiteral)
	p.registerPrefix(token.COLON, p.parseSymbolLiteral)
	p.registerPrefix(token.KEYWORD_TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.KEYWORD_FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.KEYWORD_NIL, p.parseNilLiteral)
	p.registerPrefix(token.KEYWORD_SELF, p.parseSelfExpression)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.CONSTANT, p.parseConstant)
	p.registerPrefix(token.IVAR, p.parseInstanceVariable)
	p.registerPrefix(token.CVAR, p.parseClassVariable)
	p.registerPrefix(token.GVAR, p.parseGlobalVariable)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.TILDE, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LPAREN_BEG, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACKET_ARRAY, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.KEYWORD_IF, p.parseIfExpression)
	p.registerPrefix(token.KEYWORD_UNLESS, p.parseUnlessExpression)
	p.registerPrefix(token.KEYWORD_CASE, p.parseCaseExpression)
	p.registerPrefix(token.KEYWORD_WHILE, p.parseWhileExpression)
	p.registerPrefix(token.KEYWORD_UNTIL, p.parseUntilExpression)
	p.registerPrefix(token.KEYWORD_FOR, p.parseForExpression)
	p.registerPrefix(token.KEYWORD_BEGIN, p.parseBeginExpression)
	p.registerPrefix(token.KEYWORD_YIELD, p.parseYieldExpression)
	p.registerPrefix(token.KEYWORD_SUPER, p.parseSuperExpression)
	p.registerPrefix(token.KEYWORD_NOT, p.parseNotExpression)
	p.registerPrefix(token.KEYWORD_DEFINED, p.parseDefinedExpression)
	p.registerPrefix(token.MINUS_GREATER, p.parseLambda)
	p.registerPrefix(token.REGEXP_BEGIN, p.parseRegexpLiteral)
	p.registerPrefix(token.UCOLON_COLON, p.parseTopLevelConstant)
	p.registerPrefix(token.LABEL, p.parseLabelAsSymbol)
	p.registerPrefix(token.STAR, p.parseSplatExpression)
	p.registerPrefix(token.STAR_STAR, p.parseDoubleSplatExpression)

	// Register infix parse functions
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.STAR, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.STAR_STAR, p.parseInfixExpression)
	p.registerInfix(token.EQUAL_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.BANG_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.EQUAL_EQUAL_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.LESS_EQUAL_GREATER, p.parseInfixExpression)
	p.registerInfix(token.LESS, p.parseInfixExpression)
	p.registerInfix(token.GREATER, p.parseInfixExpression)
	p.registerInfix(token.LESS_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.GREATER_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.AMPERSAND_AMPERSAND, p.parseInfixExpression)
	p.registerInfix(token.PIPE_PIPE, p.parseInfixExpression)
	p.registerInfix(token.AMPERSAND, p.parseInfixExpression)
	p.registerInfix(token.PIPE, p.parseInfixExpression)
	p.registerInfix(token.CARET, p.parseInfixExpression)
	p.registerInfix(token.LESS_LESS, p.parseInfixExpression)
	p.registerInfix(token.GREATER_GREATER, p.parseInfixExpression)
	p.registerInfix(token.EQUAL_TILDE, p.parseInfixExpression)
	p.registerInfix(token.BANG_TILDE, p.parseInfixExpression)
	p.registerInfix(token.DOT_DOT, p.parseRangeExpression)
	p.registerInfix(token.DOT_DOT_DOT, p.parseRangeExpression)
	p.registerInfix(token.EQUAL, p.parseAssignment)
	p.registerInfix(token.PLUS_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.MINUS_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.STAR_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.SLASH_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.PERCENT_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.STAR_STAR_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.PIPE_PIPE_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.AMPERSAND_AMPERSAND_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.AMPERSAND_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.PIPE_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.CARET_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.LESS_LESS_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.GREATER_GREATER_EQUAL, p.parseOpAssignment)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMethodCall)
	p.registerInfix(token.AMPERSAND_DOT, p.parseSafeNavigation)
	p.registerInfix(token.COLON_COLON, p.parseScopedConstant)
	p.registerInfix(token.QUESTION, p.parseTernaryExpression)
	p.registerInfix(token.KEYWORD_AND, p.parseAndExpression)
	p.registerInfix(token.KEYWORD_OR, p.parseOrExpression)
	p.registerInfix(token.KEYWORD_IF, p.parseModifierIf)
	p.registerInfix(token.KEYWORD_IF_MODIFIER, p.parseModifierIf)
	p.registerInfix(token.KEYWORD_UNLESS, p.parseModifierUnless)
	p.registerInfix(token.KEYWORD_UNLESS_MODIFIER, p.parseModifierUnless)
	p.registerInfix(token.KEYWORD_WHILE, p.parseModifierWhile)
	p.registerInfix(token.KEYWORD_WHILE_MODIFIER, p.parseModifierWhile)
	p.registerInfix(token.KEYWORD_UNTIL, p.parseModifierUntil)
	p.registerInfix(token.KEYWORD_UNTIL_MODIFIER, p.parseModifierUntil)
	p.registerInfix(token.KEYWORD_RESCUE, p.parseRescueModifier)
	p.registerInfix(token.KEYWORD_RESCUE_MODIFIER, p.parseRescueModifier)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead (literal: %q)",
		t.String(), p.peekToken.Type.String(), p.peekToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found (literal: %q)",
		t.String(), p.curToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	// Skip newlines, comments and ignored newlines in most cases
	// Track if we skipped a newline so we can use it for statement separation
	p.sawNewline = false
	for p.peekToken.Type == token.NEWLINE ||
		p.peekToken.Type == token.IGNORED_NEWLINE ||
		p.peekToken.Type == token.COMMENT {
		if p.peekToken.Type == token.NEWLINE {
			p.sawNewline = true
		}
		p.peekToken = p.l.NextToken()
	}
}


func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.KEYWORD_DEF:
		return p.parseMethodDefinition()
	case token.KEYWORD_CLASS:
		return p.parseClassDefinition()
	case token.KEYWORD_MODULE:
		return p.parseModuleDefinition()
	case token.KEYWORD_RETURN:
		return p.parseReturnStatement()
	case token.KEYWORD_BREAK:
		return p.parseBreakStatement()
	case token.KEYWORD_NEXT:
		return p.parseNextStatement()
	case token.KEYWORD_REDO:
		return p.parseRedoStatement()
	case token.KEYWORD_RETRY:
		return p.parseRetryStatement()
	case token.KEYWORD_ALIAS:
		return p.parseAliasStatement()
	case token.KEYWORD_UNDEF:
		return p.parseUndefStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// Literal parsing

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// Handle different bases and underscores
	literal := strings.ReplaceAll(p.curToken.Literal, "_", "")

	var value int64
	var err error

	if strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X") {
		value, err = strconv.ParseInt(literal[2:], 16, 64)
	} else if strings.HasPrefix(literal, "0o") || strings.HasPrefix(literal, "0O") {
		value, err = strconv.ParseInt(literal[2:], 8, 64)
	} else if strings.HasPrefix(literal, "0b") || strings.HasPrefix(literal, "0B") {
		value, err = strconv.ParseInt(literal[2:], 2, 64)
	} else {
		value, err = strconv.ParseInt(literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	literal := strings.ReplaceAll(p.curToken.Literal, "_", "")
	value, err := strconv.ParseFloat(literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	startToken := p.curToken
	var parts []ast.Expression
	var currentContent strings.Builder
	hasInterpolation := false

	// Move past STRING_BEGIN
	p.nextToken()

	for !p.curTokenIs(token.STRING_END) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.STRING_CONTENT:
			currentContent.WriteString(p.curToken.Literal)
		case token.EMBEXPR_BEGIN:
			// Save current content if any
			if currentContent.Len() > 0 {
				parts = append(parts, &ast.StringLiteral{
					Token: p.curToken,
					Value: currentContent.String(),
				})
				currentContent.Reset()
			}
			hasInterpolation = true
			p.nextToken() // move past #{
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				parts = append(parts, expr)
			}
			// Expect closing }
			if !p.expectPeek(token.EMBEXPR_END) {
				// Try to continue
			}
		case token.EMBVAR:
			// Variable interpolation like #@var
			if currentContent.Len() > 0 {
				parts = append(parts, &ast.StringLiteral{
					Token: p.curToken,
					Value: currentContent.String(),
				})
				currentContent.Reset()
			}
			hasInterpolation = true
			p.nextToken()
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				parts = append(parts, expr)
			}
			continue // Don't advance, parseExpression already did
		default:
			currentContent.WriteString(p.curToken.Literal)
		}
		p.nextToken()
	}

	// Add remaining content
	if currentContent.Len() > 0 || len(parts) == 0 {
		parts = append(parts, &ast.StringLiteral{
			Token: startToken,
			Value: currentContent.String(),
		})
	}

	if hasInterpolation {
		return &ast.InterpolatedString{
			Token: startToken,
			Parts: parts,
		}
	}

	// Simple string
	return &ast.StringLiteral{
		Token: startToken,
		Value: currentContent.String(),
	}
}

func (p *Parser) parseSimpleStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseSymbolLiteral() ast.Expression {
	tok := p.curToken

	// Handle :symbol or :"string" syntax
	if p.curTokenIs(token.COLON) || p.curTokenIs(token.SYMBOL_BEGIN) {
		p.nextToken()
	}

	var value string
	switch p.curToken.Type {
	case token.IDENT, token.CONSTANT, token.METHOD_NAME:
		value = p.curToken.Literal
	case token.STRING_BEGIN:
		// :"string" syntax
		str := p.parseStringLiteral()
		if sl, ok := str.(*ast.StringLiteral); ok {
			value = sl.Value
		}
	case token.STRING_CONTENT:
		value = p.curToken.Literal
	default:
		value = p.curToken.Literal
	}

	// Consume STRING_END if present
	if p.peekTokenIs(token.STRING_END) {
		p.nextToken()
	}

	return &ast.SymbolLiteral{
		Token: tok,
		Value: value,
	}
}

func (p *Parser) parseLabelAsSymbol() ast.Expression {
	// Label token like "foo:" - treat as symbol :foo for hash keys
	value := strings.TrimSuffix(p.curToken.Literal, ":")
	return &ast.SymbolLiteral{
		Token: p.curToken,
		Value: value,
	}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(token.KEYWORD_TRUE),
	}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parseSelfExpression() ast.Expression {
	return &ast.SelfExpression{Token: p.curToken}
}

func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check if this is a method call without parentheses
	// Note: LBRACKET should NOT trigger method call - it's for indexing
	// Note: STAR should NOT trigger method call - it could be multiplication
	// Splat arguments should use explicit parentheses: method(*args)
	// IMPORTANT: If we saw a newline, the next token is on a new line,
	// so this identifier is NOT a method call - it's just an identifier
	if !p.sawNewline && (p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.INTEGER) ||
		p.peekTokenIs(token.FLOAT) || p.peekTokenIs(token.STRING_BEGIN) ||
		p.peekTokenIs(token.COLON) || p.peekTokenIs(token.SYMBOL_BEGIN) ||
		p.peekTokenIs(token.KEYWORD_TRUE) || p.peekTokenIs(token.KEYWORD_FALSE) ||
		p.peekTokenIs(token.KEYWORD_NIL) ||
		p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.IVAR) ||
		p.peekTokenIs(token.CVAR) || p.peekTokenIs(token.GVAR) ||
		p.peekTokenIs(token.CONSTANT) ||
		p.peekTokenIs(token.AMPERSAND)) {
		return p.parseMethodCallWithoutParens(ident)
	}

	// Check for method call with parentheses
	if p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LPAREN_ARG) {
		return p.parseMethodCallWithParens(ident)
	}

	// Check for block attachment
	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACE_BLOCK) ||
		p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_BLOCK) {
		return p.parseMethodCallWithBlock(ident, nil)
	}

	return ident
}

func (p *Parser) parseConstant() ast.Expression {
	return &ast.Constant{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseInstanceVariable() ast.Expression {
	return &ast.InstanceVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parseClassVariable() ast.Expression {
	return &ast.ClassVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parseGlobalVariable() ast.Expression {
	return &ast.GlobalVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parseRegexpLiteral() ast.Expression {
	tok := p.curToken
	var content strings.Builder

	p.nextToken() // move past REGEXP_BEGIN

	for !p.curTokenIs(token.REGEXP_END) && !p.curTokenIs(token.EOF) {
		content.WriteString(p.curToken.Literal)
		p.nextToken()
	}

	flags := ""
	if p.curTokenIs(token.REGEXP_END) {
		// Extract flags from the end token if present
		flags = strings.TrimPrefix(p.curToken.Literal, "/")
	}

	return &ast.RegexpLiteral{
		Token: tok,
		Value: content.String(),
		Flags: flags,
	}
}

// Prefix expressions

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(UNARY)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	hash.Order = []ast.Expression{}

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	p.nextToken()

	for {
		key := p.parseExpression(LOWEST)
		if key == nil {
			return nil
		}

		// Check for label syntax (foo: value)
		if p.curTokenIs(token.LABEL) {
			// The key is already parsed as a symbol from the label
			// Move to the value
			p.nextToken()
		} else if p.peekTokenIs(token.EQUAL_GREATER) {
			// Hash rocket syntax
			p.nextToken() // move to =>
			p.nextToken() // move to value
		} else if p.peekTokenIs(token.COLON) {
			// Symbol key with colon after (JSON-like syntax isn't standard Ruby)
			p.nextToken() // move to :
			p.nextToken() // move to value
		} else {
			// Label was already consumed in parseExpression
			// just advance to value
			p.nextToken()
		}

		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}

		hash.Pairs[key] = value
		hash.Order = append(hash.Order, key)

		if p.peekTokenIs(token.RBRACE) {
			break
		}

		if !p.expectPeek(token.COMMA) {
			// Allow missing comma before closing brace
			if p.peekTokenIs(token.RBRACE) {
				break
			}
			return nil
		}
		p.nextToken()
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()

	// Check for keyword argument pattern (LABEL token like "name:")
	if p.curTokenIs(token.LABEL) {
		// Start collecting keyword arguments as a hash
		hash := p.parseImplicitHash(end)
		list = append(list, hash)
	} else if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
		// Also handle ident: pattern
		hash := p.parseImplicitHash(end)
		list = append(list, hash)
	} else {
		list = append(list, p.parseExpression(LOWEST))

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next expression

			// Check if remaining arguments are keyword args
			if p.curTokenIs(token.LABEL) || (p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON)) {
				hash := p.parseImplicitHash(end)
				list = append(list, hash)
				return list // hash consumes rest of arguments
			}

			list = append(list, p.parseExpression(LOWEST))
		}

		if !p.expectPeek(end) {
			return nil
		}
	}

	return list
}

// parseImplicitHash parses keyword arguments as an implicit hash (without braces)
func (p *Parser) parseImplicitHash(end token.Type) *ast.HashLiteral {
	hash := &ast.HashLiteral{
		Token:         p.curToken,
		Pairs:         make(map[ast.Expression]ast.Expression),
		Order:         []ast.Expression{},
		IsKeywordArgs: true,
	}

	for !p.curTokenIs(end) && !p.curTokenIs(token.EOF) {
		var keyName string

		// Parse key - handle both LABEL ("name:") and IDENT + COLON patterns
		if p.curTokenIs(token.LABEL) {
			// LABEL is "name:" - strip the colon
			keyName = strings.TrimSuffix(p.curToken.Literal, ":")
			p.nextToken() // move to value
		} else if p.curTokenIs(token.IDENT) {
			keyName = p.curToken.Literal
			if !p.expectPeek(token.COLON) {
				return hash
			}
			p.nextToken() // move to value
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected keyword argument, got %s", p.curToken.Type))
			return hash
		}

		// Create symbol key
		key := &ast.SymbolLiteral{
			Token: token.Token{Type: token.SYMBOL_BEGIN, Literal: keyName},
			Value: keyName,
		}

		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value
		hash.Order = append(hash.Order, key)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next key
		} else {
			break
		}
	}

	if !p.expectPeek(end) {
		return nil
	}

	return hash
}

// Infix expressions

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()

	// Handle right-associative operators
	if p.curTokenIs(token.STAR_STAR) {
		precedence--
	}

	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeLiteral{
		Token:     p.curToken,
		Start:     left,
		Exclusive: p.curTokenIs(token.DOT_DOT_DOT),
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.End = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseAssignment(left ast.Expression) ast.Expression {
	expression := &ast.AssignmentExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken()
	expression.Value = p.parseExpression(ASSIGNMENT - 1)

	return expression
}

func (p *Parser) parseOpAssignment(left ast.Expression) ast.Expression {
	expression := &ast.OpAssignmentExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Value = p.parseExpression(ASSIGNMENT - 1)

	return expression
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseMethodCall(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // move past .

	methodName := p.curToken.Literal

	call := &ast.MethodCall{
		Token:    tok,
		Receiver: left,
		Method:   methodName,
	}

	// Check for arguments
	if p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LPAREN_ARG) {
		p.nextToken()
		call.Arguments = p.parseExpressionList(token.RPAREN)
	}

	// Check for block
	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACE_BLOCK) ||
		p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_BLOCK) {
		p.nextToken()
		call.Block = p.parseBlock()
	}

	return call
}

func (p *Parser) parseSafeNavigation(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // move past &.

	methodName := p.curToken.Literal

	call := &ast.MethodCall{
		Token:    tok,
		Receiver: left,
		Method:   methodName,
		SafeNav:  true,
	}

	if p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LPAREN_ARG) {
		p.nextToken()
		call.Arguments = p.parseExpressionList(token.RPAREN)
	}

	return call
}

func (p *Parser) parseScopedConstant(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // move past ::

	return &ast.ScopedConstant{
		Token: tok,
		Left:  left,
		Name:  p.curToken.Literal,
	}
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	expression := &ast.TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	p.nextToken() // move past ?
	expression.Consequence = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	expression.Alternative = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseAndExpression(left ast.Expression) ast.Expression {
	expression := &ast.AndExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseOrExpression(left ast.Expression) ast.Expression {
	expression := &ast.OrExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseModifierIf(left ast.Expression) ast.Expression {
	expression := &ast.ModifierExpression{
		Token:    p.curToken,
		Body:     left,
		Modifier: "if",
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseModifierUnless(left ast.Expression) ast.Expression {
	expression := &ast.ModifierExpression{
		Token:    p.curToken,
		Body:     left,
		Modifier: "unless",
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseModifierWhile(left ast.Expression) ast.Expression {
	expression := &ast.ModifierExpression{
		Token:    p.curToken,
		Body:     left,
		Modifier: "while",
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseModifierUntil(left ast.Expression) ast.Expression {
	expression := &ast.ModifierExpression{
		Token:    p.curToken,
		Body:     left,
		Modifier: "until",
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseRescueModifier(left ast.Expression) ast.Expression {
	expression := &ast.RescueModifier{
		Token: p.curToken,
		Body:  left,
	}

	p.nextToken()
	expression.Rescue = p.parseExpression(RESCUE_MOD)

	return expression
}

// Method call helpers

func (p *Parser) parseMethodCallWithoutParens(receiver ast.Expression) ast.Expression {
	call := &ast.MethodCall{
		Token:  receiver.(*ast.Identifier).Token,
		Method: receiver.(*ast.Identifier).Value,
	}

	// Parse arguments without parentheses
	call.Arguments = p.parseArgumentsWithoutParens()

	// Check for block
	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACE_BLOCK) ||
		p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_BLOCK) {
		p.nextToken()
		call.Block = p.parseBlock()
	}

	return call
}

func (p *Parser) parseMethodCallWithParens(receiver ast.Expression) ast.Expression {
	call := &ast.MethodCall{
		Token:  receiver.(*ast.Identifier).Token,
		Method: receiver.(*ast.Identifier).Value,
	}

	p.nextToken() // move to (
	call.Arguments = p.parseExpressionList(token.RPAREN)

	// Check for block
	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACE_BLOCK) ||
		p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_BLOCK) {
		p.nextToken()
		call.Block = p.parseBlock()
	}

	return call
}

func (p *Parser) parseMethodCallWithBlock(receiver ast.Expression, args []ast.Expression) ast.Expression {
	ident := receiver.(*ast.Identifier)
	call := &ast.MethodCall{
		Token:     ident.Token,
		Method:    ident.Value,
		Arguments: args,
	}

	p.nextToken() // move to block start
	call.Block = p.parseBlock()

	return call
}

func (p *Parser) parseArgumentsWithoutParens() []ast.Expression {
	args := []ast.Expression{}

	p.nextToken()
	// Parse first argument with higher precedence to stop at modifier keywords
	args = append(args, p.parseExpression(MODIFIER))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // move to comma
		p.nextToken() // move to next arg
		args = append(args, p.parseExpression(MODIFIER))
	}

	return args
}

// Block parsing

func (p *Parser) parseBlock() *ast.Block {
	block := &ast.Block{Token: p.curToken}

	isBrace := p.curTokenIs(token.LBRACE) || p.curTokenIs(token.LBRACE_BLOCK)

	// Parse parameters if present
	if p.peekTokenIs(token.PIPE) {
		p.nextToken() // move to |
		block.Parameters = p.parseBlockParameters()
	}

	block.Body = p.parseBlockBody(isBrace)

	return block
}

func (p *Parser) parseBlockParameters() []*ast.BlockParameter {
	params := []*ast.BlockParameter{}

	p.nextToken() // move past opening |

	for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.EOF) {
		param := &ast.BlockParameter{Token: p.curToken}

		if p.curTokenIs(token.STAR) {
			param.Splat = true
			p.nextToken()
		} else if p.curTokenIs(token.STAR_STAR) {
			param.DSplat = true
			p.nextToken()
		} else if p.curTokenIs(token.AMPERSAND) {
			param.Block = true
			p.nextToken()
		}

		param.Name = p.curToken.Literal

		// Check for default value
		if p.peekTokenIs(token.EQUAL) {
			p.nextToken() // move to =
			p.nextToken() // move to default value
			param.Default = p.parseExpression(LOWEST)
		}

		params = append(params, param)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next param
		} else {
			p.nextToken() // move to closing |
		}
	}

	return params
}

func (p *Parser) parseBlockBody(isBrace bool) *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	p.nextToken()

	endToken := token.KEYWORD_END
	if isBrace {
		endToken = token.RBRACE
	}

	for !p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	return body
}

// Control flow

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// Skip optional 'then'
	if p.peekTokenIs(token.KEYWORD_THEN) {
		p.nextToken()
	}

	expression.Consequence = p.parseConsequence()

	// Handle elsif chain
	for p.curTokenIs(token.KEYWORD_ELSIF) {
		p.nextToken()
		elsif := &ast.IfExpression{Token: p.curToken}
		elsif.Condition = p.parseExpression(LOWEST)

		if p.peekTokenIs(token.KEYWORD_THEN) {
			p.nextToken()
		}

		elsif.Consequence = p.parseConsequence()

		if expression.Alternative == nil {
			expression.Alternative = elsif
		} else {
			// Find the last elsif in the chain
			current := expression.Alternative
			for current.Alternative != nil {
				current = current.Alternative
			}
			current.Alternative = elsif
		}
	}

	// Handle else
	if p.curTokenIs(token.KEYWORD_ELSE) {
		p.nextToken()
		expression.ElseBody = p.parseBlockBodyUntilEnd()
	}

	return expression
}

func (p *Parser) parseUnlessExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken, Unless: true}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.KEYWORD_THEN) {
		p.nextToken()
	}

	expression.Consequence = p.parseConsequence()

	if p.curTokenIs(token.KEYWORD_ELSE) {
		p.nextToken()
		expression.ElseBody = p.parseBlockBodyUntilEnd()
	}

	return expression
}

func (p *Parser) parseConsequence() *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_ELSIF) &&
		!p.curTokenIs(token.KEYWORD_ELSE) &&
		!p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	return body
}

func (p *Parser) parseBlockBodyUntilEnd() *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	return body
}

func (p *Parser) parseCaseExpression() ast.Expression {
	expression := &ast.CaseExpression{Token: p.curToken}

	// Check for case with subject
	if !p.peekTokenIs(token.KEYWORD_WHEN) {
		p.nextToken()
		if !p.curTokenIs(token.KEYWORD_WHEN) {
			expression.Subject = p.parseExpression(LOWEST)
			p.nextToken()
		}
	} else {
		p.nextToken()
	}

	// Parse when clauses
	for p.curTokenIs(token.KEYWORD_WHEN) {
		when := p.parseWhenClause()
		expression.Whens = append(expression.Whens, when)
	}

	// Handle else
	if p.curTokenIs(token.KEYWORD_ELSE) {
		p.nextToken()
		expression.Else = p.parseBlockBodyUntilEnd()
	}

	return expression
}

func (p *Parser) parseWhenClause() *ast.WhenClause {
	when := &ast.WhenClause{Token: p.curToken}

	p.nextToken()

	// Parse conditions
	when.Conditions = append(when.Conditions, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		when.Conditions = append(when.Conditions, p.parseExpression(LOWEST))
	}

	// Skip optional 'then'
	if p.peekTokenIs(token.KEYWORD_THEN) {
		p.nextToken()
	}

	when.Body = &ast.BlockBody{}
	when.Body.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_WHEN) &&
		!p.curTokenIs(token.KEYWORD_ELSE) &&
		!p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			when.Body.Statements = append(when.Body.Statements, stmt)
		}
		p.nextToken()
	}

	return when
}

func (p *Parser) parseWhileExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// Skip optional 'do'
	if p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_COND) {
		p.nextToken()
	}

	expression.Body = p.parseLoopBody()

	return expression
}

func (p *Parser) parseUntilExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken, Until: true}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_COND) {
		p.nextToken()
	}

	expression.Body = p.parseLoopBody()

	return expression
}

func (p *Parser) parseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: p.curToken}

	p.nextToken()
	expression.Variable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.KEYWORD_IN) {
		return nil
	}

	p.nextToken()
	expression.Iterable = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_COND) {
		p.nextToken()
	}

	expression.Body = p.parseLoopBody()

	return expression
}

func (p *Parser) parseLoopBody() *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	return body
}

func (p *Parser) parseBeginExpression() ast.Expression {
	expression := &ast.BeginExpression{Token: p.curToken}

	expression.Body = &ast.BlockBody{}
	expression.Body.Statements = []ast.Statement{}

	p.nextToken()

	// Parse body until rescue/else/ensure/end
	for !p.curTokenIs(token.KEYWORD_RESCUE) &&
		!p.curTokenIs(token.KEYWORD_ELSE) &&
		!p.curTokenIs(token.KEYWORD_ENSURE) &&
		!p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.EOF) {
		stmt := p.parseBlockContextStatement()
		if stmt != nil {
			expression.Body.Statements = append(expression.Body.Statements, stmt)
		}
		p.nextToken()
	}

	// Parse rescue clauses
	for p.curTokenIs(token.KEYWORD_RESCUE) {
		rescue := p.parseRescueClause()
		expression.Rescues = append(expression.Rescues, rescue)
	}

	// Parse else
	if p.curTokenIs(token.KEYWORD_ELSE) {
		expression.Else = &ast.BlockBody{}
		expression.Else.Statements = []ast.Statement{}
		p.nextToken()

		for !p.curTokenIs(token.KEYWORD_ENSURE) &&
			!p.curTokenIs(token.KEYWORD_END) &&
			!p.curTokenIs(token.EOF) {
			stmt := p.parseBlockContextStatement()
			if stmt != nil {
				expression.Else.Statements = append(expression.Else.Statements, stmt)
			}
			p.nextToken()
		}
	}

	// Parse ensure
	if p.curTokenIs(token.KEYWORD_ENSURE) {
		expression.Ensure = &ast.BlockBody{}
		expression.Ensure.Statements = []ast.Statement{}
		p.nextToken()

		for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) {
			stmt := p.parseBlockContextStatement()
			if stmt != nil {
				expression.Ensure.Statements = append(expression.Ensure.Statements, stmt)
			}
			p.nextToken()
		}
	}

	return expression
}

func (p *Parser) parseRescueClause() *ast.RescueClause {
	rescue := &ast.RescueClause{Token: p.curToken}

	p.nextToken() // move past 'rescue'

	// Parse exception types if present
	// rescue StandardError => e
	// rescue SyntaxError, RuntimeError => e
	// rescue => e (no exception type)
	// rescue (bare)
	if !p.curTokenIs(token.EQUAL_GREATER) &&
		!p.curTokenIs(token.KEYWORD_THEN) &&
		!p.curTokenIs(token.KEYWORD_RESCUE) &&
		!p.curTokenIs(token.KEYWORD_ELSE) &&
		!p.curTokenIs(token.KEYWORD_ENSURE) &&
		!p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.EOF) &&
		(p.curTokenIs(token.CONSTANT) || p.curTokenIs(token.IDENT)) {

		// Parse first exception type
		rescue.Exceptions = append(rescue.Exceptions, &ast.Constant{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		// Parse additional exception types separated by comma
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next exception type
			rescue.Exceptions = append(rescue.Exceptions, &ast.Constant{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})
		}

		p.nextToken() // move past last exception type
	}

	// Parse variable binding (=> e)
	if p.curTokenIs(token.EQUAL_GREATER) {
		p.nextToken() // move past =>
		rescue.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // move past variable name
	}

	// Skip optional 'then'
	if p.curTokenIs(token.KEYWORD_THEN) {
		p.nextToken()
	}

	rescue.Body = &ast.BlockBody{}
	rescue.Body.Statements = []ast.Statement{}

	for !p.curTokenIs(token.KEYWORD_RESCUE) &&
		!p.curTokenIs(token.KEYWORD_ELSE) &&
		!p.curTokenIs(token.KEYWORD_ENSURE) &&
		!p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.EOF) {
		stmt := p.parseBlockContextStatement()
		if stmt != nil {
			rescue.Body.Statements = append(rescue.Body.Statements, stmt)
		}
		p.nextToken()
	}

	return rescue
}

// Other expressions

func (p *Parser) parseYieldExpression() ast.Expression {
	expression := &ast.YieldExpression{Token: p.curToken}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		expression.Arguments = p.parseExpressionList(token.RPAREN)
	} else if !p.peekIsStatementEnd() {
		expression.Arguments = p.parseArgumentsWithoutParens()
	}

	return expression
}

func (p *Parser) parseSuperExpression() ast.Expression {
	expression := &ast.SuperExpression{Token: p.curToken}

	if p.peekTokenIs(token.LPAREN) {
		expression.HasParens = true
		p.nextToken()
		expression.Arguments = p.parseExpressionList(token.RPAREN)
	} else if !p.peekIsStatementEnd() && !p.peekTokenIs(token.KEYWORD_DO) && !p.peekTokenIs(token.LBRACE) {
		// Check if there are arguments without parens
		if p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.INTEGER) ||
			p.peekTokenIs(token.STRING_BEGIN) || p.peekTokenIs(token.SYMBOL_BEGIN) {
			expression.Arguments = p.parseArgumentsWithoutParens()
		}
	}

	return expression
}

func (p *Parser) parseNotExpression() ast.Expression {
	expression := &ast.NotExpression{Token: p.curToken}

	p.nextToken()
	expression.Expression = p.parseExpression(NOT)

	return expression
}

func (p *Parser) parseDefinedExpression() ast.Expression {
	expression := &ast.DefinedExpression{Token: p.curToken}

	// Handle defined?(expr)
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		expression.Expression = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	} else {
		p.nextToken()
		expression.Expression = p.parseExpression(LOWEST)
	}

	return expression
}

func (p *Parser) parseLambda() ast.Expression {
	lambda := &ast.Lambda{Token: p.curToken}

	// Parse optional parameters
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		lambda.Parameters = p.parseLambdaParameters()
	}

	// Expect block { } or do...end
	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACE_BLOCK) || p.peekTokenIs(token.LAMBDA_LBRACE) {
		p.nextToken()
		lambda.Body = p.parseBlockBody(true)
	} else if p.peekTokenIs(token.KEYWORD_DO) || p.peekTokenIs(token.KEYWORD_DO_LAMBDA) {
		p.nextToken()
		lambda.Body = p.parseBlockBody(false)
	}

	return lambda
}

func (p *Parser) parseLambdaParameters() []*ast.BlockParameter {
	params := []*ast.BlockParameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		param := &ast.BlockParameter{Token: p.curToken}

		if p.curTokenIs(token.STAR) {
			param.Splat = true
			p.nextToken()
		} else if p.curTokenIs(token.STAR_STAR) {
			param.DSplat = true
			p.nextToken()
		} else if p.curTokenIs(token.AMPERSAND) {
			param.Block = true
			p.nextToken()
		}

		param.Name = p.curToken.Literal

		if p.peekTokenIs(token.EQUAL) {
			p.nextToken()
			p.nextToken()
			param.Default = p.parseExpression(LOWEST)
		}

		params = append(params, param)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			p.nextToken()
		}
	}

	return params
}

func (p *Parser) parseTopLevelConstant() ast.Expression {
	p.nextToken() // move past ::

	return &ast.ScopedConstant{
		Token: p.curToken,
		Left:  nil,
		Name:  p.curToken.Literal,
	}
}

func (p *Parser) parseSplatExpression() ast.Expression {
	expression := &ast.SplatExpression{Token: p.curToken}

	p.nextToken()
	expression.Expression = p.parseExpression(UNARY)

	return expression
}

func (p *Parser) parseDoubleSplatExpression() ast.Expression {
	expression := &ast.DoubleSplatExpression{Token: p.curToken}

	p.nextToken()
	expression.Expression = p.parseExpression(UNARY)

	return expression
}

// Statements

func (p *Parser) parseMethodDefinition() *ast.MethodDefinition {
	method := &ast.MethodDefinition{Token: p.curToken}

	p.nextToken()

	// Check for singleton method (def self.foo or def obj.foo)
	if p.peekTokenIs(token.DOT) {
		method.Receiver = p.parseExpression(LOWEST)
		p.nextToken() // move past .
		p.nextToken() // move to method name
	}

	method.Name = p.curToken.Literal

	// Parse parameters
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		method.Parameters = p.parseMethodParameters()
	} else if p.peekTokenIs(token.IDENT) {
		// Parameters without parentheses
		method.Parameters = p.parseMethodParametersWithoutParens()
	}

	method.Body = p.parseMethodBody()

	return method
}

func (p *Parser) parseMethodParameters() []*ast.MethodParameter {
	params := []*ast.MethodParameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		param := &ast.MethodParameter{Token: p.curToken}

		if p.curTokenIs(token.STAR) {
			param.Splat = true
			p.nextToken()
		} else if p.curTokenIs(token.STAR_STAR) {
			param.DSplat = true
			p.nextToken()
		} else if p.curTokenIs(token.AMPERSAND) {
			param.Block = true
			p.nextToken()
		}

		// Handle LABEL token (e.g., "name:" as a single token)
		if p.curTokenIs(token.LABEL) {
			param.KeywordOnly = true
			param.Name = strings.TrimSuffix(p.curToken.Literal, ":")
			// Check if there's a default value
			if !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RPAREN) {
				p.nextToken()
				param.Default = p.parseExpression(LOWEST)
			}
		} else {
			param.Name = p.curToken.Literal

			// Check for keyword argument syntax: name: or name: default
			if p.peekTokenIs(token.COLON) {
				param.KeywordOnly = true
				p.nextToken() // consume the colon
				// Check if there's a default value after the colon
				if !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RPAREN) {
					p.nextToken()
					param.Default = p.parseExpression(LOWEST)
				}
			} else if p.peekTokenIs(token.EQUAL) {
				p.nextToken()
				p.nextToken()
				param.Default = p.parseExpression(LOWEST)
			}
		}

		params = append(params, param)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			p.nextToken()
		}
	}

	return params
}

func (p *Parser) parseMethodParametersWithoutParens() []*ast.MethodParameter {
	params := []*ast.MethodParameter{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) &&
		!p.peekIsStatementEnd() {
		param := &ast.MethodParameter{Token: p.curToken}

		if p.curTokenIs(token.STAR) {
			param.Splat = true
			p.nextToken()
		} else if p.curTokenIs(token.STAR_STAR) {
			param.DSplat = true
			p.nextToken()
		} else if p.curTokenIs(token.AMPERSAND) {
			param.Block = true
			p.nextToken()
		}

		param.Name = p.curToken.Literal

		if p.peekTokenIs(token.EQUAL) {
			p.nextToken()
			p.nextToken()
			param.Default = p.parseExpression(LOWEST)
		}

		params = append(params, param)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}

	return params
}

func (p *Parser) parseMethodBody() *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_END) &&
		!p.curTokenIs(token.KEYWORD_RESCUE) &&
		!p.curTokenIs(token.KEYWORD_ENSURE) &&
		!p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	// Handle rescue in method body
	if p.curTokenIs(token.KEYWORD_RESCUE) {
		// Wrap in begin expression
		// For simplicity, just skip to end
		for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
	}

	return body
}

func (p *Parser) parseClassDefinition() ast.Statement {
	class := &ast.ClassDefinition{Token: p.curToken}

	p.nextToken()

	// Check for singleton class (class << obj)
	if p.curTokenIs(token.LESS_LESS) {
		return p.parseSingletonClassDefinition()
	}

	class.Name = &ast.Constant{Token: p.curToken, Value: p.curToken.Literal}

	// Check for superclass
	if p.peekTokenIs(token.LESS) {
		p.nextToken()
		p.nextToken()
		class.Superclass = p.parseExpression(LOWEST)
	}

	class.Body = p.parseClassBody()

	return class
}

func (p *Parser) parseSingletonClassDefinition() *ast.SingletonClassDefinition {
	singleton := &ast.SingletonClassDefinition{Token: p.curToken}

	p.nextToken() // move past <<
	// Parse the object (e.g., self, or any expression)
	singleton.Object = p.parseExpression(LOWEST)
	singleton.Body = p.parseClassBody()

	return singleton
}

func (p *Parser) parseClassBody() *ast.BlockBody {
	body := &ast.BlockBody{}
	body.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.KEYWORD_END) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	return body
}

func (p *Parser) parseModuleDefinition() *ast.ModuleDefinition {
	module := &ast.ModuleDefinition{Token: p.curToken}

	p.nextToken()

	module.Name = &ast.Constant{Token: p.curToken, Value: p.curToken.Literal}

	module.Body = p.parseClassBody()

	return module
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if !p.peekIsStatementEnd() {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}

	if !p.peekIsStatementEnd() {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseNextStatement() *ast.NextStatement {
	stmt := &ast.NextStatement{Token: p.curToken}

	if !p.peekIsStatementEnd() {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseRedoStatement() *ast.RedoStatement {
	return &ast.RedoStatement{Token: p.curToken}
}

func (p *Parser) parseRetryStatement() *ast.RetryStatement {
	return &ast.RetryStatement{Token: p.curToken}
}

func (p *Parser) parseAliasStatement() *ast.AliasStatement {
	stmt := &ast.AliasStatement{Token: p.curToken}

	p.nextToken()
	stmt.New = p.parseExpression(LOWEST)

	p.nextToken()
	stmt.Old = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseUndefStatement() *ast.UndefStatement {
	stmt := &ast.UndefStatement{Token: p.curToken}

	p.nextToken()
	stmt.Methods = append(stmt.Methods, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Methods = append(stmt.Methods, p.parseExpression(LOWEST))
	}

	return stmt
}

// Helper functions

func (p *Parser) peekIsStatementEnd() bool {
	// If we saw a newline while skipping to peek, the statement ends
	if p.sawNewline {
		return true
	}
	return p.peekTokenIs(token.EOF) ||
		p.peekTokenIs(token.NEWLINE) ||
		p.peekTokenIs(token.SEMICOLON) ||
		p.peekTokenIs(token.KEYWORD_END) ||
		p.peekTokenIs(token.KEYWORD_ELSE) ||
		p.peekTokenIs(token.KEYWORD_ELSIF) ||
		p.peekTokenIs(token.KEYWORD_WHEN) ||
		p.peekTokenIs(token.KEYWORD_RESCUE) ||
		p.peekTokenIs(token.KEYWORD_ENSURE) ||
		p.peekTokenIs(token.RBRACE) ||
		p.peekTokenIs(token.RPAREN) ||
		p.peekTokenIs(token.RBRACKET)
}

// parseBlockContextStatement parses a statement in a block context where
// rescue/else/ensure/end keywords should terminate the statement, not be
// parsed as modifiers.
func (p *Parser) parseBlockContextStatement() ast.Statement {
	switch p.curToken.Type {
	case token.KEYWORD_DEF:
		return p.parseMethodDefinition()
	case token.KEYWORD_CLASS:
		return p.parseClassDefinition()
	case token.KEYWORD_MODULE:
		return p.parseModuleDefinition()
	case token.KEYWORD_RETURN:
		return p.parseReturnStatement()
	case token.KEYWORD_BREAK:
		return p.parseBreakStatement()
	case token.KEYWORD_NEXT:
		return p.parseNextStatement()
	case token.KEYWORD_REDO:
		return p.parseRedoStatement()
	case token.KEYWORD_RETRY:
		return p.parseRetryStatement()
	case token.KEYWORD_ALIAS:
		return p.parseAliasStatement()
	case token.KEYWORD_UNDEF:
		return p.parseUndefStatement()
	default:
		return p.parseBlockContextExpressionStatement()
	}
}

func (p *Parser) parseBlockContextExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	// Parse expression but stop before rescue/else/ensure/end modifiers
	stmt.Expression = p.parseBlockContextExpression(LOWEST)
	return stmt
}

func (p *Parser) parseBlockContextExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) &&
		!p.peekIsBlockKeyword() &&
		precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekIsBlockKeyword() bool {
	return p.peekTokenIs(token.KEYWORD_RESCUE) ||
		p.peekTokenIs(token.KEYWORD_ELSE) ||
		p.peekTokenIs(token.KEYWORD_ENSURE) ||
		p.peekTokenIs(token.KEYWORD_END) ||
		p.peekTokenIs(token.KEYWORD_ELSIF) ||
		p.peekTokenIs(token.KEYWORD_WHEN)
}
