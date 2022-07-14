package extensions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/qjpcpu/glisp"
)

func ImportJSON(env *glisp.Environment) {
	env.AddFunctionByConstructor("json/stringify", jsonMarshal)
	env.AddFunctionByConstructor("json/parse", jsonUnmarshal)
}

func jsonMarshal(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return wrongNumberArguments(name, len(args), 1)
		}
		bytes, err := glisp.Marshal(args[0])
		if err != nil {
			return glisp.SexpNull, err
		}
		return glisp.SexpStr(string(bytes)), nil
	}
}

func jsonUnmarshal(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return wrongNumberArguments(name, len(args), 1)
		}
		switch val := args[0].(type) {
		case glisp.SexpStr:
			rawBytes := []byte(string(val))
			return ParseJSON(rawBytes)
		case glisp.SexpBytes:
			return ParseJSON(val.Bytes())
		default:
			return glisp.SexpNull, fmt.Errorf("the first argument of %s must be string/bytes", name)
		}
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
		return glisp.SexpFloat(float64(val))
	case float64:
		return glisp.SexpFloat(val)
	case string:
		return glisp.SexpStr(val)
	case json.Number:
		str := val.String()
		if strings.Contains(str, ".") {
			num, _ := strconv.ParseFloat(str, 64)
			return glisp.SexpFloat(num)
		}
		expr, _ := glisp.NewSexpIntStrWithBase(str, 10)
		return expr
	}
	return glisp.SexpNull
}
