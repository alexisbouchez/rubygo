package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/object"
)

var enumeratorBuiltinsOnce sync.Once
var enumeratorBuiltinsMap map[string]*object.Builtin

func getEnumeratorBuiltins() map[string]*object.Builtin {
	enumeratorBuiltinsOnce.Do(func() {
		enumeratorBuiltinsMap = map[string]*object.Builtin{
			"each": {
				Name: "each",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					block := env.Block()
					if block == nil {
						// Return self if no block given
						return enum
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					var result object.Object = object.NIL
					for _, val := range enum.Values {
						blockEnv := object.NewEnclosedEnvironment(env)
						if block.Parameters != nil && len(block.Parameters) > 0 {
							blockEnv.Set(block.Parameters[0].Name, val)
						}
						result = evalBlockBody(block.Body, blockEnv)
						if isControlFlow(result) {
							break
						}
					}
					return result
				},
			},
			"next": {
				Name: "next",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					if enum.Index >= len(enum.Values) {
						return newError("StopIteration: iteration reached an end")
					}

					val := enum.Values[enum.Index]
					enum.Index++
					enum.Started = true
					return val
				},
			},
			"peek": {
				Name: "peek",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					if enum.Index >= len(enum.Values) {
						return newError("StopIteration: iteration reached an end")
					}

					return enum.Values[enum.Index]
				},
			},
			"rewind": {
				Name: "rewind",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					enum.Index = 0
					enum.Started = false
					return enum
				},
			},
			"to_a": {
				Name: "to_a",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					return &object.Array{Elements: enum.Values}
				},
			},
			"first": {
				Name: "first",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					if len(args) == 0 {
						if len(enum.Values) == 0 {
							return object.NIL
						}
						return enum.Values[0]
					}

					n, ok := args[0].(*object.Integer)
					if !ok {
						return newError("no implicit conversion to Integer")
					}

					count := int(n.Value)
					if count > len(enum.Values) {
						count = len(enum.Values)
					}
					return &object.Array{Elements: enum.Values[:count]}
				},
			},
			"count": {
				Name: "count",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					return &object.Integer{Value: int64(len(enum.Values))}
				},
			},
			"with_index": {
				Name: "with_index",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					block := env.Block()

					offset := int64(0)
					if len(args) > 0 {
						if n, ok := args[0].(*object.Integer); ok {
							offset = n.Value
						}
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					if block == nil {
						// Return new enumerator with indexed values
						indexedValues := make([]object.Object, len(enum.Values))
						for i, val := range enum.Values {
							indexedValues[i] = &object.Array{Elements: []object.Object{val, &object.Integer{Value: int64(i) + offset}}}
						}
						return &object.Enumerator{
							Object: enum.Object,
							Method: enum.Method + ".with_index",
							Values: indexedValues,
						}
					}

					var result object.Object = object.NIL
					for i, val := range enum.Values {
						blockEnv := object.NewEnclosedEnvironment(env)
						if block.Parameters != nil {
							if len(block.Parameters) >= 1 {
								blockEnv.Set(block.Parameters[0].Name, val)
							}
							if len(block.Parameters) >= 2 {
								blockEnv.Set(block.Parameters[1].Name, &object.Integer{Value: int64(i) + offset})
							}
						}
						result = evalBlockBody(block.Body, blockEnv)
						if isControlFlow(result) {
							break
						}
					}
					return result
				},
			},
			"map": {
				Name: "map",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					block := env.Block()

					if block == nil {
						// Return lazy enumerator
						return &object.Enumerator{
							Object: enum,
							Method: "map",
							Lazy:   true,
						}
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					results := make([]object.Object, 0, len(enum.Values))
					for _, val := range enum.Values {
						blockEnv := object.NewEnclosedEnvironment(env)
						if block.Parameters != nil && len(block.Parameters) > 0 {
							blockEnv.Set(block.Parameters[0].Name, val)
						}
						result := evalBlockBody(block.Body, blockEnv)
						if isControlFlow(result) {
							break
						}
						results = append(results, result)
					}
					return &object.Array{Elements: results}
				},
			},
			"select": {
				Name: "select",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					block := env.Block()

					if block == nil {
						// Return lazy enumerator
						return &object.Enumerator{
							Object: enum,
							Method: "select",
							Lazy:   true,
						}
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					results := make([]object.Object, 0)
					for _, val := range enum.Values {
						blockEnv := object.NewEnclosedEnvironment(env)
						if block.Parameters != nil && len(block.Parameters) > 0 {
							blockEnv.Set(block.Parameters[0].Name, val)
						}
						result := evalBlockBody(block.Body, blockEnv)
						if isControlFlow(result) {
							break
						}
						if isTruthy(result) {
							results = append(results, val)
						}
					}
					return &object.Array{Elements: results}
				},
			},
			"lazy": {
				Name: "lazy",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)
					return &object.Enumerator{
						Object:  enum.Object,
						Method:  enum.Method,
						Args:    enum.Args,
						Values:  enum.Values,
						Index:   enum.Index,
						Started: enum.Started,
						Lazy:    true,
						LazyOps: enum.LazyOps,
					}
				},
			},
			"force": {
				Name: "force",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					// Apply lazy operations
					values := enum.Values
					for _, op := range enum.LazyOps {
						switch op.Type {
						case "take":
							if op.Count < len(values) {
								values = values[:op.Count]
							}
						case "drop":
							if op.Count < len(values) {
								values = values[op.Count:]
							} else {
								values = []object.Object{}
							}
						case "map":
							if op.Block != nil {
								newValues := make([]object.Object, 0, len(values))
								for _, val := range values {
									blockEnv := object.NewEnclosedEnvironment(env)
									if op.Block.Parameters != nil && len(op.Block.Parameters) > 0 {
										blockEnv.Set(op.Block.Parameters[0].Name, val)
									}
									result := evalBlockBody(op.Block.Body, blockEnv)
									newValues = append(newValues, result)
								}
								values = newValues
							}
						case "select":
							if op.Block != nil {
								newValues := make([]object.Object, 0)
								for _, val := range values {
									blockEnv := object.NewEnclosedEnvironment(env)
									if op.Block.Parameters != nil && len(op.Block.Parameters) > 0 {
										blockEnv.Set(op.Block.Parameters[0].Name, val)
									}
									result := evalBlockBody(op.Block.Body, blockEnv)
									if isTruthy(result) {
										newValues = append(newValues, val)
									}
								}
								values = newValues
							}
						}
					}

					return &object.Array{Elements: values}
				},
			},
			"take": {
				Name: "take",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					if len(args) == 0 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					n, ok := args[0].(*object.Integer)
					if !ok {
						return newError("no implicit conversion to Integer")
					}

					if enum.Lazy {
						// For lazy enumerators, add to the chain
						newEnum := &object.Enumerator{
							Object:  enum.Object,
							Method:  enum.Method,
							Args:    enum.Args,
							Lazy:    true,
							LazyOps: append(enum.LazyOps, object.LazyOperation{Type: "take", Count: int(n.Value)}),
						}
						return newEnum
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					count := int(n.Value)
					if count > len(enum.Values) {
						count = len(enum.Values)
					}
					return &object.Array{Elements: enum.Values[:count]}
				},
			},
			"drop": {
				Name: "drop",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					enum := receiver.(*object.Enumerator)

					if len(args) == 0 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}

					n, ok := args[0].(*object.Integer)
					if !ok {
						return newError("no implicit conversion to Integer")
					}

					if enum.Lazy {
						// For lazy enumerators, add to the chain
						newEnum := &object.Enumerator{
							Object:  enum.Object,
							Method:  enum.Method,
							Args:    enum.Args,
							Lazy:    true,
							LazyOps: append(enum.LazyOps, object.LazyOperation{Type: "drop", Count: int(n.Value)}),
						}
						return newEnum
					}

					// Materialize values if needed
					if enum.Values == nil {
						materializeEnumerator(enum, env)
					}

					count := int(n.Value)
					if count > len(enum.Values) {
						count = len(enum.Values)
					}
					return &object.Array{Elements: enum.Values[count:]}
				},
			},
		}
	})
	return enumeratorBuiltinsMap
}

