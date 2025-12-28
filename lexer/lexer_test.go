package lexer

import (
	"testing"

	"github.com/alexisbouchez/rubylexer/token"
)

func TestNextToken_Empty(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF, got %v", tok.Type)
	}
}

func TestNextToken_Whitespace(t *testing.T) {
	l := New("   \t  ")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF after whitespace, got %v", tok.Type)
	}
}

func TestNextToken_Newlines(t *testing.T) {
	l := New("foo\nbar")
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "bar"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Identifiers(t *testing.T) {
	input := `foo bar_baz _private foo? foo! foo= foo123`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.IDENT, "bar_baz"},
		{token.IDENT, "_private"},
		{token.METHOD_NAME, "foo?"},
		{token.METHOD_NAME, "foo!"},
		{token.METHOD_NAME, "foo="},
		{token.IDENT, "foo123"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Constants(t *testing.T) {
	input := `Foo Bar_Baz FOO_BAR HTTP`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.CONSTANT, "Foo"},
		{token.CONSTANT, "Bar_Baz"},
		{token.CONSTANT, "FOO_BAR"},
		{token.CONSTANT, "HTTP"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_InstanceVariables(t *testing.T) {
	input := `@foo @bar_baz @_private`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IVAR, "@foo"},
		{token.IVAR, "@bar_baz"},
		{token.IVAR, "@_private"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_ClassVariables(t *testing.T) {
	input := `@@foo @@bar_baz`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.CVAR, "@@foo"},
		{token.CVAR, "@@bar_baz"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_GlobalVariables(t *testing.T) {
	input := `$foo $_ $0 $stdout $-w $: $; $/`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.GVAR, "$foo"},
		{token.GVAR, "$_"},
		{token.GVAR, "$0"},
		{token.GVAR, "$stdout"},
		{token.GVAR, "$-w"},
		{token.GVAR, "$:"},
		{token.GVAR, "$;"},
		{token.GVAR, "$/"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_NthRef(t *testing.T) {
	input := `$1 $2 $10`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.NTH_REF, "$1"},
		{token.NTH_REF, "$2"},
		{token.NTH_REF, "$10"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_BackRef(t *testing.T) {
	input := "$& $` $' $+"
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.BACK_REF, "$&"},
		{token.BACK_REF, "$`"},
		{token.BACK_REF, "$'"},
		{token.BACK_REF, "$+"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := `if else elsif end def class module begin rescue ensure return yield do while until for case when break next redo retry in and or not nil true false self super alias undef defined? then unless __FILE__ __LINE__ __ENCODING__`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.KEYWORD_IF, "if"},
		{token.KEYWORD_ELSE, "else"},
		{token.KEYWORD_ELSIF, "elsif"},
		{token.KEYWORD_END, "end"},
		{token.KEYWORD_DEF, "def"},
		{token.KEYWORD_CLASS, "class"},
		{token.KEYWORD_MODULE, "module"},
		{token.KEYWORD_BEGIN, "begin"},
		{token.KEYWORD_RESCUE, "rescue"},
		{token.KEYWORD_ENSURE, "ensure"},
		{token.KEYWORD_RETURN, "return"},
		{token.KEYWORD_YIELD, "yield"},
		{token.KEYWORD_DO, "do"},
		{token.KEYWORD_WHILE, "while"},
		{token.KEYWORD_UNTIL, "until"},
		{token.KEYWORD_FOR, "for"},
		{token.KEYWORD_CASE, "case"},
		{token.KEYWORD_WHEN, "when"},
		{token.KEYWORD_BREAK, "break"},
		{token.KEYWORD_NEXT, "next"},
		{token.KEYWORD_REDO, "redo"},
		{token.KEYWORD_RETRY, "retry"},
		{token.KEYWORD_IN, "in"},
		{token.KEYWORD_AND, "and"},
		{token.KEYWORD_OR, "or"},
		{token.KEYWORD_NOT, "not"},
		{token.KEYWORD_NIL, "nil"},
		{token.KEYWORD_TRUE, "true"},
		{token.KEYWORD_FALSE, "false"},
		{token.KEYWORD_SELF, "self"},
		{token.KEYWORD_SUPER, "super"},
		{token.KEYWORD_ALIAS, "alias"},
		{token.KEYWORD_UNDEF, "undef"},
		{token.KEYWORD_DEFINED, "defined?"},
		{token.KEYWORD_THEN, "then"},
		{token.KEYWORD_UNLESS, "unless"},
		{token.KEYWORD___FILE__, "__FILE__"},
		{token.KEYWORD___LINE__, "__LINE__"},
		{token.KEYWORD___ENCODING__, "__ENCODING__"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_BeginEnd(t *testing.T) {
	input := `BEGIN { } END { }`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.KEYWORD_BEGIN_UPCASE, "BEGIN"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.KEYWORD_END_UPCASE, "END"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestNextToken_Integers(t *testing.T) {
	input := `0 42 1_000_000 0x2A 0X2a 0o52 0O52 0b101010 0B101010 0d42 0D42`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INTEGER, "0"},
		{token.INTEGER, "42"},
		{token.INTEGER, "1_000_000"},
		{token.INTEGER, "0x2A"},
		{token.INTEGER, "0X2a"},
		{token.INTEGER, "0o52"},
		{token.INTEGER, "0O52"},
		{token.INTEGER, "0b101010"},
		{token.INTEGER, "0B101010"},
		{token.INTEGER, "0d42"},
		{token.INTEGER, "0D42"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Floats(t *testing.T) {
	input := `3.14 1.0 0.5 1.0e10 1.0E10 1.0e+10 1.0e-10 1e10 2.5e-3`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FLOAT, "3.14"},
		{token.FLOAT, "1.0"},
		{token.FLOAT, "0.5"},
		{token.FLOAT, "1.0e10"},
		{token.FLOAT, "1.0E10"},
		{token.FLOAT, "1.0e+10"},
		{token.FLOAT, "1.0e-10"},
		{token.FLOAT, "1e10"},
		{token.FLOAT, "2.5e-3"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Rational(t *testing.T) {
	input := `1r 42r 3.14r`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.RATIONAL, "1r"},
		{token.RATIONAL, "42r"},
		{token.RATIONAL, "3.14r"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Imaginary(t *testing.T) {
	input := `1i 42i 3.14i 1ri`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IMAGINARY, "1i"},
		{token.IMAGINARY, "42i"},
		{token.IMAGINARY, "3.14i"},
		{token.IMAGINARY, "1ri"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_CharLiteral(t *testing.T) {
	input := `?a ?b ?\n ?\t ?\\`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.CHAR, "?a"},
		{token.CHAR, "?b"},
		{token.CHAR, "?\\n"},
		{token.CHAR, "?\\t"},
		{token.CHAR, "?\\\\"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Operators(t *testing.T) {
	input := `+ - * / % ** & | ^ ~ << >> && || ! != !~ == === <=> >= <= > < =~ => -> += -= *= /= %= **= &= |= ^= <<= >>= &&= ||= .. ... &.`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.STAR, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.STAR_STAR, "**"},
		{token.AMPERSAND, "&"},
		{token.PIPE, "|"},
		{token.CARET, "^"},
		{token.TILDE, "~"},
		{token.LESS_LESS, "<<"},
		{token.GREATER_GREATER, ">>"},
		{token.AMPERSAND_AMPERSAND, "&&"},
		{token.PIPE_PIPE, "||"},
		{token.BANG, "!"},
		{token.BANG_EQUAL, "!="},
		{token.BANG_TILDE, "!~"},
		{token.EQUAL_EQUAL, "=="},
		{token.EQUAL_EQUAL_EQUAL, "==="},
		{token.LESS_EQUAL_GREATER, "<=>"},
		{token.GREATER_EQUAL, ">="},
		{token.LESS_EQUAL, "<="},
		{token.GREATER, ">"},
		{token.LESS, "<"},
		{token.EQUAL_TILDE, "=~"},
		{token.EQUAL_GREATER, "=>"},
		{token.MINUS_GREATER, "->"},
		{token.PLUS_EQUAL, "+="},
		{token.MINUS_EQUAL, "-="},
		{token.STAR_EQUAL, "*="},
		{token.SLASH_EQUAL, "/="},
		{token.PERCENT_EQUAL, "%="},
		{token.STAR_STAR_EQUAL, "**="},
		{token.AMPERSAND_EQUAL, "&="},
		{token.PIPE_EQUAL, "|="},
		{token.CARET_EQUAL, "^="},
		{token.LESS_LESS_EQUAL, "<<="},
		{token.GREATER_GREATER_EQUAL, ">>="},
		{token.AMPERSAND_AMPERSAND_EQUAL, "&&="},
		{token.PIPE_PIPE_EQUAL, "||="},
		{token.DOT_DOT, ".."},
		{token.DOT_DOT_DOT, "..."},
		{token.AMPERSAND_DOT, "&."},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Delimiters(t *testing.T) {
	input := `( ) [ ] { } , ; : :: .`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.COLON, ":"},
		{token.COLON_COLON, "::"},
		{token.DOT, "."},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `foo # this is a comment
bar`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.COMMENT, "# this is a comment"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "bar"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_EmbeddedDoc(t *testing.T) {
	input := `foo
=begin
this is a
multiline comment
=end
bar`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.NEWLINE, "\n"},
		{token.EMBDOC_BEGIN, "=begin"},
		{token.EMBDOC_LINE, "this is a\n"},
		{token.EMBDOC_LINE, "multiline comment\n"},
		{token.EMBDOC_END, "=end"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "bar"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_SingleQuoteString(t *testing.T) {
	input := `'hello' 'world'`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "'"},
		{token.STRING_CONTENT, "hello"},
		{token.STRING_END, "'"},
		{token.STRING_BEGIN, "'"},
		{token.STRING_CONTENT, "world"},
		{token.STRING_END, "'"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_DoubleQuoteString(t *testing.T) {
	input := `"hello" "world"`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "hello"},
		{token.STRING_END, "\""},
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "world"},
		{token.STRING_END, "\""},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_StringEscapes(t *testing.T) {
	input := `"hello\nworld" "tab\there" "quote\"here"`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "hello\\nworld"},
		{token.STRING_END, "\""},
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "tab\\there"},
		{token.STRING_END, "\""},
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "quote\\\"here"},
		{token.STRING_END, "\""},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_StringInterpolation(t *testing.T) {
	input := `"hello #{name}"`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "hello "},
		{token.EMBEXPR_BEGIN, "#{"},
		{token.IDENT, "name"},
		{token.EMBEXPR_END, "}"},
		{token.STRING_END, "\""},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_NestedInterpolation(t *testing.T) {
	input := `"outer #{inner + "nested #{deep}"} end"`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "outer "},
		{token.EMBEXPR_BEGIN, "#{"},
		{token.IDENT, "inner"},
		{token.PLUS, "+"},
		{token.STRING_BEGIN, "\""},
		{token.STRING_CONTENT, "nested "},
		{token.EMBEXPR_BEGIN, "#{"},
		{token.IDENT, "deep"},
		{token.EMBEXPR_END, "}"},
		{token.STRING_END, "\""},
		{token.EMBEXPR_END, "}"},
		{token.STRING_CONTENT, " end"},
		{token.STRING_END, "\""},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_VarInterpolation(t *testing.T) {
	input := `"#@foo #@@bar #$baz"`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "\""},
		{token.EMBVAR, "#"},
		{token.IVAR, "@foo"},
		{token.STRING_CONTENT, " "},
		{token.EMBVAR, "#"},
		{token.CVAR, "@@bar"},
		{token.STRING_CONTENT, " "},
		{token.EMBVAR, "#"},
		{token.GVAR, "$baz"},
		{token.STRING_END, "\""},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Symbols(t *testing.T) {
	input := `:foo :Bar :foo_bar :"with spaces" :'single quote'`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SYMBOL_BEGIN, ":"},
		{token.IDENT, "foo"},
		{token.SYMBOL_BEGIN, ":"},
		{token.CONSTANT, "Bar"},
		{token.SYMBOL_BEGIN, ":"},
		{token.IDENT, "foo_bar"},
		{token.SYMBOL_BEGIN, ":\""},
		{token.STRING_CONTENT, "with spaces"},
		{token.STRING_END, "\""},
		{token.SYMBOL_BEGIN, ":'"},
		{token.STRING_CONTENT, "single quote"},
		{token.STRING_END, "'"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Backtick(t *testing.T) {
	input := "`ls -la`"
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.XSTRING_BEGIN, "`"},
		{token.STRING_CONTENT, "ls -la"},
		{token.STRING_END, "`"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Regexp(t *testing.T) {
	input := `/foo/
/bar/i
/baz/mix`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.REGEXP_BEGIN, "/"},
		{token.STRING_CONTENT, "foo"},
		{token.REGEXP_END, "/"},
		{token.NEWLINE, "\n"},
		{token.REGEXP_BEGIN, "/"},
		{token.STRING_CONTENT, "bar"},
		{token.REGEXP_END, "/i"},
		{token.NEWLINE, "\n"},
		{token.REGEXP_BEGIN, "/"},
		{token.STRING_CONTENT, "baz"},
		{token.REGEXP_END, "/mix"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_RegexpWithEscape(t *testing.T) {
	input := `/foo\/bar/`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.REGEXP_BEGIN, "/"},
		{token.STRING_CONTENT, "foo\\/bar"},
		{token.REGEXP_END, "/"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PercentStrings(t *testing.T) {
	input := `%q(hello world) %Q(hello world) %(default) %{curly} %[bracket] %<angle>`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "%q("},
		{token.STRING_CONTENT, "hello world"},
		{token.STRING_END, ")"},
		{token.STRING_BEGIN, "%Q("},
		{token.STRING_CONTENT, "hello world"},
		{token.STRING_END, ")"},
		{token.STRING_BEGIN, "%("},
		{token.STRING_CONTENT, "default"},
		{token.STRING_END, ")"},
		{token.STRING_BEGIN, "%{"},
		{token.STRING_CONTENT, "curly"},
		{token.STRING_END, "}"},
		{token.STRING_BEGIN, "%["},
		{token.STRING_CONTENT, "bracket"},
		{token.STRING_END, "]"},
		{token.STRING_BEGIN, "%<"},
		{token.STRING_CONTENT, "angle"},
		{token.STRING_END, ">"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PercentNestedDelimiters(t *testing.T) {
	input := `%(outer (inner) end)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_BEGIN, "%("},
		{token.STRING_CONTENT, "outer (inner) end"},
		{token.STRING_END, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_WordArrays(t *testing.T) {
	input := `%w(foo bar baz) %W(hello world)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.WORDS_BEGIN, "%w("},
		{token.STRING_CONTENT, "foo"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "bar"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "baz"},
		{token.STRING_END, ")"},
		{token.WORDS_BEGIN, "%W("},
		{token.STRING_CONTENT, "hello"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "world"},
		{token.STRING_END, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_SymbolArrays(t *testing.T) {
	input := `%i(foo bar baz) %I(hello world)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SYMBOLS_BEGIN, "%i("},
		{token.STRING_CONTENT, "foo"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "bar"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "baz"},
		{token.STRING_END, ")"},
		{token.SYMBOLS_BEGIN, "%I("},
		{token.STRING_CONTENT, "hello"},
		{token.WORDS_SEP, " "},
		{token.STRING_CONTENT, "world"},
		{token.STRING_END, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PercentRegexp(t *testing.T) {
	input := `%r{foo/bar}i`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.REGEXP_BEGIN, "%r{"},
		{token.STRING_CONTENT, "foo/bar"},
		{token.REGEXP_END, "}i"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PercentSymbol(t *testing.T) {
	input := `%s(foo bar)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SYMBOL_BEGIN, "%s("},
		{token.STRING_CONTENT, "foo bar"},
		{token.STRING_END, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PercentBacktick(t *testing.T) {
	input := `%x(ls -la)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.XSTRING_BEGIN, "%x("},
		{token.STRING_CONTENT, "ls -la"},
		{token.STRING_END, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Heredoc(t *testing.T) {
	input := `<<EOF
hello
world
EOF
done`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.HEREDOC_BEGIN, "<<EOF"},
		{token.STRING_CONTENT, "hello\nworld\n"},
		{token.HEREDOC_END, "EOF"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "done"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_HeredocDash(t *testing.T) {
	input := `<<-EOF
  hello
  world
  EOF
done`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.HEREDOC_BEGIN, "<<-EOF"},
		{token.STRING_CONTENT, "  hello\n  world\n"},
		{token.HEREDOC_END, "EOF"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "done"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_HeredocSquiggle(t *testing.T) {
	input := `<<~EOF
  hello
  world
EOF
done`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.HEREDOC_BEGIN, "<<~EOF"},
		{token.STRING_CONTENT, "  hello\n  world\n"},
		{token.HEREDOC_END, "EOF"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "done"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_HeredocQuoted(t *testing.T) {
	input := `<<'EOF'
hello
world
EOF
done`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.HEREDOC_BEGIN, "<<'EOF'"},
		{token.STRING_CONTENT, "hello\nworld\n"},
		{token.HEREDOC_END, "EOF"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "done"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Label(t *testing.T) {
	input := `foo: bar:`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LABEL, "foo:"},
		{token.LABEL, "bar:"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LabelVsSymbol(t *testing.T) {
	input := `{foo: 1, :bar => 2}`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LBRACE, "{"},
		{token.LABEL, "foo:"},
		{token.INTEGER, "1"},
		{token.COMMA, ","},
		{token.SYMBOL_BEGIN, ":"},
		{token.IDENT, "bar"},
		{token.EQUAL_GREATER, "=>"},
		{token.INTEGER, "2"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Lambda(t *testing.T) {
	input := `-> { x } ->(x) { x }`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.MINUS_GREATER, "->"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.RBRACE, "}"},
		{token.MINUS_GREATER, "->"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_EndMarker(t *testing.T) {
	input := `foo
__END__
this is ignored data`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.NEWLINE, "\n"},
		{token.END_MARKER, "__END__"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LineContinuation(t *testing.T) {
	input := "foo \\\nbar"
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.IDENT, "bar"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Position(t *testing.T) {
	input := `foo
bar`
	l := New(input)

	tok := l.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Fatalf("expected foo at line 1, col 1, got line %d, col %d", tok.Line, tok.Column)
	}

	l.NextToken() // newline

	tok = l.NextToken()
	if tok.Line != 2 || tok.Column != 1 {
		t.Fatalf("expected bar at line 2, col 1, got line %d, col %d", tok.Line, tok.Column)
	}
}

func TestNextToken_MethodDefinition(t *testing.T) {
	input := `def foo(a, b = 1, *args, **kwargs, &block)
  a + b
end`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.KEYWORD_DEF, "def"},
		{token.IDENT, "foo"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.EQUAL, "="},
		{token.INTEGER, "1"},
		{token.COMMA, ","},
		{token.STAR, "*"},
		{token.IDENT, "args"},
		{token.COMMA, ","},
		{token.STAR_STAR, "**"},
		{token.IDENT, "kwargs"},
		{token.COMMA, ","},
		{token.AMPERSAND, "&"},
		{token.IDENT, "block"},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.NEWLINE, "\n"},
		{token.KEYWORD_END, "end"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_ClassDefinition(t *testing.T) {
	input := `class Foo < Bar
  include Baz

  attr_accessor :name

  def initialize(name)
    @name = name
  end
end`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.KEYWORD_CLASS, "class"},
		{token.CONSTANT, "Foo"},
		{token.LESS, "<"},
		{token.CONSTANT, "Bar"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "include"},
		{token.CONSTANT, "Baz"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "attr_accessor"},
		{token.SYMBOL_BEGIN, ":"},
		{token.IDENT, "name"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.KEYWORD_DEF, "def"},
		{token.IDENT, "initialize"},
		{token.LPAREN, "("},
		{token.IDENT, "name"},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},
		{token.IVAR, "@name"},
		{token.EQUAL, "="},
		{token.IDENT, "name"},
		{token.NEWLINE, "\n"},
		{token.KEYWORD_END, "end"},
		{token.NEWLINE, "\n"},
		{token.KEYWORD_END, "end"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_BlocksAndProcs(t *testing.T) {
	input := `[1, 2, 3].map { |x| x * 2 }
[1, 2, 3].map do |x|
  x * 2
end`
	l := New(input)
	tests := []struct {
		expectedType token.Type
	}{
		{token.LBRACKET},
		{token.INTEGER},
		{token.COMMA},
		{token.INTEGER},
		{token.COMMA},
		{token.INTEGER},
		{token.RBRACKET},
		{token.DOT},
		{token.IDENT},
		{token.LBRACE},
		{token.PIPE},
		{token.IDENT},
		{token.PIPE},
		{token.IDENT},
		{token.STAR},
		{token.INTEGER},
		{token.RBRACE},
		{token.NEWLINE},
		{token.LBRACKET},
		{token.INTEGER},
		{token.COMMA},
		{token.INTEGER},
		{token.COMMA},
		{token.INTEGER},
		{token.RBRACKET},
		{token.DOT},
		{token.IDENT},
		{token.KEYWORD_DO},
		{token.PIPE},
		{token.IDENT},
		{token.PIPE},
		{token.NEWLINE},
		{token.IDENT},
		{token.STAR},
		{token.INTEGER},
		{token.NEWLINE},
		{token.KEYWORD_END},
		{token.EOF},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestNextToken_SafeNavigation(t *testing.T) {
	input := `foo&.bar&.baz`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.AMPERSAND_DOT, "&."},
		{token.IDENT, "bar"},
		{token.AMPERSAND_DOT, "&."},
		{token.IDENT, "baz"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_TernaryOperator(t *testing.T) {
	input := `a ? b : c`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.QUESTION, "?"},
		{token.IDENT, "b"},
		{token.COLON, ":"},
		{token.IDENT, "c"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Ranges(t *testing.T) {
	input := `1..10 1...10 (..5) (...5)`
	l := New(input)
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INTEGER, "1"},
		{token.DOT_DOT, ".."},
		{token.INTEGER, "10"},
		{token.INTEGER, "1"},
		{token.DOT_DOT_DOT, "..."},
		{token.INTEGER, "10"},
		{token.LPAREN, "("},
		{token.DOT_DOT, ".."},
		{token.INTEGER, "5"},
		{token.RPAREN, ")"},
		{token.LPAREN, "("},
		{token.DOT_DOT_DOT, "..."},
		{token.INTEGER, "5"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d]: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_PatternMatching(t *testing.T) {
	input := `case x
in [a, b, *rest]
  a + b
in {name:, age:}
  name
end`
	l := New(input)
	tests := []struct {
		expectedType token.Type
	}{
		{token.KEYWORD_CASE},
		{token.IDENT},
		{token.NEWLINE},
		{token.KEYWORD_IN},
		{token.LBRACKET},
		{token.IDENT},
		{token.COMMA},
		{token.IDENT},
		{token.COMMA},
		{token.STAR},
		{token.IDENT},
		{token.RBRACKET},
		{token.NEWLINE},
		{token.IDENT},
		{token.PLUS},
		{token.IDENT},
		{token.NEWLINE},
		{token.KEYWORD_IN},
		{token.LBRACE},
		{token.LABEL},
		{token.COMMA},
		{token.LABEL},
		{token.RBRACE},
		{token.NEWLINE},
		{token.IDENT},
		{token.NEWLINE},
		{token.KEYWORD_END},
		{token.EOF},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d]: expected type %v, got %v (literal=%q)", i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}
