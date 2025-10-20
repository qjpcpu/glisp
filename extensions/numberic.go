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
		handle := NumericDo
		if op == Sub {
			if glisp.IsList(accum, true) {
				handle = list_NumericDoSub
			} else if glisp.IsArray(accum) {
				handle = array_NumericDoSub
			}
		}

		args.SliceStart(1).Foreach(func(expr glisp.Sexp) bool {
			accum, err = handle(op, accum, expr)
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

func list_NumericDoSub(op NumericOp, a, b glisp.Sexp) (glisp.Sexp, error) {
	if !glisp.IsList(a, true) {
		return glisp.SexpNull, fmt.Errorf("operands is not list %s", glisp.GetSexpType(a))
	}
	if !glisp.IsList(b, true) {
		return glisp.SexpNull, fmt.Errorf("operands is not list %s", glisp.GetSexpType(b))
	}
	if a == glisp.SexpNull || b == glisp.SexpNull {
		return a, nil
	}
	hash, _ := glisp.MakeHash(nil)
	b.(*glisp.SexpPair).Foreach(func(elem glisp.Sexp) bool {
		hash.HashSet(elem, glisp.SexpNull)
		return true
	})
	lb := glisp.NewListBuilder()
	a.(*glisp.SexpPair).Foreach(func(elem glisp.Sexp) bool {
		if !hash.HashExist(elem) {
			lb.Add(elem)
		}
		return true
	})
	return lb.Get(), nil
}

func array_NumericDoSub(op NumericOp, a, b glisp.Sexp) (glisp.Sexp, error) {
	arr0, ok := a.(glisp.SexpArray)
	if !ok {
		return glisp.SexpNull, fmt.Errorf("operands is not array %s", glisp.GetSexpType(a))
	}
	arr1, ok := b.(glisp.SexpArray)
	if !ok {
		return glisp.SexpNull, fmt.Errorf("operands is not array %s", glisp.GetSexpType(b))
	}
	if len(arr0) == 0 || len(arr1) == 0 {
		return a, nil
	}
	hash, _ := glisp.MakeHash(nil)
	for _, elem := range arr1 {
		hash.HashSet(elem, glisp.SexpNull)
	}
	ret := make([]glisp.Sexp, 0, len(arr0))
	for _, elem := range arr0 {
		if !hash.HashExist(elem) {
			ret = append(ret, elem)
		}
	}
	return glisp.SexpArray(ret), nil
}
