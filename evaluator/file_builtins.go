package evaluator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexisbouchez/rubylexer/object"
)

var (
	fileBuiltinsOnce sync.Once
	dirBuiltinsOnce  sync.Once

	fileBuiltinsMap map[string]*object.Builtin
	dirBuiltinsMap  map[string]*object.Builtin
)

// FileClass represents Ruby's File class
var FileClass = &object.RubyClass{
	Name:         "File",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

// DirClass represents Ruby's Dir class
var DirClass = &object.RubyClass{
	Name:         "Dir",
	Superclass:   object.ObjectClass,
	Methods:      make(map[string]object.Object),
	ClassMethods: make(map[string]object.Object),
}

func init() {
	// Initialize File class methods
	initFileClassMethods()
	// Initialize Dir class methods
	initDirClassMethods()
}

func initFileClassMethods() {
	FileClass.ClassMethods["read"] = &object.Builtin{
		Name: "read",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			content, err := ioutil.ReadFile(filename.Value)
			if err != nil {
				return newError("No such file or directory @ rb_sysopen - %s", filename.Value)
			}
			return &object.String{Value: string(content)}
		},
	}

	FileClass.ClassMethods["write"] = &object.Builtin{
		Name: "write",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments (given %d, expected 2)", len(args))
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			content, ok := args[1].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[1].Type())
			}
			err := ioutil.WriteFile(filename.Value, []byte(content.Value), 0644)
			if err != nil {
				return newError("Permission denied @ rb_sysopen - %s", filename.Value)
			}
			return &object.Integer{Value: int64(len(content.Value))}
		},
	}

	FileClass.ClassMethods["exist?"] = &object.Builtin{
		Name: "exist?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			_, err := os.Stat(filename.Value)
			return object.NativeToBool(err == nil)
		},
	}

	FileClass.ClassMethods["exists?"] = FileClass.ClassMethods["exist?"]

	FileClass.ClassMethods["file?"] = &object.Builtin{
		Name: "file?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			info, err := os.Stat(filename.Value)
			if err != nil {
				return object.FALSE
			}
			return object.NativeToBool(!info.IsDir())
		},
	}

	FileClass.ClassMethods["directory?"] = &object.Builtin{
		Name: "directory?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			info, err := os.Stat(filename.Value)
			if err != nil {
				return object.FALSE
			}
			return object.NativeToBool(info.IsDir())
		},
	}

	FileClass.ClassMethods["join"] = &object.Builtin{
		Name: "join",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			parts := make([]string, len(args))
			for i, arg := range args {
				switch a := arg.(type) {
				case *object.String:
					parts[i] = a.Value
				default:
					return newError("no implicit conversion of %s into String", arg.Type())
				}
			}
			return &object.String{Value: filepath.Join(parts...)}
		},
	}

	FileClass.ClassMethods["dirname"] = &object.Builtin{
		Name: "dirname",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			return &object.String{Value: filepath.Dir(path.Value)}
		},
	}

	FileClass.ClassMethods["basename"] = &object.Builtin{
		Name: "basename",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			base := filepath.Base(path.Value)
			// Handle optional suffix argument
			if len(args) > 1 {
				if suffix, ok := args[1].(*object.String); ok {
					if suffix.Value == ".*" {
						// Remove any extension
						ext := filepath.Ext(base)
						base = strings.TrimSuffix(base, ext)
					} else {
						base = strings.TrimSuffix(base, suffix.Value)
					}
				}
			}
			return &object.String{Value: base}
		},
	}

	FileClass.ClassMethods["extname"] = &object.Builtin{
		Name: "extname",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			return &object.String{Value: filepath.Ext(path.Value)}
		},
	}

	FileClass.ClassMethods["expand_path"] = &object.Builtin{
		Name: "expand_path",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1+)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}

			expandedPath := path.Value

			// Handle ~ expansion
			if strings.HasPrefix(expandedPath, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					expandedPath = strings.Replace(expandedPath, "~", home, 1)
				}
			}

			// Handle relative path with base directory
			if len(args) > 1 {
				if base, ok := args[1].(*object.String); ok {
					if !filepath.IsAbs(expandedPath) {
						expandedPath = filepath.Join(base.Value, expandedPath)
					}
				}
			}

			absPath, err := filepath.Abs(expandedPath)
			if err != nil {
				return &object.String{Value: expandedPath}
			}
			return &object.String{Value: absPath}
		},
	}

	FileClass.ClassMethods["delete"] = &object.Builtin{
		Name: "delete",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			count := 0
			for _, arg := range args {
				filename, ok := arg.(*object.String)
				if !ok {
					return newError("no implicit conversion of %s into String", arg.Type())
				}
				err := os.Remove(filename.Value)
				if err != nil {
					return newError("No such file or directory @ unlink_internal - %s", filename.Value)
				}
				count++
			}
			return &object.Integer{Value: int64(count)}
		},
	}

	FileClass.ClassMethods["size"] = &object.Builtin{
		Name: "size",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			info, err := os.Stat(filename.Value)
			if err != nil {
				return newError("No such file or directory @ rb_file_s_size - %s", filename.Value)
			}
			return &object.Integer{Value: info.Size()}
		},
	}
}

