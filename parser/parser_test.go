package parser

import (
	"testing"

	"github.com/alexisbouchez/rubylexer/ast"
	"github.com/alexisbouchez/rubylexer/lexer"
)

func TestIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"42", 42},
		{"1_000_000", 1000000},
		{"0x2A", 42},
		{"0o52", 42},
		{"0b101010", 42},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
		}

		literal, ok := stmt.Expression.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("expected IntegerLiteral, got %T", stmt.Expression)
		}

		if literal.Value != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, literal.Value)
		}
	}
}

func TestFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"1.0", 1.0},
		{"1.5e10", 1.5e10},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		literal, ok := stmt.Expression.(*ast.FloatLiteral)
		if !ok {
			t.Fatalf("expected FloatLiteral, got %T", stmt.Expression)
		}

		if literal.Value != tt.expected {
			t.Errorf("expected %f, got %f", tt.expected, literal.Value)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"hello\nworld"`, "hello\\nworld"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		literal, ok := stmt.Expression.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("expected StringLiteral, got %T", stmt.Expression)
		}

		if literal.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, literal.Value)
		}
	}
}

func TestSymbolLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{":foo", "foo"},
		{":Bar", "Bar"},
		{":foo_bar", "foo_bar"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		literal, ok := stmt.Expression.(*ast.SymbolLiteral)
		if !ok {
			t.Fatalf("expected SymbolLiteral, got %T", stmt.Expression)
		}

		if literal.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, literal.Value)
		}
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		literal, ok := stmt.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("expected BooleanLiteral, got %T", stmt.Expression)
		}

		if literal.Value != tt.expected {
			t.Errorf("expected %t, got %t", tt.expected, literal.Value)
		}
	}
}

func TestNilLiteral(t *testing.T) {
	l := lexer.New("nil")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.NilLiteral)
	if !ok {
		t.Fatalf("expected NilLiteral, got %T", stmt.Expression)
	}
}

func TestSelfExpression(t *testing.T) {
	l := lexer.New("self")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.SelfExpression)
	if !ok {
		t.Fatalf("expected SelfExpression, got %T", stmt.Expression)
	}
}

func TestIdentifier(t *testing.T) {
	l := lexer.New("foobar")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("expected foobar, got %s", ident.Value)
	}
}

func TestConstant(t *testing.T) {
	l := lexer.New("FooBar")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	constant, ok := stmt.Expression.(*ast.Constant)
	if !ok {
		t.Fatalf("expected Constant, got %T", stmt.Expression)
	}

	if constant.Value != "FooBar" {
		t.Errorf("expected FooBar, got %s", constant.Value)
	}
}

func TestInstanceVariable(t *testing.T) {
	l := lexer.New("@foo")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ivar, ok := stmt.Expression.(*ast.InstanceVariable)
	if !ok {
		t.Fatalf("expected InstanceVariable, got %T", stmt.Expression)
	}

	if ivar.Name != "@foo" {
		t.Errorf("expected @foo, got %s", ivar.Name)
	}
}

func TestClassVariable(t *testing.T) {
	l := lexer.New("@@foo")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	cvar, ok := stmt.Expression.(*ast.ClassVariable)
	if !ok {
		t.Fatalf("expected ClassVariable, got %T", stmt.Expression)
	}

	if cvar.Name != "@@foo" {
		t.Errorf("expected @@foo, got %s", cvar.Name)
	}
}

func TestGlobalVariable(t *testing.T) {
	l := lexer.New("$foo")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	gvar, ok := stmt.Expression.(*ast.GlobalVariable)
	if !ok {
		t.Fatalf("expected GlobalVariable, got %T", stmt.Expression)
	}

	if gvar.Name != "$foo" {
		t.Errorf("expected $foo, got %s", gvar.Name)
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-5", "-", 5},
		{"!true", "!", true},
		{"~42", "~", 42},
		{"+5", "+", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("expected PrefixExpression, got %T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Errorf("expected operator %s, got %s", tt.operator, exp.Operator)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 % 5", 5, "%", 5},
		{"5 ** 2", 5, "**", 2},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 >= 5", 5, ">=", 5},
		{"5 <= 5", 5, "<=", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"5 <=> 5", 5, "<=>", 5},
		{"true && false", true, "&&", false},
		{"true || false", true, "||", false},
		{"5 & 3", 5, "&", 3},
		{"5 | 3", 5, "|", 3},
		{"5 ^ 3", 5, "^", 3},
		{"5 << 2", 5, "<<", 2},
		{"5 >> 2", 5, ">>", 2},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("expected InfixExpression for %q, got %T", tt.input, stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Errorf("expected operator %s, got %s", tt.operator, exp.Operator)
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 * 2 * 3", "((1 * 2) * 3)"},
		{"2 ** 3 ** 2", "(2 ** (3 ** 2))"},
		{"-1 + 2", "((-1) + 2)"},
		{"!true == false", "((!true) == false)"},
		{"1 < 2 == true", "((1 < 2) == true)"},
		{"1 && 2 || 3", "((1 && 2) || 3)"},
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, actual)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2, 3]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(array.Elements))
	}
}

func TestHashLiteral(t *testing.T) {
	input := `{:a => 1, :b => 2}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expected HashLiteral, got %T", stmt.Expression)
	}

	if len(hash.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(hash.Pairs))
	}
}

