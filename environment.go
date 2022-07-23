package glisp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type PreHook func(*Environment, string, []Sexp)
type PostHook func(*Environment, string, Sexp)

type Environment struct {
	datastack        *Stack
	scopestack       *Stack
	addrstack        *Stack
	stackstack       *Stack
	symtable         map[string]int
	revsymtable      map[int]string
	builtins         map[int]*SexpFunction
	macros           map[int]*SexpFunction
	curfunc          *SexpFunction
	mainfunc         *SexpFunction
	pc               int
	nextsymbol       int
	before           []PreHook
	after            []PostHook
	extraGlobalCount int
}

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100
const StackStackSize = 5

func New() *Environment {
	env := new(Environment)
	env.datastack = NewStack(DataStackSize)
	env.scopestack = NewStack(ScopeStackSize)
	env.scopestack.PushScope()
	env.stackstack = NewStack(StackStackSize)
	env.addrstack = NewStack(CallStackSize)
	env.builtins = make(map[int]*SexpFunction)
	env.macros = make(map[int]*SexpFunction)
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = 1
	env.before = []PreHook{}
	env.after = []PostHook{}

	for key, function := range BuiltinFunctions() {
		sym := env.MakeSymbol(key)
		env.builtins[sym.number] = MakeUserFunction(key, function)
		env.AddFunction(key, function)
	}

	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	env.curfunc = env.mainfunc
	env.pc = 0

	return env
}

func (env *Environment) Clone() *Environment {
	dupenv := new(Environment)

	dupenv.datastack = env.datastack.Clone()
	dupenv.stackstack = env.stackstack.Clone()
	dupenv.scopestack = env.scopestack.Clone()
	dupenv.addrstack = env.addrstack.Clone()

	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.PushMulti(env.globalScopes()...)
	dupenv.extraGlobalCount = env.extraGlobalCount

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	return dupenv
}

func (env *Environment) Duplicate() *Environment {
	dupenv := new(Environment)
	dupenv.datastack = NewStack(DataStackSize)
	dupenv.scopestack = NewStack(ScopeStackSize)
	dupenv.stackstack = NewStack(StackStackSize)
	dupenv.addrstack = NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.PushMulti(env.globalScopes()...)
	dupenv.extraGlobalCount = env.extraGlobalCount

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	return dupenv
}

func (env *Environment) MakeSymbol(name string) SexpSymbol {
	symnum, ok := env.symtable[name]
	if ok {
		return SexpSymbol{name, symnum}
	}
	symbol := SexpSymbol{name, env.nextsymbol}
	env.symtable[name] = symbol.number
	env.revsymtable[symbol.number] = name
	env.nextsymbol++
	return symbol
}

func (env *Environment) GenSymbol(prefix string) SexpSymbol {
	symname := prefix + strconv.Itoa(env.nextsymbol)
	return env.MakeSymbol(symname)
}

func (env *Environment) CurrentFunctionSize() int {
	if env.curfunc.user {
		return 0
	}
	return len(env.curfunc.fun)
}

func (env *Environment) wrangleOptargs(fnargs, nargs int) error {
	if nargs < fnargs {
		return errors.New(
			fmt.Sprintf("Expected >%d arguments, got %d",
				fnargs, nargs))
	}
	if nargs > fnargs {
		optargs, err := env.datastack.PopExpressions(nargs - fnargs)
		if err != nil {
			return err
		}
		env.datastack.PushExpr(MakeList(optargs))
	} else {
		env.datastack.PushExpr(SexpNull)
	}
	return nil
}

func (env *Environment) CallFunction(function *SexpFunction, nargs int) error {
	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return err
		}
		prehook(env, function.name, expressions)
	}

	if function.varargs {
		err := env.wrangleOptargs(function.nargs, nargs)
		if err != nil {
			return err
		}
	} else if nargs != function.nargs {
		return errors.New(
			fmt.Sprintf("%s expected %d arguments, got %d",
				function.name, function.nargs, nargs))
	}

	if env.scopestack.IsEmpty() {
		panic("where's the global scope?")
	}
	globalScopes := env.globalScopes()
	env.stackstack.Push(env.scopestack)
	env.scopestack = NewStack(ScopeStackSize)
	env.scopestack.PushMulti(globalScopes...)

	if function.closeScope != nil {
		function.closeScope.PushAllTo(env.scopestack)
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.scopestack.PushScope()
	env.curfunc = function
	env.pc = 0

	return nil
}

