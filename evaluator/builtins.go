package evaluator

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/alexisbouchez/rubylexer/ast"
	"github.com/alexisbouchez/rubylexer/object"
)

// Lazy initialization for builtin maps to avoid initialization cycles
var (
	objectBuiltinsOnce   sync.Once
	kernelBuiltinsOnce   sync.Once
	integerBuiltinsOnce  sync.Once
	floatBuiltinsOnce    sync.Once
	stringBuiltinsOnce   sync.Once
	arrayBuiltinsOnce    sync.Once
	hashBuiltinsOnce     sync.Once
	rangeBuiltinsOnce    sync.Once
	symbolBuiltinsOnce   sync.Once
	nilBuiltinsOnce      sync.Once
	booleanBuiltinsOnce  sync.Once
	procBuiltinsOnce     sync.Once
	methodBuiltinsOnce   sync.Once

	objectBuiltinsMap  map[string]*object.Builtin
	kernelBuiltinsMap  map[string]*object.Builtin
	integerBuiltinsMap map[string]*object.Builtin
	floatBuiltinsMap   map[string]*object.Builtin
	stringBuiltinsMap  map[string]*object.Builtin
	arrayBuiltinsMap   map[string]*object.Builtin
	hashBuiltinsMap    map[string]*object.Builtin
	rangeBuiltinsMap   map[string]*object.Builtin
	symbolBuiltinsMap  map[string]*object.Builtin
	nilBuiltinsMap     map[string]*object.Builtin
	booleanBuiltinsMap map[string]*object.Builtin
	procBuiltinsMap    map[string]*object.Builtin
	methodBuiltinsMap  map[string]*object.Builtin
)

// Built-in method lookup
func getBuiltinMethod(receiver object.Object, name string) *object.Builtin {
	var typeBuiltin *object.Builtin

	switch receiver.Type() {
	case object.INTEGER_OBJ:
		typeBuiltin = getIntegerBuiltins()[name]
	case object.FLOAT_OBJ:
		typeBuiltin = getFloatBuiltins()[name]
	case object.STRING_OBJ:
		typeBuiltin = getStringBuiltins()[name]
	case object.ARRAY_OBJ:
		typeBuiltin = getArrayBuiltins()[name]
	case object.HASH_OBJ:
		typeBuiltin = getHashBuiltins()[name]
	case object.RANGE_OBJ:
		typeBuiltin = getRangeBuiltins()[name]
	case object.SYMBOL_OBJ:
		typeBuiltin = getSymbolBuiltins()[name]
	case object.NIL_OBJ:
		typeBuiltin = getNilBuiltins()[name]
	case object.BOOLEAN_OBJ:
		typeBuiltin = getBooleanBuiltins()[name]
	case object.PROC_OBJ, object.LAMBDA_OBJ:
		typeBuiltin = getProcBuiltins()[name]
	case object.METHOD_OBJ, object.BOUND_METHOD_OBJ:
		typeBuiltin = getMethodBuiltins()[name]
	case object.REGEXP_OBJ:
		typeBuiltin = getRegexpBuiltins()[name]
	case object.TIME_OBJ:
		typeBuiltin = getTimeBuiltins()[name]
	case object.DATE_OBJ:
		typeBuiltin = getDateBuiltins()[name]
	case object.CLASS_OBJ:
		// Check module builtins first (attr_accessor, include, etc.)
		if b := getModuleBuiltins()[name]; b != nil {
			return b
		}
		// Then class-specific builtins
		typeBuiltin = getClassBuiltins()[name]
	case object.MODULE_OBJ:
		typeBuiltin = getModuleBuiltins()[name]
	case object.ERROR_OBJ:
		typeBuiltin = getErrorBuiltins()[name]
	}

	if typeBuiltin != nil {
		return typeBuiltin
	}

	// Check Kernel methods
	if b := getKernelBuiltins()[name]; b != nil {
		return b
	}

	// Object methods
	return getObjectBuiltins()[name]
}