func TestHashLiteralWithLabels(t *testing.T) {
	input := `{a: 1, b: 2}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expected HashLiteral, got %T", stmt.Expression)
	}

	if len(hash.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(hash.Pairs))
	}
}

func TestRangeLiteral(t *testing.T) {
	tests := []struct {
		input     string
		exclusive bool
	}{
		{"1..10", false},
		{"1...10", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		rng, ok := stmt.Expression.(*ast.RangeLiteral)
		if !ok {
			t.Fatalf("expected RangeLiteral, got %T", stmt.Expression)
		}

		if rng.Exclusive != tt.exclusive {
			t.Errorf("expected exclusive=%t, got %t", tt.exclusive, rng.Exclusive)
		}
	}
}

func TestIndexExpression(t *testing.T) {
	input := "arr[0]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	index, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected IndexExpression, got %T", stmt.Expression)
	}

	if index.Left.String() != "arr" {
		t.Errorf("expected arr, got %s", index.Left.String())
	}
}

func TestMethodCall(t *testing.T) {
	tests := []struct {
		input  string
		method string
		args   int
	}{
		{"foo()", "foo", 0},
		{"foo(1)", "foo", 1},
		{"foo(1, 2)", "foo", 2},
		{"foo 1, 2", "foo", 2},
		{"obj.bar()", "bar", 0},
		{"obj.bar(1)", "bar", 1},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		call, ok := stmt.Expression.(*ast.MethodCall)
		if !ok {
			t.Fatalf("expected MethodCall for %q, got %T", tt.input, stmt.Expression)
		}

		if call.Method != tt.method {
			t.Errorf("expected method %s, got %s", tt.method, call.Method)
		}

		if len(call.Arguments) != tt.args {
			t.Errorf("expected %d args, got %d", tt.args, len(call.Arguments))
		}
	}
}

func TestMethodCallChain(t *testing.T) {
	input := "a.b.c"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.MethodCall)
	if !ok {
		t.Fatalf("expected MethodCall, got %T", stmt.Expression)
	}

	if call.Method != "c" {
		t.Errorf("expected method c, got %s", call.Method)
	}
}

func TestSafeNavigation(t *testing.T) {
	input := "obj&.foo"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.MethodCall)
	if !ok {
		t.Fatalf("expected MethodCall, got %T", stmt.Expression)
	}

	if !call.SafeNav {
		t.Error("expected SafeNav to be true")
	}
}

func TestAssignment(t *testing.T) {
	input := "x = 5"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	assign, ok := stmt.Expression.(*ast.AssignmentExpression)
	if !ok {
		t.Fatalf("expected AssignmentExpression, got %T", stmt.Expression)
	}

	if assign.Left.String() != "x" {
		t.Errorf("expected x, got %s", assign.Left.String())
	}
}

func TestOpAssignment(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"x += 5", "+="},
		{"x -= 5", "-="},
		{"x *= 5", "*="},
		{"x /= 5", "/="},
		{"x ||= 5", "||="},
		{"x &&= 5", "&&="},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		assign, ok := stmt.Expression.(*ast.OpAssignmentExpression)
		if !ok {
			t.Fatalf("expected OpAssignmentExpression for %q, got %T", tt.input, stmt.Expression)
		}

		if assign.Operator != tt.operator {
			t.Errorf("expected %s, got %s", tt.operator, assign.Operator)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x > 5
  puts "big"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", stmt.Expression)
	}

	if ifExp.Condition == nil {
		t.Error("expected condition")
	}

	if ifExp.Consequence == nil {
		t.Error("expected consequence")
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if x > 5
  puts "big"
else
  puts "small"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", stmt.Expression)
	}

	if ifExp.ElseBody == nil {
		t.Error("expected else body")
	}
}

func TestIfElsifElseExpression(t *testing.T) {
	input := `if x > 10
  puts "big"
elsif x > 5
  puts "medium"
else
  puts "small"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", stmt.Expression)
	}

	if ifExp.Alternative == nil {
		t.Error("expected alternative (elsif)")
	}
}

