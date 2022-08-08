package glisp

import (
	"errors"
	"fmt"
)

var NotAList = errors.New("not a list")

type SexpPair struct {
	head Sexp
	tail Sexp
}

func Cons(a Sexp, b Sexp) *SexpPair {
	return &SexpPair{a, b}
}

func (pair *SexpPair) Head() Sexp {
	return pair.head
}

func (pair *SexpPair) Tail() Sexp {
	return pair.tail
}

func (pair *SexpPair) SexpString() string {
	str := "("

	for {
		switch pair.tail.(type) {
		case *SexpPair:
			str += pair.head.SexpString() + " "
			pair = pair.tail.(*SexpPair)
			continue
		}
		break
	}

	str += pair.head.SexpString()

	if pair.tail == SexpNull {
		str += ")"
	} else {
		str += " . " + pair.tail.SexpString() + ")"
	}

	return str
}

func ListToArray(expr Sexp) ([]Sexp, error) {
	if !IsList(expr) {
		return nil, NotAList
	}
	arr := make([]Sexp, 0)

	for expr != SexpNull {
		list := expr.(*SexpPair)
		arr = append(arr, list.head)
		expr = list.tail
	}

	return arr, nil
}

func MakeList(expressions []Sexp) Sexp {
	if len(expressions) == 0 {
		return SexpNull
	}

	return Cons(expressions[0], MakeList(expressions[1:]))
}

func MapList(env *Context, fun *SexpFunction, expr Sexp) (Sexp, error) {
	if expr == SexpNull {
		return SexpNull, nil
	}

	list := &SexpPair{}
	switch e := expr.(type) {
	case *SexpPair:
		*list = *e
	default:
		return SexpNull, NotAList
	}

	var err error

	list.head, err = env.Apply(fun, []Sexp{list.head})

	if err != nil {
		return SexpNull, err
	}

	list.tail, err = MapList(env, fun, list.tail)

	if err != nil {
		return SexpNull, err
	}

	return list, nil
}

func FlatMapList(env *Environment, fun *SexpFunction, expr Sexp) (Sexp, error) {
	if expr == SexpNull {
		return SexpNull, nil
	}

	list := &SexpPair{}
	switch e := expr.(type) {
	case *SexpPair:
		*list = *e
	default:
		return SexpNull, NotAList
	}

	oldlist := list
	tail := list.tail
	head, err := env.Apply(fun, []Sexp{list.head})
	if err != nil {
		return SexpNull, err
	}

	var noHead bool
	if head == SexpNull {
		noHead = true
	} else {
		if pair, ok := head.(*SexpPair); !ok {
			return SexpNull, fmt.Errorf("flatmap function must return list but got %v", head.SexpString())
		} else {
			list.head = pair.head
			list.tail = pair.tail
			/* go to list end */
			for list.tail != SexpNull {
				list = list.tail.(*SexpPair)
			}
		}
	}

	result, err := FlatMapList(env, fun, tail)
	if err != nil {
		return SexpNull, err
	}
	if noHead {
		return result, nil
	}
	if pair, ok := result.(*SexpPair); ok {
		list.tail = pair
	}

	return oldlist, nil
}

func ConcatList(a *SexpPair, b ...Sexp) (Sexp, error) {
	for _, expr := range b {
		ret, err := concatList(a, expr)
		if err != nil {
			return ret, err
		}
		a = ret.(*SexpPair)
	}
	return a, nil
}

func concatList(a *SexpPair, b Sexp) (Sexp, error) {
	if !IsList(b) {
		return SexpNull, NotAList
	}

	if a.tail == SexpNull {
		return Cons(a.head, b), nil
	}

	switch t := a.tail.(type) {
	case *SexpPair:
		newtail, err := ConcatList(t, b)
		if err != nil {
			return SexpNull, err
		}
		return Cons(a.head, newtail), nil
	}

	return SexpNull, NotAList
}

func FoldlList(env *Context, fun *SexpFunction, lst, acc Sexp) (Sexp, error) {
	if lst == SexpNull {
		return acc, nil
	}

	list := &SexpPair{}
	switch e := lst.(type) {
	case *SexpPair:
		*list = *e
	default:
		return SexpNull, NotAList
	}

	var err error
	if acc, err = env.Apply(fun, []Sexp{list.head, acc}); err != nil {
		return SexpNull, err
	}

	return FoldlList(env, fun, list.tail, acc)
}

func FilterList(env *Context, fun *SexpFunction, list *SexpPair) (Sexp, error) {
	var head *SexpPair
	var last *SexpPair
	for {
		if ret, err := env.Apply(fun, []Sexp{list.head}); err != nil {
			return SexpNull, err
		} else if !IsBool(ret) {
			return SexpNull, errors.New("filter function must return boolean")
		} else if bool(ret.(SexpBool)) {
			if head == nil {
				head = &SexpPair{head: list.head, tail: SexpNull}
				last = head
			} else {
				cell := Cons(list.head, SexpNull)
				last.tail = cell
				last = cell
			}
		}
		if list.tail == SexpNull {
			break
		}
		if next, ok := list.tail.(*SexpPair); ok {
			list = next
		} else {
			break
		}
	}

	if head == nil {
		return SexpNull, nil
	}
	return head, nil
}
