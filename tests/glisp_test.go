package tests

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"

	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

const testDir = `.`

func TestAllScripts(t *testing.T) {
	for _, script := range listScripts(t) {
		testFile(t, script)
	}
}

func TestAllScriptsAgain(t *testing.T) {
	for _, script := range listScripts(t) {
		testFileAgain(t, script)
	}
}

func TestLoadAllFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	funcs := vm.GlobalFunctions()
	sort.Strings(funcs)
	t.Logf("all functions(%v)\n", len(funcs))
	var buf bytes.Buffer
	for i, f := range funcs {
		if i%10 == 0 {
			if buf.Len() > 0 {
				t.Logf("%s\n", buf.String())
			}
			buf.Reset()
		}
		buf.WriteString(`"` + f + `"` + "\t")
	}
	if buf.Len() > 0 {
		t.Logf("%s\n", buf.String())
	}
}

func TestOverrideBuiltin(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.AddFunction("len", func(v *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		return glisp.NewSexpInt(100), nil
	})
	ret, err := vm.EvalString(`(len [])`)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 100, ret)
}

func TestPrint(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	var buf bytes.Buffer
	vm.AddNamedFunction("print", extensions.GetPrintFunction(&buf))
	vm.LoadString(`(print "hello" 18446744073709551615)`)
	_, err := vm.Run()
	ExpectSuccess(t, err)
	ExpectEqStr(t, `hello18446744073709551615`, glisp.SexpStr(buf.String()))
}

func TestPrintln(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	var buf bytes.Buffer
	vm.AddNamedFunction("println", extensions.GetPrintFunction(&buf))
	vm.LoadString(`(println "hello" 18446744073709551615)`)
	_, err := vm.Run()
	ExpectSuccess(t, err)
	ExpectEqStr(t, "hello 18446744073709551615\n", glisp.SexpStr(buf.String()))
}

func TestPrintf(t *testing.T) {
	testPrintf(t, `(printf "%d" 10)`, `10`)
	testPrintf(t, `(printf "%v" 0.2)`, `0.2`)
	testPrintf(t, `(printf "%s" "hello")`, `hello`)
	testPrintf(t, `(printf "%v" #a)`, `97`)
	testPrintf(t, `(printf "%v" 0B37)`, `0B37`)
	testPrintf(t, `(printf "%v" '(1 2 3))`, `(1 2 3)`)
	testPrintf(t, `(printf "%v" [1 2 3])`, `[1 2 3]`)
	testPrintf(t, `(printf "%v" {'a 1})`, `{a 1}`)
	testPrintf(t, `(printf "%f%%" 3.14)`, `3.14%`)
}

func TestMakeScriptFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	fn, err := vm.MakeScriptFunction(`(+ %1 %2)`)
	ExpectSuccess(t, err)

	ret, err := vm.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, ret)

	fn, err = vm.MakeScriptFunction(`(str/start-with? %1 %2)`)
	ExpectSuccess(t, err)

	ret, err = vm.Apply(fn, []glisp.Sexp{glisp.SexpStr("abc"), glisp.SexpStr("a")})
	ExpectSuccess(t, err)
	ExpectTrue(t, ret)
}

func TestMakeScriptFunctionNoArgument(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	fn, err := vm.MakeScriptFunction(`(+ 1 2)`)
	ExpectSuccess(t, err)

	ret, err := vm.Apply(fn, nil)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, ret)
}

func TestMakeScriptFunctionArgumentNumber(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	fn, err := vm.MakeScriptFunction(`%N`)
	ExpectSuccess(t, err)

	ret, err := vm.Apply(fn, nil)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 0, ret)

	ret, err = vm.Apply(fn, []glisp.Sexp{glisp.NewSexpBytes(nil)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 1, ret)

	ret, err = vm.Apply(fn, []glisp.Sexp{glisp.NewSexpBytes(nil), glisp.NewSexpInt(1)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 2, ret)
}

func TestMakeComplexScriptFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	script := `
(defn afn [a b] (+ a b))
(defmac amac [a b] ESCAPE(* ~a ~b))
;; (%1 + %2) * (%4 - %3)
(amac (afn %1 %2) ((fn [a] (- a %3)) %4))
`
	fn, err := vm.MakeScriptFunction(strings.ReplaceAll(script, "ESCAPE", "`"))
	ExpectSuccess(t, err)

	ret, err := vm.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(2), glisp.NewSexpInt(3), glisp.NewSexpInt(5), glisp.NewSexpInt(7)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 10, ret)
}

