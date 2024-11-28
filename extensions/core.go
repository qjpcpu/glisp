package extensions

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/qjpcpu/glisp"
)

func ImportCoreUtils(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("println", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("printf", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("print", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("sprintf", GetPrintFunction(os.Stdout))
	env.AddNamedFunction("<", GetCompareFunction)
	env.AddNamedFunction(">", GetCompareFunction)
	env.AddNamedFunction("<=", GetCompareFunction)
	env.AddNamedFunction(">=", GetCompareFunction)
	env.AddNamedFunction("=", GetCompareFunction)
	env.AddNamedFunction("not=", GetCompareFunction)
	env.AddNamedFunction("!=", GetCompareFunction)
	env.AddNamedFunction("+", GetNumericFunction)
	env.AddNamedFunction("-", GetNumericFunction)
	env.AddNamedFunction("*", GetNumericFunction)
	env.AddNamedFunction("/", GetNumericFunction)
	env.AddNamedFunction("mod", GetBinaryIntFunction)
	env.AddNamedMacro("doc", GetDocFunction)
	env.AddNamedMacro("defined?", SymbolDefinedFunction)
	env.AddFuzzyMacro(`^:[^:]+$`, ExplainColonMacro)
	env.AddNamedFunction("sort", GetSortFunction)
	env.AddNamedFunction("compose", GetComposeFunction)
	/* stream */
	env.OverrideFunction("type", OverrideTypeFunction)
	env.AddNamedFunction("streamable?", IsStreamableFunction)
	env.AddNamedFunction("stream?", IsStreamFunction)
	env.AddNamedFunction("stream", StreamFunction)
	env.AddNamedFunction("map", StreamMapFunction)
	env.AddNamedFunction("flatmap", StreamFlatmapFunction)
	env.AddNamedFunction("filter", StreamFilterFunction)
	env.AddNamedFunction("take", StreamTakeFunction)
	env.AddNamedFunction("drop", StreamDropFunction)
	env.AddNamedFunction("foldl", StreamFoldlFunction)
	env.AddNamedFunction("realize", StreamRealizeFunction)
	env.AddNamedFunction("range", StreamRangeFunction)
	env.AddNamedFunction("partition", StreamPartitionFunction)
	env.AddNamedFunction("zip", StreamZipFunction)
	env.AddNamedFunction("union", StreamUnionFunction)

	/* record related */
	env.AddNamedMacro("defrecord", DefineRecord)
	env.AddNamedMacro("assoc", AssocRecordField)
	env.AddNamedFunction("record?", CheckIsRecord)
	env.AddNamedFunction("record-class?", CheckIsRecordClass)
	env.AddNamedFunction("get-record-class", GetRecordClass)
	env.AddNamedFunction("record-class-definition", ClassDefinition)
	env.AddNamedFunction("record-of?", CheckIsRecordOf)
	mustLoadScript(env.Environment, "core")

	/* buffer */
	env.AddNamedFunction("buffer", newBuffer)
	return nil
}

func GetPrintFunction(w io.Writer) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) == 0 {
				return glisp.SexpNull, fmt.Errorf("%s need at least one argument", name)
			}
			if name == `printf` || name == `sprintf` {
				if len(args) <= 1 {
					return glisp.SexpNull, fmt.Errorf("%s need at least two argument", name)
				}
				if !glisp.IsString(args[0]) {
					return glisp.SexpNull, fmt.Errorf("first argument of %s must be string", name)
				}
			}

			switch name {
			case "println":
				_, items := transformFmt("", args)
				fmt.Fprintln(w, items...)
			case "print":
				_, items := transformFmt("", args)
				fmt.Fprint(w, items...)
			case "printf":
				fmtstr, vargs := transformFmt(string(args[0].(glisp.SexpStr)), args[1:])
				fmt.Fprintf(w, fmtstr, vargs...)
			case "sprintf":
				fmtstr, vargs := transformFmt(string(args[0].(glisp.SexpStr)), args[1:])
				return glisp.SexpStr(fmt.Sprintf(fmtstr, vargs...)), nil
			}

			return glisp.SexpNull, nil
		}
	}
}

