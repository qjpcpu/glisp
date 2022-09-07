package glisp

import (
	"fmt"
	"math/big"
)

var Float64EqualityThreshold = 1e-10

type SexpFloat struct {
	v *big.Float
	/* well, sometimes we want SexpString show it's original string value */
	rawStr string
}

func NewSexpFloat(c float64) SexpFloat {
	return SexpFloat{v: big.NewFloat(c)}
}

func NewSexpFloatStr(str string) (SexpFloat, error) {
	f, ok := new(big.Float).SetString(str)
	if !ok {
		return SexpFloat{v: new(big.Float)}, fmt.Errorf("%s not float", str)
	}
	return SexpFloat{v: f, rawStr: str}, nil
}

func NewSexpFloatInt(i SexpInt) SexpFloat {
	return SexpFloat{v: new(big.Float).SetInt(i.v)}
}

func (f SexpFloat) SexpString() string {
	if f.rawStr != "" {
		return f.rawStr
	}
	return f.v.String()
}

func (f SexpFloat) ToString(prec int) string {
	if prec < 0 {
		return f.SexpString()
	}
	s := f.v.Text('f', prec)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != '0' {
			if s[i] == '.' {
				return s[:i]
			}
			return s[:i+1]
		}
	}
	return s
}

func (f SexpFloat) Div(f2 SexpFloat) SexpFloat {
	return SexpFloat{v: new(big.Float).Quo(f.v, f2.v)}
}

func (f SexpFloat) Sub(f2 SexpFloat) SexpFloat {
	return SexpFloat{v: new(big.Float).Sub(f.v, f2.v)}
}

func (f SexpFloat) Mul(f2 SexpFloat) SexpFloat {
	return SexpFloat{v: new(big.Float).Mul(f.v, f2.v)}
}

func (f SexpFloat) Add(f2 SexpFloat) SexpFloat {
	return SexpFloat{v: new(big.Float).Add(f.v, f2.v)}
}

func (f SexpFloat) Cmp(f2 SexpFloat) int {
	res := new(big.Float).Sub(f.v, f2.v)
	abs := new(big.Float).Abs(res)
	if abs.Cmp(new(big.Float).SetFloat64(Float64EqualityThreshold)) < 0 {
		return 0
	}
	return res.Sign()
}

func (f SexpFloat) ToFloat64() float64 {
	v, _ := f.v.Float64()
	return v
}

func (f SexpFloat) Round() SexpInt {
	leftInt, rightInt := f.intRange()
	l, _ := new(big.Float).Sub(f.v, new(big.Float).SetInt(leftInt)).Float64()
	r, _ := new(big.Float).Sub(new(big.Float).SetInt(rightInt), f.v).Float64()
	if l < r {
		return SexpInt{v: leftInt}
	}
	return SexpInt{v: rightInt}
}

func (f SexpFloat) intRange() (*big.Int, *big.Int) {
	if res := f.v.Cmp(big.NewFloat(0)); res >= 0 {
		leftInt := new(big.Int)
		f.v.Int(leftInt)
		rightInt := new(big.Int).Add(leftInt, big.NewInt(1))
		return leftInt, rightInt
	}
	rightInt := new(big.Int)
	f.v.Int(rightInt)
	leftInt := new(big.Int).Sub(rightInt, big.NewInt(1))
	return leftInt, rightInt
}

func (f SexpFloat) Ceil() SexpInt {
	_, rightInt := f.intRange()
	return SexpInt{v: rightInt}
}

func (f SexpFloat) Floor() SexpInt {
	leftInt, _ := f.intRange()
	return SexpInt{v: leftInt}
}
