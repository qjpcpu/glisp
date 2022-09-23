package extensions

import (
	"errors"
	"fmt"
	"time"

	"github.com/qjpcpu/glisp"
)

func ImportTime(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("time/now", TimeNow)
	env.AddNamedFunction("time/format", TimeFormatFunction)
	env.AddNamedFunction("time/parse", ParseTime)
	env.AddNamedFunction("time/add-date", TimeAddDate)
	env.AddNamedFunction("time/add", TimeAdd)
	env.AddNamedFunction("time/sub", TimeSub)
	env.AddNamedFunction("time/year", TimeYearOf)
	env.AddNamedFunction("time/month", TimeMonthOf)
	env.AddNamedFunction("time/day", TimeDayOf)
	env.AddNamedFunction("time/hour", TimeHourOf)
	env.AddNamedFunction("time/minute", TimeMinuteOf)
	env.AddNamedFunction("time/second", TimeSecondOf)
	env.AddNamedFunction("time/weekday", TimeWeekdayOf)
	return nil
}

const (
	sym_timestamp       = `timestamp`
	sym_timestamp_ms    = `timestamp-ms`
	sym_timestamp_micro = `timestamp-micro`
	sym_timestamp_nano  = `timestamp-nano`
)

type SexpTime time.Time

func (t SexpTime) SexpString() string {
	return fmt.Sprintf(`(time/parse %v '%s)`, time.Time(t).UnixMilli(), sym_timestamp_ms)
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
		return 0, fmt.Errorf("cannot compare %s to %s", glisp.Inspect(t), glisp.Inspect(b))
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
		return SexpTime(time.Now().In(time.UTC)), nil
	}
}

func TimeYearOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Year())), nil
	}
}

func TimeMonthOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Month())), nil
	}
}

func TimeDayOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Day())), nil
	}
}

func TimeHourOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Hour())), nil
	}
}

func TimeMinuteOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Minute())), nil
	}
}

func TimeSecondOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Second())), nil
	}
}

func TimeWeekdayOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf("first argument of %s must be time", name)
		}
		tm := time.Time(stm)
		return glisp.NewSexpInt64(int64(tm.Weekday())), nil
	}
}

// TimeSub t1-t2 in seconds
func TimeSub(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		for i, v := range args {
			if _, ok := v.(SexpTime); !ok {
				return glisp.SexpNull, fmt.Errorf(`the %v argument of function %s must be time`, i+1, name)
			}
		}
		t1, t2 := time.Time(args[0].(SexpTime)), time.Time(args[1].(SexpTime))
		return glisp.NewSexpInt64(int64(t1.Sub(t2).Seconds())), nil
	}
}

// (time/add time number uint)
func TimeAdd(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 3)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`first argument of function %s must be time`, name)
		}
		if !glisp.IsInt(args[1]) {
			return glisp.SexpNull, fmt.Errorf(`third argument of function %s must be integer`, name)
		}
		number := args[1].(glisp.SexpInt).ToInt()
		kind, ok := readSymOrStr(args[2])
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`second argument of function %s must be string/symbol`, name)
		}
		tm := time.Time(stm)
		switch kind {
		case "year":
			return SexpTime(tm.AddDate(number, 0, 0)), nil
		case "month":
			return SexpTime(tm.AddDate(0, number, 0)), nil
		case "day":
			return SexpTime(tm.AddDate(0, 0, number)), nil
		case "hour":
			return SexpTime(tm.Add(time.Hour * time.Duration(number))), nil
		case "minute":
			return SexpTime(tm.Add(time.Minute * time.Duration(number))), nil
		case "second":
			return SexpTime(tm.Add(time.Second * time.Duration(number))), nil
		default:
			return glisp.SexpNull, fmt.Errorf("not support time add kind %s", kind)
		}
	}
}

