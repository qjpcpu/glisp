package glisp

import "strconv"

type SexpStr string

func (s SexpStr) SexpString() string {
	return strconv.Quote(string(s))
}