func mapSexpToGoPrintableInterface(fmtstr string, sexp glisp.Sexp) (string, interface{}) {
	if fmtstr == "" {
		fmtstr = "%v"
	}
	if sexp == glisp.SexpNull {
		return fmtstr, nil
	}
	switch expr := sexp.(type) {
	case glisp.SexpStr:
		return fmtstr, string(expr)
	case glisp.SexpBool:
		return fmtstr, bool(expr)
	case glisp.SexpInt:
		return "%s", expr.Format(fmtstr)
	case glisp.SexpFloat:
		return "%s", expr.Format(fmtstr)
	case glisp.SexpSymbol:
		return fmtstr, expr.Name()
	case glisp.SexpChar:
		return fmtstr, rune(expr)
	default:
		return fmtstr, expr.SexpString()
	}
}

func parseFmtStr(str string, args []glisp.Sexp) ([]string, []int) {
	var ret []string
	var cache []rune
	var mark []int
	addSymbol := func(sym string) {
		if len(cache) > 0 {
			ret = append(ret, string(cache))
			cache = nil
		}
		ret = append(ret, sym)
		mark = append(mark, len(ret)-1)
	}
	data := []rune(str)
	foundSym := -1
	for i := 0; i < len(data); {
		b := data[i]
		if foundSym == -1 {
			if b == '%' {
				if i < len(data)-1 && data[i+1] == '%' {
					cache = append(cache, '%', '%')
					i += 2
				} else {
					foundSym = i
					i++
				}
				continue
			} else {
				cache = append(cache, b)
				i++
			}
		} else {
			if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') {
				addSymbol(string(data[foundSym : i+1]))
				foundSym = -1
			}
			i++
		}
	}
	if len(cache) > 0 {
		ret = append(ret, string(cache))
		cache = nil
	}
	if extra := len(args) - len(mark); extra > 0 {
		ret = append(ret, "%%!(EXTRA ")
		for i := 0; i < extra; i++ {
			ret = append(ret, glisp.InspectType(args[len(args)-extra+i]), "=")
			ret = append(ret, "%v")
			mark = append(mark, len(ret)-1)
			if i < extra-1 {
				ret = append(ret, ", ")
			}
		}
		ret = append(ret, ")")
	}
	if missing := len(mark) - len(args); missing > 0 {
		for i := 0; i < missing; i++ {
			v := ret[mark[len(mark)-missing+i]]
			ret[mark[len(mark)-missing+i]] = "%%!" + strings.TrimPrefix(v, "%") + "(MISSING)"
		}
		mark = mark[:len(mark)-missing]
	}
	return ret, mark
}

func transformFmt(fmtstr string, args []glisp.Sexp) (string, []interface{}) {
	if fmtstr != "" {
		fmtStrs, mark := parseFmtStr(fmtstr, args)
		var ret []interface{}
		for i, item := range args {
			f, v := mapSexpToGoPrintableInterface(fmtStrs[mark[i]], item)
			fmtStrs[mark[i]] = f
			ret = append(ret, v)
		}
		return strings.Join(fmtStrs, ""), ret
	}
	var ret []interface{}
	for _, item := range args {
		_, v := mapSexpToGoPrintableInterface("%v", item)
		ret = append(ret, v)
	}
	return "", ret
}

