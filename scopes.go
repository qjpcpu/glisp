package glisp

import (
	"errors"
	"fmt"
)

type Scope map[int]Sexp

func (s Scope) IsStackElem() {}
func (s Scope) Clone() StackElem {
	newScope := make(Scope)
	for k, v := range s {
		newScope[k] = v
	}
	return newScope
}
func (s Scope) Find(n int) (Sexp, bool) {
	if expr, ok := s[n]; ok {
		return expr, true
	}
	return SexpNull, false
}

func (s Scope) Bind(n int, e Sexp) {
	s[n] = e
}

func (stack *Stack) PushScope() {
	stack.Push(make(Scope))
}

func (stack *Stack) PopScope() error {
	_, err := stack.Pop()
	return err
}

func (stack *Stack) lookupSymbol(sym SexpSymbol, minFrame int) (Sexp, Scope, error) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos-minFrame; i++ {
			elem, err := stack.Get(i)
			if err != nil {
				return SexpNull, nil, err
			}
			scope := elem.(Scope)
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
	stack.elements[stack.tos].(Scope).Bind(sym.number, expr)
	return nil
}

func (stack *Stack) SetSymbol(sym SexpSymbol, expr Sexp) error {
	if _, scope, err := stack.lookupSymbol(sym, 0); err == nil {
		scope.Bind(sym.number, expr)
		return nil
	}
	return stack.BindSymbol(sym, expr)
}
