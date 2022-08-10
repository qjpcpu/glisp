package extensions

import (
	"github.com/qjpcpu/glisp"
)

func GetCompareFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}

		res, err := glisp.Compare(args[0], args[1])
		if err != nil {
			return glisp.SexpNull, err
		}

		cond := false
		switch name {
		case "<":
			cond = res < 0
		case ">":
			cond = res > 0
		case "<=":
			cond = res <= 0
		case ">=":
			cond = res >= 0
		case "=":
			cond = res == 0
		case "not=", "!=":
			cond = res != 0
		}

		return glisp.SexpBool(cond), nil
	}
}

func GetNumericFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}

		var err error
		accum := args[0]
		var op NumericOp
		switch name {
		case "+":
			op = Add
		case "-":
			op = Sub
		case "*":
			op = Mult
		case "/":
			op = Div
		}

		for _, expr := range args[1:] {
			accum, err = NumericDo(op, accum, expr)
			if err != nil {
				return glisp.SexpNull, err
			}
		}
		return accum, nil
	}
}
