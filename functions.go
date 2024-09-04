package glisp

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	//go:embed builtin.clj
	buitin_scripts string
)

type Function []Instruction
type UserFunction func(*Environment, []Sexp) (Sexp, error)
type NamedUserFunction func(string) UserFunction
type OverrideFunction func(*SexpFunction) UserFunction

var builtinFunctions = map[string]NamedUserFunction{
	"read":       GetReadFunction,
	"cons":       GetConsFunction,
	"car":        GetFirstFunction,
	"cdr":        GetRestFunction,
	"list?":      GetTypeQueryFunction,
	"nil?":       GetTypeQueryFunction,
	"array?":     GetTypeQueryFunction,
	"hash?":      GetTypeQueryFunction,
	"number?":    GetTypeQueryFunction,
	"int?":       GetTypeQueryFunction,
	"float?":     GetTypeQueryFunction,
	"char?":      GetTypeQueryFunction,
	"symbol?":    GetTypeQueryFunction,
	"string?":    GetTypeQueryFunction,
	"zero?":      GetTypeQueryFunction,
	"bool?":      GetTypeQueryFunction,
	"empty?":     GetTypeQueryFunction,
	"bytes?":     GetTypeQueryFunction,
	"function?":  GetTypeQueryFunction,
	"not":        GetNotFunction,
	"apply":      GetApplyFunction,
	"make-array": GetMakeArrayFunction,
	"aget":       GetArrayAccessFunction,
	"aset!":      GetArrayAccessFunction,
	"sget":       GetSgetFunction,
	"hget":       GetHashAccessFunction,
	"hset!":      GetHashAccessFunction,
	"hdel!":      GetHashAccessFunction,
	"exist?":     GetExistFunction,
	"slice":      GetSliceFunction,
	"len":        GetLenFunction,
	"append":     GetAppendFunction,
	"concat":     GetConcatFunction,
	"array":      GetConstructorFunction,
	"list":       GetConstructorFunction,
	"hash":       GetConstructorFunction,
	"symnum":     GetSymnumFunction,
	"string":     GetStringifyFunction,
	"sexp-str":   GetSexpString,
	"int":        GetAnyToInteger,
	"float":      GetAnyToFloat,
	"char":       GetAnyToChar,
	"bool":       GetAnyToBool,
	"type":       GetTypeFunction,
	"gensym":     GetGenSymFunction,
	"symbol":     GetAnyToSymbolFunction,
	"bytes":      GetAnyToBytes,
}

func GetConsFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}

		return Cons(args[0], args[1]), nil
	}
}

func GetFirstFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch expr := args[0].(type) {
		case SexpSentinel:
			if expr == SexpNull {
				return SexpNull, nil
			}
		case *SexpPair:
			return expr.head, nil
		case SexpArray:
			if len(expr) == 0 {
				return SexpNull, errors.New(`access an empty array`)
			}
			return expr[0], nil
		}

		return SexpNull, WrongType
	}
}

func GetRestFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch expr := args[0].(type) {
		case *SexpPair:
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
			return SexpNull, fmt.Errorf("first argument of aget must be array but got <%v:%v>", InspectType(args[0]), args[0].SexpString())
		}

		var i int
		switch t := args[1].(type) {
		case SexpInt:
			i = int(t.ToInt64())
		case SexpChar:
			i = int(t)
		default:
			return SexpNull, fmt.Errorf("second argument of aget must be integer but got <%v:%v>", InspectType(args[1]), args[1].SexpString())
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

		return arr, nil
	}
}

func GetSgetFunction(name string) UserFunction {
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

		if i < 0 || i >= len(string(str)) {
			return SexpNull, errors.New("string index out of bounds")
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
		case *SexpHash:
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
		case *SexpPair:
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

		var hash *SexpHash
		switch e := args[0].(type) {
		case *SexpHash:
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
			if len(args) != 3 {
				return WrongNumberArguments(name, len(args), 3)
			}
			err := hash.HashSet(args[1], args[2])
			return hash, err
		case "hdel!":
			if len(args) != 2 {
				return WrongNumberArguments(name, len(args), 2)
			}
			err := hash.HashDelete(args[1])
			return hash, err
		}

		return hash, nil
	}
}

func GetSliceFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 3 && len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2, 3)
		}

		var start int
		end := math.MaxInt
		switch t := args[1].(type) {
		case SexpInt:
			start = int(t.ToInt64())
		case SexpChar:
			start = int(t)
		default:
			return SexpNull, errors.New("Second argument of slice must be integer")
		}

		if len(args) == 3 {
			switch t := args[2].(type) {
			case SexpInt:
				end = int(t.ToInt64())
			case SexpChar:
				end = int(t)
			default:
				return SexpNull, errors.New("Third argument of slice must be integer")
			}
		}

		min := func(i, j int) int {
			if i > j {
				return j
			}
			return i
		}
		switch t := args[0].(type) {
		case SexpArray:
			if start < 0 || start > len(t) || end < start {
				return SexpNull, errors.New("index out of range")
			}
			end = min(end, len(t))
			return SexpArray(t[start:end]), nil
		case SexpStr:
			size := lenOfStr(string(t))
			if start < 0 || start > size || end < start {
				return SexpNull, errors.New("index out of range")
			}
			end = min(end, size)
			return SexpStr(sliceOfStr(string(t), start, end)), nil
		case SexpBytes:
			if start < 0 || start > len(t.Bytes()) || end < start {
				return SexpNull, errors.New("index out of range")
			}
			end = min(end, len(t.Bytes()))
			return NewSexpBytes(t.Bytes()[start:end]), nil
		}

		return SexpNull, errors.New("First argument of slice must be array or string or bytes")
	}
}

func GetLenFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		switch t := args[0].(type) {
		case SexpArray:
			return NewSexpInt(len(t)), nil
		case SexpStr:
			return NewSexpInt(lenOfStr(string(t))), nil
		case *SexpHash:
			n, err := HashCountKeys(t)
			return NewSexpInt(n), err
		case *SexpPair:
			if IsList(t) {
				arr, _ := ListToArray(t)
				return NewSexpInt(len(arr)), nil
			}
		case SexpSentinel:
			if t == SexpNull {
				return NewSexpInt(0), nil
			}
		case SexpBytes:
			return NewSexpInt(len(t.bytes)), nil
		}

		return NewSexpInt(0), fmt.Errorf("argument must be string/array/list/hash/bytes but got %s", InspectType(args[0]))
	}
}

func GetAppendFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 2 {
			return WrongNumberArguments(name, len(args), 2, Many)
		}

		switch t := args[0].(type) {
		case SexpArray:
			return SexpArray(append(t, args[1:]...)), nil
		case SexpStr:
			return AppendStr(t, args[1:]...)
		}

		return SexpNull, errors.New("First argument of append must be array or string but got " + InspectType(args[0]))
	}
}

func GetConcatFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) < 2 {
			return WrongNumberArguments(name, len(args), 2, Many)
		}
		return concatSexp(args)
	}
}

func concatSexp(args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		return SexpNull, nil
	}
	switch t := args[0].(type) {
	case SexpArray:
		return ConcatArray(t, args[1:]...)
	case SexpStr:
		return ConcatStr(t, args[1:]...)
	case *SexpPair:
		return ConcatList(t, args[1:]...)
	case SexpBytes:
		return ConcatBytes(t, args[1:]...)
	case *SexpHash:
		h, _ := MakeHash(nil)
		return ConcatHash(h, args...)
	case SexpSentinel:
		if t == SexpNull {
			return concatSexp(args[1:])
		}
	}
	return SexpNull, errors.New("expected strings/arrays/lists/bytes but got " + InspectType(args[0]))
}

func GetReadFunction(name string) UserFunction {
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

func GetEvalFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}
		newenv := env.Duplicate()
		err := newenv.LoadExpressions(args)
		if err != nil {
			return SexpNull, fmt.Errorf("failed to compile expression: %v", err)
		}
		newenv.pc = 0
		return newenv.Run()
	}
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
		case "nil?":
			result = args[0] == SexpNull
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
		case "bytes?":
			result = IsBytes(args[0])
		case "bool?":
			result = IsBool(args[0])
		case "function?":
			result = IsFunction(args[0])
		}

		return SexpBool(result), nil
	}
}

func GetNotFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}

		result := SexpBool(!IsTruthy(args[0]))
		return result, nil
	}
}

func GetApplyFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}
		var fun *SexpFunction
		var funargs SexpArray

		switch e := args[0].(type) {
		case *SexpFunction:
			fun = e
		case SexpSymbol:
			var foundFn bool
			if rfn, ok := env.FindObject(e.Name()); ok {
				if IsFunction(rfn) {
					fun = rfn.(*SexpFunction)
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
		case *SexpPair:
			var err error
			funargs, err = ListToArray(e)
			if err != nil {
				return SexpNull, err
			}
		default:
			return SexpNull, fmt.Errorf("second argument must be array or list but got %v", InspectType(args[1]))
		}

		return env.Apply(fun, funargs)
	}
}

