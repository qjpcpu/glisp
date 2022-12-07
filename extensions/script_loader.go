package extensions

import (
	"bytes"
	"embed"

	"github.com/qjpcpu/glisp"
)

//go:embed *.clj
var scripts embed.FS

func assertSuccess(err error) {
	if err != nil {
		panic(err)
	}
}
func mustLoadScript(env *glisp.Environment, name string) *glisp.Environment {
	bs, err := scripts.ReadFile(name + ".clj")
	assertSuccess(err)
	assertSuccess(env.SourceStream(bytes.NewBufferString(string(bs))))
	return env
}
