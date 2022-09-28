package glisp

import "sync"

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
	if cap(stack.elements) < size {
		stack.elements = make([]StackElem, size)
	}
	for i := len(stack.elements); i < size; i++ {
		stack.elements = append(stack.elements, nil)
	}
	return stack
}

func recycleStack(stack *Stack) {
	stack.tos = -1
	stack.elements = stack.elements[:0]
	stackPool.Put(stack)
}

var scopePool = sync.Pool{
	New: func() interface{} {
		return Scope(make(map[int]Sexp))
	},
}

func getScopeFromPool() Scope {
	return scopePool.Get().(Scope)
}

func recycleScope(s Scope) {
	for k := range s {
		delete(s, k)
	}
	scopePool.Put(s)
}
