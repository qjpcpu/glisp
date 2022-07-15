package glisp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Function []Instruction
type UserFunction func(*Environment, []Sexp) (Sexp, error)
type UserFunctionConstructor func(string) UserFunction

func GetCompareFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		res, err := Compare(args[0], args[1])
		if err != nil {
			return SexpNull, err
		}

		cond := false
		switch name {
		case "<":
			cond = res < 0
		case ">":
			cond = res > 0
		case "<=":
			cond = res <= 0
		case ">=":
			cond = res >= 0
		case "=":
			cond = res == 0
		case "not=":
			cond = res != 0
		}

		return SexpBool(cond), nil
	}
}

func GetBinaryIntFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		var op IntegerOp
		switch name {
		case "sll":
			op = ShiftLeft
		case "sra":
			op = ShiftRightArith
		case "srl":
			op = ShiftRightLog
		case "mod":
			op = Modulo
		}

		return IntegerDo(op, args[0], args[1])
	}
}

func GetBitwiseFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
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
				return SexpNull, err
			}
		}
		return accum, nil
	}
}

func ComplementFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case SexpInt:
			return t.BitNot(), nil
		case SexpChar:
			return ^t, nil
		}

		return SexpNull, errors.New("Argument to bit-not should be integer")
	}
}

func GetNumericFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		var err error
		accum := args[0]
		var op NumericOp
		switch name {
		case "+":
			op = Add
		case "-":
			op = Sub
		case "*":
			op = Mult
		case "/":
			op = Div
		}

		for _, expr := range args[1:] {
			accum, err = NumericDo(op, accum, expr)
			if err != nil {
				return SexpNull, err
			}
		}
		return accum, nil
	}
}

func ConsFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		return Cons(args[0], args[1]), nil
	}
}

func FirstFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch expr := args[0].(type) {
		case SexpPair:
			return expr.head, nil
		case SexpArray:
			return expr[0], nil
		}

		return SexpNull, WrongType
	}
}

func RestFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch expr := args[0].(type) {
		case SexpPair:
			return expr.tail, nil
		case SexpArray:
			if len(expr) == 0 {
				return expr, nil
			}
			return expr[1:], nil
		case SexpSentinel:
			if expr == SexpNull {
				return SexpNull, nil
			}
		}

		return SexpNull, WrongType
	}
}

func GetArrayAccessFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 2 {
			return WrongNumberArguments(name, len(args), 2, 3)
		}

		var arr SexpArray
		switch t := args[0].(type) {
		case SexpArray:
			arr = t
		default:
			return SexpNull, errors.New("First argument of aget must be array")
		}

		var i int
		switch t := args[1].(type) {
		case SexpInt:
			i = int(t.ToInt64())
		case SexpChar:
			i = int(t)
		default:
			return SexpNull, errors.New("Second argument of aget must be integer")
		}

		if i < 0 || i >= len(arr) {
			return SexpNull, errors.New("Array index out of bounds")
		}

		if name == "aget" {
			return arr[i], nil
		}

		if len(args) != 3 {
			return WrongNumberArguments(name, len(args), 3)
		}
		arr[i] = args[2]

		return SexpNull, nil
	}
}

func SgetFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		var str SexpStr
		switch t := args[0].(type) {
		case SexpStr:
			str = t
		default:
			return SexpNull, errors.New("First argument of sget must be string")
		}

		var i int
		switch t := args[1].(type) {
		case SexpInt:
			i = int(t.ToInt64())
		case SexpChar:
			i = int(t)
		default:
			return SexpNull, errors.New("Second argument of sget must be integer")
		}

		return SexpChar(str[i]), nil
	}
}

func GetExistFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}
		switch expr := args[0].(type) {
		case SexpHash:
			if _, err := expr.HashGet(args[1]); err != nil {
				if strings.Contains(err.Error(), "not found") {
					return SexpBool(false), nil
				}
				return SexpNull, err
			} else {
				return SexpBool(true), nil
			}
		case SexpArray:
			for _, item := range expr {
				if eq, err := Compare(item, args[1]); err != nil {
					return SexpNull, err
				} else if eq == 0 {
					return SexpBool(true), nil
				}
			}
			return SexpBool(false), nil
		case SexpPair:
			if IsList(expr) {
				ex, err := existInList(expr, args[1])
				return SexpBool(ex), err
			}
		case SexpSentinel:
			if expr == SexpNull {
				return SexpBool(false), nil
			}
		}
		return SexpNull, fmt.Errorf(`%s only support hash/array/list`, name)
	}
}

func GetHashAccessFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 2 || len(args) > 3 {
			return WrongNumberArguments(name, len(args), 2, 3)
		}

		var hash SexpHash
		switch e := args[0].(type) {
		case SexpHash:
			hash = e
		default:
			return SexpNull, errors.New("first argument of hget must be hash")
		}

		switch name {
		case "hget":
			if len(args) == 3 {
				return hash.HashGetDefault(args[1], args[2])
			}
			return hash.HashGet(args[1])
		case "hset!":
			err := hash.HashSet(args[1], args[2])
			return SexpNull, err
		case "hdel!":
			if len(args) != 2 {
				return WrongNumberArguments(name, len(args), 2)
			}
			err := hash.HashDelete(args[1])
			return SexpNull, err
		}

		return SexpNull, nil
	}
}

func SliceFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 3 {
			return WrongNumberArguments(name, len(args), 3)
		}

		var start int
		var end int
		switch t := args[1].(type) {
		case SexpInt:
			start = int(t.ToInt64())
		case SexpChar:
			start = int(t)
		default:
			return SexpNull, errors.New("Second argument of slice must be integer")
		}

		switch t := args[2].(type) {
		case SexpInt:
			end = int(t.ToInt64())
		case SexpChar:
			end = int(t)
		default:
			return SexpNull, errors.New("Third argument of slice must be integer")
		}

		switch t := args[0].(type) {
		case SexpArray:
			return SexpArray(t[start:end]), nil
		case SexpStr:
			return SexpStr(t[start:end]), nil
		}

		return SexpNull, errors.New("First argument of slice must be array or string")
	}
}

func LenFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case SexpArray:
			return NewSexpInt(len(t)), nil
		case SexpStr:
			return NewSexpInt(len(t)), nil
		case SexpHash:
			return NewSexpInt(HashCountKeys(t)), nil
		case SexpPair:
			if IsList(t) {
				arr, _ := ListToArray(t)
				return NewSexpInt(len(arr)), nil
			}
		case SexpSentinel:
			if t == SexpNull {
				return NewSexpInt(0), nil
			}
		}

		return NewSexpInt(0), fmt.Errorf("argument must be string or array but got %s", args[0])
	}
}

func AppendFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		switch t := args[0].(type) {
		case SexpArray:
			return SexpArray(append(t, args[1])), nil
		case SexpStr:
			return AppendStr(t, args[1])
		}

		return SexpNull, errors.New("First argument of append must be array or string")
	}
}

func ConcatFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		switch t := args[0].(type) {
		case SexpArray:
			return ConcatArray(t, args[1])
		case SexpStr:
			return ConcatStr(t, args[1])
		case SexpPair:
			return ConcatList(t, args[1])
		case SexpSentinel:
			if t == SexpNull {
				return args[1], nil
			}
		}

		return SexpNull, errors.New("expected strings or arrays")
	}
}

func ReadFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}
		str := ""
		switch t := args[0].(type) {
		case SexpStr:
			str = string(t)
		default:
			return SexpNull, WrongType
		}
		lexer := NewLexerFromStream(bytes.NewBuffer([]byte(str)))
		parser := Parser{lexer, env}
		return ParseExpression(&parser)
	}
}

func EvalFunction(env *Environment, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return WrongNumberArguments("eval", len(args), 1)
	}
	newenv := env.Duplicate()
	err := newenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, errors.New("failed to compile expression")
	}
	newenv.pc = 0
	return newenv.Run()
}

func GetTypeQueryFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		var result bool

		switch name {
		case "list?":
			result = IsList(args[0])
		case "null?":
			result = (args[0] == SexpNull)
		case "array?":
			result = IsArray(args[0])
		case "number?":
			result = IsNumber(args[0])
		case "float?":
			result = IsFloat(args[0])
		case "int?":
			result = IsInt(args[0])
		case "char?":
			result = IsChar(args[0])
		case "symbol?":
			result = IsSymbol(args[0])
		case "string?":
			result = IsString(args[0])
		case "hash?":
			result = IsHash(args[0])
		case "zero?":
			result = IsZero(args[0])
		case "empty?":
			result = IsEmpty(args[0])
		}

		return SexpBool(result), nil
	}
}

func GetPrintFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) == 0 {
			return SexpNull, fmt.Errorf("%s need at least one argument", name)
		}
		if name == `printf` {
			if len(args) <= 1 {
				return SexpNull, fmt.Errorf("%s need at least two argument", name)
			}
			if !IsString(args[0]) {
				return SexpNull, fmt.Errorf("first argument of %s must be string", name)
			}
		}

		var items []interface{}

		for _, item := range args {
			switch expr := item.(type) {
			case SexpStr:
				items = append(items, string(expr))
			default:
				items = append(items, expr.SexpString())
			}
		}

		switch name {
		case "println":
			fmt.Println(items...)
		case "print":
			fmt.Print(items...)
		case "printf":
			fmt.Printf(items[0].(string), items[1:]...)
		}

		return SexpNull, nil
	}
}

func NotFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		result := SexpBool(!IsTruthy(args[0]))
		return result, nil
	}
}

func ApplyFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}
		var fun SexpFunction
		var funargs SexpArray

		switch e := args[0].(type) {
		case SexpFunction:
			fun = e
		case SexpSymbol:
			var foundFn bool
			if rfn, ok := env.FindObject(e.Name()); ok {
				if IsFunction(rfn) {
					fun = rfn.(SexpFunction)
					foundFn = true
				}
			}
			if !foundFn {
				return SexpNull, fmt.Errorf(`can't find function by symbol %v`, e.Name())
			}
		default:
			return SexpNull, errors.New("first argument must be function")
		}

		switch e := args[1].(type) {
		case SexpArray:
			funargs = e
		case SexpPair:
			var err error
			funargs, err = ListToArray(e)
			if err != nil {
				return SexpNull, err
			}
		default:
			return SexpNull, errors.New("second argument must be array or list")
		}

		return env.Apply(fun, funargs)
	}
}

func MapFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}
		var fun SexpFunction

		switch e := args[0].(type) {
		case SexpFunction:
			fun = e
		default:
			return SexpNull, fmt.Errorf("first argument of map must be function had %T %v", e, e)
		}

		switch e := args[1].(type) {
		case SexpArray:
			return MapArray(env, fun, e)
		case SexpPair:
			return MapList(env, fun, e)
		}
		return SexpNull, errors.New("second argument ofr map must be array")
	}
}

func MakeArrayFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 1 {
			return WrongNumberArguments(name, len(args), 1, 2)
		}

		var size int
		switch e := args[0].(type) {
		case SexpInt:
			size = int(e.ToInt64())
		default:
			return SexpNull, errors.New("first argument must be integer")
		}

		var fill Sexp
		if len(args) == 2 {
			fill = args[1]
		} else {
			fill = SexpNull
		}

		arr := make([]Sexp, size)
		for i := range arr {
			arr[i] = fill
		}

		return SexpArray(arr), nil
	}
}

func GetConstructorFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		switch name {
		case "array":
			return SexpArray(args), nil
		case "list":
			return MakeList(args), nil
		case "hash":
			return MakeHash(args)
		}
		return SexpNull, errors.New("invalid constructor")
	}
}

func SymnumFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case SexpSymbol:
			return NewSexpInt(t.number), nil
		}
		return SexpNull, errors.New("argument must be symbol")
	}
}

func SourceFileFunction(env *Environment, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return WrongNumberArguments("source", len(args), 1)
	}

	var sourceItem func(item Sexp) error

	sourceItem = func(item Sexp) error {
		switch t := item.(type) {
		case SexpArray:
			for _, v := range t {
				if err := sourceItem(v); err != nil {
					return err
				}
			}
		case SexpPair:
			expr := item
			for expr != SexpNull {
				list := expr.(SexpPair)
				if err := sourceItem(list.head); err != nil {
					return err
				}
				expr = list.tail
			}
		case SexpStr:
			var f *os.File
			var err error

			if f, err = os.Open(string(t)); err != nil {
				return err
			}
			defer f.Close()

			if err = env.SourceFile(f); err != nil {
				return err
			}
		default:
			return fmt.Errorf("source-file: Expected `string`, `list`, `array` given type %T val %v", item, item)
		}

		return nil
	}

	for _, v := range args {
		if err := sourceItem(v); err != nil {
			return SexpNull, err
		}
	}

	return SexpNull, nil
}