func getObjectBuiltins() map[string]*object.Builtin {
	objectBuiltinsOnce.Do(func() {
		objectBuiltinsMap = map[string]*object.Builtin{
			"class": {
				Name: "class",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver.Class()
				},
			},
			"inspect": {
				Name: "inspect",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: receiver.Inspect()}
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: receiver.Inspect()}
				},
			},
			"nil?": {
				Name: "nil?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.Type() == object.NIL_OBJ)
				},
			},
			"is_a?": {
				Name: "is_a?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					class, ok := args[0].(*object.RubyClass)
					if !ok {
						return object.FALSE
					}
					receiverClass := receiver.Class()
					for receiverClass != nil {
						if receiverClass == class {
							return object.TRUE
						}
						receiverClass = receiverClass.Superclass
					}
					return object.FALSE
				},
			},
			"kind_of?": {
				Name: "kind_of?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					class, ok := args[0].(*object.RubyClass)
					if !ok {
						return object.FALSE
					}
					receiverClass := receiver.Class()
					for receiverClass != nil {
						if receiverClass == class {
							return object.TRUE
						}
						receiverClass = receiverClass.Superclass
					}
					return object.FALSE
				},
			},
			"respond_to?": {
				Name: "respond_to?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					methodName := ""
					switch n := args[0].(type) {
					case *object.String:
						methodName = n.Value
					case *object.Symbol:
						methodName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					if class := receiver.Class(); class != nil {
						if _, ok := class.LookupMethod(methodName); ok {
							return object.TRUE
						}
					}
					if getBuiltinMethod(receiver, methodName) != nil {
						return object.TRUE
					}
					return object.FALSE
				},
			},
			"send": {
				Name: "send",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1+)")
					}
					methodName := ""
					switch n := args[0].(type) {
					case *object.String:
						methodName = n.Value
					case *object.Symbol:
						methodName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}
					return callMethod(receiver, methodName, args[1:], nil, env)
				},
			},
			"object_id": {
				Name: "object_id",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(uintptr(fmt.Sprintf("%p", receiver)[2:][0]))}
				},
			},
			"==": {
				Name: "==",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return object.FALSE
					}
					return object.NativeToBool(objectsEqual(receiver, args[0]))
				},
			},
			"!=": {
				Name: "!=",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return object.TRUE
					}
					return object.NativeToBool(!objectsEqual(receiver, args[0]))
				},
			},
			"instance_eval": {
				Name: "instance_eval",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					block := env.Block()
					if block == nil {
						return newError("no block given")
					}

					// Create new environment with self set to the receiver
					evalEnv := object.NewEnclosedEnvironment(block.Env)
					evalEnv.SetSelf(receiver)

					// If the receiver is an instance, we can define singleton methods
					// For now, just evaluate the block with self set to the receiver
					return evalBlockBody(block.Body, evalEnv)
				},
			},
			"methods": {
				Name: "methods",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					methods := []object.Object{}
					seen := make(map[string]bool)

					class := receiver.Class()
					for class != nil {
						for name := range class.Methods {
							if !seen[name] {
								methods = append(methods, &object.Symbol{Value: name})
								seen[name] = true
							}
						}
						// Include methods from included modules
						for _, mod := range class.IncludedModules {
							for name := range mod.Methods {
								if !seen[name] {
									methods = append(methods, &object.Symbol{Value: name})
									seen[name] = true
								}
							}
						}
						class = class.Superclass
					}

					return &object.Array{Elements: methods}
				},
			},
			"instance_variables": {
				Name: "instance_variables",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					instance, ok := receiver.(*object.Instance)
					if !ok {
						return &object.Array{Elements: []object.Object{}}
					}

					vars := []object.Object{}
					for name := range instance.InstanceVariables {
						vars = append(vars, &object.Symbol{Value: name})
					}
					return &object.Array{Elements: vars}
				},
			},
			"instance_variable_get": {
				Name: "instance_variable_get",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					varName := ""
					switch n := args[0].(type) {
					case *object.String:
						varName = n.Value
					case *object.Symbol:
						varName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					if !strings.HasPrefix(varName, "@") {
						varName = "@" + varName
					}

					instance, ok := receiver.(*object.Instance)
					if !ok {
						return object.NIL
					}

					return instance.GetInstanceVariable(varName)
				},
			},
			"instance_variable_set": {
				Name: "instance_variable_set",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 2 {
						return newError("wrong number of arguments (given %d, expected 2)", len(args))
					}

					varName := ""
					switch n := args[0].(type) {
					case *object.String:
						varName = n.Value
					case *object.Symbol:
						varName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					if !strings.HasPrefix(varName, "@") {
						varName = "@" + varName
					}

					instance, ok := receiver.(*object.Instance)
					if !ok {
						return newError("can't set instance variable on %s", receiver.Type())
					}

					instance.SetInstanceVariable(varName, args[1])
					return args[1]
				},
			},
			"instance_variable_defined?": {
				Name: "instance_variable_defined?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					varName := ""
					switch n := args[0].(type) {
					case *object.String:
						varName = n.Value
					case *object.Symbol:
						varName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					if !strings.HasPrefix(varName, "@") {
						varName = "@" + varName
					}

					instance, ok := receiver.(*object.Instance)
					if !ok {
						return object.FALSE
					}

					_, exists := instance.InstanceVariables[varName]
					return object.NativeToBool(exists)
				},
			},
			"freeze": {
				Name: "freeze",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					// Simplified freeze - just return self (not enforced)
					return receiver
				},
			},
			"frozen?": {
				Name: "frozen?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.FALSE
				},
			},
			"dup": {
				Name: "dup",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					// Shallow copy for instances
					if instance, ok := receiver.(*object.Instance); ok {
						newIvars := make(map[string]object.Object)
						for k, v := range instance.InstanceVariables {
							newIvars[k] = v
						}
						return &object.Instance{
							Class_:            instance.Class_,
							InstanceVariables: newIvars,
						}
					}
					return receiver
				},
			},
			"clone": {
				Name: "clone",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					// Same as dup for now (full clone would copy singleton class)
					if instance, ok := receiver.(*object.Instance); ok {
						newIvars := make(map[string]object.Object)
						for k, v := range instance.InstanceVariables {
							newIvars[k] = v
						}
						return &object.Instance{
							Class_:            instance.Class_,
							InstanceVariables: newIvars,
						}
					}
					return receiver
				},
			},
			"tap": {
				Name: "tap",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					block := env.Block()
					if block == nil {
						return newError("no block given")
					}

					// Call block with self
					blockEnv := object.NewEnclosedEnvironment(block.Env)
					if len(block.Parameters) > 0 {
						blockEnv.Set(block.Parameters[0].Name, receiver)
					}
					evalBlockBody(block.Body, blockEnv)

					return receiver
				},
			},
			"to_json": getToJSONBuiltin(),
			"method": {
				Name: "method",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					methodName := ""
					switch n := args[0].(type) {
					case *object.Symbol:
						methodName = n.Value
					case *object.String:
						methodName = n.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					// Look up the method
					if class := receiver.Class(); class != nil {
						if method, ok := class.LookupMethod(methodName); ok {
							if m, ok := method.(*object.Method); ok {
								// Return a bound method
								return &object.Method{
									Name:       m.Name,
									Parameters: m.Parameters,
									Body:       m.Body,
									Env:        m.Env,
									Receiver:   receiver,
								}
							}
							if b, ok := method.(*object.Builtin); ok {
								return &object.BoundMethod{
									Name:     methodName,
									Receiver: receiver,
									Builtin:  b,
								}
							}
						}
					}

					// Check builtins
					if builtin := getBuiltinMethod(receiver, methodName); builtin != nil {
						return &object.BoundMethod{
							Name:     methodName,
							Receiver: receiver,
							Builtin:  builtin,
						}
					}

					return newError("undefined method `%s'", methodName)
				},
			},
		}
	})
	return objectBuiltinsMap
}

