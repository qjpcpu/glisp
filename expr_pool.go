package glisp

import (
	"math"
	"math/bits"
	"sync"
	"unsafe"
)

var builtinPool Pool

type Pool struct {
	pools [32]sync.Pool
}

func GetSlice(size int) []Sexp {
	return builtinPool.Get(size)
}

func PutSlice(buf []Sexp) {
	builtinPool.Put(buf)
}

func (p *Pool) Get(size int) []Sexp {
	if size <= 0 {
		return nil
	}
	if size > math.MaxInt32 {
		return make([]Sexp, size)
	}
	idx := index(uint32(size))
	ptr, _ := p.pools[idx].Get().(*Sexp)
	if ptr == nil {
		return make([]Sexp, size, 1<<idx)
	}
	return unsafe.Slice(ptr, 1<<idx)[:size]
}

func (p *Pool) Put(buf []Sexp) {
	size := cap(buf)
	if size == 0 || size > math.MaxInt32 {
		return
	}
	idx := index(uint32(size))
	if size != 1<<idx { // this Sexp slice is not from Pool.Get(), put it into the previous interval of idx
		idx--
	}
	// Store the pointer to the underlying array instead of the pointer to the slice itself,
	// which circumvents the escape of buf from the stack to the heap.
	p.pools[idx].Put(unsafe.SliceData(buf))
}

func index(n uint32) uint32 {
	return uint32(bits.Len32(n - 1))
}
