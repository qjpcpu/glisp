package tests

import (
	"io/ioutil"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

func newFullEnv() *glisp.Environment {
	return loadAllExtensions(glisp.New())
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
