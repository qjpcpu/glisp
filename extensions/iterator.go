package extensions

import (
	"fmt"
	"strings"

	"github.com/qjpcpu/glisp"
)

type Iterable interface {
	glisp.Sexp
	Next() (glisp.Sexp, bool)
}

type iStream interface {
	glisp.Sexp
	Next(*glisp.Environment) (glisp.Sexp, bool, error)
}

var (
	_ iStream = &ListIterator{}
	_ iStream = &ZipListIterator{}
	_ iStream = &ArrayIterator{}
	_ iStream = &BytesIterator{}
	_ iStream = &StringIterator{}
	_ iStream = &IterableStream{}
	_ iStream = &mapIterator{}
	_ iStream = &flatmapIterator{}
	_ iStream = &filterIterator{}
	_ iStream = &takeIterator{}
	_ iStream = &dropIterator{}
	_ iStream = &HashIterator{}
	_ iStream = &RangeIterator{}
	_ iStream = &partitionIterator{}
	_ iStream = &UnionIterator{}
)

type IterableStream struct {
	expr Iterable
}

func (iter *IterableStream) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, iter.expr.SexpString())
}

func (iter *IterableStream) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	expr, ok := iter.expr.Next()
	return expr, ok, nil
}

type ListIterator struct {
	expr glisp.Sexp
}

func (iter *ListIterator) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, iter.expr.SexpString())
}

func (iter *ListIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.expr == glisp.SexpNull {
		return glisp.SexpNull, false, nil
	}
	pair, ok := iter.expr.(*glisp.SexpPair)
	if !ok {
		expr := iter.expr
		iter.expr = glisp.SexpNull
		return expr, true, nil
	}
	iter.expr = pair.Tail()
	return pair.Head(), true, nil
}

type ArrayIterator struct {
	expr glisp.SexpArray
	idx  int
}

func (iter *ArrayIterator) SexpString() string {
	arr := iter.expr[iter.idx:]
	return fmt.Sprintf(`(stream %s)`, arr.SexpString())
}

func (iter *ArrayIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.idx >= len(iter.expr) {
		return glisp.SexpNull, false, nil
	}
	iter.idx++
	return iter.expr[iter.idx-1], true, nil
}

type BytesIterator struct {
	expr glisp.SexpBytes
	idx  int
}

func (iter *BytesIterator) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, glisp.NewSexpBytes(iter.expr.Bytes()[iter.idx:]).SexpString())
}

func (iter *BytesIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.idx >= len(iter.expr.Bytes()) {
		return glisp.SexpNull, false, nil
	}
	iter.idx++
	return glisp.SexpChar(iter.expr.Bytes()[iter.idx-1]), true, nil
}

type StringIterator struct {
	expr []rune
	idx  int
}

func (iter *StringIterator) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, glisp.SexpStr(iter.expr[iter.idx:]).SexpString())
}

func (iter *StringIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.idx >= len(iter.expr) {
		return glisp.SexpNull, false, nil
	}
	iter.idx++
	return glisp.SexpChar(iter.expr[iter.idx-1]), true, nil
}

type mapIterator struct {
	iStream
	f *glisp.SexpFunction
}

func (iter *mapIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	elem, ok, err := iter.iStream.Next(env)
	if err != nil || !ok {
		return glisp.SexpNull, false, err
	}
	ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
	return ret, true, err
}

type flatmapIterator struct {
	iStream
	f     *glisp.SexpFunction
	inner iStream
}

func (iter *flatmapIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
START:
	if iter.inner == nil {
		elem, ok, err := iter.iStream.Next(env)
		if err != nil || !ok {
			return glisp.SexpNull, false, err
		}
		ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
		if err != nil {
			return glisp.SexpNull, false, err
		}
		if IsStream(ret) {
			iter.inner = ret.(iStream)
		} else if IsStreamable(ret) {
			iter.inner = expr2Stream(ret)
		} else {
			return glisp.SexpNull, false, fmt.Errorf("flatmap element(%s) is not streamable", glisp.InspectType(ret))
		}
	}
	elem, ok, err := iter.inner.Next(env)
	if err != nil {
		return glisp.SexpNull, false, err
	}
	if !ok {
		iter.inner = nil
		goto START
	}
	return elem, true, nil
}

