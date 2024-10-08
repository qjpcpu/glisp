package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
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
	testPrintf(t, `(printf "%%%f%%%%" 3.14)`, `%3.14%%`)
	testPrintf(t, `(printf "%.2f%%" 3.1415926)`, `3.14%`)
	testPrintf(t, `(printf "%b" 100)`, `1100100`)
	testPrintf(t, `(printf "%v:%v" "a" "b" "c" 1)`, `a:b%!(EXTRA string=c, int=1)`)
	testPrintf(t, `(printf "%v" "a"  1)`, `a%!(EXTRA int=1)`)
	testPrintf(t, `(printf "%v:%v" "a")`, `a:%!v(MISSING)`)
	testPrintf(t, `(printf "%v:%v:%v" "a")`, `a:%!v(MISSING):%!v(MISSING)`)
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
	bytes, err := os.ReadFile(file)
	ExpectSuccess(t, err)
	t.Logf("TESTing %s", file)
	vm := loadAllExtensions(glisp.New())
	err = vm.LoadString(string(bytes))
	ExpectSuccess(t, err)
	_, err = vm.Run()
	ExpectSuccess(t, err)
	t.Logf("TEST %s OK", file)
}

func listScripts(t *testing.T) []string {
	files, err := os.ReadDir(testDir)
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
	name := vm.GenSymbol("__anon")
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
		script := fmt.Sprintf(`(car (http/get -i "%s"))`, url)
		env := newFullEnv()
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqHashKV(t, ret, "StatusCode", "200")
		ExpectEqHashKV(t, ret, "Status", "200 OK")
	})
}

func TestMultiHTTP(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(->> (http/get ["%s" "%s"]) (map json/parse) (list-to-array))`, url, url)
		env := newFullEnv()
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		arr := ret.(glisp.SexpArray)
		ExpectEqInteger(t, int64(2), glisp.NewSexpInt(len(arr)))
	})
}

func TestHTTPProxy(t *testing.T) {
	WithHttpServer2(func(addr, path string) {
		env := newFullEnv()
		env.AddFunction("proxy", func(_ *glisp.Environment, a []glisp.Sexp) (glisp.Sexp, error) {
			return extensions.MakeDialer(func(_ context.Context, _ string, _ string) (net.Conn, error) {
				return net.Dial("tcp", addr)
			}), nil
		})
		script := fmt.Sprintf(`(car (http/get -i "http://www.any.com%s" -x (proxy)))`, path)
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqHashKV(t, ret, "StatusCode", "200")
		ExpectEqHashKV(t, ret, "Status", "200 OK")
		ExpectEqHashKV(t, ret, "Content-Type", "text/plain; charset=utf-8")

		script = fmt.Sprintf(`(def cli (http/get -i -x (proxy))) (cli "http://www.any.com%s")`, path)
		ret, err = env.EvalString(script)
		ExpectSuccess(t, err)
		header := ret.(*glisp.SexpPair).Head()
		body := ret.(*glisp.SexpPair).Tail()
		ExpectEqHashKV(t, header, "StatusCode", "200")
		ExpectEqHashKV(t, header, "Status", "200 OK")
		ExpectEqHashKV(t, header, "Content-Type", "text/plain; charset=utf-8")
		ExpectContainsStr(t, glisp.SexpStr(string(body.(glisp.SexpBytes).Bytes())), `"url":"/echo"`)
	})
}

func TestHTTPBadURL(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get -i "%s")`, "httptx://aa:db:4k")
		ExpectScriptErr(t, script, `build request fail parse`)
	})
}

func TestAddHeader(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get -H {'header-aaa "header-value"} "%s")`, url)
		ExpectScriptErr(t, script, `-H option value must be a string but got "hash"`)
		script = fmt.Sprintf(`(http/get -H "header-aaa: header-value" "%s")`, url)
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
		script := fmt.Sprintf(`(http/get -H "content-type:application/x-www-form-urlencoded" "%s" -d {"f1" "v1" 'f2 123})`, url)
		ExpectScriptSuccess(t, script, `"body":"f1=v1&f2=123"`, `"Content-Type":"application/x-www-form-urlencoded"`)
	})
}

func TestHTTPWithData(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get "%s" -d "text")`, url)
		ExpectScriptSuccess(t, script, `"body":"text"`)

		script = fmt.Sprintf(`(http/get "%s" -d (bytes "text"))`, url)
		ExpectScriptSuccess(t, script, `"body":"text"`)

		script = fmt.Sprintf(`(http/get "%s" -d 123)`, url)
		ExpectScriptSuccess(t, script, `"body":"123"`)

		script = fmt.Sprintf(`(http/get "%s" -d (fn [] 1))`, url)
		ExpectScriptErr(t, script, `build request fail bad value of -d`)

		script = fmt.Sprintf(`(http/get "%s" -d '())`, url)
		ExpectScriptSuccess(t, script, `"body":""`)
	})
}

