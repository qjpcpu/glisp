package extensions

import (
	"bytes"
	"encoding/json"
	"fmt"

	_ "embed"
	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/qjson"
)

var (
	//go:embed json_utils.lisp
	json_scripts string
)

func ImportJSON(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("json/stringify", jsonMarshal)
	env.AddNamedFunction("json/parse", jsonUnmarshal)
	env.AddNamedFunction("json/query", QueryJSONSexp)
	env.AddNamedFunction("json/set", SetJSONSexp)
	env.AddNamedFunction("json/del", DelJSONSexp)
	return env.SourceStream(bytes.NewBufferString(json_scripts))
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
}

func ParseJSON(rawBytes []byte) (glisp.Sexp, error) {
	if len(rawBytes) == 0 {
		return glisp.SexpNull, nil
	}
	tree, err := qjson.Decode(rawBytes)
	if err != nil {
		return glisp.SexpNull, fmt.Errorf("decode json fail %v", err)
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
