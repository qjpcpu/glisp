package glisp

import (
	"errors"
	"fmt"
)

// Opcode is the type for virtual machine instructions.
type Opcode byte

const (
	// OpInvalid is an invalid instruction
	OpInvalid Opcode = iota

	// Stack manipulation
	OpPush
	OpPushClosure
	OpPop
	OpDup

	// Variable access
	OpGet
	OpPut
	OpBindDynFun

	// Control flow
	OpJump     // Unconditional relative jump
	OpGoto     // Unconditional absolute jump
	OpBranch   // Conditional relative jump
	OpReturn   // Return from function
	OpCall     // Call a function by symbol
	OpPrepare  // Prepare for tail call
	OpDispatch // Call a function from stack

	// Scope
	OpAddScope
	OpRemoveScope

	// Special instructions
	OpExplode
	OpSquash
	OpVectorize
	OpHashize
	OpBindlist
	OpRefSym
)

// Instruction represents a single bytecode instruction for the VM.
// Instead of an interface, it's a concrete struct.
type Instruction struct {
	Op Opcode

	// Operands for different instructions
	Expr       Sexp          // For OpPush
	ClosedFunc *SexpFunction // For OpPushClosure
	Sym        SexpSymbol    // For OpGet, OpPut, OpCall, OpPrepare
	IsSet      bool          // For OpPut
	Nargs      int           // For OpCall, OpPrepare, OpDispatch
	Loc        int           // For OpJump, OpGoto, OpBranch
	Direction  bool          // For OpBranch
	Err        error         // For OpReturn
	DynamicErr bool          // For OpReturn
}

// InstrString provides a human-readable representation of the instruction.
func (i Instruction) InstrString() string {
	switch i.Op {
	case OpPush:
		return "push " + i.Expr.SexpString()
	case OpPushClosure:
		return "pushC " + i.ClosedFunc.SexpString()
	case OpPop:
		return "pop"
	case OpDup:
		return "dup"
	case OpGet:
		return fmt.Sprintf("get %s", i.Sym.name)
	case OpPut:
		return fmt.Sprintf("put %s", i.Sym.name)
	case OpBindDynFun:
		return "bind dynamic function"
	case OpJump:
		return fmt.Sprintf("jump %d", i.Loc)
	case OpGoto:
		return fmt.Sprintf("goto %d", i.Loc)
	case OpBranch:
		format := "brn %d"
		if i.Direction {
			format = "br %d"
		}
		return fmt.Sprintf(format, i.Loc)
	case OpReturn:
		if i.Err == nil {
			return "ret"
		}
		return "ret \"" + i.Err.Error() + "\""
	case OpCall:
		return fmt.Sprintf("call %s %d", i.Sym.name, i.Nargs)
	case OpPrepare:
		return fmt.Sprintf("preparecall %s %d", i.Sym.name, i.Nargs)
	case OpDispatch:
		return fmt.Sprintf("dispatch %d", i.Nargs)
	case OpAddScope:
		return "add scope"
	case OpRemoveScope:
		return "rem scope"
	case OpExplode:
		return "explode"
	case OpSquash:
		return "squash"
	case OpVectorize:
		return "vectorize"
	case OpHashize:
		return "hashize"
	case OpBindlist:
		return "bindlist"
	case OpRefSym:
		return "ref symbol"
	default:
		return "invalid"
	}
}

var OutOfBounds error = errors.New("jump out of bounds")