func testPrintf(t *testing.T, script string, expect string) {
	vm := loadAllExtensions(glisp.New())
	var buf bytes.Buffer
	vm.AddNamedFunction("printf", extensions.GetPrintFunction(&buf))
	vm.LoadString(script)
	_, err := vm.Run()
	ExpectSuccess(t, err)
	ExpectEqStr(t, expect, glisp.SexpStr(buf.String()))
}

type testSexp struct{}

func (*testSexp) SexpString() string { return "xxxx" }

type testSexp2 struct{}

func (testSexp2) SexpString() string { return "yyyy" }

func TestCustomType(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	fn, _ := vm.FindObject("type")
	ret, err := vm.Apply(fn.(*glisp.SexpFunction), []glisp.Sexp{&testSexp{}})
	ExpectSuccess(t, err)
	ExpectEqStr(t, "*tests.testSexp", ret)
	ret, err = vm.Apply(fn.(*glisp.SexpFunction), []glisp.Sexp{testSexp2{}})
	ExpectSuccess(t, err)
	ExpectEqStr(t, "tests.testSexp2", ret)
}

func testFile(t *testing.T, file string) {
	bytes, err := ioutil.ReadFile(file)
	ExpectSuccess(t, err)
	vm := loadAllExtensions(glisp.New())
	err = vm.LoadString(string(bytes))
	ExpectSuccess(t, err)
	_, err = vm.Run()
	ExpectSuccess(t, err)
	t.Logf("TEST %s OK", file)
}

func testFileAgain(t *testing.T, file string) {
	testit := func(fn func([]glisp.Sexp) string, expressions []glisp.Sexp) {
		vm := loadAllExtensions(glisp.New())
		err := vm.LoadString(fn(expressions))
		ExpectSuccess(t, err)

		_, err = vm.Run()
		t.Log("==========", file, "============\n", fn(expressions))
		ExpectSuccess(t, err)
		t.Logf("TEST %s OK", file)
	}
	vm := loadAllExtensions(glisp.New())
	expressions, err := vm.ParseFile(file)
	ExpectSuccess(t, err)
	testit(glisp.FormatCompact, expressions)
	testit(glisp.FormatPretty, expressions)
}

func listScripts(t *testing.T) []string {
	files, err := ioutil.ReadDir(testDir)
	ExpectSuccess(t, err)

	var scripts []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".lisp") {
			scripts = append(scripts, filepath.Join(testDir, file.Name()))
		}
	}
	return scripts
}

func TestSandBox(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.PushGlobalScope()
	_, err := vm.EvalString(`(defn sb [a b ] (+ a b))`)
	ExpectSuccess(t, err)

	ret, err := vm.ApplyByName("sb", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, ret)

	vm.PopGlobalScope()
	_, err = vm.ApplyByName("sb", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectError(t, err, "function sb not found")
}

func TestWrapExpressionsAsFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	script := `(+ 2 3)`
	name := vm.GenSymbol("__anoy")
	script = fmt.Sprintf(`(defn %s [] %s)`, name.Name(), script)
	_, err := vm.EvalString(script)
	ExpectSuccess(t, err)

	sym, ok := vm.FindObject(name.Name())
	ExpectTrue(t, glisp.SexpBool(ok))

	ret, err := vm.Apply(sym.(*glisp.SexpFunction), nil)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 5, ret)
}

func TestApplyMessupPC0(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	_, err := vm.EvalString(`(+ 1 2)`)
	ExpectSuccess(t, err)
	expr, err := vm.ApplyByName("+", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)

	_, err = vm.EvalString(`(+ 1 2)`)
	ExpectSuccess(t, err)
	expr, err = vm.ApplyByName("+", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)
}