type filterIterator struct {
	iStream
	f *glisp.SexpFunction
}

func (iter *filterIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	for {
		elem, ok, err := iter.iStream.Next(env)
		if err != nil || !ok {
			return glisp.SexpNull, false, err
		}
		ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
		if err != nil {
			return glisp.SexpNull, false, err
		}
		if !glisp.IsBool(ret) {
			return glisp.SexpNull, false, fmt.Errorf("filter function should return bool but got %s", glisp.InspectType(ret))
		}
		if bool(ret.(glisp.SexpBool)) {
			return elem, true, nil
		}
	}
}

type takeIterator struct {
	iStream
	count uint64
	f     *glisp.SexpFunction
}

func (iter *takeIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.f != nil {
		elem, ok, err := iter.iStream.Next(env)
		if err != nil || !ok {
			return glisp.SexpNull, false, err
		}
		ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
		if err != nil {
			return glisp.SexpNull, false, err
		}
		if !glisp.IsBool(ret) {
			return glisp.SexpNull, false, fmt.Errorf("take function should return bool but got %s", glisp.InspectType(ret))
		}
		return elem, bool(ret.(glisp.SexpBool)), nil
	}
	if iter.count == 0 {
		return glisp.SexpNull, false, nil
	}
	iter.count--
	return iter.iStream.Next(env)
}

type dropIterator struct {
	iStream
	count uint64
	f     *glisp.SexpFunction
}

func (iter *dropIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.f != nil {
		for {
			elem, ok, err := iter.iStream.Next(env)
			if err != nil || !ok {
				return glisp.SexpNull, false, err
			}
			ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
			if err != nil {
				return glisp.SexpNull, false, err
			}
			if !glisp.IsBool(ret) {
				return glisp.SexpNull, false, fmt.Errorf("drop function should return bool but got %s", glisp.InspectType(ret))
			}
			if bool(ret.(glisp.SexpBool)) {
				continue
			}
			return elem, true, nil
		}
	}
	for ; iter.count > 0; iter.count-- {
		_, ok, err := iter.iStream.Next(env)
		if err != nil || !ok {
			return glisp.SexpNull, false, err
		}
	}
	return iter.iStream.Next(env)
}

type HashIterator struct {
	expr *glisp.SexpHash
	idx  int
}

func (iter *HashIterator) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, iter.expr.SexpString())
}

func (iter *HashIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.idx >= len(iter.expr.KeyOrder) {
		return glisp.SexpNull, false, nil
	}
	iter.idx++
	key := iter.expr.KeyOrder[iter.idx-1]
	if iter.expr.HashExist(key) {
		val, err := iter.expr.HashGet(key)
		if err != nil {
			return glisp.SexpNull, false, err
		}
		return glisp.Cons(key, val), true, nil
	}
	return glisp.SexpNull, false, nil
}

type RecordIterator struct {
	expr SexpRecord
	idx  int
}

func (iter *RecordIterator) SexpString() string {
	return fmt.Sprintf(`(stream %s)`, iter.expr.SexpString())
}

func (iter *RecordIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	fs := iter.expr.Class().Fields()
	if iter.idx >= len(fs) {
		return glisp.SexpNull, false, nil
	}
	key := glisp.SexpStr(fs[iter.idx].Name)
	val := iter.expr.GetFieldDefault(fs[iter.idx].Name, glisp.SexpNull)
	iter.idx++
	return glisp.Cons(key, val), true, nil
}

type RangeIterator struct {
	isDefault      bool
	from, to, step glisp.SexpInt
}

func newDefaultRange() *RangeIterator {
	return &RangeIterator{isDefault: true, from: glisp.NewSexpInt(0), step: glisp.NewSexpInt(1)}
}

func newRange(from glisp.SexpInt, to glisp.SexpInt, step glisp.SexpInt) *RangeIterator {
	return &RangeIterator{from: from, to: to, step: step}
}

func (iter *RangeIterator) SexpString() string {
	if iter.isDefault {
		return "(range)"
	}
	return fmt.Sprintf(`(range %v %v %v)`, iter.from.SexpString(), iter.to.SexpString(), iter.step.SexpString())
}

