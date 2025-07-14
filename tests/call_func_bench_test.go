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
		env.PushGlobalScope()
		env.Bind("any", glisp.NewSexpInt(100))
		env.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(i), glisp.NewSexpInt(i % 10)})
		env.PopGlobalScope()
	}
}

func BenchmarkCallUserFunction(b *testing.B) {
	env := newFullEnv()
	env.AddFunction("user_minus", func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		a, b := args[0].(glisp.SexpInt), args[1].(glisp.SexpInt)
		return a.Sub(b), nil
	})
	env.EvalString(`
(defn test-user-function [a b]
   (user_minus a b)
)`)
	obj, _ := env.FindObject("test-user-function")
	fn := obj.(*glisp.SexpFunction)
	for i := 0; i < b.N; i++ {
		env.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(i), glisp.NewSexpInt(i % 10)})
	}
}

func BenchmarkCallUserInstruction(b *testing.B) {
	env := newFullEnv()
	env.AddInstruction("user_minus", func(env *glisp.UserInstrContext) (glisp.Sexp, error) {
		b, a := env.PopExpr().(glisp.SexpInt), env.PopExpr().(glisp.SexpInt)
		return a.Sub(b), nil
	})
	env.EvalString(`
(defn test-user-function [a b]
   (user_minus a b)
)`)
	obj, _ := env.FindObject("test-user-function")
	fn := obj.(*glisp.SexpFunction)
	for i := 0; i < b.N; i++ {
		env.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(i), glisp.NewSexpInt(i % 10)})
	}
}