func (env *Environment) Bind(name string, expr Sexp) error {
	sym := env.MakeSymbol(name)
	return env.scopestack.BindSymbol(sym, expr)
}

func (env *Environment) PushGlobalScope() error {
	if env.scopestack.Top() != env.extraGlobalCount {
		return errors.New("not in global scope")
	}
	env.scopestack.PushScope()
	env.extraGlobalCount++
	return nil
}

func (env *Environment) PopGlobalScope() error {
	if env.scopestack.Top() != env.extraGlobalCount {
		return errors.New("not in global scope")
	}
	if env.extraGlobalCount <= 0 {
		return errors.New("no extra global scope")
	}
	if err := env.scopestack.PopScope(); err != nil {
		return err
	}
	env.extraGlobalCount--
	return nil
}

func (env *Environment) globalScopes() []StackElem {
	return env.scopestack.elements[0 : env.extraGlobalCount+1]
}

func (env *Environment) ReturnFromFunction() error {
	for _, posthook := range env.after {
		retval, err := env.datastack.GetExpr(0)
		if err != nil {
			return err
		}
		posthook(env, env.curfunc.name, retval)
	}

	var err error
	env.curfunc, env.pc, err = env.addrstack.PopAddr()
	if err != nil {
		return err
	}
	scopestack, err := env.stackstack.Pop()
	if err != nil {
		return err
	}
	env.scopestack = scopestack.(*Stack)

	return nil
}

func (env *Environment) CallUserFunction(
	function *SexpFunction, name string, nargs int) error {

	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return err
		}
		prehook(env, function.name, expressions)
	}

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error calling %s: %v", name, err))
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.curfunc = function
	env.pc = -1

	res, err := function.userfun(env, args)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error calling %s: %v", name, err))
	}
	env.datastack.PushExpr(res)

	for _, posthook := range env.after {
		posthook(env, name, res)
	}

	env.curfunc, env.pc, _ = env.addrstack.PopAddr()
	return nil
}

func (env *Environment) ParseStream(in io.Reader) ([]Sexp, error) {
	lexer := NewLexerFromStream(bufio.NewReader(in))

	var err error
	var exp []Sexp

	exp, err = ParseTokens(env, lexer)
	if err != nil {
		return nil, fmt.Errorf("Error on line %d: %v\n", lexer.Linenum(), err)
	}

	return exp, nil
}

// ParseFile, used in the generator at read time to dynamiclly add more defs from other files
func (env *Environment) ParseFile(file string) ([]Sexp, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var exp []Sexp

	exp, err = env.ParseStream(in)

	in.Close()

	return exp, err
}

func (env *Environment) SourceExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)
	if !env.ReachedEnd() {
		gen.AddInstruction(PopInstr(0))
	}
	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	curfunc := env.curfunc
	curpc := env.pc

	env.curfunc = MakeFunction("__source", 0, false, gen.instructions)
	env.pc = 0

	env.datastack.PushExpr(SexpNull)

	if _, err = env.Run(); err != nil {
		return err
	}

	env.datastack.PopExpr()

	env.pc = curpc
	env.curfunc = curfunc

	return nil
}

// SourceStream, load this in via a __source dynamic function, after it runs it no longer exists
func (env *Environment) SourceStream(stream io.Reader) error {
	expressions, err := env.ParseStream(stream)

	if err != nil {
		return err
	}

	return env.SourceExpressions(expressions)
}

func (env *Environment) SourceFile(file *os.File) error {
	return env.SourceStream(bufio.NewReader(file))
}

func (env *Environment) LoadExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)
	if !env.ReachedEnd() {
		gen.AddInstruction(PopInstr(0))
	}
	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	env.mainfunc.fun = append(env.mainfunc.fun, gen.instructions...)
	env.curfunc = env.mainfunc

	return nil
}

// LoadStream, load this in via running a __main function and setting main on the environment
func (env *Environment) LoadStream(stream io.Reader) error {
	expressions, err := env.ParseStream(stream)

	if err != nil {
		return err
	}

	return env.LoadExpressions(expressions)
}

func (env *Environment) EvalString(str string) (Sexp, error) {
	err := env.LoadString(str)
	if err != nil {
		return SexpNull, err
	}

	return env.Run()
}

func (env *Environment) LoadFile(file *os.File) error {
	return env.LoadStream(bufio.NewReader(file))
}

func (env *Environment) LoadString(str string) error {
	return env.LoadStream(bytes.NewBuffer([]byte(str)))
}

