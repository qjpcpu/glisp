package tests

import (
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
	extensions.ImportIO(vm)
	return vm
}
