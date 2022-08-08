package extensions

import (
	"encoding/base64"
	"fmt"

	"github.com/qjpcpu/glisp"
)

func ImportBase64(env *glisp.Environment) {
	env.AddFunction("base64/decode", Base64StringToBytes)
	env.AddFunction("base64/encode", BytesToBase64String)
}

func Base64StringToBytes(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.Function
	if len(args) != 1 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpStr)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
	}
	bytes, err := base64.StdEncoding.DecodeString(string(str))
	if err != nil {
		return glisp.SexpNull, err
	}
	return glisp.NewSexpBytes(bytes), nil
}

func BytesToBase64String(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.Function
	if len(args) != 1 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpBytes)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be bytes`, name)
	}
	bs := base64.StdEncoding.EncodeToString(str.Bytes())
	return glisp.SexpStr(bs), nil
}
