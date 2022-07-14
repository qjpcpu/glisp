package extensions

import (
	"github.com/qjpcpu/glisp"
	"errors"
	"fmt"
)

type SexpChannel chan glisp.Sexp

func (ch SexpChannel) SexpString() string {
	return "[chan]"
}

func (ch SexpChannel) TypeName() string {
	return "channel"
}

func MakeChanFunction(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) > 1 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	size := 0
	if len(args) == 1 {
		switch t := args[0].(type) {
		case glisp.SexpInt:
			size = t.ToInt()
		default:
			return glisp.SexpNull, errors.New(
				fmt.Sprintf("argument to %s must be int", `make-chan`))
		}
	}

	return SexpChannel(make(chan glisp.Sexp, size)), nil
}

func GetChanTxFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 1 {
			return glisp.SexpNull, glisp.WrongNargs
		}
		var channel chan glisp.Sexp
		switch t := args[0].(type) {
		case SexpChannel:
			channel = chan glisp.Sexp(t)
		default:
			return glisp.SexpNull, errors.New(
				fmt.Sprintf("argument 0 of %s must be channel", name))
		}

		if name == "send!" {
			if len(args) != 2 {
				return glisp.SexpNull, glisp.WrongNargs
			}
			channel <- args[1]
			return glisp.SexpNull, nil
		}

		return <-channel, nil
	}
}

func ImportChannels(env *glisp.Environment) {
	env.AddFunction("make-chan", MakeChanFunction)
	env.AddFunctionByConstructor("send!", GetChanTxFunction)
	env.AddFunctionByConstructor("<!", GetChanTxFunction)
}