func TestApplyMessupPC(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	_, err := vm.EvalString(`(defn myfn [a b] (+ a b))`)
	ExpectSuccess(t, err)
	expr, err := vm.ApplyByName("myfn", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)

	_, err = vm.EvalString(`(defn myfn2 [a b] (+ a b))`)
	ExpectSuccess(t, err)
	expr, err = vm.ApplyByName("myfn2", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)

	expr, err = vm.EvalString(`(defn myfn3 [a b] (* a (myfn2 b 1))) (myfn2 2 3)`)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 5, expr)
	expr, err = vm.ApplyByName("myfn3", []glisp.Sexp{glisp.NewSexpInt(3), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 9, expr)

	expr, err = vm.EvalString(`(defn myfn3 [a b] (* a (myfn2 b 1))) (apply myfn2 [2 3])`)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 5, expr)
	expr, err = vm.ApplyByName("myfn3", []glisp.Sexp{glisp.NewSexpInt(3), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 9, expr)
}

func TestApplyMessupPCWithMacro(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	expr, err := vm.EvalString("(defmac myfnmac [a b] `(+ ~a ~b)) (defn myfn [a b] (myfnmac a b)) (+ 1 2)")
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)
	expr, err = vm.ApplyByName("myfn", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)

	expr, err = vm.EvalString("(defmac myfnmac2 [a b] `(+ ~a ~b)) (defn myfn2 [a b] (myfnmac2 a b)) (+ 1 2)")
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)
	expr, err = vm.ApplyByName("myfn2", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)

	expr, err = vm.EvalString("(defmac myfnmac3 [a b] `(+ ~a ~b)) (defn myfn3 [a b] (myfnmac3 (myfnmac2 a 1) (myfnmac b 1))) (+ 1 2)")
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, expr)
	expr, err = vm.ApplyByName("myfn3", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 5, expr)
}

func TestCrossApply(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.AddFunction("f0", func(env *glisp.Environment, expr []glisp.Sexp) (glisp.Sexp, error) {
		ret, err := env.ApplyByName("my-plus", expr)
		ExpectSuccess(t, err)
		return ret, err
	})
	_, err := vm.EvalString("(defn my-plus [a b] (+ a b))")
	ExpectSuccess(t, err)
	ret, err := vm.ApplyByName("f0", []glisp.Sexp{glisp.NewSexpInt(100), glisp.NewSexpInt(200)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 300, ret)
}

func TestCloneEnv(t *testing.T) {
	vm := glisp.New()
	vm2 := vm.Clone()
	vm.AddFunction("f0", func(env *glisp.Environment, expr []glisp.Sexp) (glisp.Sexp, error) {
		return glisp.NewSexpInt(100), nil
	})
	ret, err := vm.ApplyByName("f0", nil)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 100, ret)

	_, ok := vm2.FindObject("f0")
	ExpectFalse(t, glisp.SexpBool(ok))
}

func TestSourceFile(t *testing.T) {
	WithTempFile(`(+ 1 2)`, func(file string) {
		script := fmt.Sprintf(`(source-file '("%s"))`, file)
		ExpectScriptSuccess(t, script)
	})
}

func TestHTTPIncludeHeaderInOutput(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get '-i "%s")`, url)
		ExpectScriptSuccess(t, script, "HTTP/1.1 200 OK", `Content-Type: text/plain; charset=utf-8`, `"url":"/echo"`)
	})
}

func TestHTTPBadURL(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get '-i "%s")`, "ehttp://aa:db:4k")
		ExpectScriptErr(t, script, `build request fail parse "ehttp://aa:db:4k"`)
	})
}

func TestAddHeader(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get '-H {'header-aaa "header-value"} "%s")`, url)
		ExpectScriptErr(t, script, `-H option value must be a string but got "hash"`)
		script = fmt.Sprintf(`(http/get '-H "header-aaa: header-value" "%s")`, url)
		ExpectScriptSuccess(t, script, `"Header-Aaa":"header-value"`)
	})
}

