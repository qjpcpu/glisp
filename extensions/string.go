package extensions

import (
	"crypto/md5"
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
	env.AddFunctionByConstructor("str/md5", strMD5())
	env.AddFunctionByConstructor("str/mask", strMask)
}

func StringPredict(fn func(string, string) bool) glisp.UserFunctionConstructor {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 2 {
				return glisp.WrongNumberArguments(name, len(args), 2)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !glisp.IsString(args[1]) {
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
				return glisp.WrongNumberArguments(name, len(args), 2)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !glisp.IsString(args[1]) {
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
				return glisp.WrongNumberArguments(name, len(args), 2)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !glisp.IsString(args[1]) {
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
				return glisp.WrongNumberArguments(name, len(args), 1)
			}
			if !glisp.IsString(args[0]) {
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
				return glisp.WrongNumberArguments(name, len(args), 1)
			}
			if !glisp.IsString(args[0]) {
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
				return glisp.WrongNumberArguments(name, len(args), 2)
			}
			list, ok := args[0].(glisp.SexpArray)
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be array`, name)
			}
			if !glisp.IsString(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			var items []string
			for _, v := range list {
				if !glisp.IsString(v) {
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
				return glisp.WrongNumberArguments(name, len(args), 2)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !glisp.IsString(args[1]) {
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
				return glisp.WrongNumberArguments(name, len(args), 3)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			if !glisp.IsString(args[1]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			if !glisp.IsString(args[2]) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
			}
			return glisp.SexpStr(fn(toStr(args[0]), toStr(args[1]), toStr(args[2]))), nil
		}
	}
}

func strMask(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 4 {
			return glisp.WrongNumberArguments(name, len(args), 4)
		}
		if !glisp.IsString(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
		}
		if !glisp.IsInt(args[1]) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be integer`, name)
		}
		if !glisp.IsInt(args[2]) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be integer`, name)
		}
		if !glisp.IsString(args[3]) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be string`, name)
		}
		str := []rune(args[0].(glisp.SexpStr))
		index := args[1].(glisp.SexpInt).ToInt()
		length := args[2].(glisp.SexpInt).ToInt()
		mask := string(args[3].(glisp.SexpStr))
		if index < 0 || index >= len(str) {
			/* mask nothing */
			return args[0], nil
		}
		if length <= 0 && length != -1 {
			return glisp.SexpNull, fmt.Errorf(`%s length must greater than 0 or equal -1`, name)
		}
		if mask == "" {
			return glisp.SexpNull, fmt.Errorf(`%s blank mask`, name)
		}
		end := index + length
		if end > len(str) {
			end = len(str)
		}
		if length == -1 {
			end = len(str)
		}
		var ret strings.Builder
		ret.WriteString(string(str[0:index]))
		ret.WriteString(strings.Repeat(mask, len(str[index:end])))
		ret.WriteString(string(str[end:]))
		return glisp.SexpStr(ret.String()), nil
	}
}
func isDigit() glisp.UserFunctionConstructor {
	return StringBool(func(s string) bool {
		for _, b := range s {
			if b < '0' || b > '9' {
				return false
			}
		}
		return len(s) > 0
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
		return len(s) > 0
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

func strMD5() glisp.UserFunctionConstructor {
	return StringMap(func(s string) string {
		return fmt.Sprintf("%x", md5.Sum([]byte(s)))
	})
}

func toStr(e glisp.Sexp) string {
	return string(e.(glisp.SexpStr))
}
