// Package evaluator implements the Ruby interpreter.
package evaluator

import (
	"fmt"
	"math"

	"github.com/alexisbouchez/rubylexer/ast"
	"github.com/alexisbouchez/rubylexer/object"
)

// Eval evaluates an AST node.
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Program
	case *ast.Program:
		return evalProgram(node, env)

	// Statements
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.MethodDefinition:
		return evalMethodDefinition(node, env)

	case *ast.ClassDefinition:
		return evalClassDefinition(node, env)

	case *ast.ModuleDefinition:
		return evalModuleDefinition(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.BreakStatement:
		var val object.Object = object.NIL
		if node.Value != nil {
			val = Eval(node.Value, env)
			if isError(val) {
				return val
			}
		}
		return &object.BreakValue{Value: val}

	case *ast.NextStatement:
		var val object.Object = object.NIL
		if node.Value != nil {
			val = Eval(node.Value, env)
			if isError(val) {
				return val
			}
		}
		return &object.NextValue{Value: val}

	case *ast.RedoStatement:
		return newError("redo not yet supported")

	case *ast.RetryStatement:
		return newError("retry not yet supported")

	// Literals
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.InterpolatedString:
		return evalInterpolatedString(node, env)

	case *ast.SymbolLiteral:
		return &object.Symbol{Value: node.Value}

	case *ast.BooleanLiteral:
		return object.NativeToBool(node.Value)

	case *ast.NilLiteral:
		return object.NIL

	case *ast.SelfExpression:
		self := env.Self()
		if self == nil {
			return object.NIL
		}
		return self

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.RangeLiteral:
		return evalRangeLiteral(node, env)

	case *ast.RegexpLiteral:
		re, err := object.NewRegexp(node.Value, node.Flags)
		if err != nil {
			return newError("invalid regular expression: %s", err)
		}
		return re

	// Variables
	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.Constant:
		return evalConstant(node, env)

	case *ast.InstanceVariable:
		return evalInstanceVariable(node, env)

	case *ast.ClassVariable:
		return evalClassVariable(node, env)

	case *ast.GlobalVariable:
		return evalGlobalVariable(node, env)

	case *ast.ScopedConstant:
		return evalScopedConstant(node, env)

	// Expressions
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.AssignmentExpression:
		return evalAssignment(node, env)

	case *ast.OpAssignmentExpression:
		return evalOpAssignment(node, env)

	case *ast.IndexExpression:
		return evalIndexExpression(node, env)

	case *ast.MethodCall:
		return evalMethodCall(node, env)

	// Control Flow
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.TernaryExpression:
		return evalTernaryExpression(node, env)

	case *ast.ModifierExpression:
		return evalModifierExpression(node, env)

	case *ast.CaseExpression:
		return evalCaseExpression(node, env)

	case *ast.WhileExpression:
		return evalWhileExpression(node, env)

	case *ast.ForExpression:
		return evalForExpression(node, env)

	case *ast.BeginExpression:
		return evalBeginExpression(node, env)

	// Other
	case *ast.YieldExpression:
		return evalYieldExpression(node, env)

	case *ast.Lambda:
		return evalLambda(node, env)

	case *ast.NotExpression:
		val := Eval(node.Expression, env)
		if isError(val) {
			return val
		}
		return object.NativeToBool(!isTruthy(val))

	case *ast.AndExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		if !isTruthy(left) {
			return left
		}
		return Eval(node.Right, env)

	case *ast.OrExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		if isTruthy(left) {
			return left
		}
		return Eval(node.Right, env)

	case *ast.RescueModifier:
		result := Eval(node.Body, env)
		if isError(result) {
			return Eval(node.Rescue, env)
		}
		return result

	case *ast.SplatExpression:
		return evalSplatExpression(node, env)

	case *ast.DoubleSplatExpression:
		return evalDoubleSplatExpression(node, env)

	case *ast.DefinedExpression:
		return evalDefinedExpression(node, env)

	default:
		return newError("unknown node type: %T", node)
	}
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object = object.NIL

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockBody(body *ast.BlockBody, env *object.Environment) object.Object {
	var result object.Object = object.NIL

	for _, statement := range body.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ ||
				rt == object.BREAK_VALUE_OBJ ||
				rt == object.NEXT_VALUE_OBJ ||
				rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// Literal evaluation

func evalInterpolatedString(node *ast.InterpolatedString, env *object.Environment) object.Object {
	var result string

	for _, part := range node.Parts {
		val := Eval(part, env)
		if isError(val) {
			return val
		}
		result += objectToString(val)
	}

	return &object.String{Value: result}
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	order := make([]object.HashKey, 0, len(node.Pairs))

	for _, keyNode := range node.Order {
		valueNode := node.Pairs[keyNode]

		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
		order = append(order, hashed)
	}

	return &object.Hash{Pairs: pairs, Order: order, IsKeywordArgs: node.IsKeywordArgs}
}

func evalRangeLiteral(node *ast.RangeLiteral, env *object.Environment) object.Object {
	start := Eval(node.Start, env)
	if isError(start) {
		return start
	}

	end := Eval(node.End, env)
	if isError(end) {
		return end
	}

	return &object.Range{Start: start, End: end, Exclusive: node.Exclusive}
}

// Variable evaluation

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// Check if it's a method call with no arguments
	self := env.Self()
	if self != nil {
		if class := self.Class(); class != nil {
			if method, ok := class.LookupMethod(node.Value); ok {
				return applyMethod(method, self, []object.Object{}, nil, env)
			}
		}
	}

	// Check Kernel methods
	if builtin, ok := object.KernelModule.Methods[node.Value]; ok {
		return applyMethod(builtin, self, []object.Object{}, nil, env)
	}

	return newError("undefined local variable or method `%s'", node.Value)
}

