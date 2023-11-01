package extensions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/qjson"
)

func ImportJSON(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("json/stringify", jsonMarshal)
	env.AddNamedFunction("json/parse", jsonUnmarshal)
	env.AddNamedFunction("json/query", QueryJSONSexp)
	env.AddNamedFunction("json/set", SetJSONSexp)
	env.AddNamedFunction("json/del", DelJSONSexp)
	mustLoadScript(env.Environment, "json")
	return nil
}

func jsonMarshal(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 1, 2)
		}
		bs, err := glisp.Marshal(args[0])
		if err != nil {
			return glisp.SexpNull, err
		}
		if len(args) == 2 && glisp.IsBool(args[1]) && bool(args[1].(glisp.SexpBool)) {
			buf := new(bytes.Buffer)
			json.Indent(buf, bs, "", "  ")
			return glisp.SexpStr(buf.String()), nil
		}
		return glisp.SexpStr(string(bs)), nil
	}
}

func jsonUnmarshal(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 1, 2)
		}
		makeRes := func(s glisp.Sexp, err error) (glisp.Sexp, error) {
			if len(args) == 2 && err != nil {
				return args[1], nil
			}
			return s, err
		}
		switch val := args[0].(type) {
		case glisp.SexpStr:
			rawBytes := []byte(string(val))
			return makeRes(ParseJSON(rawBytes))
		case glisp.SexpBytes:
			return makeRes(ParseJSON(val.Bytes()))
		case glisp.SexpInt:
			return val, nil
		case glisp.SexpBool:
			return val, nil
		case glisp.SexpFloat:
			return val, nil
		default:
			if len(args) == 2 {
				return args[1], nil
			}
			return glisp.SexpNull, fmt.Errorf("the first argument of %s must be string/bytes/int/bool/float but got %v", name, glisp.InspectType(args[0]))
		}
	}
}

func ParseJSON(rawBytes []byte) (glisp.Sexp, error) {
	if len(bytes.TrimSpace(rawBytes)) == 0 {
		return glisp.SexpNull, errors.New(`unexpected end of JSON empty input`)
	}
	tree, err := qjson.Decode(rawBytes)
	if err != nil {
		return glisp.SexpNull, fmt.Errorf("decode json fail %v, input is %v", err, string(rawBytes))
	}
	defer tree.Release()
	return mapJsonNodeToSexp(tree.Root), nil
}

func mapJsonKeyNodeToSexp(v *qjson.Node) glisp.Sexp {
	/* in fact, json key should only be string */
	switch v.Type {
	case qjson.Bool:
		return glisp.SexpBool(v.Value == `true`)
	case qjson.Integer:
		v, _ := glisp.NewSexpIntStr(v.Value)
		return v
	case qjson.String:
		return glisp.SexpStr(v.AsString())
	}
	return glisp.SexpStr(v.Value)
}

func mapJsonNodeToSexp(v *qjson.Node) glisp.Sexp {
	if v == nil {
		return glisp.SexpNull
	}
	switch v.Type {
	case qjson.Null:
		return glisp.SexpNull
	case qjson.Object:
		arr := make(glisp.SexpArray, 0, len(v.ObjectValues)*2)
		for _, elem := range v.ObjectValues {
			arr = append(arr,
				mapJsonKeyNodeToSexp(elem.Key),
				mapJsonNodeToSexp(elem.Value),
			)
		}
		hash, _ := glisp.MakeHash(arr)
		return hash
	case qjson.Array:
		arr := make(glisp.SexpArray, 0, len(v.ArrayValues))
		for _, item := range v.ArrayValues {
			arr = append(arr, mapJsonNodeToSexp(item))
		}
		return arr
	case qjson.Bool:
		return glisp.SexpBool(v.Value == `true`)
	case qjson.Integer:
		expr, _ := glisp.NewSexpIntStr(v.Value)
		return expr
	case qjson.Float:
		v, _ := glisp.NewSexpFloatStr(v.Value)
		return v
	case qjson.String:
		return glisp.SexpStr(v.AsString())
	}
	return glisp.SexpNull
}

func stdUnmarshal(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	return dec.Decode(v)
}