func getKernelBuiltins() map[string]*object.Builtin {
	kernelBuiltinsOnce.Do(func() {
		kernelBuiltinsMap = map[string]*object.Builtin{
			"puts": {
				Name: "puts",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					for _, arg := range args {
						fmt.Println(objectToString(arg))
					}
					if len(args) == 0 {
						fmt.Println()
					}
					return object.NIL
				},
			},
			"print": {
				Name: "print",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					for _, arg := range args {
						fmt.Print(objectToString(arg))
					}
					return object.NIL
				},
			},
			"p": {
				Name: "p",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					for _, arg := range args {
						fmt.Println(arg.Inspect())
					}
					if len(args) == 1 {
						return args[0]
					}
					return &object.Array{Elements: args}
				},
			},
			"gets": {
				Name: "gets",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					var input string
					fmt.Scanln(&input)
					return &object.String{Value: input}
				},
			},
			"require": {
				Name: "require",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					filename, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
					return RequireFile(filename.Value, env)
				},
			},
			"require_relative": {
				Name: "require_relative",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					filename, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
					return RequireRelativeFile(filename.Value, env)
				},
			},
			"load": {
				Name: "load",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					filename, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
					return LoadFile(filename.Value, env)
				},
			},
			"raise": {
				Name: "raise",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					var message string
					var class *object.RubyClass = object.RuntimeErrorClass

					switch len(args) {
					case 0:
						message = "unhandled exception"
					case 1:
						switch arg := args[0].(type) {
						case *object.String:
							message = arg.Value
						case *object.RubyClass:
							class = arg
							message = arg.Name
						case *object.Exception:
							return &object.Error{Message: arg.Message, Class_: arg.Class_}
						default:
							message = args[0].Inspect()
						}
					default:
						if c, ok := args[0].(*object.RubyClass); ok {
							class = c
						}
						if s, ok := args[1].(*object.String); ok {
							message = s.Value
						}
					}

					return &object.Error{Message: message, Class_: class}
				},
			},
			"exit": {
				Name: "exit",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					code := 0
					if len(args) > 0 {
						if c, ok := args[0].(*object.Integer); ok {
							code = int(c.Value)
						}
					}
					os.Exit(code)
					return object.NIL
				},
			},
			"sleep": {
				Name: "sleep",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: 0}
				},
			},
			"rand": {
				Name: "rand",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Float{Value: 0.5}
				},
			},
			"lambda": {
				Name: "lambda",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					block := env.Block()
					if block == nil {
						return newError("tried to create Proc object without a block")
					}
					return &object.Lambda{
						Parameters: block.Parameters,
						Body:       block.Body,
						Env:        block.Env,
					}
				},
			},
			"proc": {
				Name: "proc",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					block := env.Block()
					if block == nil {
						return newError("tried to create Proc object without a block")
					}
					return block
				},
			},
			"block_given?": {
				Name: "block_given?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(env.Block() != nil)
				},
			},
			"loop": {
				Name: "loop",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					block := env.Block()
					if block == nil {
						return newError("no block given (yield)")
					}
					for {
						result := callBlock(block, []object.Object{}, env)
						if _, ok := result.(*object.BreakValue); ok {
							return object.NIL
						}
						if isError(result) {
							return result
						}
					}
				},
			},
		}
	})
	return kernelBuiltinsMap
}

func getIntegerBuiltins() map[string]*object.Builtin {
	integerBuiltinsOnce.Do(func() {
		integerBuiltinsMap = map[string]*object.Builtin{
			"to_i": {
				Name: "to_i",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"to_f": {
				Name: "to_f",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Float{Value: float64(receiver.(*object.Integer).Value)}
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					base := 10
					if len(args) > 0 {
						if b, ok := args[0].(*object.Integer); ok {
							base = int(b.Value)
						}
					}
					return &object.String{Value: fmt.Sprintf("%s", formatInt(receiver.(*object.Integer).Value, base))}
				},
			},
			"abs": {
				Name: "abs",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Integer).Value
					if val < 0 {
						val = -val
					}
					return &object.Integer{Value: val}
				},
			},
			"times": {
				Name: "times",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					n := receiver.(*object.Integer).Value
					block := env.Block()
					if block == nil {
						return receiver
					}
					for i := int64(0); i < n; i++ {
						result := callBlock(block, []object.Object{&object.Integer{Value: i}}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"upto": {
				Name: "upto",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					start := receiver.(*object.Integer).Value
					end, ok := args[0].(*object.Integer)
					if !ok {
						return newError("no implicit conversion of %s into Integer", args[0].Type())
					}
					block := env.Block()
					if block == nil {
						return receiver
					}
					for i := start; i <= end.Value; i++ {
						result := callBlock(block, []object.Object{&object.Integer{Value: i}}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"downto": {
				Name: "downto",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					start := receiver.(*object.Integer).Value
					end, ok := args[0].(*object.Integer)
					if !ok {
						return newError("no implicit conversion of %s into Integer", args[0].Type())
					}
					block := env.Block()
					if block == nil {
						return receiver
					}
					for i := start; i >= end.Value; i-- {
						result := callBlock(block, []object.Object{&object.Integer{Value: i}}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"even?": {
				Name: "even?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.(*object.Integer).Value%2 == 0)
				},
			},
			"odd?": {
				Name: "odd?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.(*object.Integer).Value%2 != 0)
				},
			},
			"zero?": {
				Name: "zero?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.(*object.Integer).Value == 0)
				},
			},
			"positive?": {
				Name: "positive?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.(*object.Integer).Value > 0)
				},
			},
			"negative?": {
				Name: "negative?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(receiver.(*object.Integer).Value < 0)
				},
			},
		}
	})
	return integerBuiltinsMap
}

func getFloatBuiltins() map[string]*object.Builtin {
	floatBuiltinsOnce.Do(func() {
		floatBuiltinsMap = map[string]*object.Builtin{
			"to_i": {
				Name: "to_i",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(receiver.(*object.Float).Value)}
				},
			},
			"to_f": {
				Name: "to_f",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: fmt.Sprintf("%g", receiver.(*object.Float).Value)}
				},
			},
			"abs": {
				Name: "abs",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					if val < 0 {
						val = -val
					}
					return &object.Float{Value: val}
				},
			},
			"round": {
				Name: "round",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					return &object.Integer{Value: int64(val + 0.5)}
				},
			},
			"ceil": {
				Name: "ceil",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					return &object.Integer{Value: int64(val) + 1}
				},
			},
			"floor": {
				Name: "floor",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					return &object.Integer{Value: int64(val)}
				},
			},
			"nan?": {
				Name: "nan?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					return object.NativeToBool(val != val)
				},
			},
			"infinite?": {
				Name: "infinite?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					val := receiver.(*object.Float).Value
					if val > 1e308 {
						return &object.Integer{Value: 1}
					}
					if val < -1e308 {
						return &object.Integer{Value: -1}
					}
					return object.NIL
				},
			},
		}
	})
	return floatBuiltinsMap
}