func evalConstant(node *ast.Constant, env *object.Environment) object.Object {
	if val, ok := env.GetConstant(node.Value); ok {
		return val
	}

	// Check built-in classes
	switch node.Value {
	case "Object":
		return object.ObjectClass
	case "Class":
		return object.ClassClass
	case "Module":
		return object.ModuleClass
	case "Integer":
		return object.IntegerClass
	case "Float":
		return object.FloatClass
	case "String":
		return object.StringClass
	case "Symbol":
		return object.SymbolClass
	case "Array":
		return object.ArrayClass
	case "Hash":
		return object.HashClass
	case "Range":
		return object.RangeClass
	case "Regexp":
		return object.RegexpClass
	case "Proc":
		return object.ProcClass
	case "TrueClass":
		return object.TrueClass
	case "FalseClass":
		return object.FalseClass
	case "NilClass":
		return object.NilClass
	case "Exception":
		return object.ExceptionClass
	case "StandardError":
		return object.StandardErrorClass
	case "RuntimeError":
		return object.RuntimeErrorClass
	case "ArgumentError":
		return object.ArgumentErrorClass
	case "TypeError":
		return object.TypeError
	case "NameError":
		return object.NameErrorClass
	case "NoMethodError":
		return object.NoMethodErrorClass
	case "Kernel":
		return object.KernelModule
	case "Comparable":
		return object.ComparableModule
	case "Enumerable":
		return object.EnumerableModule
	case "File":
		return FileClass
	case "Dir":
		return DirClass
	}

	return newError("uninitialized constant %s", node.Value)
}

func evalInstanceVariable(node *ast.InstanceVariable, env *object.Environment) object.Object {
	self := env.Self()
	if self == nil {
		return object.NIL
	}

	if instance, ok := self.(*object.Instance); ok {
		return instance.GetInstanceVariable(node.Name)
	}

	return object.NIL
}

func evalClassVariable(node *ast.ClassVariable, env *object.Environment) object.Object {
	// Class variables are complex - simplified implementation
	if val, ok := env.Get(node.Name); ok {
		return val
	}
	return object.NIL
}

var globalVariables = make(map[string]object.Object)

func evalGlobalVariable(node *ast.GlobalVariable, env *object.Environment) object.Object {
	if val, ok := globalVariables[node.Name]; ok {
		return val
	}
	return object.NIL
}

func evalScopedConstant(node *ast.ScopedConstant, env *object.Environment) object.Object {
	if node.Left == nil {
		// Top-level constant (::Foo)
		return evalConstant(&ast.Constant{Token: node.Token, Value: node.Name}, env)
	}

	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	switch obj := left.(type) {
	case *object.RubyClass:
		if val, ok := obj.Constants[node.Name]; ok {
			return val
		}
	case *object.RubyModule:
		if val, ok := obj.Constants[node.Name]; ok {
			return val
		}
	}

	return newError("uninitialized constant %s::%s", left.Inspect(), node.Name)
}

// Prefix expression

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperator(right)
	case "+":
		return evalPlusPrefixOperator(right)
	case "~":
		return evalTildeOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperator(right object.Object) object.Object {
	return object.NativeToBool(!isTruthy(right))
}

func evalMinusPrefixOperator(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return newError("undefined method `-@' for %s", right.Type())
	}
}

func evalPlusPrefixOperator(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return obj
	case *object.Float:
		return obj
	default:
		return newError("undefined method `+@' for %s", right.Type())
	}
}

func evalTildeOperator(right object.Object) object.Object {
	if obj, ok := right.(*object.Integer); ok {
		return &object.Integer{Value: ^obj.Value}
	}
	return newError("undefined method `~' for %s", right.Type())
}

// Infix expressions

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalStringIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.REGEXP_OBJ:
		return evalStringRegexpInfixExpression(operator, left, right)
	case left.Type() == object.REGEXP_OBJ && right.Type() == object.STRING_OBJ:
		return evalRegexpStringInfixExpression(operator, left, right)
	case left.Type() == object.ARRAY_OBJ:
		return evalArrayInfixExpression(operator, left, right)
	case operator == "==":
		return object.NativeToBool(objectsEqual(left, right))
	case operator == "!=":
		return object.NativeToBool(!objectsEqual(left, right))
	case operator == "===":
		return evalCaseEquality(left, right)
	case operator == "&&":
		if !isTruthy(left) {
			return left
		}
		return right
	case operator == "||":
		if isTruthy(left) {
			return left
		}
		return right
	default:
		return newError("undefined method `%s' for %s", operator, left.Type())
	}
}

