package extensions

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/qjpcpu/glisp"
)

type SexpRegexp regexp.Regexp

func (re SexpRegexp) SexpString() string {
	r := regexp.Regexp(re)
	return fmt.Sprintf(`(regexp-compile "%v")`, r.String())
}

func regexpFindIndex(
	needle regexp.Regexp, haystack string) (glisp.Sexp, error) {

	loc := needle.FindStringIndex(haystack)

	arr := make([]glisp.Sexp, len(loc))
	for i := range arr {
		arr[i] = glisp.Sexp(glisp.NewSexpInt(loc[i]))
	}

	return glisp.SexpArray(arr), nil
}

func GetRegexpFind(name string) glisp.GlispUserFunction {
	return func(env *glisp.Glisp, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, glisp.WrongNargs
		}
		var haystack string
		switch t := args[1].(type) {
		case glisp.SexpStr:
			haystack = string(t)
		default:
			return glisp.SexpNull,
				fmt.Errorf("2nd argument of %v should be a string", name)
		}

		var needle regexp.Regexp
		switch t := args[0].(type) {
		case SexpRegexp:
			needle = regexp.Regexp(t)
		default:
			return glisp.SexpNull,
				fmt.Errorf("1st argument of %v should be a compiled regular expression", name)
		}

		switch name {
		case "regexp-find":
			str := needle.FindString(haystack)
			return glisp.SexpStr(str), nil
		case "regexp-find-index":
			return regexpFindIndex(needle, haystack)
		case "regexp-match":
			matches := needle.MatchString(haystack)
			return glisp.SexpBool(matches), nil
		}

		return glisp.SexpNull, errors.New("unknown function")
	}
}

func RegexpCompile(env *glisp.Glisp, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) < 1 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var re string
	switch t := args[0].(type) {
	case glisp.SexpStr:
		re = string(t)
	default:
		return glisp.SexpNull,
			errors.New("argument of regexp-compile should be a string")
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return glisp.SexpNull, errors.New(
			fmt.Sprintf("error during regexp-compile: '%v'", err))
	}

	return glisp.Sexp(SexpRegexp(*r)), nil
}

func ImportRegex(env *glisp.Glisp) {
	env.AddFunction("regexp-compile", RegexpCompile)
	env.AddFunctionByConstructor("regexp-find-index", GetRegexpFind)
	env.AddFunctionByConstructor("regexp-find", GetRegexpFind)
	env.AddFunctionByConstructor("regexp-match", GetRegexpFind)
}