func (iter *RangeIterator) Next(*glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.isDefault {
		ret := iter.from
		iter.from = iter.from.Add(iter.step)
		return ret, true, nil
	}
	if r, _ := glisp.Compare(iter.from, iter.to); r < 0 {
		ret := iter.from
		iter.from = iter.from.Add(iter.step)
		return ret, true, nil
	}
	return iter.from, false, nil
}

type includePartitionSeparatorPolicy int

const (
	excludeSep      includePartitionSeparatorPolicy = 0
	includeSepLeft  includePartitionSeparatorPolicy = 1
	includeSepRight includePartitionSeparatorPolicy = 2
)

type partitionIterator struct {
	iStream
	size            int
	f               *glisp.SexpFunction
	separatorPolicy includePartitionSeparatorPolicy
	done            bool
	prev            glisp.Sexp
}

func (iter *partitionIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	if iter.done {
		return glisp.SexpNull, false, nil
	}
	if iter.f == nil {
		group := glisp.NewListBuilder()
		for i := 0; i < iter.size; i++ {
			elem, ok, err := iter.iStream.Next(env)
			if err != nil {
				return glisp.SexpNull, false, err
			}
			if !ok {
				iter.done = true
				return group.Get(), group.Size() > 0, nil
			}
			group.Add(elem)
		}
		return group.Get(), group.Size() > 0, nil
	}
	group := glisp.NewListBuilder()
	for {
		if iter.separatorPolicy == includeSepLeft && iter.prev != nil {
			group.Add(iter.prev)
			iter.prev = nil
		}
		elem, ok, err := iter.iStream.Next(env)
		if err != nil {
			return glisp.SexpNull, false, err
		}
		if !ok {
			iter.done = true
			return group.Get(), group.Size() > 0, nil
		}
		ret, err := env.Apply(iter.f, []glisp.Sexp{elem})
		if err != nil {
			return glisp.SexpNull, false, err
		}
		if !glisp.IsBool(ret) {
			return glisp.SexpNull, false, fmt.Errorf("partition function must return bool but get %v", glisp.InspectType(ret))
		}
		if !bool(ret.(glisp.SexpBool)) {
			group.Add(elem)
			continue
		}
		if iter.separatorPolicy == excludeSep {
			if group.Size() > 0 {
				break
			} else {
				continue
			}
		} else if iter.separatorPolicy == includeSepLeft {
			if group.Size() > 0 {
				iter.prev = elem
				break
			} else {
				group.Add(elem)
			}
		} else {
			group.Add(elem)
			break
		}
	}
	return group.Get(), group.Size() > 0, nil
}

type ZipListIterator struct {
	expr []iStream
	size int
}

func (iter *ZipListIterator) SexpString() string {
	arr := make([]string, len(iter.expr))
	for i, s := range iter.expr {
		arr[i] = s.SexpString()
	}
	return `(zip ` + strings.Join(arr, " ") + ")"
}

func (iter *ZipListIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	elem := glisp.NewListBuilder()
	for i := 0; i < iter.size; i++ {
		v, ok, err := iter.expr[i].Next(env)
		if err != nil || !ok {
			return glisp.SexpNull, false, err
		}
		elem.Add(v)
	}

	return elem.Get(), true, nil
}

type UnionIterator struct {
	expr []iStream
}

func (iter *UnionIterator) SexpString() string {
	arr := make([]string, len(iter.expr))
	for i, s := range iter.expr {
		arr[i] = s.SexpString()
	}
	return `(union ` + strings.Join(arr, " ") + ")"
}

func (iter *UnionIterator) Next(env *glisp.Environment) (glisp.Sexp, bool, error) {
	for cur := iter.expr[0]; ; {
		elem, ok, err := cur.Next(env)
		if err != nil {
			return glisp.SexpNull, false, err
		} else if ok {
			return elem, true, nil
		}
		if len(iter.expr) > 1 {
			iter.expr = iter.expr[1:]
			cur = iter.expr[0]
		} else {
			break
		}
	}
	return glisp.SexpNull, false, nil
}
