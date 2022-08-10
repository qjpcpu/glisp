package extensions

import (
	"errors"
	"fmt"

	"github.com/qjpcpu/glisp"
)

func ImportMathUtils(env *glisp.Environment) error {
	env.AddNamedFunction("sla", GetBinaryIntFunction)
	env.AddNamedFunction("sra", GetBinaryIntFunction)
	env.AddNamedFunction("bit-and", GetBitwiseFunction)
	env.AddNamedFunction("bit-or", GetBitwiseFunction)
	env.AddNamedFunction("bit-xor", GetBitwiseFunction)
	env.AddNamedFunction("bit-not", ComplementFunction)
	env.AddNamedFunction("sll8", GetLogicalShiftFunction)
	env.AddNamedFunction("sll16", GetLogicalShiftFunction)
	env.AddNamedFunction("sll32", GetLogicalShiftFunction)
	env.AddNamedFunction("sll64", GetLogicalShiftFunction)
	env.AddNamedFunction("srl8", GetLogicalShiftFunction)
	env.AddNamedFunction("srl16", GetLogicalShiftFunction)
	env.AddNamedFunction("srl32", GetLogicalShiftFunction)
	env.AddNamedFunction("srl64", GetLogicalShiftFunction)
	return nil
}

type IntegerOp int

const (
	ShiftLeftArith IntegerOp = iota
	ShiftRightArith
	ShiftLeftLogical
	ShiftRightLogical
	Modulo
	BitAnd
	BitOr
	BitXor
)

func IntegerDo(op IntegerOp, a, b glisp.Sexp) (glisp.Sexp, error) {
	var ia glisp.SexpInt
	var ib glisp.SexpInt

	switch i := a.(type) {
	case glisp.SexpInt:
		ia = i
	case glisp.SexpChar:
		ia = glisp.NewSexpInt(int(i))
	default:
		return glisp.SexpNull, glisp.WrongType
	}

	switch i := b.(type) {
	case glisp.SexpInt:
		ib = i
	case glisp.SexpChar:
		ib = glisp.NewSexpInt(int(i))
	default:
		return glisp.SexpNull, glisp.WrongType
	}

	switch op {
	case ShiftLeftArith:
		return ia.ShiftLeft(ib), nil
	case ShiftRightArith:
		return ia.ShiftRight(ib), nil
	case Modulo:
		if ib.IsZero() {
			return glisp.SexpNull, errors.New(`division by zero`)
		}
		return ia.Mod(ib), nil
	case BitAnd:
		return ia.And(ib), nil
	case BitOr:
		return ia.Or(ib), nil
	case BitXor:
		return ia.Xor(ib), nil
	}
	return glisp.SexpNull, errors.New("unrecognized shift operation")
}

func GetBitwiseFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}

		var op IntegerOp
		switch name {
		case "bit-and":
			op = BitAnd
		case "bit-or":
			op = BitOr
		case "bit-xor":
			op = BitXor
		}

		accum := args[0]
		var err error

		for _, expr := range args[1:] {
			accum, err = IntegerDo(op, accum, expr)
			if err != nil {
				return glisp.SexpNull, err
			}
		}
		return accum, nil
	}
}

func ComplementFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case glisp.SexpInt:
			return t.BitNot(), nil
		case glisp.SexpChar:
			return ^t, nil
		}

		return glisp.SexpNull, errors.New("Argument to bit-not should be integer")
	}
}

func GetBinaryIntFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}

		var op IntegerOp
		switch name {
		case "sla":
			op = ShiftLeftArith
		case "sra":
			op = ShiftRightArith
		case "mod":
			op = Modulo
		}

		return IntegerDo(op, args[0], args[1])
	}
}

func GetLogicalShiftFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}

		if !glisp.IsInt(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be integer", name)
		}
		if !glisp.IsInt(args[1]) {
			return glisp.SexpNull, fmt.Errorf("second argument of %s must be integer", name)
		}
		shift := uint(args[1].(glisp.SexpInt).ToUint64())
		snum := args[0].(glisp.SexpInt)
		switch name {
		case "sll8":
			if snum.Sign() >= 0 {
				val := uint8(snum.ToUint64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint8(snum.ToInt64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "sll16":
			if snum.Sign() >= 0 {
				val := uint16(snum.ToUint64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint16(snum.ToInt64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "sll32":
			if snum.Sign() >= 0 {
				val := uint32(snum.ToUint64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint32(snum.ToInt64()) << shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "sll64":
			if snum.Sign() >= 0 {
				val := snum.ToUint64() << shift
				return glisp.NewSexpUint64(val), nil
			} else {
				val := uint64(snum.ToInt64()) << shift
				return glisp.NewSexpUint64(val), nil
			}
		case "srl8":
			if snum.Sign() >= 0 {
				val := uint8(snum.ToUint64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint8(snum.ToInt64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "srl16":
			if snum.Sign() >= 0 {
				val := uint16(snum.ToUint64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint16(snum.ToInt64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "srl32":
			if snum.Sign() >= 0 {
				val := uint32(snum.ToUint64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			} else {
				val := uint32(snum.ToInt64()) >> shift
				return glisp.NewSexpUint64(uint64(val)), nil
			}
		case "srl64":
			if snum.Sign() >= 0 {
				val := snum.ToUint64() >> shift
				return glisp.NewSexpUint64(val), nil
			} else {
				val := uint64(snum.ToInt64()) >> shift
				return glisp.NewSexpUint64(val), nil
			}
		}

		return glisp.SexpNull, errors.New("unrecognized shift operation")
	}
}
