// Package object defines the Ruby object system.
package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"time"

	"github.com/alexisbouchez/rubylexer/ast"
)

// Type represents the type of an object.
type Type string

const (
	INTEGER_OBJ      Type = "INTEGER"
	FLOAT_OBJ        Type = "FLOAT"
	STRING_OBJ       Type = "STRING"
	SYMBOL_OBJ       Type = "SYMBOL"
	BOOLEAN_OBJ      Type = "BOOLEAN"
	NIL_OBJ          Type = "NIL"
	ARRAY_OBJ        Type = "ARRAY"
	HASH_OBJ         Type = "HASH"
	RANGE_OBJ        Type = "RANGE"
	REGEXP_OBJ       Type = "REGEXP"
	RETURN_VALUE_OBJ Type = "RETURN_VALUE"
	BREAK_VALUE_OBJ  Type = "BREAK_VALUE"
	NEXT_VALUE_OBJ   Type = "NEXT_VALUE"
	ERROR_OBJ        Type = "ERROR"
	PROC_OBJ         Type = "PROC"
	LAMBDA_OBJ       Type = "LAMBDA"
	METHOD_OBJ       Type = "METHOD"
	BUILTIN_OBJ      Type = "BUILTIN"
	CLASS_OBJ        Type = "CLASS"
	MODULE_OBJ       Type = "MODULE"
	INSTANCE_OBJ     Type = "INSTANCE"
	EXCEPTION_OBJ    Type = "EXCEPTION"
	TIME_OBJ         Type = "TIME"
	DATE_OBJ         Type = "DATE"
)

// Object is the base interface for all Ruby objects.
type Object interface {
	Type() Type
	Inspect() string
	Class() *RubyClass
	IsTruthy() bool
}

// Hashable is implemented by objects that can be hash keys.
type Hashable interface {
	HashKey() HashKey
}

// HashKey is used for hash map keys.
type HashKey struct {
	Type  Type
	Value uint64
}

// Integer represents a Ruby Integer.
type Integer struct {
	Value int64
}

func (i *Integer) Type() Type      { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Class() *RubyClass { return IntegerClass }
func (i *Integer) IsTruthy() bool  { return true }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// Float represents a Ruby Float.
type Float struct {
	Value float64
}

func (f *Float) Type() Type      { return FLOAT_OBJ }
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }
func (f *Float) Class() *RubyClass { return FloatClass }
func (f *Float) IsTruthy() bool  { return true }

// String represents a Ruby String.
type String struct {
	Value string
}

func (s *String) Type() Type      { return STRING_OBJ }
func (s *String) Inspect() string { return fmt.Sprintf("%q", s.Value) }
func (s *String) Class() *RubyClass { return StringClass }
func (s *String) IsTruthy() bool  { return true }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// Symbol represents a Ruby Symbol.
type Symbol struct {
	Value string
}

func (s *Symbol) Type() Type      { return SYMBOL_OBJ }
func (s *Symbol) Inspect() string { return ":" + s.Value }
func (s *Symbol) Class() *RubyClass { return SymbolClass }
func (s *Symbol) IsTruthy() bool  { return true }
func (s *Symbol) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// Boolean represents Ruby true or false.
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type      { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Class() *RubyClass {
	if b.Value {
		return TrueClass
	}
	return FalseClass
}
func (b *Boolean) IsTruthy() bool { return b.Value }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: b.Type(), Value: value}
}

// Nil represents Ruby nil.
type Nil struct{}

func (n *Nil) Type() Type      { return NIL_OBJ }
func (n *Nil) Inspect() string { return "nil" }
func (n *Nil) Class() *RubyClass { return NilClass }
func (n *Nil) IsTruthy() bool  { return false }

// Array represents a Ruby Array.
type Array struct {
	Elements []Object
}

