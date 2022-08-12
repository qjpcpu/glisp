package glisp

import (
	"fmt"
	"strconv"
	"strings"
)

func SearchSexp(name string) UserFunction {
	return func(env *Environment, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return WrongNumberArguments(name, len(args), 2)
		}
		if args[0] == SexpNull {
			return SexpNull, nil
		}
		if !IsHash(args[0]) && !IsArray(args[0]) {
			return SexpNull, fmt.Errorf("first argument of %s must be hash/array", name)
		}
		if !IsString(args[1]) {
			return SexpNull, fmt.Errorf("second argument of %s must be string", name)
		}
		p, ok := makeStPath(string(args[1].(SexpStr)))
		if !ok {
			return SexpNull, nil
		}
		return findSexp(args[0], p), nil
	}
}

const (
	sharpSym               = "#"
	arrayElemEq            = "=="
	arrayElemContains      = "="
	arrayElemNotEq         = "!=="
	arrayElemNotContains   = "!="
	arrayElemGreaterThan   = ">"
	arrayElemGreaterEqThan = ">="
	arrayElemLessThan      = "<"
	arrayElemLessEqThan    = "<="
)

type stPath struct {
	Name     string
	Selector string
	Op       string
	Val      string
}

func (sp stPath) isArrayElemSelector() bool {
	return sp.Name == "#"
}

func (sp stPath) isInteger() bool {
	_, err := strconv.ParseUint(sp.Name, 10, 64)
	return err == nil
}

func (sp stPath) asInteger() int {
	i, _ := strconv.ParseUint(sp.Name, 10, 64)
	return int(i)
}

func findSexp(node Sexp, paths []stPath) Sexp {
	if len(paths) == 0 {
		return node
	}
	if node == nil || node == SexpNull {
		return SexpNull
	}
	p := paths[0]
	switch expr := node.(type) {
	case *SexpHash:
		keys := expr.KeyOrder
		for _, key := range keys {
			if sexprToStr(key) == p.Name {
				val, _ := expr.HashGet(key)
				return findSexp(val, paths[1:])
			}
		}
	case SexpArray:
		if p.isArrayElemSelector() {
			var list SexpArray
			fromList := expr
			if p.Op != "" {
				fromList = filterArrayNodeBySelector(expr, p)
			}
			for _, n := range fromList {
				if out := findSexp(n, paths[1:]); out != nil && out != SexpNull {
					list = append(list, out)
				}
			}
			return list
		} else if p.isInteger() && p.asInteger() < len(expr) {
			return findSexp(expr[p.asInteger()], paths[1:])
		}
	}
	return SexpNull
}

func filterArrayNodeBySelector(nodes SexpArray, path stPath) SexpArray {
	paths, ok := makeStPath(path.Selector)
	if !ok {
		return nil
	}
	val := strings.TrimSuffix(strings.TrimPrefix(path.Val, `"`), `"`)
	var list SexpArray
	for _, n := range nodes {
		if out := findSexp(n, paths); out != nil {
			if isElemMatched(out, path.Op, val) {
				list = append(list, n)
			}
		}
	}
	return list
}

func isElemMatched(n Sexp, op string, val string) bool {
	switch n.(type) {
	case SexpArray:
	case *SexpHash:
	default:
		v := sexprToStr(n)
		switch op {
		case arrayElemContains:
			return strings.Contains(v, val)
		case arrayElemEq:
			return v == val
		case arrayElemNotEq:
			return v != val
		case arrayElemNotContains:
			return !strings.Contains(v, val)
		case arrayElemGreaterThan:
			if IsNumber(n) {
				if strings.Contains(val, ".") || strings.Contains(v, ".") {
					f, _ := NewSexpFloatStr(val)
					r, _ := Compare(n, f)
					return r > 0
				} else {
					f, _ := NewSexpIntStr(val)
					r, _ := Compare(n, f)
					return r > 0
				}
			} else {
				return v > val
			}
		case arrayElemGreaterEqThan:
			if IsNumber(n) {
				if strings.Contains(val, ".") || strings.Contains(v, ".") {
					f, _ := NewSexpFloatStr(val)
					r, _ := Compare(n, f)
					return r >= 0
				} else {
					f, _ := NewSexpIntStr(val)
					r, _ := Compare(n, f)
					return r >= 0
				}
			} else {
				return v >= val
			}
		case arrayElemLessEqThan:
			if IsNumber(n) {
				if strings.Contains(val, ".") || strings.Contains(v, ".") {
					f, _ := NewSexpFloatStr(val)
					r, _ := Compare(n, f)
					return r <= 0
				} else {
					f, _ := NewSexpIntStr(val)
					r, _ := Compare(n, f)
					return r <= 0
				}
			} else {
				return v <= val
			}
		case arrayElemLessThan:
			if IsNumber(n) {
				if strings.Contains(val, ".") || strings.Contains(v, ".") {
					f, _ := NewSexpFloatStr(val)
					r, _ := Compare(n, f)
					return r < 0
				} else {
					f, _ := NewSexpIntStr(val)
					r, _ := Compare(n, f)
					return r < 0
				}
			} else {
				return v < val
			}
		}
	}
	return false
}

