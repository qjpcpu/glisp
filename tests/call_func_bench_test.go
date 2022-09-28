package tests

import (
	"testing"

	"github.com/qjpcpu/glisp"
)

func BenchmarkCallFunction(b *testing.B) {
	env := newFullEnv()
	env.EvalString(`(defn plus [a b]
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
