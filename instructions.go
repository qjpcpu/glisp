package glisp

import (
	"errors"
	"fmt"
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
	UserInstr  userInstrData // For OpUserInstr
}

type userInstrData struct {
	name      string
	nargs     int
	userinstr UserInstruction
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
	case OpUserInstr:
		return i.UserInstr.name
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

type UserInstrContext struct {
	*Environment
	nargs int
	name  string
}

type UserInstruction func(*UserInstrContext) (Sexp, error)

func newUserInstrCtx(name string, env *Environment, nargs int) *UserInstrContext {
	return &UserInstrContext{name: name, Environment: env, nargs: nargs}
}

func (ctx *UserInstrContext) PopExpr() Sexp {
	if ctx.nargs <= 0 {
		return NewErrorWith(fmt.Errorf("userinstr:%s no argument left on stack", ctx.name))
	}
	ctx.nargs--
	expr, err := ctx.Environment.datastack.PopExpr()
	if err != nil {
		return NewErrorWith(err)
	}
	return expr
}

type userInstr struct {
	nargs     int
	name      string
	userinstr UserInstruction
}

func (i userInstr) InstrString() string {
	return i.name
}

func (i userInstr) Execute(env *Environment) error {
	expr, err := i.userinstr(newUserInstrCtx(i.name, env, i.nargs))
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}
