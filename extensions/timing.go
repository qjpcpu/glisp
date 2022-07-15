package extensions

import (
	"errors"
	"fmt"
	"time"

	"github.com/qjpcpu/glisp"
)

type SexpTime time.Time

func (t SexpTime) SexpString() string {
	return time.Time(t).String()
}

func (t SexpTime) TypeName() string {
	return `time`
}

func (t SexpTime) MarshalJSON() ([]byte, error) {
	tm := (time.Time)(t)
	return glisp.Marshal(glisp.SexpStr(tm.Format(`2006-01-02 15:04:05`)))
}

func (t SexpTime) Cmp(b glisp.Comparable) (int, error) {
	if _, ok := b.(SexpTime); !ok {
		return 0, fmt.Errorf("cannot compare %T(%s) to %T(%s)", t, t.SexpString(), b, b.SexpString())
	}
	t1 := time.Time(t)
	t2 := time.Time(b.(SexpTime))
	if t1.Equal(t2) {
		return 0, nil
	} else if t1.Before(t2) {
		return -1, nil
	} else {
		return 1, nil
	}
}

func TimeNow(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 0 {
			return glisp.WrongNumberArguments(name, len(args), 0)
		}
		return SexpTime(time.Now()), nil
	}
}

/*
  (time/parse 1655967400280) => parse unix timestamp to SexpTime
  (time/parse "2015-02-23 23:54:55") => parse time by value, use default layout 2006-01-02 15:04:05
  (time/parse "2006-Jan-02" "2014-Feb-04") => parse time by layout and value
  (time/parse "2006-Jan-02" "2014-Feb-04") => parse time by layout and value
  (time/parse "2006-Jan-02" "2014-Feb-04" "Asia/Shanghai") => parse time by layout and value and location
*/
func ParseTime(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		switch len(args) {
		case 1:
			arg := args[0]
			switch val := arg.(type) {
			case glisp.SexpInt:
				return SexpTime(time.Unix(arg.(glisp.SexpInt).ToInt64(), 0)), nil
			case glisp.SexpStr:
				tm, err := time.Parse(`2006-01-02 15:04:05`, string(val))
				if err != nil {
					return glisp.SexpNull, err
				}
				return SexpTime(tm), nil
			default:
				return glisp.SexpNull, fmt.Errorf(`%s with unsupported argument %v`, name, args[0].SexpString())
			}
		case 2, 3:
			layout, ok := readSymOrStr(args[0])
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`%s with unsupported argument %v`, name, args[0].SexpString())
			}
			value, ok := readSymOrStr(args[1])
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`%s with unsupported argument %v`, name, args[0].SexpString())
			}
			var parseTimeFn func() (time.Time, error)
			if len(args) == 3 {
				loc, ok := readSymOrStr(args[2])
				if !ok {
					return glisp.SexpNull, fmt.Errorf(`%s with unsupported argument %v`, name, args[0].SexpString())
				}
				location, err := time.LoadLocation(loc)
				if err != nil {
					return glisp.SexpNull, err
				}
				parseTimeFn = func() (time.Time, error) { return time.ParseInLocation(layout, value, location) }
			} else {
				parseTimeFn = func() (time.Time, error) { return time.Parse(layout, value) }
			}
			tm, err := parseTimeFn()
			if err != nil {
				return glisp.SexpNull, err
			}
			return SexpTime(tm), nil
		}
		return glisp.WrongNumberArguments(name, len(args), 1, 2)
	}
}

/*
  (time/format SexpTime 'timestamp) => SexpTime to unix timestamp
  (time/format SexpTime 'timestamp-ms) => SexpTime to unix timestamp mills
  (time/format SexpTime "2006-01-02 15:04:05") => SexpTime to string by layout
*/
func GetTimeFormatFunction(fname string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, fmt.Errorf(`wrong argument number of function %s`, fname)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`first argument of function %s must be time`, fname)
		}
		var format string
		if sym, ok := args[1].(glisp.SexpSymbol); ok {
			format = sym.Name()
		} else if layout, ok := args[1].(glisp.SexpStr); ok {
			format = string(layout)
		} else {
			return glisp.SexpNull, fmt.Errorf(`second argument of function %s must be symbol/string`, fname)
		}
		tm := time.Time(stm)
		switch format {
		case "timestamp":
			return glisp.NewSexpInt64(tm.Unix()), nil
		case "timestamp-ms":
			return glisp.NewSexpInt64(tm.UnixMilli()), nil
		case "":
			return glisp.SexpNull, errors.New(`blank time format symbol`)
		default:
			return glisp.SexpStr(tm.Format(format)), nil
		}
	}
}

func readSymOrStr(s glisp.Sexp) (string, bool) {
	switch s.(type) {
	case glisp.SexpSymbol:
		return s.(glisp.SexpSymbol).Name(), true
	case glisp.SexpStr:
		return string(s.(glisp.SexpStr)), true
	}
	return "", false
}

func ImportTime(env *glisp.Environment) {
	env.AddFunctionByConstructor("time/now", TimeNow)
	env.AddFunctionByConstructor("time/format", GetTimeFormatFunction)
	env.AddFunctionByConstructor("time/parse", ParseTime)
}
