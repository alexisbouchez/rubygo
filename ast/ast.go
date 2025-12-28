// Package ast defines the Abstract Syntax Tree for Ruby.
package ast

import (
	"bytes"
	"strings"

	"github.com/alexisbouchez/rubylexer/token"
)

// Node represents a node in the AST.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a float value.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// InterpolatedString represents a string with interpolation.
type InterpolatedString struct {
	Token token.Token
	Parts []Expression // StringLiteral or interpolated expressions
}

func (is *InterpolatedString) expressionNode()      {}
func (is *InterpolatedString) TokenLiteral() string { return is.Token.Literal }
func (is *InterpolatedString) String() string {
	var out bytes.Buffer
	out.WriteString("\"")
	for _, part := range is.Parts {
		if sl, ok := part.(*StringLiteral); ok {
			out.WriteString(sl.Value)
		} else {
			out.WriteString("#{")
			out.WriteString(part.String())
			out.WriteString("}")
		}
	}
	out.WriteString("\"")
	return out.String()
}

// SymbolLiteral represents a symbol.
type SymbolLiteral struct {
	Token token.Token
	Value string
}

func (sl *SymbolLiteral) expressionNode()      {}
func (sl *SymbolLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *SymbolLiteral) String() string       { return ":" + sl.Value }

// RegexpLiteral represents a regular expression.
type RegexpLiteral struct {
	Token token.Token
	Value string
	Flags string
}

func (rl *RegexpLiteral) expressionNode()      {}
func (rl *RegexpLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RegexpLiteral) String() string       { return "/" + rl.Value + "/" + rl.Flags }

// NilLiteral represents nil.
type NilLiteral struct {
	Token token.Token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "nil" }

// BooleanLiteral represents true or false.
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

// SelfExpression represents self.
type SelfExpression struct {
	Token token.Token
}

func (se *SelfExpression) expressionNode()      {}
func (se *SelfExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SelfExpression) String() string       { return "self" }

// Identifier represents a local variable or method name.
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// Constant represents a constant.
type Constant struct {
	Token token.Token
	Value string
}

func (c *Constant) expressionNode()      {}
func (c *Constant) TokenLiteral() string { return c.Token.Literal }
func (c *Constant) String() string       { return c.Value }

// InstanceVariable represents an instance variable (@foo).
type InstanceVariable struct {
	Token token.Token
	Name  string
}

func (iv *InstanceVariable) expressionNode()      {}
func (iv *InstanceVariable) TokenLiteral() string { return iv.Token.Literal }
func (iv *InstanceVariable) String() string       { return iv.Name }

// ClassVariable represents a class variable (@@foo).
type ClassVariable struct {
	Token token.Token
	Name  string
}

func (cv *ClassVariable) expressionNode()      {}
func (cv *ClassVariable) TokenLiteral() string { return cv.Token.Literal }
func (cv *ClassVariable) String() string       { return cv.Name }

// GlobalVariable represents a global variable ($foo).
type GlobalVariable struct {
	Token token.Token
	Name  string
}

