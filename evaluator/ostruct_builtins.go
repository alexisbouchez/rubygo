package evaluator

import (
	"strings"

	"github.com/alexisbouchez/rubylexer/object"
)

// OpenStructClass represents Ruby's OpenStruct class
var OpenStructClass = &object.RubyClass{
	Name:         "OpenStruct",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

func init() {
	initOpenStructMethods()
}

func initOpenStructMethods() {
	OpenStructClass.ClassMethods["new"] = &object.Builtin{
		Name: "new",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			instance := &object.Instance{
				Class_:            OpenStructClass,
				InstanceVariables: make(map[string]object.Object),
			}

			// If a hash is passed, set initial attributes
			if len(args) > 0 {
				if hash, ok := args[0].(*object.Hash); ok {
					for _, hk := range hash.Order {
						pair := hash.Pairs[hk]
						keyName := ""
						switch k := pair.Key.(type) {
						case *object.String:
							keyName = k.Value
						case *object.Symbol:
							keyName = k.Value
						}
						if keyName != "" {
							instance.InstanceVariables["@"+keyName] = pair.Value
						}
					}
				}
			}

			return instance
		},
	}

	// method_missing handles dynamic attribute access
	OpenStructClass.Methods["method_missing"] = &object.Builtin{
		Name: "method_missing",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.NIL
			}

			inst := receiver.(*object.Instance)
			methodName := ""
			if sym, ok := args[0].(*object.Symbol); ok {
				methodName = sym.Value
			}

			// Check if it's a setter (ends with =)
			if strings.HasSuffix(methodName, "=") {
				attrName := strings.TrimSuffix(methodName, "=")
				if len(args) > 1 {
					inst.InstanceVariables["@"+attrName] = args[1]
					return args[1]
				}
				return object.NIL
			}

			// Getter
			if val, exists := inst.InstanceVariables["@"+methodName]; exists {
				return val
			}
			return object.NIL
		},
	}

	OpenStructClass.Methods["to_h"] = &object.Builtin{
		Name: "to_h",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			inst := receiver.(*object.Instance)
			pairs := make(map[object.HashKey]object.HashPair)
			order := make([]object.HashKey, 0)

			for name, val := range inst.InstanceVariables {
				attrName := strings.TrimPrefix(name, "@")
				key := &object.Symbol{Value: attrName}
				hk := key.HashKey()
				pairs[hk] = object.HashPair{Key: key, Value: val}
				order = append(order, hk)
			}

			return &object.Hash{Pairs: pairs, Order: order}
		},
	}

	OpenStructClass.Methods["[]"] = &object.Builtin{
		Name: "[]",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.NIL
			}
			inst := receiver.(*object.Instance)

			attrName := ""
			switch k := args[0].(type) {
			case *object.String:
				attrName = k.Value
			case *object.Symbol:
				attrName = k.Value
			}

			if val, exists := inst.InstanceVariables["@"+attrName]; exists {
				return val
			}
			return object.NIL
		},
	}

	OpenStructClass.Methods["[]="] = &object.Builtin{
		Name: "[]=",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return object.NIL
			}
			inst := receiver.(*object.Instance)

			attrName := ""
			switch k := args[0].(type) {
			case *object.String:
				attrName = k.Value
			case *object.Symbol:
				attrName = k.Value
			}

			inst.InstanceVariables["@"+attrName] = args[1]
			return args[1]
		},
	}

	OpenStructClass.Methods["=="] = &object.Builtin{
		Name: "==",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			inst := receiver.(*object.Instance)
			other, ok := args[0].(*object.Instance)
			if !ok || other.Class_ != OpenStructClass {
				return object.FALSE
			}

			if len(inst.InstanceVariables) != len(other.InstanceVariables) {
				return object.FALSE
			}

			for k, v := range inst.InstanceVariables {
				if ov, exists := other.InstanceVariables[k]; !exists || !objectsEqual(v, ov) {
					return object.FALSE
				}
			}
			return object.TRUE
		},
	}

	OpenStructClass.Methods["inspect"] = &object.Builtin{
		Name: "inspect",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			inst := receiver.(*object.Instance)
			parts := []string{}
			for name, val := range inst.InstanceVariables {
				attrName := strings.TrimPrefix(name, "@")
				parts = append(parts, attrName+"="+val.Inspect())
			}
			return &object.String{Value: "#<OpenStruct " + strings.Join(parts, ", ") + ">"}
		},
	}
}
