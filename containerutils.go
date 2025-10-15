package glisp

import (
	"bytes"
	"errors"
	"strings"
)

func ConcatStr(str SexpStr, exprs Args) (SexpStr, error) {
	var sb strings.Builder
	sb.WriteString(string(str))
	var err error
	exprs.Foreach(func(expr Sexp) bool {
		switch t := expr.(type) {
		case SexpStr:
			sb.WriteString(string(t))
		default:
			err = errors.New("second argument is not a string")
		}
		return true
	})
	if err != nil {
		return SexpStr(""), err
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

func ConcatBytes(str SexpBytes, exprs Args) (SexpBytes, error) {
	var sb bytes.Buffer
	sb.Write(str.Bytes())
	var err error
	exprs.Foreach(func(expr Sexp) bool {
		switch t := expr.(type) {
		case SexpBytes:
			sb.Write(t.Bytes())
		default:
			err = errors.New("second argument is not bytes")
		}
		return err == nil
	})
	if err != nil {
		return NewSexpBytes(nil), err
	}
	return NewSexpBytes(sb.Bytes()), nil
}
