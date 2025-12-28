package evaluator

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alexisbouchez/rubylexer/object"
)

var (
	timeBuiltinsOnce sync.Once
	dateBuiltinsOnce sync.Once

	timeBuiltinsMap map[string]*object.Builtin
	dateBuiltinsMap map[string]*object.Builtin
)

// getTimeBuiltins returns instance methods for Time
func getTimeBuiltins() map[string]*object.Builtin {
	timeBuiltinsOnce.Do(func() {
		timeBuiltinsMap = make(map[string]*object.Builtin)
		for name, method := range TimeClass.Methods {
			if builtin, ok := method.(*object.Builtin); ok {
				timeBuiltinsMap[name] = builtin
			}
		}
	})
	return timeBuiltinsMap
}

// getDateBuiltins returns instance methods for Date
func getDateBuiltins() map[string]*object.Builtin {
	dateBuiltinsOnce.Do(func() {
		dateBuiltinsMap = make(map[string]*object.Builtin)
		for name, method := range DateClass.Methods {
			if builtin, ok := method.(*object.Builtin); ok {
				dateBuiltinsMap[name] = builtin
			}
		}
	})
	return dateBuiltinsMap
}

// TimeClass represents Ruby's Time class
var TimeClass = &object.RubyClass{
	Name:         "Time",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

// DateClass represents Ruby's Date class
var DateClass = &object.RubyClass{
	Name:         "Date",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

func init() {
	initTimeClassMethods()
	initTimeInstanceMethods()
	initDateClassMethods()
	initDateInstanceMethods()
}

func initTimeClassMethods() {
	TimeClass.ClassMethods["now"] = &object.Builtin{
		Name: "now",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			return &object.Time{Value: time.Now()}
		},
	}

	TimeClass.ClassMethods["new"] = &object.Builtin{
		Name: "new",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) == 0 {
				return &object.Time{Value: time.Now()}
			}

			year, month, day := 0, 1, 1
			hour, min, sec := 0, 0, 0

			if len(args) > 0 {
				if i, ok := args[0].(*object.Integer); ok {
					year = int(i.Value)
				}
			}
			if len(args) > 1 {
				if i, ok := args[1].(*object.Integer); ok {
					month = int(i.Value)
				}
			}
			if len(args) > 2 {
				if i, ok := args[2].(*object.Integer); ok {
					day = int(i.Value)
				}
			}
			if len(args) > 3 {
				if i, ok := args[3].(*object.Integer); ok {
					hour = int(i.Value)
				}
			}
			if len(args) > 4 {
				if i, ok := args[4].(*object.Integer); ok {
					min = int(i.Value)
				}
			}
			if len(args) > 5 {
				if i, ok := args[5].(*object.Integer); ok {
					sec = int(i.Value)
				}
			}

			t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
			return &object.Time{Value: t}
		},
	}

	TimeClass.ClassMethods["at"] = &object.Builtin{
		Name: "at",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			switch ts := args[0].(type) {
			case *object.Integer:
				return &object.Time{Value: time.Unix(ts.Value, 0)}
			case *object.Float:
				sec := int64(ts.Value)
				nsec := int64((ts.Value - float64(sec)) * 1e9)
				return &object.Time{Value: time.Unix(sec, nsec)}
			default:
				return newError("no implicit conversion of %s into Integer", args[0].Type())
			}
		},
	}

	TimeClass.ClassMethods["utc"] = &object.Builtin{
		Name: "utc",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			year, month, day := 0, 1, 1
			hour, min, sec := 0, 0, 0

			if len(args) > 0 {
				if i, ok := args[0].(*object.Integer); ok {
					year = int(i.Value)
				}
			}
			if len(args) > 1 {
				if i, ok := args[1].(*object.Integer); ok {
					month = int(i.Value)
				}
			}
			if len(args) > 2 {
				if i, ok := args[2].(*object.Integer); ok {
					day = int(i.Value)
				}
			}
			if len(args) > 3 {
				if i, ok := args[3].(*object.Integer); ok {
					hour = int(i.Value)
				}
			}
			if len(args) > 4 {
				if i, ok := args[4].(*object.Integer); ok {
					min = int(i.Value)
				}
			}
			if len(args) > 5 {
				if i, ok := args[5].(*object.Integer); ok {
					sec = int(i.Value)
				}
			}

			t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
			return &object.Time{Value: t}
		},
	}

	TimeClass.ClassMethods["parse"] = &object.Builtin{
		Name: "parse",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			// Try common formats
			formats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05",
				"2006-01-02 15:04:05 -0700",
				"2006-01-02T15:04:05",
				"2006-01-02",
				"01/02/2006",
				"Jan 2, 2006",
				"January 2, 2006",
			}

			for _, format := range formats {
				if t, err := time.Parse(format, str.Value); err == nil {
					return &object.Time{Value: t}
				}
			}

			return newError("no time information in %q", str.Value)
		},
	}
}

