package glisp

import (
	"fmt"
	"math/big"
	"strconv"
)

type SexpInt struct {
	v *big.Int
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
