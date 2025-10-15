package glisp

import (
	"fmt"
	"reflect"
)

type ITypeName interface {
	TypeName() string
}

type Stringer interface {
	Stringify() string
}

func IsIType(s Sexp) bool {
	_, ok := s.(ITypeName)
	return ok
}

func GetGenSymFunction(name string) UserFunction {
	return func(env *Environment, args Args) (Sexp, error) {
		if args.Len() != 0 {
			return SexpNull, fmt.Errorf(`%s expect 0 argument but got %v`, name, args.Len())
		}
		return env.GenSymbol("__anon"), nil
	}
}

func GetAnyToSymbolFunction(name string) UserFunction {
	return func(env *Environment, args Args) (Sexp, error) {
		if args.Len() != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		switch val := args.Get(0).(type) {
		case SexpStr:
			return env.MakeSymbol(string(args.Get(0).(SexpStr))), nil
		case SexpSymbol:
			return val, nil
		}
		return SexpNull, fmt.Errorf(`%s first argument bad type %v`, name, args.Get(0))
	}
}

func GetTypeFunction(name string) UserFunction {
	return func(env *Environment, args Args) (Sexp, error) {
		if args.Len() != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		var present string
		switch expr := args.Get(0).(type) {
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
		case SexpError:
			present = `error`
		case ITypeName:
			present = expr.TypeName()
		default:
			present = reflect.TypeOf(args.Get(0)).String()
		}
		return SexpStr(present), nil
	}
}

func InspectType(expr Sexp) string {
	if expr == SexpNull {
		return `nil`
	}
	t, _ := GetTypeFunction(`type`)(nil, MakeArgs(expr))
	return string(t.(SexpStr))
}
