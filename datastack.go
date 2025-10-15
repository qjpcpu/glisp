package glisp

import (
	"errors"
	"fmt"
	"io"
)

type DataStackElem struct {
	expr Sexp
}

func (d *DataStackElem) IsStackElem() {}

func (stack *Stack) PushExpr(expr Sexp) {
	elem := dataElemPool.Get().(*DataStackElem)
	elem.expr = expr
	stack.Push(elem)
}

func (stack *Stack) PopExpr() (Sexp, error) {
	elem, err := stack.Pop()
	if err != nil {
		return nil, err
	}
	data := elem.(*DataStackElem)
	expr := data.expr
	recycleDataElem(data)
	return expr, nil
}

func (stack *Stack) GetExpressions(n int) ([]Sexp, error) {
	return stack.getExpressions(n, false)
}

func (stack *Stack) getExpressions(n int, recycle bool) ([]Sexp, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return nil, errors.New("not enough items on stack")
	}
	arr := GetSlice(n)
	for i := 0; i < n; i++ {
		elem := stack.elements[stack_start+i].(*DataStackElem)
		arr[i] = elem.expr
		if recycle {
			recycleDataElem(elem)
		}
	}
	return arr, nil
}

func (stack *Stack) PeekArgs(n int) (Args, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return Args{}, errors.New("not enough items on stack")
	}
	return Args{
		len:      n,
		argsList: stack.elements[stack_start : stack_start+n],
	}, nil
}

func (stack *Stack) DropArgs(n int) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return
	}
	for i := 0; i < n; i++ {
		elem := stack.elements[stack_start+i].(*DataStackElem)
		recycleDataElem(elem)
	}
	stack.tos -= n
}

func (stack *Stack) PopExpressions(n int) ([]Sexp, error) {
	expressions, err := stack.getExpressions(n, true)
	if err != nil {
		return nil, err
	}
	stack.tos -= n
	return expressions, nil
}

func (stack *Stack) GetExpr(n int) (Sexp, error) {
	elem, err := stack.Get(n)
	if err != nil {
		return nil, err
	}
	return elem.(*DataStackElem).expr, nil
}

func (stack *Stack) PrintStack(w io.Writer) {
	fmt.Fprintf(w, "\t%d elements\n", stack.tos+1)
	for i := 0; i <= stack.tos; i++ {
		expr := stack.elements[i].(*DataStackElem).expr
		fmt.Fprintln(w, "\t"+expr.SexpString())
	}
}
