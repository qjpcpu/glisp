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

func ImportContainerUtils(env *glisp.Environment) error {
	return env.SourceStream(bytes.NewBufferString(container_scripts))
}
