# GLisp

![Coverage Status](./tests/codcov.svg)

GLISP is a dialect of LISP designed as an embedded extension language for Go. It is implemented in pure Go as a bytecode interpreter. As a result, the interpreter can be compiled or cross-compiled for any platform Go runs on.

## Features

*   [x] **Rich Data Types**: Float, Int, Char, String, Bytes, Symbol, List, Array, and Hash.
*   [x] **Comprehensive Operators**:
    *   Arithmetic: `+`, `-`, `*`, `/`, `mod`
    *   Shift: `sla`, `sra`
    *   Bitwise: `bit-and`, `bit-or`, `bit-xor`
*   [x] **Big Integer Support**
*   [x] **Control Flow**:
    *   Comparison: `<`, `>`, `<=`, `>=`, `=`, `not=`
    *   Short-circuit booleans: `and`, `or`
    *   Conditionals: `cond`
*   [x] **Core LISP Functionality**:
    *   REPL (Read-Eval-Print Loop)
    *   Lambdas (`fn`) and Bindings (`def`, `defn`, `let`)
    *   Tail-call optimization
    *   Powerful Macro System with syntax quoting (backticks)
*   [x] **Go Integration**:
    *   Seamless Go API
    *   Concurrency support with Channels and Goroutines

---

## Performance

