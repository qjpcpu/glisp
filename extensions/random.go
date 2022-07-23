package extensions

import (
	"github.com/qjpcpu/glisp"
	"math/rand"
	"time"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomFunction(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	return glisp.NewSexpFloat(defaultRand.Float64()), nil
}

func ImportRandom(env *glisp.Environment) {
	env.AddFunction("random", RandomFunction)
}