func evalStringRegexpInfixExpression(operator string, left, right object.Object) object.Object {
	str := left.(*object.String).Value
	re := right.(*object.Regexp)

	switch operator {
	case "=~":
		if re.Compiled == nil {
			return object.NIL
		}
		loc := re.Compiled.FindStringIndex(str)
		if loc == nil {
			return object.NIL
		}
		return &object.Integer{Value: int64(loc[0])}
	case "!~":
		if re.Compiled == nil {
			return object.TRUE
		}
		return object.NativeToBool(!re.Compiled.MatchString(str))
	default:
		return newError("undefined method `%s' for String", operator)
	}
}

func evalRegexpStringInfixExpression(operator string, left, right object.Object) object.Object {
	re := left.(*object.Regexp)
	str := right.(*object.String).Value

	switch operator {
	case "=~":
		if re.Compiled == nil {
			return object.NIL
		}
		loc := re.Compiled.FindStringIndex(str)
		if loc == nil {
			return object.NIL
		}
		return &object.Integer{Value: int64(loc[0])}
	case "!~":
		if re.Compiled == nil {
			return object.TRUE
		}
		return object.NativeToBool(!re.Compiled.MatchString(str))
	case "===":
		if re.Compiled == nil {
			return object.FALSE
		}
		return object.NativeToBool(re.Compiled.MatchString(str))
	default:
		return newError("undefined method `%s' for Regexp", operator)
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("ZeroDivisionError: divided by 0")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("ZeroDivisionError: divided by 0")
		}
		return &object.Integer{Value: leftVal % rightVal}
	case "**":
		return &object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "<":
		return object.NativeToBool(leftVal < rightVal)
	case ">":
		return object.NativeToBool(leftVal > rightVal)
	case "<=":
		return object.NativeToBool(leftVal <= rightVal)
	case ">=":
		return object.NativeToBool(leftVal >= rightVal)
	case "==":
		return object.NativeToBool(leftVal == rightVal)
	case "!=":
		return object.NativeToBool(leftVal != rightVal)
	case "<=>":
		if leftVal < rightVal {
			return &object.Integer{Value: -1}
		} else if leftVal > rightVal {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	case "&":
		return &object.Integer{Value: leftVal & rightVal}
	case "|":
		return &object.Integer{Value: leftVal | rightVal}
	case "^":
		return &object.Integer{Value: leftVal ^ rightVal}
	case "<<":
		return &object.Integer{Value: leftVal << uint(rightVal)}
	case ">>":
		return &object.Integer{Value: leftVal >> uint(rightVal)}
	default:
		return newError("undefined method `%s' for Integer", operator)
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	var leftVal, rightVal float64

	switch l := left.(type) {
	case *object.Integer:
		leftVal = float64(l.Value)
	case *object.Float:
		leftVal = l.Value
	}

	switch r := right.(type) {
	case *object.Integer:
		rightVal = float64(r.Value)
	case *object.Float:
		rightVal = r.Value
	}

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("ZeroDivisionError: divided by 0")
		}
		return &object.Float{Value: leftVal / rightVal}
	case "%":
		return &object.Float{Value: math.Mod(leftVal, rightVal)}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "<":
		return object.NativeToBool(leftVal < rightVal)
	case ">":
		return object.NativeToBool(leftVal > rightVal)
	case "<=":
		return object.NativeToBool(leftVal <= rightVal)
	case ">=":
		return object.NativeToBool(leftVal >= rightVal)
	case "==":
		return object.NativeToBool(leftVal == rightVal)
	case "!=":
		return object.NativeToBool(leftVal != rightVal)
	case "<=>":
		if leftVal < rightVal {
			return &object.Integer{Value: -1}
		} else if leftVal > rightVal {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	default:
		return newError("undefined method `%s' for Float", operator)
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return object.NativeToBool(leftVal == rightVal)
	case "!=":
		return object.NativeToBool(leftVal != rightVal)
	case "<":
		return object.NativeToBool(leftVal < rightVal)
	case ">":
		return object.NativeToBool(leftVal > rightVal)
	case "<=":
		return object.NativeToBool(leftVal <= rightVal)
	case ">=":
		return object.NativeToBool(leftVal >= rightVal)
	case "<=>":
		if leftVal < rightVal {
			return &object.Integer{Value: -1}
		} else if leftVal > rightVal {
			return &object.Integer{Value: 1}
		}
		return &object.Integer{Value: 0}
	default:
		return newError("undefined method `%s' for String", operator)
	}
}

func evalStringIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	str := left.(*object.String).Value
	n := right.(*object.Integer).Value

	switch operator {
	case "*":
		if n < 0 {
			return newError("ArgumentError: negative argument")
		}
		var result string
		for i := int64(0); i < n; i++ {
			result += str
		}
		return &object.String{Value: result}
	default:
		return newError("undefined method `%s' for String", operator)
	}
}

