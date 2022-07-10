package tests

import (
	"bytes"
	"errors"
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

func TestScope(t *testing.T) {
	vm := glisp.NewGlisp()
	key := "aaa"
	if err := vm.BindObject(key, glisp.NewSexpInt(1)); err != nil {
		t.Fatal("bind object fail", err)
	}
	obj, ok := vm.FindObject(key)
	if !ok {
		t.Fatal("locate symbol fail")
	}
	if !glisp.IsInt(obj) || obj.(glisp.SexpInt).ToInt() != 1 {
		t.Fatalf("bad symbol,should be int %v", 1)
	}
	/* add scope */
	vm.AddScope()
	if err := vm.BindObject(key, glisp.NewSexpInt(2)); err != nil {
		t.Fatal("bind object fail", err)
	}
	obj, ok = vm.FindObject(key)
	if !ok {
		t.Fatal("locate symbol fail")
	}
	if !glisp.IsInt(obj) || obj.(glisp.SexpInt).ToInt() != 2 {
		t.Fatalf("bad symbol,should be int %v", 2)
	}
	/* pop scope */
	vm.PopScope()
	obj, ok = vm.FindObject(key)
	if !ok {
		t.Fatal("locate symbol fail")
	}
	if !glisp.IsInt(obj) || obj.(glisp.SexpInt).ToInt() != 1 {
		t.Fatalf("bad symbol,should be int %v", 1)
	}
}

func TestLoadAllFunction(t *testing.T) {
	vm := loadAllExtensions(glisp.NewGlisp())
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

func testFile(t *testing.T, file string) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("read file %s fail %v", file, err)
	}
	vm := registerTestingFunc(loadAllExtensions(glisp.NewGlisp()))
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

func loadAllExtensions(vm *glisp.Glisp) *glisp.Glisp {
	vm.ImportEval()
	extensions.ImportJSON(vm)
	extensions.ImportChannels(vm)
	extensions.ImportCoreUtils(vm)
	extensions.ImportCoroutines(vm)
	extensions.ImportRandom(vm)
	extensions.ImportRegex(vm)
	extensions.ImportTime(vm)
	extensions.ImportString(vm)
	return vm
}

func registerTestingFunc(vm *glisp.Glisp) *glisp.Glisp {
	vm.AddFunctionByConstructor("test/read-file", testReadFile)
	return vm
}

func testReadFile(name string) glisp.GlispUserFunction {
	return func(env *glisp.Glisp, args []glisp.Sexp) (glisp.Sexp, error) {
		bytes, _ := ioutil.ReadFile(string(args[0].(glisp.SexpStr)))
		return glisp.SexpStr(string(bytes)), nil
	}
}

func TestContextFunction(t *testing.T) {
	vm := glisp.NewGlisp()
	var value int
	vm.AddFunction("test/echo", func(env *glisp.Glisp, args []glisp.Sexp) (glisp.Sexp, error) {
		val, ok := env.FindObject("ctx")
		if !ok {
			return glisp.SexpNull, errors.New("context lost")
		}
		if !glisp.IsInt(val) {
			return glisp.SexpNull, errors.New("context lost")
		}
		if val.(glisp.SexpInt).ToInt() != value {
			return glisp.SexpNull, errors.New("context lost")
		}
		return val, nil
	})
	callEcho := func() (glisp.Sexp, error) {
		fn, ok := vm.FindObject("test/echo")
		if !ok {
			return glisp.SexpNull, errors.New("test/echo function not found")
		}
		if !glisp.IsFunction(fn) {
			return glisp.SexpNull, errors.New("test/echo function not found")
		}
		return vm.Apply(fn.(glisp.SexpFunction), nil)
	}
	if _, err := callEcho(); err == nil {
		t.Fatal("should not success")
	}

	vm.AddScope()
	value = 1
	vm.BindObject("ctx", glisp.NewSexpInt(value))
	if expr, err := callEcho(); err != nil {
		t.Fatalf("call echo fail %v", err)
	} else if expr.(glisp.SexpInt).ToInt() != value {
		t.Fatalf("find bad context value %v != %v", expr.(glisp.SexpInt).ToInt(), value)
	}
	vm.PopScope()

	vm.AddScope()
	value = 2
	vm.BindObject("ctx", glisp.NewSexpInt(value))
	if expr, err := callEcho(); err != nil {
		t.Fatalf("call echo fail %v", err)
	} else if expr.(glisp.SexpInt).ToInt() != value {
		t.Fatalf("find bad context value %v != %v", expr.(glisp.SexpInt).ToInt(), value)
	}
	vm.PopScope()

	if _, err := callEcho(); err == nil {
		t.Fatal("should not success")
	}
}
