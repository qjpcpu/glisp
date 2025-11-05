package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

func newFullEnv() *glisp.Environment {
	return loadAllExtensions(glisp.New())
}

func loadAllExtensions(vm *glisp.Environment) *glisp.Environment {
	if err := extensions.ImportAll(vm); err != nil {
		panic(err)
	}
	vm.AddFunction("my-counter", MakeCounter)
	vm.RegisterType("stream", &Counter{})
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

func ExpectEqHashKV(t *testing.T, hash glisp.Sexp, key string, expectVal string) {
	if !glisp.IsHash(hash) {
		t.Log(getTestStack())
		t.Fatalf("should get hash but got %#T", hash)
	}
	h := hash.(*glisp.SexpHash)
	value, _ := h.HashGetDefault(glisp.SexpStr(key), glisp.SexpNull)
	if v := string(value.(glisp.SexpStr)); v != expectVal {
		t.Log(getTestStack())
		t.Fatalf("should get %v but got %v", expectVal, v)
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

func ExpectPanic(t *testing.T, fn func()) {
	f := func() (r any) {
		defer func() {
			r = recover()
		}()
		fn()
		return
	}
	r := f()
	if r == nil {
		t.Fatal("expect panic but success")
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

func ExpectEqString(t *testing.T, s1, s2 string) {
	if s1 != s2 {
		t.Fatalf("%s != %s", s1, s2)
	}
}

func ExpectEqAny(t *testing.T, s1, s2 any) {
	if !reflect.DeepEqual(s1, s2) {
		t.Fatalf("%v != %v", s1, s2)
	}
}

func ExpectEqList(t *testing.T, s1, s2 []any) {
	if len(s1) != len(s2) {
		t.Fatalf("%v != %v", s1, s2)
	}
	j1, _ := json.Marshal(s1)
	j2, _ := json.Marshal(s2)
	if string(j1) != string(j2) {
		t.Fatalf("%v != %v", s1, s2)
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

func WithHttpServer2(fn func(string, string)) {
	server := NewMockServer()
	defer server.ServeBackground()()
	fn(server.Address, "/echo")
}

type Counter struct {
	Len    int
	cursor int
}

func (c *Counter) SexpString() string {
	return fmt.Sprintf(`(my-counter %v)`, c.Len)
}

func MakeCounter(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
	return &Counter{Len: args.Get(0).(glisp.SexpInt).ToInt()}, nil
}

func GetCounterNumber(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
	num := args.Get(0).(*Counter).cursor
	return glisp.NewSexpInt(num), nil
}

func (c *Counter) Explain(env *glisp.Environment, n string, args glisp.Args) (glisp.Sexp, error) {
	return glisp.SexpStr("OK"), nil
}

func (c *Counter) Next() (glisp.Sexp, bool) {
	if c.cursor < c.Len {
		c.cursor++
		return glisp.NewSexpInt(c.cursor), true
	}
	return glisp.SexpNull, false
}

func MakeErrStream(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
	msg := "error occur"
	if args.Len() > 0 {
		msg = string(args.Get(0).(glisp.SexpStr))
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

type runeReader struct {
	input chan rune
	stopc chan struct{}
}

func NewRuneReader() *runeReader {
	return &runeReader{input: make(chan rune, 1024), stopc: make(chan struct{})}
}

func (rr *runeReader) ReadRune() (r rune, size int, err error) {
	select {
	case r = <-rr.input:
		size = utf8.RuneLen(r)
	case <-rr.stopc:
		err = io.EOF
		return
	}
	return
}

type repl struct {
	reader *runeReader
	lexer  *glisp.Lexer
	env    *glisp.Environment
}

func NewREPL() *repl {
	env := glisp.New()
	extensions.ImportAll(env)
	stream := NewRuneReader()
	lexer := glisp.NewLexerFromStream(stream)
	return &repl{
		reader: stream,
		lexer:  lexer,
		env:    env,
	}
}

func (self *repl) InputLine(str string) {
	bs := []rune(str)
	for _, b := range bs {
		self.reader.input <- b
	}
}

func (self *repl) REPLOnce() (glisp.Sexp, error) {
	expr, err := glisp.ParseExpression(glisp.NewParser(self.lexer, self.env))
	if err != nil {
		return glisp.SexpNull, errors.New(self.env.GetStackTrace(err))
	}
	if err = self.env.LoadExpressions([]glisp.Sexp{expr}); err != nil {
		return glisp.SexpNull, errors.New(self.env.GetStackTrace(err))
	}
	ret, err := self.env.Run()
	if err != nil {
		return glisp.SexpNull, errors.New(self.env.GetStackTrace(err))
	}
	return ret, err
}

func (self *repl) Close() {
	close(self.reader.stopc)
}
