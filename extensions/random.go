package extensions

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/qjpcpu/glisp"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomIntegerFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 0 && args.Len() != 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 0, 1)
		}
		if args.Len() == 1 {
			if !glisp.IsInt(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf("first argument should be integer but got %v", glisp.InspectType(args.Get(0)))
			}
			num := args.Get(0).(glisp.SexpInt)
			if num.Sign() <= 0 {
				return glisp.SexpNull, errors.New("first argument should greater than 0")
			}
			return num.Random(defaultRand), nil
		}
		return glisp.NewSexpInt(100).Random(defaultRand), nil
	}
}

func RandomFloatFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 0 {
			return glisp.WrongNumberArguments(name, args.Len(), 0)
		}
		return glisp.NewSexpFloat(defaultRand.Float64()), nil
	}
}

func ImportRandom(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("rand", RandomIntegerFunction)
	env.AddNamedFunction("randf", RandomFloatFunction)
	return nil
}