func initTimeInstanceMethods() {
	TimeClass.Methods["year"] = &object.Builtin{
		Name: "year",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Year())}
		},
	}

	TimeClass.Methods["month"] = &object.Builtin{
		Name: "month",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Month())}
		},
	}

	TimeClass.Methods["mon"] = TimeClass.Methods["month"]

	TimeClass.Methods["day"] = &object.Builtin{
		Name: "day",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Day())}
		},
	}

	TimeClass.Methods["mday"] = TimeClass.Methods["day"]

	TimeClass.Methods["hour"] = &object.Builtin{
		Name: "hour",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Hour())}
		},
	}

	TimeClass.Methods["min"] = &object.Builtin{
		Name: "min",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Minute())}
		},
	}

	TimeClass.Methods["sec"] = &object.Builtin{
		Name: "sec",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Second())}
		},
	}

	TimeClass.Methods["nsec"] = &object.Builtin{
		Name: "nsec",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Nanosecond())}
		},
	}

	TimeClass.Methods["wday"] = &object.Builtin{
		Name: "wday",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.Weekday())}
		},
	}

	TimeClass.Methods["yday"] = &object.Builtin{
		Name: "yday",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: int64(t.Value.YearDay())}
		},
	}

	TimeClass.Methods["to_i"] = &object.Builtin{
		Name: "to_i",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Integer{Value: t.Value.Unix()}
		},
	}

	TimeClass.Methods["to_f"] = &object.Builtin{
		Name: "to_f",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Float{Value: float64(t.Value.UnixNano()) / 1e9}
		},
	}

	TimeClass.Methods["to_s"] = &object.Builtin{
		Name: "to_s",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.String{Value: t.Value.Format("2006-01-02 15:04:05 -0700")}
		},
	}

	TimeClass.Methods["inspect"] = TimeClass.Methods["to_s"]

	TimeClass.Methods["strftime"] = &object.Builtin{
		Name: "strftime",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			format, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			t := receiver.(*object.Time)
			return &object.String{Value: rubyStrftime(t.Value, format.Value)}
		},
	}

	TimeClass.Methods["utc"] = &object.Builtin{
		Name: "utc",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Time{Value: t.Value.UTC()}
		},
	}

	TimeClass.Methods["getutc"] = TimeClass.Methods["utc"]
	TimeClass.Methods["gmtime"] = TimeClass.Methods["utc"]

	TimeClass.Methods["localtime"] = &object.Builtin{
		Name: "localtime",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return &object.Time{Value: t.Value.Local()}
		},
	}

	TimeClass.Methods["utc?"] = &object.Builtin{
		Name: "utc?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Location() == time.UTC)
		},
	}

	TimeClass.Methods["zone"] = &object.Builtin{
		Name: "zone",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			zone, _ := t.Value.Zone()
			return &object.String{Value: zone}
		},
	}

	TimeClass.Methods["utc_offset"] = &object.Builtin{
		Name: "utc_offset",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			_, offset := t.Value.Zone()
			return &object.Integer{Value: int64(offset)}
		},
	}

	TimeClass.Methods["+"] = &object.Builtin{
		Name: "+",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			t := receiver.(*object.Time)
			switch secs := args[0].(type) {
			case *object.Integer:
				return &object.Time{Value: t.Value.Add(time.Duration(secs.Value) * time.Second)}
			case *object.Float:
				return &object.Time{Value: t.Value.Add(time.Duration(secs.Value * float64(time.Second)))}
			default:
				return newError("can't convert %s into exact number", args[0].Type())
			}
		},
	}

	TimeClass.Methods["-"] = &object.Builtin{
		Name: "-",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			t := receiver.(*object.Time)
			switch other := args[0].(type) {
			case *object.Integer:
				return &object.Time{Value: t.Value.Add(-time.Duration(other.Value) * time.Second)}
			case *object.Float:
				return &object.Time{Value: t.Value.Add(-time.Duration(other.Value * float64(time.Second)))}
			case *object.Time:
				diff := t.Value.Sub(other.Value)
				return &object.Float{Value: diff.Seconds()}
			default:
				return newError("can't convert %s into exact number", args[0].Type())
			}
		},
	}

	TimeClass.Methods["<"] = &object.Builtin{
		Name: "<",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return newError("comparison of Time with %s failed", args[0].Type())
			}
			return object.NativeToBool(t.Value.Before(other.Value))
		},
	}

	TimeClass.Methods[">"] = &object.Builtin{
		Name: ">",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return newError("comparison of Time with %s failed", args[0].Type())
			}
			return object.NativeToBool(t.Value.After(other.Value))
		},
	}

	TimeClass.Methods["<="] = &object.Builtin{
		Name: "<=",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return newError("comparison of Time with %s failed", args[0].Type())
			}
			return object.NativeToBool(!t.Value.After(other.Value))
		},
	}

	TimeClass.Methods[">="] = &object.Builtin{
		Name: ">=",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return newError("comparison of Time with %s failed", args[0].Type())
			}
			return object.NativeToBool(!t.Value.Before(other.Value))
		},
	}

	TimeClass.Methods["=="] = &object.Builtin{
		Name: "==",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return object.FALSE
			}
			return object.NativeToBool(t.Value.Equal(other.Value))
		},
	}

	TimeClass.Methods["<=>"] = &object.Builtin{
		Name: "<=>",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.NIL
			}
			t := receiver.(*object.Time)
			other, ok := args[0].(*object.Time)
			if !ok {
				return object.NIL
			}
			if t.Value.Before(other.Value) {
				return &object.Integer{Value: -1}
			} else if t.Value.After(other.Value) {
				return &object.Integer{Value: 1}
			}
			return &object.Integer{Value: 0}
		},
	}

	TimeClass.Methods["sunday?"] = &object.Builtin{
		Name: "sunday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Sunday)
		},
	}

	TimeClass.Methods["monday?"] = &object.Builtin{
		Name: "monday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Monday)
		},
	}

	TimeClass.Methods["tuesday?"] = &object.Builtin{
		Name: "tuesday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Tuesday)
		},
	}

	TimeClass.Methods["wednesday?"] = &object.Builtin{
		Name: "wednesday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Wednesday)
		},
	}

	TimeClass.Methods["thursday?"] = &object.Builtin{
		Name: "thursday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Thursday)
		},
	}

	TimeClass.Methods["friday?"] = &object.Builtin{
		Name: "friday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Friday)
		},
	}

	TimeClass.Methods["saturday?"] = &object.Builtin{
		Name: "saturday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			t := receiver.(*object.Time)
			return object.NativeToBool(t.Value.Weekday() == time.Saturday)
		},
	}
}

