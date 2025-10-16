package glisp

import (
	"fmt"
)

type StackStack struct {
	tos      int
	elements []*ScopeStack
}

func NewStackStack(size int) *StackStack {
	stack := new(StackStack)
	stack.tos = -1
	if count := size - cap(stack.elements); count > 0 {
		stack.elements = append(stack.elements, make([]*ScopeStack, count)...)
	}
	if count := size - len(stack.elements); count > 0 {
		stack.elements = stack.elements[:cap(stack.elements)]
	}
	return stack
}

func (stack *StackStack) Clone() *StackStack {
	ret := NewStackStack(len(stack.elements))
	ret.tos = stack.tos
	for i := 0; i <= stack.tos; i++ {
		ret.elements[i] = stack.elements[i]
	}
	return ret
}

func (stack *StackStack) Top() int {
	return stack.tos
}

func (stack *StackStack) IsEmpty() bool {
	return stack.tos < 0
}

func (stack *StackStack) Push(elem *ScopeStack) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *StackStack) Get(n int) (*ScopeStack, error) {
	if stack.tos-n < 0 {
		return nil, fmt.Errorf("invalid stack access asked for %v Top was %v", n, stack.tos)
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *StackStack) Pop() (*ScopeStack, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}
	stack.tos--
	return elem, nil
}
