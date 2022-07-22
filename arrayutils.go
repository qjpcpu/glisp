package glisp

import (
	"errors"
)

func MapArray(env *Environment, fun SexpFunction, arr SexpArray) (SexpArray, error) {
	result := make([]Sexp, len(arr))
	var err error

	for i := range arr {
		result[i], err = env.Apply(fun, arr[i:i+1])
		if err != nil {
			return SexpArray(result), err
		}
	}

	return SexpArray(result), nil
}

func FilterArray(env *Environment, fun SexpFunction, arr SexpArray) (SexpArray, error) {
	result := make([]Sexp, 0, len(arr))

	for i := range arr {
		item := arr[i]
		ret, err := env.Apply(fun, []Sexp{item})
		if err != nil {
			return SexpArray(result), err
		}
		pass, ok := ret.(SexpBool)
		if !ok {
			return SexpArray{}, errors.New("filter function must return boolean")
		} else if pass {
			result = append(result, item)
		}
	}

	return SexpArray(result), nil
}

func ConcatArray(arr SexpArray, exprs ...Sexp) (SexpArray, error) {
	ret := make(SexpArray, len(arr))
	copy(ret, arr)
	for _, expr := range exprs {
		switch t := expr.(type) {
		case SexpArray:
			ret = append(ret, t...)
		default:
			return arr, errors.New("second argument is not an array")
		}
	}

	return ret, nil
}

func FoldlArray(env *Environment, fun SexpFunction, lst Sexp, acc Sexp) (Sexp, error) {
	if lst == SexpNull {
		return acc, nil
	}
	var list SexpArray
	switch e := lst.(type) {
	case SexpArray:
		list = e
	default:
		return SexpNull, errors.New("not a array")
	}

	if len(list) == 0 {
		return acc, nil
	}

	var err error
	if acc, err = env.Apply(fun, []Sexp{list[0], acc}); err != nil {
		return SexpNull, err
	}

	return FoldlArray(env, fun, SexpArray(list[1:]), acc)
}