func initDateClassMethods() {
	DateClass.ClassMethods["today"] = &object.Builtin{
		Name: "today",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			now := time.Now()
			t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			return &object.Date{Value: t}
		},
	}

	DateClass.ClassMethods["new"] = &object.Builtin{
		Name: "new",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			year, month, day := -4712, 1, 1 // Ruby Date default

			if len(args) > 0 {
				if i, ok := args[0].(*object.Integer); ok {
					year = int(i.Value)
				}
			}
			if len(args) > 1 {
				if i, ok := args[1].(*object.Integer); ok {
					month = int(i.Value)
				}
			}
			if len(args) > 2 {
				if i, ok := args[2].(*object.Integer); ok {
					day = int(i.Value)
				}
			}

			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			return &object.Date{Value: t}
		},
	}

	DateClass.ClassMethods["parse"] = &object.Builtin{
		Name: "parse",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			formats := []string{
				"2006-01-02",
				"01/02/2006",
				"02/01/2006",
				"Jan 2, 2006",
				"January 2, 2006",
				"2 Jan 2006",
				"2006/01/02",
			}

			for _, format := range formats {
				if t, err := time.Parse(format, str.Value); err == nil {
					return &object.Date{Value: t}
				}
			}

			return newError("invalid date: %q", str.Value)
		},
	}
}