func GetMakeArrayFunction(name string) UserFunction {
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

func GetSymnumFunction(name string) UserFunction {
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

func GetSourceFileFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) == 0 {
			return WrongNumberArguments(name, len(args), 1, Many)
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
			case *SexpPair:
				expr := item
				for expr != SexpNull {
					list := expr.(*SexpPair)
					if err := sourceItem(list.head); err != nil {
						return err
					}
					expr = list.tail
				}
			case SexpStr:
				var f io.ReadCloser
				var err error

				if f, err = env.fileReader.Open(string(t)); err != nil {
					return err
				}
				defer f.Close()

				if err = env.SourceStream(f); err != nil {
					return err
				}
			default:
				return fmt.Errorf("source-file: Expected `string`, `list`, `array` given %v", InspectType(item))
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
}

var MissingFunction = &SexpFunction{
	name:    "__missing",
	user:    true,
	nargs:   0,
	varargs: false,
	fun:     nil,
}

type FuntionOption func(*SexpFunction)

func WithDoc(doc string) FuntionOption {
	return func(f *SexpFunction) {
		if doc != `` {
			f.doc = doc
		}
	}
}

func withNameRegexp(exp string) FuntionOption {
	return func(f *SexpFunction) {
		if exp != "" {
			f.nameRegexp = regexp.MustCompile(exp)
		}
	}
}

func MakeFunction(name string, nargs int, varargs bool, fun Function, opts ...FuntionOption) *SexpFunction {
	var sfun = &SexpFunction{}
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.varargs = varargs
	sfun.fun = fun
	return setFuncOpts(sfun, opts...)
}

func MakeUserFunction(name string, ufun UserFunction, opts ...FuntionOption) *SexpFunction {
	var sfun = &SexpFunction{}
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return setFuncOpts(sfun, opts...)
}

func setFuncOpts(f *SexpFunction, opts ...FuntionOption) *SexpFunction {
	for _, fn := range opts {
		fn(f)
	}
	return f
}

func BuiltinFunctions() map[string]UserFunction {
	ret := make(map[string]UserFunction)
	for name, cons := range builtinFunctions {
		ret[name] = cons(name)
	}
	return ret
}

// GetStringifyFunction return s-expr's string representation
func GetStringifyFunction(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return WrongNumberArguments(name, len(args), 1, 2)
		}
		var sb strings.Builder
		if err := sexpToString(&sb, args); err != nil {
			return SexpNull, err
		}
		return SexpStr(sb.String()), nil
	}
}

func GetSexpString(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}
		return SexpStr(args[0].SexpString()), nil
	}
}

func GetAnyToInteger(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return WrongNumberArguments(name, len(args), 1)
		}
		switch val := args[0].(type) {
		case SexpChar:
			return NewSexpInt64(int64(val)), nil
		case SexpFloat:
			integer := new(big.Int)
			val.v.Int(integer)
			return SexpInt{v: integer}, nil
		case SexpInt:
			return val, nil
		case SexpStr:
			return NewSexpIntStr(string(val))
		case SexpBytes:
			return NewSexpIntBytes(val.Bytes()), nil
		}
		return SexpNull, fmt.Errorf(`%s argument should be char/float/str/int/bytes but got %v`, name, InspectType(args[0]))
	}
}

func GetAnyToChar(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		switch val := args[0].(type) {
		case SexpInt:
			return SexpChar(rune(val.ToInt())), nil
		case SexpStr:
			rs := []rune(val)
			if len(rs) != 1 {
				return SexpNull, fmt.Errorf("%s expect string only contains 1 char but got %v", name, len(rs))
			}
			return SexpChar(rs[0]), nil
		}
		return SexpNull, fmt.Errorf(`%s argument should be integer but got %v`, name, InspectType(args[0]))
	}
}

func GetAnyToBool(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		switch val := args[0].(type) {
		case SexpBool:
			return val, nil
		case SexpStr:
			return SexpBool(string(val) == `true`), nil
		}
		return SexpNull, fmt.Errorf(`%s argument should be string/bool but got %v`, name, InspectType(args[0]))
	}
}

func GetAnyToBytes(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		switch v := args[0].(type) {
		case SexpStr:
			return NewSexpBytes([]byte(string(v))), nil
		case SexpInt:
			return NewSexpBytes(v.ToBytes()), nil
		default:
			return SexpNull, fmt.Errorf(`%s argument should be string/int but got %v`, name, InspectType(args[0]))
		}
	}
}

func GetAnyToFloat(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		switch val := args[0].(type) {
		case SexpStr:
			return NewSexpFloatStr(string(val))
		case SexpFloat:
			return val, nil
		case SexpInt:
			return NewSexpFloatInt(val), nil
		}
		return SexpNull, fmt.Errorf(`%s argument should be string/int/float but got %v`, name, InspectType(args[0]))
	}
}

func lenOfStr(s string) int {
	return utf8.RuneCountInString(s)
}

func sliceOfStr(s string, i, j int) string {
	runes := []rune(s)
	return string(runes[i:j])
}

func sexpToString(sb *strings.Builder, args []Sexp) error {
	if args[0] == SexpNull {
		return nil
	}
	switch val := args[0].(type) {
	case SexpBytes:
		sb.WriteString(string(val.bytes))
	case SexpStr:
		sb.WriteString(string(val))
	case SexpFloat:
		if len(args) == 2 {
			if !IsInt(args[1]) {
				return fmt.Errorf("prec should be integer but got %v", InspectType(args[1]))
			}
			sb.WriteString(val.ToString(args[1].(SexpInt).ToInt()))
			return nil
		}
		sb.WriteString(val.SexpString())
	case SexpChar:
		sb.WriteRune(rune(val))
	case SexpArray:
		for _, elem := range val {
			if err := sexpToString(sb, []Sexp{elem}); err != nil {
				return err
			}
		}
	case *SexpPair:
		var err error
		val.Foreach(func(e Sexp) bool {
			err = sexpToString(sb, []Sexp{e})
			return err == nil
		})
	case Stringer:
		sb.WriteString(val.Stringify())
	default:
		sb.WriteString(args[0].SexpString())
	}
	return nil
}
