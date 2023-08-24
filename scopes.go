package glisp

import (
	"errors"
	"fmt"
)

type ScopeElem struct {
	k int
	v Sexp
}

type Scope struct {
	vals []*ScopeElem
}

func (s *Scope) IsStackElem() {}
func (s *Scope) Clone() StackElem {
	newScope := getScopeFromPool()
	for _, v := range s.vals {
		if size := len(newScope.vals); size < cap(newScope.vals) {
			newScope.vals = newScope.vals[:size+1]
			if newScope.vals[size] == nil {
				newScope.vals[size] = &ScopeElem{}
			}
			newScope.vals[size].k = v.k
			newScope.vals[size].v = v.v
		} else {
			newScope.vals = append(newScope.vals, &ScopeElem{k: v.k, v: v.v})
		}
	}
	return newScope
}
func (s *Scope) Find(n int) (Sexp, bool) {
	if idx, ok := s.search(n); ok {
		return s.vals[idx].v, true
	}
	return SexpNull, false
}
func (s *Scope) Bind(n int, e Sexp) {
	if idx, ok := s.search(n); ok {
		s.vals[idx].v = e
	} else {
		if size := len(s.vals); size < cap(s.vals) {
			s.vals = s.vals[:size+1]
			if s.vals[size] == nil {
				s.vals[size] = &ScopeElem{}
			}
		} else {
			s.vals = append(s.vals, &ScopeElem{})
		}
		for i := len(s.vals) - 1; i > idx; i-- {
			s.vals[i].k = s.vals[i-1].k
			s.vals[i].v = s.vals[i-1].v
		}
		s.vals[idx] = &ScopeElem{k: n, v: e}
	}
}

func (s *Scope) search(target int) (int, bool) {
	start, end := 0, len(s.vals)
	for start < end {
		mid := (start + end) / 2
		if s.vals[mid].k < target {
			start = mid + 1
		} else {
			end = mid
		}
	}
	if start < len(s.vals) && s.vals[start].k == target {
		return start, true
	}
	return start, false
}

func (stack *Stack) PushScope() {
	stack.Push(getScopeFromPool())
}

func (stack *Stack) PopScope() error {
	s, err := stack.Pop()
	recycleScope(s.(*Scope))
	return err
}

func (stack *Stack) lookupSymbol(sym SexpSymbol, minFrame int) (Sexp, *Scope, error) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos-minFrame; i++ {
			elem, err := stack.Get(i)
			if err != nil {
				return SexpNull, nil, err
			}
			scope := elem.(*Scope)
			expr, ok := scope.Find(sym.number)
			if ok {
				return expr, scope, nil
			}
		}
	}
	return SexpNull, nil, fmt.Errorf("symbol `%v` not found", sym.Name())
}

func (stack *Stack) LookupSymbol(sym SexpSymbol) (Sexp, error) {
	expr, _, err := stack.lookupSymbol(sym, 0)
	return expr, err
}

func (stack *Stack) BindSymbol(sym SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		return errors.New("no scope available")
	}
	stack.elements[stack.tos].(*Scope).Bind(sym.number, expr)
	return nil
}

func (stack *Stack) SetSymbol(sym SexpSymbol, expr Sexp) error {
	if _, scope, err := stack.lookupSymbol(sym, 0); err == nil {
		scope.Bind(sym.number, expr)
		return nil
	}
	return stack.BindSymbol(sym, expr)
}
