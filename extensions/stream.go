package extensions

import (
	"errors"
	"fmt"

	"github.com/qjpcpu/glisp"
)

func StreamFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		if !IsStreamable(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`type %s is not streamable`, glisp.InspectType(args[0]))
		}
		return expr2Stream(args[0]), nil
	}
}

func expr2Stream(v glisp.Sexp) iStream {
	if IsStream(v) {
		return v.(iStream)
	}
	switch expr := v.(type) {
	case *glisp.SexpPair:
		return &ListIterator{expr: expr}
	case glisp.SexpArray:
		return &ArrayIterator{expr: expr}
	case glisp.SexpBytes:
		return &BytesIterator{expr: expr}
	case glisp.SexpStr:
		return &StringIterator{expr: []rune(expr)}
	case *glisp.SexpHash:
		return &HashIterator{expr: expr}
	}
	if v == glisp.SexpNull {
		return &ListIterator{expr: glisp.SexpNull}
	}
	iter := v.(Iterable)
	return &IterableStream{expr: iter}
}

func IsStreamFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		return glisp.SexpBool(IsStream(args[0])), nil
	}
}

func IsStreamableFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		return glisp.SexpBool(IsStreamable(args[0])), nil
	}
}

func IsStream(expr glisp.Sexp) bool {
	_, ok := expr.(iStream)
	return ok
}

func IsStreamable(expr glisp.Sexp) bool {
	if expr == glisp.SexpNull {
		return true
	}
	if _, ok := expr.(Iterable); ok {
		return true
	}
	if IsStream(expr) {
		return true
	}
	switch expr.(type) {
	case *glisp.SexpPair:
		return true
	case glisp.SexpArray:
		return true
	case glisp.SexpBytes:
		return true
	case glisp.SexpStr:
		return true
	case *glisp.SexpHash:
		return true
	}
	return false
}

func StreamMapFunction(name string) glisp.UserFunction {
	normalfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		fun := args[0].(*glisp.SexpFunction)
		switch e := args[1].(type) {
		case glisp.SexpArray:
			return glisp.MapArray(env, fun, e)
		case *glisp.SexpPair:
			return glisp.MapList(env, fun, e)
		case *glisp.SexpHash:
			return glisp.MapHash(env, fun, e)
		case glisp.SexpSentinel:
			if e == glisp.SexpNull {
				return glisp.SexpNull, nil
			}
		}
		return glisp.SexpNull, errors.New("second argument of map must be array/list but got " + glisp.InspectType(args[1]))
	}
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !glisp.IsFunction(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !IsStream(args[1]) {
			return normalfn(env, args)
		}
		f, stream := args[0].(*glisp.SexpFunction), args[1].(iStream)
		return &mapIterator{iStream: stream, f: f}, nil
	}
}

func StreamFlatmapFunction(name string) glisp.UserFunction {
	normalfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		fun := args[0].(*glisp.SexpFunction)
		switch e := args[1].(type) {
		case glisp.SexpArray:
			return glisp.FlatMapArray(env, fun, e)
		case *glisp.SexpHash:
			return glisp.FlatMapHash(env, fun, e)
		case *glisp.SexpPair:
			return glisp.FlatMapList(env, fun, e)
		case glisp.SexpSentinel:
			if e == glisp.SexpNull {
				return glisp.SexpNull, nil
			}
		}
		return glisp.SexpNull, errors.New("second argument of map must be array/list")
	}
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !glisp.IsFunction(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !IsStream(args[1]) {
			return normalfn(env, args)
		}
		f, stream := args[0].(*glisp.SexpFunction), args[1].(iStream)
		return &flatmapIterator{iStream: stream, f: f}, nil
	}
}

func StreamFilterFunction(name string) glisp.UserFunction {
	normalfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		fun := args[0].(*glisp.SexpFunction)
		switch e := args[1].(type) {
		case glisp.SexpArray:
			return glisp.FilterArray(env, fun, e)
		case *glisp.SexpPair:
			return glisp.FilterList(env, fun, e)
		case *glisp.SexpHash:
			return glisp.FilterHash(env, fun, e)
		case glisp.SexpSentinel:
			if e == glisp.SexpNull {
				return e, nil
			}
		}
		return glisp.SexpNull, fmt.Errorf("second argument of %s must be array/list but got %s", name, glisp.InspectType(args[1]))
	}
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !glisp.IsFunction(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !IsStream(args[1]) {
			return normalfn(env, args)
		}
		f, stream := args[0].(*glisp.SexpFunction), args[1].(iStream)
		return &filterIterator{iStream: stream, f: f}, nil
	}
}

func StreamTakeFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !IsStream(args[1]) {
			return glisp.SexpNull, fmt.Errorf("second argument of %s must be stream, but got %v", name, glisp.InspectType(args[1]))
		}
		if glisp.IsInt(args[0]) {
			num, stream := args[0].(glisp.SexpInt), args[1].(iStream)
			return &takeIterator{iStream: stream, count: num.ToUint64()}, nil
		} else if glisp.IsFunction(args[0]) {
			f, stream := args[0].(*glisp.SexpFunction), args[1].(iStream)
			return &takeIterator{iStream: stream, f: f}, nil
		}
		return glisp.SexpNull, fmt.Errorf(`first argument of %s must be int/function, but got %v`, name, glisp.InspectType(args[0]))
	}
}

func StreamDropFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !IsStream(args[1]) {
			return glisp.SexpNull, fmt.Errorf("second argument of %s must be stream, but got %v", name, glisp.InspectType(args[1]))
		}
		if glisp.IsInt(args[0]) {
			num, stream := args[0].(glisp.SexpInt), args[1].(iStream)
			return &dropIterator{iStream: stream, count: num.ToUint64()}, nil
		} else if glisp.IsFunction(args[0]) {
			f, stream := args[0].(*glisp.SexpFunction), args[1].(iStream)
			return &dropIterator{iStream: stream, f: f}, nil
		}
		return glisp.SexpNull, fmt.Errorf(`first argument of %s must be int/function, but got %v`, name, glisp.InspectType(args[0]))
	}
}

