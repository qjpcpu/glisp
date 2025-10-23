package glisp

import (
	"fmt"
)

func GetCompareFunction(name string) UserFunction {
	return func(env *Environment, args Args) (Sexp, error) {
		if args.Len() < 2 {
			return WrongNumberArguments(name, args.Len(), 2, Many)
		}
		cond, err := compareArgs(name, args)
		return SexpBool(cond), err
	}
}

func GetNumericFunction(name string) UserFunction {
	return func(env *Environment, args Args) (Sexp, error) {
		if args.Len() < 1 {
			return WrongNumberArguments(name, args.Len(), 1)
		}
		return simpleArithmetic(name, args)
	}
}

func simpleArithmetic(name string, args Args) (Sexp, error) {
	if args.Len() <= 1 {
		return WrongNumberArguments(name, args.Len(), 1, 2)
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
		if IsList(accum, true) {
			handle = list_NumericDoSub
		} else if IsArray(accum) {
			handle = array_NumericDoSub
		}
	}

	args.SliceStart(1).Foreach(func(expr Sexp) bool {
		accum, err = handle(op, accum, expr)
		return err == nil
	})
	if err != nil {
		return SexpNull, err
	}
	return accum, nil
}

func list_NumericDoSub(op NumericOp, a, b Sexp) (Sexp, error) {
	if !IsList(a, true) {
		return SexpNull, fmt.Errorf("operands is not list %s", GetSexpType(a))
	}
	if !IsList(b, true) {
		return SexpNull, fmt.Errorf("operands is not list %s", GetSexpType(b))
	}
	if a == SexpNull || b == SexpNull {
		return a, nil
	}
	presenceMap := make(map[string]struct{})
	b.(*SexpPair).Foreach(func(elem Sexp) bool {
		if key, err := HashExpr(elem); err == nil {
			presenceMap[key] = struct{}{}
		}
		return true
	})
	lb := NewListBuilder()
	a.(*SexpPair).Foreach(func(elem Sexp) bool {
		key, _ := HashExpr(elem)
		if _, ok := presenceMap[key]; !ok {
			lb.Add(elem)
		}
		return true
	})
	return lb.Get(), nil
}

func array_NumericDoSub(op NumericOp, a, b Sexp) (Sexp, error) {
	arr0, ok := a.(SexpArray)
	if !ok {
		return SexpNull, fmt.Errorf("operands is not array %s", GetSexpType(a))
	}
	arr1, ok := b.(SexpArray)
	if !ok {
		return SexpNull, fmt.Errorf("operands is not array %s", GetSexpType(b))
	}
	if len(arr0) == 0 || len(arr1) == 0 {
		return a, nil
	}
	presenceMap := make(map[string]struct{})
	for _, elem := range arr1 {
		if key, err := HashExpr(elem); err == nil {
			presenceMap[key] = struct{}{}
		}
	}
	ret := make([]Sexp, 0, len(arr0))
	for _, elem := range arr0 {
		key, _ := HashExpr(elem)
		if _, ok := presenceMap[key]; !ok {
			ret = append(ret, elem)
		}
	}
	return SexpArray(ret), nil
}