func TestHTTPWithJSON(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/post "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"body":"[1,2,3]"`, `"Content-Type":"application/json"`)

		script = fmt.Sprintf(`(http/post "%s" -d '(1 2 3))`, url)
		ExpectScriptSuccess(t, script, `"body":"[1,2,3]"`, `"Content-Type":"application/json"`)

		script = fmt.Sprintf(`(http/post "%s" -d {'k1 1 'k2 "str"})`, url)
		ExpectScriptSuccess(t, script, `"body":"{\"k1\":1,\"k2\":\"str\"}"`, `"Content-Type":"application/json"`)
	})
}

func TestHTTPWith404(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/post -i "%s" '-d [1 2 3])`, url+"/xxx")
		env := newFullEnv()
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		header := ret.(*glisp.SexpPair).Head()
		ExpectEqHashKV(t, header, "StatusCode", "404")
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

		script := fmt.Sprintf(`(http/get "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		fi, _ := os.CreateTemp(os.TempDir(), "http2")
		file := fi.Name()
		script = fmt.Sprintf(`(http/get -o "%s" "%s" -d [1 2 3])`, file, url)
		ExpectScriptSuccess(t, script)
		data, _ := os.ReadFile(file)
		ExpectEqString(t, string(data), `{"args":{},"body":"[1,2,3]","headers":{"Accept-Encoding":"gzip","Content-Length":"7","Content-Type":"application/json","User-Agent":"Go-http-client/1.1"},"method":"GET","url":"/echo"}`)
		os.RemoveAll(file)

		script = fmt.Sprintf(`(http/get -X "POST" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		script = fmt.Sprintf(`(http/get -v -X "POST" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		script = fmt.Sprintf(`(http/post "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"POST"`)

		script = fmt.Sprintf(`(http/post -X "GET" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"POST"`)

		script = fmt.Sprintf(`(http/put "%s" '-d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PUT"`)

		script = fmt.Sprintf(`(http/put -X "GET" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PUT"`)

		script = fmt.Sprintf(`(http/patch "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PATCH"`)

		script = fmt.Sprintf(`(http/patch -X "GET" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"PATCH"`)

		script = fmt.Sprintf(`(http/head -i "%s")`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/head -X "GET" "%s")`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/delete "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"DELETE"`)

		script = fmt.Sprintf(`(http/delete -X "GET" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"DELETE"`)

		script = fmt.Sprintf(`(http/options "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"OPTIONS"`)

		script = fmt.Sprintf(`(http/options -X "GET" "%s" -d [1 2 3])`, url)
		ExpectScriptSuccess(t, script, `"method":"OPTIONS"`)

		script = fmt.Sprintf(`(cdr (http/curl "%s" -d [1 2 3]))`, url)
		ExpectScriptSuccess(t, script, `"method":"GET"`)

		for _, method := range []string{`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`} {
			script = fmt.Sprintf(`(cdr (http/curl -X "%s" "%s" -d [1 2 3]))`, method, url)
			ExpectScriptSuccess(t, script, fmt.Sprintf(`"method":"%s"`, method))
		}
	})
}

func TestHTTPRaw(t *testing.T) {
	WithHttpServer(func(url string) {
		env := newFullEnv()
		script := fmt.Sprintf(`(car (http/curl -X "POST" -i "%s" -d [1 2 3]))`, url+"/xxx")
		ret, err := env.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqInteger(t, 404, ret)
	})
}

func TestHTTPMultiHeaders(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get  "%s" -H "h1:v1" -H "h2:v2")`, url)
		ExpectScriptSuccess(t, script, `"H1":"v1"`, `"H2":"v2"`)

		script = fmt.Sprintf(`(def args (list  "%s" '-H "h1:v1" '-H "h2:v2")) (apply http/get args)`, url)
		ExpectScriptSuccess(t, script, `"H1":"v1"`, `"H2":"v2"`)
	})
}

func TestHTTPTimeout(t *testing.T) {
	WithHttpServer(func(url string) {
		script := fmt.Sprintf(`(http/get  "%s" -timeout 5)`, url)
		ExpectScriptSuccess(t, script)

		script = fmt.Sprintf(`(http/get  "%s" -timeout "1ns")`, url)
		ExpectScriptErr(t, script, `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)

		script = fmt.Sprintf(`(http/get  "%s" -timeout (fn [] 1))`, url)
		ExpectScriptErr(t, script, `-timeout value should be integer or duration string such as`)

		script = fmt.Sprintf(`(http/get  "%s" -timeout "xxxyyy")`, url)
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

