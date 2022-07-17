package glisp

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

type Sexp interface {
	SexpString() string
}

type SexpSentinel int

const (
	SexpNull SexpSentinel = iota
	SexpEnd
	SexpMarker
)

func (sent SexpSentinel) SexpString() string {
	if sent == SexpNull {
		return "()"
	}
	if sent == SexpEnd {
		return "End"
	}
	if sent == SexpMarker {
		return "Marker"
	}

	return ""
}

type SexpPair struct {
	head Sexp
	tail Sexp
}

func Cons(a Sexp, b Sexp) SexpPair {
	return SexpPair{a, b}
}

func (pair SexpPair) Head() Sexp {
	return pair.head
}

func (pair SexpPair) Tail() Sexp {
	return pair.tail
}

func (pair SexpPair) SexpString() string {
	str := "("

	for {
		switch pair.tail.(type) {
		case SexpPair:
			str += pair.head.SexpString() + " "
			pair = pair.tail.(SexpPair)
			continue
		}
		break
	}

	str += pair.head.SexpString()

	if pair.tail == SexpNull {
		str += ")"
	} else {
		str += " . " + pair.tail.SexpString() + ")"
	}

	return str
}

type SexpArray []Sexp

type sexpHash struct {
	Map      map[int][]SexpPair
	KeyOrder []Sexp // must user pointers here, else hset! will fail to update.
	NumKeys  int
}

type SexpHash = *sexpHash

type SexpInt struct {
	v *big.Int
}

type SexpBool bool
type SexpFloat float64
type SexpChar rune
type SexpStr string

var SexpFloatSize = reflect.TypeOf(SexpFloat(0.0)).Bits()

func (arr SexpArray) SexpString() string {
	if len(arr) == 0 {
		return "[]"
	}

	str := "[" + arr[0].SexpString()
	for _, sexp := range arr[1:] {
		str += " " + sexp.SexpString()
	}
	str += "]"
	return str
}

func (hash SexpHash) SexpString() string {
	str := "{"
	for _, arr := range hash.Map {
		for _, pair := range arr {
			str += pair.head.SexpString() + " "
			str += pair.tail.SexpString() + " "
		}
	}
	if len(str) > 1 {
		return str[:len(str)-1] + "}"
	}
	return str + "}"
}

func (b SexpBool) SexpString() string {
	if b {
		return "true"
	}
	return "false"
}

func NewSexpInt(i int) SexpInt {
	return SexpInt{v: big.NewInt(int64(i))}
}

func NewSexpInt64(i int64) SexpInt {
	return SexpInt{v: new(big.Int).SetInt64(i)}
}

func NewSexpUint64(i uint64) SexpInt {
	return SexpInt{v: new(big.Int).SetUint64(i)}
}
func NewSexpIntStr(str string) (SexpInt, error) {
	v, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return SexpInt{}, fmt.Errorf(`%s not number`, str)
	}
	return SexpInt{v: v}, nil
}

func NewSexpIntStrWithBase(str string, base int) (SexpInt, error) {
	switch base {
	case 10:
	case 16:
	case 8:
	case 2:
	default:
		base = 10
	}
	if bigint, ok := big.NewInt(0).SetString(str, base); ok {
		return SexpInt{v: bigint}, nil
	}
	return SexpInt{}, fmt.Errorf(`can't parse %s to number`, str)
}

func (i SexpInt) SexpString() string {
	return i.v.String()
}

func (i SexpInt) BitNot() SexpInt {
	return SexpInt{v: new(big.Int).Not(i.v)}
}

func (i SexpInt) Xor(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Xor(i.v, j.v)}
}

func (i SexpInt) Add(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Add(i.v, j.v)}
}

func (i SexpInt) Sub(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Sub(i.v, j.v)}
}

func (i SexpInt) Mul(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Mul(i.v, j.v)}
}

func (i SexpInt) Div(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Div(i.v, j.v)}
}

func (i SexpInt) ShiftLeft(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Lsh(i.v, uint(j.v.Uint64()))}
}

func (i SexpInt) ShiftRight(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Rsh(i.v, uint(j.v.Uint64()))}
}

func (i SexpInt) Mod(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Mod(i.v, j.v)}
}

func (i SexpInt) And(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).And(i.v, j.v)}
}

func (i SexpInt) Or(j SexpInt) SexpInt {
	return SexpInt{v: new(big.Int).Or(i.v, j.v)}
}

func (i SexpInt) IsZero() bool {
	return i.v.Cmp(big.NewInt(0)) == 0
}

func (i SexpInt) IsInt64() bool {
	return i.v.IsInt64()
}

func (i SexpInt) IsUint64() bool {
	return i.v.IsUint64()
}

func (i SexpInt) ToInt64() int64 {
	return i.v.Int64()
}

func (i SexpInt) ToInt() int {
	return int(i.ToInt64())
}

func (i SexpInt) ToUint64() uint64 {
	return i.v.Uint64()
}

func (i SexpInt) ToFloat64() (float64, error) {
	return strconv.ParseFloat(i.v.String(), 64)
}

func (i SexpInt) Sign() int {
	return i.v.Sign()
}

func (f SexpFloat) SexpString() string {
	return strconv.FormatFloat(float64(f), 'g', 5, SexpFloatSize)
}

func (c SexpChar) SexpString() string {
	/* char is ' */
	if int32(c) == 39 {
		return `#'`
	}
	return "#" + strings.Trim(strconv.QuoteRune(rune(c)), "'")
}

func (s SexpStr) SexpString() string {
	return strconv.Quote(string(s))
}

type SexpSymbol struct {
	name   string
	number int
}

func (sym SexpSymbol) SexpString() string {
	return sym.name
}

func (sym SexpSymbol) Name() string {
	return sym.name
}

func (sym SexpSymbol) Number() int {
	return sym.number
}

type SexpFunction struct {
	name       string
	user       bool
	nargs      int
	varargs    bool
	fun        Function
	userfun    UserFunction
	closeScope *Stack
}

func (sf SexpFunction) SexpString() string {
	return "fn [" + sf.name + "]"
}

func IsTruthy(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpBool:
		return bool(e)
	case SexpInt:
		return e.v.Cmp(big.NewInt(0)) != 0
	case SexpChar:
		return e != 0
	case SexpSentinel:
		return e != SexpNull
	}
	return true
}

type SexpStackmark struct {
	sym SexpSymbol
}

func (mark SexpStackmark) SexpString() string {
	return "stackmark " + mark.sym.name
}