func getStringBuiltins() map[string]*object.Builtin {
	stringBuiltinsOnce.Do(func() {
		stringBuiltinsMap = map[string]*object.Builtin{
			"length": {
				Name: "length",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.String).Value))}
				},
			},
			"size": {
				Name: "size",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.String).Value))}
				},
			},
			"to_i": {
				Name: "to_i",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					var val int64
					fmt.Sscanf(receiver.(*object.String).Value, "%d", &val)
					return &object.Integer{Value: val}
				},
			},
			"to_f": {
				Name: "to_f",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					var val float64
					fmt.Sscanf(receiver.(*object.String).Value, "%f", &val)
					return &object.Float{Value: val}
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"to_sym": {
				Name: "to_sym",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Symbol{Value: receiver.(*object.String).Value}
				},
			},
			"upcase": {
				Name: "upcase",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: strings.ToUpper(receiver.(*object.String).Value)}
				},
			},
			"downcase": {
				Name: "downcase",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: strings.ToLower(receiver.(*object.String).Value)}
				},
			},
			"capitalize": {
				Name: "capitalize",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					if len(s) == 0 {
						return receiver
					}
					return &object.String{Value: strings.ToUpper(s[:1]) + strings.ToLower(s[1:])}
				},
			},
			"reverse": {
				Name: "reverse",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					runes := []rune(receiver.(*object.String).Value)
					for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
						runes[i], runes[j] = runes[j], runes[i]
					}
					return &object.String{Value: string(runes)}
				},
			},
			"strip": {
				Name: "strip",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: strings.TrimSpace(receiver.(*object.String).Value)}
				},
			},
			"chomp": {
				Name: "chomp",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					return &object.String{Value: strings.TrimRight(s, "\n\r")}
				},
			},
			"split": {
				Name: "split",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					sep := " "
					if len(args) > 0 {
						if sepObj, ok := args[0].(*object.String); ok {
							sep = sepObj.Value
						}
					}
					parts := strings.Split(s, sep)
					elements := make([]object.Object, len(parts))
					for i, p := range parts {
						elements[i] = &object.String{Value: p}
					}
					return &object.Array{Elements: elements}
				},
			},
			"include?": {
				Name: "include?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					substr, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
					return object.NativeToBool(strings.Contains(receiver.(*object.String).Value, substr.Value))
				},
			},
			"start_with?": {
				Name: "start_with?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					for _, arg := range args {
						if prefix, ok := arg.(*object.String); ok {
							if strings.HasPrefix(s, prefix.Value) {
								return object.TRUE
							}
						}
					}
					return object.FALSE
				},
			},
			"end_with?": {
				Name: "end_with?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					for _, arg := range args {
						if suffix, ok := arg.(*object.String); ok {
							if strings.HasSuffix(s, suffix.Value) {
								return object.TRUE
							}
						}
					}
					return object.FALSE
				},
			},
			"replace": {
				Name: "replace",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 2 {
						return newError("wrong number of arguments (given %d, expected 2)", len(args))
					}
					old, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
					new, ok := args[1].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[1].Type())
					}
					return &object.String{Value: strings.Replace(receiver.(*object.String).Value, old.Value, new.Value, 1)}
				},
			},
			"gsub": {
				Name: "gsub",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 2 {
						return newError("wrong number of arguments (given %d, expected 2)", len(args))
					}
					s := receiver.(*object.String).Value

					// Check if first arg is Regexp or String
					switch pattern := args[0].(type) {
					case *object.Regexp:
						newStr, ok := args[1].(*object.String)
						if !ok {
							return newError("no implicit conversion of %s into String", args[1].Type())
						}
						return &object.String{Value: pattern.ReplaceAll(s, newStr.Value)}
					case *object.String:
						newStr, ok := args[1].(*object.String)
						if !ok {
							return newError("no implicit conversion of %s into String", args[1].Type())
						}
						return &object.String{Value: strings.ReplaceAll(s, pattern.Value, newStr.Value)}
					default:
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
				},
			},
			"sub": {
				Name: "sub",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 2 {
						return newError("wrong number of arguments (given %d, expected 2)", len(args))
					}
					s := receiver.(*object.String).Value

					switch pattern := args[0].(type) {
					case *object.Regexp:
						newStr, ok := args[1].(*object.String)
						if !ok {
							return newError("no implicit conversion of %s into String", args[1].Type())
						}
						if pattern.Compiled != nil {
							// Only replace first match
							loc := pattern.Compiled.FindStringIndex(s)
							if loc != nil {
								return &object.String{Value: s[:loc[0]] + newStr.Value + s[loc[1]:]}
							}
						}
						return &object.String{Value: s}
					case *object.String:
						newStr, ok := args[1].(*object.String)
						if !ok {
							return newError("no implicit conversion of %s into String", args[1].Type())
						}
						return &object.String{Value: strings.Replace(s, pattern.Value, newStr.Value, 1)}
					default:
						return newError("no implicit conversion of %s into String", args[0].Type())
					}
				},
			},
			"match": {
				Name: "match",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					s := receiver.(*object.String).Value

					switch pattern := args[0].(type) {
					case *object.Regexp:
						matches := pattern.Match(s)
						if matches == nil {
							return object.NIL
						}
						elements := make([]object.Object, len(matches))
						for i, m := range matches {
							elements[i] = &object.String{Value: m}
						}
						return &object.Array{Elements: elements}
					case *object.String:
						// Convert string to regexp
						re, err := object.NewRegexp(pattern.Value, "")
						if err != nil {
							return object.NIL
						}
						matches := re.Match(s)
						if matches == nil {
							return object.NIL
						}
						elements := make([]object.Object, len(matches))
						for i, m := range matches {
							elements[i] = &object.String{Value: m}
						}
						return &object.Array{Elements: elements}
					default:
						return newError("wrong argument type %s (expected Regexp)", args[0].Type())
					}
				},
			},
			"scan": {
				Name: "scan",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					s := receiver.(*object.String).Value

					var re *object.Regexp
					switch pattern := args[0].(type) {
					case *object.Regexp:
						re = pattern
					case *object.String:
						var err error
						re, err = object.NewRegexp(pattern.Value, "")
						if err != nil {
							return newError("invalid regular expression: %s", err)
						}
					default:
						return newError("wrong argument type %s (expected Regexp)", args[0].Type())
					}

					allMatches := re.MatchAll(s)
					elements := make([]object.Object, 0)
					for _, matches := range allMatches {
						if len(matches) > 1 {
							// Has capture groups
							subElements := make([]object.Object, len(matches)-1)
							for i, m := range matches[1:] {
								subElements[i] = &object.String{Value: m}
							}
							elements = append(elements, &object.Array{Elements: subElements})
						} else {
							elements = append(elements, &object.String{Value: matches[0]})
						}
					}
					return &object.Array{Elements: elements}
				},
			},
			"=~": {
				Name: "=~",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return object.NIL
					}
					s := receiver.(*object.String).Value

					var re *object.Regexp
					switch pattern := args[0].(type) {
					case *object.Regexp:
						re = pattern
					case *object.String:
						var err error
						re, err = object.NewRegexp(pattern.Value, "")
						if err != nil {
							return object.NIL
						}
					default:
						return object.NIL
					}

					if re.Compiled == nil {
						return object.NIL
					}

					loc := re.Compiled.FindStringIndex(s)
					if loc == nil {
						return object.NIL
					}
					return &object.Integer{Value: int64(loc[0])}
				},
			},
			"!~": {
				Name: "!~",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return object.TRUE
					}
					s := receiver.(*object.String).Value

					var re *object.Regexp
					switch pattern := args[0].(type) {
					case *object.Regexp:
						re = pattern
					case *object.String:
						var err error
						re, err = object.NewRegexp(pattern.Value, "")
						if err != nil {
							return object.TRUE
						}
					default:
						return object.TRUE
					}

					if re.Compiled == nil {
						return object.TRUE
					}

					return object.NativeToBool(!re.Compiled.MatchString(s))
				},
			},
			"empty?": {
				Name: "empty?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(len(receiver.(*object.String).Value) == 0)
				},
			},
			"chars": {
				Name: "chars",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					chars := make([]object.Object, 0, len(s))
					for _, c := range s {
						chars = append(chars, &object.String{Value: string(c)})
					}
					return &object.Array{Elements: chars}
				},
			},
			"bytes": {
				Name: "bytes",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					bytes := make([]object.Object, len(s))
					for i, b := range []byte(s) {
						bytes[i] = &object.Integer{Value: int64(b)}
					}
					return &object.Array{Elements: bytes}
				},
			},
			"each_char": {
				Name: "each_char",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					s := receiver.(*object.String).Value
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, c := range s {
						result := callBlock(block, []object.Object{&object.String{Value: string(c)}}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
		}
	})
	return stringBuiltinsMap
}

