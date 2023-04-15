package extensions

import (
	"io"
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