func TestListBuilder(t *testing.T) {
	if glisp.NewListBuilder().Get() != glisp.SexpNull {
		t.Fatal("should get null expr")
	}
	b := glisp.NewListBuilder()
	expr := b.Add(glisp.NewSexpInt(1)).
		Add(glisp.NewSexpInt(2)).
		Get().
		SexpString()
	if expr != `(1 2)` {
		t.Fatal("should get list")
	}
}

func TestListFuzzyMacro(t *testing.T) {
	testMacro := func(script string) {
		vm := loadAllExtensions(glisp.New())
		vm.AddFuzzyMacro(`^fuzzy.+$`, func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			/* always return 1024 */
			return glisp.NewSexpInt(1024), nil
		})
		ret, err := vm.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqInteger(t, 1024, ret)
	}
	testMacro(`(fuzzy-x 1 2 3)`)
	testMacro(`(fuzzyAny 1 2 3)`)
}

func TestListFuzzyMacroName(t *testing.T) {
	testMacro := func(script, docstr string) {
		vm := loadAllExtensions(glisp.New())
		vm.AddFuzzyMacro(`^fuzzy-\d+$`, func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			num := strings.TrimPrefix(string(args[0].(glisp.SexpStr)), "fuzzy-")
			return glisp.SexpStr("number:" + num), nil
		})
		ret, err := vm.EvalString(script)
		ExpectSuccess(t, err)
		ExpectEqStr(t, docstr, ret)
	}
	testMacro(`(fuzzy-123)`, `number:123`)
	testMacro(`(fuzzy-456)`, `number:456`)
}

func TestDefineClassInGolang(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	extensions.NewRecordClassBuilder("test/Options").
		AddField("Name", "string").
		AddField("Age", "int").
		Build(vm)
	script := `(def p (->test/Options Name "hello" Age (+ 1 2))) (assert (= "hello" (:Name p))) (assert (= 3 (:Age p)))`
	_, err := vm.EvalString(script)
	ExpectSuccess(t, err)
}

func TestFloat64(t *testing.T) {
	sf := glisp.NewSexpFloat(113.965916)
	if f := sf.ToFloat64(); f != 113.965916 {
		t.Fatal("not equal", 113.965916, f)
	}
	bs, _ := sf.MarshalJSON()
	if string(bs) != "113.965916" {
		t.Fatal("not equal", string(bs))
	}
	sf, _ = glisp.NewSexpFloatStr("113.965916")
	if f := sf.ToFloat64(); f != 113.965916 {
		t.Fatal("not equal", 113.965916, f)
	}
	bs, _ = sf.MarshalJSON()
	if string(bs) != "113.965916" {
		t.Fatal("not equal", string(bs))
	}

	sf = glisp.NewSexpFloat(114.711036)
	if f := sf.ToFloat64(); f != 114.711036 {
		t.Fatal("not equal", 114.711036, f)
	}
	bs, _ = sf.MarshalJSON()
	if string(bs) != "114.711036" {
		t.Fatal("not equal", string(bs))
	}
	sf, _ = glisp.NewSexpFloatStr("114.711036")
	if f := sf.ToFloat64(); f != 114.711036 {
		t.Fatal("not equal", 114.711036, f)
	}
	if string(bs) != "114.711036" {
		t.Fatal("not equal", string(bs))
	}
}

func TestGoRecord(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	ret, err := vm.EvalString(`(defrecord GoR (Str string) (Int int) (Bool bool) (Bytes bytes "stream") (List list) (Array array) (Hash hash)) (->GoR Str "text")`)
	ExpectSuccess(t, err)
	r := extensions.ToGoRecord(ret.(extensions.SexpRecord))
	ExpectEqString(t, "text", r.GetStringField("Str"))
	r.SetStringField("Str", "text2")
	ExpectEqString(t, "text2", r.GetStringField("Str"))
	r.SetIntField("Int", 1)
	ExpectEqAny(t, int64(1), r.GetIntField("Int"))
	r.SetUintField("Int", 1)
	ExpectEqAny(t, uint64(1), r.GetUintField("Int"))
	r.SetBoolField("Bool", true)
	ExpectEqAny(t, true, r.GetBoolField("Bool"))
	r.SetBytesField("Bytes", []byte("abc"))
	ExpectEqAny(t, []byte("abc"), r.GetBytesField("Bytes"))
	ExpectEqAny(t, "stream", r.GetTag("Bytes"))

	r.SetHashField("Hash", map[string]interface{}{"key": "value"})
	ExpectEqAny(t, r.GetHashField("Hash"), map[string]interface{}{"key": "value"})
	r.SetHashField("Hash", map[string]interface{}{"key": 1024})

	r.SetListField("List", []interface{}{1, 2, 3})
	ExpectEqList(t, r.GetListField("List"), []interface{}{1, 2, 3})

	r.SetListField("Array", []interface{}{1, 2, 3})
	ExpectEqList(t, r.GetListField("Array"), []interface{}{1, 2, 3})
}