func getArrayBuiltins() map[string]*object.Builtin {
	arrayBuiltinsOnce.Do(func() {
		arrayBuiltinsMap = map[string]*object.Builtin{
			"length": {
				Name: "length",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Array).Elements))}
				},
			},
			"size": {
				Name: "size",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Array).Elements))}
				},
			},
			"first": {
				Name: "first",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					if len(arr.Elements) == 0 {
						return object.NIL
					}
					if len(args) > 0 {
						n := args[0].(*object.Integer).Value
						if n > int64(len(arr.Elements)) {
							n = int64(len(arr.Elements))
						}
						return &object.Array{Elements: arr.Elements[:n]}
					}
					return arr.Elements[0]
				},
			},
			"last": {
				Name: "last",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					if len(arr.Elements) == 0 {
						return object.NIL
					}
					if len(args) > 0 {
						n := args[0].(*object.Integer).Value
						if n > int64(len(arr.Elements)) {
							n = int64(len(arr.Elements))
						}
						return &object.Array{Elements: arr.Elements[len(arr.Elements)-int(n):]}
					}
					return arr.Elements[len(arr.Elements)-1]
				},
			},
			"push": {
				Name: "push",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					arr.Elements = append(arr.Elements, args...)
					return arr
				},
			},
			"pop": {
				Name: "pop",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					if len(arr.Elements) == 0 {
						return object.NIL
					}
					last := arr.Elements[len(arr.Elements)-1]
					arr.Elements = arr.Elements[:len(arr.Elements)-1]
					return last
				},
			},
			"shift": {
				Name: "shift",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					if len(arr.Elements) == 0 {
						return object.NIL
					}
					first := arr.Elements[0]
					arr.Elements = arr.Elements[1:]
					return first
				},
			},
			"unshift": {
				Name: "unshift",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					arr.Elements = append(args, arr.Elements...)
					return arr
				},
			},
			"reverse": {
				Name: "reverse",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					newElements := make([]object.Object, len(arr.Elements))
					for i, j := 0, len(arr.Elements)-1; i <= j; i, j = i+1, j-1 {
						newElements[i] = arr.Elements[j]
						newElements[j] = arr.Elements[i]
					}
					return &object.Array{Elements: newElements}
				},
			},
			"sort": {
				Name: "sort",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					newElements := make([]object.Object, len(arr.Elements))
					copy(newElements, arr.Elements)
					sort.Slice(newElements, func(i, j int) bool {
						result := evalInfixExpression("<=>", newElements[i], newElements[j])
						if intResult, ok := result.(*object.Integer); ok {
							return intResult.Value < 0
						}
						return false
					})
					return &object.Array{Elements: newElements}
				},
			},
			"join": {
				Name: "join",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					sep := ""
					if len(args) > 0 {
						if s, ok := args[0].(*object.String); ok {
							sep = s.Value
						}
					}
					parts := make([]string, len(arr.Elements))
					for i, elem := range arr.Elements {
						parts[i] = objectToString(elem)
					}
					return &object.String{Value: strings.Join(parts, sep)}
				},
			},
			"include?": {
				Name: "include?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					arr := receiver.(*object.Array)
					for _, elem := range arr.Elements {
						if objectsEqual(elem, args[0]) {
							return object.TRUE
						}
					}
					return object.FALSE
				},
			},
			"index": {
				Name: "index",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					arr := receiver.(*object.Array)
					for i, elem := range arr.Elements {
						if objectsEqual(elem, args[0]) {
							return &object.Integer{Value: int64(i)}
						}
					}
					return object.NIL
				},
			},
			"empty?": {
				Name: "empty?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(len(receiver.(*object.Array).Elements) == 0)
				},
			},
			"each": {
				Name: "each",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"each_with_index": {
				Name: "each_with_index",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for i, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem, &object.Integer{Value: int64(i)}}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"map": {
				Name: "map",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newElements := make([]object.Object, 0, len(arr.Elements))
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						newElements = append(newElements, result)
					}
					return &object.Array{Elements: newElements}
				},
			},
			"collect": {
				Name: "collect",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					// Call map implementation directly
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newElements := make([]object.Object, 0, len(arr.Elements))
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						newElements = append(newElements, result)
					}
					return &object.Array{Elements: newElements}
				},
			},
			"select": {
				Name: "select",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newElements := make([]object.Object, 0)
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						if isTruthy(result) {
							newElements = append(newElements, elem)
						}
					}
					return &object.Array{Elements: newElements}
				},
			},
			"reject": {
				Name: "reject",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newElements := make([]object.Object, 0)
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						if !isTruthy(result) {
							newElements = append(newElements, elem)
						}
					}
					return &object.Array{Elements: newElements}
				},
			},
			"reduce": {
				Name: "reduce",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return newError("no block given")
					}

					var acc object.Object
					startIdx := 0
					if len(args) > 0 {
						acc = args[0]
					} else if len(arr.Elements) > 0 {
						acc = arr.Elements[0]
						startIdx = 1
					} else {
						return object.NIL
					}

					for i := startIdx; i < len(arr.Elements); i++ {
						result := callBlock(block, []object.Object{acc, arr.Elements[i]}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						acc = result
					}
					return acc
				},
			},
			"find": {
				Name: "find",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						if isTruthy(result) {
							return elem
						}
					}
					return object.NIL
				},
			},
			"any?": {
				Name: "any?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					if block == nil {
						return object.NativeToBool(len(arr.Elements) > 0)
					}
					for _, elem := range arr.Elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isTruthy(result) {
							return object.TRUE
						}
					}
					return object.FALSE
				},
			},
			"all?": {
				Name: "all?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					for _, elem := range arr.Elements {
						var result object.Object
						if block != nil {
							result = callBlock(block, []object.Object{elem}, env)
						} else {
							result = elem
						}
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if !isTruthy(result) {
							return object.FALSE
						}
					}
					return object.TRUE
				},
			},
			"none?": {
				Name: "none?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					block := env.Block()
					for _, elem := range arr.Elements {
						var result object.Object
						if block != nil {
							result = callBlock(block, []object.Object{elem}, env)
						} else {
							result = elem
						}
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isTruthy(result) {
							return object.FALSE
						}
					}
					return object.TRUE
				},
			},
			"compact": {
				Name: "compact",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					newElements := make([]object.Object, 0)
					for _, elem := range arr.Elements {
						if elem.Type() != object.NIL_OBJ {
							newElements = append(newElements, elem)
						}
					}
					return &object.Array{Elements: newElements}
				},
			},
			"flatten": {
				Name: "flatten",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					return &object.Array{Elements: flattenArray(arr.Elements)}
				},
			},
			"uniq": {
				Name: "uniq",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					arr := receiver.(*object.Array)
					seen := make(map[string]bool)
					newElements := make([]object.Object, 0)
					for _, elem := range arr.Elements {
						key := elem.Inspect()
						if !seen[key] {
							seen[key] = true
							newElements = append(newElements, elem)
						}
					}
					return &object.Array{Elements: newElements}
				},
			},
			"to_a": {
				Name: "to_a",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: receiver.Inspect()}
				},
			},
		}
	})
	return arrayBuiltinsMap
}

