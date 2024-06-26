package glisp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Many = -1

var WrongType error = errors.New("operands have invalid type")

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
	return SexpNull, fmt.Errorf(`%s expect %s argument(s) but got %v`, funcname, strings.Join(exp, ","), current)
}

func WrongGeneratorNumberArguments(funcname string, current int, expect ...int) error {
	_, err := WrongNumberArguments(funcname, current, expect...)
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyFuncMap(fm map[int]*SexpFunction) map[int]*SexpFunction {
	out := make(map[int]*SexpFunction)
	for k, v := range fm {
		out[k] = v
	}
	return out
}
