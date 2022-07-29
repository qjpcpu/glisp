package tests

import (
	"strings"
	"testing"
)

func TestCompareFloatWithString(t *testing.T) {
	scritp := `(= 1.0 "a")`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpFloat(1.0) to glisp.SexpStr("a")`)
}

func TestCompareIntWithString(t *testing.T) {
	scritp := `(= 1 "a")`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpInt(1) to glisp.SexpStr("a")`)
}

func TestCompareStringWithInt(t *testing.T) {
	scritp := `(= "a" 1)`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareCharWithHash(t *testing.T) {
	scritp := `(= #a {})`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpChar(#a) to *glisp.sexpHash({})`)
}

func TestCompareBytesWithHash(t *testing.T) {
	scritp := `(= 0B6869 {})`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpBytes(0B6869) to *glisp.sexpHash({})`)
}

func TestCompareHashWithList(t *testing.T) {
	scritp := `(=  {} '(1))`
	expectErrorContains(t, scritp, `cannot compare *glisp.sexpHash({}) to glisp.SexpPair((1))`)
}

func TestCompareListWithHash(t *testing.T) {
	scritp := `(= '(1) {})`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpPair((1)) to *glisp.sexpHash({})`)
}

func TestCompareBoolWithInt(t *testing.T) {
	scritp := `(= true 1)`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpBool(true) to glisp.SexpInt(1)`)
}

func TestCompareListStringWithListInt(t *testing.T) {
	scritp := `(= '("a") '(1))`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareArrayStringWithArrayInt(t *testing.T) {
	scritp := `(= ["a"] [1])`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareArrayWithInt(t *testing.T) {
	scritp := `(= [] 1)`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpArray([]) to glisp.SexpInt(1)`)
}

func TestNotComparable(t *testing.T) {
	scritp := `(= (make-chan) 1)`
	expectErrorContains(t, scritp, `cannot compare extensions.SexpChannel([chan]) to glisp.SexpInt(1)`)
}

func TestErrorExistInList(t *testing.T) {
	scritp := `(exist? '("a") 1)`
	expectErrorContains(t, scritp, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestConcatStringErr(t *testing.T) {
	script := `(concat "" 1)`
	expectErrorContains(t, script, `second argument is not a string`)
}

func TestConcatBytesErr(t *testing.T) {
	script := `(concat 0B6869 1)`
	expectErrorContains(t, script, `second argument is not bytes`)
}

func TestAppendStringErr(t *testing.T) {
	script := `(append "" 1)`
	expectErrorContains(t, script, `second argument is not a char`)
}

func TestHashKey(t *testing.T) {
	script := `(exist? {} {})`
	expectErrorContains(t, script, `cannot hash type *glisp.sexpHash`)
}

func TestEvalNothing(t *testing.T) {
	script := `(eval (begin))`
	expectErrorContains(t, script, `generating (eval (begin))`)
}

func TestExistArrayNotCompare(t *testing.T) {
	script := `(exist? [1] "a")`
	expectErrorContains(t, script, `compare glisp.SexpInt(1) to glisp.SexpStr("a")`)
}

func TestApplySymbolNotFound(t *testing.T) {
	script := `(apply 'not-found-symbol "a")`
	expectErrorContains(t, script, `can't find function by symbol not-found-symbol`)
}

func TestApplyArgMustBeList(t *testing.T) {
	script := `(apply + (cons 1 2))`
	expectErrorContains(t, script, `not a list`)
}

func TestWrongArgumentNumber(t *testing.T) {
	mustErr := func(script string) {
		expectErrorOccur(t, script)
	}
	mustErr(`(gensym 1)`)
	mustErr(`(str2sym)`)
	mustErr(`(str2sym 1)`)
	mustErr(`(sym2str)`)
	mustErr(`(sym2str 1)`)
	mustErr(`(typestr)`)
	mustErr(`(>)`)
	mustErr(`(+)`)
	mustErr(`(+ 1 "a")`)
	mustErr(`(cons)`)
	mustErr(`(car)`)
	mustErr(`(cdr)`)
	mustErr(`(car 1)`)
	mustErr(`(cdr 1)`)
	mustErr(`(aget)`)
	mustErr(`(aget 1)`)
	mustErr(`(aget 1 1)`)
	mustErr(`(list?)`)
	mustErr(`(aget [1] "")`)
	mustErr(`(aget [1] -1)`)
	mustErr(`(aset! [1] #a)`)
	mustErr(`(aset! [1] 0)`)
	mustErr(`(sget 0)`)
	mustErr(`(sget 0 0)`)
	mustErr(`(sget "" #a)`)
	mustErr(`(sget "" 0B6869)`)
	mustErr(`(exist?)`)
	mustErr(`(exist? 1 1)`)
	mustErr(`(hget {} 1 2 3 4)`)
	mustErr(`(hget 0 1 2 )`)
	mustErr(`(hdel! {} 1 2)`)
	mustErr(`(slice 1 2 3 4)`)
	mustErr(`(slice [] (make-chan) 1)`)
	mustErr(`(slice [] 1 (make-chan))`)
	mustErr(`(slice (make-chan) 1 1)`)
	mustErr(`(slice [] 1 1)`)
	mustErr(`(slice "" 1 1)`)
	mustErr(`(slice 0B6869 10 100)`)
	mustErr(`(slice 0B6869 10 1)`)
	mustErr(`(slice 0B6869 #a 100)`)
	mustErr(`(slice 0B6869 10 #b)`)
	mustErr(`(len)`)
	mustErr(`(len 1)`)
	mustErr(`(append)`)
	mustErr(`(append 1 1)`)
	mustErr(`(concat)`)
	mustErr(`(concat 1 1)`)
	mustErr(`(read)`)
	mustErr(`(read 1)`)
	mustErr(`(eval)`)
	mustErr(`(not)`)
	mustErr(`(apply)`)
	mustErr(`(apply 1 1)`)
	mustErr(`(apply '+ 1)`)
	mustErr(`(map)`)
	mustErr(`(map 1 1)`)
	mustErr(`(map + 1)`)
	mustErr(`(filer)`)
	mustErr(`(filer 1 1)`)
	mustErr(`(make-array)`)
	mustErr(`(make-array "")`)
	mustErr(`(symnum)`)
	mustErr(`(symnum 1)`)
	mustErr(`(str)`)
	mustErr(`(str2int)`)
	mustErr(`(str2int 1)`)
	mustErr(`(bytes2str)`)
	mustErr(`(bytes2str 1)`)
	mustErr(`(str2bytes 1)`)
	mustErr(`(str2bytes)`)
	mustErr(`(str2float)`)
	mustErr(`(str2float 1)`)
	mustErr(`(float2int)`)
	mustErr(`(float2int "")`)
	mustErr(`(float2str)`)
	mustErr(`(float2str {})`)
	mustErr(`(float2str 1.0 "")`)
	mustErr(`(round)`)
	mustErr(`(round "")`)
	mustErr(`(foldl)`)
	mustErr(`(foldl + 1 1)`)
	mustErr(`(foldl "+" 1 1)`)
	mustErr(`(filter)`)
	mustErr(`(filter + 1)`)
	mustErr(`(filter "+" 1)`)
}

func expectErrorContains(t *testing.T, script string, keyword string) {
	expectErrorMatch(t, script, func(err error) bool { return strings.Contains(err.Error(), keyword) })
}

func expectErrorOccur(t *testing.T, script string) {
	expectErrorMatch(t, script, func(err error) bool { return true })
}

func expectErrorMatch(t *testing.T, script string, expect func(err error) bool) {
	env := newFullEnv()
	expr, err := env.EvalString(script)
	if err == nil {
		t.Fatalf("expect error occur, but success with %v\n", expr.SexpString())
	}
	if !expect(err) {
		t.Fatalf("error not match expect: %v", err)
	}
}
