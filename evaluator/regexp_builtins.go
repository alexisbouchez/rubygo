package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/object"
)

var (
	regexpBuiltinsOnce sync.Once
	regexpBuiltinsMap  map[string]*object.Builtin
)

func getRegexpBuiltins() map[string]*object.Builtin {
	regexpBuiltinsOnce.Do(func() {
		regexpBuiltinsMap = map[string]*object.Builtin{
			"match": {
				Name: "match",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					re := receiver.(*object.Regexp)
					str, ok := args[0].(*object.String)
					if !ok {
						return newError("no implicit conversion of %s into String", args[0].Type())
					}

					matches := re.Match(str.Value)
					if matches == nil {
						return object.NIL
					}

					// Return a MatchData-like object (as an array for now)
					elements := make([]object.Object, len(matches))
					for i, m := range matches {
						elements[i] = &object.String{Value: m}
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
					re := receiver.(*object.Regexp)
					str, ok := args[0].(*object.String)
					if !ok {
						return object.NIL
					}

					if re.Compiled == nil {
						return object.NIL
					}

					loc := re.Compiled.FindStringIndex(str.Value)
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
					re := receiver.(*object.Regexp)
					str, ok := args[0].(*object.String)
					if !ok {
						return object.TRUE
					}

					if re.Compiled == nil {
						return object.TRUE
					}

					return object.NativeToBool(!re.Compiled.MatchString(str.Value))
				},
			},
			"===": {
				Name: "===",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return object.FALSE
					}
					re := receiver.(*object.Regexp)
					str, ok := args[0].(*object.String)
					if !ok {
						return object.FALSE
					}

					if re.Compiled == nil {
						return object.FALSE
					}

					return object.NativeToBool(re.Compiled.MatchString(str.Value))
				},
			},
			"source": {
				Name: "source",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					re := receiver.(*object.Regexp)
					return &object.String{Value: re.Pattern}
				},
			},
			"options": {
				Name: "options",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					re := receiver.(*object.Regexp)
					// Return flags as an integer (simplified)
					var opts int64
					for _, c := range re.Flags {
						switch c {
						case 'i':
							opts |= 1
						case 'm':
							opts |= 2
						case 'x':
							opts |= 4
						}
					}
					return &object.Integer{Value: opts}
				},
			},
			"to_s": {
				Name: "to_s",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					re := receiver.(*object.Regexp)
					return &object.String{Value: re.Inspect()}
				},
			},
			"inspect": {
				Name: "inspect",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					re := receiver.(*object.Regexp)
					return &object.String{Value: re.Inspect()}
				},
			},
		}
	})
	return regexpBuiltinsMap
}
