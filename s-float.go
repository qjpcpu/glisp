package glisp

import (
	"reflect"
	"strconv"
)

type SexpFloat float64

var SexpFloatSize = reflect.TypeOf(SexpFloat(0.0)).Bits()

func (f SexpFloat) SexpString() string {
	return strconv.FormatFloat(float64(f), 'g', 5, SexpFloatSize)
}
