package glisp

import (
	"fmt"
)

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
