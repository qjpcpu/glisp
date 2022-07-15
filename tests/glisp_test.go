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

func TestOverwriteBuiltinFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.New())
	vm.AddFunction("+", func(e *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		a := args[0].(glisp.SexpInt)
		b := args[1].(glisp.SexpInt)
		return a.Add(b).Add(glisp.NewSexpInt(100)), nil
	})
	if err := vm.LoadString(`(+ 2 3)`); err != nil {
		t.Fatal(err)
	}
	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}
	if !glisp.IsInt(ret) {
		t.Fatalf("result is not int %s", ret.SexpString())
	}
	if ret.(glisp.SexpInt).ToInt() != 105 {
		t.Fatalf("result should be 105 but got %v", ret.SexpString())
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
