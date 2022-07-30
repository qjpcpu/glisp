package glisp

import (
	"fmt"
	"reflect"
)

type ITypeName interface {
	TypeName() string
}

func GenSymFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 0 {
			return SexpNull, fmt.Errorf(`%s expect 0 argument but got %v`, name, len(args))
		}
		return env.GenSymbol("__anon"), nil
	}
}

func Str2SymFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		if !IsString(args[0]) {
			return SexpNull, fmt.Errorf(`%s first argument bad type %v`, name, args[0])
		}
		return env.MakeSymbol(string(args[0].(SexpStr))), nil
	}
}

func Sym2StrFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		if !IsSymbol(args[0]) {
			return SexpNull, fmt.Errorf(`%s first argument bad type %v`, name, args[0])
		}
		return SexpStr(args[0].(SexpSymbol).Name()), nil
	}
}

func GetTypeFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		var present string
		switch expr := args[0].(type) {
		case SexpSymbol:
			present = `symbol`
		case SexpStr:
			present = `string`
		case SexpArray:
			present = `array`
		case SexpBool:
			present = `bool`
		case SexpChar:
			present = `char`
		case SexpFloat:
			present = `float`
		case *SexpFunction:
			present = `function`
		case *SexpHash:
			present = `hash`
		case SexpInt:
			present = `int`
		case *SexpPair:
			present = `list`
		case SexpSentinel:
			switch expr {
			case SexpNull:
				present = "list"
			case SexpEnd:
				present = "<end>"
			case SexpMarker:
				present = "<marker>"
			}
		case SexpBytes:
			present = `bytes`
		case ITypeName:
			present = expr.TypeName()
		default:
			present = reflect.TypeOf(args[0]).String()
		}
		return SexpStr(present), nil
	}
}