func (env *Environment) AddFunction(name string, function UserFunction) {
	env.Bind(name, MakeUserFunction(name, function))
}

func (env *Environment) AddFunctionByConstructor(name string, function UserFunctionConstructor) {
	env.Bind(name, MakeUserFunction(name, function(name)))
}

func (env *Environment) AddMacro(name string, function UserFunction) {
	sym := env.MakeSymbol(name)
	env.macros[sym.number] = MakeUserFunction(name, function)
}

func (env *Environment) ImportEval() {
	env.AddFunction("source-file", SourceFileFunction)
	env.AddFunction("eval", EvalFunction)
}

func (env *Environment) DumpFunctionByName(name string) error {
	obj, found := env.FindObject(name)
	if !found {
		return fmt.Errorf("%q not found", name)
	}

	var fun Function
	switch t := obj.(type) {
	case *SexpFunction:
		if !t.user {
			fun = t.fun
		} else {
			return errors.New("not a glisp function")
		}
	default:
		return errors.New("not a function")
	}
	DumpFunction(fun)
	return nil
}

func DumpFunction(fun Function) {
	for _, instr := range fun {
		fmt.Println("\t" + instr.InstrString())
	}
}

func (env *Environment) DumpEnvironment() {
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		DumpFunction(env.curfunc.fun)
	}
	fmt.Println("Stack:")
	env.datastack.PrintStack()
	fmt.Printf("PC: %d\n", env.pc)
}

func (env *Environment) ReachedEnd() bool {
	return env.pc == env.CurrentFunctionSize()
}

func (env *Environment) GetStackTrace(err error) string {
	str := fmt.Sprintf("error in %s:%d: %v\n",
		env.curfunc.name, env.pc, err)
	for !env.addrstack.IsEmpty() {
		fun, pos, _ := env.addrstack.PopAddr()
		str += fmt.Sprintf("in %s:%d\n", fun.name, pos)
	}
	return str
}

func (env *Environment) Clear() {
	env.datastack.tos = -1
	env.scopestack.tos = 0
	env.addrstack.tos = -1
	env.extraGlobalCount = 0
	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	env.curfunc = env.mainfunc
	env.pc = 0
}

func (env *Environment) FindObject(name string) (Sexp, bool) {
	sym := env.MakeSymbol(name)
	obj, err := env.scopestack.LookupSymbol(sym)
	if err != nil {
		return SexpNull, false
	}
	return obj, true
}

func (env *Environment) ApplyByName(fun string, args []Sexp) (Sexp, error) {
	f, ok := env.FindObject(fun)
	if !ok {
		return SexpNull, fmt.Errorf("function %s not found", fun)
	}
	fn, ok := f.(*SexpFunction)
	if !ok {
		return SexpNull, fmt.Errorf("%s(%T) is not a function", fun, f)
	}
	return env.Apply(fn, args)
}

func (env *Environment) Apply(fun *SexpFunction, args []Sexp) (Sexp, error) {
	if fun.user {
		return fun.userfun(env, args)
	}

	env.pc = -2
	for _, expr := range args {
		env.datastack.PushExpr(expr)
	}

	err := env.CallFunction(fun, len(args))
	if err != nil {
		return SexpNull, err
	}

	return env.Run()
}

func (env *Environment) Run() (Sexp, error) {
	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc.fun[env.pc]
		err := instr.Execute(env)
		if err != nil {
			return SexpNull, err
		}
	}

	if env.datastack.IsEmpty() {
		env.DumpEnvironment()
		os.Exit(-1)
	}

	return env.datastack.PopExpr()
}

func (env *Environment) AddPreHook(fun PreHook) {
	env.before = append(env.before, fun)
}

func (env *Environment) AddPostHook(fun PostHook) {
	env.after = append(env.after, fun)
}

func (env *Environment) GlobalFunctions() []string {
	var ret []string
	for _, scope := range env.globalScopes() {
		for _, v := range scope.(Scope) {
			if fn, ok := v.(*SexpFunction); ok {
				ret = append(ret, fn.name)
			}
		}
	}
	for _, fn := range env.macros {
		ret = append(ret, fn.name)
	}
	ret = append(ret,
		"and", "or", "cond",
		"quote",
		"def", "fn", "defn",
		"begin",
		"let", "let*",
		"assert",
		"defmac",
		"macexpand",
		"syntax-quote",
		"include",
	)
	return ret
}
