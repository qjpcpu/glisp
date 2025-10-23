package tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qjpcpu/glisp"
)

func TestCompareFloatWithString(t *testing.T) {
	script := `(= 1.0 "a")`
	ExpectScriptErr(t, script, `cannot compare float to string`)
}

func TestTypeInstrCoverage(t *testing.T) {
	env := newFullEnv()
	fn := glisp.GetTypeFunction(`type`)
	expr, err := fn(env, glisp.MakeArgs(glisp.SexpEnd))
	ExpectSuccess(t, err)
	ExpectEqStr(t, `<end>`, expr)
	expr, err = fn(env, glisp.MakeArgs(glisp.SexpMarker))
	ExpectSuccess(t, err)
	ExpectEqStr(t, `<marker>`, expr)
}

func TestOverrideFunction(t *testing.T) {
	env := newFullEnv()
	newf := func(*glisp.SexpFunction) glisp.UserFunction {
		return func(*glisp.Environment, glisp.Args) (glisp.Sexp, error) { return glisp.SexpNull, nil }
	}
	err := env.OverrideFunction("xxxxx", newf)
	ExpectError(t, err, "function `xxxxx` not found")
	err = env.OverrideFunction("nil", newf)
	ExpectError(t, err, "`nil` is not a function")
}

func TestMapListFail(t *testing.T) {
	script := `(map (fn [a] (+ a 1)) '("a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(map (fn [a] (+ a 1)) '(1 "a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(map (fn [a] (+ a 1)) 1)`
	ExpectScriptErr(t, script, `second argument of map must be array/list`)
}

