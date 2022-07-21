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
