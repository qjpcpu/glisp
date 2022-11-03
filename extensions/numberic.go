package extensions

import (
	"fmt"
	"sort"

	"github.com/qjpcpu/glisp"
)

func GetCompareFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 2 {
			return glisp.WrongNumberArguments(name, len(args), 2, glisp.Many)
		}

		for i := 1; i < len(args); i++ {
			res, err := glisp.Compare(args[i-1], args[i])
			if err != nil {
				return glisp.SexpNull, err
			}

			var cond bool
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
			if !cond {
				return glisp.SexpBool(false), nil
			}
		}

		return glisp.SexpBool(true), nil
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

func GetSortFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 1, 2)
		}
		if !glisp.IsFunction(args[0]) && len(args) == 2 {
			return glisp.SexpNull, fmt.Errorf("first argument must be function but got %v", glisp.InspectType(args[0]))
		}
		var f *glisp.SexpFunction
		var coll glisp.Sexp
		if len(args) == 2 {
			f = args[0].(*glisp.SexpFunction)
			coll = args[1]
		} else {
			v, _ := env.FindObject("<=")
			f = v.(*glisp.SexpFunction)
			coll = args[0]
		}

		var arr []glisp.Sexp
		var isList bool
		if coll == glisp.SexpNull {
			return coll, nil
		} else if glisp.IsArray(coll) {
			arr = coll.(glisp.SexpArray)
		} else if glisp.IsList(coll) {
			isList = true
			arr, _ = glisp.ListToArray(coll)
		} else {
			return glisp.SexpNull, fmt.Errorf("second argument must be array/list but got %v", glisp.InspectType(coll))
		}
		sort.SliceStable(arr, func(i, j int) bool {
			res, _ := env.Apply(f, []glisp.Sexp{arr[i], arr[j]})
			return glisp.IsBool(res) && bool(res.(glisp.SexpBool))
		})
		if isList {
			return glisp.MakeList(arr), nil
		}
		return glisp.SexpArray(arr), nil
	}
}
