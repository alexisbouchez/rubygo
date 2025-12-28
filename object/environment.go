package object

// Environment holds variable bindings.
type Environment struct {
	store             map[string]Object
	outer             *Environment
	constants         map[string]Object
	self              Object
	block             *Proc
	currentClass      *RubyClass
	currentModule     *RubyModule
	singletonTarget   Object           // Target object for singleton class (class << obj)
	currentMethod     string           // Current method name (for super)
	methodArgs        []Object         // Original method arguments (for super without args)
	definingClass     *RubyClass       // Class where current method is defined
	currentVisibility MethodVisibility // Current visibility for method definitions
	visibilitySet     bool             // Whether visibility was explicitly set
	activeRefinements []*RubyModule    // Active refinements in lexical scope
}

// NewEnvironment creates a new environment.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	c := make(map[string]Object)
	return &Environment{store: s, outer: nil, constants: c}
}

// NewEnclosedEnvironment creates an enclosed environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a variable from the environment.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set sets a variable in the current environment.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// SetLocal sets a variable in the local environment only (no lookup in outer scopes).
func (e *Environment) SetLocal(name string, val Object) Object {
	e.store[name] = val
	return val
}

// Update updates a variable, looking up the scope where it was defined.
func (e *Environment) Update(name string, val Object) Object {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	// Variable not found, set in current scope
	e.store[name] = val
	return val
}

// GetConstant retrieves a constant.
func (e *Environment) GetConstant(name string) (Object, bool) {
	obj, ok := e.constants[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.GetConstant(name)
	}
	return obj, ok
}

// SetConstant sets a constant.
func (e *Environment) SetConstant(name string, val Object) Object {
	e.constants[name] = val
	return val
}

// Self returns the current self object.
func (e *Environment) Self() Object {
	if e.self != nil {
		return e.self
	}
	if e.outer != nil {
		return e.outer.Self()
	}
	return nil
}

// SetSelf sets the self object.
func (e *Environment) SetSelf(self Object) {
	e.self = self
}

// Block returns the current block.
func (e *Environment) Block() *Proc {
	if e.block != nil {
		return e.block
	}
	if e.outer != nil {
		return e.outer.Block()
	}
	return nil
}

// SetBlock sets the current block.
func (e *Environment) SetBlock(block *Proc) {
	e.block = block
}

// Outer returns the outer environment.
func (e *Environment) Outer() *Environment {
	return e.outer
}

// CurrentClass returns the current class context for method definitions.
func (e *Environment) CurrentClass() *RubyClass {
	if e.currentClass != nil {
		return e.currentClass
	}
	if e.outer != nil {
		return e.outer.CurrentClass()
	}
	return nil
}

// SetCurrentClass sets the current class context.
func (e *Environment) SetCurrentClass(class *RubyClass) {
	e.currentClass = class
}

// CurrentModule returns the current module context for method definitions.
func (e *Environment) CurrentModule() *RubyModule {
	if e.currentModule != nil {
		return e.currentModule
	}
	if e.outer != nil {
		return e.outer.CurrentModule()
	}
	return nil
}

// SetCurrentModule sets the current module context.
func (e *Environment) SetCurrentModule(mod *RubyModule) {
	e.currentModule = mod
}

// SingletonTarget returns the singleton target object (for class << obj).
func (e *Environment) SingletonTarget() Object {
	if e.singletonTarget != nil {
		return e.singletonTarget
	}
	if e.outer != nil {
		return e.outer.SingletonTarget()
	}
	return nil
}

// SetSingletonTarget sets the singleton target object.
func (e *Environment) SetSingletonTarget(obj Object) {
	e.singletonTarget = obj
}

// CurrentMethod returns the current method name (for super calls).
func (e *Environment) CurrentMethod() string {
	if e.currentMethod != "" {
		return e.currentMethod
	}
	if e.outer != nil {
		return e.outer.CurrentMethod()
	}
	return ""
}

// SetCurrentMethod sets the current method name.
func (e *Environment) SetCurrentMethod(name string) {
	e.currentMethod = name
}

// MethodArgs returns the original method arguments (for super without args).
func (e *Environment) MethodArgs() []Object {
	if e.methodArgs != nil {
		return e.methodArgs
	}
	if e.outer != nil {
		return e.outer.MethodArgs()
	}
	return nil
}

// SetMethodArgs sets the original method arguments.
func (e *Environment) SetMethodArgs(args []Object) {
	e.methodArgs = args
}

// DefiningClass returns the class where the current method is defined.
func (e *Environment) DefiningClass() *RubyClass {
	if e.definingClass != nil {
		return e.definingClass
	}
	if e.outer != nil {
		return e.outer.DefiningClass()
	}
	return nil
}

// SetDefiningClass sets the class where the current method is defined.
func (e *Environment) SetDefiningClass(class *RubyClass) {
	e.definingClass = class
}

// CurrentVisibility returns the current visibility for method definitions.
func (e *Environment) CurrentVisibility() MethodVisibility {
	if e.visibilitySet {
		return e.currentVisibility
	}
	if e.outer != nil {
		return e.outer.CurrentVisibility()
	}
	return VisibilityPublic
}

// SetCurrentVisibility sets the current visibility for method definitions.
func (e *Environment) SetCurrentVisibility(v MethodVisibility) {
	e.currentVisibility = v
	e.visibilitySet = true
}

// LocalVariableNames returns a list of all local variable names in this environment.
func (e *Environment) LocalVariableNames() []string {
	names := make([]string, 0, len(e.store))
	for name := range e.store {
		names = append(names, name)
	}
	return names
}

// ActiveRefinements returns all active refinements in the current lexical scope.
func (e *Environment) ActiveRefinements() []*RubyModule {
	if e.activeRefinements != nil {
		return e.activeRefinements
	}
	if e.outer != nil {
		return e.outer.ActiveRefinements()
	}
	return nil
}

// AddRefinement adds a module's refinements to the active refinements.
func (e *Environment) AddRefinement(mod *RubyModule) {
	if e.activeRefinements == nil {
		// Copy from parent if exists
		if e.outer != nil {
			parent := e.outer.ActiveRefinements()
			if parent != nil {
				e.activeRefinements = make([]*RubyModule, len(parent))
				copy(e.activeRefinements, parent)
			} else {
				e.activeRefinements = make([]*RubyModule, 0)
			}
		} else {
			e.activeRefinements = make([]*RubyModule, 0)
		}
	}
	e.activeRefinements = append(e.activeRefinements, mod)
}

// LookupRefinedMethod looks for a method in active refinements for the given class.
func (e *Environment) LookupRefinedMethod(class *RubyClass, methodName string) (Object, bool) {
	refinements := e.ActiveRefinements()
	if refinements == nil {
		return nil, false
	}
	// Check refinements in reverse order (most recently added first)
	for i := len(refinements) - 1; i >= 0; i-- {
		mod := refinements[i]
		if mod.Refinements != nil {
			if ref, ok := mod.Refinements[class]; ok {
				if method, ok := ref.Methods[methodName]; ok {
					return method, true
				}
			}
		}
	}
	return nil, false
}