func GetDocFunction(name string) glisp.UserFunction {
	userfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		name := args[0].(glisp.SexpSymbol).Name()
		var doc string
		if expr, ok := env.FindObject(name); ok && glisp.IsFunction(expr) {
			doc = expr.(*glisp.SexpFunction).Doc()
		} else if mac, ok := env.FindMacro(name); ok {
			doc = mac.Doc()
		} else {
			doc = glisp.QueryBuiltinDoc(name)
		}
		if doc == `` {
			doc = `No document found.`
		}
		return glisp.SexpStr(doc), nil
	}
	sexpfn := glisp.MakeUserFunction(name, userfn)
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		if !glisp.IsSymbol(args[0]) {
			return glisp.SexpNull, fmt.Errorf("argument of %s should be symbol but got %v", name, glisp.InspectType(args[0]))
		}
		return glisp.MakeList([]glisp.Sexp{
			env.MakeSymbol("println"),
			glisp.MakeList([]glisp.Sexp{
				sexpfn,
				glisp.MakeList([]glisp.Sexp{
					env.MakeSymbol("quote"),
					args[0].(glisp.SexpSymbol),
				}),
			}),
		}), nil
	}
}

func SymbolDefinedFunction(name string) glisp.UserFunction {
	userfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		var name string
		switch args[0].(type) {
		case *glisp.SexpFunction:
			return glisp.SexpBool(true), nil
		case glisp.SexpSymbol:
			name = args[0].(glisp.SexpSymbol).Name()
		case glisp.SexpStr:
			name = string(args[0].(glisp.SexpStr))
		default:
			return glisp.SexpNull, fmt.Errorf("can't guess %v definition", glisp.InspectType(args[0]))
		}
		if _, ok := env.FindObject(name); ok {
			return glisp.SexpBool(true), nil
		} else if _, ok = env.FindMacro(name); ok {
			return glisp.SexpBool(true), nil
		} else {
			return glisp.SexpBool(false), nil
		}
	}
	sexpfn := glisp.MakeUserFunction(name, userfn)
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		newArgs := append([]glisp.Sexp{sexpfn}, args...)
		return glisp.MakeList(newArgs), nil
	}
}

type ExplainSexp interface {
	Explain(*glisp.Environment, string, []glisp.Sexp) (glisp.Sexp, error)
}

func ExplainColonMacro(name string) glisp.UserFunction {
	sexpfn := glisp.MakeUserFunction(name, func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if ex, ok := args[1].(ExplainSexp); ok {
			return ex.Explain(env, string(args[0].(glisp.SexpStr)), args[2:])
		}
		/* nil explain anything as nil */
		if args[1] == glisp.SexpNull {
			return glisp.SexpNull, nil
		}
		return glisp.SexpNull, fmt.Errorf("type `%s` can't explain `%s`", glisp.InspectType(args[1]), args[0].SexpString())
	})
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 2 {
			return glisp.WrongNumberArguments(name, len(args), 2, glisp.Many)
		}
		if !glisp.IsString(args[0]) {
			return glisp.SexpNull, fmt.Errorf("%s first argument must be string but got %s", name, glisp.InspectType(args[0]))
		}
		colon := string(args[0].(glisp.SexpStr))
		vargs := []glisp.Sexp{sexpfn, glisp.SexpStr(colon[1:])}
		vargs = append(vargs, args[1:]...)
		return glisp.MakeList(vargs), nil
	}
}

func GetComposeFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 2 {
			return glisp.WrongNumberArguments(name, len(args), 2, glisp.Many)
		}
		for _, fn := range args {
			if !glisp.IsFunction(fn) {
				return glisp.SexpNull, fmt.Errorf("argument should be function but got %v", glisp.InspectType(fn))
			}
		}
		return glisp.MakeUserFunction(env.GenSymbol("__compose").Name(), func(_env *glisp.Environment, _args []glisp.Sexp) (glisp.Sexp, error) {
			for i := len(args) - 1; i >= 0; i-- {
				fn := args[i].(*glisp.SexpFunction)
				ret, err := _env.Apply(fn, _args)
				if err != nil {
					return glisp.SexpNull, err
				}
				_args = []glisp.Sexp{ret}
			}
			/* len(_args) is greater than 0, because function always return something */
			return _args[0], nil
		}), nil
	}
}