func TestUnlessExpression(t *testing.T) {
	input := `unless x > 5
  puts "small"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", stmt.Expression)
	}

	if !ifExp.Unless {
		t.Error("expected Unless to be true")
	}
}

func TestTernaryExpression(t *testing.T) {
	input := "x > 5 ? true : false"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ternary, ok := stmt.Expression.(*ast.TernaryExpression)
	if !ok {
		t.Fatalf("expected TernaryExpression, got %T", stmt.Expression)
	}

	if ternary.Condition == nil {
		t.Error("expected condition")
	}
}

func TestModifierIf(t *testing.T) {
	input := "puts x if x > 5"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	mod, ok := stmt.Expression.(*ast.ModifierExpression)
	if !ok {
		t.Fatalf("expected ModifierExpression, got %T", stmt.Expression)
	}

	if mod.Modifier != "if" {
		t.Errorf("expected if, got %s", mod.Modifier)
	}
}

func TestCaseExpression(t *testing.T) {
	input := `case x
when 1
  puts "one"
when 2
  puts "two"
else
  puts "other"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	caseExp, ok := stmt.Expression.(*ast.CaseExpression)
	if !ok {
		t.Fatalf("expected CaseExpression, got %T", stmt.Expression)
	}

	if len(caseExp.Whens) != 2 {
		t.Errorf("expected 2 when clauses, got %d", len(caseExp.Whens))
	}
}

func TestWhileExpression(t *testing.T) {
	input := `while x > 0
  x -= 1
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	whileExp, ok := stmt.Expression.(*ast.WhileExpression)
	if !ok {
		t.Fatalf("expected WhileExpression, got %T", stmt.Expression)
	}

	if whileExp.Until {
		t.Error("expected Until to be false")
	}
}

func TestUntilExpression(t *testing.T) {
	input := `until x == 0
  x -= 1
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	untilExp, ok := stmt.Expression.(*ast.WhileExpression)
	if !ok {
		t.Fatalf("expected WhileExpression, got %T", stmt.Expression)
	}

	if !untilExp.Until {
		t.Error("expected Until to be true")
	}
}

func TestForExpression(t *testing.T) {
	input := `for i in 1..10
  puts i
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	forExp, ok := stmt.Expression.(*ast.ForExpression)
	if !ok {
		t.Fatalf("expected ForExpression, got %T", stmt.Expression)
	}

	if forExp.Variable == nil {
		t.Error("expected variable")
	}
}

func TestBeginRescue(t *testing.T) {
	input := `begin
  risky_operation
rescue StandardError => e
  handle_error(e)
ensure
  cleanup
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	beginExp, ok := stmt.Expression.(*ast.BeginExpression)
	if !ok {
		t.Fatalf("expected BeginExpression, got %T", stmt.Expression)
	}

	if len(beginExp.Rescues) != 1 {
		t.Errorf("expected 1 rescue, got %d", len(beginExp.Rescues))
	}

	if beginExp.Ensure == nil {
		t.Error("expected ensure")
	}
}

func TestMethodDefinition(t *testing.T) {
	input := `def foo(a, b = 1, *args, **kwargs, &block)
  a + b
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	method, ok := program.Statements[0].(*ast.MethodDefinition)
	if !ok {
		t.Fatalf("expected MethodDefinition, got %T", program.Statements[0])
	}

	if method.Name != "foo" {
		t.Errorf("expected foo, got %s", method.Name)
	}

	if len(method.Parameters) != 5 {
		t.Errorf("expected 5 parameters, got %d", len(method.Parameters))
	}
}

func TestSingletonMethodDefinition(t *testing.T) {
	input := `def self.foo
  42
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	method, ok := program.Statements[0].(*ast.MethodDefinition)
	if !ok {
		t.Fatalf("expected MethodDefinition, got %T", program.Statements[0])
	}

	if method.Receiver == nil {
		t.Error("expected receiver")
	}
}

