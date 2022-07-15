package glisp

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

type Comparable interface {
	Sexp
	// Cmp compares x and y and returns:
	//
	//   -1 if x <  y
	//    0 if x == y
	//   +1 if x >  y
	//
	Cmp(Comparable) (int, error)
}

func Compare(a Sexp, b Sexp) (int, error) {
	switch at := a.(type) {
	case SexpInt:
		return compareInt(at, b)
	case SexpChar:
		return compareChar(at, b)
	case SexpFloat:
		return compareFloat(at, b)
	case SexpBool:
		return compareBool(at, b)
	case SexpStr:
		return compareString(at, b)
	case SexpSymbol:
		return compareSymbol(at, b)
	case SexpPair:
		return comparePair(at, b)
	case SexpArray:
		return compareArray(at, b)
	case SexpSentinel:
		if at == SexpNull && b == SexpNull {
			return 0, nil
		} else {
			return -1, nil
		}
	case SexpBytes:
		return compareBytes(at, b)
	}
	if isComparable(a) && isComparable(b) {
		return a.(Comparable).Cmp(b.(Comparable))
	}
	return 0, fmt.Errorf("cannot compare %T(%s) to %T(%s)", a, a.SexpString(), b, b.SexpString())
}

func signumFloat(f SexpFloat) int {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func signumInt(i SexpInt) int {
	return i.Sign()
}

func compareFloat(f SexpFloat, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return compareFloatAndInt(f, e), nil
	case SexpFloat:
		return signumFloat(f - e), nil
	case SexpChar:
		return signumFloat(f - SexpFloat(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", f, f.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func compareIntAndFloat(e SexpInt, f SexpFloat) int {
	return -compareFloatAndInt(f, e)
}

func compareFloatAndInt(f SexpFloat, e SexpInt) int {
	if e.IsInt64() {
		return signumFloat(f - SexpFloat(e.ToInt64()))
	} else if e.IsUint64() {
		return signumFloat(f - SexpFloat(e.ToUint64()))
	} else {
		ev, _ := big.NewFloat(0).SetString(e.v.String())
		return big.NewFloat(float64(f)).Cmp(ev)
	}
}

func compareBetweenInt(f, e SexpInt) int {
	return f.v.Cmp(e.v)
}

func compareInt(i SexpInt, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return compareBetweenInt(i, e), nil
	case SexpFloat:
		return compareIntAndFloat(i, e), nil
	case SexpChar:
		si, _ := NewSexpIntStr(strconv.FormatInt(int64(byte(e)), 10))
		return compareBetweenInt(i, si), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", i, i.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func compareChar(c SexpChar, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return compareBetweenInt(NewSexpInt(int(c)), e), nil
	case SexpFloat:
		return signumFloat(SexpFloat(c) - e), nil
	case SexpChar:
		ci := NewSexpInt64(int64(byte(c)))
		ei := NewSexpInt64(int64(byte(e)))
		return compareBetweenInt(ci, ei), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", c, c.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func compareString(s SexpStr, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpStr:
		return bytes.Compare([]byte(s), []byte(e)), nil
	case SexpBytes:
		return bytes.Compare([]byte(s), e.bytes), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", s, s.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func compareBytes(s SexpBytes, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpBytes:
		return bytes.Compare(s.bytes, e.bytes), nil
	case SexpStr:
		return bytes.Compare(s.bytes, []byte(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", s, s.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func compareSymbol(sym SexpSymbol, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpSymbol:
		return compareBetweenInt(NewSexpInt(sym.number), NewSexpInt(e.number)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", sym, sym.SexpString(), expr, expr.SexpString())
	return 0, errors.New(errmsg)
}

func comparePair(a SexpPair, b Sexp) (int, error) {
	var bp SexpPair
	switch t := b.(type) {
	case SexpPair:
		bp = t
	default:
		errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", a, a.SexpString(), b, b.SexpString())
		return 0, errors.New(errmsg)
	}
	res, err := Compare(a.head, bp.head)
	if err != nil {
		return 0, err
	}
	if res != 0 {
		return res, nil
	}
	return Compare(a.tail, bp.tail)
}

func compareArray(a SexpArray, b Sexp) (int, error) {
	var ba SexpArray
	switch t := b.(type) {
	case SexpArray:
		ba = t
	default:
		errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", a, a.SexpString(), b, b.SexpString())
		return 0, errors.New(errmsg)
	}
	var length int
	if len(a) < len(ba) {
		length = len(a)
	} else {
		length = len(ba)
	}

	for i := 0; i < length; i++ {
		res, err := Compare(a[i], ba[i])
		if err != nil {
			return 0, err
		}
		if res != 0 {
			return res, nil
		}
	}

	return compareBetweenInt(NewSexpInt(len(a)), NewSexpInt(len(ba))), nil
}

func compareBool(a SexpBool, b Sexp) (int, error) {
	var bb SexpBool
	switch bt := b.(type) {
	case SexpBool:
		bb = bt
	default:
		errmsg := fmt.Sprintf("cannot compare %T(%s) to %T(%s)", a, a.SexpString(), b, b.SexpString())
		return 0, errors.New(errmsg)
	}

	// true > false
	if a && bb {
		return 0, nil
	}
	if a {
		return 1, nil
	}
	if bb {
		return -1, nil
	}
	return 0, nil
}

func existInList(a Sexp, element Sexp) (bool, error) {
	for {
		if a == SexpNull {
			return false, nil
		}
		expr := a.(SexpPair)
		eq, err := Compare(expr.head, element)
		if err != nil {
			return false, err
		}
		if eq == 0 {
			return true, nil
		}
		a = expr.tail
	}
}
