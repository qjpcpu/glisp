package extensions

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/qjpcpu/glisp"
)

func ImportString(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("str/start-with?", StringPredict(strings.HasPrefix))
	env.AddNamedFunction("str/end-with?", StringPredict(strings.HasSuffix))
	env.AddNamedFunction("str/contains?", StringPredict(strings.Contains))
	env.AddNamedFunction("str/equal-fold?", StringPredict(strings.EqualFold))
	env.AddNamedFunction("str/title", StringMap(strings.Title))
	env.AddNamedFunction("str/lower", StringMap(strings.ToLower))
	env.AddNamedFunction("str/upper", StringMap(strings.ToUpper))
	env.AddNamedFunction("str/replace", StringReplace)
	env.AddNamedFunction("str/repeat", StringRepeat)
	env.AddNamedFunction("str/trim-prefix", StringMap2(strings.TrimPrefix))
	env.AddNamedFunction("str/trim-suffix", StringMap2(strings.TrimSuffix))
	env.AddNamedFunction("str/trim-space", StringMap(strings.TrimSpace))
	env.AddNamedFunction("str/count", StringSearch(strings.Count))
	env.AddNamedFunction("str/index", StringSearch(stringIndex))
	env.AddNamedFunction("str/split", StringSplit)
	env.AddNamedFunction("str/join", StringJoin(strings.Join))
	env.AddNamedFunction("str/digit?", isDigit())
	env.AddNamedFunction("str/alpha?", isAlpha())
	env.AddNamedFunction("str/title?", isTitle())
	env.AddNamedFunction("str/integer?", isInteger())
	env.AddNamedFunction("str/float?", isFloat())
	env.AddNamedFunction("str/bool?", isBool())
	env.AddNamedFunction("str/md5", strMD5())
	env.AddNamedFunction("str/sha256", strSHA256())
	env.AddNamedFunction("str/mask", strMask)
	mustLoadScript(env.Environment, "str")
	return nil
}

func StringPredict(fn func(string, string) bool) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 2 && args.Len() != 3 {
				return glisp.WrongNumberArguments(name, args.Len(), 2, 3)
			}
			if !glisp.IsString(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			if !glisp.IsString(args.Get(1)) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
			}
			s, substr := toStr(args.Get(0)), toStr(args.Get(1))
			if args.Len() == 3 && glisp.IsBool(args.Get(2)) && bool(args.Get(2).(glisp.SexpBool)) {
				s = strings.ToLower(s)
				substr = strings.ToLower(substr)
			}
			return glisp.SexpBool(fn(s, substr)), nil
		}
	}
}

func StringSearch(fn func(string, string) int) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 2 {
				return glisp.WrongNumberArguments(name, args.Len(), 2)
			}
			if !glisp.IsString(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			if !glisp.IsString(args.Get(1)) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
			}
			return glisp.NewSexpInt((fn(toStr(args.Get(0)), toStr(args.Get(1))))), nil
		}
	}
}

func StringSplit(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 && args.Len() != 3 {
			return glisp.WrongNumberArguments(name, args.Len(), 2, 3)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsString(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
		}
		var list []string
		if args.Len() == 3 && glisp.IsInt(args.Get(2)) {
			list = strings.SplitN(toStr(args.Get(0)), toStr(args.Get(1)), args.Get(2).(glisp.SexpInt).ToInt())
		} else {
			list = strings.Split(toStr(args.Get(0)), toStr(args.Get(1)))
		}
		var array glisp.SexpArray
		for _, str := range list {
			array = append(array, glisp.SexpStr(str))
		}
		return array, nil
	}
}

func StringMap(fn func(string) string) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 1 {
				return glisp.WrongNumberArguments(name, args.Len(), 1)
			}
			if !glisp.IsString(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			return glisp.SexpStr((fn(toStr(args.Get(0))))), nil
		}
	}
}

func StringBool(fn func(string) bool) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 1 {
				return glisp.WrongNumberArguments(name, args.Len(), 1)
			}
			if !glisp.IsString(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			return glisp.SexpBool((fn(toStr(args.Get(0))))), nil
		}
	}
}

