package tests

import (
	"bytes"
	"fmt"
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
	fn, err := vm.MakeScriptFunction(`(+ @arg0 @arg1)`, ``)
	ExpectSuccess(t, err)

	ret, err := vm.Apply(fn, []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	ExpectSuccess(t, err)
	ExpectEqInteger(t, 3, ret)

	fn, err = vm.MakeScriptFunction(`(str/start-with? @arg0 @arg1)`, ``)
	ExpectSuccess(t, err)

	ret, err = vm.Apply(fn, []glisp.Sexp{glisp.SexpStr("abc"), glisp.SexpStr("a")})
	ExpectSuccess(t, err)
	ExpectTrue(t, ret)
}

func TestMakeComplexScriptFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	script := `
(defn afn [a b] (+ a b))
(defmac amac [a b] ESCAPE(* ~a ~b))
;; (@arg0 + @arg1) * (@arg3 - @arg2)
(amac (afn @arg0 @arg1) ((fn [a] (- a @arg2)) @arg3))
`
	fn, err := vm.MakeScriptFunction(strings.ReplaceAll(script, "ESCAPE", "`"), "")
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
	fn, _ := vm.FindObject("typestr")
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