func OverrideTypeFunction(orig *glisp.SexpFunction) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) == 1 && IsStream(args[0]) && !glisp.IsIType(args[0]) {
			return glisp.SexpStr(`stream`), nil
		}
		return env.Apply(orig, args)
	}
}

func StreamFoldlFunction(name string) glisp.UserFunction {
	normalFn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		fun := args[0].(*glisp.SexpFunction)
		switch e := args[2].(type) {
		case glisp.SexpArray:
			return glisp.FoldlArray(env, fun, e, args[1])
		case *glisp.SexpPair:
			return glisp.FoldlList(env, fun, e, args[1])
		case *glisp.SexpHash:
			return glisp.FoldlHash(env, fun, e, args[1])
		case glisp.SexpSentinel:
			if e == glisp.SexpNull {
				return args[1], nil
			}
		}
		return glisp.SexpNull, fmt.Errorf("third argument of %s must be array/list", name)
	}
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 3)
		}
		if !glisp.IsFunction(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !IsStream(args[2]) {
			return normalFn(env, args)
		}
		f, acc, stream := args[0].(*glisp.SexpFunction), args[1], args[2].(iStream)
		for {
			elem, ok, err := stream.Next(env)
			if err != nil {
				return glisp.SexpNull, err
			}
			if !ok {
				break
			}
			ret, err := env.Apply(f, []glisp.Sexp{elem, acc})
			if err != nil {
				return glisp.SexpNull, err
			}
			acc = ret
		}
		return acc, nil
	}
}

func StreamRealizeFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		if !IsStream(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`type %s is not stream`, glisp.InspectType(args[0]))
		}
		stream := args[0].(iStream)
		builder := glisp.NewListBuilder()
		for {
			elem, ok, err := stream.Next(env)
			if err != nil {
				return glisp.SexpNull, err
			}
			if !ok {
				break
			}
			builder.Add(elem)
		}
		return builder.Get(), nil
	}
}

func StreamRangeFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		var numbers []glisp.SexpInt
		for _, arg := range args {
			if !glisp.IsInt(arg) {
				return glisp.SexpNull, fmt.Errorf("all arguments of %s must be int but got %v", name, glisp.InspectType(arg))
			}
			numbers = append(numbers, arg.(glisp.SexpInt))
		}
		switch len(args) {
		case 0:
			// (range)
			return newDefaultRange(), nil
		case 1:
			// (range end)
			return newRange(glisp.NewSexpInt(0), numbers[0], glisp.NewSexpInt(1)), nil
		case 2:
			// (range start end)
			return newRange(numbers[0], numbers[1], glisp.NewSexpInt(1)), nil
		case 3:
			// (range start end step)
			return newRange(numbers[0], numbers[1], numbers[2]), nil
		default:
			return glisp.WrongNumberArguments(name, len(args), 0, 1, 2, 3)
		}
	}
}

func StreamPartitionFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 && len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 2, 3)
		}
		if !IsStream(args[len(args)-1]) {
			return glisp.SexpNull, fmt.Errorf("last argument of %s must be stream, but got %v", name, glisp.InspectType(args[len(args)-1]))
		}
		stream := args[len(args)-1].(iStream)
		if len(args) == 2 {
			switch expr := args[0].(type) {
			case glisp.SexpInt:
				return &partitionIterator{iStream: stream, size: expr.ToInt()}, nil
			case *glisp.SexpFunction:
				return &partitionIterator{iStream: stream, f: expr}, nil
			}
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be int/function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !glisp.IsFunction(args[0]) {
			return glisp.SexpNull, fmt.Errorf(`first argument of %s must be function, but got %v`, name, glisp.InspectType(args[0]))
		}
		if !glisp.IsBool(args[1]) {
			return glisp.SexpNull, fmt.Errorf(`second argument of %s must be bool, but got %v`, name, glisp.InspectType(args[1]))
		}
		if bool(args[1].(glisp.SexpBool)) {
			return &partitionIterator{iStream: stream, f: args[0].(*glisp.SexpFunction), separatorPolicy: includeSepLeft}, nil
		}
		return &partitionIterator{iStream: stream, f: args[0].(*glisp.SexpFunction), separatorPolicy: includeSepRight}, nil
	}
}

func StreamZipFunction(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 2 {
			return glisp.WrongNumberArguments(name, len(args), 2, glisp.Many)
		}
		expr := make([]iStream, len(args))
		for i, stream := range args {
			if !IsStream(stream) {
				return glisp.SexpNull, fmt.Errorf("every argument of %s must be stream but %v-th is %v", name, i+1, glisp.InspectType(stream))
			}
			expr[i] = args[i].(iStream)
		}
		return &ZipListIterator{expr: expr, size: len(args)}, nil
	}
}

func StreamUnionFunction(name string) glisp.UserFunction {
	concat := glisp.GetConcatFunction(name)
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) < 2 {
			return glisp.WrongNumberArguments(name, len(args), 2, glisp.Many)
		}
		if !IsStream(args[0]) {
			return concat(env, args)
		}
		expr := make([]iStream, len(args))
		for i, stream := range args {
			if !IsStream(stream) {
				return glisp.SexpNull, fmt.Errorf("every argument of %s must be stream/streamable but %v-th is %v", name, i+1, glisp.InspectType(stream))
			} else {
				expr[i] = args[i].(iStream)
			}
		}
		return &UnionIterator{expr: expr}, nil
	}
}
