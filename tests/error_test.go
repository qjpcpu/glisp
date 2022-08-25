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
	ExpectScriptErr(t, script, `cannot compare 1.0<float> to "a"<string>`)
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
	script := `(filter (fn [a b] (+ "a" 1)) {'a 1})`
	ExpectScriptErr(t, script, `operands have invalid type`)
	script = `(filter (fn [k v] (+ 1 1)) {'a 1})`
	ExpectScriptErr(t, script, `filter function must return boolean`)
}

func TestFoldlHash(t *testing.T) {
	script := `(foldl (fn [a b c] (+ "a" 1)) 0 {'a 1})`
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
	ExpectScriptErr(t, script, `cannot compare 1<int> to "a"<string>`)
}

func TestIsZero(t *testing.T) {
	ExpectTrue(t, glisp.SexpBool(glisp.IsZero(glisp.SexpChar(0))))
	ExpectFalse(t, glisp.SexpBool(glisp.IsTruthy(glisp.SexpChar(0))))
	ExpectTrue(t, glisp.SexpBool(glisp.IsNumber(glisp.SexpChar(0))))
}

func TestCompareStringWithInt(t *testing.T) {
	script := `(= "a" 1)`
	ExpectScriptErr(t, script, `cannot compare "a"<string> to 1<int>`)
}

func TestCompareCharWithHash(t *testing.T) {
	script := `(= #a {})`
	ExpectScriptErr(t, script, `cannot compare #a<char> to {}<hash>`)
}

func TestCompareBytesWithHash(t *testing.T) {
	script := `(= 0B6869 {})`
	ExpectScriptErr(t, script, `cannot compare 0B6869<bytes> to {}<hash>`)
}

func TestCompareHashWithList(t *testing.T) {
	script := `(=  {} '(1))`
	ExpectScriptErr(t, script, `cannot compare {}<hash> to (1)<list>`)
}

func TestCompareListWithHash(t *testing.T) {
	script := `(= '(1) {})`
	ExpectScriptErr(t, script, `cannot compare (1)<list> to {}<hash>`)
}

func TestCompareBoolWithInt(t *testing.T) {
	script := `(= true 1)`
	ExpectScriptErr(t, script, `cannot compare true<bool> to 1<int>`)
}

func TestCompareListStringWithListInt(t *testing.T) {
	script := `(= '("a") '(1))`
	ExpectScriptErr(t, script, `cannot compare "a"<string> to 1<int>`)
}

func TestCompareArrayStringWithArrayInt(t *testing.T) {
	script := `(= ["a"] [1])`
	ExpectScriptErr(t, script, `cannot compare "a"<string> to 1<int>`)
}

func TestCompareArrayWithInt(t *testing.T) {
	script := `(= [] 1)`
	ExpectScriptErr(t, script, `cannot compare []<array> to 1<int>`)
}

func TestDiv0(t *testing.T) {
	script := `(/ 1 0)`
	ExpectScriptErr(t, script, `division by zero`)
}

func TestNotComparable(t *testing.T) {
	script := `(= (make-chan) 1)`
	ExpectScriptErr(t, script, `cannot compare [chan]<channel> to 1<int>`)
}