func evalArrayInfixExpression(operator string, left, right object.Object) object.Object {
	leftArr := left.(*object.Array)

	switch operator {
	case "+":
		if rightArr, ok := right.(*object.Array); ok {
			elements := make([]object.Object, 0, len(leftArr.Elements)+len(rightArr.Elements))
			elements = append(elements, leftArr.Elements...)
			elements = append(elements, rightArr.Elements...)
			return &object.Array{Elements: elements}
		}
	case "*":
		if n, ok := right.(*object.Integer); ok {
			elements := make([]object.Object, 0, len(leftArr.Elements)*int(n.Value))
			for i := int64(0); i < n.Value; i++ {
				elements = append(elements, leftArr.Elements...)
			}
			return &object.Array{Elements: elements}
		}
	case "<<":
		leftArr.Elements = append(leftArr.Elements, right)
		return leftArr
	}

	return newError("undefined method `%s' for Array", operator)
}

func evalCaseEquality(left, right object.Object) object.Object {
	// === operator behavior depends on the left operand
	switch l := left.(type) {
	case *object.RubyClass:
		// Class === obj checks if obj is an instance of the class
		if right.Class() == l {
			return object.TRUE
		}
		// Check superclasses
		class := right.Class()
		for class != nil {
			if class == l {
				return object.TRUE
			}
			class = class.Superclass
		}
		return object.FALSE
	case *object.Range:
		return evalRangeIncludes(l, right)
	case *object.Regexp:
		if str, ok := right.(*object.String); ok {
			// Simplified regex matching
			return object.NativeToBool(simpleMatch(l.Pattern, str.Value))
		}
	}

	return object.NativeToBool(objectsEqual(left, right))
}

// Assignment

func evalAssignment(node *ast.AssignmentExpression, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	switch target := node.Left.(type) {
	case *ast.Identifier:
		return env.Set(target.Value, val)
	case *ast.InstanceVariable:
		return setInstanceVariable(target.Name, val, env)
	case *ast.ClassVariable:
		return env.Set(target.Name, val)
	case *ast.GlobalVariable:
		globalVariables[target.Name] = val
		return val
	case *ast.Constant:
		return env.SetConstant(target.Value, val)
	case *ast.IndexExpression:
		return evalIndexAssignment(target, val, env)
	case *ast.MethodCall:
		// Handle setter method calls: obj.attr = value -> obj.attr=(value)
		receiver := Eval(target.Receiver, env)
		if isError(receiver) {
			return receiver
		}
		setterName := target.Method + "="
		return callMethod(receiver, setterName, []object.Object{val}, nil, env)
	default:
		return newError("invalid assignment target: %T", node.Left)
	}
}

func evalOpAssignment(node *ast.OpAssignmentExpression, env *object.Environment) object.Object {
	// Get current value
	var currentVal object.Object
	switch target := node.Left.(type) {
	case *ast.Identifier:
		currentVal, _ = env.Get(target.Value)
	case *ast.InstanceVariable:
		currentVal = evalInstanceVariable(target, env)
	case *ast.IndexExpression:
		currentVal = evalIndexExpression(target, env)
	default:
		return newError("invalid assignment target: %T", node.Left)
	}

	if currentVal == nil {
		currentVal = object.NIL
	}

	// Handle ||= and &&= specially
	switch node.Operator {
	case "||=":
		if isTruthy(currentVal) {
			return currentVal
		}
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return evalAssignment(&ast.AssignmentExpression{
			Token: node.Token,
			Left:  node.Left,
			Value: node.Value,
		}, env)
	case "&&=":
		if !isTruthy(currentVal) {
			return currentVal
		}
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return evalAssignment(&ast.AssignmentExpression{
			Token: node.Token,
			Left:  node.Left,
			Value: node.Value,
		}, env)
	}

	// Evaluate right side
	rightVal := Eval(node.Value, env)
	if isError(rightVal) {
		return rightVal
	}

	// Apply operator
	var op string
	switch node.Operator {
	case "+=":
		op = "+"
	case "-=":
		op = "-"
	case "*=":
		op = "*"
	case "/=":
		op = "/"
	case "%=":
		op = "%"
	case "**=":
		op = "**"
	case "&=":
		op = "&"
	case "|=":
		op = "|"
	case "^=":
		op = "^"
	case "<<=":
		op = "<<"
	case ">>=":
		op = ">>"
	}

	result := evalInfixExpression(op, currentVal, rightVal)
	if isError(result) {
		return result
	}

	// Assign result
	switch target := node.Left.(type) {
	case *ast.Identifier:
		return env.Set(target.Value, result)
	case *ast.InstanceVariable:
		return setInstanceVariable(target.Name, result, env)
	case *ast.IndexExpression:
		return evalIndexAssignment(target, result, env)
	}

	return result
}

func setInstanceVariable(name string, val object.Object, env *object.Environment) object.Object {
	self := env.Self()
	if self == nil {
		return newError("cannot set instance variable outside of object")
	}

	if instance, ok := self.(*object.Instance); ok {
		instance.SetInstanceVariable(name, val)
		return val
	}

	return newError("cannot set instance variable on %s", self.Type())
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	return evalIndex(left, index)
}

