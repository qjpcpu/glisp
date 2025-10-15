package extensions

import (
	"fmt"
	"sort"

	"github.com/qjpcpu/glisp"
)

func GetCompareFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() < 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2, glisp.Many)
		}

		for i := 1; i < args.Len(); i++ {
			res, err := glisp.Compare(args.Get(i-1), args.Get(i))
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
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() < 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 1)
		}

		var err error
		accum := args.Get(0)
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

		args.SliceStart(1).Foreach(func(expr glisp.Sexp) bool {
			accum, err = NumericDo(op, accum, expr)
			return err == nil
		})
		if err != nil {
			return glisp.SexpNull, err
		}
		return accum, nil
	}
}

func GetSortFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 && args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 1, 2)
		}
		if !glisp.IsFunction(args.Get(0)) && args.Len() == 2 {
			return glisp.SexpNull, fmt.Errorf("first argument must be function but got %v", glisp.InspectType(args.Get(0)))
		}
		var f *glisp.SexpFunction
		var coll glisp.Sexp
		if args.Len() == 2 {
			f = args.Get(0).(*glisp.SexpFunction)
			coll = args.Get(1)
		} else {
			v, _ := env.FindObject("<=")
			f = v.(*glisp.SexpFunction)
			coll = args.Get(0)
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
			res, _ := env.Apply(f, glisp.MakeArgs(arr[i], arr[j]))
			return glisp.IsBool(res) && bool(res.(glisp.SexpBool))
		})
		if isList {
			return glisp.MakeList(arr), nil
		}
		return glisp.SexpArray(arr), nil
	}
}