var MissingFunction = SexpFunction{"__missing", true, 0, false, nil, nil, nil}

func MakeFunction(name string, nargs int, varargs bool,
	fun Function) SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.varargs = varargs
	sfun.fun = fun
	return sfun
}

func MakeUserFunction(name string, ufun UserFunction) SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return sfun
}

func BuiltinFunctions() map[string]UserFunction {
	ret := make(map[string]UserFunction)
	for name, cons := range builtinFunctions {
		ret[name] = cons(name)
	}
	return ret
}

var builtinFunctions = map[string]UserFunctionConstructor{
	"<":          GetCompareFunction,
	">":          GetCompareFunction,
	"<=":         GetCompareFunction,
	">=":         GetCompareFunction,
	"=":          GetCompareFunction,
	"not=":       GetCompareFunction,
	"sll":        GetBinaryIntFunction,
	"sra":        GetBinaryIntFunction,
	"srl":        GetBinaryIntFunction,
	"mod":        GetBinaryIntFunction,
	"+":          GetNumericFunction,
	"-":          GetNumericFunction,
	"*":          GetNumericFunction,
	"/":          GetNumericFunction,
	"bit-and":    GetBitwiseFunction,
	"bit-or":     GetBitwiseFunction,
	"bit-xor":    GetBitwiseFunction,
	"bit-not":    ComplementFunction,
	"read":       ReadFunction,
	"cons":       ConsFunction,
	"car":        FirstFunction,
	"cdr":        RestFunction,
	"list?":      GetTypeQueryFunction,
	"null?":      GetTypeQueryFunction,
	"array?":     GetTypeQueryFunction,
	"hash?":      GetTypeQueryFunction,
	"number?":    GetTypeQueryFunction,
	"int?":       GetTypeQueryFunction,
	"float?":     GetTypeQueryFunction,
	"char?":      GetTypeQueryFunction,
	"symbol?":    GetTypeQueryFunction,
	"string?":    GetTypeQueryFunction,
	"zero?":      GetTypeQueryFunction,
	"empty?":     GetTypeQueryFunction,
	"println":    GetPrintFunction,
	"print":      GetPrintFunction,
	"printf":     GetPrintFunction,
	"not":        NotFunction,
	"apply":      ApplyFunction,
	"map":        MapFunction,
	"make-array": MakeArrayFunction,
	"aget":       GetArrayAccessFunction,
	"aset!":      GetArrayAccessFunction,
	"sget":       SgetFunction,
	"hget":       GetHashAccessFunction,
	"hset!":      GetHashAccessFunction,
	"hdel!":      GetHashAccessFunction,
	"exist?":     GetExistFunction,
	"slice":      SliceFunction,
	"len":        LenFunction,
	"append":     AppendFunction,
	"concat":     ConcatFunction,
	"array":      GetConstructorFunction,
	"list":       GetConstructorFunction,
	"hash":       GetConstructorFunction,
	"symnum":     SymnumFunction,
	"str":        StringifyFunction,
	"str2int":    StringToNumber,
	"str2float":  StringToFloat,
	"typestr":    GetTypeFunction,
	"gensym":     GenSymFunction,
	"sym2str":    Sym2StrFunction,
	"str2sym":    Str2SymFunction,
	"str2bytes":  StringToBytes,
	"bytes2str":  BytesToString,
}

func StringifyFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		return SexpStr(args[0].SexpString()), nil
	}
}

func StringToNumber(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(SexpStr)
		if !ok {
			return SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		return NewSexpIntStr(string(str))
	}
}

func BytesToString(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(SexpBytes)
		if !ok {
			return SexpNull, fmt.Errorf(`%s argument should be bytes`, name)
		}
		return SexpStr(string(str.bytes)), nil
	}
}

func StringToBytes(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(SexpStr)
		if !ok {
			return SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		return NewSexpBytes([]byte(str)), nil
	}
}

func StringToFloat(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(SexpStr)
		if !ok {
			return SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		f, err := strconv.ParseFloat(string(str), 64)
		if err != nil {
			return SexpNull, err
		}
		return SexpFloat(f), nil
	}
}
