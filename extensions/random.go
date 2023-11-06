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
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 0 && len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 0, 1)
		}
		if len(args) == 1 {
			if !glisp.IsInt(args[0]) {
				return glisp.SexpNull, fmt.Errorf("first argument should be integer but got %v", glisp.InspectType(args[0]))
			}
			num := args[0].(glisp.SexpInt)
			if num.Sign() <= 0 {
				return glisp.SexpNull, errors.New("first argument should greater than 0")
			}
			return num.Random(defaultRand), nil
		}
		return glisp.NewSexpInt(100).Random(defaultRand), nil
	}
}

func RandomFloatFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 0 {
			return glisp.WrongNumberArguments(name, len(args), 0)
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
