package evaluator

import (
	"github.com/alexisbouchez/rubylexer/object"
)

// StructClass represents Ruby's Struct class
var StructClass = &object.RubyClass{
	Name:         "Struct",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

func init() {
	initStructClassMethods()
}

func initStructClassMethods() {
	StructClass.ClassMethods["new"] = &object.Builtin{
		Name: "new",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			// Struct.new(:name, :age) creates a new struct class
			members := make([]string, 0, len(args))
			for _, arg := range args {
				switch a := arg.(type) {
				case *object.Symbol:
					members = append(members, a.Value)
				case *object.String:
					members = append(members, a.Value)
				default:
					return newError("invalid member name: %s", arg.Type())
				}
			}

			return createStructClass(members)
		},
	}
}

func createStructClass(members []string) *object.RubyClass {
	structClass := &object.RubyClass{
		Name:         "Struct",
		Superclass:   StructClass,
		Methods:      make(map[string]object.Object),
		ClassMethods: make(map[string]object.Object),
		StructMembers: members,
	}

	// Add 'new' class method to create instances
	structClass.ClassMethods["new"] = &object.Builtin{
		Name: "new",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			class := receiver.(*object.RubyClass)
			instance := &object.Instance{
				Class_:            class,
				InstanceVariables: make(map[string]object.Object),
			}

			// Set instance variables from args
			for i, member := range class.StructMembers {
				if i < len(args) {
					instance.InstanceVariables["@"+member] = args[i]
				} else {
					instance.InstanceVariables["@"+member] = object.NIL
				}
			}

			return instance
		},
	}

	// Add 'members' class method
	structClass.ClassMethods["members"] = &object.Builtin{
		Name: "members",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			class := receiver.(*object.RubyClass)
			elements := make([]object.Object, len(class.StructMembers))
			for i, m := range class.StructMembers {
				elements[i] = &object.Symbol{Value: m}
			}
			return &object.Array{Elements: elements}
		},
	}

	// Add getter and setter for each member
	for _, member := range members {
		m := member // capture
		ivarName := "@" + m

		// Getter
		structClass.Methods[m] = &object.Builtin{
			Name: m,
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				if inst, ok := receiver.(*object.Instance); ok {
					if val, exists := inst.InstanceVariables[ivarName]; exists {
						return val
					}
				}
				return object.NIL
			},
		}

		// Setter
		structClass.Methods[m+"="] = &object.Builtin{
			Name: m + "=",
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("wrong number of arguments")
				}
				if inst, ok := receiver.(*object.Instance); ok {
					inst.InstanceVariables[ivarName] = args[0]
					return args[0]
				}
				return object.NIL
			},
		}
	}

	// Add instance methods
	structClass.Methods["members"] = &object.Builtin{
		Name: "members",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			inst := receiver.(*object.Instance)
			elements := make([]object.Object, len(inst.Class_.StructMembers))
			for i, m := range inst.Class_.StructMembers {
				elements[i] = &object.Symbol{Value: m}
			}
			return &object.Array{Elements: elements}
		},
	}

	structClass.Methods["to_a"] = &object.Builtin{
		Name: "to_a",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			inst := receiver.(*object.Instance)
			elements := make([]object.Object, len(inst.Class_.StructMembers))
			for i, m := range inst.Class_.StructMembers {
				if val, exists := inst.InstanceVariables["@"+m]; exists {
					elements[i] = val
				} else {
					elements[i] = object.NIL
				}
			}
			return &object.Array{Elements: elements}
		},
	}

	structClass.Methods["to_h"] = &object.Builtin{
		Name: "to_h",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			inst := receiver.(*object.Instance)
			pairs := make(map[object.HashKey]object.HashPair)
			order := make([]object.HashKey, 0, len(inst.Class_.StructMembers))

			for _, m := range inst.Class_.StructMembers {
				key := &object.Symbol{Value: m}
				hk := key.HashKey()
				var val object.Object = object.NIL
				if v, exists := inst.InstanceVariables["@"+m]; exists {
					val = v
				}
				pairs[hk] = object.HashPair{Key: key, Value: val}
				order = append(order, hk)
			}

			return &object.Hash{Pairs: pairs, Order: order}
		},
	}

	structClass.Methods["[]"] = &object.Builtin{
		Name: "[]",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments")
			}
			inst := receiver.(*object.Instance)

			var member string
			switch a := args[0].(type) {
			case *object.Symbol:
				member = a.Value
			case *object.String:
				member = a.Value
			case *object.Integer:
				idx := int(a.Value)
				if idx < 0 || idx >= len(inst.Class_.StructMembers) {
					return newError("index out of range")
				}
				member = inst.Class_.StructMembers[idx]
			default:
				return newError("invalid key type")
			}

			if val, exists := inst.InstanceVariables["@"+member]; exists {
				return val
			}
			return object.NIL
		},
	}

	structClass.Methods["[]="] = &object.Builtin{
		Name: "[]=",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments")
			}
			inst := receiver.(*object.Instance)

			var member string
			switch a := args[0].(type) {
			case *object.Symbol:
				member = a.Value
			case *object.String:
				member = a.Value
			case *object.Integer:
				idx := int(a.Value)
				if idx < 0 || idx >= len(inst.Class_.StructMembers) {
					return newError("index out of range")
				}
				member = inst.Class_.StructMembers[idx]
			default:
				return newError("invalid key type")
			}

			inst.InstanceVariables["@"+member] = args[1]
			return args[1]
		},
	}

	structClass.Methods["=="] = &object.Builtin{
		Name: "==",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			inst := receiver.(*object.Instance)
			other, ok := args[0].(*object.Instance)
			if !ok || inst.Class_ != other.Class_ {
				return object.FALSE
			}

			for _, m := range inst.Class_.StructMembers {
				v1 := inst.InstanceVariables["@"+m]
				v2 := other.InstanceVariables["@"+m]
				if !objectsEqual(v1, v2) {
					return object.FALSE
				}
			}
			return object.TRUE
		},
	}

	return structClass
}