func (gv *GlobalVariable) expressionNode()      {}
func (gv *GlobalVariable) TokenLiteral() string { return gv.Token.Literal }
func (gv *GlobalVariable) String() string       { return gv.Name }

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := make([]string, len(al.Elements))
	for i, e := range al.Elements {
		elements[i] = e.String()
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// HashLiteral represents a hash literal.
type HashLiteral struct {
	Token         token.Token
	Pairs         map[Expression]Expression
	Order         []Expression // To maintain key order
	IsKeywordArgs bool         // True if this is an implicit keyword arguments hash
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := make([]string, 0, len(hl.Pairs))
	for _, key := range hl.Order {
		value := hl.Pairs[key]
		pairs = append(pairs, key.String()+" => "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// RangeLiteral represents a range (1..10 or 1...10).
type RangeLiteral struct {
	Token     token.Token
	Start     Expression
	End       Expression
	Exclusive bool // true for ..., false for ..
}

func (rl *RangeLiteral) expressionNode()      {}
func (rl *RangeLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RangeLiteral) String() string {
	var out bytes.Buffer
	if rl.Start != nil {
		out.WriteString(rl.Start.String())
	}
	if rl.Exclusive {
		out.WriteString("...")
	} else {
		out.WriteString("..")
	}
	if rl.End != nil {
		out.WriteString(rl.End.String())
	}
	return out.String()
}

// PrefixExpression represents a prefix operator expression.
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix operator expression.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// AssignmentExpression represents variable assignment.
type AssignmentExpression struct {
	Token token.Token
	Left  Expression
	Value Expression
}

func (ae *AssignmentExpression) expressionNode()      {}
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ae.Left.String())
	out.WriteString(" = ")
	out.WriteString(ae.Value.String())
	return out.String()
}

// OpAssignmentExpression represents compound assignment (+=, -=, etc.).
type OpAssignmentExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Value    Expression
}

func (oa *OpAssignmentExpression) expressionNode()      {}
func (oa *OpAssignmentExpression) TokenLiteral() string { return oa.Token.Literal }
func (oa *OpAssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString(oa.Left.String())
	out.WriteString(" " + oa.Operator + " ")
	out.WriteString(oa.Value.String())
	return out.String()
}

// MultipleAssignment represents parallel assignment (a, b = 1, 2).
type MultipleAssignment struct {
	Token  token.Token
	Left   []Expression
	Right  []Expression
}

func (ma *MultipleAssignment) expressionNode()      {}
func (ma *MultipleAssignment) TokenLiteral() string { return ma.Token.Literal }
func (ma *MultipleAssignment) String() string {
	var out bytes.Buffer
	lefts := make([]string, len(ma.Left))
	for i, l := range ma.Left {
		lefts[i] = l.String()
	}
	rights := make([]string, len(ma.Right))
	for i, r := range ma.Right {
		rights[i] = r.String()
	}
	out.WriteString(strings.Join(lefts, ", "))
	out.WriteString(" = ")
	out.WriteString(strings.Join(rights, ", "))
	return out.String()
}

// MethodCall represents a method call.
type MethodCall struct {
	Token     token.Token
	Receiver  Expression // nil if implicit self
	Method    string
	Arguments []Expression
	Block     *Block
	SafeNav   bool // true if using &.
}

func (mc *MethodCall) expressionNode()      {}
func (mc *MethodCall) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCall) String() string {
	var out bytes.Buffer
	if mc.Receiver != nil {
		out.WriteString(mc.Receiver.String())
		if mc.SafeNav {
			out.WriteString("&.")
		} else {
			out.WriteString(".")
		}
	}
	out.WriteString(mc.Method)
	if len(mc.Arguments) > 0 || mc.Block == nil {
		out.WriteString("(")
		args := make([]string, len(mc.Arguments))
		for i, a := range mc.Arguments {
			args[i] = a.String()
		}
		out.WriteString(strings.Join(args, ", "))
		out.WriteString(")")
	}
	if mc.Block != nil {
		out.WriteString(" ")
		out.WriteString(mc.Block.String())
	}
	return out.String()
}

// IndexExpression represents array/hash indexing.
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	return out.String()
}

// Block represents a block (do...end or {...}).
type Block struct {
	Token      token.Token
	Parameters []*BlockParameter
	Body       *BlockBody
}

func (b *Block) expressionNode()      {}
func (b *Block) TokenLiteral() string { return b.Token.Literal }
func (b *Block) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	if len(b.Parameters) > 0 {
		out.WriteString("|")
		params := make([]string, len(b.Parameters))
		for i, p := range b.Parameters {
			params[i] = p.String()
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString("| ")
	}
	out.WriteString(b.Body.String())
	out.WriteString(" }")
	return out.String()
}

// BlockParameter represents a block parameter.
type BlockParameter struct {
	Token   token.Token
	Name    string
	Splat   bool // *args
	DSplat  bool // **kwargs
	Block   bool // &block
	Default Expression
}

func (bp *BlockParameter) String() string {
	var out bytes.Buffer
	if bp.Splat {
		out.WriteString("*")
	}
	if bp.DSplat {
		out.WriteString("**")
	}
	if bp.Block {
		out.WriteString("&")
	}
	out.WriteString(bp.Name)
	if bp.Default != nil {
		out.WriteString(" = ")
		out.WriteString(bp.Default.String())
	}
	return out.String()
}

// BlockBody represents the body of a block.
type BlockBody struct {
	Statements []Statement
}

