package glisp

import (
	"errors"
	"fmt"
)

type Instruction interface {
	InstrString() string
	Execute(env *Environment) error
}

type JumpInstr struct {
	location int
}

var OutOfBounds error = errors.New("jump out of bounds")

func (j JumpInstr) InstrString() string {
	return fmt.Sprintf("jump %d", j.location)
}

func (j JumpInstr) Execute(env *Environment) error {
	newpc := env.pc + j.location
	if newpc < 0 || newpc > env.CurrentFunctionSize() {
		return OutOfBounds
	}
	env.pc = newpc
	return nil
}

type GotoInstr struct {
	location int
}

func (g GotoInstr) InstrString() string {
	return fmt.Sprintf("goto %d", g.location)
}

func (g GotoInstr) Execute(env *Environment) error {
	if g.location < 0 || g.location > env.CurrentFunctionSize() {
		return OutOfBounds
	}
	env.pc = g.location
	return nil
}

type BranchInstr struct {
	direction bool
	location  int
}

func (b BranchInstr) InstrString() string {
	var format string
	if b.direction {
		format = "br %d"
	} else {
		format = "brn %d"
	}
	return fmt.Sprintf(format, b.location)
}

func (b BranchInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	if b.direction == IsTruthy(expr) {
		return JumpInstr{b.location}.Execute(env)
	}
	env.pc++
	return nil
}

type PushInstrClosure struct {
	expr *SexpFunction
}

func (p PushInstrClosure) InstrString() string {
	return "pushC " + p.expr.SexpString()
}

func (p PushInstrClosure) Execute(env *Environment) error {
	fn := p.expr.Clone()
	fn.closeScope = env.scopestack.Clone() // for a non script fuction I have no idea what it accesses, so we clone the whole thing
	env.datastack.PushExpr(fn)
	env.pc++
	return nil
}

type PushInstr struct {
	expr Sexp
}

func (p PushInstr) InstrString() string {
	return "push " + p.expr.SexpString()
}

func (p PushInstr) Execute(env *Environment) error {
	env.datastack.PushExpr(p.expr)
	env.pc++
	return nil
}

type PopInstr int

func (p PopInstr) InstrString() string {
	return "pop"
}

func (p PopInstr) Execute(env *Environment) error {
	_, err := env.datastack.PopExpr()
	env.pc++
	return err
}

type DupInstr int

func (d DupInstr) InstrString() string {
	return "dup"
}

