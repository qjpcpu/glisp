package glisp

import (
	"errors"
	"fmt"
	"strings"
)

const CompactHashThreshold = 16

type hashkey struct {
	t string
	k string
}

type hashval struct {
	k, v Sexp
	i    int
}

type SexpHash struct {
	Map      map[hashkey]hashval
	keys     []hashkey
	del      []int
	delCount int
}

func (hash *SexpHash) SexpString() string {
	str := "{"
	for _, elem := range hash.Map {
		str += elem.k.SexpString() + " "
		str += elem.v.SexpString() + " "
	}
	if len(str) > 1 {
		return str[:len(str)-1] + "}"
	}
	return str + "}"
}

func hashExpr(e Sexp) (hashkey, error) {
	switch expr := e.(type) {
	case SexpSymbol:
		return hashkey{t: "symbol", k: expr.SexpString()}, nil
	case SexpStr:
		return hashkey{t: "string", k: expr.SexpString()}, nil
	case SexpBool:
		return hashkey{t: "bool", k: expr.SexpString()}, nil
	case SexpChar:
		return hashkey{t: "char", k: expr.SexpString()}, nil
	case SexpInt:
		return hashkey{t: "int", k: expr.SexpString()}, nil
	default:
		return hashkey{}, fmt.Errorf("can't hash type %s", GetSexpType(e))
	}
}

func HashExpr(e Sexp) (string, error) {
	k, err := hashExpr(e)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	sb.WriteString(k.t)
	sb.WriteString(":")
	sb.WriteString(k.k)
	return sb.String(), nil
}

func MakeHash(args []Sexp) (*SexpHash, error) {
	if len(args)%2 != 0 {
		return &SexpHash{}, errors.New("hash requires even number of arguments")
	}

	hash := &SexpHash{
		Map: make(map[hashkey]hashval),
	}
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		err := hash.HashSet(key, val)
		if err != nil {
			return hash, err
		}
	}
	return hash, nil
}

func (hash *SexpHash) HashExist(key Sexp) bool {
	_, err := hash.HashGet(key)
	return err == nil
}

func (hash *SexpHash) HashGet(key Sexp) (Sexp, error) {
	// this is kind of a hack
	// SexpEnd can't be created by user
	// so there is no way it would actually show up in the map
	val, err := hash.HashGetDefault(key, SexpEnd)
	if err != nil {
		return SexpNull, err
	}
	if val == SexpEnd {
		return SexpNull, fmt.Errorf("key %s not found", key.SexpString())
	}
	return val, nil
}

func (hash *SexpHash) HashGetDefault(key Sexp, defaultval Sexp) (Sexp, error) {
	hkey, err := hashExpr(key)
	if err != nil {
		return SexpNull, err
	}
	elem, ok := hash.Map[hkey]
	if !ok {
		return defaultval, nil
	}
	return elem.v, nil
}

func (hash *SexpHash) HashSet(key Sexp, val Sexp) error {
	hkey, err := hashExpr(key)
	if err != nil {
		return err
	}
	elem, ok := hash.Map[hkey]
	if !ok {
		hash.Map[hkey] = hashval{k: key, v: val, i: len(hash.keys)}
		hash.keys = append(hash.keys, hkey)
		hash.del = append(hash.del, 0)
	} else {
		hash.Map[hkey] = hashval{k: key, v: val, i: elem.i}
	}
	return nil
}

func (hash *SexpHash) HashDelete(key Sexp) error {
	hkey, err := hashExpr(key)
	if err != nil {
		return err
	}
	elem, ok := hash.Map[hkey]
	// if it doesn't exist, no need to delete it
	if !ok {
		return nil
	}

	delete(hash.Map, hkey)
	hash.del[elem.i] = 1
	hash.delCount++
	hash.compact()
	return nil
}

func (hash *SexpHash) compact() {
	if hash.delCount > CompactHashThreshold {
		var delCount int
		for i := range hash.keys {
			if hash.del[i] == 1 {
				hash.del[i] = 0
				delCount++
			} else if delCount > 0 {
				key := hash.keys[i]
				val := hash.Map[key]
				val.i = i - delCount
				hash.Map[key] = val
				hash.keys[i-delCount] = hash.keys[i]
			}
		}
		size := len(hash.keys)
		hash.keys = hash.keys[:size-delCount]
		hash.del = hash.del[:size-delCount]
		hash.delCount = 0
	}
}

func HashCountKeys(hash *SexpHash) (int, error) {
	return len(hash.Map), nil
}

