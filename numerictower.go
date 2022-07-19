package glisp

import (
	"errors"
	"strconv"
)

type IntegerOp int

const (
	ShiftLeftArith IntegerOp = iota
	ShiftRightArith
	ShiftLeftLogical
	ShiftRightLogical
	Modulo
	BitAnd
	BitOr
	BitXor
)

var WrongType error = errors.New("operands have invalid type")

func IntegerDo(op IntegerOp, a, b Sexp) (Sexp, error) {
	var ia SexpInt
	var ib SexpInt

	switch i := a.(type) {
	case SexpInt:
		ia = i
	case SexpChar:
		ia = NewSexpInt(int(i))
	default:
		return SexpNull, WrongType
	}

	switch i := b.(type) {
	case SexpInt:
		ib = i
	case SexpChar:
		ib = NewSexpInt(int(i))
	default:
		return SexpNull, WrongType
	}

	switch op {
	case ShiftLeftArith:
		return ia.ShiftLeft(ib), nil
	case ShiftRightArith:
		return ia.ShiftRight(ib), nil
	case Modulo:
		return ia.Mod(ib), nil
	case BitAnd:
		return ia.And(ib), nil
	case BitOr:
		return ia.Or(ib), nil
	case BitXor:
		return ia.Xor(ib), nil
	}
	return SexpNull, errors.New("unrecognized shift operation")
}

type NumericOp int

const (
	Add NumericOp = iota
	Sub
	Mult
	Div
)

func NumericFloatDo(op NumericOp, a, b SexpFloat) Sexp {
	switch op {
	case Add:
		return a + b
	case Sub:
		return a - b
	case Mult:
		return a * b
	case Div:
		return a / b
	}
	return SexpNull
}

func NumericIntDo(op NumericOp, a, b SexpInt) Sexp {
	switch op {
	case Add:
		return a.Add(b)
	case Sub:
		return a.Sub(b)
	case Mult:
		return a.Mul(b)
	case Div:
		if a.Mod(b).ToInt() == 0 {
			return a.Div(b)
		} else {
			return SexpFloat(float64(a.ToInt64())) / SexpFloat(float64(b.ToInt64()))
		}
	}
	return SexpNull
}

func NumericMatchFloat(op NumericOp, a SexpFloat, b Sexp) (Sexp, error) {
	var fb SexpFloat
	switch tb := b.(type) {
	case SexpFloat:
		fb = tb
	case SexpInt:
		i, err := strconv.ParseFloat(tb.v.String(), 64)
		if err != nil {
			return SexpNull, err
		}
		fb = SexpFloat(i)
	case SexpChar:
		fb = SexpFloat(tb)
	default:
		return SexpNull, WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a SexpInt, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case SexpFloat:
		f, err := a.ToFloat64()
		if err != nil {
			return SexpNull, err
		}
		return NumericFloatDo(op, SexpFloat(f), tb), nil
	case SexpInt:
		return NumericIntDo(op, a, tb), nil
	case SexpChar:
		return NumericIntDo(op, a, NewSexpInt(int(tb))), nil
	}
	return SexpNull, WrongType
}

func NumericMatchChar(op NumericOp, a SexpChar, b Sexp) (Sexp, error) {
	var res Sexp
	switch tb := b.(type) {
	case SexpFloat:
		res = NumericFloatDo(op, SexpFloat(a), tb)
	case SexpInt:
		res = NumericIntDo(op, NewSexpInt(int(a)), tb)
	case SexpChar:
		res = NumericIntDo(op, NewSexpInt(int(a)), NewSexpInt(int(tb)))
	default:
		return SexpNull, WrongType
	}
	switch tres := res.(type) {
	case SexpFloat:
		return tres, nil
	case SexpInt:
		return SexpChar(tres.ToInt()), nil
	}
	return SexpNull, errors.New("unexpected result")
}

func NumericDo(op NumericOp, a, b Sexp) (Sexp, error) {
	switch ta := a.(type) {
	case SexpFloat:
		return NumericMatchFloat(op, ta, b)
	case SexpInt:
		return NumericMatchInt(op, ta, b)
	case SexpChar:
		return NumericMatchChar(op, ta, b)
	}
	return SexpNull, WrongType
}
