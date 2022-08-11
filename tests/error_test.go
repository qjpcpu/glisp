package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qjpcpu/glisp"
)

func TestCompareFloatWithString(t *testing.T) {
	script := `(= 1.0 "a")`
	expectErrorContains(t, script, `cannot compare glisp.SexpFloat(1.0) to glisp.SexpStr("a")`)
}

func TestMapListFail(t *testing.T) {
	script := `(map (fn [a] (+ a 1)) '("a"))`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(map (fn [a] (+ a 1)) '(1 "a"))`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(map (fn [a] (+ a 1)) 1)`
	expectErrorContains(t, script, `second argument of map must be array/list`)
}

func TestFlatMapListFail(t *testing.T) {
	script := `(flatmap (fn [a] (+ a 1)) '("a"))`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(flatmap (fn [a] (+ a 1)) '(1))`
	expectErrorContains(t, script, `flatmap function must return list but got 2`)
	script = `(flatmap (fn [a] (list (+ a 1))) '(1 "a"))`
	expectErrorContains(t, script, `operands have invalid type`)
}

func TestFilterListFail(t *testing.T) {
	script := `(filter (fn [a] (+ "3" 1)) '("a"))`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(filter (fn [a] (+ 3 1)) '("a"))`
	expectErrorContains(t, script, `filter function must return boolean`)
}

func TestFoldlListFail(t *testing.T) {
	script := `(foldl + 0 '("a"))`
	expectErrorContains(t, script, `operands have invalid type`)
}

func TestMapArrayFail(t *testing.T) {
	script := `(map (fn [a] (+ a 1)) ["a"])`
	expectErrorContains(t, script, `operands have invalid type`)
}

func TestFilterArrayFail(t *testing.T) {
	script := `(filter (fn [a] (+ "a" 1)) ["a"])`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(filter (fn [a] (+ 1 1)) ["a"])`
	expectErrorContains(t, script, `filter function must return boolean`)
}

func TestFlatMapArrayFail(t *testing.T) {
	script := `(flatmap (fn [a] (+ "a" 1)) ["a"])`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(flatmap (fn [a] (+ 1 1)) ["a"])`
	expectErrorContains(t, script, `flatmap function must return array`)
}

func TestFilterHash(t *testing.T) {
	script := `(filter (fn [a b] (+ "a" 1)) {'a 1})`
	expectErrorContains(t, script, `operands have invalid type`)
	script = `(filter (fn [k v] (+ 1 1)) {'a 1})`
	expectErrorContains(t, script, `filter function must return boolean`)
}

func TestFoldlHash(t *testing.T) {
	script := `(foldl (fn [a b c] (+ "a" 1)) 0 {'a 1})`
	expectErrorContains(t, script, `operands have invalid type`)
}

type AnyStruct struct{ Number int }

func (AnyStruct) SexpString() string { return "" }

func TestMarshalAny(t *testing.T) {
	env := newFullEnv()
	env.Bind("g_var", AnyStruct{Number: 1024})
	script := `(= "{\"Number\":1024}" (json/stringify g_var))`
	ret, err := env.EvalString(script)
	if err != nil || !bool(ret.(glisp.SexpBool)) {
		t.Fatal("marshal any")
	}
}

func TestCompareIntWithString(t *testing.T) {
	script := `(= 1 "a")`
	expectErrorContains(t, script, `cannot compare glisp.SexpInt(1) to glisp.SexpStr("a")`)
}

func TestIsZero(t *testing.T) {
	if !glisp.IsZero(glisp.SexpChar(0)) {
		t.Fatal("zero char")
	}
	if glisp.IsTruthy(glisp.SexpChar(0)) {
		t.Fatal("zero char")
	}
	if !glisp.IsNumber(glisp.SexpChar(0)) {
		t.Fatal("zero char")
	}
}

func TestCompareStringWithInt(t *testing.T) {
	script := `(= "a" 1)`
	expectErrorContains(t, script, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareCharWithHash(t *testing.T) {
	script := `(= #a {})`
	expectErrorContains(t, script, `cannot compare glisp.SexpChar(#a) to *glisp.SexpHash({})`)
}

func TestCompareBytesWithHash(t *testing.T) {
	script := `(= 0B6869 {})`
	expectErrorContains(t, script, `cannot compare glisp.SexpBytes(0B6869) to *glisp.SexpHash({})`)
}

func TestCompareHashWithList(t *testing.T) {
	script := `(=  {} '(1))`
	expectErrorContains(t, script, `cannot compare *glisp.SexpHash({}) to *glisp.SexpPair((1))`)
}

func TestCompareListWithHash(t *testing.T) {
	script := `(= '(1) {})`
	expectErrorContains(t, script, `cannot compare *glisp.SexpPair((1)) to *glisp.SexpHash({})`)
}

func TestCompareBoolWithInt(t *testing.T) {
	script := `(= true 1)`
	expectErrorContains(t, script, `cannot compare glisp.SexpBool(true) to glisp.SexpInt(1)`)
}

func TestCompareListStringWithListInt(t *testing.T) {
	script := `(= '("a") '(1))`
	expectErrorContains(t, script, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareArrayStringWithArrayInt(t *testing.T) {
	script := `(= ["a"] [1])`
	expectErrorContains(t, script, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
}

func TestCompareArrayWithInt(t *testing.T) {
	script := `(= [] 1)`
	expectErrorContains(t, script, `cannot compare glisp.SexpArray([]) to glisp.SexpInt(1)`)
}

func TestDiv0(t *testing.T) {
	script := `(/ 1 0)`
	expectErrorContains(t, script, `division by zero`)
}

func TestNotComparable(t *testing.T) {
	script := `(= (make-chan) 1)`
	expectErrorContains(t, script, `cannot compare extensions.SexpChannel([chan]) to glisp.SexpInt(1)`)
}

func TestErrorExistInList(t *testing.T) {
	script := `(exist? '("a") 1)`
	expectErrorContains(t, script, `cannot compare glisp.SexpStr("a") to glisp.SexpInt(1)`)
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
	expectErrorContains(t, script, `cannot hash type *glisp.SexpHash`)
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

func TestJSONParse(t *testing.T) {
	script := `(json/parse (time/now))`
	expectErrorContains(t, script, `the first argument of json/parse must be`)
}

func TestApplyArgMustBeList(t *testing.T) {
	script := `(apply + (cons 1 2))`
	expectErrorContains(t, script, `not a list`)
}

func TestDefensiveCornor(t *testing.T) {
	if glisp.SexpEnd.SexpString() != `End` {
		t.Fatal(glisp.SexpEnd.SexpString())
	}
	if glisp.SexpMarker.SexpString() != `Marker` {
		t.Fatal(glisp.SexpMarker.SexpString())
	}
	if glisp.SexpSentinel(100).SexpString() != "" {
		t.Fatal("should be emtpy")
	}
	if _, err := glisp.NewSexpIntStr("abc"); err == nil {
		t.Fatal("should be error")
	}
	if _, err := glisp.NewSexpIntStrWithBase("abc", 12); err == nil {
		t.Fatal("should be error")
	}
	if _, err := glisp.NewSexpFloatStr("abc"); err == nil {
		t.Fatal("should be error")
	}
	if glisp.NewSexpFloat(1).SexpString() != "1" {
		t.Fatal("should be 1")
	}
	if _, err := glisp.NewSexpBytesByHex("世界"); err == nil {
		t.Fatal("should be error")
	}
}

func TestRecover(t *testing.T) {
	env := newFullEnv()
	env.PushGlobalScope()
	if _, err := env.EvalString(`(def g_var 1023) (/ 1 0)`); err == nil {
		t.Fatal("must error")
	}
	_, ok := env.FindObject(`g_var`)
	if !ok {
		t.Fatal("should find g_var")
	}
	env.Clear()
	_, ok = env.FindObject(`g_var`)
	if ok {
		t.Fatal("should not find g_var")
	}
}

func TestWrongArgumentNumber(t *testing.T) {
	mustErr := func(script string) {
		expectErrorOccur(t, script)
	}
	mustErr(`(gensym 1)`)
	mustErr(`(sprintf)`)
	mustErr(`(sprintf 1)`)
	mustErr(`(sprintf 1 2)`)
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
	mustErr(`(/ 1 "")`)
	mustErr(`(/ "" 1)`)
	mustErr(`(/ #a "")`)
	mustErr(`(/ #a 0)`)
	mustErr(`(/ 1.0 "")`)
	mustErr(`(assert)`)
	mustErr(`({'a})`)
	mustErr(`(rand 1 2)`)
	mustErr(`(rand "")`)
	mustErr(`(rand 0)`)
	mustErr(`(randf 1)`)
	mustErr(`(json/stringify)`)
	mustErr(`(json/parse)`)
	mustErr(`(regexp-compile)`)
	mustErr(`(regexp-compile 1)`)
	mustErr(`(regexp-compile "\\")`)
	mustErr(`(regexp-find)`)
	mustErr(`(regexp-find "h" 1)`)
	mustErr(`(regexp-find "h" "a")`)
	mustErr(`(time/now "a")`)
	mustErr(`(time/year)`)
	mustErr(`(time/year 1)`)
	mustErr(`(time/month)`)
	mustErr(`(time/month 1)`)
	mustErr(`(time/day)`)
	mustErr(`(time/day 1)`)
	mustErr(`(time/hour)`)
	mustErr(`(time/hour 1)`)
	mustErr(`(time/minute)`)
	mustErr(`(time/minute 1)`)
	mustErr(`(time/second)`)
	mustErr(`(time/second 1)`)
	mustErr(`(time/weekday)`)
	mustErr(`(time/weekday 1)`)
	mustErr(`(time/sub)`)
	mustErr(`(time/sub 1 2)`)
	mustErr(`(time/add)`)
	mustErr(`(time/add 1 2 3)`)
	mustErr(`(time/add (time/now) "a" 3)`)
	mustErr(`(time/add (time/now) 1 "a")`)
	mustErr(`(time/add (time/now) 1 1)`)
	mustErr(`(time/add-date)`)
	mustErr(`(time/add-date 1 2 3 4)`)
	mustErr(`(time/add-date (time/now) "a" 3 4)`)
	mustErr(`(time/parse)`)
	mustErr(`(time/parse "xyz")`)
	mustErr(`(time/parse 1.2)`)
	mustErr(`(time/parse 1.2 1)`)
	mustErr(`(time/parse "a" 1)`)
	mustErr(`(time/parse "a" 1 2)`)
	mustErr(`(time/format)`)
	mustErr(`(time/format 1 2)`)
	mustErr(`(time/format (time/now) 2)`)
	mustErr(`(time/format (time/now) "")`)
	mustErr(`(str/start-with?)`)
	mustErr(`(str/start-with? 1 2)`)
	mustErr(`(str/start-with? "" 2)`)
	mustErr(`(str/count)`)
	mustErr(`(str/count 1 2)`)
	mustErr(`(str/count "" 2)`)
	mustErr(`(str/split)`)
	mustErr(`(str/split 1 2)`)
	mustErr(`(str/split "" 2)`)
	mustErr(`(str/lower)`)
	mustErr(`(str/lower 1)`)
	mustErr(`(str/digit?)`)
	mustErr(`(str/digit? 1)`)
	mustErr(`(str/join)`)
	mustErr(`(str/join 1 2)`)
	mustErr(`(str/join [] 2)`)
	mustErr(`(str/join [1] "")`)
	mustErr(`(str/join [] "" 2)`)
	mustErr(`(str/trim-prefix)`)
	mustErr(`(str/trim-prefix 1 2)`)
	mustErr(`(str/trim-prefix "" 2)`)
	mustErr(`(str/replace)`)
	mustErr(`(str/replace 1 2 3)`)
	mustErr(`(str/replace "" 2 3)`)
	mustErr(`(str/replace "" "" 3)`)
	mustErr(`(str/mask)`)
	mustErr(`(str/mask 1 2 3 4)`)
	mustErr(`(str/mask "" "" 3 4)`)
	mustErr(`(str/mask "" 1 "" 4)`)
	mustErr(`(str/mask "" 1 1 4)`)
	mustErr(`(str/mask "" 1 0 4)`)
	mustErr(`(str/mask "abc" 1 0 "*")`)
	mustErr(`(str/mask "abc" 1 0 "")`)
	mustErr(`(str/mask "abc" 1 -1 "")`)
	mustErr(`(io/read-file)`)
	mustErr(`(io/read-file 1)`)
	mustErr(`(io/read-file "not-exist-file")`)
	mustErr(`(io/write-file)`)
	mustErr(`(io/write-file 1 2)`)
	mustErr(`(io/write-file "./test" 2)`)

	notExistDir := `~/tmp/not-exist-dir/`
	home, _ := os.UserHomeDir()
	os.RemoveAll(filepath.Join(home, strings.TrimPrefix(notExistDir, "~")))
	mustErr(fmt.Sprintf(`(io/write-file "%s" 2)`, notExistDir+"not-exit-file"))
	os.RemoveAll(filepath.Join(home, strings.TrimPrefix(notExistDir, "~")))

	mustErr(`(io/remove-file)`)
	mustErr(`(io/remove-file 1)`)
	mustErr(`(io/file-exist?)`)
	mustErr(`(io/file-exist? 1)`)
	mustErr(`(flatmap)`)
	mustErr(`(flatmap 1 2)`)
	mustErr(`(flatmap (fn [] 1) 2)`)
	mustErr(`(compose)`)
	mustErr(`(compose 1 2)`)
	mustErr(`(make-chan 1 2 3)`)
	mustErr(`(make-chan "a")`)
	mustErr(`(send!)`)
	mustErr(`(send! "a")`)
	mustErr(`(send! (make-chan))`)
	mustErr(`(base64/encode)`)
	mustErr(`(base64/encode 1)`)
	mustErr(`(base64/decode)`)
	mustErr(`(base64/decode 1)`)
	mustErr(`(base64/decode "^&*^*&%")`)
	mustErr(`(mod "" "")`)
	mustErr(`(mod 1 "")`)
	mustErr(`(mod 1 0)`)
	mustErr(`(bit-and)`)
	mustErr(`(bit-and "" "")`)
	mustErr(`(bit-not)`)
	mustErr(`(bit-not "")`)
	mustErr(`(sla)`)
	mustErr(`(sll8)`)
	mustErr(`(sll8 "" "")`)
	mustErr(`(sll8 1 "")`)
	mustErr(`(char2int)`)
	mustErr(`(char2int "")`)
	mustErr(`(char2str)`)
	mustErr(`(char2str "")`)
	mustErr(`(int2char)`)
	mustErr(`(int2char "")`)
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