func (hash *SexpHash) Visit(fn func(Sexp, Sexp) bool) {
	for i := range len(hash.keys) {
		if hash.del[i] == 0 {
			elem := hash.Map[hash.keys[i]]
			if !fn(elem.k, elem.v) {
				break
			}
		}
	}
}

func (hash *SexpHash) ReverseVisit(fn func(Sexp, Sexp) bool) {
	for i := len(hash.keys) - 1; i >= 0; i-- {
		if hash.del[i] == 0 {
			elem := hash.Map[hash.keys[i]]
			if !fn(elem.k, elem.v) {
				break
			}
		}
	}
}

func HashIsEmpty(hash *SexpHash) bool {
	return len(hash.Map) == 0
}

func FilterHash(env *Environment, fun *SexpFunction, hash *SexpHash) (*SexpHash, error) {
	result, err := MakeHash(nil)
	if err != nil {
		return hash, err
	}

	hash.Visit(func(key Sexp, val Sexp) bool {
		ret, err0 := env.Apply(fun, MakeArgs(Cons(key, val)))
		if err0 != nil {
			err = err0
			return false
		}
		pass, ok := ret.(SexpBool)
		if !ok {
			err = errors.New("filter function must return boolean")
			return false
		}
		if pass {
			result.HashSet(key, val)
		}
		return true
	})

	if err != nil {
		return hash, err
	}

	return result, nil
}

func FoldlHash(env *Environment, fun *SexpFunction, hash *SexpHash, acc Sexp) (Sexp, error) {
	if len(hash.Map) == 0 {
		return acc, nil
	}

	var err error
	hash.Visit(func(k Sexp, v Sexp) bool {
		if acc, err = env.Apply(fun, MakeArgs(Cons(k, v), acc)); err != nil {
			return false
		}
		return true
	})

	if err != nil {
		return acc, err
	}

	return acc, nil
}

func FlatMapHash(env *Environment, fun *SexpFunction, arr *SexpHash) (Sexp, error) {
	result := NewListBuilder()

	var err error
	arr.Visit(func(key Sexp, val Sexp) bool {
		res, err0 := env.Apply(fun, MakeArgs(Cons(key, val)))
		if err0 != nil {
			err = err0
			return false
		}
		if res == SexpNull {
			return true
		}
		if IsArray(res) {
			arr := res.(SexpArray)
			result.Add(arr...)
		} else if IsList(res, true) {
			arr, err0 := ListToArray(res)
			if err0 != nil {
				err = err0
				return false
			}
			result.Add(arr...)
		} else {
			err = errors.New("flatmap function must return array/list")
			return false
		}
		return true
	})
	if err != nil {
		return SexpNull, err
	}
	return result.Get(), nil
}

func (hash *SexpHash) Explain(env *Environment, field string, args Args) (Sexp, error) {
	if args.Len() > 1 {
		return WrongNumberArguments("hash field accessor", args.Len(), 0, 1)
	}
	kstr := SexpStr(field)
	if v, err := hash.HashGet(kstr); err == nil {
		return v, nil
	}
	ksym := env.MakeSymbol(field)
	if v, err := hash.HashGet(ksym); err == nil {
		return v, nil
	}
	if args.Len() == 1 {
		return args.Get(0), nil
	}
	return SexpNull, fmt.Errorf("field %s not found", field)
}

func ConcatHash(h *SexpHash, exprs Args) (Sexp, error) {
	var err error
	exprs.Foreach(func(e Sexp) bool {
		if !IsHash(e) {
			err = fmt.Errorf("expect hash but got %s", InspectType(e))
			return false
		}
		eh := e.(*SexpHash)
		eh.Visit(func(key Sexp, val Sexp) bool {
			h.HashSet(key, val)
			return true
		})
		return true
	})
	if err != nil {
		return h, err
	}
	return h, nil
}

func (hash *SexpHash) Iter() *HashIter {
	return &HashIter{hash: hash}
}

type HashIter struct {
	hash *SexpHash
	idx  int
}

func (iter *HashIter) Next() (Sexp, Sexp, bool) {
	if iter.idx >= len(iter.hash.keys) {
		return SexpNull, SexpNull, false
	}
	for ; iter.idx < len(iter.hash.keys); iter.idx++ {
		if iter.hash.del[iter.idx] == 0 {
			key := iter.hash.keys[iter.idx]
			val := iter.hash.Map[key]
			iter.idx++
			return val.k, val.v, true
		}
	}
	return SexpNull, SexpNull, false
}