func TestRequestFail(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get "%s")`, "http://127.0.0.1:1122")
		ExpectScriptErr(t, script, `dial tcp 127.0.0.1:1122: connect: connection refused`)
	})
}

func TestHTTPForm(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get '-H "content-type:application/x-www-form-urlencoded" "%s" '-d {"f1" "v1" 'f2 123})`, url)
		ExpectScriptSuccess(t, script, `"body":"f1=v1&f2=123"`, `"Content-Type":"application/x-www-form-urlencoded"`)
	})
}

func TestHTTPWithData(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get "%s" '-d "text")`, url)
		ExpectScriptSuccess(t, script, `"body":"text"`)

		script = fmt.Sprintf(`(http/get "%s" '-d (bytes "text"))`, url)
		ExpectScriptSuccess(t, script, `"body":"text"`)

		script = fmt.Sprintf(`(http/get "%s" '-d 123)`, url)
		ExpectScriptSuccess(t, script, `"body":"123"`)

		script = fmt.Sprintf(`(http/get "%s" '-d (fn [] 1))`, url)
		ExpectScriptErr(t, script, `build request fail bad value of -d`)

		script = fmt.Sprintf(`(http/get "%s" '-d '())`, url)
		ExpectScriptSuccess(t, script, `"body":""`)
	})
}

func TestHTTPWithJSON(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/post "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"body":"[1,2,3]"`, `"Content-Type":"application/json"`)

		script = fmt.Sprintf(`(http/post "%s" '-d '(1 2 3))`, url)
		ExpectScriptSuccess(t, script, `"body":"[1,2,3]"`, `"Content-Type":"application/json"`)

		script = fmt.Sprintf(`(http/post "%s" '-d {'k1 1 'k2 "str"})`, url)
		ExpectScriptSuccess(t, script, `"body":"{\"k1\":1,\"k2\":\"str\"}"`, `"Content-Type":"application/json"`)
	})
}

func TestHTTPWith404(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/post '-i "%s" '-d [1 2 3])`, url+"/xxx")
		ExpectScriptSuccess(t, script, `HTTP/1.1 404 Not Found`)
	})
}

func TestHTTPMethods(t *testing.T) {
	WithHttpServer(func(url string) {
		f, _ := os.CreateTemp(os.TempDir(), "http")
		oldstderr := os.Stderr
		os.Stderr = f
		defer func() {
			os.Stderr = oldstderr
			os.RemoveAll(f.Name())
		}()

		script := fmt.Sprintf(`(http/get "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		script = fmt.Sprintf(`(http/get '-X "POST" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		script = fmt.Sprintf(`(http/get '-v '-X "POST" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		script = fmt.Sprintf(`(http/post "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"POST"`)

		script = fmt.Sprintf(`(http/post '-X "GET" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"POST"`)

		script = fmt.Sprintf(`(http/put "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PUT"`)

		script = fmt.Sprintf(`(http/put '-X "GET" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PUT"`)

		script = fmt.Sprintf(`(http/patch "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PATCH"`)

		script = fmt.Sprintf(`(http/patch '-X "GET" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PATCH"`)

		script = fmt.Sprintf(`(http/head '-i "%s")`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/head '-X "GET" "%s")`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/delete "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"DELETE"`)

		script = fmt.Sprintf(`(http/delete '-X "GET" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"DELETE"`)

		script = fmt.Sprintf(`(http/options "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"OPTIONS"`)

		script = fmt.Sprintf(`(http/options '-X "GET" "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"OPTIONS"`)

		script = fmt.Sprintf(`(cdr (http/curl "%s" '-d [1 2 3]))`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		for _, method := range []string{`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`} {
			script = fmt.Sprintf(`(cdr (http/curl '-X "%s" "%s" '-d [1 2 3]))`, method, url)
			ExpectScriptSuccess(t, script, fmt.Sprintf(`"method":"%s"`, method))
		}
	})
}

func TestHTTPRaw(t *testing.T) {
	WithHttpServer(func(url string) {
		env := newFullEnv()
		script := fmt.Sprintf(`(car (http/curl '-X "POST" '-i "%s" '-d [1 2 3]))`, url+"/xxx")
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqInteger(t, 404, ret)
	})
}

func TestHTTPMultiHeaders(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get  "%s" '-H "h1:v1" '-H "h2:v2")`, url)
		ExpectScriptSuccess(t, script, `"H1":"v1"`, `"H2":"v2"`)
	})
}

func TestHTTPTimeout(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get  "%s" '-timeout 5)`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/get  "%s" '-timeout "1ns")`, url)
		ExpectScriptErr(t, script, `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)

		script = fmt.Sprintf(`(http/get  "%s" '-timeout (fn [] 1))`, url)
		ExpectScriptErr(t, script, `-timeout value should be integer or duration string such as`)

		script = fmt.Sprintf(`(http/get  "%s" '-timeout "xxxyyy")`, url)
		ExpectScriptErr(t, script, `bad -timeout`)
	})
}

