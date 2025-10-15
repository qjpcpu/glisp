package extensions

import (
	"errors"
	"fmt"
	"github.com/qjpcpu/glisp"
)

type SexpChannel chan glisp.Sexp

func (ch SexpChannel) SexpString() string {
	return "[chan]"
}

func (ch SexpChannel) TypeName() string {
	return "channel"
}

func MakeChanFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() > 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 0, 1)
		}

		size := 0
		if args.Len() == 1 {
			switch t := args.Get(0).(type) {
			case glisp.SexpInt:
				size = t.ToInt()
			default:
				return glisp.SexpNull, errors.New(
					fmt.Sprintf("argument to %s must be int", `make-chan`))
			}
		}

		return SexpChannel(make(chan glisp.Sexp, size)), nil
	}
}

func ChanTxFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() < 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 1, 2)
		}
		var channel chan glisp.Sexp
		switch t := args.Get(0).(type) {
		case SexpChannel:
			channel = chan glisp.Sexp(t)
		default:
			return glisp.SexpNull, errors.New(
				fmt.Sprintf("argument 0 of %s must be channel", name))
		}

		if name == "send!" {
			if args.Len() != 2 {
				return glisp.WrongNumberArguments(name, args.Len(), 2)
			}
			channel <- args.Get(1)
			return glisp.SexpNull, nil
		}

		return <-channel, nil
	}
}

func ImportChannels(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("make-chan", MakeChanFunction)
	env.AddNamedFunction("send!", ChanTxFunction)
	env.AddNamedFunction("<!", ChanTxFunction)
	return nil
}