func getHashBuiltins() map[string]*object.Builtin {
	hashBuiltinsOnce.Do(func() {
		hashBuiltinsMap = map[string]*object.Builtin{
			"keys": {
				Name: "keys",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					keys := make([]object.Object, 0, len(hash.Pairs))
					for _, key := range hash.Order {
						keys = append(keys, hash.Pairs[key].Key)
					}
					return &object.Array{Elements: keys}
				},
			},
			"values": {
				Name: "values",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					values := make([]object.Object, 0, len(hash.Pairs))
					for _, key := range hash.Order {
						values = append(values, hash.Pairs[key].Value)
					}
					return &object.Array{Elements: values}
				},
			},
			"length": {
				Name: "length",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Hash).Pairs))}
				},
			},
			"size": {
				Name: "size",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Hash).Pairs))}
				},
			},
			"empty?": {
				Name: "empty?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(len(receiver.(*object.Hash).Pairs) == 0)
				},
			},
			"has_key?": {
				Name: "has_key?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					hash := receiver.(*object.Hash)
					key, ok := args[0].(object.Hashable)
					if !ok {
						return object.FALSE
					}
					_, exists := hash.Pairs[key.HashKey()]
					return object.NativeToBool(exists)
				},
			},
			"has_value?": {
				Name: "has_value?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					hash := receiver.(*object.Hash)
					for _, pair := range hash.Pairs {
						if objectsEqual(pair.Value, args[0]) {
							return object.TRUE
						}
					}
					return object.FALSE
				},
			},
			"each": {
				Name: "each",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, key := range hash.Order {
						pair := hash.Pairs[key]
						result := callBlock(block, []object.Object{pair.Key, pair.Value}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"each_key": {
				Name: "each_key",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, key := range hash.Order {
						result := callBlock(block, []object.Object{hash.Pairs[key].Key}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"each_value": {
				Name: "each_value",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					block := env.Block()
					if block == nil {
						return receiver
					}
					for _, key := range hash.Order {
						result := callBlock(block, []object.Object{hash.Pairs[key].Value}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"map": {
				Name: "map",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newElements := make([]object.Object, 0, len(hash.Pairs))
					for _, key := range hash.Order {
						pair := hash.Pairs[key]
						result := callBlock(block, []object.Object{pair.Key, pair.Value}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						newElements = append(newElements, result)
					}
					return &object.Array{Elements: newElements}
				},
			},
			"select": {
				Name: "select",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					block := env.Block()
					if block == nil {
						return receiver
					}
					newPairs := make(map[object.HashKey]object.HashPair)
					newOrder := make([]object.HashKey, 0)
					for _, key := range hash.Order {
						pair := hash.Pairs[key]
						result := callBlock(block, []object.Object{pair.Key, pair.Value}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
						if isTruthy(result) {
							newPairs[key] = pair
							newOrder = append(newOrder, key)
						}
					}
					return &object.Hash{Pairs: newPairs, Order: newOrder}
				},
			},
			"merge": {
				Name: "merge",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					newPairs := make(map[object.HashKey]object.HashPair)
					newOrder := make([]object.HashKey, 0, len(hash.Pairs))

					for _, key := range hash.Order {
						newPairs[key] = hash.Pairs[key]
						newOrder = append(newOrder, key)
					}

					for _, arg := range args {
						if other, ok := arg.(*object.Hash); ok {
							for _, key := range other.Order {
								if _, exists := newPairs[key]; !exists {
									newOrder = append(newOrder, key)
								}
								newPairs[key] = other.Pairs[key]
							}
						}
					}

					return &object.Hash{Pairs: newPairs, Order: newOrder}
				},
			},
			"to_a": {
				Name: "to_a",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					hash := receiver.(*object.Hash)
					elements := make([]object.Object, 0, len(hash.Pairs))
					for _, key := range hash.Order {
						pair := hash.Pairs[key]
						elements = append(elements, &object.Array{Elements: []object.Object{pair.Key, pair.Value}})
					}
					return &object.Array{Elements: elements}
				},
			},
			"to_h": {
				Name: "to_h",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: receiver.Inspect()}
				},
			},
			"delete": {
				Name: "delete",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					hash := receiver.(*object.Hash)
					key, ok := args[0].(object.Hashable)
					if !ok {
						return object.NIL
					}
					hashed := key.HashKey()
					pair, exists := hash.Pairs[hashed]
					if !exists {
						return object.NIL
					}
					delete(hash.Pairs, hashed)
					for i, k := range hash.Order {
						if k == hashed {
							hash.Order = append(hash.Order[:i], hash.Order[i+1:]...)
							break
						}
					}
					return pair.Value
				},
			},
			"fetch": {
				Name: "fetch",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1+)")
					}
					hash := receiver.(*object.Hash)
					key, ok := args[0].(object.Hashable)
					if !ok {
						return newError("unusable as hash key: %s", args[0].Type())
					}
					if pair, exists := hash.Pairs[key.HashKey()]; exists {
						return pair.Value
					}
					if len(args) > 1 {
						return args[1]
					}
					return newError("KeyError: key not found: %s", args[0].Inspect())
				},
			},
		}
	})
	return hashBuiltinsMap
}