func initDirClassMethods() {
	DirClass.ClassMethods["pwd"] = &object.Builtin{
		Name: "pwd",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			pwd, err := os.Getwd()
			if err != nil {
				return newError("couldn't get current directory")
			}
			return &object.String{Value: pwd}
		},
	}

	DirClass.ClassMethods["getwd"] = DirClass.ClassMethods["pwd"]

	DirClass.ClassMethods["chdir"] = &object.Builtin{
		Name: "chdir",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				// chdir to home directory
				home, err := os.UserHomeDir()
				if err != nil {
					return newError("couldn't find HOME directory")
				}
				os.Chdir(home)
				return &object.Integer{Value: 0}
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			err := os.Chdir(path.Value)
			if err != nil {
				return newError("No such file or directory @ dir_chdir - %s", path.Value)
			}
			return &object.Integer{Value: 0}
		},
	}

	DirClass.ClassMethods["entries"] = &object.Builtin{
		Name: "entries",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			files, err := ioutil.ReadDir(path.Value)
			if err != nil {
				return newError("No such file or directory @ dir_initialize - %s", path.Value)
			}
			entries := make([]object.Object, 0, len(files)+2)
			entries = append(entries, &object.String{Value: "."})
			entries = append(entries, &object.String{Value: ".."})
			for _, f := range files {
				entries = append(entries, &object.String{Value: f.Name()})
			}
			return &object.Array{Elements: entries}
		},
	}

	DirClass.ClassMethods["glob"] = &object.Builtin{
		Name: "glob",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1+)")
			}
			pattern, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			matches, err := filepath.Glob(pattern.Value)
			if err != nil {
				return &object.Array{Elements: []object.Object{}}
			}
			elements := make([]object.Object, len(matches))
			for i, m := range matches {
				elements[i] = &object.String{Value: m}
			}
			return &object.Array{Elements: elements}
		},
	}

	DirClass.ClassMethods["mkdir"] = &object.Builtin{
		Name: "mkdir",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			perm := os.FileMode(0755)
			if len(args) > 1 {
				if mode, ok := args[1].(*object.Integer); ok {
					perm = os.FileMode(mode.Value)
				}
			}
			err := os.Mkdir(path.Value, perm)
			if err != nil {
				return newError("File exists @ dir_s_mkdir - %s", path.Value)
			}
			return &object.Integer{Value: 0}
		},
	}

	DirClass.ClassMethods["exist?"] = &object.Builtin{
		Name: "exist?",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			info, err := os.Stat(path.Value)
			if err != nil {
				return object.FALSE
			}
			return object.NativeToBool(info.IsDir())
		},
	}

	DirClass.ClassMethods["exists?"] = DirClass.ClassMethods["exist?"]

	DirClass.ClassMethods["rmdir"] = &object.Builtin{
		Name: "rmdir",
		Fn: func(receiver object.Object, env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments (given 0, expected 1)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("no implicit conversion of %s into String", args[0].Type())
			}
			err := os.Remove(path.Value)
			if err != nil {
				return newError("Directory not empty @ dir_s_rmdir - %s", path.Value)
			}
			return &object.Integer{Value: 0}
		},
	}

	DirClass.ClassMethods["delete"] = DirClass.ClassMethods["rmdir"]
	DirClass.ClassMethods["unlink"] = DirClass.ClassMethods["rmdir"]
}

// getFileBuiltins returns class methods for File
func getFileBuiltins() map[string]*object.Builtin {
	fileBuiltinsOnce.Do(func() {
		fileBuiltinsMap = make(map[string]*object.Builtin)
		for name, method := range FileClass.ClassMethods {
			if builtin, ok := method.(*object.Builtin); ok {
				fileBuiltinsMap[name] = builtin
			}
		}
	})
	return fileBuiltinsMap
}

// getDirBuiltins returns class methods for Dir
func getDirBuiltins() map[string]*object.Builtin {
	dirBuiltinsOnce.Do(func() {
		dirBuiltinsMap = make(map[string]*object.Builtin)
		for name, method := range DirClass.ClassMethods {
			if builtin, ok := method.(*object.Builtin); ok {
				dirBuiltinsMap[name] = builtin
			}
		}
	})
	return dirBuiltinsMap
}