func TestClassDefinition(t *testing.T) {
	input := `class Foo < Bar
  def initialize
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class, ok := program.Statements[0].(*ast.ClassDefinition)
	if !ok {
		t.Fatalf("expected ClassDefinition, got %T", program.Statements[0])
	}

	if class.Name.Value != "Foo" {
		t.Errorf("expected Foo, got %s", class.Name.Value)
	}

	if class.Superclass == nil {
		t.Error("expected superclass")
	}
}

func TestModuleDefinition(t *testing.T) {
	input := `module Foo
  def bar
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	module, ok := program.Statements[0].(*ast.ModuleDefinition)
	if !ok {
		t.Fatalf("expected ModuleDefinition, got %T", program.Statements[0])
	}

	if module.Name.Value != "Foo" {
		t.Errorf("expected Foo, got %s", module.Name.Value)
	}
}

func TestBlock(t *testing.T) {
	input := "[1, 2, 3].each { |x| puts x }"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.MethodCall)
	if !ok {
		t.Fatalf("expected MethodCall, got %T", stmt.Expression)
	}

	if call.Block == nil {
		t.Error("expected block")
	}

	if len(call.Block.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(call.Block.Parameters))
	}
}

func TestDoBlock(t *testing.T) {
	input := `[1, 2, 3].each do |x|
  puts x
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.MethodCall)
	if !ok {
		t.Fatalf("expected MethodCall, got %T", stmt.Expression)
	}

	if call.Block == nil {
		t.Error("expected block")
	}
}

func TestLambda(t *testing.T) {
	input := "->(x) { x * 2 }"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	lambda, ok := stmt.Expression.(*ast.Lambda)
	if !ok {
		t.Fatalf("expected Lambda, got %T", stmt.Expression)
	}

	if len(lambda.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(lambda.Parameters))
	}
}

func TestReturnStatement(t *testing.T) {
	input := "return 5"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	ret, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", program.Statements[0])
	}

	if ret.Value == nil {
		t.Error("expected return value")
	}
}

func TestBreakNextRedo(t *testing.T) {
	tests := []struct {
		input string
		kind  string
	}{
		{"break", "break"},
		{"next", "next"},
		{"redo", "redo"},
		{"retry", "retry"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Errorf("expected 1 statement for %q", tt.input)
		}
	}
}

func TestYield(t *testing.T) {
	input := "yield 1, 2"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	yield, ok := stmt.Expression.(*ast.YieldExpression)
	if !ok {
		t.Fatalf("expected YieldExpression, got %T", stmt.Expression)
	}

	if len(yield.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(yield.Arguments))
	}
}

func TestSuper(t *testing.T) {
	input := "super(1, 2)"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	super, ok := stmt.Expression.(*ast.SuperExpression)
	if !ok {
		t.Fatalf("expected SuperExpression, got %T", stmt.Expression)
	}

	if len(super.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(super.Arguments))
	}
}

func TestScopedConstant(t *testing.T) {
	input := "Foo::Bar::Baz"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	scoped, ok := stmt.Expression.(*ast.ScopedConstant)
	if !ok {
		t.Fatalf("expected ScopedConstant, got %T", stmt.Expression)
	}

	if scoped.Name != "Baz" {
		t.Errorf("expected Baz, got %s", scoped.Name)
	}
}

func TestNotExpression(t *testing.T) {
	input := "not x"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	notExp, ok := stmt.Expression.(*ast.NotExpression)
	if !ok {
		t.Fatalf("expected NotExpression, got %T", stmt.Expression)
	}

	if notExp.Expression == nil {
		t.Error("expected expression")
	}
}

func TestAndOrExpressions(t *testing.T) {
	tests := []struct {
		input string
		isAnd bool
	}{
		{"x and y", true},
		{"x or y", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		if tt.isAnd {
			_, ok := stmt.Expression.(*ast.AndExpression)
			if !ok {
				t.Fatalf("expected AndExpression, got %T", stmt.Expression)
			}
		} else {
			_, ok := stmt.Expression.(*ast.OrExpression)
			if !ok {
				t.Fatalf("expected OrExpression, got %T", stmt.Expression)
			}
		}
	}
}

func TestRescueModifier(t *testing.T) {
	input := "risky rescue default"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	rescue, ok := stmt.Expression.(*ast.RescueModifier)
	if !ok {
		t.Fatalf("expected RescueModifier, got %T", stmt.Expression)
	}

	if rescue.Body == nil {
		t.Error("expected body")
	}
}

func TestInterpolatedString(t *testing.T) {
	input := `"hello #{name}"`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	str, ok := stmt.Expression.(*ast.InterpolatedString)
	if !ok {
		t.Fatalf("expected InterpolatedString, got %T", stmt.Expression)
	}

	if len(str.Parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(str.Parts))
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}
