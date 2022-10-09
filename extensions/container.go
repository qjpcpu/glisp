package extensions

import (
	"bytes"

	_ "embed"

	"github.com/qjpcpu/glisp"
)

var (
	//go:embed container.lisp
	container_scripts string
)

func ImportContainerUtils(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	return env.SourceStream(bytes.NewBufferString(container_scripts))
}