To evaluate the performance of GLisp, we have created a separate benchmark project: [glisp-benchmark](https://github.com/qjpcpu/glisp-benchmark).

This project benchmarks GLisp against other popular embedded scripting languages for Go, including:

*   [goja](https://github.com/dop251/goja) (JavaScript)
*   [go-lua](https://github.com/Shopify/go-lua) (Lua)
*   [zygo](https://github.com/glycerine/zygomys) (Lisp)

The benchmarks cover various scenarios such as Fibonacci calculation, function calls, closures, and concurrency. The results show that GLisp demonstrates competitive performance across many test cases, making it a solid choice for performance-sensitive applications.

We encourage you to check out the benchmark project for detailed performance data and testing methodologies.

---

## Language Guide

### Reader Syntax

#### Atoms
GLISP supports seven atomic types: ints, floats, strings, chars, bools, bytes, and symbols.

```clojure
; Numbers
3          ; an int
-21        ; a negative int
0x41       ; hexadecimal int
0o755      ; octal int
0b1110     ; binary int
4.1        ; a float
1.3e20     ; float in scientific notation

; Characters
#c         ; the character 'c'
#\n       ; the newline character

; Strings
"hello world" ; a string
#`raw string`  ; a raw string literal

; Other Atoms
asdf       ; a symbol
true       ; boolean true
false      ; boolean false
0B676c...  ; byte stream (hex encoded)
```
*Semicolons (`;`) are used for single-line comments.*

#### Collections

**Lists**: Standard cons-cell lists, delimited by parentheses.
```clojure
(a-function arg1 (another-function arg2))
'(1 2 3)      ; A quoted list
(1 . 2)      ; A cons pair
```

**Arrays**: Correspond to Go slices, delimited by square brackets.
```clojure
[1 2 3 4]
```

**Hashes**: Mutable hashmaps (Go maps internally), delimited by curly braces.
```clojure
{'a 3 'b 2}  ; Maps symbol 'a' to 3 and 'b' to 2
```

#### Quoting
The quote symbol (`'`) prevents evaluation, treating the following expression as data.
```clojure
'(+ 1 2)   ; results in the list (+ 1 2), not the number 3
'a-symbol  ; results in the symbol a-symbol
```

### Functions, Bindings, and Control Flow

#### Functions (`fn`, `defn`)
Define anonymous functions with `fn` and named functions with `defn`.

```clojure
; Anonymous function
(fn [a b] (+ a b))

; Named function
(defn add-three [a] (+ a 3))

; Shorthand lambda syntax
#(+ 6 %)     ; Equivalent to: (fn [x] (+ 6 x))
#(+ %1 %2)   ; Equivalent to: (fn [x y] (+ x y))
```

#### Bindings (`def`, `let`, `set!`)
- `def`: Creates a binding in the current scope.
- `let` / `let*`: Creates a new scope with local bindings. `let*` allows later bindings to refer to earlier ones.
- `set!`: Modifies an existing binding, searching up the scope stack if necessary. Use with care.

```clojure
(def a 3)

(let [a 3 b 4] (* a b)) ; returns 12

(let* [a 2 b (+ a 1)] (+ a b)) ; returns 5
```

#### Conditionals (`cond`, `and`, `or`)
`cond` is the primary conditional form. `and` and `or` provide short-circuit evaluation.
In GLISP, `false` and `nil` (`'()`) are "falsy". All other values are "truthy".

```clojure
(cond
    (< x 0) "negative"
    (> x 0) "positive"
    "zero") ; default case

(and truthy-val falsy-val) ; returns falsy-val
(or falsy-val truthy-val)  ; returns truthy-val
```

### Macros (`defmac`)
Macros enable syntactic extension by transforming code at compile time. They are defined with `defmac`.
- `` ` `` (syntax-quote): Creates a template for code expansion.
- `~` (unquote): Evaluates an expression within a syntax-quoted template.
- `~@` (splicing-unquote): Splices a list's elements into the template.

```clojure
(defmac when [predicate & body]
  `(cond ~predicate
      (begin
        ~@body)
      '()))
```

### Built-in Functions

GLISP provides a rich set of built-in functions. Here is a partial list.

- **Type Introspection**: `list?`, `array?`, `number?`, `string?`, `symbol?`, `type`, etc.
- **Generic Collections**: `len`, `append`, `concat`, `slice`, `exist?`.
- **List Operations**: `cons`, `car`, `cdr`, `list`.
- **Array Operations**: `make-array`, `aget`, `aset!`.
- **Hashmap Operations**: `hget`, `hset!`, `hdel!`.
- **Higher-Order Functions**: `map`, `flatmap`, `filter`, `foldl`, `compose`.
- **Symbolic**: `gensym`, `symbol`, `str`.
- **Printing**: `println`, `print`, `printf`.
- **Sequencing**: `begin`.

### Extensions

GLISP's functionality can be extended with modules.

- **Core**: Arithmetic (`+`, `-`, `*`, `/`), comparisons (`<`, `=`, `>`), metaprogramming (`alias`, `currying`, `override`), threading macros (`->`, `->>`).
- **Math**: Bitwise operations (`bit-and`, `sla`), float functions (`round`, `floor`).
- **JSON**: `json/parse`, `json/stringify`, and powerful querying with `json/query`.
- **Strings**: `str/split`, `str/contains?`, `str/replace`, `str/trim-space`, and more.
- **Time**: `time/now`, `time/parse`, `time/format`, and time arithmetic.
- **OS**: `os/read-file`, `os/write-file`, `os/exec`.
- **HTTP**: `http/get`, `http/post`, `http/curl` for making HTTP requests.
- **And more**: Base64, Random, Regexp, etc.

---

## Go API Guide

### Getting Started with the Go API

To embed GLISP in your Go application, start by creating a `glisp.Environment`.

```go
package main

import (
	"fmt"
	"github.com/qjpcpu/glisp"
	ext  "github.com/qjpcpu/glisp/extensions"
)

func main() {
	// 1. Create a new interpreter environment
	env := glisp.New()
	ext.ImportCoreUtils(env) // Load standard library

	// 2. Load and compile GLISP code
	err := env.LoadString(`(defn add [a b] (+ a b))`)
	if err != nil {
		panic(err)
	}

	// 3. Run the code
	_, err = env.Run()
	if err != nil {
		panic(err)
	}

	// 4. Call a LISP function from Go
	addFunc, err := env.FindFunction("add")
	if err != nil {
		panic(err)
	}

	args := glisp.MakeArgs(glisp.SexpInt(10), glisp.SexpInt(20))
	result, err := glisp.Apply(addFunc, args)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result of (add 10 20) is: %s", result.SexpString())
}
```

### GLISP Types in Go

GLISP values are represented by the `Sexp` interface. Most types map directly to Go types (`SexpInt` -> `int64`, `SexpStr` -> `string`), while others are special structs (`SexpPair`, `SexpHash`).

### Extending GLISP with Go Functions

You can expose Go functions to your GLISP environment. A Go function must have the following signature:

```go
func MyGoFunction(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
    if args.Len() != 2 {
        return glisp.SexpNull, glisp.WrongNargs
    }
    // ... your logic here ...
    return glisp.SexpInt(42), nil
}
```

Register it in the environment:
```go
env.AddFunction("my-go-function", MyGoFunction)
```

### Error Handling

If `Run()` or `Apply()` returns an error, the environment's state is compromised. You can get a stack trace with `GetStackTrace()` and must call `Clear()` to reset the VM before running more code.

```go
expr, err := env.Run()
if err != nil {
    stackTrace := env.GetStackTrace(err)
    fmt.Println(stackTrace)
    env.Clear() // Reset the environment
}
```

## Running the REPL

To start the interactive REPL:
```bash
cd ./repl/glisp
go build && ./glisp
```

Inside the REPL, you can use `(doc function-name)` to get documentation for any function.
```
glisp> (doc map)

Usage: (map f coll)

Returns a sequence consisting of the result of applying f to
the set of first items of each coll, followed by applying f to the
set of second items in each coll, until any one of the colls is
exhausted. coll can be array or list.
```
