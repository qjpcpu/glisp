package extensions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qjpcpu/glisp"
)

func ImportJSON(env *glisp.Environment) {
	env.AddFunction("json/stringify", jsonMarshal)
	env.AddFunction("json/parse", jsonUnmarshal)
}

func jsonMarshal(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 1 {
		return glisp.WrongNumberArguments(env.CallName(), len(args), 1)
	}
	bytes, err := glisp.Marshal(args[0])
	if err != nil {
		return glisp.SexpNull, err
	}
	return glisp.SexpStr(string(bytes)), nil
}

func jsonUnmarshal(env *glisp.Context, args []glisp.Sexp) (glisp.Sexp, error) {
	name := env.CallName()
	if len(args) != 1 {
		return glisp.WrongNumberArguments(name, len(args), 1)
	}
	switch val := args[0].(type) {
	case glisp.SexpStr:
		rawBytes := []byte(string(val))
		return ParseJSON(rawBytes)
	case glisp.SexpBytes:
		return ParseJSON(val.Bytes())
	case glisp.SexpInt:
		return val, nil
	case glisp.SexpBool:
		return val, nil
	case glisp.SexpFloat:
		return val, nil
	default:
		return glisp.SexpNull, fmt.Errorf("the first argument of %s must be string/bytes", name)
	}
}

func ParseJSON(rawBytes []byte) (glisp.Sexp, error) {
	var v interface{}
	if len(rawBytes) == 0 {
		return glisp.SexpNull, nil
	}
	if err := stdUnmarshal(rawBytes, &v); err != nil {
		return glisp.SexpNull, fmt.Errorf("decode json fail %v", err)
	}
	return mapInterfaceToSexp(v), nil
}

func stdUnmarshal(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	return dec.Decode(v)
}

func mapInterfaceToSexp(v interface{}) glisp.Sexp {
	if v == nil {
		return glisp.SexpNull
	}
	switch val := v.(type) {
	case map[string]interface{}:
		arr := make(glisp.SexpArray, 0, 10)
		for k, v := range val {
			arr = append(arr,
				glisp.SexpStr(k),
				mapInterfaceToSexp(v),
			)
		}
		hash, _ := glisp.MakeHash(arr)
		return hash
	case []interface{}:
		arr := make(glisp.SexpArray, 0, 10)
		for _, item := range val {
			arr = append(arr, mapInterfaceToSexp(item))
		}
		return arr
	case bool:
		return glisp.SexpBool(val)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		expr, _ := glisp.NewSexpIntStrWithBase(fmt.Sprint(val), 10)
		return expr
	case float32:
		return glisp.NewSexpFloat(float64(val))
	case float64:
		return glisp.NewSexpFloat(val)
	case string:
		return glisp.SexpStr(val)
	case json.Number:
		str := val.String()
		if strings.Contains(str, ".") {
			num, _ := glisp.NewSexpFloatStr(str)
			return num
		}
		expr, _ := glisp.NewSexpIntStrWithBase(str, 10)
		return expr
	}
	return glisp.SexpNull
}