func getRangeBuiltins() map[string]*object.Builtin {
	rangeBuiltinsOnce.Do(func() {
		rangeBuiltinsMap = map[string]*object.Builtin{
			"to_a": {
				Name: "to_a",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					r := receiver.(*object.Range)
					return &object.Array{Elements: expandRange(r)}
				},
			},
			"each": {
				Name: "each",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					r := receiver.(*object.Range)
					block := env.Block()
					if block == nil {
						return receiver
					}
					elements := expandRange(r)
					for _, elem := range elements {
						result := callBlock(block, []object.Object{elem}, env)
						if bv, ok := result.(*object.BreakValue); ok {
							return bv.Value
						}
						if isError(result) {
							return result
						}
					}
					return receiver
				},
			},
			"include?": {
				Name: "include?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					r := receiver.(*object.Range)
					return evalRangeIncludes(r, args[0])
				},
			},
			"first": {
				Name: "first",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					r := receiver.(*object.Range)
					return r.Start
				},
			},
			"last": {
				Name: "last",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					r := receiver.(*object.Range)
					return r.End
				},
			},
			"size": {
				Name: "size",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					r := receiver.(*object.Range)
					startInt, ok1 := r.Start.(*object.Integer)
					endInt, ok2 := r.End.(*object.Integer)
					if !ok1 || !ok2 {
						return object.NIL
					}
					size := endInt.Value - startInt.Value
					if !r.Exclusive {
						size++
					}
					if size < 0 {
						size = 0
					}
					return &object.Integer{Value: size}
				},
			},
		}
	})
	return rangeBuiltinsMap
}

func getSymbolBuiltins() map[string]*object.Builtin {
	symbolBuiltinsOnce.Do(func() {
		symbolBuiltinsMap = map[string]*object.Builtin{
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: receiver.(*object.Symbol).Value}
				},
			},
			"to_sym": {
				Name: "to_sym",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
			"upcase": {
				Name: "upcase",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Symbol{Value: strings.ToUpper(receiver.(*object.Symbol).Value)}
				},
			},
			"downcase": {
				Name: "downcase",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Symbol{Value: strings.ToLower(receiver.(*object.Symbol).Value)}
				},
			},
			"length": {
				Name: "length",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Symbol).Value))}
				},
			},
			"size": {
				Name: "size",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Integer{Value: int64(len(receiver.(*object.Symbol).Value))}
				},
			},
			"empty?": {
				Name: "empty?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NativeToBool(len(receiver.(*object.Symbol).Value) == 0)
				},
			},
		}
	})
	return symbolBuiltinsMap
}

func getNilBuiltins() map[string]*object.Builtin {
	nilBuiltinsOnce.Do(func() {
		nilBuiltinsMap = map[string]*object.Builtin{
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: ""}
				},
			},
			"to_a": {
				Name: "to_a",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Array{Elements: []object.Object{}}
				},
			},
			"to_h": {
				Name: "to_h",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
				},
			},
			"nil?": {
				Name: "nil?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.TRUE
				},
			},
			"inspect": {
				Name: "inspect",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return &object.String{Value: "nil"}
				},
			},
		}
	})
	return nilBuiltinsMap
}

func getBooleanBuiltins() map[string]*object.Builtin {
	booleanBuiltinsOnce.Do(func() {
		booleanBuiltinsMap = map[string]*object.Builtin{
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if receiver.(*object.Boolean).Value {
						return &object.String{Value: "true"}
					}
					return &object.String{Value: "false"}
				},
			},
		}
	})
	return booleanBuiltinsMap
}

func getProcBuiltins() map[string]*object.Builtin {
	procBuiltinsOnce.Do(func() {
		procBuiltinsMap = map[string]*object.Builtin{
			"call": {
				Name: "call",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch proc := receiver.(type) {
					case *object.Proc:
						return callBlock(proc, args, env)
					case *object.Lambda:
						return callBlock(&object.Proc{
							Parameters: proc.Parameters,
							Body:       proc.Body,
							Env:        proc.Env,
						}, args, env)
					default:
						return newError("not a callable object")
					}
				},
			},
			"arity": {
				Name: "arity",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch proc := receiver.(type) {
					case *object.Proc:
						return &object.Integer{Value: int64(len(proc.Parameters))}
					case *object.Lambda:
						return &object.Integer{Value: int64(len(proc.Parameters))}
					default:
						return &object.Integer{Value: 0}
					}
				},
			},
			"lambda?": {
				Name: "lambda?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					_, isLambda := receiver.(*object.Lambda)
					return object.NativeToBool(isLambda)
				},
			},
			"to_proc": {
				Name: "to_proc",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return receiver
				},
			},
		}
	})
	return procBuiltinsMap
}

