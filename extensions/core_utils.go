package extensions

import (
	"bytes"

	_ "embed"
	"github.com/qjpcpu/glisp"
)

var (
	//go:embed core_utils.lisp
	core_scripts string
)

func ImportCoreUtils(env *glisp.Environment) error {
	return env.SourceStream(bytes.NewBufferString(core_scripts))
}
