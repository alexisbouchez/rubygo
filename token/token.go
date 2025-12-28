// Package token defines Ruby lexer token types and utilities.
package token

// Type represents the type of a token.
type Type int

const (
	// Special tokens
	ILLEGAL Type = iota
	EOF
	NEWLINE
	IGNORED_NEWLINE // Newline that can be ignored (line continuation, etc.)
	COMMENT
	EMBDOC_BEGIN // =begin
	EMBDOC_LINE  // Lines within =begin...=end
	EMBDOC_END   // =end
	END_MARKER   // __END__

	// Identifiers and literals
	IDENT         // foo, bar
	CONSTANT      // Foo, BAR
	IVAR          // @foo
	CVAR          // @@foo
	GVAR          // $foo
	NTH_REF       // $1, $2, etc.
	BACK_REF      // $&, $`, $', $+
	LABEL         // foo:
	METHOD_NAME   // def foo?, foo!, foo=
	INTEGER       // 42, 0x2A, 0o52, 0b101010, 1_000
	FLOAT         // 3.14, 1.0e10
	RATIONAL      // 1r, 3.14r
	IMAGINARY     // 1i, 3.14i
	CHAR          // ?a, ?\n

	// Keywords
	keyword_beg
	KEYWORD___ENCODING__
	KEYWORD___FILE__
	KEYWORD___LINE__
	KEYWORD_ALIAS
	KEYWORD_AND
	KEYWORD_BEGIN
	KEYWORD_BEGIN_UPCASE // BEGIN { }
	KEYWORD_BREAK
	KEYWORD_CASE
	KEYWORD_CLASS
	KEYWORD_DEF
	KEYWORD_DEFINED
	KEYWORD_DO
	KEYWORD_DO_BLOCK  // do in block context
	KEYWORD_DO_COND   // do in condition context
	KEYWORD_DO_LAMBDA // do for lambda
	KEYWORD_ELSE
	KEYWORD_ELSIF
	KEYWORD_END
	KEYWORD_END_UPCASE // END { }
	KEYWORD_ENSURE
	KEYWORD_FALSE
	KEYWORD_FOR
	KEYWORD_IF
	KEYWORD_IF_MODIFIER
	KEYWORD_IN
	KEYWORD_MODULE
	KEYWORD_NEXT
	KEYWORD_NIL
	KEYWORD_NOT
	KEYWORD_OR
	KEYWORD_REDO
	KEYWORD_RESCUE
	KEYWORD_RESCUE_MODIFIER
	KEYWORD_RETRY
	KEYWORD_RETURN
	KEYWORD_SELF
	KEYWORD_SUPER
	KEYWORD_THEN
	KEYWORD_TRUE
	KEYWORD_UNDEF
	KEYWORD_UNLESS
	KEYWORD_UNLESS_MODIFIER
	KEYWORD_UNTIL
	KEYWORD_UNTIL_MODIFIER
	KEYWORD_WHEN
	KEYWORD_WHILE
	KEYWORD_WHILE_MODIFIER
	KEYWORD_YIELD
	keyword_end

	// Operators
	AMPERSAND                 // &
	AMPERSAND_AMPERSAND       // &&
	AMPERSAND_AMPERSAND_EQUAL // &&=
	AMPERSAND_DOT             // &.
	AMPERSAND_EQUAL           // &=
	BANG                      // !
	BANG_EQUAL                // !=
	BANG_TILDE                // !~
	CARET                     // ^
	CARET_EQUAL               // ^=
	COLON                     // :
	COLON_COLON               // ::
	COMMA                     // ,
	DOT                       // .
	DOT_DOT                   // ..
	DOT_DOT_DOT               // ...
	EQUAL                     // =
	EQUAL_EQUAL               // ==
	EQUAL_EQUAL_EQUAL         // ===
	EQUAL_GREATER             // =>
	EQUAL_TILDE               // =~
	GREATER                   // >
	GREATER_EQUAL             // >=
	GREATER_GREATER           // >>
	GREATER_GREATER_EQUAL     // >>=
	LESS                      // <
	LESS_EQUAL                // <=
	LESS_EQUAL_GREATER        // <=>
	LESS_LESS                 // <<
	LESS_LESS_EQUAL           // <<=
	MINUS                     // -
	MINUS_EQUAL               // -=
	MINUS_GREATER             // ->
	PERCENT                   // %
	PERCENT_EQUAL             // %=
	PIPE                      // |
	PIPE_EQUAL                // |=
	PIPE_PIPE                 // ||
	PIPE_PIPE_EQUAL           // ||=
	PLUS                      // +
	PLUS_EQUAL                // +=
	QUESTION                  // ?
	SEMICOLON                 // ;
	SLASH                     // /
	SLASH_EQUAL               // /=
	STAR                      // *
	STAR_EQUAL                // *=
	STAR_STAR                 // **
	STAR_STAR_EQUAL           // **=
	TILDE                     // ~
	BACKSLASH                 // \

	// Unary operators (prefix)
	UPLUS     // unary +
	UMINUS    // unary -
	USTAR     // unary * (splat)
	USTAR_STAR // unary ** (double splat)
	UAMPERSAND // unary & (block argument)
	UCOLON_COLON // :: at expression beginning

	// Brackets and delimiters
	LPAREN        // (
	LPAREN_ARG    // ( in argument context
	LPAREN_BEG    // ( at expression beginning
	RPAREN        // )
	LBRACKET      // [
	LBRACKET_ARRAY // [ for array literal
	RBRACKET      // ]
	LBRACE        // {
	LBRACE_ARG    // { in argument context
	LBRACE_BLOCK  // { for block
	RBRACE        // }
	BRACKET_LEFT_RIGHT       // []
	BRACKET_LEFT_RIGHT_EQUAL // []=

	// String-related tokens
	STRING_BEGIN   // ', ", %q, %Q, %
	STRING_CONTENT // string content
	STRING_END     // closing quote
	XSTRING_BEGIN  // ` or %x
	SYMBOL_BEGIN   // : or %s
	REGEXP_BEGIN   // / or %r
	REGEXP_END     // / with flags
	WORDS_BEGIN    // %w, %W
	WORDS_SEP      // whitespace separator in word arrays
	SYMBOLS_BEGIN  // %i, %I
	HEREDOC_BEGIN  // <<IDENT, <<-IDENT, <<~IDENT
	HEREDOC_END    // heredoc terminator

	// Interpolation
	EMBEXPR_BEGIN // #{
	EMBEXPR_END   // } closing interpolation
	EMBVAR        // #@ or #$

	// Lambda
	LAMBDA_BEGIN // ->
	LAMBDA_LBRACE // { after ->

	// Range operators in special contexts
	BDOT2 // (.. at expression beginning
	BDOT3 // (... at expression beginning

	// For error recovery
	MISSING // placeholder for missing token
)

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    Type
	Literal string
	Line    int
	Column  int
	Offset  int
}

