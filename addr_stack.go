package glisp

import (
	"fmt"
)

type Address struct {
	function *SexpFunction
	position int
}

type AddrStack struct {
	tos      int
	elements []Address
}

func NewAddrStack(size int) *AddrStack {
	stack := new(AddrStack)
	stack.tos = -1
	if count := size - cap(stack.elements); count > 0 {
		stack.elements = append(stack.elements, make([]Address, count)...)
	}
	if count := size - len(stack.elements); count > 0 {
		stack.elements = stack.elements[:cap(stack.elements)]
	}
	return stack
}

func (stack *AddrStack) Clone() *AddrStack {
	ret := NewAddrStack(len(stack.elements))
	ret.tos = stack.tos
	for i := 0; i <= stack.tos; i++ {
		ret.elements[i] = stack.elements[i]
	}
	return ret
}

func (stack *AddrStack) Top() int {
	return stack.tos
}

func (stack *AddrStack) IsEmpty() bool {
	return stack.tos < 0
}

func (stack *AddrStack) Push(elem Address) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *AddrStack) Get(n int) (Address, error) {
	if stack.tos-n < 0 {
		return Address{}, fmt.Errorf("invalid stack access asked for %v Top was %v", n, stack.tos)
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *AddrStack) Pop() (Address, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return Address{}, err
	}
	stack.tos--
	return elem, nil
}

func (stack *AddrStack) PushAddr(function *SexpFunction, pc int) {
	stack.Push(Address{function: function, position: pc})
}

func (stack *AddrStack) PopAddr() (fn *SexpFunction, pos int, err error) {
	addr, err := stack.Pop()
	if err != nil {
		return MissingFunction, 0, err
	}
	fn = addr.function
	pos = addr.position
	return
}
