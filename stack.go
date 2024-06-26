package glisp

import (
	"fmt"
)

type Clonable interface {
	StackElem
	Clone() StackElem
}

type StackElem interface {
	IsStackElem()
}

type Stack struct {
	tos      int
	elements []StackElem
}

func NewStack(size int) *Stack {
	return getStackFromPool(size)
}

func (stack *Stack) Clone() *Stack {
	ret := getStackFromPool(len(stack.elements))
	ret.tos = stack.tos
	for i := 0; i <= stack.tos; i++ {
		if cb, ok := stack.elements[i].(Clonable); ok {
			ret.elements[i] = cb.Clone()
		} else {
			ret.elements[i] = stack.elements[i]
		}
	}

	return ret
}

func (stack *Stack) Top() int {
	return stack.tos
}

func (stack *Stack) PushAllTo(target *Stack) int {
	if stack.tos < 0 {
		return 0
	}

	for _, v := range stack.elements[0 : stack.tos+1] {
		target.Push(v)
	}

	return stack.tos + 1
}

func (stack *Stack) IsEmpty() bool {
	return stack.tos < 0
}

func (stack *Stack) PushMulti(elems ...StackElem) {
	for i := range elems {
		stack.Push(elems[i])
	}
}

func (stack *Stack) Push(elem StackElem) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *Stack) Get(n int) (StackElem, error) {
	if stack.tos-n < 0 {
		return nil, fmt.Errorf("invalid stack access asked for %v Top was %v", n, stack.tos)
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *Stack) Pop() (StackElem, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}
	stack.tos--
	return elem, nil
}

func (stack *Stack) IsStackElem() {}