func (bb *BlockBody) String() string {
	var out bytes.Buffer
	for i, s := range bb.Statements {
		if i > 0 {
			out.WriteString("; ")
		}
		out.WriteString(s.String())
	}
	return out.String()
}

// Lambda represents a lambda (-> { }).
type Lambda struct {
	Token      token.Token
	Parameters []*BlockParameter
	Body       *BlockBody
}

func (l *Lambda) expressionNode()      {}
func (l *Lambda) TokenLiteral() string { return l.Token.Literal }
func (l *Lambda) String() string {
	var out bytes.Buffer
	out.WriteString("->")
	if len(l.Parameters) > 0 {
		out.WriteString("(")
		params := make([]string, len(l.Parameters))
		for i, p := range l.Parameters {
			params[i] = p.String()
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(")")
	}
	out.WriteString(" { ")
	out.WriteString(l.Body.String())
	out.WriteString(" }")
	return out.String()
}

// IfExpression represents an if/unless expression.
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockBody
	Alternative *IfExpression // for elsif chain
	ElseBody    *BlockBody    // for else
	Unless      bool          // true if unless
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	if ie.Unless {
		out.WriteString("unless ")
	} else {
		out.WriteString("if ")
	}
	out.WriteString(ie.Condition.String())
	out.WriteString("\n")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("\nelsif ")
		out.WriteString(ie.Alternative.String())
	}
	if ie.ElseBody != nil {
		out.WriteString("\nelse\n")
		out.WriteString(ie.ElseBody.String())
	}
	out.WriteString("\nend")
	return out.String()
}

// TernaryExpression represents a ? b : c.
type TernaryExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString(te.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(te.Consequence.String())
	out.WriteString(" : ")
	out.WriteString(te.Alternative.String())
	return out.String()
}

// ModifierExpression represents modifier if/unless/while/until.
type ModifierExpression struct {
	Token     token.Token
	Body      Expression
	Condition Expression
	Modifier  string // "if", "unless", "while", "until"
}

func (me *ModifierExpression) expressionNode()      {}
func (me *ModifierExpression) TokenLiteral() string { return me.Token.Literal }
func (me *ModifierExpression) String() string {
	var out bytes.Buffer
	out.WriteString(me.Body.String())
	out.WriteString(" ")
	out.WriteString(me.Modifier)
	out.WriteString(" ")
	out.WriteString(me.Condition.String())
	return out.String()
}

// CaseExpression represents a case/when expression.
type CaseExpression struct {
	Token   token.Token
	Subject Expression // can be nil for case without subject
	Whens   []*WhenClause
	Else    *BlockBody
}

func (ce *CaseExpression) expressionNode()      {}
func (ce *CaseExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CaseExpression) String() string {
	var out bytes.Buffer
	out.WriteString("case")
	if ce.Subject != nil {
		out.WriteString(" ")
		out.WriteString(ce.Subject.String())
	}
	out.WriteString("\n")
	for _, w := range ce.Whens {
		out.WriteString(w.String())
		out.WriteString("\n")
	}
	if ce.Else != nil {
		out.WriteString("else\n")
		out.WriteString(ce.Else.String())
		out.WriteString("\n")
	}
	out.WriteString("end")
	return out.String()
}

// WhenClause represents a when clause in a case expression.
type WhenClause struct {
	Token      token.Token
	Conditions []Expression
	Body       *BlockBody
}

func (wc *WhenClause) String() string {
	var out bytes.Buffer
	out.WriteString("when ")
	conds := make([]string, len(wc.Conditions))
	for i, c := range wc.Conditions {
		conds[i] = c.String()
	}
	out.WriteString(strings.Join(conds, ", "))
	out.WriteString("\n")
	out.WriteString(wc.Body.String())
	return out.String()
}

