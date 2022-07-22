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

func AppendStr(str SexpStr, exprs ...Sexp) (SexpStr, error) {
	var sb strings.Builder
	sb.WriteString(string(str))
	for _, expr := range exprs {
		switch t := expr.(type) {
		case SexpChar:
			sb.WriteRune(rune(t))
		default:
			return SexpStr(""), errors.New("second argument is not a char")
		}
	}

	return SexpStr(sb.String()), nil
}
