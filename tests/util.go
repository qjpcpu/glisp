package tests

import (
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
	vm.ImportEval()
	extensions.ImportCoreUtils(vm)
	extensions.ImportContainerUtils(vm)
	extensions.ImportJSON(vm)
	extensions.ImportMathUtils(vm)
	extensions.ImportBase64(vm)
	extensions.ImportChannels(vm)
	extensions.ImportCoroutines(vm)
	extensions.ImportRandom(vm)
	extensions.ImportRegex(vm)
	extensions.ImportTime(vm)
	extensions.ImportString(vm)
	extensions.ImportOS(vm)
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
		if !strings.Contains(err.Error(), key) {
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

func ExpectScriptSuccess(t *testing.T, script string) {
	env := newFullEnv()
	_, err := env.EvalString(script)
	ExpectSuccess(t, err)
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
