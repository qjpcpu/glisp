package glisp

import (
	"sync"
)

var stackPool = sync.Pool{
	New: func() interface{} {
		stack := new(Stack)
		stack.tos = -1
		return stack
	},
}

func getStackFromPool(size int) *Stack {
	stack := stackPool.Get().(*Stack)
	stack.tos = -1
	if count := size - cap(stack.elements); count > 0 {
		stack.elements = append(stack.elements, make([]StackElem, count)...)
	}
	if count := size - len(stack.elements); count > 0 {
		stack.elements = stack.elements[:cap(stack.elements)]
	}
	return stack
}

func recycleStack(stack *Stack) {
	stack.tos = -1
	stack.elements = stack.elements[:0]
	stackPool.Put(stack)
}

var dataElemPool = sync.Pool{
	New: func() interface{} {
		return &DataStackElem{}
	},
}

func recycleDataElem(e *DataStackElem) {
	e.expr = nil
	dataElemPool.Put(e)
}

var addressPool = sync.Pool{
	New: func() interface{} {
		return &Address{}
	},
}

func getAddressFromPool() *Address {
	addr := addressPool.Get().(*Address)
	return addr
}

func recycleAddress(addr *Address) {
	addr.function = nil
	addressPool.Put(addr)
}