func evalIndex(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndex(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndex(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.RANGE_OBJ:
		return evalStringRangeIndex(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndex(array, index object.Object) object.Object {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arr.Elements))

	if idx < 0 {
		idx = max + idx
	}

	if idx < 0 || idx >= max {
		return object.NIL
	}

	return arr.Elements[idx]
}

func evalHashIndex(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return object.NIL
	}

	return pair.Value
}

func evalStringIndex(str, index object.Object) object.Object {
	s := str.(*object.String)
	idx := index.(*object.Integer).Value
	max := int64(len(s.Value))

	if idx < 0 {
		idx = max + idx
	}

	if idx < 0 || idx >= max {
		return object.NIL
	}

	return &object.String{Value: string(s.Value[idx])}
}

func evalStringRangeIndex(str, index object.Object) object.Object {
	s := str.(*object.String)
	r := index.(*object.Range)

	startObj, ok := r.Start.(*object.Integer)
	if !ok {
		return newError("no implicit conversion of %s into Integer", r.Start.Type())
	}
	endObj, ok := r.End.(*object.Integer)
	if !ok {
		return newError("no implicit conversion of %s into Integer", r.End.Type())
	}

	start := startObj.Value
	end := endObj.Value
	max := int64(len(s.Value))

	if start < 0 {
		start = max + start
	}
	if end < 0 {
		end = max + end
	}
	if !r.Exclusive {
		end++
	}

	if start < 0 || start > max {
		return object.NIL
	}
	if end > max {
		end = max
	}
	if start > end {
		return &object.String{Value: ""}
	}

	return &object.String{Value: s.Value[start:end]}
}

func evalIndexAssignment(node *ast.IndexExpression, val object.Object, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	switch obj := left.(type) {
	case *object.Array:
		idx := index.(*object.Integer).Value
		if idx < 0 {
			idx = int64(len(obj.Elements)) + idx
		}
		if idx >= 0 && idx < int64(len(obj.Elements)) {
			obj.Elements[idx] = val
		} else if idx == int64(len(obj.Elements)) {
			obj.Elements = append(obj.Elements, val)
		} else {
			// Fill with nil
			for int64(len(obj.Elements)) <= idx {
				obj.Elements = append(obj.Elements, object.NIL)
			}
			obj.Elements[idx] = val
		}
		return val
	case *object.Hash:
		key, ok := index.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", index.Type())
		}
		hashed := key.HashKey()
		obj.Pairs[hashed] = object.HashPair{Key: index, Value: val}
		obj.Order = append(obj.Order, hashed)
		return val
	default:
		return newError("index assignment not supported: %s", left.Type())
	}
}

// Method calls

func evalMethodCall(node *ast.MethodCall, env *object.Environment) object.Object {
	var receiver object.Object

	if node.Receiver != nil {
		receiver = Eval(node.Receiver, env)
		if isError(receiver) {
			return receiver
		}
		// Handle safe navigation
		if node.SafeNav && receiver == object.NIL {
			return object.NIL
		}
	} else {
		receiver = env.Self()
		if receiver == nil {
			receiver = object.NIL
		}
	}

	// Evaluate arguments
	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	// Check if the last argument is a hash with symbol keys (keyword arguments)
	if len(args) > 0 {
		if hash, ok := args[len(args)-1].(*object.Hash); ok {
			// Check if all keys are symbols
			allSymbols := true
			for _, key := range hash.Order {
				pair := hash.Pairs[key]
				if _, ok := pair.Key.(*object.Symbol); !ok {
					allSymbols = false
					break
				}
			}
			if allSymbols && len(hash.Pairs) > 0 {
				hash.IsKeywordArgs = true
			}
		}
	}

	// Create block if present
	var block *object.Proc
	if node.Block != nil {
		block = &object.Proc{
			Parameters: node.Block.Parameters,
			Body:       node.Block.Body,
			Env:        env,
		}
	}

	return callMethod(receiver, node.Method, args, block, env)
}

func callMethod(receiver object.Object, methodName string, args []object.Object, block *object.Proc, env *object.Environment) object.Object {
	// Check if receiver is a class (class method call)
	if class, ok := receiver.(*object.RubyClass); ok {
		if method, ok := class.LookupClassMethod(methodName); ok {
			return applyMethod(method, receiver, args, block, env)
		}
		// Check for 'new' method
		if methodName == "new" {
			return createInstance(class, args, block, env)
		}
	}

	// Look up instance method
	if class := receiver.Class(); class != nil {
		if method, ok := class.LookupMethod(methodName); ok {
			return applyMethod(method, receiver, args, block, env)
		}
	}

	// Check built-in methods
	if builtin := getBuiltinMethod(receiver, methodName); builtin != nil {
		// Create a new environment with the block set
		callEnv := object.NewEnclosedEnvironment(env)
		callEnv.SetSelf(receiver)
		if block != nil {
			callEnv.SetBlock(block)
		}
		return builtin.Fn(receiver, callEnv, args...)
	}

	// Check for method_missing (but not if we're already calling method_missing)
	if methodName != "method_missing" {
		if class := receiver.Class(); class != nil {
			if mmMethod, ok := class.LookupMethod("method_missing"); ok {
				// Prepend method name as first argument
				mmArgs := make([]object.Object, 0, len(args)+1)
				mmArgs = append(mmArgs, &object.Symbol{Value: methodName})
				mmArgs = append(mmArgs, args...)
				return applyMethod(mmMethod, receiver, mmArgs, block, env)
			}
		}
	}

	return newError("undefined method `%s' for %s", methodName, receiver.Inspect())
}