// materializeEnumerator collects all values from the enumerator's source
func materializeEnumerator(enum *object.Enumerator, env *object.Environment) {
	enum.Values = []object.Object{}

	switch obj := enum.Object.(type) {
	case *object.Array:
		enum.Values = obj.Elements
	case *object.Range:
		enum.Values = expandRange(obj)
	case *object.Hash:
		for _, key := range obj.Order {
			pair := obj.Pairs[key]
			enum.Values = append(enum.Values, &object.Array{Elements: []object.Object{pair.Key, pair.Value}})
		}
	case *object.String:
		switch enum.Method {
		case "each_char":
			for _, r := range obj.Value {
				enum.Values = append(enum.Values, &object.String{Value: string(r)})
			}
		case "each_byte":
			for _, b := range []byte(obj.Value) {
				enum.Values = append(enum.Values, &object.Integer{Value: int64(b)})
			}
		case "each_line":
			lines := splitLines(obj.Value)
			for _, line := range lines {
				enum.Values = append(enum.Values, &object.String{Value: line})
			}
		default:
			for _, r := range obj.Value {
				enum.Values = append(enum.Values, &object.String{Value: string(r)})
			}
		}
	case *object.Enumerator:
		// Nested enumerator - materialize the inner one first
		if obj.Values == nil {
			materializeEnumerator(obj, env)
		}
		enum.Values = obj.Values
	}
}

// splitLines splits a string by newlines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// isControlFlow checks if result is a control flow object
func isControlFlow(result object.Object) bool {
	if result == nil {
		return false
	}
	switch result.Type() {
	case object.RETURN_VALUE_OBJ, object.BREAK_VALUE_OBJ, object.NEXT_VALUE_OBJ, object.RETRY_VALUE_OBJ:
		return true
	case object.ERROR_OBJ:
		if err, ok := result.(*object.Error); ok && !err.Caught {
			return true
		}
	}
	return false
}
