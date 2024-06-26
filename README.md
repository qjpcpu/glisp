# GLisp

![Image text](./tests/codcov.svg)

This is a LISP dialect designed as an embedded extension language for the Go
programming language. It is implemented in pure Go, so it can be easily ported
to all systems and architectures that Go targets.

Here is a list of what features are implemented and not implemented so far.

 * [x] Float, Int, Char, String, Bytes, Symbol, List, Array, and Hash datatypes
 * [x] Arithmetic (`+`, `-`, `*`, `/`, `mod`)
 * [x] Shift Operators (`sla`, `sra`)
 * [x] Big integer
 * [x] Bitwise operations (`bit-and`, `bit-or`, `bit-xor`)
 * [x] Comparison operations (`<`, `>`, `<=`, `>=`, `=`, and `not=`)
 * [x] Short-circuit boolean operators (`and` and `or`)
 * [x] Conditionals (`cond`)
 * [x] A Basic Repl
 * [x] Lambdas (`fn`)
 * [x] Bindings (`def`, `defn`, and `let`)
 * [x] Tail-call optimization
 * [x] Go API
 * [x] Macro System
 * [x] Syntax quoting (backticks)
 * [x] Channel and goroutine support

The full documentation can be found in the [Wiki](https://github.com/qjpcpu/glisp/wiki).