func makeStPath(p string) ([]stPath, bool) {
	var paths []stPath
	proj := map[byte]byte{
		'(': ')',
		'"': '"',
	}
	data := []byte(strings.TrimPrefix(p, "."))
	var start int
	for i := 0; i < len(data); {
		if data[i] == '\\' {
			if i+1 < len(data) && data[i+1] == '.' {
				data[i] = 0
			}
			i += 2
			continue
		} else if data[i] == '#' && i+1 < len(data) && data[i+1] == '(' {
			if closeIdx := findCloseSym(data, i+2, len(data), '(', proj); closeIdx == -1 {
				return nil, false
			} else {
				i = closeIdx + 1
				if closeIdx == len(data)-1 {
					paths = append(paths, stPath{Name: removeByte(string(data[start:]), 0)})
				}
				continue
			}
		} else if data[i] == '.' && i > start {
			paths = append(paths, stPath{Name: removeByte(string(data[start:i]), 0)})
			start = i + 1
		} else if i == len(data)-1 {
			paths = append(paths, stPath{Name: removeByte(string(data[start:]), 0)})
			start = i + 1
		}
		i++
	}
	for i, path := range paths {
		paths[i] = reformatStPath(path)
	}
	return paths, true
}

func reformatStPath(p stPath) stPath {
	if len(p.Name) > 3 && p.Name[0] == '#' && p.Name[1] == '(' && p.Name[len(p.Name)-1] == ')' {
		step := p.Name[2 : len(p.Name)-1]
		p.Name = "#"
		p.Selector, p.Op, p.Val = reformatStStep(step)
	}
	return p
}

func reformatStStep(step string) (selector string, op string, val string) {
	if idx := strings.Index(step, arrayElemNotEq); idx >= 0 {
		selector = strings.TrimSpace(step[:idx])
		op = arrayElemNotEq
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemNotContains); idx >= 0 {
		selector = strings.TrimSpace(step[:idx])
		op = arrayElemNotContains
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemEq); idx >= 0 {
		selector = strings.TrimSpace(step[:idx])
		op = arrayElemEq
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemGreaterEqThan); idx >= 0 {
		selector = strings.TrimSpace(step[:idx])
		op = arrayElemGreaterEqThan
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemLessEqThan); idx >= 0 {
		selector = strings.TrimSpace(step[:idx])
		op = arrayElemLessEqThan
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemContains); idx >= 0 {
		op = arrayElemContains
		selector = strings.TrimSpace(step[:idx])
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemLessThan); idx >= 0 {
		op = arrayElemLessThan
		selector = strings.TrimSpace(step[:idx])
		val = strings.TrimSpace(step[idx+len(op):])
	} else if idx = strings.Index(step, arrayElemGreaterThan); idx >= 0 {
		op = arrayElemGreaterThan
		selector = strings.TrimSpace(step[:idx])
		val = strings.TrimSpace(step[idx+len(op):])
	}
	return
}

func removeByte(s string, b byte) string {
	var cnt int
	data := []byte(s)
	for i := 0; i < len(data); i++ {
		if data[i] == b {
			cnt++
		} else if cnt > 0 {
			data[i-cnt] = data[i]
		}
	}
	return string(data[:len(data)-cnt])
}

func findCloseSym(data []byte, from, to int, openSym byte, proj map[byte]byte) int {
	closeSym, ok := proj[openSym]
	if !ok {
		return -1
	}
	for i := from; i < to && i < len(data); {
		if data[i] == '\\' {
			i += 2
			continue
		}
		if data[i] == closeSym {
			return i
		}
		if _, ok := proj[data[i]]; ok {
			if nextCloseIndex := findCloseSym(data, i+1, to, data[i], proj); nextCloseIndex == -1 {
				return -1
			} else {
				i = nextCloseIndex + 1
				continue
			}
		}
		i++
	}
	return -1
}

func sexprToStr(expr Sexp) string {
	switch s := expr.(type) {
	case SexpStr:
		return string(s)
	default:
		return expr.SexpString()
	}
}
