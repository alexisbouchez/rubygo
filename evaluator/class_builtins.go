package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/ast"
	"github.com/alexisbouchez/rubylexer/object"
)

var (
	classBuiltinsOnce  sync.Once
	moduleBuiltinsOnce sync.Once

	classBuiltinsMap  map[string]*object.Builtin
	moduleBuiltinsMap map[string]*object.Builtin
)

// getClassBuiltins returns builtins for Class objects
func getClassBuiltins() map[string]*object.Builtin {
	classBuiltinsOnce.Do(func() {
		classBuiltinsMap = map[string]*object.Builtin{
			"new": {
				Name: "new",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					class := receiver.(*object.RubyClass)
					return createInstance(class, args, env.Block(), env)
				},
			},
			"name": {
				Name: "name",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					class := receiver.(*object.RubyClass)
					return &object.String{Value: class.Name}
				},
			},
			"superclass": {
				Name: "superclass",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					class := receiver.(*object.RubyClass)
					if class.Superclass == nil {
						return object.NIL
					}
					return class.Superclass
				},
			},
			"ancestors": {
				Name: "ancestors",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					class := receiver.(*object.RubyClass)
					ancestors := []object.Object{}
					current := class
					for current != nil {
						ancestors = append(ancestors, current)
						// Add included modules
						for _, mod := range current.IncludedModules {
							ancestors = append(ancestors, mod)
						}
						current = current.Superclass
					}
					return &object.Array{Elements: ancestors}
				},
			},
			"instance_methods": {
				Name: "instance_methods",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					class := receiver.(*object.RubyClass)
					includeSuper := true
					if len(args) > 0 {
						if b, ok := args[0].(*object.Boolean); ok {
							includeSuper = b.Value
						}
					}

					methods := []object.Object{}
					seen := make(map[string]bool)

					current := class
					for current != nil {
						for name := range current.Methods {
							if !seen[name] {
								methods = append(methods, &object.Symbol{Value: name})
								seen[name] = true
							}
						}
						if !includeSuper {
							break
						}
						current = current.Superclass
					}
					return &object.Array{Elements: methods}
				},
			},
			"method_defined?": {
				Name: "method_defined?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					if len(args) < 1 {
						return newError("wrong number of arguments (given 0, expected 1)")
					}
					class := receiver.(*object.RubyClass)
					methodName := getMethodName(args[0])
					if methodName == "" {
						return newError("no implicit conversion of %s into Symbol", args[0].Type())
					}
					_, found := class.LookupMethod(methodName)
					return object.NativeToBool(found)
				},
			},
		}
	})
	return classBuiltinsMap
}

// getModuleBuiltins returns builtins for Module objects (also used by Class)
func getModuleBuiltins() map[string]*object.Builtin {
	moduleBuiltinsOnce.Do(func() {
		moduleBuiltinsMap = map[string]*object.Builtin{
			"attr_reader": {
				Name: "attr_reader",
				Fn:   attrReaderFn,
			},
			"attr_writer": {
				Name: "attr_writer",
				Fn:   attrWriterFn,
			},
			"attr_accessor": {
				Name: "attr_accessor",
				Fn:   attrAccessorFn,
			},
			"include": {
				Name: "include",
				Fn:   includeFn,
			},
			"extend": {
				Name: "extend",
				Fn:   extendFn,
			},
			"prepend": {
				Name: "prepend",
				Fn:   prependFn,
			},
			"define_method": {
				Name: "define_method",
				Fn:   defineMethodFn,
			},
			"alias_method": {
				Name: "alias_method",
				Fn:   aliasMethodFn,
			},
			"private": {
				Name: "private",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return setVisibility(receiver, env, object.VisibilityPrivate, args...)
				},
			},
			"protected": {
				Name: "protected",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return setVisibility(receiver, env, object.VisibilityProtected, args...)
				},
			},
			"public": {
				Name: "public",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return setVisibility(receiver, env, object.VisibilityPublic, args...)
				},
			},
			"module_function": {
				Name: "module_function",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					return object.NIL
				},
			},
			"class_eval": {
				Name: "class_eval",
				Fn:   classEvalFn,
			},
			"module_eval": {
				Name: "module_eval",
				Fn:   classEvalFn, // Same as class_eval
			},
			"refine": {
				Name: "refine",
				Fn:   refineFn,
			},
			"using": {
				Name: "using",
				Fn:   usingFn,
			},
		}
	})
	return moduleBuiltinsMap
}

// classEvalFn evaluates a block in the context of the class/module
func classEvalFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	block := env.Block()
	if block == nil {
		return newError("no block given")
	}

	// Create new environment with self set to the class/module
	evalEnv := object.NewEnclosedEnvironment(block.Env)
	evalEnv.SetSelf(receiver)

	// If this is a class, set it as the current class for method definitions
	if class, ok := receiver.(*object.RubyClass); ok {
		evalEnv.SetCurrentClass(class)
	} else if mod, ok := receiver.(*object.RubyModule); ok {
		// For modules, we need a way to add methods
		// Create a temporary class wrapper
		evalEnv.SetCurrentModule(mod)
	}

	// Evaluate the block
	return evalBlockBody(block.Body, evalEnv)
}