// Position returns a human-readable position string.
func (t Token) Position() string {
	return ""
}

var tokenNames = map[Type]string{
	ILLEGAL:         "ILLEGAL",
	EOF:             "EOF",
	NEWLINE:         "NEWLINE",
	IGNORED_NEWLINE: "IGNORED_NEWLINE",
	COMMENT:         "COMMENT",
	EMBDOC_BEGIN:    "EMBDOC_BEGIN",
	EMBDOC_LINE:     "EMBDOC_LINE",
	EMBDOC_END:      "EMBDOC_END",
	END_MARKER:      "__END__",

	IDENT:       "IDENT",
	CONSTANT:    "CONSTANT",
	IVAR:        "IVAR",
	CVAR:        "CVAR",
	GVAR:        "GVAR",
	NTH_REF:     "NTH_REF",
	BACK_REF:    "BACK_REF",
	LABEL:       "LABEL",
	METHOD_NAME: "METHOD_NAME",
	INTEGER:     "INTEGER",
	FLOAT:       "FLOAT",
	RATIONAL:    "RATIONAL",
	IMAGINARY:   "IMAGINARY",
	CHAR:        "CHAR",

	KEYWORD___ENCODING__:    "__ENCODING__",
	KEYWORD___FILE__:        "__FILE__",
	KEYWORD___LINE__:        "__LINE__",
	KEYWORD_ALIAS:           "alias",
	KEYWORD_AND:             "and",
	KEYWORD_BEGIN:           "begin",
	KEYWORD_BEGIN_UPCASE:    "BEGIN",
	KEYWORD_BREAK:           "break",
	KEYWORD_CASE:            "case",
	KEYWORD_CLASS:           "class",
	KEYWORD_DEF:             "def",
	KEYWORD_DEFINED:         "defined?",
	KEYWORD_DO:              "do",
	KEYWORD_DO_BLOCK:        "do",
	KEYWORD_DO_COND:         "do",
	KEYWORD_DO_LAMBDA:       "do",
	KEYWORD_ELSE:            "else",
	KEYWORD_ELSIF:           "elsif",
	KEYWORD_END:             "end",
	KEYWORD_END_UPCASE:      "END",
	KEYWORD_ENSURE:          "ensure",
	KEYWORD_FALSE:           "false",
	KEYWORD_FOR:             "for",
	KEYWORD_IF:              "if",
	KEYWORD_IF_MODIFIER:     "if",
	KEYWORD_IN:              "in",
	KEYWORD_MODULE:          "module",
	KEYWORD_NEXT:            "next",
	KEYWORD_NIL:             "nil",
	KEYWORD_NOT:             "not",
	KEYWORD_OR:              "or",
	KEYWORD_REDO:            "redo",
	KEYWORD_RESCUE:          "rescue",
	KEYWORD_RESCUE_MODIFIER: "rescue",
	KEYWORD_RETRY:           "retry",
	KEYWORD_RETURN:          "return",
	KEYWORD_SELF:            "self",
	KEYWORD_SUPER:           "super",
	KEYWORD_THEN:            "then",
	KEYWORD_TRUE:            "true",
	KEYWORD_UNDEF:           "undef",
	KEYWORD_UNLESS:          "unless",
	KEYWORD_UNLESS_MODIFIER: "unless",
	KEYWORD_UNTIL:           "until",
	KEYWORD_UNTIL_MODIFIER:  "until",
	KEYWORD_WHEN:            "when",
	KEYWORD_WHILE:           "while",
	KEYWORD_WHILE_MODIFIER:  "while",
	KEYWORD_YIELD:           "yield",

	AMPERSAND:                 "&",
	AMPERSAND_AMPERSAND:       "&&",
	AMPERSAND_AMPERSAND_EQUAL: "&&=",
	AMPERSAND_DOT:             "&.",
	AMPERSAND_EQUAL:           "&=",
	BANG:                      "!",
	BANG_EQUAL:                "!=",
	BANG_TILDE:                "!~",
	CARET:                     "^",
	CARET_EQUAL:               "^=",
	COLON:                     ":",
	COLON_COLON:               "::",
	COMMA:                     ",",
	DOT:                       ".",
	DOT_DOT:                   "..",
	DOT_DOT_DOT:               "...",
	EQUAL:                     "=",
	EQUAL_EQUAL:               "==",
	EQUAL_EQUAL_EQUAL:         "===",
	EQUAL_GREATER:             "=>",
	EQUAL_TILDE:               "=~",
	GREATER:                   ">",
	GREATER_EQUAL:             ">=",
	GREATER_GREATER:           ">>",
	GREATER_GREATER_EQUAL:     ">>=",
	LESS:                      "<",
	LESS_EQUAL:                "<=",
	LESS_EQUAL_GREATER:        "<=>",
	LESS_LESS:                 "<<",
	LESS_LESS_EQUAL:           "<<=",
	MINUS:                     "-",
	MINUS_EQUAL:               "-=",
	MINUS_GREATER:             "->",
	PERCENT:                   "%",
	PERCENT_EQUAL:             "%=",
	PIPE:                      "|",
	PIPE_EQUAL:                "|=",
	PIPE_PIPE:                 "||",
	PIPE_PIPE_EQUAL:           "||=",
	PLUS:                      "+",
	PLUS_EQUAL:                "+=",
	QUESTION:                  "?",
	SEMICOLON:                 ";",
	SLASH:                     "/",
	SLASH_EQUAL:               "/=",
	STAR:                      "*",
	STAR_EQUAL:                "*=",
	STAR_STAR:                 "**",
	STAR_STAR_EQUAL:           "**=",
	TILDE:                     "~",
	BACKSLASH:                 "\\",

	UPLUS:        "+@",
	UMINUS:       "-@",
	USTAR:        "*",
	USTAR_STAR:   "**",
	UAMPERSAND:   "&",
	UCOLON_COLON: "::",

	LPAREN:                   "(",
	LPAREN_ARG:               "(",
	LPAREN_BEG:               "(",
	RPAREN:                   ")",
	LBRACKET:                 "[",
	LBRACKET_ARRAY:           "[",
	RBRACKET:                 "]",
	LBRACE:                   "{",
	LBRACE_ARG:               "{",
	LBRACE_BLOCK:             "{",
	RBRACE:                   "}",
	BRACKET_LEFT_RIGHT:       "[]",
	BRACKET_LEFT_RIGHT_EQUAL: "[]=",

	STRING_BEGIN:   "STRING_BEGIN",
	STRING_CONTENT: "STRING_CONTENT",
	STRING_END:     "STRING_END",
	XSTRING_BEGIN:  "XSTRING_BEGIN",
	SYMBOL_BEGIN:   "SYMBOL_BEGIN",
	REGEXP_BEGIN:   "REGEXP_BEGIN",
	REGEXP_END:     "REGEXP_END",
	WORDS_BEGIN:    "WORDS_BEGIN",
	WORDS_SEP:      "WORDS_SEP",
	SYMBOLS_BEGIN:  "SYMBOLS_BEGIN",
	HEREDOC_BEGIN:  "HEREDOC_BEGIN",
	HEREDOC_END:    "HEREDOC_END",

	EMBEXPR_BEGIN: "EMBEXPR_BEGIN",
	EMBEXPR_END:   "EMBEXPR_END",
	EMBVAR:        "EMBVAR",

	LAMBDA_BEGIN:  "LAMBDA_BEGIN",
	LAMBDA_LBRACE: "LAMBDA_LBRACE",

	BDOT2: "BDOT2",
	BDOT3: "BDOT3",

	MISSING: "MISSING",
}