func (d DupInstr) Execute(env *Environment) error {
	expr, err := env.datastack.GetExpr(0)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type GetInstr struct {
	sym SexpSymbol
}

func (g GetInstr) InstrString() string {
	return fmt.Sprintf("get %s", g.sym.name)
}

func (g GetInstr) Execute(env *Environment) error {
	expr, err := env.scopestack.LookupSymbol(g.sym)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type PutInstr struct {
	sym SexpSymbol
}

func (p PutInstr) InstrString() string {
	return fmt.Sprintf("put %s", p.sym.name)
}

func (p PutInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	env.pc++
	return env.scopestack.BindSymbol(p.sym, expr)
}

type BindDynFunInstr struct {
	sym SexpSymbol
}

func (p BindDynFunInstr) InstrString() string {
	return "bind dynamic function"
}

func (p BindDynFunInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	if !IsFunction(expr) {
		return fmt.Errorf("%s is not a function", expr.SexpString())
	}
	name, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	if !IsSymbol(name) {
		return fmt.Errorf("bad function name %s", name.SexpString())
	}

	env.pc++
	return env.scopestack.BindSymbol(name.(SexpSymbol), expr)
}

type CallInstr struct {
	sym   SexpSymbol
	nargs int
}

func (c CallInstr) InstrString() string {
	return fmt.Sprintf("call %s %d", c.sym.name, c.nargs)
}

func (c CallInstr) Execute(env *Environment) error {
	funcobj, err := env.scopestack.LookupSymbol(c.sym)
	if err != nil {
		f, ok := env.builtins[c.sym.number]
		if ok {
			return env.CallUserFunction(f, c.sym.name, c.nargs)
		}
		return err
	}
	switch f := funcobj.(type) {
	case *SexpFunction:
		if !f.user {
			return env.CallFunction(f, c.nargs)
		}
		return env.CallUserFunction(f, c.sym.name, c.nargs)
	}
	return fmt.Errorf("%s is not a function", c.sym.name)
}

type PrepareCallInstr struct {
	sym   SexpSymbol
	nargs int
}

func (c PrepareCallInstr) InstrString() string {
	return fmt.Sprintf("preparecall %s %d", c.sym.name, c.nargs)
}

func (c PrepareCallInstr) Execute(env *Environment) error {
	if err := c.execute(env); err != nil {
		return err
	}
	env.pc++
	return nil
}

func (c PrepareCallInstr) execute(env *Environment) error {
	funcobj, err := env.scopestack.LookupSymbol(c.sym)
	if err != nil {
		_, ok := env.builtins[c.sym.number]
		if ok {
			return nil
		}
		return err
	}
	switch f := funcobj.(type) {
	case *SexpFunction:
		if !f.user && f.varargs {
			return env.wrangleOptargs(f.nargs, c.nargs)
		}
		return nil
	}
	return nil
}

type DispatchInstr struct {
	nargs int
}

func (d DispatchInstr) InstrString() string {
	return fmt.Sprintf("dispatch %d", d.nargs)
}

func (d DispatchInstr) Execute(env *Environment) error {
	funcobj, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	switch f := funcobj.(type) {
	case *SexpFunction:
		if !f.user {
			return env.CallFunction(f, d.nargs)
		}
		return env.CallUserFunction(f, f.name, d.nargs)
	}
	return fmt.Errorf("%s not a function", funcobj.SexpString())
}

type ReturnInstr struct {
	err        error
	dynamicErr bool
}

func (r ReturnInstr) Execute(env *Environment) error {
	if r.err != nil {
		return r.err
	}
	if r.dynamicErr {
		elem, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if IsString(elem) {
			return errors.New(string(elem.(SexpStr)))
		}
		return errors.New(elem.SexpString())
	}
	return env.ReturnFromFunction()
}

func (r ReturnInstr) InstrString() string {
	if r.err == nil {
		return "ret"
	}
	return "ret \"" + r.err.Error() + "\""
}

type AddScopeInstr int

func (a AddScopeInstr) InstrString() string {
	return "add scope"
}

func (a AddScopeInstr) Execute(env *Environment) error {
	env.scopestack.PushScope()
	env.pc++
	return nil
}

type RemoveScopeInstr int

func (a RemoveScopeInstr) InstrString() string {
	return "rem scope"
}

func (a RemoveScopeInstr) Execute(env *Environment) error {
	env.pc++
	return env.scopestack.PopScope()
}

type ExplodeInstr int

func (e ExplodeInstr) InstrString() string {
	return "explode"
}

func (e ExplodeInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	arr, err := ListToArray(expr)
	if err != nil {
		return err
	}

	for _, val := range arr {
		env.datastack.PushExpr(val)
	}
	env.pc++
	return nil
}

type SquashInstr int

func (s SquashInstr) InstrString() string {
	return "squash"
}

func (s SquashInstr) Execute(env *Environment) error {
	var list Sexp = SexpNull

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		list = Cons(expr, list)
	}
	env.datastack.PushExpr(list)
	env.pc++
	return nil
}

type VectorizeInstr int

func (s VectorizeInstr) InstrString() string {
	return "vectorize"
}

func (s VectorizeInstr) Execute(env *Environment) error {
	vec := make([]Sexp, 0)

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		vec = append([]Sexp{expr}, vec...)
	}
	env.datastack.PushExpr(SexpArray(vec))
	env.pc++
	return nil
}

type HashizeInstr struct {
	HashLen int
}

func (s HashizeInstr) InstrString() string {
	return "hashize"
}

func (s HashizeInstr) Execute(env *Environment) error {
	a := make([]Sexp, 0)

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		a = append(a, expr)
	}
	hash, err := MakeHash(a)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(hash)
	env.pc++
	return nil
}

type BindlistInstr struct{}

func (b BindlistInstr) InstrString() string {
	return `bindlist`
}

func (b BindlistInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	switch arr := expr.(type) {
	case SexpArray:
		if len(arr)%2 != 0 {
			return errors.New("bind list length must be even")
		}

		for i := 0; i*2+1 < len(arr); i++ {
			if !IsSymbol(arr[i*2]) {
				return errors.New("odd argument of bind list must be symbol but got " + Inspect(arr[i*2]))
			}
			env.scopestack.BindSymbol(arr[i*2].(SexpSymbol), arr[i*2+1])
		}
	case *SexpHash:
		var err error
		arr.Foreach(func(k Sexp, v Sexp) bool {
			if IsSymbol(k) {
				env.scopestack.BindSymbol(k.(SexpSymbol), v)
				return true
			}
			err = errors.New("hash key must be symbol but got " + Inspect(k))
			return false
		})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(`bad let binding type %v`, Inspect(expr))
	}

	env.pc++
	return nil
}

type RefSymInstr struct {
}

func (p RefSymInstr) InstrString() string {
	return "ref symbol"
}

func (p RefSymInstr) Execute(env *Environment) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	if !IsSymbol(expr) {
		return fmt.Errorf("%s is not a symbol", expr.SexpString())
	}
	if t, ok := env.FindObject(expr.(SexpSymbol).Name()); ok {
		env.datastack.PushExpr(t)
	} else {
		env.datastack.PushExpr(SexpNull)
	}

	env.pc++
	return nil
}
