package extensions

import (
	"fmt"
	"os"

	"encoding/csv"

	"github.com/qjpcpu/glisp"
)

func ImportCSV(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("csv/read-file", ReadCSVFile)
	env.AddNamedFunction("csv/write-file", WriteCSVFile)
	return nil
}

func ReadCSVFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1,2 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s 1st argument should be string but got %v`, name, glisp.InspectType(args[0]))
		}
		filename := replaceHomeDirSymbol(string(str))
		fd, err := os.Open(filename)
		if err != nil {
			return glisp.SexpNull, err
		}
		defer fd.Close()
		reader := csv.NewReader(fd)
		records, err := reader.ReadAll()
		if err != nil {
			return glisp.SexpNull, err
		}
		var rows glisp.SexpArray
		if len(args) == 2 && glisp.IsSymbol(args[1]) && args[1].SexpString() == "hash" {
			var header glisp.SexpArray
			for i, row := range records {
				if i == 0 {
					for _, col := range row {
						header = append(header, glisp.SexpStr(col))
					}
				} else {
					var sexpRow []glisp.Sexp
					for j, col := range row {
						sexpRow = append(sexpRow,
							header[j],
							glisp.SexpStr(col),
						)
					}
					h, _ := glisp.MakeHash(sexpRow)
					rows = append(rows, h)
				}
			}
		} else {
			for _, row := range records {
				var sexpRow glisp.SexpArray
				for _, col := range row {
					sexpRow = append(sexpRow, glisp.SexpStr(col))
				}
				rows = append(rows, sexpRow)
			}
		}
		return rows, nil
	}
}

func WriteCSVFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1,2 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s 1st argument should be string but got %v`, name, glisp.InspectType(args[0]))
		}
		filename := replaceHomeDirSymbol(string(str))
		var records [][]string
		if recs, ok := args[1].(glisp.SexpArray); ok {
			var err error
			if records, err = extractCSVRows(recs); err != nil {
				return glisp.SexpNull, err
			}
		} else if args[1] == glisp.SexpNull {
		} else {
			return glisp.SexpNull, fmt.Errorf(`%s 2nd argument should be [][]string but got %v`, name, glisp.InspectType(args[1]))
		}
		fd, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return glisp.SexpNull, err
		}
		defer fd.Close()
		writer := csv.NewWriter(fd)
		writer.WriteAll(records)
		return glisp.SexpNull, nil
	}
}

func extractCSVRows(rows glisp.SexpArray) (ret [][]string, err error) {
	if len(rows) == 0 {
		return
	}
	if glisp.IsHash(rows[0]) {
		var columns []glisp.Sexp
		for i, row := range rows {
			hash, ok := row.(*glisp.SexpHash)
			if !ok {
				err = fmt.Errorf("csv row should be hash but got %v", glisp.InspectType(row))
				return
			}
			if i == 0 {
				var cols, vals []string
				hash.Foreach(func(k glisp.Sexp, v glisp.Sexp) bool {
					columns = append(columns, k)
					cols = append(cols, toCSVString(k))
					vals = append(vals, toCSVString(v))
					return true
				})
				ret = append(ret, cols, vals)
			} else {
				var vals []string
				for _, k := range columns {
					v, _ := hash.HashGetDefault(k, glisp.SexpStr(""))
					vals = append(vals, toCSVString(v))
				}
				ret = append(ret, vals)
			}
		}
		return
	}
	for _, row := range rows {
		arr := row.(glisp.SexpArray)
		var vals []string
		for _, col := range arr {
			vals = append(vals, toCSVString(col))
		}
		ret = append(ret, vals)
	}
	return
}

func toCSVString(s glisp.Sexp) string {
	switch val := s.(type) {
	case glisp.SexpStr:
		return string(val)
	case *glisp.SexpHash, glisp.SexpArray:
		v, _ := glisp.Marshal(val)
		return string(v)
	}
	return s.SexpString()
}
