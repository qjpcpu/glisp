package extensions

import (
	"errors"

	"github.com/qjpcpu/glisp"
)

type NumericOp int

const (
	Add NumericOp = iota
	Sub
	Mult
	Div
)

func NumericFloatDo(op NumericOp, a, b glisp.SexpFloat) glisp.Sexp {
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
	return glisp.SexpNull
}

func NumericIntDo(op NumericOp, a, b glisp.SexpInt) glisp.Sexp {
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
			return glisp.NewSexpFloatInt(a).Div(glisp.NewSexpFloatInt(b))
		}
	}
	return glisp.SexpNull
}

func NumericMatchFloat(op NumericOp, a glisp.SexpFloat, b glisp.Sexp) (glisp.Sexp, error) {
	var fb glisp.SexpFloat
	switch tb := b.(type) {
	case glisp.SexpFloat:
		fb = tb
	case glisp.SexpInt:
		fb = glisp.NewSexpFloatInt(tb)
	case glisp.SexpChar:
		fb = glisp.NewSexpFloat(float64(tb))
	default:
		return glisp.SexpNull, glisp.WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a glisp.SexpInt, b glisp.Sexp) (glisp.Sexp, error) {
	switch tb := b.(type) {
	case glisp.SexpFloat:
		return NumericFloatDo(op, glisp.NewSexpFloatInt(a), tb), nil
	case glisp.SexpInt:
		if tb.IsZero() && op == Div {
			return glisp.SexpNull, errors.New(`division by zero`)
		}
		return NumericIntDo(op, a, tb), nil
	case glisp.SexpChar:
		return NumericIntDo(op, a, glisp.NewSexpInt(int(tb))), nil
	}
	return glisp.SexpNull, glisp.WrongType
}

func NumericMatchChar(op NumericOp, a glisp.SexpChar, b glisp.Sexp) (glisp.Sexp, error) {
	var res glisp.Sexp
	switch tb := b.(type) {
	case glisp.SexpFloat:
		res = NumericFloatDo(op, glisp.NewSexpFloat(float64(a)), tb)
	case glisp.SexpInt:
		if tb.IsZero() {
			return glisp.SexpNull, errors.New(`division by zero`)
		}
		res = NumericIntDo(op, glisp.NewSexpInt(int(a)), tb)
	case glisp.SexpChar:
		res = NumericIntDo(op, glisp.NewSexpInt(int(a)), glisp.NewSexpInt(int(tb)))
	default:
		return glisp.SexpNull, glisp.WrongType
	}
	switch tres := res.(type) {
	case glisp.SexpFloat:
		return tres, nil
	case glisp.SexpInt:
		return glisp.SexpChar(tres.ToInt()), nil
	}
	return glisp.SexpNull, errors.New("unexpected result")
}

func NumericDo(op NumericOp, a, b glisp.Sexp) (glisp.Sexp, error) {
	switch ta := a.(type) {
	case glisp.SexpFloat:
		return NumericMatchFloat(op, ta, b)
	case glisp.SexpInt:
		return NumericMatchInt(op, ta, b)
	case glisp.SexpChar:
		return NumericMatchChar(op, ta, b)
	}
	return glisp.SexpNull, glisp.WrongType
}