func applyMethod(method object.Object, receiver object.Object, args []object.Object, block *object.Proc, env *object.Environment) object.Object {
	switch m := method.(type) {
	case *object.Method:
		extendedEnv := object.NewEnclosedEnvironment(m.Env)
		extendedEnv.SetSelf(receiver)
		if block != nil {
			extendedEnv.SetBlock(block)
		}

		// Separate positional and keyword arguments
		var positionalArgs []object.Object
		var kwArgs *object.Hash

		for _, arg := range args {
			if hash, ok := arg.(*object.Hash); ok && hash.IsKeywordArgs {
				kwArgs = hash
			} else {
				positionalArgs = append(positionalArgs, arg)
			}
		}

		// Bind parameters
		argIdx := 0
		for _, param := range m.Parameters {
			if param.Splat {
				// Collect remaining positional args
				remaining := []object.Object{}
				for argIdx < len(positionalArgs) {
					remaining = append(remaining, positionalArgs[argIdx])
					argIdx++
				}
				extendedEnv.Set(param.Name, &object.Array{Elements: remaining})
			} else if param.DSplat {
				// Collect remaining keyword args
				if kwArgs != nil {
					extendedEnv.Set(param.Name, kwArgs)
				} else {
					extendedEnv.Set(param.Name, &object.Hash{
						Pairs: make(map[object.HashKey]object.HashPair),
					})
				}
			} else if param.Block {
				// Block parameter handled separately
			} else if param.KeywordOnly {
				// Keyword-only parameter
				if kwArgs != nil {
					key := object.Symbol{Value: param.Name}
					if pair, ok := kwArgs.Pairs[key.HashKey()]; ok {
						extendedEnv.Set(param.Name, pair.Value)
					} else if param.Default != nil {
						defaultVal := Eval(param.Default, extendedEnv)
						extendedEnv.Set(param.Name, defaultVal)
					} else {
						return newError("missing keyword: %s", param.Name)
					}
				} else if param.Default != nil {
					defaultVal := Eval(param.Default, extendedEnv)
					extendedEnv.Set(param.Name, defaultVal)
				} else {
					return newError("missing keyword: %s", param.Name)
				}
			} else {
				// Regular positional parameter
				if argIdx < len(positionalArgs) {
					extendedEnv.Set(param.Name, positionalArgs[argIdx])
					argIdx++
				} else if param.Default != nil {
					defaultVal := Eval(param.Default, extendedEnv)
					extendedEnv.Set(param.Name, defaultVal)
				} else {
					extendedEnv.Set(param.Name, object.NIL)
				}
			}
		}

		result := evalBlockBody(m.Body, extendedEnv)
		return unwrapReturnValue(result)

	case *object.Builtin:
		return m.Fn(receiver, env, args...)

	default:
		return newError("not a method: %s", method.Type())
	}
}

func createInstance(class *object.RubyClass, args []object.Object, block *object.Proc, env *object.Environment) object.Object {
	instance := &object.Instance{
		Class_:            class,
		InstanceVariables: make(map[string]object.Object),
	}

	// Call initialize if it exists
	if method, ok := class.LookupMethod("initialize"); ok {
		instanceEnv := object.NewEnclosedEnvironment(env)
		instanceEnv.SetSelf(instance)
		applyMethod(method, instance, args, block, instanceEnv)
	}

	return instance
}

// Control flow

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	conditionMet := isTruthy(condition)
	if node.Unless {
		conditionMet = !conditionMet
	}

	if conditionMet {
		return evalBlockBody(node.Consequence, env)
	} else if node.Alternative != nil {
		return evalIfExpression(node.Alternative, env)
	} else if node.ElseBody != nil {
		return evalBlockBody(node.ElseBody, env)
	}

	return object.NIL
}

func evalTernaryExpression(node *ast.TernaryExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	}
	return Eval(node.Alternative, env)
}

func evalModifierExpression(node *ast.ModifierExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	conditionMet := isTruthy(condition)
	switch node.Modifier {
	case "unless", "until":
		conditionMet = !conditionMet
	}

	if conditionMet {
		return Eval(node.Body, env)
	}

	return object.NIL
}

func evalCaseExpression(node *ast.CaseExpression, env *object.Environment) object.Object {
	var subject object.Object
	if node.Subject != nil {
		subject = Eval(node.Subject, env)
		if isError(subject) {
			return subject
		}
	}

	for _, when := range node.Whens {
		for _, condExpr := range when.Conditions {
			cond := Eval(condExpr, env)
			if isError(cond) {
				return cond
			}

			matched := false
			if subject != nil {
				// Use === for matching
				result := evalCaseEquality(cond, subject)
				matched = isTruthy(result)
			} else {
				matched = isTruthy(cond)
			}

			if matched {
				return evalBlockBody(when.Body, env)
			}
		}
	}

	if node.Else != nil {
		return evalBlockBody(node.Else, env)
	}

	return object.NIL
}