func (a *Array) Type() Type { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.Inspect()
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
func (a *Array) Class() *RubyClass { return ArrayClass }
func (a *Array) IsTruthy() bool  { return true }

// HashPair represents a key-value pair in a Hash.
type HashPair struct {
	Key   Object
	Value Object
}

// Hash represents a Ruby Hash.
type Hash struct {
	Pairs        map[HashKey]HashPair
	Order        []HashKey // Maintain insertion order
	IsKeywordArgs bool      // True when this hash represents keyword arguments
}

func (h *Hash) Type() Type { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := make([]string, 0, len(h.Pairs))
	for _, key := range h.Order {
		pair := h.Pairs[key]
		pairs = append(pairs, fmt.Sprintf("%s => %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
func (h *Hash) Class() *RubyClass { return HashClass }
func (h *Hash) IsTruthy() bool  { return true }

// Range represents a Ruby Range.
type Range struct {
	Start     Object
	End       Object
	Exclusive bool
}

func (r *Range) Type() Type { return RANGE_OBJ }
func (r *Range) Inspect() string {
	op := ".."
	if r.Exclusive {
		op = "..."
	}
	return fmt.Sprintf("%s%s%s", r.Start.Inspect(), op, r.End.Inspect())
}
func (r *Range) Class() *RubyClass { return RangeClass }
func (r *Range) IsTruthy() bool  { return true }

// Regexp represents a Ruby Regexp.
type Regexp struct {
	Pattern  string
	Flags    string
	Compiled *regexp.Regexp
}

// NewRegexp creates a new Regexp object with compiled pattern.
func NewRegexp(pattern, flags string) (*Regexp, error) {
	// Convert Ruby regex flags to Go regex flags
	goPattern := pattern
	if strings.Contains(flags, "i") {
		goPattern = "(?i)" + goPattern
	}
	if strings.Contains(flags, "m") {
		// Ruby's m = Go's s (dot matches newline)
		goPattern = "(?s)" + goPattern
	}
	if strings.Contains(flags, "x") {
		// Extended mode - ignore whitespace (simplified)
		goPattern = "(?x)" + goPattern
	}

	compiled, err := regexp.Compile(goPattern)
	if err != nil {
		return nil, err
	}

	return &Regexp{
		Pattern:  pattern,
		Flags:    flags,
		Compiled: compiled,
	}, nil
}

func (r *Regexp) Type() Type      { return REGEXP_OBJ }
func (r *Regexp) Inspect() string { return "/" + r.Pattern + "/" + r.Flags }
func (r *Regexp) Class() *RubyClass { return RegexpClass }
func (r *Regexp) IsTruthy() bool  { return true }

// Match returns the match data for the string.
func (r *Regexp) Match(s string) []string {
	if r.Compiled == nil {
		return nil
	}
	return r.Compiled.FindStringSubmatch(s)
}

// MatchAll returns all matches for the string.
func (r *Regexp) MatchAll(s string) [][]string {
	if r.Compiled == nil {
		return nil
	}
	return r.Compiled.FindAllStringSubmatch(s, -1)
}

// ReplaceFirst replaces the first match.
func (r *Regexp) ReplaceFirst(s, replacement string) string {
	if r.Compiled == nil {
		return s
	}
	return r.Compiled.ReplaceAllStringFunc(s, func(match string) string {
		return replacement
	})
}

// ReplaceAll replaces all matches.
func (r *Regexp) ReplaceAll(s, replacement string) string {
	if r.Compiled == nil {
		return s
	}
	return r.Compiled.ReplaceAllString(s, replacement)
}

// Time represents a Ruby Time object.
type Time struct {
	Value time.Time
}

func (t *Time) Type() Type         { return TIME_OBJ }
func (t *Time) Inspect() string    { return t.Value.Format("2006-01-02 15:04:05 -0700") }
func (t *Time) Class() *RubyClass  { return nil } // Set dynamically
func (t *Time) IsTruthy() bool     { return true }

// Date represents a Ruby Date object.
type Date struct {
	Value time.Time
}

func (d *Date) Type() Type         { return DATE_OBJ }
func (d *Date) Inspect() string    { return d.Value.Format("2006-01-02") }
func (d *Date) Class() *RubyClass  { return nil } // Set dynamically
func (d *Date) IsTruthy() bool     { return true }

// ReturnValue wraps a return value.
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() Type      { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }
func (rv *ReturnValue) Class() *RubyClass { return nil }
func (rv *ReturnValue) IsTruthy() bool  { return rv.Value.IsTruthy() }

// BreakValue wraps a break value.
type BreakValue struct {
	Value Object
}

func (bv *BreakValue) Type() Type      { return BREAK_VALUE_OBJ }
func (bv *BreakValue) Inspect() string { return bv.Value.Inspect() }
func (bv *BreakValue) Class() *RubyClass { return nil }
func (bv *BreakValue) IsTruthy() bool  { return bv.Value.IsTruthy() }

// NextValue wraps a next value.
type NextValue struct {
	Value Object
}

func (nv *NextValue) Type() Type      { return NEXT_VALUE_OBJ }
func (nv *NextValue) Inspect() string { return nv.Value.Inspect() }
func (nv *NextValue) Class() *RubyClass { return nil }
func (nv *NextValue) IsTruthy() bool  { return nv.Value.IsTruthy() }

// Error represents a Ruby error.
type Error struct {
	Message   string
	Class_    *RubyClass
	Backtrace []string
}

func (e *Error) Type() Type      { return ERROR_OBJ }
func (e *Error) Inspect() string { return "ERROR: " + e.Message }
func (e *Error) Class() *RubyClass {
	if e.Class_ != nil {
		return e.Class_
	}
	return RuntimeErrorClass
}
func (e *Error) IsTruthy() bool { return true }

// Exception represents a Ruby exception.
type Exception struct {
	Message   string
	Class_    *RubyClass
	Backtrace []string
}

func (e *Exception) Type() Type      { return EXCEPTION_OBJ }
func (e *Exception) Inspect() string { return fmt.Sprintf("#<%s: %s>", e.Class_.Name, e.Message) }
func (e *Exception) Class() *RubyClass { return e.Class_ }
func (e *Exception) IsTruthy() bool  { return true }

// Proc represents a Ruby Proc/block.
type Proc struct {
	Parameters []*ast.BlockParameter
	Body       *ast.BlockBody
	Env        *Environment
}

func (p *Proc) Type() Type      { return PROC_OBJ }
func (p *Proc) Inspect() string { return "#<Proc>" }
func (p *Proc) Class() *RubyClass { return ProcClass }
func (p *Proc) IsTruthy() bool  { return true }

// Lambda represents a Ruby Lambda.
type Lambda struct {
	Parameters []*ast.BlockParameter
	Body       *ast.BlockBody
	Env        *Environment
}

func (l *Lambda) Type() Type      { return LAMBDA_OBJ }
func (l *Lambda) Inspect() string { return "#<Proc (lambda)>" }
func (l *Lambda) Class() *RubyClass { return ProcClass }
func (l *Lambda) IsTruthy() bool  { return true }

// Method represents a Ruby method.
type Method struct {
	Name       string
	Parameters []*ast.MethodParameter
	Body       *ast.BlockBody
	Env        *Environment
	Receiver   Object
}

func (m *Method) Type() Type      { return METHOD_OBJ }
func (m *Method) Inspect() string { return fmt.Sprintf("#<Method: %s>", m.Name) }
func (m *Method) Class() *RubyClass { return MethodClass }
func (m *Method) IsTruthy() bool  { return true }

// BuiltinFunction is a Go function callable from Ruby.
type BuiltinFunction func(receiver Object, env *Environment, args ...Object) Object

// Builtin represents a built-in method.
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() Type      { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string { return fmt.Sprintf("#<Builtin: %s>", b.Name) }
func (b *Builtin) Class() *RubyClass { return nil }
func (b *Builtin) IsTruthy() bool  { return true }

// RubyClass represents a Ruby class.
type RubyClass struct {
	Name            string
	Superclass      *RubyClass
	Methods         map[string]Object // Method objects (Method or Builtin)
	ClassMethods    map[string]Object // Class methods
	Constants       map[string]Object
	IncludedModules []*RubyModule
	StructMembers   []string // For Struct subclasses
}

func (c *RubyClass) Type() Type      { return CLASS_OBJ }
func (c *RubyClass) Inspect() string { return c.Name }
func (c *RubyClass) Class() *RubyClass { return ClassClass }
func (c *RubyClass) IsTruthy() bool  { return true }

// LookupMethod looks up a method in the class hierarchy.
func (c *RubyClass) LookupMethod(name string) (Object, bool) {
	// Check this class
	if method, ok := c.Methods[name]; ok {
		return method, true
	}
	// Check included modules
	for i := len(c.IncludedModules) - 1; i >= 0; i-- {
		if method, ok := c.IncludedModules[i].Methods[name]; ok {
			return method, true
		}
	}
	// Check superclass
	if c.Superclass != nil {
		return c.Superclass.LookupMethod(name)
	}
	return nil, false
}

// LookupClassMethod looks up a class method.
func (c *RubyClass) LookupClassMethod(name string) (Object, bool) {
	if method, ok := c.ClassMethods[name]; ok {
		return method, true
	}
	if c.Superclass != nil {
		return c.Superclass.LookupClassMethod(name)
	}
	return nil, false
}

// RubyModule represents a Ruby module.
type RubyModule struct {
	Name      string
	Methods   map[string]Object
	Constants map[string]Object
}

func (m *RubyModule) Type() Type      { return MODULE_OBJ }
func (m *RubyModule) Inspect() string { return m.Name }
func (m *RubyModule) Class() *RubyClass { return ModuleClass }
func (m *RubyModule) IsTruthy() bool  { return true }

// Instance represents an instance of a Ruby class.
type Instance struct {
	Class_            *RubyClass
	InstanceVariables map[string]Object
}

func (i *Instance) Type() Type      { return INSTANCE_OBJ }
func (i *Instance) Inspect() string { return fmt.Sprintf("#<%s:0x%p>", i.Class_.Name, i) }
func (i *Instance) Class() *RubyClass { return i.Class_ }
func (i *Instance) IsTruthy() bool  { return true }

// GetInstanceVariable gets an instance variable.
func (i *Instance) GetInstanceVariable(name string) Object {
	if val, ok := i.InstanceVariables[name]; ok {
		return val
	}
	return NIL
}

// SetInstanceVariable sets an instance variable.
func (i *Instance) SetInstanceVariable(name string, val Object) {
	i.InstanceVariables[name] = val
}

// Singleton objects
var (
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NIL   = &Nil{}
)

// NativeToBool converts a Go bool to a Ruby Boolean.
func NativeToBool(b bool) *Boolean {
	if b {
		return TRUE
	}
	return FALSE
}

// Built-in classes (initialized in init)
var (
	ObjectClass       *RubyClass
	ClassClass        *RubyClass
	ModuleClass       *RubyClass
	BasicObjectClass  *RubyClass
	IntegerClass      *RubyClass
	FloatClass        *RubyClass
	StringClass       *RubyClass
	SymbolClass       *RubyClass
	ArrayClass        *RubyClass
	HashClass         *RubyClass
	RangeClass        *RubyClass
	RegexpClass       *RubyClass
	ProcClass         *RubyClass
	MethodClass       *RubyClass
	TrueClass         *RubyClass
	FalseClass        *RubyClass
	NilClass          *RubyClass
	ExceptionClass    *RubyClass
	StandardErrorClass *RubyClass
	RuntimeErrorClass *RubyClass
	ArgumentErrorClass *RubyClass
	TypeError         *RubyClass
	NameErrorClass    *RubyClass
	NoMethodErrorClass *RubyClass
	IOClass           *RubyClass
	KernelModule      *RubyModule
	ComparableModule  *RubyModule
	EnumerableModule  *RubyModule
)

func init() {
	// Initialize class hierarchy
	BasicObjectClass = &RubyClass{
		Name:         "BasicObject",
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	ObjectClass = &RubyClass{
		Name:         "Object",
		Superclass:   BasicObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	ModuleClass = &RubyClass{
		Name:         "Module",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	ClassClass = &RubyClass{
		Name:         "Class",
		Superclass:   ModuleClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Numeric classes
	IntegerClass = &RubyClass{
		Name:         "Integer",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	FloatClass = &RubyClass{
		Name:         "Float",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// String and Symbol
	StringClass = &RubyClass{
		Name:         "String",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	SymbolClass = &RubyClass{
		Name:         "Symbol",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Collections
	ArrayClass = &RubyClass{
		Name:         "Array",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	HashClass = &RubyClass{
		Name:         "Hash",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	RangeClass = &RubyClass{
		Name:         "Range",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	RegexpClass = &RubyClass{
		Name:         "Regexp",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Proc and Method
	ProcClass = &RubyClass{
		Name:         "Proc",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	MethodClass = &RubyClass{
		Name:         "Method",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Boolean and Nil classes
	TrueClass = &RubyClass{
		Name:         "TrueClass",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	FalseClass = &RubyClass{
		Name:         "FalseClass",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	NilClass = &RubyClass{
		Name:         "NilClass",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Exception classes
	ExceptionClass = &RubyClass{
		Name:         "Exception",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	StandardErrorClass = &RubyClass{
		Name:         "StandardError",
		Superclass:   ExceptionClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	RuntimeErrorClass = &RubyClass{
		Name:         "RuntimeError",
		Superclass:   StandardErrorClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	ArgumentErrorClass = &RubyClass{
		Name:         "ArgumentError",
		Superclass:   StandardErrorClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	TypeError = &RubyClass{
		Name:         "TypeError",
		Superclass:   StandardErrorClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	NameErrorClass = &RubyClass{
		Name:         "NameError",
		Superclass:   StandardErrorClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	NoMethodErrorClass = &RubyClass{
		Name:         "NoMethodError",
		Superclass:   NameErrorClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	IOClass = &RubyClass{
		Name:         "IO",
		Superclass:   ObjectClass,
		Methods:      make(map[string]Object),
		ClassMethods: make(map[string]Object),
		Constants:    make(map[string]Object),
	}

	// Modules
	KernelModule = &RubyModule{
		Name:      "Kernel",
		Methods:   make(map[string]Object),
		Constants: make(map[string]Object),
	}

	ComparableModule = &RubyModule{
		Name:      "Comparable",
		Methods:   make(map[string]Object),
		Constants: make(map[string]Object),
	}

	EnumerableModule = &RubyModule{
		Name:      "Enumerable",
		Methods:   make(map[string]Object),
		Constants: make(map[string]Object),
	}

	// Include Kernel in Object
	ObjectClass.IncludedModules = append(ObjectClass.IncludedModules, KernelModule)
}