// String returns the string representation of the token type.
func (t Type) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Keywords maps keyword strings to their token types.
var Keywords = map[string]Type{
	"__ENCODING__": KEYWORD___ENCODING__,
	"__FILE__":     KEYWORD___FILE__,
	"__LINE__":     KEYWORD___LINE__,
	"alias":        KEYWORD_ALIAS,
	"and":          KEYWORD_AND,
	"begin":        KEYWORD_BEGIN,
	"BEGIN":        KEYWORD_BEGIN_UPCASE,
	"break":        KEYWORD_BREAK,
	"case":         KEYWORD_CASE,
	"class":        KEYWORD_CLASS,
	"def":          KEYWORD_DEF,
	"defined?":     KEYWORD_DEFINED,
	"do":           KEYWORD_DO,
	"else":         KEYWORD_ELSE,
	"elsif":        KEYWORD_ELSIF,
	"end":          KEYWORD_END,
	"END":          KEYWORD_END_UPCASE,
	"ensure":       KEYWORD_ENSURE,
	"false":        KEYWORD_FALSE,
	"for":          KEYWORD_FOR,
	"if":           KEYWORD_IF,
	"in":           KEYWORD_IN,
	"module":       KEYWORD_MODULE,
	"next":         KEYWORD_NEXT,
	"nil":          KEYWORD_NIL,
	"not":          KEYWORD_NOT,
	"or":           KEYWORD_OR,
	"redo":         KEYWORD_REDO,
	"rescue":       KEYWORD_RESCUE,
	"retry":        KEYWORD_RETRY,
	"return":       KEYWORD_RETURN,
	"self":         KEYWORD_SELF,
	"super":        KEYWORD_SUPER,
	"then":         KEYWORD_THEN,
	"true":         KEYWORD_TRUE,
	"undef":        KEYWORD_UNDEF,
	"unless":       KEYWORD_UNLESS,
	"until":        KEYWORD_UNTIL,
	"when":         KEYWORD_WHEN,
	"while":        KEYWORD_WHILE,
	"yield":        KEYWORD_YIELD,
}