func initDateInstanceMethods() {
	DateClass.Methods["year"] = &object.Builtin{
		Name: "year",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Integer{Value: int64(d.Value.Year())}
		},
	}

	DateClass.Methods["month"] = &object.Builtin{
		Name: "month",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Integer{Value: int64(d.Value.Month())}
		},
	}

	DateClass.Methods["mon"] = DateClass.Methods["month"]

	DateClass.Methods["day"] = &object.Builtin{
		Name: "day",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Integer{Value: int64(d.Value.Day())}
		},
	}

	DateClass.Methods["mday"] = DateClass.Methods["day"]

	DateClass.Methods["wday"] = &object.Builtin{
		Name: "wday",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Integer{Value: int64(d.Value.Weekday())}
		},
	}

	DateClass.Methods["yday"] = &object.Builtin{
		Name: "yday",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Integer{Value: int64(d.Value.YearDay())}
		},
	}

	DateClass.Methods["to_s"] = &object.Builtin{
		Name: "to_s",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.String{Value: d.Value.Format("2006-01-02")}
		},
	}

	DateClass.Methods["inspect"] = DateClass.Methods["to_s"]

	DateClass.Methods["strftime"] = &object.Builtin{
		Name: "strftime",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			format, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			d := receiver.(*object.Date)
			return &object.String{Value: rubyStrftime(d.Value, format.Value)}
		},
	}

	DateClass.Methods["+"] = &object.Builtin{
		Name: "+",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			d := receiver.(*object.Date)
			days, ok := args[0].(*object.Integer)
			if !ok {
				return newError("expected Integer")
			}
			return &object.Date{Value: d.Value.AddDate(0, 0, int(days.Value))}
		},
	}

	DateClass.Methods["-"] = &object.Builtin{
		Name: "-",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			d := receiver.(*object.Date)
			switch other := args[0].(type) {
			case *object.Integer:
				return &object.Date{Value: d.Value.AddDate(0, 0, -int(other.Value))}
			case *object.Date:
				diff := d.Value.Sub(other.Value)
				return &object.Integer{Value: int64(diff.Hours() / 24)}
			default:
				return newError("expected Integer or Date")
			}
		},
	}

	DateClass.Methods["<"] = &object.Builtin{
		Name: "<",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			d := receiver.(*object.Date)
			other, ok := args[0].(*object.Date)
			if !ok {
				return newError("comparison of Date with %s failed", args[0].Type())
			}
			return object.NativeToBool(d.Value.Before(other.Value))
		},
	}

	DateClass.Methods[">"] = &object.Builtin{
		Name: ">",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			d := receiver.(*object.Date)
			other, ok := args[0].(*object.Date)
			if !ok {
				return newError("comparison of Date with %s failed", args[0].Type())
			}
			return object.NativeToBool(d.Value.After(other.Value))
		},
	}

	DateClass.Methods["=="] = &object.Builtin{
		Name: "==",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.FALSE
			}
			d := receiver.(*object.Date)
			other, ok := args[0].(*object.Date)
			if !ok {
				return object.FALSE
			}
			return object.NativeToBool(d.Value.Equal(other.Value))
		},
	}

	DateClass.Methods["sunday?"] = &object.Builtin{
		Name: "sunday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return object.NativeToBool(d.Value.Weekday() == time.Sunday)
		},
	}

	DateClass.Methods["monday?"] = &object.Builtin{
		Name: "monday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return object.NativeToBool(d.Value.Weekday() == time.Monday)
		},
	}

	DateClass.Methods["saturday?"] = &object.Builtin{
		Name: "saturday?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return object.NativeToBool(d.Value.Weekday() == time.Saturday)
		},
	}

	DateClass.Methods["to_time"] = &object.Builtin{
		Name: "to_time",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			d := receiver.(*object.Date)
			return &object.Time{Value: d.Value}
		},
	}
}

// rubyStrftime converts Ruby strftime format to Go time
func rubyStrftime(t time.Time, format string) string {
	// Ruby strftime -> Go format replacements
	replacements := map[string]string{
		"%Y": "2006",
		"%y": "06",
		"%m": "01",
		"%d": "02",
		"%e": "_2",
		"%H": "15",
		"%I": "03",
		"%M": "04",
		"%S": "05",
		"%p": "PM",
		"%P": "pm",
		"%Z": "MST",
		"%z": "-0700",
		"%A": "Monday",
		"%a": "Mon",
		"%B": "January",
		"%b": "Jan",
		"%j": "", // Day of year - handle separately
		"%w": "", // Weekday - handle separately
		"%U": "", // Week number - handle separately
		"%W": "", // Week number - handle separately
		"%%": "%",
		"%n": "\n",
		"%t": "\t",
	}

	result := format

	// Handle special cases first
	result = strings.ReplaceAll(result, "%j", fmt.Sprintf("%03d", t.YearDay()))
	result = strings.ReplaceAll(result, "%w", strconv.Itoa(int(t.Weekday())))
	_, week := t.ISOWeek()
	result = strings.ReplaceAll(result, "%W", fmt.Sprintf("%02d", week))
	result = strings.ReplaceAll(result, "%U", fmt.Sprintf("%02d", week))

	for rubyFmt, goFmt := range replacements {
		if goFmt != "" {
			result = strings.ReplaceAll(result, rubyFmt, t.Format(goFmt))
		}
	}

	return result
}
