package glisp

import (
	"errors"
)

var WrongType error = errors.New("operands have invalid type")

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
		return a.Add(b)
	case Sub:
		return a.Sub(b)
	case Mult:
		return a.Mul(b)
	case Div:
		return a.Div(b)
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
			return NewSexpFloatInt(a).Div(NewSexpFloatInt(b))
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
		fb = NewSexpFloatInt(tb)
	case SexpChar:
		fb = NewSexpFloat(float64(tb))
	default:
		return SexpNull, WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a SexpInt, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case SexpFloat:
		return NumericFloatDo(op, NewSexpFloatInt(a), tb), nil
	case SexpInt:
		if tb.IsZero() && op == Div {
			return SexpNull, errors.New(`division by zero`)
		}
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
		res = NumericFloatDo(op, NewSexpFloat(float64(a)), tb)
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
