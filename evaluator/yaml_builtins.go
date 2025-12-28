package evaluator

import (
	"github.com/alexisbouchez/rubylexer/object"
	"gopkg.in/yaml.v3"
)

// YAMLModule represents Ruby's YAML module
var YAMLModule = &object.RubyModule{
	Name:      "YAML",
	Methods:   make(map[string]object.Object),
	Constants: make(map[string]object.Object),
}

func init() {
	initYAMLMethods()
}

func initYAMLMethods() {
	YAMLModule.Methods["load"] = &object.Builtin{
		Name: "load",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			var data interface{}
			if err := yaml.Unmarshal([]byte(str.Value), &data); err != nil {
				return newError("YAML parse error: %s", err.Error())
			}

			return yamlToRuby(data)
		},
	}

	YAMLModule.Methods["safe_load"] = YAMLModule.Methods["load"]
	YAMLModule.Methods["parse"] = YAMLModule.Methods["load"]

	YAMLModule.Methods["dump"] = &object.Builtin{
		Name: "dump",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			return rubyToYAML(args[0])
		},
	}

	YAMLModule.Methods["load_file"] = &object.Builtin{
		Name: "load_file",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			// Read file
			content := FileClass.ClassMethods["read"].(*object.Builtin).Fn(nil, env, filename)
			if err, isErr := content.(*object.Error); isErr {
				return err
			}

			// Parse YAML
			return YAMLModule.Methods["load"].(*object.Builtin).Fn(nil, env, content)
		},
	}
}

// yamlToRuby converts a Go value from YAML to a Ruby object
func yamlToRuby(data interface{}) object.Object {
	switch v := data.(type) {
	case nil:
		return object.NIL
	case bool:
		return object.NativeToBool(v)
	case int:
		return &object.Integer{Value: int64(v)}
	case int64:
		return &object.Integer{Value: v}
	case float64:
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, elem := range v {
			elements[i] = yamlToRuby(elem)
		}
		return &object.Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		order := make([]object.HashKey, 0, len(v))
		for key, val := range v {
			keyObj := &object.String{Value: key}
			hashed := keyObj.HashKey()
			pairs[hashed] = object.HashPair{Key: keyObj, Value: yamlToRuby(val)}
			order = append(order, hashed)
		}
		return &object.Hash{Pairs: pairs, Order: order}
	case map[interface{}]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		order := make([]object.HashKey, 0, len(v))
		for key, val := range v {
			keyObj := yamlToRuby(key)
			if hashable, ok := keyObj.(object.Hashable); ok {
				hashed := hashable.HashKey()
				pairs[hashed] = object.HashPair{Key: keyObj, Value: yamlToRuby(val)}
				order = append(order, hashed)
			}
		}
		return &object.Hash{Pairs: pairs, Order: order}
	default:
		return &object.String{Value: ""}
	}
}

// rubyToYAML converts a Ruby object to YAML string
func rubyToYAML(obj object.Object) object.Object {
	data := rubyToGo(obj)
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return newError("YAML dump error: %s", err.Error())
	}
	return &object.String{Value: string(bytes)}
}
