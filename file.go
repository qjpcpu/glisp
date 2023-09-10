package glisp

import (
	"io"
	"os"
)

type FileReader interface {
	Open(file string) (io.ReadCloser, error)
}

func DefaultFileReader() FileReader { return osFileReader{} }

type osFileReader struct{}

func (osFileReader) Open(file string) (io.ReadCloser, error) {
	return os.Open(file)
}
