package extensions

import (
	"bytes"
	"fmt"
	"io"

	"github.com/qjpcpu/glisp"
)

type Namer interface {
	Name() string
}

type SexpIO struct {
	err error
	io  interface{}
}

func NewIO(obj interface{}) *SexpIO {
	switch obj.(type) {
	case io.Reader, io.Writer:
		return &SexpIO{io: obj}
	}
	return &SexpIO{io: obj, err: fmt.Errorf("%#T is not IO object", obj)}
}

func (w *SexpIO) SexpString() string {
	if t, ok := w.io.(glisp.Sexp); ok {
		return t.SexpString()
	}
	return "io"
}

func (w *SexpIO) TypeName() string {
	if t, ok := w.io.(glisp.ITypeName); ok {
		return t.TypeName()
	}
	return "io"
}

func (w *SexpIO) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if val, ok := w.io.(io.Writer); ok {
		return val.Write(p)
	}
	return 0, fmt.Errorf("%s is not writable", w.TypeName())
}

func (w *SexpIO) Read(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if val, ok := w.io.(io.Reader); ok {
		return val.Read(p)
	}
	return 0, fmt.Errorf("%s is not readable", w.TypeName())
}

func (w *SexpIO) Close() error {
	if w.err != nil {
		return w.err
	}
	if val, ok := w.io.(io.Closer); ok {
		return val.Close()
	}
	return fmt.Errorf("%s is not closable", w.TypeName())
}

func (w *SexpIO) Stringify() string {
	if w.err != nil {
		return ""
	}
	if val, ok := w.io.(glisp.Stringer); ok {
		return val.Stringify()
	}
	return w.SexpString()
}

func (w *SexpIO) Explain(env *glisp.Environment, sym string, args []glisp.Sexp) (glisp.Sexp, error) {
	switch sym {
	case "close":
		return glisp.SexpNull, w.Close()
	case "println", "printf", "print":
		return GetPrintFunction(w)(sym)(env, args)
	case "name":
		if n, ok := w.io.(Namer); ok {
			return glisp.SexpStr(n.Name()), nil
		}
		return glisp.SexpStr("anonWriter"), nil
	case "write":
		if len(args) != 1 {
			return glisp.WrongNumberArguments(":write", len(args), 1)
		}
		var n int
		var err error
		if glisp.IsBytes(args[0]) {
			n, err = w.Write(args[0].(glisp.SexpBytes).Bytes())
		} else if glisp.IsString(args[0]) {
			n, err = w.Write([]byte(string(args[0].(glisp.SexpStr))))
		} else {
			return glisp.SexpNull, fmt.Errorf("must write bytes/string to file but got %v", glisp.InspectType(args[0]))
		}
		return glisp.NewSexpInt(n), err
	default:
		return glisp.SexpNull, fmt.Errorf("not support :%s", sym)
	}
}

func NewBuffer() *SexpIO { return NewIO(&sexpBuffer{buffer: new(bytes.Buffer)}) }

type sexpBuffer struct {
	buffer *bytes.Buffer
}

func (b *sexpBuffer) Read(p []byte) (int, error) {
	return b.buffer.Read(p)
}

func (b *sexpBuffer) Write(p []byte) (int, error) {
	return b.buffer.Write(p)
}

func (b *sexpBuffer) SexpString() string {
	return "*buffer*"
}

func (b *sexpBuffer) TypeName() string {
	return "*buffer*"
}

func (b *sexpBuffer) Close() error {
	return nil
}

func (b *sexpBuffer) Stringify() string {
	return b.buffer.String()
}

func newBuffer(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		return NewBuffer(), nil
	}
}
