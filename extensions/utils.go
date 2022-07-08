package extensions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qjpcpu/glisp"
)

func wrongNumberArguments(funcname string, current int, expect ...int) (glisp.Sexp, error) {
	exp := make([]string, len(expect))
	for i, n := range expect {
		exp[i] = strconv.FormatInt(int64(n), 10)
	}
	return glisp.SexpNull, fmt.Errorf(`%s expect %s argument but got %v`, funcname, strings.Join(exp, "/"), current)
}