// (time/add-date time year month day)
func TimeAddDate(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 4 {
			return glisp.WrongNumberArguments(name, len(args), 4)
		}
		stm, ok := args[0].(SexpTime)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`first argument of function %s must be time`, name)
		}
		for i := 1; i < len(args); i++ {
			if !glisp.IsInt(args[i]) {
				return glisp.SexpNull, fmt.Errorf(`the %v argument of function %s must be integer`, i+1, name)
			}
		}
		tm := time.Time(stm)
		return SexpTime(tm.AddDate(args[1].(glisp.SexpInt).ToInt(), args[2].(glisp.SexpInt).ToInt(), args[3].(glisp.SexpInt).ToInt())), nil
	}
}

func ParseTime(name string) glisp.UserFunction {
	layoutCandidates := []string{
		`2006-01-02 15:04:05`,
		time.RFC3339,
		`2006-01-02`,
		`15:04:05`,
	}
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		switch len(args) {
		case 1:
			arg := args[0]
			switch val := arg.(type) {
			case glisp.SexpInt:
				return SexpTime(time.Unix(arg.(glisp.SexpInt).ToInt64(), 0)), nil
			case glisp.SexpStr:
				err := fmt.Errorf("can't parse time `%s`", string(val))
				for _, layout := range layoutCandidates {
					if tm, err0 := time.ParseInLocation(layout, string(val), time.UTC); err0 == nil {
						return SexpTime(tm), nil
					} else {
						err = err0
					}
				}
				return glisp.SexpNull, err
			default:
				return glisp.SexpNull, fmt.Errorf(`%s with unsupported argument %v`, name, args[0].SexpString())
			}
		case 2, 3:
			if len(args) == 2 && glisp.IsInt(args[0]) && glisp.IsSymbol(args[1]) {
				switch args[1].(glisp.SexpSymbol).Name() {
				case sym_timestamp:
					tm := time.Unix(args[0].(glisp.SexpInt).ToInt64(), 0)
					return SexpTime(tm), nil
				case sym_timestamp_ms:
					tm := time.UnixMilli(args[0].(glisp.SexpInt).ToInt64())
					return SexpTime(tm), nil
				case sym_timestamp_micro:
					tm := time.UnixMicro(args[0].(glisp.SexpInt).ToInt64())
					return SexpTime(tm), nil
				case sym_timestamp_nano:
					number := args[0].(glisp.SexpInt)
					sec := number.Div(glisp.NewSexpUint64(1e9))
					nsec := number.Mod(glisp.NewSexpUint64(1e9))
					tm := time.Unix(sec.ToInt64(), nsec.ToInt64())
					return SexpTime(tm), nil
				}
			}
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
				parseTimeFn = func() (time.Time, error) { return time.ParseInLocation(layout, value, time.UTC) }
			}
			tm, err := parseTimeFn()
			if err != nil {
				return glisp.SexpNull, fmt.Errorf(`parse time %s with layout %s fail %v`, value, layout, err)
			}
			return SexpTime(tm), nil
		}
		return glisp.WrongNumberArguments(name, len(args), 1, 2)
	}
}

func TimeFormatFunction(fname string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 && len(args) != 3 {
			return glisp.SexpNull, glisp.WrongGeneratorNumberArguments(fname, len(args), 2, 3)
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
		if len(args) == 3 && !glisp.IsString(args[2]) {
			return glisp.SexpNull, fmt.Errorf(`third argument of function %s must be string but got %s`, fname, glisp.InspectType(args[2]))
		}
		tm := time.Time(stm)
		switch format {
		case sym_timestamp:
			return glisp.NewSexpInt64(tm.Unix()), nil
		case sym_timestamp_ms:
			return glisp.NewSexpInt64(tm.UnixMilli()), nil
		case sym_timestamp_micro:
			return glisp.NewSexpInt64(tm.UnixMicro()), nil
		case sym_timestamp_nano:
			return glisp.NewSexpInt64(tm.UnixNano()), nil
		case "":
			return glisp.SexpNull, errors.New(`blank time format symbol`)
		default:
			if len(args) == 3 {
				locStr := string(string(args[2].(glisp.SexpStr)))
				loc, err := time.LoadLocation(locStr)
				if err != nil {
					return glisp.SexpNull, fmt.Errorf("parse location `%s` fail %v", locStr, err)
				}
				return glisp.SexpStr(tm.In(loc).Format(format)), nil
			}
			return glisp.SexpStr(tm.UTC().Format(format)), nil
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
