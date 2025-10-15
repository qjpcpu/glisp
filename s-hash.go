package glisp

import (
	"errors"
	"fmt"
	"hash/fnv"
)

type SexpHash struct {
	Map      map[int][]*SexpPair
	KeyOrder []Sexp // must user pointers here, else hset! will fail to update.
	NumKeys  int
}

func (hash *SexpHash) SexpString() string {
	str := "{"
	for _, arr := range hash.Map {
		for _, pair := range arr {
			str += pair.head.SexpString() + " "
			str += pair.tail.SexpString() + " "
		}
	}
	if len(str) > 1 {
		return str[:len(str)-1] + "}"
	}
	return str + "}"
}

func HashExpression(expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		if e.IsInt64() {
			return e.ToInt(), nil
		}
		hasher := fnv.New32()
		_, err := hasher.Write([]byte(e.SexpString()))
		if err != nil {
			return 0, err
		}
		return int(hasher.Sum32()), nil
	case SexpChar:
		return int(e), nil
	case SexpSymbol:
		return e.number, nil
	case SexpStr:
		hasher := fnv.New32()
		_, err := hasher.Write([]byte(e))
		if err != nil {
			return 0, err
		}
		return int(hasher.Sum32()), nil
	case SexpBool:
		if bool(e) {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot hash type %v", InspectType(expr))
}

func MakeHash(args []Sexp) (*SexpHash, error) {
	if len(args)%2 != 0 {
		return &SexpHash{},
			errors.New("hash requires even number of arguments")
	}

	var memberCount int
	hash := &SexpHash{
		Map:      make(map[int][]*SexpPair),
		KeyOrder: []Sexp{},
		NumKeys:  memberCount,
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
	hashval, err := HashExpression(key)
	if err != nil {
		return SexpNull, err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		return defaultval, nil
	}

	for _, pair := range arr {
		res, err := Compare(pair.head, key)
		if err == nil && res == 0 {
			return pair.tail, nil
		}
	}
	return defaultval, nil
}

func (hash *SexpHash) HashSet(key Sexp, val Sexp) error {
	hashval, err := HashExpression(key)
	if err != nil {
		return err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		hash.Map[hashval] = []*SexpPair{Cons(key, val)}
		hash.KeyOrder = append(hash.KeyOrder, key)
		(hash.NumKeys)++
		return nil
	}

	found := false
	for i, pair := range arr {
		res, err := Compare(pair.head, key)
		if err == nil && res == 0 {
			arr[i] = Cons(key, val)
			found = true
		}
	}

	if !found {
		arr = append(arr, Cons(key, val))
		hash.KeyOrder = append(hash.KeyOrder, key)
		(hash.NumKeys)++
	}

	hash.Map[hashval] = arr

	return nil
}

func (hash *SexpHash) HashDelete(key Sexp) error {
	hashval, err := HashExpression(key)
	if err != nil {
		return err
	}
	arr, ok := hash.Map[hashval]

	// if it doesn't exist, no need to delete it
	if !ok {
		return nil
	}

	(hash.NumKeys)--
	for i, pair := range arr {
		res, err := Compare(pair.head, key)
		if err == nil && res == 0 {
			if len(arr) == 1 {
				for j, k := range hash.KeyOrder {
					if kr, kerr := Compare(k, key); kerr == nil && kr == 0 {
						hash.KeyOrder = append((hash.KeyOrder)[0:j], (hash.KeyOrder)[j+1:]...)
						break
					}
				}
			}
			hash.Map[hashval] = append(arr[0:i], arr[i+1:]...)
			break
		}
	}

	return nil
}

func HashCountKeys(hash *SexpHash) (int, error) {
	var num int
	for _, arr := range hash.Map {
		num += len(arr)
	}
	if num != hash.NumKeys {
		return 0, fmt.Errorf("HashCountKeys disagreement on count: num=%d, (*hash.NumKeys)=%d", num, hash.NumKeys)
	}
	return num, nil
}

func (hash *SexpHash) Foreach(fn func(Sexp, Sexp) bool) {
	keys := hash.KeyOrder
	for _, key := range keys {
		val, _ := hash.HashGet(key)
		if !fn(key, val) {
			break
		}
	}
}

func HashIsEmpty(hash *SexpHash) bool {
	for _, arr := range hash.Map {
		if len(arr) > 0 {
			return false
		}
	}
	return true
}

func FilterHash(env *Environment, fun *SexpFunction, hash *SexpHash) (*SexpHash, error) {
	result, err := MakeHash(nil)
	if err != nil {
		return hash, err
	}

	hash.Foreach(func(key Sexp, val Sexp) bool {
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
	if hash.NumKeys == 0 {
		return acc, nil
	}

	var err error
	hash.Foreach(func(k Sexp, v Sexp) bool {
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

	for _, key := range arr.KeyOrder {
		val, err := arr.HashGet(key)
		if err != nil {
			return SexpNull, err
		}
		res, err := env.Apply(fun, MakeArgs(Cons(key, val)))
		if err != nil {
			return SexpNull, err
		}
		if res == SexpNull {
			continue
		}
		if IsArray(res) {
			arr := res.(SexpArray)
			result.Add(arr...)
		} else if IsList(res) {
			arr, _ := ListToArray(res)
			result.Add(arr...)
		} else {
			return SexpNull, errors.New("flatmap function must return array/list")
		}
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
		for _, key := range eh.KeyOrder {
			val, _ := eh.HashGet(key)
			h.HashSet(key, val)
		}
		return true
	})
	if err != nil {
		return h, err
	}
	return h, nil
}