func TestFlatMapListFail(t *testing.T) {
	script := `(flatmap (fn [a] (+ a 1)) '("a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(flatmap (fn [a] (+ a 1)) '(1))`
	ExpectScriptErr(t, script, `flatmap function must return list but got 2`)
	script = `(flatmap (fn [a] (list (+ a 1))) '(1 "a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
}

func TestFilterListFail(t *testing.T) {
	script := `(filter (fn [a] (+ "3" 1)) '("a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(filter (fn [a] (+ 3 1)) '("a"))`
	ExpectScriptErr(t, script, `filter function must return boolean`)
}

func TestFoldlListFail(t *testing.T) {
	script := `(foldl + 0 '("a"))`
	ExpectScriptErr(t, script, `operands have invalid type`)
}

func TestMapArrayFail(t *testing.T) {
	script := `(map (fn [a] (+ a 1)) ["a"])`
	ExpectScriptErr(t, script, `operands have invalid type`)
}

func TestFilterArrayFail(t *testing.T) {
	script := `(filter (fn [a] (+ "a" 1)) ["a"])`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(filter (fn [a] (+ 1 1)) ["a"])`
	ExpectScriptErr(t, script, `filter function must return boolean`)
}

func TestFlatMapArrayFail(t *testing.T) {
	script := `(flatmap (fn [a] (+ "a" 1)) ["a"])`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(flatmap (fn [a] (+ 1 1)) ["a"])`
	ExpectScriptErr(t, script, `flatmap function must return array`)
}

func TestFilterHash(t *testing.T) {
	script := `(filter (fn [ab] (+ "a" 1)) {'a 1})`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(filter (fn [kv] (+ 1 1)) {'a 1})`
	ExpectScriptErr(t, script, `filter function must return boolean`)
}

func TestFoldlHash(t *testing.T) {
	script := `(foldl (fn [ab c] (+ "a" 1)) 0 {'a 1})`
	ExpectScriptErr(t, script, `operands have invalid type`)
}

type AnyStruct struct{ Number int }

func (AnyStruct) SexpString() string { return "" }

func TestMarshalAny(t *testing.T) {
	env := newFullEnv()
	env.Bind("g_var", AnyStruct{Number: 1024})
	script := `(= "{\"Number\":1024}" (json/stringify g_var))`
	ret, err := env.EvalString(script)
	ExpectSuccess(t, err)
	ExpectTrue(t, ret)
}

func TestCompareIntWithString(t *testing.T) {
	script := `(= 1 "a")`
	ExpectScriptErr(t, script, `cannot compare int to string`)
}

func TestIsZero(t *testing.T) {
	ExpectTrue(t, glisp.SexpBool(glisp.IsZero(glisp.SexpChar(0))))
	ExpectTrue(t, glisp.SexpBool(glisp.IsTruthy(glisp.SexpChar(0))))
	ExpectTrue(t, glisp.SexpBool(glisp.IsNumber(glisp.SexpChar(0))))
}

func TestCompareStringWithInt(t *testing.T) {
	script := `(= "a" 1)`
	ExpectScriptErr(t, script, `cannot compare string to int`)
}

func TestCompareCharWithHash(t *testing.T) {
	script := `(= #a {})`
	ExpectScriptErr(t, script, `cannot compare char to hash`)
}

func TestCompareBytesWithHash(t *testing.T) {
	script := `(= 0B6869 {})`
	ExpectScriptErr(t, script, `cannot compare bytes to hash`)
}

func TestCompareHashWithList(t *testing.T) {
	script := `(=  {} '(1))`
	ExpectScriptErr(t, script, `cannot compare hash to list`)
}

func TestCompareListWithHash(t *testing.T) {
	script := `(= '(1) {})`
	ExpectScriptErr(t, script, `cannot compare list to hash`)
}

func TestCompareBoolWithInt(t *testing.T) {
	script := `(= true 1)`
	ExpectScriptErr(t, script, `cannot compare bool to int`)
}

func TestCompareListStringWithListInt(t *testing.T) {
	script := `(= '("a") '(1))`
	ExpectScriptErr(t, script, `cannot compare string to int`)
}

func TestCompareArrayStringWithArrayInt(t *testing.T) {
	script := `(= ["a"] [1])`
	ExpectScriptErr(t, script, `cannot compare string to int`)
}

func TestCompareArrayWithInt(t *testing.T) {
	script := `(= [] 1)`
	ExpectScriptErr(t, script, `cannot compare array to int`)
}

func TestDiv0(t *testing.T) {
	script := `(/ 1 0)`
	ExpectScriptErr(t, script, `division by zero`)
}

func TestNotComparable(t *testing.T) {
	script := `(= (make-chan) 1)`
	ExpectScriptErr(t, script, `cannot compare channel to int`)
}

func TestErrorExistInList(t *testing.T) {
	script := `(exist? '("a") 1)`
	ExpectScriptErr(t, script, `cannot compare string to int`)
}

func TestConcatStringErr(t *testing.T) {
	script := `(concat "" 1)`
	ExpectScriptErr(t, script, `second argument is not a string`)
}

func TestConcatBytesErr(t *testing.T) {
	script := `(concat 0B6869 1)`
	ExpectScriptErr(t, script, `second argument is not bytes`)
}

func TestAppendStringErr(t *testing.T) {
	script := `(append "" 1)`
	ExpectScriptErr(t, script, `second argument is not a char`)
}

func TestEvalNothing(t *testing.T) {
	script := `(eval (begin))`
	ExpectScriptErr(t, script, `generating (eval (begin))`)
}

func TestExistArrayNotCompare(t *testing.T) {
	script := `(exist? [1] "a")`
	ExpectScriptErr(t, script, `cannot compare int to string`)
}

func TestApplySymbolNotFound(t *testing.T) {
	script := `(apply 'not-found-symbol "a")`
	ExpectScriptErr(t, script, `can't find function by symbol not-found-symbol`)
}

func TestJSONParse(t *testing.T) {
	script := `(json/parse (time/now))`
	ExpectScriptErr(t, script, `the first argument of json/parse must be`)
}

func TestApplyArgMustBeList(t *testing.T) {
	script := `(apply + (cons 1 2))`
	ExpectScriptErr(t, script, `expect list but got (1 . 2)`)
}

func TestDefensiveCornor(t *testing.T) {
	ExpectEqStr(t, `End`, glisp.SexpStr(glisp.SexpEnd.SexpString()))
	ExpectEqStr(t, `Marker`, glisp.SexpStr(glisp.SexpMarker.SexpString()))
	ExpectEqStr(t, ``, glisp.SexpStr(glisp.SexpSentinel(100).SexpString()))
	_, err := glisp.NewSexpIntStr("abc")
	ExpectError(t, err)

	_, err = glisp.NewSexpIntStrWithBase("abc", 12)
	ExpectError(t, err)

	_, err = glisp.NewSexpFloatStr("abc")
	ExpectError(t, err)

	ExpectEqStr(t, `1`, glisp.SexpStr(glisp.NewSexpFloat(1).SexpString()))

	_, err = glisp.NewSexpBytesByHex("世界")
	ExpectError(t, err)
}

func TestRecover(t *testing.T) {
	env := newFullEnv()
	env.PushScope()
	_, err := env.EvalString(`(def g_var 1023) (/ 1 0)`)
	ExpectError(t, err, "division by zero")

	_, ok := env.FindObject(`g_var`)
	ExpectTrue(t, glisp.SexpBool(ok))

	env.Clear()
	_, ok = env.FindObject(`g_var`)
	ExpectFalse(t, glisp.SexpBool(ok))
}

func TestScriptFunctionFail(t *testing.T) {
	env := newFullEnv()
	fn, err := env.MakeScriptFunction(`(1 2 3)`)
	ExpectSuccess(t, err)
	_, err = env.Apply(fn, glisp.MakeArgs())
	ExpectError(t, err, "not a function")
}

func TestLetListFail(t *testing.T) {
	env := newFullEnv()
	err := env.SourceStream(bytes.NewBufferString(`(let (1) 1)`))
	ExpectError(t, err, `not a function`)

	err = env.SourceStream(bytes.NewBufferString(`(let (list 1) 1)`))
	ExpectError(t, err, "bad let binding type")

	err = env.SourceStream(bytes.NewBufferString(`(let ((fn [] {1 2})) 1)`))
	ExpectError(t, err, "hash key must be symbol but got int")

	err = env.SourceStream(bytes.NewBufferString(`(let ((fn [] [1])) 1)`))
	ExpectError(t, err, "bind list length must be even")

	err = env.SourceStream(bytes.NewBufferString(`(let ((fn [] [1 2])) 1)`))
	ExpectError(t, err, "odd argument of bind list must be symbol")
}

func TestDumpEnv(t *testing.T) {
	env := newFullEnv()
	err := env.SourceStream(bytes.NewBufferString(`(let (1) 1)`))
	ExpectError(t, err)
	var buf bytes.Buffer
	env.DumpEnvironment(&buf)
	ExpectNonEmptyStr(t, buf.String())
	ExpectNonEmptyStr(t, env.GetStackTrace(err))
	var buf1 bytes.Buffer
	env.DumpFunctionByName(&buf1, "+")
	err = env.DumpFunctionByName(&buf1, "xxxxyyy")
	ExpectError(t, err, `"xxxxyyy" not found`)
}

func TestApplyByNameFail(t *testing.T) {
	env := newFullEnv()
	env.Bind("aaa", glisp.NewSexpInt(1))
	_, err := env.ApplyByName("aaa", glisp.MakeArgs())
	ExpectError(t, err, "is not a function")
}

func TestWrongArgumentNumber(t *testing.T) {
	ExpectScriptErr(t, `(gensym 1)`)
	ExpectScriptErr(t, `(sprintf)`)
	ExpectScriptErr(t, `(sprintf 1)`)
	ExpectScriptErr(t, `(sprintf 1 2)`)
	ExpectScriptErr(t, `(symbol)`)
	ExpectScriptErr(t, `(symbol 1)`)
	ExpectScriptErr(t, `(str)`)
	ExpectScriptErr(t, `(type)`)
	ExpectScriptErr(t, `(>)`)
	ExpectScriptErr(t, `(+)`)
	ExpectScriptErr(t, `(+ 1 "a")`)
	ExpectScriptErr(t, `(cons)`)
	ExpectScriptErr(t, `(car)`)
	ExpectScriptErr(t, `(cdr)`)
	ExpectScriptErr(t, `(car 1)`)
	ExpectScriptErr(t, `(cdr 1)`)
	ExpectScriptErr(t, `(aget)`)
	ExpectScriptErr(t, `(aindex)`)
	ExpectScriptErr(t, `(aget 1)`)
	ExpectScriptErr(t, `(aget 1 1)`)
	ExpectScriptErr(t, `(list?)`)
	ExpectScriptErr(t, `(aget [1] "")`)
	ExpectScriptErr(t, `(aget [1] -1)`)
	ExpectScriptErr(t, `(aset! [1] #a)`)
	ExpectScriptErr(t, `(aset! [1] 0)`)
	ExpectScriptErr(t, `(sget 0)`)
	ExpectScriptErr(t, `(sget 0 0)`)
	ExpectScriptErr(t, `(sget "" #a)`)
	ExpectScriptErr(t, `(sget "" 0B6869)`)
	ExpectScriptErr(t, `(exist?)`)
	ExpectScriptErr(t, `(exist? 1 1)`)
	ExpectScriptErr(t, `(hget {} 1 2 3 4)`)
	ExpectScriptErr(t, `(hget 0 1 2 )`)
	ExpectScriptErr(t, `(hdel! {} 1 2)`)
	ExpectScriptErr(t, `(slice 1 2 3 4)`)
	ExpectScriptErr(t, `(slice [] (make-chan) 1)`)
	ExpectScriptErr(t, `(slice [] 1 (make-chan))`)
	ExpectScriptErr(t, `(slice (make-chan) 1 1)`)
	ExpectScriptErr(t, `(slice [] 1 1)`)
	ExpectScriptErr(t, `(slice "" 1 1)`)
	ExpectScriptErr(t, `(slice 0B6869 10 100)`)
	ExpectScriptErr(t, `(slice 0B6869 10 1)`)
	ExpectScriptErr(t, `(slice 0B6869 #a 100)`)
	ExpectScriptErr(t, `(slice 0B6869 10 #b)`)
	ExpectScriptErr(t, `(len)`)
	ExpectScriptErr(t, `(len 1)`)
	ExpectScriptErr(t, `(append)`)
	ExpectScriptErr(t, `(append 1 1)`)
	ExpectScriptErr(t, `(concat)`)
	ExpectScriptErr(t, `(concat 1 1)`)
	ExpectScriptErr(t, `(read)`)
	ExpectScriptErr(t, `(read 1)`)
	ExpectScriptErr(t, `(eval)`)
	ExpectScriptErr(t, `(not)`)
	ExpectScriptErr(t, `(apply)`)
	ExpectScriptErr(t, `(apply 1 1)`)
	ExpectScriptErr(t, `(apply '+ 1)`)
	ExpectScriptErr(t, `(map)`)
	ExpectScriptErr(t, `(map 1 1)`)
	ExpectScriptErr(t, `(map + 1)`)
	ExpectScriptErr(t, `(filer)`)
	ExpectScriptErr(t, `(filer 1 1)`)
	ExpectScriptErr(t, `(make-array)`)
	ExpectScriptErr(t, `(make-array "")`)
	ExpectScriptErr(t, `(symnum)`)
	ExpectScriptErr(t, `(symnum 1)`)
	ExpectScriptErr(t, `(str)`)
	ExpectScriptErr(t, `(int)`)
	ExpectScriptErr(t, `(int '())`)
	ExpectScriptErr(t, `(string)`)
	ExpectScriptSuccess(t, `(bytes 1)`)
	ExpectScriptErr(t, `(bytes)`)
	ExpectScriptErr(t, `(float)`)
	ExpectScriptErr(t, `(float '())`)
	ExpectScriptErr(t, `(int)`)
	ExpectScriptErr(t, `(int "")`)
	ExpectScriptErr(t, `(string)`)
	ExpectScriptErr(t, `(string 1.0 "")`)
	ExpectScriptErr(t, `(round)`)
	ExpectScriptErr(t, `(round "")`)
	ExpectScriptErr(t, `(ceil)`)
	ExpectScriptErr(t, `(ceil "")`)
	ExpectScriptErr(t, `(floor)`)
	ExpectScriptErr(t, `(floor "")`)
	ExpectScriptErr(t, `(foldl)`)
	ExpectScriptErr(t, `(foldl + 1 1)`)
	ExpectScriptErr(t, `(foldl "+" 1 1)`)
	ExpectScriptErr(t, `(filter)`)
	ExpectScriptErr(t, `(filter + 1)`)
	ExpectScriptErr(t, `(filter "+" 1)`)
	ExpectScriptErr(t, `(/ 1 "")`)
	ExpectScriptErr(t, `(/ "" 1)`)
	ExpectScriptErr(t, `(/ #a "")`)
	ExpectScriptErr(t, `(hset! {} "")`)
	ExpectScriptErr(t, `(/ #a 0)`)
	ExpectScriptErr(t, `(/ 1.0 "")`)
	ExpectScriptErr(t, `(assert)`)
	ExpectScriptErr(t, `({'a})`)
	ExpectScriptErr(t, `(rand 1 2)`)
	ExpectScriptErr(t, `(rand "")`)
	ExpectScriptErr(t, `(rand 0)`)
	ExpectScriptErr(t, `(randf 1)`)
	ExpectScriptErr(t, `(json/stringify)`)
	ExpectScriptErr(t, `(json/parse)`)
	ExpectScriptErr(t, `(regexp/compile)`)
	ExpectScriptErr(t, `(regexp/compile 1)`)
	ExpectScriptErr(t, `(regexp/compile "\\")`)
	ExpectScriptErr(t, `(regexp/find)`)
	ExpectScriptErr(t, `(regexp/find 2 1)`)
	ExpectScriptErr(t, `(regexp/find 1 "a")`)
	ExpectScriptErr(t, `(regexp/replace 1 "a")`)
	ExpectScriptErr(t, `(regexp/replace 1 "a" "B")`)
	ExpectScriptErr(t, `(time/now "a")`)
	ExpectScriptErr(t, `(time/year)`)
	ExpectScriptErr(t, `(time/year 1)`)
	ExpectScriptErr(t, `(time/month)`)
	ExpectScriptErr(t, `(time/month 1)`)
	ExpectScriptErr(t, `(time/day)`)
	ExpectScriptErr(t, `(time/day 1)`)
	ExpectScriptErr(t, `(time/hour)`)
	ExpectScriptErr(t, `(time/hour 1)`)
	ExpectScriptErr(t, `(time/minute)`)
	ExpectScriptErr(t, `(time/minute 1)`)
	ExpectScriptErr(t, `(time/second)`)
	ExpectScriptErr(t, `(time/second 1)`)
	ExpectScriptErr(t, `(time/weekday)`)
	ExpectScriptErr(t, `(time/weekday 1)`)
	ExpectScriptErr(t, `(time/sub)`)
	ExpectScriptErr(t, `(time/sub 1 2)`)
	ExpectScriptErr(t, `(time/add)`)
	ExpectScriptErr(t, `(time/add 1 2 3)`)
	ExpectScriptErr(t, `(time/add (time/now) "a" 3)`)
	ExpectScriptErr(t, `(time/add (time/now) 1 "a")`)
	ExpectScriptErr(t, `(time/add (time/now) 1 1)`)
	ExpectScriptErr(t, `(time/add-date)`)
	ExpectScriptErr(t, `(time/add-date 1 2 3 4)`)
	ExpectScriptErr(t, `(time/add-date (time/now) "a" 3 4)`)
	ExpectScriptErr(t, `(time/parse)`)
	ExpectScriptErr(t, `(time/parse "xyz")`)
	ExpectScriptErr(t, `(time/parse 1.2)`)
	ExpectScriptErr(t, `(time/parse 1.2 1)`)
	ExpectScriptErr(t, `(time/parse "a" 1)`)
	ExpectScriptErr(t, `(time/parse "a" 1 2)`)
	ExpectScriptErr(t, `(time/format)`)
	ExpectScriptErr(t, `(time/format 1 2)`)
	ExpectScriptErr(t, `(time/format (time/now) 2)`)
	ExpectScriptErr(t, `(time/format (time/now) "")`)
	ExpectScriptErr(t, `(str/start-with?)`)
	ExpectScriptErr(t, `(str/start-with? 1 2)`)
	ExpectScriptErr(t, `(str/start-with? "" 2)`)
	ExpectScriptErr(t, `(str/count)`)
	ExpectScriptErr(t, `(str/count 1 2)`)
	ExpectScriptErr(t, `(str/count "" 2)`)
	ExpectScriptErr(t, `(str/split)`)
	ExpectScriptErr(t, `(str/split 1 2)`)
	ExpectScriptErr(t, `(str/split "" 2)`)
	ExpectScriptErr(t, `(str/lower)`)
	ExpectScriptErr(t, `(str/lower 1)`)
	ExpectScriptErr(t, `(str/digit?)`)
	ExpectScriptErr(t, `(str/digit? 1)`)
	ExpectScriptErr(t, `(str/join)`)
	ExpectScriptErr(t, `(str/join 1 2)`)
	ExpectScriptErr(t, `(str/join [] 2)`)
	ExpectScriptErr(t, `(str/join [1] "")`)
	ExpectScriptErr(t, `(str/join [] "" 2)`)
	ExpectScriptErr(t, `(str/join [1 "2"] "")`)
	ExpectScriptErr(t, `(str/join '(1 "2") "")`)
	ExpectScriptErr(t, `(str/trim-prefix)`)
	ExpectScriptErr(t, `(str/trim-prefix 1 2)`)
	ExpectScriptErr(t, `(str/trim-prefix "" 2)`)
	ExpectScriptErr(t, `(str/replace)`)
	ExpectScriptErr(t, `(str/replace 1 2 3)`)
	ExpectScriptErr(t, `(str/replace "" 2 3)`)
	ExpectScriptErr(t, `(str/replace "" "" 3)`)
	ExpectScriptErr(t, `(str/mask)`)
	ExpectScriptErr(t, `(str/mask 1 2 3 4)`)
	ExpectScriptErr(t, `(str/mask "" "" 3 4)`)
	ExpectScriptErr(t, `(str/mask "" 1 "" 4)`)
	ExpectScriptErr(t, `(str/mask "" 1 1 4)`)
	ExpectScriptErr(t, `(str/mask "" 1 0 4)`)
	ExpectScriptErr(t, `(str/mask "abc" 1 0 "*")`)
	ExpectScriptErr(t, `(str/mask "abc" 1 0 "")`)
	ExpectScriptErr(t, `(str/mask "abc" 1 -1 "")`)
	ExpectScriptErr(t, `(os/read-file)`)
	ExpectScriptErr(t, `(os/read-file 1)`)
	ExpectScriptErr(t, `(os/read-file "not-exist-file")`)
	ExpectScriptErr(t, `(os/open-file 1 2)`)
	ExpectScriptErr(t, `(os/open-file 1)`)
	ExpectScriptErr(t, `(:not-exist-method (os/open-file "/tmp/glisp-tmp-file")`)
	ExpectScriptErr(t, `(:write (os/open-file "/tmp/glisp-tmp-file")`)
	ExpectScriptErr(t, `(:write (os/open-file "/tmp/glisp-tmp-file" 1)`)
	ExpectScriptSuccess(t, `(assert (= "string" (type (:name (os/open-file)))))`)
	ExpectScriptErr(t, `(os/open-file "not-exist-dir/not-exist-file")`)
	ExpectScriptErr(t, `(os/write-file)`)
	ExpectScriptErr(t, `(os/write-file 1 2)`)
	ExpectScriptErr(t, `(os/write-file "./test" 2)`)

	notExistDir := `~/tmp/not-exist-dir/`
	home, _ := os.UserHomeDir()
	os.RemoveAll(filepath.Join(home, strings.TrimPrefix(notExistDir, "~")))
	ExpectScriptErr(t, fmt.Sprintf(`(os/write-file "%s" 2)`, notExistDir+"not-exit-file"))
	os.RemoveAll(filepath.Join(home, strings.TrimPrefix(notExistDir, "~")))

	ExpectScriptErr(t, `(os/remove-file)`)
	ExpectScriptErr(t, `(os/remove-file 1)`)
	ExpectScriptErr(t, `(os/file-exist?)`)
	ExpectScriptErr(t, `(os/file-exist? 1)`)
	ExpectScriptErr(t, `(flatmap)`)
	ExpectScriptErr(t, `(flatmap 1 2)`)
	ExpectScriptErr(t, `(flatmap (fn [] 1) 2)`)
	ExpectScriptErr(t, `(compose)`)
	ExpectScriptErr(t, `(compose 1 2)`)
	ExpectScriptErr(t, `(make-chan 1 2 3)`)
	ExpectScriptErr(t, `(make-chan "a")`)
	ExpectScriptErr(t, `(send!)`)
	ExpectScriptErr(t, `(send! "a")`)
	ExpectScriptErr(t, `(send! (make-chan))`)
	ExpectScriptErr(t, `(base64/encode)`)
	ExpectScriptErr(t, `(base64/encode 1)`)
	ExpectScriptErr(t, `(base64/decode)`)
	ExpectScriptErr(t, `(base64/decode 1)`)
	ExpectScriptErr(t, `(base64/decode "^&*^*&%")`)
	ExpectScriptErr(t, `(mod "" "")`)
	ExpectScriptErr(t, `(mod 1 "")`)
	ExpectScriptErr(t, `(mod 1 0)`)
	ExpectScriptErr(t, `(bit-and)`)
	ExpectScriptErr(t, `(bit-and "" "")`)
	ExpectScriptErr(t, `(bit-not)`)
	ExpectScriptErr(t, `(bit-not "")`)
	ExpectScriptErr(t, `(sla)`)
	ExpectScriptErr(t, `(sll8)`)
	ExpectScriptErr(t, `(sll8 "" "")`)
	ExpectScriptErr(t, `(sll8 1 "")`)
	ExpectScriptErr(t, `(int)`)
	ExpectScriptErr(t, `(int "")`)
	ExpectScriptErr(t, `(string)`)
	ExpectScriptErr(t, `(sexp-str)`)
	ExpectScriptErr(t, `(bool)`)
	ExpectScriptErr(t, `(char "aaa")`)
	ExpectScriptErr(t, `(char)`)
	ExpectScriptErr(t, `(char "")`)
	ExpectScriptErr(t, `(json/query)`)
	ExpectScriptErr(t, `(json/query 1 2)`)
	ExpectScriptErr(t, `(json/query {} 2)`)
	ExpectScriptErr(t, `(os/exec)`)
	ExpectScriptErr(t, `(os/exec {})`)
	ExpectScriptErr(t, `(assert (= 0 (car (os/exec "aaa"))))`)
	ExpectScriptErr(t, `(json/set)`)
	ExpectScriptErr(t, `(json/set 1 2 3)`)
	ExpectScriptErr(t, `(json/set {} 2 3)`)
	ExpectScriptErr(t, `(json/set {} "" 3)`)
	ExpectScriptErr(t, `(json/set {"a" +} "a.1" 3)`)
	ExpectScriptErr(t, `(json/set [] "a" 3)`, `strconv.Atoi: parsing`)
	ExpectScriptErr(t, `(json/set {"a" 1} "a.b" 2)`, `must set on hash/array but got int`)
	ExpectScriptErr(t, `(json/set [{"a" 1}] "0.a.b" 2)`, `must set on hash/array but got int`)
	ExpectScriptErr(t, `(json/del)`)
	ExpectScriptErr(t, `(json/del 1 2)`)
	ExpectScriptErr(t, `(json/del [] 2)`)
	ExpectScriptErr(t, `(json/del [] "a")`, `strconv.Atoi: parsing`)
	ExpectScriptErr(t, `(json/del [1] "0.a")`, `must del on hash/array but got int`)
	ExpectScriptErr(t, `(defn (list 1) [] 1)`, `bad function name`)
	ExpectScriptErr(t, `(def x 1) (x)`, `is not a function`)
	ExpectScriptErr(t, `(assert (= 1 2))`, `Assertion failed: (= 1 2)`)
	ExpectScriptErr(t, `(source-file)`, "expect 1,... argument(s) but got 0")
	ExpectScriptErr(t, `(source-file 1)`, "Expected `string`, `list`, `array` given int")
	ExpectScriptErr(t, `(source-file "not-exist-source-file")`, "no such file or directory")
	ExpectScriptErr(t, `(source-file ["not-exist-source-file"])`, "no such file or directory")
	ExpectScriptErr(t, `(source-file '("not-exist-source-file"))`, "no such file or directory")
	WithTempFile(`(`, func(file string) {
		ExpectScriptErr(t, fmt.Sprintf(`(source-file '("%s"))`, file), "Unexpected end of input")
	})
	ExpectScriptErr(t, `((compose (fn [e] e) (fn [e] (xxx))) 1)`, "symbol", "not found")

	_, err := glisp.GetConstructorFunction("")(glisp.New(), glisp.MakeArgs())
	ExpectError(t, err, "invalid constructor")
	h, _ := glisp.MakeHash(glisp.MakeArgs())
	_, err = glisp.GetHashAccessFunction("")(glisp.New(), glisp.MakeArgs(h, glisp.NewSexpInt(0)))
	ExpectSuccess(t, err)

	ExpectScriptErr(t, `(http/get)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(http/post)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(http/put)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(http/patch)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(http/delete)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(http/get -H 111)`, `-H option value must be a string but got "int"`)
	ExpectScriptErr(t, `(http/get -H)`, `-H need an argument but got nothing`)
	ExpectScriptErr(t, `(http/get -H "aaa")`, `bad format aaa, -H option value must like header:value`)
	ExpectScriptErr(t, `(http/get 1)`, `unknown option 1("int")`)
	ExpectScriptErr(t, `(http/curl -X 123 "http://127.0.0.1:9880")`, `-X Method need string but got "int"`)
	ExpectScriptErr(t, `(concat [] 1)`, `second argument(int) is not an array`)
	ExpectScriptErr(t, `(json/parse "{{")`, `json/parse: decode json fail unexpected json char {`)
	ExpectScriptErr(t, `((`, `Error on line 1,2: Unexpected end of input`)
	multiLine := "(len #`" + "\na)`)\n)"
	ExpectScriptErr(t, multiLine, `Error on line 3,1`)
	ExpectScriptErr(t, `(os/env)`, `no arguments`)
	ExpectScriptErr(t, `(os/setenv)`, `expect 2 argument(s) but got 0`)
	ExpectScriptErr(t, `(os/env 1)`, `env variable should be string but got int`)
	ExpectScriptErr(t, `(os/setenv 1 1)`, `env variable should be string but got int`)
	ExpectScriptErr(t, `(os/setenv "d" 1)`, `env variable should be string but got int`)
	ExpectScriptErr(t, `(os/setenv "" "")`, `env variable name can't be empty`)
	ExpectScriptErr(t, `(len)`, `expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(len 1)`, `argument must be string/array/list/hash/bytes but got int`)
	ExpectScriptErr(t, `(bool 1)`, `bool argument should be string/bool`)
	ExpectScriptErr(t, `(doc)`, `doc expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(doc 1)`, `should be symbol`)
	ExpectScriptErr(t, `(assert false 100)`, `100`)
	ExpectScriptErr(t, `(assert false "error message-x")`, `error message-x`)
	ExpectScriptErr(t, `(assert false (begin ((fn [a] (sprintf "--%s--" a)) "ERROR") ))`, `--ERROR--`)
	ExpectScriptErr(t, `(time/format (time/now) "2006-01-02 15:04:05" "Asia/s")`, `unknown time zone Asia/s`)
	ExpectScriptErr(t, `(json/parse nil)`, `the first argument of json/parse must be string/bytes/int/bool/float but got nil`)
	ExpectScriptErr(t, `(car [])`, `access an empty array`)
	ExpectScriptSuccess(t, `(nil? (cdr []))`)
	ExpectScriptSuccess(t, `(sexp-str (stream {}))`)
	ExpectScriptErr(t, `(stream 1)`, `type int is not streamable`)
	ExpectScriptErr(t, `(stream)`, `stream expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(stream?)`, `stream? expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(streamable?)`, `streamable? expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(map 1 (stream (my-counter 1)))`, `first argument of map must be function, but got int`)
	ExpectScriptErr(t, `(flatmap 1 (stream (my-counter 1)))`, `first argument of flatmap must be function, but got int`)
	ExpectScriptErr(t, `(realize (flatmap (fn [e] (int "x")) (stream (my-counter 1))))`, `x not number`)
	ExpectScriptErr(t, `(realize (filter (fn [e] (int "x")) (stream (my-counter 1))))`, `x not number`)
	ExpectScriptErr(t, `(realize (take (fn [e] (int "x")) (stream (my-counter 1))))`, `x not number`)
	ExpectScriptErr(t, `(realize (drop (fn [e] (int "x")) (stream (my-counter 1))))`, `x not number`)
	ExpectScriptErr(t, `(filter 1 (stream (my-counter 1)))`, `first argument of filter must be function, but got int`)
	ExpectScriptErr(t, `(take)`, `expected 2 arguments, got 0`)
	ExpectScriptErr(t, `(take 1  (my-counter 1))`, `second argument of take must be stream, but got go:*tests.Counter`)
	ExpectScriptErr(t, `(take ""  (stream (my-counter 1)))`, `first argument of take must be int/function, but got string`)
	ExpectScriptErr(t, `(drop)`, `expected 2 arguments, got 0`)
	ExpectScriptErr(t, `(realize (drop (fn [e] 1) (stream [1 2 3])))`, `drop function should return bool but got int`)
	ExpectScriptErr(t, `(realize (take (fn [e] 1) (stream [1 2 3])))`, `take function should return bool but got int`)
	ExpectScriptErr(t, `(drop 1  (my-counter 1))`, `second argument of drop must be stream, but got go:*tests.Counter`)
	ExpectScriptErr(t, `(drop ""  (stream (my-counter 1)))`, `first argument of drop must be int/function, but got string`)
	ExpectScriptErr(t, `(flatten)`, `expected 1 arguments, got 0`)
	ExpectScriptErr(t, `(flatten 1)`, `second argument of map must be array/list`)
	ExpectScriptErr(t, `(realize)`, `realize expect 1 argument(s) but got 0`)
	ExpectScriptErr(t, `(realize 1)`, `type int is not stream`)
	ExpectScriptErr(t, `(flatten)`, `flatten expected 1 arguments, got 0`)
	ExpectScriptErr(t, `(flatten 1)`, `second argument of map must be array/list`)
	ExpectScriptErr(t, `(->> (stream (my-counter 100)) (foldl #(+ %1 %2) ""))`, `operands have invalid type`)
	ExpectScriptSuccess(t, `(sexp-str (stream (my-counter 100)))`)
	ExpectScriptErr(t, `(realize (flatten (stream [1])))`, `element(int) is not streamable`)
	ExpectScriptErr(t, `(realize (flatmap (fn [e] e) (stream [1])))`, `flatmap element(int) is not streamable`)
	ExpectScriptErr(t, `(realize (filter (fn [e] 1) (stream [1])) )`, `filter function should return bool but got int`)
	ExpectScriptErr(t, `(->> (err-stream "error-stream") (take 1) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (err-stream "error-stream") (take (fn [e] true)) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (err-stream "error-stream") (drop 1) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (err-stream "error-stream") (drop (fn [e] true)) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (stream [(err-stream "error-stream")]) (flatmap (fn [e] e)) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (stream [(err-stream "error-stream")]) (flatten) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(->> (err-stream "error-stream") (foldl (fn [e c] 1) 0) (realize))`, `error-stream`)
	ExpectScriptErr(t, `(range 1 1 1 1)`, `range expect 0,1,2,3 argument(s) but got 4`)
	ExpectScriptErr(t, `(range "x" 1 1 1)`, `all arguments of range must be int but got string`)
	ExpectScriptErr(t, `(partition)`, `partition expect 2,3 argument(s) but got 0`)
	ExpectScriptErr(t, `(partition 1 1)`, `last argument of partition must be stream, but got int`)
	ExpectScriptErr(t, `(partition "s" (range))`, `first argument of partition must be int/function, but got string`)
	ExpectScriptErr(t, `(partition "s" 1 (range))`, `first argument of partition must be function, but got string`)
	ExpectScriptErr(t, `(partition (fn [] 1) 1 (range))`, `second argument of partition must be bool, but got int`)
	ExpectScriptErr(t, `(realize (partition 1 (err-stream "x")))`, `Error calling realize: x`)
	ExpectScriptErr(t, `(realize (partition (fn [e] 1) (err-stream "x")))`, `Error calling realize: x`)
	ExpectScriptErr(t, `(realize (partition (fn [e] (int "x")) (range 3)))`, `x not number`)
	ExpectScriptErr(t, `(realize (partition (fn [e] (int "1")) (range 3)))`, `partition function must return bool but get int`)
	ExpectScriptErr(t, `(zip)`, `zip expect 2,... argument(s) but got 0`)
	ExpectScriptErr(t, `(zip 1 1)`, `every argument of zip must be stream but 1-th is int`)
	ExpectScriptSuccess(t, `(sexp-str (zip (range 1) (range 2)))`)
	ExpectScriptErr(t, `(os/run)`, `no command arguments`)
	ExpectScriptErr(t, `(os/run 1)`, `cmd must be string`)
	ExpectScriptSuccess(t, `(os/run "echo x >/dev/null")`)
	ExpectScriptErr(t, `(sort)`, `sort expect 1,2 argument(s) but got 0`)
	ExpectScriptErr(t, `(sort 1 2)`, `first argument must be function but got`)
	ExpectScriptErr(t, `(sort + 2)`, `second argument must be array/list but got`)
	ExpectScriptErr(t, `(union)`, `union expect 2,... argument(s) but got 0`)
	ExpectScriptErr(t, `(union 1 1)`, `expected strings/arrays/lists/bytes but got int`)
	ExpectScriptErr(t, `(union (range 1 2) 1)`, `every argument of union must be stream/streamable but 2-th is int`)
	ExpectScriptErr(t, `(realize (union (range 3) (err-stream)))`, `error occur`)
	ExpectScriptErr(t, `(time/zero 1)`, `time/zero expect 0 argument(s) but got 1`)
	ExpectScriptErr(t, `#'100`, `sharp-quote 100`)
	ExpectScriptErr(t, `#'(int 1)`, `is not a symbol`)
	ExpectScriptErr(t, `#'(string "xyzw")`, `is not a symbol`)
	ExpectScriptSuccess(t, `#'(symbol "xyzw")`)
	ExpectScriptErr(t, `(:xyzgfw)`, `expect 2,... argument(s) but got 1`)
	ExpectScriptErr(t, `(:xzyfewfw true false)`, "type `bool` can't explain")
	ExpectScriptErr(t, `(:fjslfsdf "s" false)`, "type `string` can't explain")
	ExpectScriptSuccess(t, `(:s (my-counter 10))`, "OK")
	ExpectScriptSuccess(t, `(:colon-symbol (my-counter 10))`, "OK")
	ExpectScriptErr(t, `(def h {"a" 1 'b 2}) (:a h 1 2)`, `hash field accessor expect 0,1 argument(s) but got 2`)
	ExpectScriptErr(t, `(def h {"a" 1 'b 2}) (:c h)`, `field c not found`)
	ExpectScriptErr(t, `(def h {"a" 1 'b 2}) (:c h)`, `field c not found`)
	ExpectScriptErr(t, `(defrecord)`, `expect 1,... argument(s) but got 0`)
	ExpectScriptErr(t, `(defrecord "sw")`, `first argument must be symbol`)
	ExpectScriptErr(t, `(defrecord X 1)`, `field definition should be list`)
	ExpectScriptErr(t, `(defrecord X (1))`, `field definition format must be`)
	ExpectScriptErr(t, `(defrecord X (1 2))`, `field definition format must be`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (->Class T)`, `count must be even`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (->Class "Name" 1)`, `field name must be symbol`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (->Class Name1 1)`, `not contains a field`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (->Class Name 1)`, `expect string but got int`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (:Name p 2 3)`, `record field accessor need not more than one argument`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (:Name1 p)`, `not have a field named`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc)`, `expect 3 argument(s)`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc 1 Name 3)`, `first argument must be record`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc p 2 333)`, `second argument must be symbol`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc p Name 12)`, `expect string but got int`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc p Name2 12)`, `not found`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (assoc p 123 12)`, `second argument must be symbol/string`)
	ExpectScriptErr(t, `(defrecord Class (Name hash<string,int>)) (def p (->Class Name "Tom"))`, `expect hash<string,int> but got string`)
	ExpectScriptErr(t, `(defrecord Class (Name hash<string,int>)) (def p (->Class Name {"a" "b"}))`, `expect int but got string`)
	ExpectScriptErr(t, `(defrecord Class (Name hash<string,int>)) (def p (->Class Name {1 "b"}))`, `expect string but got int`)
	ExpectScriptErr(t, `(defrecord Class (Name list<int>)) (def p (->Class Name {1 "b"}))`, `expect list<int> but got hash`)
	ExpectScriptErr(t, `(record?)`, `expect 1 argument(s)`)
	ExpectScriptErr(t, `(record-of?)`, `expect 2 argument(s)`)
	ExpectScriptErr(t, `(record-class-definition)`, `expect 1 argument(s)`)
	ExpectScriptErr(t, `(record-class-definition 1)`, `first argument must be record class but got int`)
	ExpectScriptErr(t, `(record-class?)`, `expect 1 argument(s)`)
	ExpectScriptErr(t, `(record-of? 1 2)`, `first argument must be record`)
	ExpectScriptErr(t, `(defrecord Class (Name string)) (def p (->Class Name "Tom")) (record-of? p 1)`, `second argument must be record class`)
	ExpectScriptErr(t, `(os/read-dir 1)`, `argument should be string`)
	ExpectScriptErr(t, `(os/read-dir)`, `expect 1,2 argument(s) but got 0`)
	ExpectScriptErr(t, `(str/repeat)`, `expect 2 argument(s)`)
	ExpectScriptErr(t, `(str/repeat 1 1)`, `first argument should be string`)
	ExpectScriptErr(t, `(str/repeat "" "")`, `second argument should be int`)
	ExpectScriptErr(t, `(os/mkdir)`, `expect 1 argument but got`)
	ExpectScriptErr(t, `(os/mkdir 1)`, `argument should be string`)
	ExpectScriptSuccess(t, `(os/mkdir "this-is-a-test-dir")`)
	ExpectScriptSuccess(t, `(os/remove-file "this-is-a-test-dir")`)
	ExpectScriptSuccess(t, `(os/read-dir "." 'file)`)
	ExpectScriptSuccess(t, `(os/read-dir "." 'dir)`)
	WithTempFile(`a,b,c`, func(file string) {
		ExpectScriptErr(t, fmt.Sprintf(`(csv/read "%s" 1 2)`, file), "csv/read expect 1,2 argument but got 3")
		ExpectScriptErr(t, fmt.Sprintf(`(csv/read 1 "%s")`, file), "csv/read 1st argument should be string/reader but got int")
		ExpectScriptErr(t, fmt.Sprintf(`(csv/read "xxxxxx")`), "no such file or directory")
	})
	WithTempFile(`a,b,c`, func(file string) {
		ExpectScriptErr(t, fmt.Sprintf(`(csv/write "%s" 1 2)`, file), "csv/write expect 1,2 argument but got 3")
		ExpectScriptErr(t, fmt.Sprintf(`(csv/write 1 "%s")`, file), "csv/write 1st argument should be string/writer but got int")
		ExpectScriptErr(t, fmt.Sprintf(`(csv/write "%s" 1)`, file), "csv/write 2nd argument should be [][]string but got int")
		ExpectScriptErr(t, fmt.Sprintf(`(csv/write "%s" [{"a" 1} 1])`, file), "csv row should be hash but got int")
	})
	ExpectScriptErr(t, `(> (time/now) 1)`, `cannot compare time to int`)
	ExpectScriptErr(t, `(time/format (time/now) "2006-01-02 15:04:05" 1)`, `time/format: third argument of function time/format must be string but got int`)
	ExpectScriptErr(t, `(time/parse 1 2)`, `time/parse with unsupported argument 2`)
	ExpectScriptErr(t, `(time/parse "2014-Feb-04" 1)`, `time/parse with unsupported argument`)
	ExpectScriptErr(t, `(time/parse "2014-Feb-04" "abc")`, `parse time 2014-Feb-04 with layout abc fail`)
	ExpectScriptErr(t, `(time/parse 1 "abc")`, `time/parse with unsupported argument`)
	ExpectScriptErr(t, `(time/parse "2014-Feb-04" "2006-Jan-02" 1)`, `time/parse with unsupported argument`)
	ExpectScriptErr(t, `(time/parse "2014-Feb-04" "2006-Jan-02" "ak")`, `time/parse: unknown time zone`)
}
