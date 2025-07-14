package glisp

type Address struct {
	function *SexpFunction
	position int
}

func (a *Address) IsStackElem() {}

func (stack *Stack) PushAddr(function *SexpFunction, pc int) {
	addr := getAddressFromPool()
	addr.function = function
	addr.position = pc
	stack.Push(addr)
}

func (stack *Stack) PopAddr() (fn *SexpFunction, pos int, err error) {
	elem, err := stack.Pop()
	if err != nil {
		return MissingFunction, 0, err
	}
	addr := elem.(*Address)
	fn = addr.function
	pos = addr.position
	recycleAddress(addr)
	return
}
