package extensions

import (
	"crypto/md5"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/qjpcpu/glisp"
)

func ImportString(env *glisp.Environment) {
	env.AddNamedFunction("str/start-with?", StringPredict(strings.HasPrefix))
	env.AddNamedFunction("str/end-with?", StringPredict(strings.HasSuffix))
	env.AddNamedFunction("str/contains?", StringPredict(strings.Contains))
	env.AddNamedFunction("str/title", StringMap(strings.Title))
	env.AddNamedFunction("str/lower", StringMap(strings.ToLower))
	env.AddNamedFunction("str/upper", StringMap(strings.ToUpper))
	env.AddNamedFunction("str/replace", StringMap3(strings.ReplaceAll))
	env.AddNamedFunction("str/trim-prefix", StringMap2(strings.TrimPrefix))
	env.AddNamedFunction("str/trim-suffix", StringMap2(strings.TrimSuffix))
	env.AddNamedFunction("str/trim-space", StringMap(strings.TrimSpace))
	env.AddNamedFunction("str/count", StringSearch(strings.Count))
	env.AddNamedFunction("str/index", StringSearch(strings.Index))
	env.AddNamedFunction("str/split", StringSplit(strings.Split))
	env.AddNamedFunction("str/join", StringJoin(strings.Join))
	env.AddNamedFunction("str/len", StringInt(utf8.RuneCountInString))
	env.AddNamedFunction("str/digit?", isDigit())
	env.AddNamedFunction("str/alpha?", isAlpha())
	env.AddNamedFunction("str/title?", isTitle())
	env.AddNamedFunction("str/integer?", isInteger())
	env.AddNamedFunction("str/float?", isFloat())
	env.AddNamedFunction("str/bool?", isBool())
	env.AddNamedFunction("str/md5", strMD5())
	env.AddNamedFunction("str/mask", strMask)
}

func StringPredict(fn func(string, string) bool) glisp.NamedUserFunction {
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

func StringSearch(fn func(string, string) int) glisp.NamedUserFunction {
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

func StringSplit(fn func(string, string) []string) glisp.NamedUserFunction {
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

func StringMap(fn func(string) string) glisp.NamedUserFunction {
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

func StringBool(fn func(string) bool) glisp.NamedUserFunction {
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

func StringInt(fn func(string) int) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 1 {
				return glisp.WrongNumberArguments(name, len(args), 1)
			}
			if !glisp.IsString(args[0]) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string`, name)
			}
			return glisp.NewSexpInt((fn(toStr(args[0])))), nil
		}
	}
}

func StringJoin(fn func([]string, string) string) glisp.NamedUserFunction {
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

func StringMap2(fn func(string, string) string) glisp.NamedUserFunction {
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

func StringMap3(fn func(string, string, string) string) glisp.NamedUserFunction {
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

func isDigit() glisp.NamedUserFunction {
	return StringBool(func(s string) bool {
		for _, b := range s {
			if b < '0' || b > '9' {
				return false
			}
		}
		return len(s) > 0
	})
}

func isInteger() glisp.NamedUserFunction {
	return StringBool(func(s string) bool {
		_, err := glisp.NewSexpIntStr(s)
		return err == nil
	})
}

func isFloat() glisp.NamedUserFunction {
	return StringBool(func(s string) bool {
		_, err := glisp.NewSexpFloatStr(s)
		return err == nil
	})
}

func isBool() glisp.NamedUserFunction {
	return StringBool(func(s string) bool {
		return s == `true` || s == `false`
	})
}

func isAlpha() glisp.NamedUserFunction {
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

func isTitle() glisp.NamedUserFunction {
	return StringBool(func(s string) bool {
		if len(s) == 0 || !('A' <= s[0] && s[0] <= 'Z') {
			return false
		}
		if len(s) > 1 {
			return s[1] > 'Z' || s[1] < 'A'
		}
		return true
	})
}

func strMD5() glisp.NamedUserFunction {
	return StringMap(func(s string) string {
		return fmt.Sprintf("%x", md5.Sum([]byte(s)))
	})
}

func toStr(e glisp.Sexp) string {
	return string(e.(glisp.SexpStr))
}
