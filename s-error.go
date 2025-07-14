package glisp

import (
	"errors"
	"fmt"
	"strconv"
)

type SexpError struct {
	Err error
}

func NewErrorWith(e error) SexpError {
	return SexpError{Err: e}
}

func NewError(msg string) SexpError {
	return SexpError{Err: errors.New(msg)}
}

func (s SexpError) SexpString() string {
	return fmt.Sprintf("(error %s)", strconv.Quote(s.Err.Error()))
}

func (s SexpError) TypeName() string {
	return "error"
}

func (s SexpError) Stringify() string {
	return s.Err.Error()
}
