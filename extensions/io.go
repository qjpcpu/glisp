package extensions

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qjpcpu/glisp"
)

func ImportIO(env *glisp.Environment) {
	env.AddFunction("io/read-file", ReadFile)
	env.AddFunction("io/write-file", WriteFile)
	env.AddFunction("io/file-exist?", ExistFile)
	env.AddFunction("io/remove-file", RemoveFile)
}

func ReadFile(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.CallName()
	if len(args) != 1 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpStr)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
	}
	filename := replaceHomeDirSymbol(string(str))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return glisp.SexpNull, err
	}
	return glisp.NewSexpBytes(data), nil
}

func WriteFile(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.CallName()
	if len(args) != 2 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpStr)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
	}
	filename := replaceHomeDirSymbol(string(str))
	if _, err := os.Stat(filepath.Dir(filename)); err != nil && os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(filename), 0755)
	}
	switch data := args[1].(type) {
	case glisp.SexpStr:
		return glisp.SexpNull, ioutil.WriteFile(filename, []byte(data), 0644)
	case glisp.SexpBytes:
		return glisp.SexpNull, ioutil.WriteFile(filename, data.Bytes(), 0644)
	default:
		return glisp.SexpNull, fmt.Errorf("%s expect write string/bytes to file", name)
	}
}

func RemoveFile(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.CallName()
	if len(args) != 1 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpStr)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
	}
	filename := replaceHomeDirSymbol(string(str))
	return glisp.SexpNull, os.RemoveAll(filename)
}

func ExistFile(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.CallName()
	if len(args) != 1 {
		return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
	}
	str, ok := args[0].(glisp.SexpStr)
	if !ok {
		return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
	}
	filename := replaceHomeDirSymbol(string(str))
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return glisp.SexpBool(false), nil
	}
	return glisp.SexpBool(true), nil
}

func replaceHomeDirSymbol(file string) string {
	if strings.HasPrefix(file, `~`) {
		if dir, err := os.UserHomeDir(); err == nil {
			file = filepath.Join(dir, strings.TrimPrefix(file, `~`))
		}
	}
	return file
}