func getMethodBuiltins() map[string]*object.Builtin {
	methodBuiltinsOnce.Do(func() {
		methodBuiltinsMap = map[string]*object.Builtin{
			"call": {
				Name: "call",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						// User-defined method with bound receiver
						return callUserMethod(m, m.Receiver, args, env)
					case *object.BoundMethod:
						if m.Builtin != nil {
							return m.Builtin.Fn(m.Receiver, env, args...)
						}
						if m.Method != nil {
							return callUserMethod(m.Method, m.Receiver, args, env)
						}
					}
					return newError("not a callable method object")
				},
			},
			"name": {
				Name: "name",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						return &object.Symbol{Value: m.Name}
					case *object.BoundMethod:
						return &object.Symbol{Value: m.Name}
					}
					return object.NIL
				},
			},
			"arity": {
				Name: "arity",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						return &object.Integer{Value: int64(len(m.Parameters))}
					case *object.BoundMethod:
						if m.Method != nil {
							return &object.Integer{Value: int64(len(m.Method.Parameters))}
						}
						// For builtins, we can't determine arity easily
						return &object.Integer{Value: -1}
					}
					return &object.Integer{Value: 0}
				},
			},
			"receiver": {
				Name: "receiver",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						if m.Receiver != nil {
							return m.Receiver
						}
					case *object.BoundMethod:
						return m.Receiver
					}
					return object.NIL
				},
			},
			"owner": {
				Name: "owner",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						if m.Receiver != nil {
							return m.Receiver.Class()
						}
					case *object.BoundMethod:
						if m.Receiver != nil {
							return m.Receiver.Class()
						}
					}
					return object.NIL
				},
			},
			"to_proc": {
				Name: "to_proc",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					switch m := receiver.(type) {
					case *object.Method:
						return &object.Proc{
							Parameters: convertMethodParamsToBlockParams(m.Parameters),
							Body:       m.Body,
							Env:        m.Env,
						}
					case *object.BoundMethod:
						if m.Method != nil {
							return &object.Proc{
								Parameters: convertMethodParamsToBlockParams(m.Method.Parameters),
								Body:       m.Method.Body,
								Env:        m.Method.Env,
							}
						}
					}
					return object.NIL
				},
			},
			"unbind": {
				Name: "unbind",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					// Returns an UnboundMethod - for now just return the method without receiver
					switch m := receiver.(type) {
					case *object.Method:
						return &object.Method{
							Name:       m.Name,
							Parameters: m.Parameters,
							Body:       m.Body,
							Env:        m.Env,
							Receiver:   nil,
						}
					case *object.BoundMethod:
						if m.Method != nil {
							return &object.Method{
								Name:       m.Method.Name,
								Parameters: m.Method.Parameters,
								Body:       m.Method.Body,
								Env:        m.Method.Env,
								Receiver:   nil,
							}
						}
					}
					return object.NIL
				},
			},
		}
	})
	return methodBuiltinsMap
}

// Helper functions

func formatInt(val int64, base int) string {
	switch base {
	case 2:
		return fmt.Sprintf("%b", val)
	case 8:
		return fmt.Sprintf("%o", val)
	case 16:
		return fmt.Sprintf("%x", val)
	default:
		return fmt.Sprintf("%d", val)
	}
}

func flattenArray(elements []object.Object) []object.Object {
	var result []object.Object
	for _, elem := range elements {
		if arr, ok := elem.(*object.Array); ok {
			result = append(result, flattenArray(arr.Elements)...)
		} else {
			result = append(result, elem)
		}
	}
	return result
}

func initKernelMethods() {
	for name, builtin := range getKernelBuiltins() {
		object.KernelModule.Methods[name] = builtin
	}
}

func init() {
	initKernelMethods()
}

// callUserMethod calls a user-defined method with a specific receiver
func callUserMethod(method *object.Method, receiver object.Object, args []object.Object, env *object.Environment) object.Object {
	methodEnv := object.NewEnclosedEnvironment(method.Env)
	methodEnv.SetSelf(receiver)

	// Bind parameters
	for i, param := range method.Parameters {
		if i < len(args) {
			methodEnv.Set(param.Name, args[i])
		} else if param.Default != nil {
			// Use default value
			defaultVal := Eval(param.Default, methodEnv)
			methodEnv.Set(param.Name, defaultVal)
		} else {
			methodEnv.Set(param.Name, object.NIL)
		}
	}

	result := evalBlockBody(method.Body, methodEnv)
	if rv, ok := result.(*object.ReturnValue); ok {
		return rv.Value
	}
	return result
}

// convertMethodParamsToBlockParams converts method parameters to block parameters
func convertMethodParamsToBlockParams(params []*ast.MethodParameter) []*ast.BlockParameter {
	blockParams := make([]*ast.BlockParameter, len(params))
	for i, p := range params {
		blockParams[i] = &ast.BlockParameter{
			Name: p.Name,
		}
	}
	return blockParams
}

// Error builtins

var errorBuiltinsOnce sync.Once
var errorBuiltinsMap map[string]*object.Builtin

func getErrorBuiltins() map[string]*object.Builtin {
	errorBuiltinsOnce.Do(func() {
		errorBuiltinsMap = map[string]*object.Builtin{
			"message": {
				Name: "message",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					err := receiver.(*object.Error)
					return &object.String{Value: err.Message}
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					err := receiver.(*object.Error)
					return &object.String{Value: err.Message}
				},
			},
			"backtrace": {
				Name: "backtrace",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					err := receiver.(*object.Error)
					elements := make([]object.Object, len(err.Backtrace))
					for i, line := range err.Backtrace {
						elements[i] = &object.String{Value: line}
					}
					return &object.Array{Elements: elements}
				},
			},
		}
	})
	return errorBuiltinsMap
}