// LookupIdent returns the token type for an identifier (keyword or ident/constant).
func LookupIdent(ident string) Type {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	// Check if it's a constant (starts with uppercase)
	if len(ident) > 0 && ident[0] >= 'A' && ident[0] <= 'Z' {
		return CONSTANT
	}
	return IDENT
}

// IsKeyword returns true if the token type is a keyword.
func (t Type) IsKeyword() bool {
	return t > keyword_beg && t < keyword_end
}

// IsLiteral returns true if the token type is a literal.
func (t Type) IsLiteral() bool {
	switch t {
	case INTEGER, FLOAT, RATIONAL, IMAGINARY, CHAR, STRING_CONTENT:
		return true
	}
	return false
}

// IsOperator returns true if the token type is an operator.
func (t Type) IsOperator() bool {
	switch t {
	case AMPERSAND, AMPERSAND_AMPERSAND, AMPERSAND_AMPERSAND_EQUAL, AMPERSAND_DOT, AMPERSAND_EQUAL,
		BANG, BANG_EQUAL, BANG_TILDE, CARET, CARET_EQUAL, DOT, DOT_DOT, DOT_DOT_DOT,
		EQUAL, EQUAL_EQUAL, EQUAL_EQUAL_EQUAL, EQUAL_GREATER, EQUAL_TILDE,
		GREATER, GREATER_EQUAL, GREATER_GREATER, GREATER_GREATER_EQUAL,
		LESS, LESS_EQUAL, LESS_EQUAL_GREATER, LESS_LESS, LESS_LESS_EQUAL,
		MINUS, MINUS_EQUAL, MINUS_GREATER, PERCENT, PERCENT_EQUAL,
		PIPE, PIPE_EQUAL, PIPE_PIPE, PIPE_PIPE_EQUAL,
		PLUS, PLUS_EQUAL, SLASH, SLASH_EQUAL, STAR, STAR_EQUAL, STAR_STAR, STAR_STAR_EQUAL,
		TILDE, UPLUS, UMINUS:
		return true
	}
	return false
}
