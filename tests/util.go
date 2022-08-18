package tests

import (
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

func ShouldEqInteger(t *testing.T, expect int64, expr glisp.Sexp, err error) {
	if err != nil {
		t.Fatalf("should eq %v but got err %v", expect, err)
	}
	if !glisp.IsInt(expr) {
		t.Fatalf("should get integer but got %#T", expr)
	}
	if v := expr.(glisp.SexpInt).ToInt64(); v != expect {
		t.Fatalf("should get %v but got %v", expect, v)
	}
}

func ShouldNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("should no error but got err %v", err)
	}
}
