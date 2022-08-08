package extensions

import (
	"errors"
	"math/rand"
	"time"

	"github.com/qjpcpu/glisp"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomIntegerFunction(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 0 && len(args) != 1 {
		return glisp.WrongNumberArguments(env.Function, len(args), 0, 1)
	}
	if len(args) == 1 {
		if !glisp.IsInt(args[0]) {
			return glisp.SexpNull, errors.New("first argument should be integer")
		}
		num := args[0].(glisp.SexpInt)
		if num.Sign() <= 0 {
			return glisp.SexpNull, errors.New("first argument should greater than 0")
		}
		return num.Random(defaultRand), nil
	}
	return glisp.NewSexpInt(100).Random(defaultRand), nil
}

func RandomFloatFunction(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 0 {
		return glisp.WrongNumberArguments(env.Function, len(args), 0)
	}
	return glisp.NewSexpFloat(defaultRand.Float64()), nil
}

func ImportRandom(env *glisp.Environment) {
	env.AddFunction("rand", RandomIntegerFunction)
	env.AddFunction("randf", RandomFloatFunction)
}
