package tests

import (
	"bytes"
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

func TestPrint(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	var buf bytes.Buffer
	vm.AddFunctionByConstructor("print", extensions.GetPrintFunction(&buf))
	vm.LoadString(`(print "hello")`)
	_, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}
	if ret := buf.String(); ret != "hello" {
		t.Fatalf("(%s)!=(%s)", ret, "hello")
	}
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
}

func testPrintf(t *testing.T, script string, expect string) {
	vm := loadAllExtensions(glisp.New())
	var buf bytes.Buffer
	vm.AddFunctionByConstructor("printf", extensions.GetPrintFunction(&buf))
	vm.LoadString(script)
	_, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); out != expect {
		t.Fatalf("[test printf] expect %s but got %s", expect, out)
	}
}

type testSexp struct{}

func (*testSexp) SexpString() string { return "xxxx" }

type testSexp2 struct{}

func (testSexp2) SexpString() string { return "yyyy" }

func TestCustomType(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	fn, _ := vm.FindObject("typestr")
	ret, _ := vm.Apply(fn.(glisp.SexpFunction), []glisp.Sexp{&testSexp{}})
	if string(ret.(glisp.SexpStr)) != "*tests.testSexp" {
		t.Fatal("bad type", ret.SexpString())
	}
	ret, _ = vm.Apply(fn.(glisp.SexpFunction), []glisp.Sexp{testSexp2{}})
	if string(ret.(glisp.SexpStr)) != "tests.testSexp2" {
		t.Fatal("bad type", ret.SexpString())
	}
}

func testFile(t *testing.T, file string) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("read file %s fail %v", file, err)
	}
	vm := registerTestingFunc(loadAllExtensions(glisp.New()))
	err = vm.LoadString(string(bytes))
	if err != nil {
		t.Fatalf("parse file %s fail %v", file, err)
	}
	_, err = vm.Run()
	if err != nil {
		t.Fatalf("run file %s fail %v", file, err)
	}
	t.Logf("TEST %s OK", file)
}

func listScripts(t *testing.T) []string {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatal("load scripts fail ", err)
	}

	var scripts []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".lisp") {
			scripts = append(scripts, filepath.Join(testDir, file.Name()))
		}
	}
	return scripts
}

func loadAllExtensions(vm *glisp.Environment) *glisp.Environment {
	vm.ImportEval()
	extensions.ImportJSON(vm)
	extensions.ImportMathUtils(vm)
	extensions.ImportBase64(vm)
	extensions.ImportChannels(vm)
	extensions.ImportCoreUtils(vm)
	extensions.ImportCoroutines(vm)
	extensions.ImportRandom(vm)
	extensions.ImportRegex(vm)
	extensions.ImportTime(vm)
	extensions.ImportString(vm)
	return vm
}

func registerTestingFunc(vm *glisp.Environment) *glisp.Environment {
	vm.AddFunctionByConstructor("test/read-file", testReadFile)
	return vm
}

func testReadFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		bytes, _ := ioutil.ReadFile(string(args[0].(glisp.SexpStr)))
		return glisp.SexpStr(string(bytes)), nil
	}
}

func TestSandBox(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.PushGlobalScope()
	if _, err := vm.EvalString(`(defn sb [a b ] (+ a b))`); err != nil {
		t.Fatal(err)
	}
	ret, err := vm.ApplyByName("sb", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	if err != nil {
		t.Fatal(err)
	}
	if !glisp.IsInt(ret) {
		t.Fatal(ret.SexpString() + " is not =3")
	}
	vm.PopGlobalScope()
	_, err = vm.ApplyByName("sb", []glisp.Sexp{glisp.NewSexpInt(1), glisp.NewSexpInt(2)})
	if err == nil {
		t.Fatal("should not found function")
	}
}