func TestParseJSONStable(t *testing.T) {
	fn := func(data string) glisp.SexpStr {
		expr, _ := extensions.ParseJSON([]byte(data))
		out, _ := glisp.Marshal(expr)
		return glisp.SexpStr(string(out))
	}
	for i := 0; i < 1000; i++ {
		orig := `{"a":1,"b":2,3:"c"}`
		ExpectEqStr(t, orig, fn(orig))
		orig = `{"b":2,3:"c","a":1}`
		ExpectEqStr(t, orig, fn(orig))
	}
}

func TestParseJSON(t *testing.T) {
	fn := func(data string) glisp.SexpStr {
		expr, _ := extensions.ParseJSON([]byte(data))
		out, _ := glisp.Marshal(expr)
		return glisp.SexpStr(string(out))
	}
	orig := `{"a":null,"b":"val","c":[],"d":[null],"e":false,"f":12,"g":3.14,true:1}`
	ExpectEqStr(t, orig, fn(orig))
	orig = `null`
	ExpectEqStr(t, orig, fn(orig))
}

func TestJSONPrettyPrint(t *testing.T) {
	testFn := func(expect, script string) {
		var buf bytes.Buffer
		vm := loadAllExtensions(glisp.New())
		vm.AddNamedFunction("print", extensions.GetPrintFunction(&buf))
		vm.AddNamedFunction("println", extensions.GetPrintFunction(&buf))
		vm.AddNamedFunction("printf", extensions.GetPrintFunction(&buf))
		_, err := vm.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqStr(t, strings.TrimSpace(expect), glisp.SexpStr(strings.TrimSpace(buf.String())))
	}
	testFn(
		`1
true
3.14
null
"str"
`,
		`(json/q 1) (json/q true) (json/q 3.14) (json/q '()) (json/q "str")`,
	)
	testFn(
		`[
    1
    2
    3
    ......
    <len=3>
]`,
		`(json/q [1 2 3])`,
	)
	testFn(
		`{
    "a": 1
    "b": 2
    "c": [<len=3>]
    "d": {"a1","b1"}
}`,
		`(json/q {"a" 1 "b" 2 "c" [1 2 3] "d" {"a1" 1 "b1" 2}})`,
	)
	testFn(
		`{
    "a": "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-...<len=68>"
}`,
		`(json/q {"a" "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-*/%$"})`,
	)
	testFn(
		`{
  "a": "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-*/%$"
}`,
		`(json/q {"a" "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-*/%$"} true)`,
	)
}

func TestDoc(t *testing.T) {
	testDoc := func(script, docstr string) {
		vm := loadAllExtensions(glisp.New())
		var buf bytes.Buffer
		vm.AddNamedFunction("println", extensions.GetPrintFunction(&buf))
		_, err := vm.EvalString(script)
		ExpectSuccess(t, err)
		ExpectContainsStr(t, glisp.SexpStr(buf.String()), docstr)
	}
	testDoc(`(defmac xyz [] "doc-xyz" 1) (doc xyz)`, `doc-xyz`)
	testDoc(`(defn xyz [] "doc-xyz" 1) (doc xyz)`, `doc-xyz`)
	testDoc(`(defn xyz [] 1) (doc xyz)`, `No document found.`)
}
