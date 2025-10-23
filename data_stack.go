package glisp

import (
	"errors"
	"fmt"
	"io"
)

type DataStack struct {
	tos      int
	elements []Sexp
}

func NewDataStack(size int) *DataStack {
	stack := new(DataStack)
	stack.tos = -1
	if count := size - cap(stack.elements); count > 0 {
		stack.elements = append(stack.elements, make([]Sexp, count)...)
	}
	if count := size - len(stack.elements); count > 0 {
		stack.elements = stack.elements[:cap(stack.elements)]
	}
	return stack
}

func (stack *DataStack) Clone() *DataStack {
	ret := NewDataStack(len(stack.elements))
	ret.tos = stack.tos
	for i := 0; i <= stack.tos; i++ {
		ret.elements[i] = stack.elements[i]
	}
	return ret
}

func (stack *DataStack) Top() int {
	return stack.tos
}

func (stack *DataStack) IsEmpty() bool {
	return stack.tos < 0
}

func (stack *DataStack) Push(elem Sexp) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *DataStack) Get(n int) (Sexp, error) {
	if stack.tos-n < 0 {
		return nil, fmt.Errorf("invalid stack access asked for %v Top was %v", n, stack.tos)
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *DataStack) Pop() (Sexp, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}
	stack.tos--
	return elem, nil
}

func (stack *DataStack) PushExpr(expr Sexp) {
	stack.Push(expr)
}

func (stack *DataStack) PopExpr() (Sexp, error) {
	return stack.Pop()
}

func (stack *DataStack) GetExpressions(n int) ([]Sexp, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return nil, errors.New("not enough items on stack")
	}
	return stack.elements[stack_start : stack_start+n], nil
}

func (stack *DataStack) PeekArgs(n int) (Args, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return Args{}, errors.New("not enough items on stack")
	}
	return Args{
		len:      n,
		argsList: stack.elements[stack_start : stack_start+n],
	}, nil
}

func (stack *DataStack) DropExpr(n int) {
	if n <= 0 {
		return
	}
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return
	}
	stack.tos -= n
}

func (stack *DataStack) PopExpressions(n int) ([]Sexp, error) {
	expressions, err := stack.GetExpressions(n)
	if err != nil {
		return nil, err
	}
	stack.tos -= n
	return expressions, nil
}

func (stack *DataStack) GetExpr(n int) (Sexp, error) {
	return stack.Get(n)
}

func (stack *DataStack) PrintStack(w io.Writer) {
	fmt.Fprintf(w, "\t%d elements\n", stack.tos+1)
	for i := 0; i <= stack.tos; i++ {
		expr := stack.elements[i]
		fmt.Fprintln(w, "\t"+expr.SexpString())
	}
}