func getMethodName(arg object.Object) string {
	switch a := arg.(type) {
	case *object.String:
		return a.Value
	case *object.Symbol:
		return a.Value
	}
	return ""
}

func attrReaderFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	class, ok := receiver.(*object.RubyClass)
	if !ok {
		if mod, ok := receiver.(*object.RubyModule); ok {
			for _, arg := range args {
				name := getMethodName(arg)
				if name == "" {
					return newError("no implicit conversion of %s into Symbol", arg.Type())
				}
				// Create getter method
				mod.Methods[name] = createGetterMethod(name)
			}
			return object.NIL
		}
		return newError("attr_reader called on non-class/module")
	}

	for _, arg := range args {
		name := getMethodName(arg)
		if name == "" {
			return newError("no implicit conversion of %s into Symbol", arg.Type())
		}
		// Create getter method
		class.Methods[name] = createGetterMethod(name)
	}
	return object.NIL
}

func attrWriterFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	class, ok := receiver.(*object.RubyClass)
	if !ok {
		if mod, ok := receiver.(*object.RubyModule); ok {
			for _, arg := range args {
				name := getMethodName(arg)
				if name == "" {
					return newError("no implicit conversion of %s into Symbol", arg.Type())
				}
				// Create setter method
				mod.Methods[name+"="] = createSetterMethod(name)
			}
			return object.NIL
		}
		return newError("attr_writer called on non-class/module")
	}

	for _, arg := range args {
		name := getMethodName(arg)
		if name == "" {
			return newError("no implicit conversion of %s into Symbol", arg.Type())
		}
		// Create setter method
		class.Methods[name+"="] = createSetterMethod(name)
	}
	return object.NIL
}

func attrAccessorFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	attrReaderFn(receiver, env, args...)
	attrWriterFn(receiver, env, args...)
	return object.NIL
}

func createGetterMethod(name string) *object.Builtin {
	ivarName := "@" + name
	return &object.Builtin{
		Name: name,
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if instance, ok := receiver.(*object.Instance); ok {
				return instance.GetInstanceVariable(ivarName)
			}
			return object.NIL
		},
	}
}

func createSetterMethod(name string) *object.Builtin {
	ivarName := "@" + name
	return &object.Builtin{
		Name: name + "=",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			if instance, ok := receiver.(*object.Instance); ok {
				instance.SetInstanceVariable(ivarName, args[0])
				return args[0]
			}
			return newError("can't set instance variable on %s", receiver.Type())
		},
	}
}

func includeFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	class, classOk := receiver.(*object.RubyClass)
	mod, modOk := receiver.(*object.RubyModule)

	if !classOk && !modOk {
		return newError("include called on non-class/module")
	}

	for _, arg := range args {
		includedMod, ok := arg.(*object.RubyModule)
		if !ok {
			return newError("wrong argument type %s (expected Module)", arg.Type())
		}
		if classOk {
			class.IncludedModules = append(class.IncludedModules, includedMod)
		} else if modOk {
			// Copy methods from included module to this module
			for name, method := range includedMod.Methods {
				if _, exists := mod.Methods[name]; !exists {
					mod.Methods[name] = method
				}
			}
		}
	}
	return receiver
}

func extendFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	// extend adds module methods as singleton/class methods
	for _, arg := range args {
		mod, ok := arg.(*object.RubyModule)
		if !ok {
			return newError("wrong argument type %s (expected Module)", arg.Type())
		}

		switch recv := receiver.(type) {
		case *object.RubyClass:
			// Add module methods as class methods
			for name, method := range mod.Methods {
				recv.ClassMethods[name] = method
			}
		case *object.Instance:
			// For instances, we'd need singleton classes (not fully implemented)
			// For now, add to the class's class methods
			class := recv.Class_
			for name, method := range mod.Methods {
				class.ClassMethods[name] = method
			}
		}
	}
	return receiver
}

func prependFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	class, classOk := receiver.(*object.RubyClass)
	if !classOk {
		return newError("prepend called on non-class")
	}

	// Prepend inserts modules at the beginning of the lookup chain
	for i := len(args) - 1; i >= 0; i-- {
		mod, ok := args[i].(*object.RubyModule)
		if !ok {
			return newError("wrong argument type %s (expected Module)", args[i].Type())
		}
		// Insert at beginning
		class.IncludedModules = append([]*object.RubyModule{mod}, class.IncludedModules...)
	}
	return receiver
}

func defineMethodFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("wrong number of arguments (given 0, expected 1+)")
	}

	name := getMethodName(args[0])
	if name == "" {
		return newError("no implicit conversion of %s into Symbol", args[0].Type())
	}

	// Get the block or proc
	var proc *object.Proc
	if len(args) > 1 {
		// Second arg is a proc/lambda
		switch p := args[1].(type) {
		case *object.Proc:
			proc = p
		case *object.Lambda:
			proc = &object.Proc{
				Parameters: p.Parameters,
				Body:       p.Body,
				Env:        p.Env,
			}
		default:
			return newError("wrong argument type %s (expected Proc)", args[1].Type())
		}
	} else {
		// Use block
		proc = env.Block()
		if proc == nil {
			return newError("tried to create Proc object without a block")
		}
	}

	// Convert proc to method
	method := &object.Method{
		Name:       name,
		Parameters: convertBlockParamsToMethodParams(proc.Parameters),
		Body:       proc.Body,
		Env:        proc.Env,
	}

	switch recv := receiver.(type) {
	case *object.RubyClass:
		recv.Methods[name] = method
	case *object.RubyModule:
		recv.Methods[name] = method
	default:
		return newError("define_method called on non-class/module")
	}

	return &object.Symbol{Value: name}
}

func convertBlockParamsToMethodParams(blockParams []*ast.BlockParameter) []*ast.MethodParameter {
	params := make([]*ast.MethodParameter, len(blockParams))
	for i, bp := range blockParams {
		params[i] = &ast.MethodParameter{
			Name:  bp.Name,
			Splat: bp.Splat,
		}
	}
	return params
}

func aliasMethodFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("wrong number of arguments (given %d, expected 2)", len(args))
	}

	newName := getMethodName(args[0])
	oldName := getMethodName(args[1])
	if newName == "" || oldName == "" {
		return newError("no implicit conversion into Symbol")
	}

	var methods map[string]object.Object
	switch recv := receiver.(type) {
	case *object.RubyClass:
		methods = recv.Methods
	case *object.RubyModule:
		methods = recv.Methods
	default:
		return newError("alias_method called on non-class/module")
	}

	if method, ok := methods[oldName]; ok {
		methods[newName] = method
		return &object.Symbol{Value: newName}
	}

	return newError("undefined method `%s'", oldName)
}

func setVisibility(receiver object.Object, env *object.Environment, visibility object.MethodVisibility, args ...object.Object) object.Object {
	// If no args, set default visibility for subsequent method definitions
	if len(args) == 0 {
		env.SetCurrentVisibility(visibility)
		return receiver
	}

	// With args, change visibility of specific methods
	var methods map[string]object.Object
	switch recv := receiver.(type) {
	case *object.RubyClass:
		methods = recv.Methods
	case *object.RubyModule:
		methods = recv.Methods
	default:
		return object.NIL
	}

	for _, arg := range args {
		name := getMethodName(arg)
		if name == "" {
			continue
		}
		if method, ok := methods[name]; ok {
			if m, ok := method.(*object.Method); ok {
				m.Visibility = visibility
			}
		}
	}

	return args[0]
}

// refineFn implements Module#refine - creates a refinement for a class
func refineFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("wrong number of arguments (given 0, expected 1)")
	}

	// Must be called on a module
	mod, ok := receiver.(*object.RubyModule)
	if !ok {
		return newError("refine must be called on a module")
	}

	// First arg must be a class
	targetClass, ok := args[0].(*object.RubyClass)
	if !ok {
		return newError("wrong argument type %s (expected Class)", args[0].Type())
	}

	// Must have a block
	block := env.Block()
	if block == nil {
		return newError("no block given")
	}

	// Create or get the refinement for this class
	if mod.Refinements == nil {
		mod.Refinements = make(map[*object.RubyClass]*object.Refinement)
	}

	refinement, exists := mod.Refinements[targetClass]
	if !exists {
		refinement = &object.Refinement{
			TargetClass: targetClass,
			Methods:     make(map[string]object.Object),
		}
		mod.Refinements[targetClass] = refinement
	}

	// Evaluate the block to define methods
	// Methods defined in this block should go into the refinement
	refineEnv := object.NewEnclosedEnvironment(block.Env)
	refineEnv.SetSelf(refinement)

	// We need a special context for method definitions in refinements
	// Set a marker so evalMethodDefinition knows to add to refinement
	result := evalBlockBody(block.Body, refineEnv)

	// Copy methods from the environment/block execution to the refinement
	// This is handled by checking self in method definition

	return result
}

// usingFn implements Module#using - activates refinements in current scope
func usingFn(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("wrong number of arguments (given 0, expected 1)")
	}

	// First arg must be a module with refinements
	mod, ok := args[0].(*object.RubyModule)
	if !ok {
		return newError("wrong argument type %s (expected Module)", args[0].Type())
	}

	// Add to active refinements
	env.AddRefinement(mod)

	return object.NIL
}
