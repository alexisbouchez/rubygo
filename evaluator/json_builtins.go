package evaluator

import (
	"encoding/json"
	"fmt"

	"github.com/alexisbouchez/rubylexer/object"
)

// JSONModule represents Ruby's JSON module
var JSONModule = &object.RubyModule{
	Name:      "JSON",
	Methods:   make(map[string]object.Object),
	Constants: make(map[string]object.Object),
}

func init() {
	initJSONMethods()
}

func initJSONMethods() {
	JSONModule.Methods["parse"] = &object.Builtin{
		Name: "parse",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			var data interface{}
			if err := json.Unmarshal([]byte(str.Value), &data); err != nil {
				return newError("JSON parse error: %s", err.Error())
			}

			return jsonToRuby(data)
		},
	}

	JSONModule.Methods["generate"] = &object.Builtin{
		Name: "generate",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			return rubyToJSON(args[0], false)
		},
	}

	JSONModule.Methods["dump"] = JSONModule.Methods["generate"]

	JSONModule.Methods["pretty_generate"] = &object.Builtin{
		Name: "pretty_generate",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			return rubyToJSON(args[0], true)
		},
	}
}

// jsonToRuby converts a Go value from JSON to a Ruby object
func jsonToRuby(data interface{}) object.Object {
	switch v := data.(type) {
	case nil:
		return object.NIL
	case bool:
		return object.NativeToBool(v)
	case float64:
		// JSON numbers are float64, check if it's an integer
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, elem := range v {
			elements[i] = jsonToRuby(elem)
		}
		return &object.Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		order := make([]object.HashKey, 0, len(v))
		for key, val := range v {
			keyObj := &object.String{Value: key}
			hashed := keyObj.HashKey()
			pairs[hashed] = object.HashPair{Key: keyObj, Value: jsonToRuby(val)}
			order = append(order, hashed)
		}
		return &object.Hash{Pairs: pairs, Order: order}
	default:
		return newError("unknown JSON type: %T", data)
	}
}

// rubyToJSON converts a Ruby object to JSON string
func rubyToJSON(obj object.Object, pretty bool) object.Object {
	data := rubyToGo(obj)
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		return newError("JSON generate error: %s", err.Error())
	}
	return &object.String{Value: string(bytes)}
}

// rubyToGo converts a Ruby object to a Go value for JSON encoding
func rubyToGo(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Nil:
		return nil
	case *object.Boolean:
		return o.Value
	case *object.Integer:
		return o.Value
	case *object.Float:
		return o.Value
	case *object.String:
		return o.Value
	case *object.Symbol:
		return o.Value
	case *object.Array:
		result := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			result[i] = rubyToGo(elem)
		}
		return result
	case *object.Hash:
		result := make(map[string]interface{})
		for _, hk := range o.Order {
			pair := o.Pairs[hk]
			keyStr := ""
			switch k := pair.Key.(type) {
			case *object.String:
				keyStr = k.Value
			case *object.Symbol:
				keyStr = k.Value
			default:
				keyStr = fmt.Sprintf("%v", pair.Key.Inspect())
			}
			result[keyStr] = rubyToGo(pair.Value)
		}
		return result
	case *object.Time:
		return o.Value.Format("2006-01-02T15:04:05Z07:00")
	case *object.Date:
		return o.Value.Format("2006-01-02")
	default:
		return obj.Inspect()
	}
}

// AddToJSONMethod adds to_json method to object builtins
func getToJSONBuiltin() *object.Builtin {
	return &object.Builtin{
		Name: "to_json",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			return rubyToJSON(receiver, false)
		},
	}
}
