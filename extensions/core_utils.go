package extensions

import (
	"bytes"

	_ "embed"
	"github.com/qjpcpu/glisp"
)

var (
	//go:embed alias.lisp
	alias_macro string
)

func ImportCoreUtils(env *glisp.Glisp) error {
	return env.SourceStream(bytes.NewBufferString(alias_macro))
}