func evalWhileExpression(node *ast.WhileExpression, env *object.Environment) object.Object {
	var result object.Object = object.NIL

	for {
		condition := Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}

		conditionMet := isTruthy(condition)
		if node.Until {
			conditionMet = !conditionMet
		}

		if !conditionMet {
			break
		}

		result = evalBlockBody(node.Body, env)

		if rv, ok := result.(*object.ReturnValue); ok {
			return rv
		}
		if _, ok := result.(*object.BreakValue); ok {
			return object.NIL
		}
		if _, ok := result.(*object.NextValue); ok {
			continue
		}
		if isError(result) {
			return result
		}
	}

	return result
}

func evalForExpression(node *ast.ForExpression, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	var elements []object.Object

	switch iter := iterable.(type) {
	case *object.Array:
		elements = iter.Elements
	case *object.Range:
		elements = expandRange(iter)
	default:
		return newError("cannot iterate over %s", iterable.Type())
	}

	var result object.Object = object.NIL
	varName := ""

	switch v := node.Variable.(type) {
	case *ast.Identifier:
		varName = v.Value
	default:
		return newError("invalid for loop variable")
	}

	for _, elem := range elements {
		env.Set(varName, elem)
		result = evalBlockBody(node.Body, env)

		if rv, ok := result.(*object.ReturnValue); ok {
			return rv
		}
		if _, ok := result.(*object.BreakValue); ok {
			return object.NIL
		}
		if _, ok := result.(*object.NextValue); ok {
			continue
		}
		if isError(result) {
			return result
		}
	}

	return result
}

func evalBeginExpression(node *ast.BeginExpression, env *object.Environment) object.Object {
	result := evalBlockBody(node.Body, env)

	// Check for errors/exceptions
	if err, ok := result.(*object.Error); ok {
		// Try to match rescue clauses
		for _, rescue := range node.Rescues {
			if matchesRescue(err, rescue) {
				rescueEnv := object.NewEnclosedEnvironment(env)
				if rescue.Variable != nil {
					rescueEnv.Set(rescue.Variable.Value, err)
				}
				result = evalBlockBody(rescue.Body, rescueEnv)
				break
			}
		}
	}

	// Execute ensure block
	if node.Ensure != nil {
		evalBlockBody(node.Ensure, env)
	}

	return result
}

func matchesRescue(err *object.Error, rescue *ast.RescueClause) bool {
	if len(rescue.Exceptions) == 0 {
		return true // Bare rescue matches all
	}

	// Check if error class matches any exception type
	// Simplified - just check by name
	for range rescue.Exceptions {
		// In a full implementation, we'd check class hierarchy
		return true
	}

	return false
}

// Other

