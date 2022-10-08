package tests

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

func newFullEnv() *glisp.Environment {
	return loadAllExtensions(glisp.New())
}

func loadAllExtensions(vm *glisp.Environment) *glisp.Environment {
	imports := []func(*glisp.Environment) error{
		func(e *glisp.Environment) error {
			return e.ImportEval()
		},
		extensions.ImportCoreUtils,
		extensions.ImportContainerUtils,
		extensions.ImportJSON,
		extensions.ImportMathUtils,
		extensions.ImportBase64,
		extensions.ImportChannels,
		extensions.ImportCoroutines,
		extensions.ImportRandom,
		extensions.ImportRegex,
		extensions.ImportTime,
		extensions.ImportString,
		extensions.ImportOS,
		extensions.ImportHTTP,
		extensions.ImportStream,
	}
	for _, fn := range imports {
		if err := fn(vm); err != nil {
			panic(err)
		}
	}
	vm.AddFunction("my-counter", MakeCounter)
	vm.AddFunction("err-stream", MakeErrStream)
	vm.AddFunction("get-my-counter", GetCounterNumber)
	return vm
}

func ExpectEqInteger(t *testing.T, expect int64, expr glisp.Sexp) {
	if !glisp.IsInt(expr) {
		t.Log(getTestStack())
		t.Fatalf("should get integer but got %#T", expr)
	}
	if v := expr.(glisp.SexpInt).ToInt64(); v != expect {
		t.Log(getTestStack())
		t.Fatalf("should get %v but got %v", expect, v)
	}
}

func ExpectEqStr(t *testing.T, expect string, expr glisp.Sexp) {
	if !glisp.IsString(expr) {
		t.Log(getTestStack())
		t.Fatalf("should get string but got %#T", expr)
	}
	if v := string(expr.(glisp.SexpStr)); v != expect {
		t.Log(getTestStack())
		t.Fatalf("should get %v but got %v", expect, v)
	}
}

func ExpectContainsStr(t *testing.T, expr glisp.Sexp, keywords ...string) {
	if !glisp.IsString(expr) {
		t.Log(getTestStack())
		t.Fatalf("should get string but got %#T", expr)
	}
	for _, key := range keywords {
		if v := string(expr.(glisp.SexpStr)); !strings.Contains(v, key) {
			t.Log(getTestStack())
			t.Fatalf("should contains %v but got %v", key, v)
		}
	}
}

func ExpectTrue(t *testing.T, expr glisp.Sexp) {
	ExpectEqBool(t, true, expr)
}

func ExpectFalse(t *testing.T, expr glisp.Sexp) {
	ExpectEqBool(t, false, expr)
}

func ExpectEqBool(t *testing.T, expect bool, expr glisp.Sexp) {
	if !glisp.IsBool(expr) {
		t.Log(getTestStack())
		t.Fatalf("should get bool but got %#T", expr)
	}
	if v := bool(expr.(glisp.SexpBool)); v != expect {
		t.Log(getTestStack())
		t.Fatalf("should get %v but got %v", expect, v)
	}
}

func ExpectSuccess(t *testing.T, err error) {
	if err != nil {
		t.Log(getTestStack())
		t.Fatalf("should no error but got err %v", err)
	}
}

func ExpectError(t *testing.T, err error, keywords ...string) {
	if err == nil {
		t.Log(getTestStack())
		t.Fatalf("should get error but success")
	}
	for _, key := range keywords {
		if !strings.Contains(err.Error(), key) || key == `` {
			t.Log(getTestStack())
			t.Fatalf("error message should contains (%s), err: %v", key, err.Error())
		}
	}
}

func ExpectEmptyStr(t *testing.T, str string) {
	if str != "" {
		t.Log(getTestStack())
		t.Fatalf("expect empty string but got %s", str)
	}
}

func ExpectNonEmptyStr(t *testing.T, str string) {
	if str == "" {
		t.Log(getTestStack())
		t.Fatalf("expect non-empty string but got empty")
	}
}

func ExpectScriptErr(t *testing.T, script string, keywords ...string) {
	env := newFullEnv()
	_, err := env.EvalString(script)
	ExpectError(t, err, keywords...)
	env.DumpEnvironment(ioutil.Discard)
}

func ExpectScriptSuccess(t *testing.T, script string, keywords ...string) {
	env := newFullEnv()
	ret, err := env.EvalString(script)
	ExpectSuccess(t, err)
	if len(keywords) > 0 {
		if glisp.IsBytes(ret) {
			ret = glisp.SexpStr(string(ret.(glisp.SexpBytes).Bytes()))
		}
		ExpectContainsStr(t, ret, keywords...)
	}
}

func getTestStack() string {
	bs := debug.Stack()
	arr := strings.Split(string(bs), "\n")
	i, j := len(arr)-1, len(arr)-1
	for ; i >= 0; i-- {
		if strings.Contains(arr[i], `testing.tRunner`) {
			j = i
		}
		if strings.Contains(arr[i], `glisp/tests/util.go`) {
			i = i + 1
			break
		}
	}
	return strings.Join(arr[i:j], "\n")
}

func WithTempFile(content string, cb func(string)) {
	file, err := ioutil.TempFile(os.TempDir(), "glisp")
	if err != nil {
		return
	}
	file.WriteString(content)
	name := file.Name()
	file.Close()
	defer os.RemoveAll(name)
	cb(name)
}

func WithHttpServer(fn func(string)) {
	server := NewMockServer()
	defer server.ServeBackground()()
	url := server.URLPrefix + `/echo`
	fn(url)
}

type Counter struct {
	Len    int
	cursor int
}

func (c *Counter) SexpString() string {
	return fmt.Sprintf(`(my-counter %v)`, c.Len)
}

func MakeCounter(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	return &Counter{Len: args[0].(glisp.SexpInt).ToInt()}, nil
}

func GetCounterNumber(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	num := args[0].(*Counter).cursor
	return glisp.NewSexpInt(num), nil
}

func (c *Counter) Next() (glisp.Sexp, bool) {
	if c.cursor < c.Len {
		c.cursor++
		return glisp.NewSexpInt(c.cursor), true
	}
	return glisp.SexpNull, false
}

func MakeErrStream(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	msg := "error occur"
	if len(args) > 0 {
		msg = string(args[0].(glisp.SexpStr))
	}
	return &errStream{msg: msg}, nil
}

type errStream struct {
	msg string
}

func (e *errStream) SexpString() string { return fmt.Sprintf(`(err-stream "%s")`, e.msg) }

func (e *errStream) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	return glisp.SexpNull, false, errors.New(e.msg)
}
