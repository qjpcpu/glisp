package extensions

import (
	"errors"
	"fmt"

	"github.com/qjpcpu/glisp"
)

func ImportMathUtils(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
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
	env.AddNamedFunction("round", GetRoundFloat)
	env.AddNamedFunction("ceil", GetCeilFloat)
	env.AddNamedFunction("floor", GetFloorFloat)
	env.AddNamedFunction("0b", FormatInt(2))
	env.AddNamedFunction("0o", FormatInt(8))
	env.AddNamedFunction("0x", FormatInt(16))
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
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() < 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2, glisp.Many)
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

		accum := args.Get(0)
		var err error

		args.SliceStart(1).Foreach(func(expr glisp.Sexp) bool {
			accum, err = IntegerDo(op, accum, expr)
			return err == nil
		})
		if err != nil {
			return glisp.SexpNull, err
		}
		return accum, nil
	}
}

func ComplementFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 1)
		}

		switch t := args.Get(0).(type) {
		case glisp.SexpInt:
			return t.BitNot(), nil
		case glisp.SexpChar:
			return ^t, nil
		}

		return glisp.SexpNull, errors.New("Argument to bit-not should be integer")
	}
}

func GetBinaryIntFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2)
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

		return IntegerDo(op, args.Get(0), args.Get(1))
	}
}

func GetLogicalShiftFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2)
		}

		if !glisp.IsInt(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be integer but got %v", name, glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsInt(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf("second argument of %s must be integer but got %v", name, glisp.InspectType(args.Get(1)))
		}
		shift := uint(args.Get(1).(glisp.SexpInt).ToUint64())
		snum := args.Get(0).(glisp.SexpInt)
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

func GetRoundFloat(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		switch val := args.Get(0).(type) {
		case glisp.SexpFloat:
			return val.Round(), nil
		case glisp.SexpInt:
			return val, nil
		}
		return glisp.SexpNull, fmt.Errorf(`%s argument should be float but got %v`, name, glisp.InspectType(args.Get(0)))
	}
}

func GetCeilFloat(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		switch val := args.Get(0).(type) {
		case glisp.SexpFloat:
			return val.Ceil(), nil
		case glisp.SexpInt:
			return val, nil
		}
		return glisp.SexpNull, fmt.Errorf(`%s argument should be float but got %v`, name, glisp.InspectType(args.Get(0)))
	}
}

func GetFloorFloat(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		switch val := args.Get(0).(type) {
		case glisp.SexpFloat:
			return val.Floor(), nil
		case glisp.SexpInt:
			return val, nil
		}
		return glisp.SexpNull, fmt.Errorf(`%s argument should be float but got %v`, name, glisp.InspectType(args.Get(0)))
	}
}

func FormatInt(bit int) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 1 {
				return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
			}
			if !glisp.IsInt(args.Get(0)) {
				return glisp.SexpNull, fmt.Errorf("first argument of %s must be int but got %s", name, glisp.InspectType(args.Get(0)))
			}
			num := args.Get(0).(glisp.SexpInt)
			switch bit {
			case 2:
				return glisp.SexpStr("0b" + num.Format("%b")), nil
			case 8:
				return glisp.SexpStr("0o" + num.Format("%o")), nil
			case 16:
				return glisp.SexpStr("0x" + num.Format("%x")), nil
			}
			return glisp.SexpStr(num.SexpString()), nil
		}
	}
}
