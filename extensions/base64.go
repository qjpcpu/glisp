package extensions

import (
	"encoding/base64"
	"fmt"

	"github.com/qjpcpu/glisp"
)

func ImportBase64(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("base64/decode", Base64StringToBytes)
	env.AddNamedFunction("base64/encode", BytesToBase64String)
	return nil
}

func Base64StringToBytes(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args[0]))
		}
		bytes, err := base64.StdEncoding.DecodeString(string(str))
		if err != nil {
			return glisp.SexpNull, err
		}
		return glisp.NewSexpBytes(bytes), nil
	}
}

func BytesToBase64String(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpBytes)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be bytes but got %v`, name, glisp.InspectType(args[0]))
		}
		bs := base64.StdEncoding.EncodeToString(str.Bytes())
		return glisp.SexpStr(bs), nil
	}
}
