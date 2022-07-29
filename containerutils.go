package glisp

import (
	"bytes"
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

func ConcatBytes(str SexpBytes, exprs ...Sexp) (SexpBytes, error) {
	var sb bytes.Buffer
	sb.Write(str.Bytes())
	for _, expr := range exprs {
		switch t := expr.(type) {
		case SexpBytes:
			sb.Write(t.Bytes())
		default:
			return NewSexpBytes(nil), errors.New("second argument is not bytes")
		}
	}
	return NewSexpBytes(sb.Bytes()), nil
}
