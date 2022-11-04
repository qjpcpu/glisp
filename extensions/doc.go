package extensions

import (
	_ "embed"

	"github.com/qjpcpu/glisp"
)

var (
	//go:embed documentation.txt
	documentations string
	docMap         = make(map[string]string)
)

func getDoc(funcName string) glisp.FuntionOption {
	return glisp.WithDoc(docMap[funcName])
}

func init() {
	docMap = glisp.ParseDoc(documentations)
	documentations = ``
}

type AutoAddFunctionWithDoc struct {
	*glisp.Environment
}

func autoAddDoc(env *glisp.Environment) *AutoAddFunctionWithDoc {
	return &AutoAddFunctionWithDoc{Environment: env}
}

func (env *AutoAddFunctionWithDoc) AddNamedFunction(name string, function glisp.NamedUserFunction, opts ...glisp.FuntionOption) {
	env.Environment.AddNamedFunction(name, function, getDoc(name))
}

func (env *AutoAddFunctionWithDoc) OverrideFunction(name string, f glisp.OverrideFunction) error {
	return env.Environment.OverrideFunction(name, f, getDoc(name))
}

func (env *AutoAddFunctionWithDoc) AddNamedMacro(name string, function glisp.NamedUserFunction, opts ...glisp.FuntionOption) {
	env.Environment.AddMacro(name, function(name), getDoc(name))
}

func (env *AutoAddFunctionWithDoc) AddFuzzyMacro(name string, function glisp.NamedUserFunction, opts ...glisp.FuntionOption) {
	env.Environment.AddFuzzyMacro(name, function(name), getDoc(name))
}
