package evaluator

import (
	"sync"

	"github.com/alexisbouchez/rubylexer/object"
)

var tracePointBuiltinsOnce sync.Once
var tracePointBuiltinsMap map[string]*object.Builtin

func getTracePointBuiltins() map[string]*object.Builtin {
	tracePointBuiltinsOnce.Do(func() {
		tracePointBuiltinsMap = map[string]*object.Builtin{
			"enable": {
				Name: "enable",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if !tp.Enabled {
						tp.Enabled = true
						object.AddActiveTracePoint(tp)
					}

					// If block given, enable only for the block
					block := env.Block()
					if block != nil {
						result := evalBlockBody(block.Body, block.Env)
						tp.Enabled = false
						object.RemoveActiveTracePoint(tp)
						return result
					}

					return tp
				},
			},
			"disable": {
				Name: "disable",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					wasEnabled := tp.Enabled
					if tp.Enabled {
						tp.Enabled = false
						object.RemoveActiveTracePoint(tp)
					}

					// If block given, disable only for the block
					block := env.Block()
					if block != nil {
						result := evalBlockBody(block.Body, block.Env)
						if wasEnabled {
							tp.Enabled = true
							object.AddActiveTracePoint(tp)
						}
						return result
					}

					return object.NativeToBool(wasEnabled)
				},
			},
			"enabled?": {
				Name: "enabled?",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					return object.NativeToBool(tp.Enabled)
				},
			},
			"event": {
				Name: "event",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					return &object.Symbol{Value: string(tp.Event)}
				},
			},
			"method_id": {
				Name: "method_id",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if tp.MethodID == "" {
						return object.NIL
					}
					return &object.Symbol{Value: tp.MethodID}
				},
			},
			"path": {
				Name: "path",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if tp.Path == "" {
						return object.NIL
					}
					return &object.String{Value: tp.Path}
				},
			},
			"lineno": {
				Name: "lineno",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					return &object.Integer{Value: int64(tp.LineNo)}
				},
			},
			"self": {
				Name: "self",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if tp.Self_ == nil {
						return object.NIL
					}
					return tp.Self_
				},
			},
			"return_value": {
				Name: "return_value",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if tp.ReturnVal == nil {
						return object.NIL
					}
					return tp.ReturnVal
				},
			},
			"raised_exception": {
				Name: "raised_exception",
				Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
					tp := receiver.(*object.TracePoint)
					if tp.RaisedExc == nil {
						return object.NIL
					}
					return tp.RaisedExc
				},
			},
		}
	})
	return tracePointBuiltinsMap
}

// TracePointNew creates a new TracePoint (called as TracePoint.new)
func TracePointNew(env *object.Environment, args []object.Object) object.Object {
	block := env.Block()
	if block == nil {
		return newError("must be called with a block")
	}

	tp := &object.TracePoint{
		Events:  make([]object.TracePointEvent, 0, len(args)),
		Block:   block,
		Enabled: false,
	}

	// Parse event types
	for _, arg := range args {
		sym, ok := arg.(*object.Symbol)
		if !ok {
			return newError("wrong argument type %s (expected Symbol)", arg.Type())
		}
		switch sym.Value {
		case "call":
			tp.Events = append(tp.Events, object.TraceEventCall)
		case "return":
			tp.Events = append(tp.Events, object.TraceEventReturn)
		case "line":
			tp.Events = append(tp.Events, object.TraceEventLine)
		case "raise":
			tp.Events = append(tp.Events, object.TraceEventRaise)
		case "b_call":
			tp.Events = append(tp.Events, object.TraceEventBCall)
		case "b_return":
			tp.Events = append(tp.Events, object.TraceEventBReturn)
		case "class":
			tp.Events = append(tp.Events, object.TraceEventClass)
		case "end":
			tp.Events = append(tp.Events, object.TraceEventEnd)
		default:
			return newError("unknown event: %s", sym.Value)
		}
	}

	// If no events specified, trace all
	if len(tp.Events) == 0 {
		tp.Events = []object.TracePointEvent{
			object.TraceEventCall,
			object.TraceEventReturn,
			object.TraceEventLine,
			object.TraceEventRaise,
		}
	}

	return tp
}

// FireTraceEvent fires trace events to all active trace points
func FireTraceEvent(event object.TracePointEvent, methodID, path string, lineno int, self, returnVal, raisedExc object.Object, env *object.Environment) {
	tracePoints := object.GetActiveTracePoints()
	for _, tp := range tracePoints {
		if !tp.Enabled {
			continue
		}

		// Check if this trace point is interested in this event
		found := false
		for _, e := range tp.Events {
			if e == event {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		// Set event info
		tp.Event = event
		tp.MethodID = methodID
		tp.Path = path
		tp.LineNo = lineno
		tp.Self_ = self
		tp.ReturnVal = returnVal
		tp.RaisedExc = raisedExc

		// Call the block with the trace point as argument
		if tp.Block != nil {
			blockEnv := object.NewEnclosedEnvironment(tp.Block.Env)
			if len(tp.Block.Parameters) > 0 {
				blockEnv.Set(tp.Block.Parameters[0].Name, tp)
			}
			evalBlockBody(tp.Block.Body, blockEnv)
		}
	}
}
