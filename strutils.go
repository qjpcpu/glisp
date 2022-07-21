package glisp

import (
	"errors"
	"strings"
)

func ConcatStr(str SexpStr, exprs ...Sexp) (SexpStr, error) {
	var sb strings.Builder
	sb.WriteString(string(str))
	for _, expr := range exprs {
		switch t := expr.(type) {
		case SexpStr:
			sb.WriteString(string(t))
		default:
			return SexpStr(""), errors.New("second argument is not a string")
		}
	}

	return SexpStr(sb.String()), nil
}

func AppendStr(str SexpStr, expr Sexp) (SexpStr, error) {
	var chr SexpChar
	switch t := expr.(type) {
	case SexpChar:
		chr = t
	default:
		return SexpStr(""), errors.New("second argument is not a char")
	}

	return str + SexpStr(chr), nil
}
