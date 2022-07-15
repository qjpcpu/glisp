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
	env.AddFunctionByConstructor("println", GetPrintFunction(os.Stdout))
	env.AddFunctionByConstructor("printf", GetPrintFunction(os.Stdout))
	env.AddFunctionByConstructor("print", GetPrintFunction(os.Stdout))
	return env.SourceStream(bytes.NewBufferString(core_scripts))
}

func GetPrintFunction(w io.Writer) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) == 0 {
				return glisp.SexpNull, fmt.Errorf("%s need at least one argument", name)
			}
			if name == `printf` {
				if len(args) <= 1 {
					return glisp.SexpNull, fmt.Errorf("%s need at least two argument", name)
				}
				if !glisp.IsString(args[0]) {
					return glisp.SexpNull, fmt.Errorf("first argument of %s must be string", name)
				}
			}

			var items []interface{}

			for _, item := range args {
				switch expr := item.(type) {
				case glisp.SexpStr:
					items = append(items, string(expr))
				default:
					items = append(items, expr.SexpString())
				}
			}

			switch name {
			case "println":
				fmt.Fprintln(w, items...)
			case "print":
				fmt.Fprint(w, items...)
			case "printf":
				fmt.Fprintf(w, items[0].(string), items[1:]...)
			}

			return glisp.SexpNull, nil
		}
	}
}
