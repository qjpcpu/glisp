package glisp

import (
	"fmt"
	"math/big"
)

var Float64EqualityThreshold = 1e-10

type SexpFloat struct{ v *big.Float }

func NewSexpFloat(c float64) SexpFloat {
	return SexpFloat{v: big.NewFloat(c)}
}

func NewSexpFloatStr(str string) (SexpFloat, error) {
	f, ok := new(big.Float).SetString(str)
	if !ok {
		return SexpFloat{v: new(big.Float)}, fmt.Errorf("%s not float", str)
	}
	return SexpFloat{v: f}, nil
}

func NewSexpFloatInt(i SexpInt) SexpFloat {
	return SexpFloat{v: new(big.Float).SetInt(i.v)}
}

func (f SexpFloat) SexpString() string {
	return f.v.String()
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
	integer := new(big.Int)
	f.v.Int(integer)
	integer1 := new(big.Int).Add(integer, big.NewInt(1))
	l, _ := new(big.Float).Sub(f.v, new(big.Float).SetInt(integer)).Float64()
	r, _ := new(big.Float).Sub(new(big.Float).SetInt(integer1), f.v).Float64()
	if l < 0 {
		l = -l
	}
	if r < 0 {
		r = -r
	}
	if l < r {
		return SexpInt{v: integer}
	}
	return SexpInt{v: integer1}
}
