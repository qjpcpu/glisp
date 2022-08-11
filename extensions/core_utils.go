package extensions

import (
	"bytes"
	"fmt"
	"io"
	"os"

	_ "embed"

	"github.com/qjpcpu/glisp"
)

var (
	//go:embed core_utils.lisp
	core_scripts string
)

func ImportCoreUtils(env *glisp.Environment) error {
	env.AddNamedFunction("println", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("printf", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("print", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("sprintf", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("<", GetCompareFunction)
	env.AddNamedFunction(">", GetCompareFunction)
	env.AddNamedFunction("<=", GetCompareFunction)
	env.AddNamedFunction(">=", GetCompareFunction)
	env.AddNamedFunction("=", GetCompareFunction)
	env.AddNamedFunction("not=", GetCompareFunction)
	env.AddNamedFunction("!=", GetCompareFunction)
	env.AddNamedFunction("+", GetNumericFunction)
	env.AddNamedFunction("-", GetNumericFunction)
	env.AddNamedFunction("*", GetNumericFunction)
	env.AddNamedFunction("/", GetNumericFunction)
	env.AddNamedFunction("mod", GetBinaryIntFunction)
	return env.SourceStream(bytes.NewBufferString(core_scripts))
}

func GetPrintFunction(w io.Writer) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) == 0 {
				return glisp.SexpNull, fmt.Errorf("%s need at least one argument", name)
			}
			if name == `printf` || name == `sprintf` {
				if len(args) <= 1 {
					return glisp.SexpNull, fmt.Errorf("%s need at least two argument", name)
				}
				if !glisp.IsString(args[0]) {
					return glisp.SexpNull, fmt.Errorf("first argument of %s must be string", name)
				}
			}

			var items []interface{}
			for _, item := range args {
				items = append(items, mapSexpToGoPrintableInterface(item))
			}

			switch name {
			case "println":
				fmt.Fprintln(w, items...)
			case "print":
				fmt.Fprint(w, items...)
			case "printf":
				fmt.Fprintf(w, refactFmtStr(items[0].(string)), items[1:]...)
			case "sprintf":
				return glisp.SexpStr(fmt.Sprintf(refactFmtStr(items[0].(string)), items[1:]...)), nil
			}

			return glisp.SexpNull, nil
		}
	}
}

func mapSexpToGoPrintableInterface(sexp glisp.Sexp) interface{} {
	if sexp == glisp.SexpNull {
		return nil
	}
	switch expr := sexp.(type) {
	case glisp.SexpStr:
		return string(expr)
	case glisp.SexpBool:
		return bool(expr)
	case glisp.SexpInt:
		if expr.IsInt64() {
			return expr.ToInt64()
		} else if expr.IsUint64() {
			return expr.ToUint64()
		} else {
			return expr.SexpString()
		}
	case glisp.SexpFloat:
		return expr.SexpString()
	case glisp.SexpSymbol:
		return expr.Name()
	case glisp.SexpChar:
		return rune(expr)
	default:
		return expr.SexpString()
	}
}

func refactFmtStr(str string) string {
	data := []rune(str)
	var ret []rune
	var foundSym bool
	for i := 0; i < len(data); {
		b := data[i]
		if !foundSym {
			if b == '%' {
				if i < len(data)-1 && data[i+1] == '%' {
					ret = append(ret, '%', '%')
					i += 2
				} else {
					foundSym = true
					i++
				}
				continue
			} else {
				ret = append(ret, b)
				i++
			}
		} else {
			if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') {
				ret = append(ret, '%', 'v')
				foundSym = false
			}
			i++
		}
	}
	return string(ret)
}
