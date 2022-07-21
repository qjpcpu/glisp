package glisp

import (
	"fmt"
	"strconv"
	"strings"
)

const Many = -1

func WrongNumberArguments(funcname string, current int, expect ...int) (Sexp, error) {
	exp := make([]string, len(expect))
	for i, n := range expect {
		if n == Many {
			exp[i] = "..."
			break
		} else {
			exp[i] = strconv.FormatInt(int64(n), 10)
		}
	}
	return SexpNull, fmt.Errorf(`%s expect %s argument but got %v`, funcname, strings.Join(exp, ","), current)
}

func WrongGeneratorNumberArguments(funcname string, current int, expect ...int) error {
	_, err := WrongNumberArguments(funcname, current, expect...)
	return err
}
