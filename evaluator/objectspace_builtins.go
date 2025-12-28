package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/object"
)

var objectSpaceModuleOnce sync.Once
var objectSpaceModule *object.RubyModule

// GetObjectSpaceModule returns the ObjectSpace module.
func GetObjectSpaceModule() *object.RubyModule {
	objectSpaceModuleOnce.Do(func() {
		objectSpaceModule = &object.RubyModule{
			Name:      "ObjectSpace",
			Methods:   make(map[string]object.Object),
			Constants: make(map[string]object.Object),
		}

		// each_object - iterate over all objects of a given type
		objectSpaceModule.Methods["each_object"] = &object.Builtin{
			Name: "each_object",
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				block := env.Block()
				if block == nil {
					return newError("no block given")
				}

				var filterClass *object.RubyClass

				// If a class is given, filter by it
				if len(args) > 0 {
					if class, ok := args[0].(*object.RubyClass); ok {
						filterClass = class
					} else {
						return newError("class or module required")
					}
				}

				count := int64(0)
				objects := object.GetTrackedObjects()

				for _, obj := range objects {
					// Filter by class if specified
					if filterClass != nil {
						objClass := obj.Class()
						if objClass == nil || objClass != filterClass {
							// Check if it's an instance of the class
							if inst, ok := obj.(*object.Instance); ok {
								if inst.Class_ != filterClass {
									continue
								}
							} else {
								continue
							}
						}
					}

					// Call the block with the object
					blockEnv := object.NewEnclosedEnvironment(block.Env)
					if len(block.Parameters) > 0 {
						blockEnv.Set(block.Parameters[0].Name, obj)
					}
					evalBlockBody(block.Body, blockEnv)
					count++
				}

				return &object.Integer{Value: count}
			},
		}

		// count_objects - count objects by type
		objectSpaceModule.Methods["count_objects"] = &object.Builtin{
			Name: "count_objects",
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				counts := object.CountObjectsByType()

				hash := &object.Hash{
					Pairs: make(map[object.HashKey]object.HashPair),
					Order: make([]object.HashKey, 0),
				}

				for typ, count := range counts {
					key := &object.Symbol{Value: string(typ)}
					hashKey := key.HashKey()
					hash.Pairs[hashKey] = object.HashPair{
						Key:   key,
						Value: &object.Integer{Value: int64(count)},
					}
					hash.Order = append(hash.Order, hashKey)
				}

				return hash
			},
		}

		// _id2ref - get object by ID (not fully implemented since Go doesn't support this directly)
		objectSpaceModule.Methods["_id2ref"] = &object.Builtin{
			Name: "_id2ref",
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("wrong number of arguments (given 0, expected 1)")
				}

				id, ok := args[0].(*object.Integer)
				if !ok {
					return newError("not a valid object id")
				}

				// Search for object with this ID
				objects := object.GetTrackedObjects()
				for _, obj := range objects {
					if object.GetObjectID(obj) == id.Value {
						return obj
					}
				}

				return newError("RangeError: object not found for object id %d", id.Value)
			},
		}

		// garbage_collect - trigger GC (no-op in this implementation, Go handles GC)
		objectSpaceModule.Methods["garbage_collect"] = &object.Builtin{
			Name: "garbage_collect",
			Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
				// In Ruby, this triggers GC. In Go, we can't force GC explicitly in a meaningful way
				// but we can suggest it
				// runtime.GC() - we could call this but it's not usually needed
				return object.NIL
			},
		}
	})
	return objectSpaceModule
}
