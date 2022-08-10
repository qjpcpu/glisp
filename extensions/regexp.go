package extensions

import (
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/qjpcpu/glisp"
)

func ImportRegex(env *glisp.Environment) {
	env.AddNamedFunction("regexp-compile", RegexpCompile)
	env.AddNamedFunction("regexp-find-index", RegexpFind)
	env.AddNamedFunction("regexp-find", RegexpFind)
	env.AddNamedFunction("regexp-match", RegexpFind)
}

type SexpRegexp struct {
	r *regexp.Regexp
}

func (re *SexpRegexp) SexpString() string {
	return fmt.Sprintf(`(regexp-compile %v)`, glisp.SexpStr(re.r.String()).SexpString())
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
				fmt.Errorf("2nd argument of %v should be a string", name)
		}

		var needle *regexp.Regexp
		switch t := args[0].(type) {
		case *SexpRegexp:
			needle = t.r
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

func RegexpCompile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}

		var re string
		switch t := args[0].(type) {
		case glisp.SexpStr:
			re = string(t)
		default:
			return glisp.SexpNull,
				errors.New("argument of regexp-compile should be a string")
		}

		if r, ok := regCache.Get(re); ok {
			return glisp.Sexp(&SexpRegexp{r: r}), nil
		}

		r, err := regexp.Compile(re)

		if err != nil {
			return glisp.SexpNull, errors.New(
				fmt.Sprintf("error during regexp-compile: '%v'", err))
		}

		regCache.Set(re, r)

		return glisp.Sexp(&SexpRegexp{r: r}), nil
	}
}

type regexpCache struct {
	rmap    map[string]*regexp.Regexp
	rw      *sync.RWMutex
	maxSize int
}

func (c *regexpCache) Get(k string) (*regexp.Regexp, bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	r, ok := c.rmap[k]
	return r, ok
}

func (c *regexpCache) Set(k string, v *regexp.Regexp) {
	c.rw.Lock()
	c.rw.Unlock()
	if len(c.rmap) > c.maxSize {
		c.rmap = make(map[string]*regexp.Regexp)
	}
	c.rmap[k] = v
}

var regCache = &regexpCache{
	rmap:    make(map[string]*regexp.Regexp),
	rw:      new(sync.RWMutex),
	maxSize: 100,
}