func TestOSCmd(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	buf := new(bytes.Buffer)
	vm.AddNamedFunction("os/exec!", extensions.ExecCommand(&extensions.CommandOptions{
		Stdout:        buf,
		Stderr:        buf,
		AssertSuccess: true,
	}))
	_, err := vm.EvalString(`(os/exec! "echo -n aaa")`)
	ExpectSuccess(t, err)
	ExpectEqString(t, buf.String(), "aaa")

	buf.Reset()
	_, err = vm.EvalString(`(os/exec! "lsl")`)
	ExpectError(t, err)
	ExpectContainsStr(t, glisp.SexpStr(buf.String()), "bash: lsl: command not found")
}

func TestOSCmd2(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	buf := extensions.NewBuffer()
	vm.BindGlobal("stdout", buf)
	vm.BindGlobal("stderr", buf)
	_, err := vm.EvalString(`(os/exec! {"cmd" "echo -n aaa" "stdout" stdout "stderr" stderr})`)
	ExpectSuccess(t, err)
	ExpectEqString(t, buf.Stringify(), "aaa")
}

func TestFileReader(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.SetFileReader(OnceFileReader(map[string]string{
		"file1.lisp": `(set! g (+ g 1))`,
	}))
	vm.BindGlobal("g", glisp.NewSexpInt(0))
	ret, err := vm.EvalString(`(include "file1.lisp") (include "file1.lisp") g`)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 1, ret)
}

func TestFileReader2(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.SetFileReader(OnceFileReader(map[string]string{
		"file1.lisp": `(set! g (+ g 1))`,
	}))
	vm.BindGlobal("g", glisp.NewSexpInt(0))
	ret, err := vm.EvalString(`(source-file "file1.lisp") (source-file "file1.lisp") g`)
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 1, ret)
}

type loadModuleOnce struct {
	files map[string]string
}

func OnceFileReader(files map[string]string) glisp.FileReader {
	return &loadModuleOnce{files: files}
}
func (f *loadModuleOnce) Open(file string) (io.ReadCloser, error) {
	if str, ok := f.files[file]; ok {
		delete(f.files, file)
		return str2Stream(str), nil
	}
	return str2Stream("nil"), nil
}

func str2Stream(s string) io.ReadCloser {
	buf := bytes.NewBufferString(s)
	return ioutil.NopCloser(buf)
}

func TestWriter(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewBuffer())
	ret, err := vm.EvalString("(type buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "*buffer*")

	ret, err = vm.EvalString("(sexp-str buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "*buffer*")

	_, err = vm.EvalString("(:close buf)")
	ExpectSuccess(t, err)

	ret, err = vm.EvalString("(:name buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "anonWriter")

	_, err = vm.EvalString("(:not-exist-method buf)")
	ExpectError(t, err, "not support :not-exist-method")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewBuffer())
	_, err = vm.EvalString("(:write buf 1 2 3)")
	ExpectError(t, err, ":write expect 1 argument(s) but got 3")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewBuffer())
	_, err = vm.EvalString("(:write buf 1)")
	ExpectError(t, err, "must write bytes/string to file but got int")
}

func TestReader(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewBuffer())
	ret, err := vm.EvalString("(type buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "*buffer*")

	ret, err = vm.EvalString("(sexp-str buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "*buffer*")

	vm.BindGlobal("buf", extensions.NewIO(1))
	ret, err = vm.EvalString("(sexp-str buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "io")
	ret, err = vm.EvalString("(type buf)")
	ExpectSuccess(t, err)
	ExpectContainsStr(t, ret, "io")
}

func TestIO(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewIO(glisp.NewSexpInt(1)))
	_, err := vm.EvalString("(:close buf)")
	ExpectError(t, err, "is not IO object")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewIO(glisp.NewSexpInt(1)))
	_, err = vm.EvalString("(:readx buf 1)")
	ExpectError(t, err, "not support :readx")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewIO(glisp.NewSexpInt(1)))
	_, err = vm.EvalString("(:write buf 1)")
	ExpectError(t, err, "must write bytes/string to file but got")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewIO(glisp.NewSexpInt(1)))
	_, err = vm.EvalString("(:write buf \"x\")")
	ExpectError(t, err, "is not IO object")

	vm = loadAllExtensions(glisp.New())
	vm.BindGlobal("buf", extensions.NewIO(rd(0)))
	_, err = vm.EvalString("(:write buf \"x\")")
	ExpectError(t, err, "io is not writable")

}

type rd int

func (rd) Read(p []byte) (n int, err error) { return }

type wt int

func (wt) Write(p []byte) (int, error) { return 0, nil }
