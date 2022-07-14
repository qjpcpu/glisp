package extensions

import (
	"fmt"
	"strings"

	"github.com/qjpcpu/glisp"
)

func ImportString(env *glisp.Environment) {
	env.AddFunctionByConstructor("str/start-with?", StringPredict(strings.HasPrefix))
	env.AddFunctionByConstructor("str/end-with?", StringPredict(strings.HasSuffix))
	env.AddFunctionByConstructor("str/contains?", StringPredict(strings.Contains))
	env.AddFunctionByConstructor("str/title", StringMap(strings.Title))
	env.AddFunctionByConstructor("str/lower", StringMap(strings.ToLower))
	env.AddFunctionByConstructor("str/upper", StringMap(strings.ToUpper))
	env.AddFunctionByConstructor("str/replace", StringMap3(strings.ReplaceAll))
	env.AddFunctionByConstructor("str/trim-prefix", StringMap2(strings.TrimPrefix))
	env.AddFunctionByConstructor("str/trim-suffix", StringMap2(strings.TrimSuffix))
	env.AddFunctionByConstructor("str/trim-space", StringMap(strings.TrimSpace))
	env.AddFunctionByConstructor("str/count", StringSearch(strings.Count))
	env.AddFunctionByConstructor("str/index", StringSearch(strings.Index))
	env.AddFunctionByConstructor("str/split", StringSplit(strings.Split))
	env.AddFunctionByConstructor("str/join", StringJoin(strings.Join))
	env.AddFunctionByConstructor("str/digit?", isDigit())
	env.AddFunctionByConstructor("str/alpha?", isAlpha())
	env.AddFunctionByConstructor("str/title?", isTitle())
}

func StringPredict(fn func(string, string) bool) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return wrongNumberArguments(name, len(args), 2)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			return glisp.SexpBool(fn(toStr(args[0]), toStr(args[1]))), nil
		}
	}
}

func StringSearch(fn func(string, string) int) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return wrongNumberArguments(name, len(args), 2)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			return glisp.NewSexpInt((fn(toStr(args[0]), toStr(args[1])))), nil
		}
	}
}

func StringSplit(fn func(string, string) []string) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return wrongNumberArguments(name, len(args), 2)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			list := fn(toStr(args[0]), toStr(args[1]))
			var array glisp.SexpArray
			for _, str := range list {
				array = append(array, glisp.SexpStr(str))
			}
			return array, nil
		}
	}
}

func StringMap(fn func(string) string) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 1 {
				return wrongNumberArguments(name, len(args), 1)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			return glisp.SexpStr((fn(toStr(args[0])))), nil
		}
	}
}

func StringBool(fn func(string) bool) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 1 {
				return wrongNumberArguments(name, len(args), 1)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			return glisp.SexpBool((fn(toStr(args[0])))), nil
		}
	}
}

func StringJoin(fn func([]string, string) string) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return wrongNumberArguments(name, len(args), 2)
			}
			list, ok := args[0].(glisp.SexpArray)
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be array`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			var items []string
			for _, v := range list {
				if !isSexpStr(v) {
					return glisp.SexpNull, fmt.Errorf(`%s first argument should be array of string`, name)
				}
				items = append(items, string(v.(glisp.SexpStr)))
			}
			return glisp.SexpStr(fn(items, string(args[1].(glisp.SexpStr)))), nil
		}
	}
}

func StringMap2(fn func(string, string) string) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return wrongNumberArguments(name, len(args), 2)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			return glisp.SexpStr(fn(toStr(args[0]), toStr(args[1]))), nil
		}
	}
}

func StringMap3(fn func(string, string, string) string) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 3 {
				return wrongNumberArguments(name, len(args), 3)
			}
			if !isSexpStr(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !isSexpStr(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			if !isSexpStr(args[2]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			return glisp.SexpStr(fn(toStr(args[0]), toStr(args[1]), toStr(args[2]))), nil
		}
	}
}

func isDigit() glisp.UserFunctionConstructor {
	return StringBool(func(s string) bool {
		for _, b := range s {
			if b < '0' || b > '9' {
				return false
			}
		}
		return true
	})
}

func isAlpha() glisp.UserFunctionConstructor {
	return StringBool(func(s string) bool {
		for _, b := range s {
			if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') {
			} else {
				return false
			}
		}
		return true
	})
}

func isTitle() glisp.UserFunctionConstructor {
	return StringBool(func(s string) bool {
		if len(s) == 0 || !('A' <= s[0] && s[0] <= 'Z') {
			return false
		}
		for i := 1; i < len(s); i++ {
			b := s[i]
			if 'a' <= b && b <= 'z' {
			} else {
				return false
			}
		}
		return true
	})
}

func isSexpStr(e glisp.Sexp) bool {
	_, ok := e.(glisp.SexpStr)
	return ok
}

func toStr(e glisp.Sexp) string {
	return string(e.(glisp.SexpStr))
}
