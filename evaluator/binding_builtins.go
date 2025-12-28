package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/lexer"
	"github.com/alexisbouchez/rubylexer/object"
	"github.com/alexisbouchez/rubylexer/parser"
)

var bindingBuiltinsOnce sync.Once
var bindingBuiltinsMap map[string]*object.Builtin

func getBindingBuiltins() map[string]*object.Builtin {
	bindingBuiltinsOnce.Do(func() {
		bindingBuiltinsMap = map[string]*object.Builtin{
			"local_variables": {
				Name: "local_variables",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					binding := receiver.(*object.Binding)
					names := binding.Env.LocalVariableNames()
					symbols := make([]object.Object, len(names))
					for i, name := range names {
						symbols[i] = &object.Symbol{Value: name}
					}
					return &object.Array{Elements: symbols}
				},
			},
			"local_variable_get": {
				Name: "local_variable_get",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) == 0 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					binding := receiver.(*object.Binding)
					var name string

					switch arg := args[0].(type) {
					case *object.Symbol:
						name = arg.Value
					case *object.String:
						name = arg.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					val, ok := binding.Env.Get(name)
					if !ok {
						return newError("NameError: local variable `%s' is not defined", name)
					}
					return val
				},
			},
			"local_variable_set": {
				Name: "local_variable_set",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 2 {
						return newError("wrong number of arguments (given %d, expected 2)", len(args))
					}

					binding := receiver.(*object.Binding)
					var name string

					switch arg := args[0].(type) {
					case *object.Symbol:
						name = arg.Value
					case *object.String:
						name = arg.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					binding.Env.Set(name, args[1])
					return args[1]
				},
			},
			"local_variable_defined?": {
				Name: "local_variable_defined?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) == 0 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					binding := receiver.(*object.Binding)
					var name string

					switch arg := args[0].(type) {
					case *object.Symbol:
						name = arg.Value
					case *object.String:
						name = arg.Value
					default:
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}

					_, ok := binding.Env.Get(name)
					if ok {
						return object.TRUE
					}
					return object.FALSE
				},
			},
			"receiver": {
				Name: "receiver",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					binding := receiver.(*object.Binding)
					if binding.Receiver != nil {
						return binding.Receiver
					}
					return object.NIL
				},
			},
			"eval": {
				Name: "eval",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) == 0 {
						return newError("wrong number of arguments (given 0, expected 1..3)")
					}

					code, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}

					binding := receiver.(*object.Binding)
					return evalInBinding(code.Value, binding)
				},
			},
		}
	})
	return bindingBuiltinsMap
}

// evalInBinding evaluates code in the context of a binding
func evalInBinding(code string, binding *object.Binding) object.Object {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return newError("SyntaxError: %s", p.Errors()[0])
	}

	// Evaluate in the binding's environment
	return Eval(program, binding.Env)
}