func evalYieldExpression(node *ast.YieldExpression, env *object.Environment) object.Object {
	block := env.Block()
	if block == nil {
		return newError("no block given (yield)")
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return callBlock(block, args, env)
}

func callBlock(block *object.Proc, args []object.Object, env *object.Environment) object.Object {
	blockEnv := object.NewEnclosedEnvironment(block.Env)

	for i, param := range block.Parameters {
		if i < len(args) {
			blockEnv.Set(param.Name, args[i])
		} else {
			blockEnv.Set(param.Name, object.NIL)
		}
	}

	result := evalBlockBody(block.Body, blockEnv)

	// Unwrap next/break
	if nv, ok := result.(*object.NextValue); ok {
		return nv.Value
	}
	if bv, ok := result.(*object.BreakValue); ok {
		return bv.Value
	}

	return result
}

func evalLambda(node *ast.Lambda, env *object.Environment) object.Object {
	return &object.Lambda{
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
	}
}

func evalMethodDefinition(node *ast.MethodDefinition, env *object.Environment) object.Object {
	method := &object.Method{
		Name:       node.Name,
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
	}

	// Check for current class context (for class_eval)
	if currentClass := env.CurrentClass(); currentClass != nil {
		if node.Receiver != nil {
			currentClass.ClassMethods[node.Name] = method
		} else {
			currentClass.Methods[node.Name] = method
		}
		return &object.Symbol{Value: node.Name}
	}

	// Check for current module context (for module_eval)
	if currentModule := env.CurrentModule(); currentModule != nil {
		currentModule.Methods[node.Name] = method
		return &object.Symbol{Value: node.Name}
	}

	// Add to current class or module based on self
	self := env.Self()
	if self != nil {
		if class, ok := self.(*object.RubyClass); ok {
			if node.Receiver != nil {
				// Class method
				class.ClassMethods[node.Name] = method
			} else {
				class.Methods[node.Name] = method
			}
			return &object.Symbol{Value: node.Name}
		}
	}

	// Top-level method goes to Object
	object.ObjectClass.Methods[node.Name] = method
	return &object.Symbol{Value: node.Name}
}

func evalClassDefinition(node *ast.ClassDefinition, env *object.Environment) object.Object {
	var superclass *object.RubyClass = object.ObjectClass

	if node.Superclass != nil {
		superclassObj := Eval(node.Superclass, env)
		if isError(superclassObj) {
			return superclassObj
		}
		var ok bool
		superclass, ok = superclassObj.(*object.RubyClass)
		if !ok {
			return newError("superclass must be a Class")
		}
	}

	class := &object.RubyClass{
		Name:         node.Name.Value,
		Superclass:   superclass,
		Methods:      make(map[string]object.Object),
		ClassMethods: make(map[string]object.Object),
		Constants:    make(map[string]object.Object),
	}

	// Store in constants
	env.SetConstant(node.Name.Value, class)

	// Evaluate class body with class as self
	classEnv := object.NewEnclosedEnvironment(env)
	classEnv.SetSelf(class)
	evalBlockBody(node.Body, classEnv)

	return class
}

func evalModuleDefinition(node *ast.ModuleDefinition, env *object.Environment) object.Object {
	module := &object.RubyModule{
		Name:      node.Name.Value,
		Methods:   make(map[string]object.Object),
		Constants: make(map[string]object.Object),
	}

	env.SetConstant(node.Name.Value, module)

	moduleEnv := object.NewEnclosedEnvironment(env)
	moduleEnv.SetSelf(module)
	evalBlockBody(node.Body, moduleEnv)

	return module
}

func evalSplatExpression(node *ast.SplatExpression, env *object.Environment) object.Object {
	val := Eval(node.Expression, env)
	if isError(val) {
		return val
	}

	if arr, ok := val.(*object.Array); ok {
		return arr
	}

	return &object.Array{Elements: []object.Object{val}}
}

func evalDoubleSplatExpression(node *ast.DoubleSplatExpression, env *object.Environment) object.Object {
	val := Eval(node.Expression, env)
	if isError(val) {
		return val
	}

	if hash, ok := val.(*object.Hash); ok {
		return hash
	}

	return newError("no implicit conversion of %s into Hash", val.Type())
}

func evalDefinedExpression(node *ast.DefinedExpression, env *object.Environment) object.Object {
	switch expr := node.Expression.(type) {
	case *ast.Identifier:
		if _, ok := env.Get(expr.Value); ok {
			return &object.String{Value: "local-variable"}
		}
		return object.NIL
	case *ast.InstanceVariable:
		self := env.Self()
		if inst, ok := self.(*object.Instance); ok {
			if _, ok := inst.InstanceVariables[expr.Name]; ok {
				return &object.String{Value: "instance-variable"}
			}
		}
		return object.NIL
	case *ast.GlobalVariable:
		if _, ok := globalVariables[expr.Name]; ok {
			return &object.String{Value: "global-variable"}
		}
		return object.NIL
	case *ast.Constant:
		if _, ok := env.GetConstant(expr.Value); ok {
			return &object.String{Value: "constant"}
		}
		return object.NIL
	case *ast.MethodCall:
		return &object.String{Value: "method"}
	default:
		return &object.String{Value: "expression"}
	}
}

// Helper functions

func isTruthy(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.IsTruthy()
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch a := a.(type) {
	case *object.Integer:
		return a.Value == b.(*object.Integer).Value
	case *object.Float:
		return a.Value == b.(*object.Float).Value
	case *object.String:
		return a.Value == b.(*object.String).Value
	case *object.Symbol:
		return a.Value == b.(*object.Symbol).Value
	case *object.Boolean:
		return a.Value == b.(*object.Boolean).Value
	case *object.Nil:
		return true
	default:
		return a == b
	}
}

func objectToString(obj object.Object) string {
	switch o := obj.(type) {
	case *object.String:
		return o.Value
	case *object.Integer:
		return fmt.Sprintf("%d", o.Value)
	case *object.Float:
		return fmt.Sprintf("%g", o.Value)
	case *object.Boolean:
		return fmt.Sprintf("%t", o.Value)
	case *object.Nil:
		return ""
	case *object.Symbol:
		return o.Value
	default:
		return obj.Inspect()
	}
}

func expandRange(r *object.Range) []object.Object {
	var elements []object.Object

	startInt, ok := r.Start.(*object.Integer)
	if !ok {
		return elements
	}
	endInt, ok := r.End.(*object.Integer)
	if !ok {
		return elements
	}

	start := startInt.Value
	end := endInt.Value
	if r.Exclusive {
		end--
	}

	for i := start; i <= end; i++ {
		elements = append(elements, &object.Integer{Value: i})
	}

	return elements
}

func evalRangeIncludes(r *object.Range, val object.Object) object.Object {
	elements := expandRange(r)
	for _, elem := range elements {
		if objectsEqual(elem, val) {
			return object.TRUE
		}
	}
	return object.FALSE
}

func simpleMatch(pattern, str string) bool {
	// Very simplified pattern matching
	// In a real implementation, use Go's regexp package
	return len(pattern) > 0 && len(str) > 0
}