func StringJoin(fn func([]string, string) string) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 2 {
				return glisp.WrongNumberArguments(name, args.Len(), 2)
			}
			if args.Get(0) == glisp.SexpNull {
				return glisp.SexpStr(""), nil
			}
			var items []string
			switch list := args.Get(0).(type) {
			case glisp.SexpArray:
				for _, v := range list {
					if !glisp.IsString(v) {
						return glisp.SexpNull, fmt.Errorf(`%s first argument should be array of string but got %v`, name, glisp.InspectType(v))
					}
					items = append(items, string(v.(glisp.SexpStr)))
				}
			case *glisp.SexpPair:
				var err error
				list.Foreach(func(v glisp.Sexp) bool {
					if !glisp.IsString(v) {
						err = fmt.Errorf(`%s first argument should be list of string but got %v`, name, glisp.InspectType(v))
						return false
					}
					items = append(items, string(v.(glisp.SexpStr)))
					return true
				})
				if err != nil {
					return glisp.SexpNull, err
				}
			default:
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be array/list but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			if !glisp.IsString(args.Get(1)) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
			}

			return glisp.SexpStr(fn(items, string(args.Get(1).(glisp.SexpStr)))), nil
		}
	}
}

func StringMap2(fn func(string, string) string) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 2 {
				return glisp.WrongNumberArguments(name, args.Len(), 2)
			}
			if !glisp.IsString(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			if !glisp.IsString(args.Get(1)) {
				return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
			}
			return glisp.SexpStr(fn(toStr(args.Get(0)), toStr(args.Get(1)))), nil
		}
	}
}

func StringReplace(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 3 && args.Len() != 4 {
			return glisp.WrongNumberArguments(name, args.Len(), 3, 4)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsString(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(1)))
		}
		if !glisp.IsString(args.Get(2)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(2)))
		}
		if args.Len() == 4 && glisp.IsInt(args.Get(3)) {
			return glisp.SexpStr(strings.Replace(toStr(args.Get(0)), toStr(args.Get(1)), toStr(args.Get(2)), args.Get(3).(glisp.SexpInt).ToInt())), nil
		}
		return glisp.SexpStr(strings.ReplaceAll(toStr(args.Get(0)), toStr(args.Get(1)), toStr(args.Get(2)))), nil
	}
}

func StringRepeat(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsInt(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be int but got %v`, name, glisp.InspectType(args.Get(1)))
		}
		return glisp.SexpStr(strings.Repeat(toStr(args.Get(0)), args.Get(1).(glisp.SexpInt).ToInt())), nil
	}
}

func strMask(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 4 {
			return glisp.WrongNumberArguments(name, args.Len(), 4)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf(`%s first argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsInt(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be integer but got %v`, name, glisp.InspectType(args.Get(1)))
		}
		if !glisp.IsInt(args.Get(2)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be integer but got %v`, name, glisp.InspectType(args.Get(2)))
		}
		if !glisp.IsString(args.Get(3)) {
			return glisp.SexpNull, fmt.Errorf(`%s second argument should be string but got %v`, name, glisp.InspectType(args.Get(3)))
		}
		str := []rune(args.Get(0).(glisp.SexpStr))
		index := args.Get(1).(glisp.SexpInt).ToInt()
		length := args.Get(2).(glisp.SexpInt).ToInt()
		mask := string(args.Get(3).(glisp.SexpStr))
		if index < 0 || index >= len(str) {
			/* mask nothing */
			return args.Get(0), nil
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

func strSHA256() glisp.NamedUserFunction {
	return StringMap(func(s string) string {
		return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
	})
}

func toStr(e glisp.Sexp) string {
	return string(e.(glisp.SexpStr))
}

func stringIndex(s, substr string) int {
	idx := strings.Index(s, substr)
	if idx != -1 {
		return utf8.RuneCountInString(s[:idx])
	}
	return idx
}