func TestErrorExistInList(t *testing.T) {
	script := `(exist? '("a") 1)`
	ExpectScriptErr(t, script, `cannot compare "a"<string> to 1<int>`)
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

func TestHashKey(t *testing.T) {
	script := `(exist? {} {})`
	ExpectScriptErr(t, script, `cannot hash type {}<hash>`)
}

func TestEvalNothing(t *testing.T) {
	script := `(eval (begin))`
	ExpectScriptErr(t, script, `generating (eval (begin))`)
}

func TestExistArrayNotCompare(t *testing.T) {
	script := `(exist? [1] "a")`
	ExpectScriptErr(t, script, `cannot compare 1<int> to "a"<string>`)
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
	ExpectScriptErr(t, script, `expect list but got (1 . 2)<list>`)
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
	env.PushGlobalScope()
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
	fn, err := env.MakeScriptFunction(`(1 2 3)`, ``)
	ExpectSuccess(t, err)
	_, err = env.Apply(fn, []glisp.Sexp{})
	ExpectError(t, err, "not a function")
}

func TestLetListFail(t *testing.T) {
	env := newFullEnv()
	err := env.SourceStream(bytes.NewBufferString(`(let (1) 1)`))
	ExpectError(t, err, `not a function`)

	err = env.SourceStream(bytes.NewBufferString(`(let (list 1) 1)`))
	ExpectError(t, err, "not an array")

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
	_, err := env.ApplyByName("aaa", nil)
	ExpectError(t, err, "is not a function")
}

func TestWrongArgumentNumber(t *testing.T) {
	ExpectScriptErr(t, `(gensym 1)`)
	ExpectScriptErr(t, `(sprintf)`)
	ExpectScriptErr(t, `(sprintf 1)`)
	ExpectScriptErr(t, `(sprintf 1 2)`)
	ExpectScriptErr(t, `(str2sym)`)
	ExpectScriptErr(t, `(str2sym 1)`)
	ExpectScriptErr(t, `(sym2str)`)
	ExpectScriptErr(t, `(sym2str 1)`)
	ExpectScriptErr(t, `(typestr)`)
	ExpectScriptErr(t, `(>)`)
	ExpectScriptErr(t, `(+)`)
	ExpectScriptErr(t, `(+ 1 "a")`)
	ExpectScriptErr(t, `(cons)`)
	ExpectScriptErr(t, `(car)`)
	ExpectScriptErr(t, `(cdr)`)
	ExpectScriptErr(t, `(car 1)`)
	ExpectScriptErr(t, `(cdr 1)`)
	ExpectScriptErr(t, `(aget)`)
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
	ExpectScriptErr(t, `(str2int)`)
	ExpectScriptErr(t, `(str2int 1)`)
	ExpectScriptErr(t, `(bytes2str)`)
	ExpectScriptErr(t, `(bytes2str 1)`)
	ExpectScriptErr(t, `(str2bytes 1)`)
	ExpectScriptErr(t, `(str2bytes)`)
	ExpectScriptErr(t, `(str2float)`)
	ExpectScriptErr(t, `(str2float 1)`)
	ExpectScriptErr(t, `(float2int)`)
	ExpectScriptErr(t, `(float2int "")`)
	ExpectScriptErr(t, `(float2str)`)
	ExpectScriptErr(t, `(float2str {})`)
	ExpectScriptErr(t, `(float2str 1.0 "")`)
	ExpectScriptErr(t, `(round)`)
	ExpectScriptErr(t, `(round "")`)
	ExpectScriptErr(t, `(foldl)`)
	ExpectScriptErr(t, `(foldl + 1 1)`)
	ExpectScriptErr(t, `(foldl "+" 1 1)`)
	ExpectScriptErr(t, `(filter)`)
	ExpectScriptErr(t, `(filter + 1)`)
	ExpectScriptErr(t, `(filter "+" 1)`)
	ExpectScriptErr(t, `(/ 1 "")`)
	ExpectScriptErr(t, `(/ "" 1)`)
	ExpectScriptErr(t, `(/ #a "")`)
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
	ExpectScriptErr(t, `(regexp-compile)`)
	ExpectScriptErr(t, `(regexp-compile 1)`)
	ExpectScriptErr(t, `(regexp-compile "\\")`)
	ExpectScriptErr(t, `(regexp-find)`)
	ExpectScriptErr(t, `(regexp-find "h" 1)`)
	ExpectScriptErr(t, `(regexp-find "h" "a")`)
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
	ExpectScriptErr(t, `(char2int)`)
	ExpectScriptErr(t, `(char2int "")`)
	ExpectScriptErr(t, `(char2str)`)
	ExpectScriptErr(t, `(char2str "")`)
	ExpectScriptErr(t, `(int2char)`)
	ExpectScriptErr(t, `(int2char "")`)
	ExpectScriptErr(t, `(json/query)`)
	ExpectScriptErr(t, `(json/query 1 2)`)
	ExpectScriptErr(t, `(json/query {} 2)`)
	ExpectScriptErr(t, `(os/exec)`)
	ExpectScriptErr(t, `(os/exec {})`)
	ExpectScriptErr(t, `(os/exec "aaa")`)
	ExpectScriptErr(t, `(json/set)`)
	ExpectScriptErr(t, `(json/set 1 2 3)`)
	ExpectScriptErr(t, `(json/set {} 2 3)`)
	ExpectScriptErr(t, `(json/set {} "" 3)`)
	ExpectScriptErr(t, `(json/set {"a" +} "a.1" 3)`)
	ExpectScriptErr(t, `(json/set [] "a" 3)`, `strconv.Atoi: parsing`)
	ExpectScriptErr(t, `(json/set {"a" 1} "a.b" 2)`, `must set on hash/array but got 1<int>`)
	ExpectScriptErr(t, `(json/set [{"a" 1}] "0.a.b" 2)`, `must set on hash/array but got 1<int>`)
	ExpectScriptErr(t, `(json/del)`)
	ExpectScriptErr(t, `(json/del 1 2)`)
	ExpectScriptErr(t, `(json/del [] 2)`)
	ExpectScriptErr(t, `(json/del [] "a")`, `strconv.Atoi: parsing`)
	ExpectScriptErr(t, `(json/del [1] "0.a")`, `must del on hash/array but got 1<int>`)
	ExpectScriptErr(t, `(defn (list 1) [] 1)`, `bad function name`)
	ExpectScriptErr(t, `(def x 1) (x)`, `is not a function`)
	ExpectScriptErr(t, `(assert (= 1 2))`, `Assertion failed: (= 1 2)`)
	ExpectScriptErr(t, `(source-file)`, "expect 1,... argument but got 0")
	ExpectScriptErr(t, `(source-file 1)`, "Expected `string`, `list`, `array` given 1<int>")
	ExpectScriptErr(t, `(source-file "not-exist-source-file")`, "no such file or directory")
	ExpectScriptErr(t, `(source-file ["not-exist-source-file"])`, "no such file or directory")
	ExpectScriptErr(t, `(source-file '("not-exist-source-file"))`, "no such file or directory")
	WithTempFile(`(`, func(file string) {
		ExpectScriptErr(t, fmt.Sprintf(`(source-file '("%s"))`, file), "Error on line 1: Unexpected end of input")
	})
	ExpectScriptErr(t, `((compose (fn [e] e) (fn [e] (xxx))) 1)`, "symbol", "not found")

	_, err := glisp.GetConstructorFunction("")(glisp.New(), nil)
	ExpectError(t, err, "invalid constructor")
	h, _ := glisp.MakeHash(nil)
	_, err = glisp.GetHashAccessFunction("")(glisp.New(), []glisp.Sexp{h, glisp.NewSexpInt(0)})
	ExpectSuccess(t, err)

	ExpectScriptErr(t, `(http/get)`, `expect 1,... argument but got 0`)
	ExpectScriptErr(t, `(http/post)`, `expect 1,... argument but got 0`)
	ExpectScriptErr(t, `(http/put)`, `expect 1,... argument but got 0`)
	ExpectScriptErr(t, `(http/patch)`, `expect 1,... argument but got 0`)
	ExpectScriptErr(t, `(http/delete)`, `expect 1,... argument but got 0`)
	ExpectScriptErr(t, `(http/get "-H" 111)`, `-H option value must be a string but got "int"`)
	ExpectScriptErr(t, `(http/get '-H)`, `-H need an argument but got nothing`)
	ExpectScriptErr(t, `(http/get '-H "aaa")`, `bad format aaa, -H option value must like header:value`)
	ExpectScriptErr(t, `(http/get 1)`, `unknown option 1("int")`)
	ExpectScriptErr(t, `(http/curl '-X 123 "http://127.0.0.1:9880")`, `-X Method need string but got "int"`)
}
