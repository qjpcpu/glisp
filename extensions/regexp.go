package extensions

import (
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/qjpcpu/glisp"
)

func ImportRegex(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("regexp/compile", RegexpCompile)
	env.AddNamedFunction("regexp/find-index", RegexpFind)
	env.AddNamedFunction("regexp/find", RegexpFind)
	env.AddNamedFunction("regexp/match", RegexpFind)
	env.AddNamedFunction("regexp/replace", RegexpReplace)
	return nil
}

type SexpRegexp struct {
	r *regexp.Regexp
}

func (re *SexpRegexp) SexpString() string {
	return fmt.Sprintf(`(regexp/compile %v)`, glisp.SexpStr(re.r.String()).SexpString())
}

func regexpFindIndex(needle *regexp.Regexp, haystack string) (glisp.Sexp, error) {
	loc := needle.FindStringIndex(haystack)

	arr := make([]glisp.Sexp, len(loc))
	for i := range arr {
		arr[i] = glisp.Sexp(glisp.NewSexpInt(loc[i]))
	}

	return glisp.SexpArray(arr), nil
}

func RegexpFind(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		var haystack string
		switch t := args[1].(type) {
		case glisp.SexpStr:
			haystack = string(t)
		default:
			return glisp.SexpNull,
				fmt.Errorf("2nd argument of %v should be a string but got %v", name, glisp.InspectType(args[1]))
		}

		var needle *regexp.Regexp
		switch t := args[0].(type) {
		case *SexpRegexp:
			needle = t.r
		case glisp.SexpStr:
			reg, err := compileRegexp(string(t))
			if err != nil {
				return glisp.SexpNull, err
			}
			needle = reg.r
		default:
			return glisp.SexpNull,
				fmt.Errorf("1st argument of %v should be a compiled regular expression but got %v", name, glisp.InspectType(args[0]))
		}

		switch name {
		case "regexp/find":
			str := needle.FindString(haystack)
			return glisp.SexpStr(str), nil
		case "regexp/find-index":
			return regexpFindIndex(needle, haystack)
		case "regexp/match":
			matches := needle.MatchString(haystack)
			return glisp.SexpBool(matches), nil
		}

		return glisp.SexpNull, errors.New("unknown function")
	}
}

func RegexpReplace(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 3)
		}
		var needle *regexp.Regexp
		switch t := args[0].(type) {
		case *SexpRegexp:
			needle = t.r
		case glisp.SexpStr:
			reg, err := compileRegexp(string(t))
			if err != nil {
				return glisp.SexpNull, err
			}
			needle = reg.r
		default:
			return glisp.SexpNull,
				fmt.Errorf("1st argument of %v should be a compiled regular expression but got %v", name, glisp.InspectType(args[0]))
		}

		var src string
		switch t := args[1].(type) {
		case glisp.SexpStr:
			src = string(t)
		default:
			return glisp.SexpNull,
				fmt.Errorf("2nd argument of %v should be a string but got %v", name, glisp.InspectType(args[1]))
		}

		var repl string
		switch t := args[2].(type) {
		case glisp.SexpStr:
			repl = string(t)
		default:
			return glisp.SexpNull,
				fmt.Errorf("3nd argument of %v should be a string but got %v", name, glisp.InspectType(args[2]))
		}

		return glisp.SexpStr(needle.ReplaceAllString(src, repl)), nil
	}
}

func RegexpCompile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case glisp.SexpStr:
			re, err := compileRegexp(string(t))
			if err != nil {
				return glisp.SexpNull, err
			}
			return re, nil
		default:
			return glisp.SexpNull,
				fmt.Errorf("argument of regexp/compile should be a string but got %v", glisp.InspectType(args[0]))
		}
	}
}

func compileRegexp(re string) (*SexpRegexp, error) {
	if r, ok := regCache.Get(re); ok {
		return r, nil
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error during regexp/compile: '%v'", err))
	}

	sr := &SexpRegexp{r: r}
	regCache.Set(re, sr)

	return sr, nil
}

type regexpCache struct {
	rmap    map[string]*SexpRegexp
	rw      *sync.RWMutex
	maxSize int
}

func (c *regexpCache) Get(k string) (*SexpRegexp, bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	r, ok := c.rmap[k]
	return r, ok
}

func (c *regexpCache) Set(k string, v *SexpRegexp) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if len(c.rmap) > c.maxSize {
		c.rmap = make(map[string]*SexpRegexp)
	}
	c.rmap[k] = v
}

var regCache = &regexpCache{
	rmap:    make(map[string]*SexpRegexp),
	rw:      new(sync.RWMutex),
	maxSize: 100,
}