// WhileExpression represents a while/until loop.
type WhileExpression struct {
	Token     token.Token
	Condition Expression
	Body      *BlockBody
	Until     bool // true if until
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	var out bytes.Buffer
	if we.Until {
		out.WriteString("until ")
	} else {
		out.WriteString("while ")
	}
	out.WriteString(we.Condition.String())
	out.WriteString("\n")
	out.WriteString(we.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// ForExpression represents a for loop.
type ForExpression struct {
	Token    token.Token
	Variable Expression
	Iterable Expression
	Body     *BlockBody
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fe.Variable.String())
	out.WriteString(" in ")
	out.WriteString(fe.Iterable.String())
	out.WriteString("\n")
	out.WriteString(fe.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// BeginExpression represents a begin/rescue/ensure block.
type BeginExpression struct {
	Token   token.Token
	Body    *BlockBody
	Rescues []*RescueClause
	Else    *BlockBody
	Ensure  *BlockBody
}

func (be *BeginExpression) expressionNode()      {}
func (be *BeginExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BeginExpression) String() string {
	var out bytes.Buffer
	out.WriteString("begin\n")
	out.WriteString(be.Body.String())
	for _, r := range be.Rescues {
		out.WriteString("\n")
		out.WriteString(r.String())
	}
	if be.Else != nil {
		out.WriteString("\nelse\n")
		out.WriteString(be.Else.String())
	}
	if be.Ensure != nil {
		out.WriteString("\nensure\n")
		out.WriteString(be.Ensure.String())
	}
	out.WriteString("\nend")
	return out.String()
}

// RescueClause represents a rescue clause.
type RescueClause struct {
	Token      token.Token
	Exceptions []Expression
	Variable   *Identifier
	Body       *BlockBody
}

func (rc *RescueClause) String() string {
	var out bytes.Buffer
	out.WriteString("rescue")
	if len(rc.Exceptions) > 0 {
		out.WriteString(" ")
		excs := make([]string, len(rc.Exceptions))
		for i, e := range rc.Exceptions {
			excs[i] = e.String()
		}
		out.WriteString(strings.Join(excs, ", "))
	}
	if rc.Variable != nil {
		out.WriteString(" => ")
		out.WriteString(rc.Variable.String())
	}
	out.WriteString("\n")
	out.WriteString(rc.Body.String())
	return out.String()
}

// MethodDefinition represents a method definition.
type MethodDefinition struct {
	Token      token.Token
	Name       string
	Receiver   Expression // for singleton methods (def self.foo)
	Parameters []*MethodParameter
	Body       *BlockBody
}

func (md *MethodDefinition) statementNode()       {}
func (md *MethodDefinition) TokenLiteral() string { return md.Token.Literal }
func (md *MethodDefinition) String() string {
	var out bytes.Buffer
	out.WriteString("def ")
	if md.Receiver != nil {
		out.WriteString(md.Receiver.String())
		out.WriteString(".")
	}
	out.WriteString(md.Name)
	if len(md.Parameters) > 0 {
		out.WriteString("(")
		params := make([]string, len(md.Parameters))
		for i, p := range md.Parameters {
			params[i] = p.String()
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(")")
	}
	out.WriteString("\n")
	out.WriteString(md.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// MethodParameter represents a method parameter.
type MethodParameter struct {
	Token   token.Token
	Name    string
	Splat   bool       // *args
	DSplat  bool       // **kwargs
	Block   bool       // &block
	Default Expression // default value
	KeywordOnly bool   // keyword-only parameter
}

func (mp *MethodParameter) String() string {
	var out bytes.Buffer
	if mp.Splat {
		out.WriteString("*")
	}
	if mp.DSplat {
		out.WriteString("**")
	}
	if mp.Block {
		out.WriteString("&")
	}
	out.WriteString(mp.Name)
	if mp.KeywordOnly {
		out.WriteString(":")
		if mp.Default != nil {
			out.WriteString(" ")
			out.WriteString(mp.Default.String())
		}
	} else if mp.Default != nil {
		out.WriteString(" = ")
		out.WriteString(mp.Default.String())
	}
	return out.String()
}

// ClassDefinition represents a class definition.
type ClassDefinition struct {
	Token      token.Token
	Name       *Constant
	Superclass Expression
	Body       *BlockBody
}

func (cd *ClassDefinition) statementNode()       {}
func (cd *ClassDefinition) TokenLiteral() string { return cd.Token.Literal }
func (cd *ClassDefinition) String() string {
	var out bytes.Buffer
	out.WriteString("class ")
	out.WriteString(cd.Name.String())
	if cd.Superclass != nil {
		out.WriteString(" < ")
		out.WriteString(cd.Superclass.String())
	}
	out.WriteString("\n")
	out.WriteString(cd.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// SingletonClassDefinition represents a singleton class definition (class << obj).
type SingletonClassDefinition struct {
	Token  token.Token
	Object Expression
	Body   *BlockBody
}

func (scd *SingletonClassDefinition) statementNode()       {}
func (scd *SingletonClassDefinition) TokenLiteral() string { return scd.Token.Literal }
func (scd *SingletonClassDefinition) String() string {
	var out bytes.Buffer
	out.WriteString("class << ")
	out.WriteString(scd.Object.String())
	out.WriteString("\n")
	out.WriteString(scd.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// ModuleDefinition represents a module definition.
type ModuleDefinition struct {
	Token token.Token
	Name  *Constant
	Body  *BlockBody
}

func (md *ModuleDefinition) statementNode()       {}
func (md *ModuleDefinition) TokenLiteral() string { return md.Token.Literal }
func (md *ModuleDefinition) String() string {
	var out bytes.Buffer
	out.WriteString("module ")
	out.WriteString(md.Name.String())
	out.WriteString("\n")
	out.WriteString(md.Body.String())
	out.WriteString("\nend")
	return out.String()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return")
	if rs.Value != nil {
		out.WriteString(" ")
		out.WriteString(rs.Value.String())
	}
	return out.String()
}

// BreakStatement represents a break statement.
type BreakStatement struct {
	Token token.Token
	Value Expression
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string {
	var out bytes.Buffer
	out.WriteString("break")
	if bs.Value != nil {
		out.WriteString(" ")
		out.WriteString(bs.Value.String())
	}
	return out.String()
}

// NextStatement represents a next statement.
type NextStatement struct {
	Token token.Token
	Value Expression
}

func (ns *NextStatement) statementNode()       {}
func (ns *NextStatement) TokenLiteral() string { return ns.Token.Literal }
func (ns *NextStatement) String() string {
	var out bytes.Buffer
	out.WriteString("next")
	if ns.Value != nil {
		out.WriteString(" ")
		out.WriteString(ns.Value.String())
	}
	return out.String()
}

// RedoStatement represents a redo statement.
type RedoStatement struct {
	Token token.Token
}

func (rs *RedoStatement) statementNode()       {}
func (rs *RedoStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RedoStatement) String() string       { return "redo" }

// RetryStatement represents a retry statement.
type RetryStatement struct {
	Token token.Token
}

func (rs *RetryStatement) statementNode()       {}
func (rs *RetryStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RetryStatement) String() string       { return "retry" }

// YieldExpression represents yield.
type YieldExpression struct {
	Token     token.Token
	Arguments []Expression
}

func (ye *YieldExpression) expressionNode()      {}
func (ye *YieldExpression) TokenLiteral() string { return ye.Token.Literal }
func (ye *YieldExpression) String() string {
	var out bytes.Buffer
	out.WriteString("yield")
	if len(ye.Arguments) > 0 {
		out.WriteString("(")
		args := make([]string, len(ye.Arguments))
		for i, a := range ye.Arguments {
			args[i] = a.String()
		}
		out.WriteString(strings.Join(args, ", "))
		out.WriteString(")")
	}
	return out.String()
}

// SuperExpression represents super.
type SuperExpression struct {
	Token     token.Token
	Arguments []Expression
	HasParens bool // true if super() vs super
}

func (se *SuperExpression) expressionNode()      {}
func (se *SuperExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SuperExpression) String() string {
	var out bytes.Buffer
	out.WriteString("super")
	if se.HasParens || len(se.Arguments) > 0 {
		out.WriteString("(")
		args := make([]string, len(se.Arguments))
		for i, a := range se.Arguments {
			args[i] = a.String()
		}
		out.WriteString(strings.Join(args, ", "))
		out.WriteString(")")
	}
	return out.String()
}

// DefinedExpression represents defined?(expr).
type DefinedExpression struct {
	Token      token.Token
	Expression Expression
}

func (de *DefinedExpression) expressionNode()      {}
func (de *DefinedExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DefinedExpression) String() string {
	var out bytes.Buffer
	out.WriteString("defined?(")
	out.WriteString(de.Expression.String())
	out.WriteString(")")
	return out.String()
}

// AliasStatement represents an alias statement.
type AliasStatement struct {
	Token token.Token
	New   Expression
	Old   Expression
}

func (as *AliasStatement) statementNode()       {}
func (as *AliasStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AliasStatement) String() string {
	var out bytes.Buffer
	out.WriteString("alias ")
	out.WriteString(as.New.String())
	out.WriteString(" ")
	out.WriteString(as.Old.String())
	return out.String()
}

// UndefStatement represents an undef statement.
type UndefStatement struct {
	Token   token.Token
	Methods []Expression
}

func (us *UndefStatement) statementNode()       {}
func (us *UndefStatement) TokenLiteral() string { return us.Token.Literal }
func (us *UndefStatement) String() string {
	var out bytes.Buffer
	out.WriteString("undef ")
	methods := make([]string, len(us.Methods))
	for i, m := range us.Methods {
		methods[i] = m.String()
	}
	out.WriteString(strings.Join(methods, ", "))
	return out.String()
}

// ScopedConstant represents a scoped constant (Foo::Bar).
type ScopedConstant struct {
	Token token.Token
	Left  Expression
	Name  string
}

func (sc *ScopedConstant) expressionNode()      {}
func (sc *ScopedConstant) TokenLiteral() string { return sc.Token.Literal }
func (sc *ScopedConstant) String() string {
	var out bytes.Buffer
	if sc.Left != nil {
		out.WriteString(sc.Left.String())
	}
	out.WriteString("::")
	out.WriteString(sc.Name)
	return out.String()
}

// SplatExpression represents *expr.
type SplatExpression struct {
	Token      token.Token
	Expression Expression
}

func (se *SplatExpression) expressionNode()      {}
func (se *SplatExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SplatExpression) String() string {
	var out bytes.Buffer
	out.WriteString("*")
	out.WriteString(se.Expression.String())
	return out.String()
}

// DoubleSplatExpression represents **expr.
type DoubleSplatExpression struct {
	Token      token.Token
	Expression Expression
}

func (dse *DoubleSplatExpression) expressionNode()      {}
func (dse *DoubleSplatExpression) TokenLiteral() string { return dse.Token.Literal }
func (dse *DoubleSplatExpression) String() string {
	var out bytes.Buffer
	out.WriteString("**")
	out.WriteString(dse.Expression.String())
	return out.String()
}

// BlockArgExpression represents &expr.
type BlockArgExpression struct {
	Token      token.Token
	Expression Expression
}

func (bae *BlockArgExpression) expressionNode()      {}
func (bae *BlockArgExpression) TokenLiteral() string { return bae.Token.Literal }
func (bae *BlockArgExpression) String() string {
	var out bytes.Buffer
	out.WriteString("&")
	out.WriteString(bae.Expression.String())
	return out.String()
}

// NotExpression represents not expr.
type NotExpression struct {
	Token      token.Token
	Expression Expression
}

func (ne *NotExpression) expressionNode()      {}
func (ne *NotExpression) TokenLiteral() string { return ne.Token.Literal }
func (ne *NotExpression) String() string {
	var out bytes.Buffer
	out.WriteString("not ")
	out.WriteString(ne.Expression.String())
	return out.String()
}

// AndExpression represents expr and expr.
type AndExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (ae *AndExpression) expressionNode()      {}
func (ae *AndExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AndExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ae.Left.String())
	out.WriteString(" and ")
	out.WriteString(ae.Right.String())
	out.WriteString(")")
	return out.String()
}

// OrExpression represents expr or expr.
type OrExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (oe *OrExpression) expressionNode()      {}
func (oe *OrExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *OrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" or ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")
	return out.String()
}

// RescueModifier represents expr rescue expr.
type RescueModifier struct {
	Token   token.Token
	Body    Expression
	Rescue  Expression
}

func (rm *RescueModifier) expressionNode()      {}
func (rm *RescueModifier) TokenLiteral() string { return rm.Token.Literal }
func (rm *RescueModifier) String() string {
	var out bytes.Buffer
	out.WriteString(rm.Body.String())
	out.WriteString(" rescue ")
	out.WriteString(rm.Rescue.String())
	return out.String()
}

// MagicComment represents __FILE__, __LINE__, __ENCODING__.
type MagicComment struct {
	Token token.Token
	Kind  string // "FILE", "LINE", "ENCODING"
}

func (mc *MagicComment) expressionNode()      {}
func (mc *MagicComment) TokenLiteral() string { return mc.Token.Literal }
func (mc *MagicComment) String() string       { return "__" + mc.Kind + "__" }
