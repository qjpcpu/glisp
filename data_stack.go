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

// PeekArgsUntil peeks at arguments from the top of the stack downwards until
// the predicate `isEnd` returns true.
//
// It returns an `Args` object containing all elements from the top of the stack
// down to and including the element that satisfied the predicate.
//
// Key characteristics:
//
//  1. Inclusion of the End Element: The returned `Args` slice *includes* the
//     element for which `isEnd` returned true.
//
//  2. Argument Order: The elements in the returned `Args` are ordered from the
//     deepest element (the one that matched `isEnd`) to the shallowest (the one
//     at the very top of the stack). For example, `args.Get(0)` will be the
//     element that satisfied `isEnd`. This is the natural order as they appear
//     in the underlying `elements` array, but is the reverse of the LIFO (Last-In, First-Out)
//     order you would get from calling `Pop()` repeatedly.
//
//  3. Discarding the End Element: Since the end element is always at index 0
//     of the returned `Args`, you can easily get a slice of the arguments
//     without the end element by calling `args.SliceStart(1)`. This is a common
//     pattern, as seen in the implementation of `OpVectorize`.
func (stack *DataStack) PeekArgsUntil(isEnd func(Sexp) bool) (Args, error) {
	for i := 1; ; i++ {
		elem, err := stack.Get(i - 1)
		if err != nil {
			return Args{}, err
		}
		if isEnd(elem) {
			return stack.PeekArgs(i)
		}
	}
}

func (stack *DataStack) PeekArgs(n int) (Args, error) {
	if n <= 0 {
		return Args{}, nil
	}
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
