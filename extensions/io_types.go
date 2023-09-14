package extensions

import (
	"fmt"
	"io"

	"github.com/qjpcpu/glisp"
)

type SexpWriter struct {
	w io.Writer
}

func NewWriter(w io.Writer) *SexpWriter {
	return &SexpWriter{w: w}
}

func (w *SexpWriter) SexpString() string {
	return "io.Writer"
}

func (w *SexpWriter) TypeName() string {
	return "io.Writer"
}

func (w *SexpWriter) Explain(env *glisp.Environment, sym string, args []glisp.Sexp) (glisp.Sexp, error) {
	switch sym {
	case "close":
		if c, ok := w.w.(io.Closer); ok {
			return glisp.SexpNull, c.Close()
		}
		return glisp.SexpNull, nil
	case "println", "printf", "print":
		return GetPrintFunction(w.w)(sym)(env, args)
	case "write":
		if len(args) != 1 {
			return glisp.WrongNumberArguments(":write", len(args), 1)
		}
		var n int
		var err error
		if glisp.IsBytes(args[0]) {
			n, err = w.w.Write(args[0].(glisp.SexpBytes).Bytes())
		} else if glisp.IsString(args[0]) {
			n, err = w.w.Write([]byte(string(args[0].(glisp.SexpStr))))
		} else {
			return glisp.SexpNull, fmt.Errorf("must write bytes to file but got %v", glisp.InspectType(args[0]))
		}
		return glisp.NewSexpInt(n), err
	default:
		return glisp.SexpNull, fmt.Errorf("no support :%s", sym)
	}
}

type SexpReader struct {
	r io.Reader
}

func NewReader(r io.Reader) *SexpReader {
	return &SexpReader{r: r}
}

func (r *SexpReader) SexpString() string {
	return "io.Reader"
}

func (r *SexpReader) TypeName() string {
	return "io.Reader"
}
