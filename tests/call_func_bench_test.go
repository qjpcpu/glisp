package tests

import (
	"testing"

	"github.com/qjpcpu/glisp"
)

func BenchmarkCallFunction(b *testing.B) {
	env := newFullEnv()
	env.EvalString(`
(defn plus [a b]
(cond (> b 0) (plus a (- b 1))
      (+ a b))
)`)
	obj, _ := env.FindObject("plus")
	fn := obj.(*glisp.SexpFunction)
	for i := 0; i < b.N; i++ {
		env.Bind("any", glisp.NewSexpInt(100))
		env.Apply(fn, glisp.MakeArgs(glisp.NewSexpInt(i), glisp.NewSexpInt(i%10)))
	}
}

func BenchmarkCallUserFunction(b *testing.B) {
	env := newFullEnv()
	env.AddFunction("user_minus", func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		a, b := args.Get(0).(glisp.SexpInt), args.Get(1).(glisp.SexpInt)
		return a.Sub(b), nil
	})
	env.EvalString(`
(defn test-user-function [a b]
   (user_minus a b)
)`)
	obj, _ := env.FindObject("test-user-function")
	fn := obj.(*glisp.SexpFunction)
	for i := 0; i < b.N; i++ {
		env.Apply(fn, glisp.MakeArgs(glisp.NewSexpInt(i), glisp.NewSexpInt(i%10)))
	}
}
