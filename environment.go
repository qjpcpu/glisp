package glisp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
)

type Environment struct {
	datastack   *Stack
	scopestack  *ScopeStack
	addrstack   *Stack
	stackstack  *Stack
	symtable    map[string]int
	revsymtable map[int]string
	builtins    map[int]*SexpFunction
	userInstr   map[string]UserInstruction
	macros      *FuncMap
	curfunc     *SexpFunction
	mainfunc    *SexpFunction
	pc          int
	nextsymbol  *nextSymbol
	fileReader  FileReader
}

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100
const StackStackSize = 5

func New() *Environment {
	env := new(Environment)
	env.datastack = NewStack(DataStackSize)
	env.scopestack = NewScopeStack()
	env.scopestack.PushScope()
	env.stackstack = NewStack(StackStackSize)
	env.addrstack = NewStack(CallStackSize)
	env.builtins = make(map[int]*SexpFunction)
	env.macros = NewFuncMap()
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = &nextSymbol{counter: 1}
	env.fileReader = DefaultFileReader()
	env.userInstr = make(map[string]UserInstruction)

	for key, function := range BuiltinFunctions() {
		sym := env.MakeSymbol(key)
		env.builtins[sym.number] = MakeUserFunction(key, function, WithDoc(QueryBuiltinDoc(key)))
		env.AddFunction(key, function, WithDoc(QueryBuiltinDoc(key)))
	}

	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	env.curfunc = env.mainfunc
	env.pc = 0

	env.SourceStream(bytes.NewBufferString(buitin_scripts))
	return env
}

func (env *Environment) Clone() *Environment {
	dupenv := new(Environment)

	dupenv.datastack = env.datastack.Clone()
	dupenv.stackstack = env.stackstack.Clone()
	dupenv.scopestack = env.scopestack.Clone()
	dupenv.addrstack = env.addrstack.Clone()
	dupenv.fileReader = env.fileReader

	dupenv.builtins = copyFuncMap(env.builtins)
	dupenv.macros = env.macros.Clone()
	dupenv.symtable = make(map[string]int)
	for k, v := range env.symtable {
		dupenv.symtable[k] = v
	}
	dupenv.revsymtable = make(map[int]string)
	for k, v := range env.revsymtable {
		dupenv.revsymtable[k] = v
	}
	dupenv.nextsymbol = env.nextsymbol.Clone()
	dupenv.userInstr = make(map[string]UserInstruction)
	for k, v := range env.userInstr {
		dupenv.userInstr[k] = v
	}

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0))
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	return dupenv
}

func (env *Environment) Duplicate() *Environment {
	dupenv := new(Environment)
	dupenv.datastack = NewStack(DataStackSize)
	dupenv.scopestack = env.scopestack.ForkBottom()
	dupenv.stackstack = NewStack(StackStackSize)
	dupenv.addrstack = NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.macros = env.macros.Clone()
	// Create copies of the symbol tables and the symbol generator
	// to prevent race conditions if multiple duplicated environments are used
	// concurrently. This makes each duplicated environment safe for defining
	// new symbols in its own context.
	dupenv.symtable = make(map[string]int)
	for k, v := range env.symtable {
		dupenv.symtable[k] = v
	}
	dupenv.revsymtable = make(map[int]string)
	for k, v := range env.revsymtable {
		dupenv.revsymtable[k] = v
	}
	dupenv.nextsymbol = env.nextsymbol
	dupenv.fileReader = env.fileReader
	dupenv.userInstr = env.userInstr

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
	symbol := SexpSymbol{name, int(env.nextsymbol.Get())}
	env.symtable[name] = symbol.number
	env.revsymtable[symbol.number] = name
	env.nextsymbol.Incr()
	return symbol
}

func (env *Environment) GenSymbol(optionalPrefix ...string) SexpSymbol {
	prefix := "__anon"
	if len(optionalPrefix) > 0 && optionalPrefix[0] != "" {
		prefix = optionalPrefix[0]
	}
	symname := prefix + strconv.FormatInt(env.nextsymbol.Get(), 10)
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
		PutSlice(optargs)
	} else {
		env.datastack.PushExpr(SexpNull)
	}
	return nil
}

func (env *Environment) CallFunction(function *SexpFunction, nargs int) error {
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
		return errors.New("where's the global scope?")
	}
	env.stackstack.Push(env.scopestack)

	if function.closeScope != nil {
		env.scopestack = function.closeScope.Fork()
	} else {
		env.scopestack = env.scopestack.ForkBottom()
	}

	env.addrstack.PushAddr(env.curfunc, min(env.pc+1, len(env.curfunc.fun)))
	env.scopestack.PushScope()
	env.curfunc = function
	env.pc = 0

	return nil
}

func (env *Environment) Bind(name string, expr Sexp) error {
	sym := env.MakeSymbol(name)
	return env.scopestack.BindSymbol(sym, expr)
}

func (env *Environment) BindGlobal(name string, expr Sexp) error {
	sym := env.MakeSymbol(name)
	if env.scopestack.IsEmpty() {
		return errors.New("no scope available")
	}
	env.scopestack.BindSymbol(sym, expr, BIND_GLOBAL)
	return nil
}

func (env *Environment) PushScope() error {
	env.scopestack.PushScope()
	return nil
}

func (env *Environment) PopScope() error {
	return env.scopestack.PopScope()
}

func (env *Environment) ReturnFromFunction() error {
	var err error
	env.curfunc, env.pc, err = env.addrstack.PopAddr()
	if err != nil {
		return err
	}
	scopestack, err := env.stackstack.Pop()
	if err != nil {
		return err
	}
	// The current scopestack is about to be replaced. We must clear it
	// to decrement the reference counts of any layers it was sharing.
	env.scopestack.Clear()
	env.scopestack = scopestack.(*ScopeStack)

	return nil
}

func (env *Environment) CallUserFunction(function *SexpFunction, name string, nargs int) error {
	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return fmt.Errorf("Error calling %s: %v", name, err)
	}

	env.addrstack.PushAddr(env.curfunc, min(env.pc+1, len(env.curfunc.fun)))
	env.curfunc = function
	env.pc = -1

	res, err := function.userfun(env, args)
	PutSlice(args)
	if err != nil {
		return fmt.Errorf("Error calling %s: %v", name, err)
	}
	env.datastack.PushExpr(res)

	env.curfunc, env.pc, _ = env.addrstack.PopAddr()
	return nil
}

func (env *Environment) ParseStream(in io.Reader) ([]Sexp, error) {
	lexer := NewLexerFromStream(bufio.NewReader(in))

	var err error
	var exp []Sexp

	exp, err = ParseTokens(env, lexer)
	if err != nil {
		return nil, fmt.Errorf("Error on line %d,%d: %v\n", lexer.Linenum(), lexer.LineOffset(), err)
	}

	return exp, nil
}

// ParseFile, used in the generator at read time to dynamiclly add more defs from other files
func (env *Environment) ParseFile(file string) ([]Sexp, error) {
	in, err := env.fileReader.Open(file)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	var exp []Sexp

	exp, err = env.ParseStream(in)

	return exp, err
}

func (env *Environment) SetFileReader(fr FileReader) {
	env.fileReader = fr
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

func (env *Environment) AddFunction(name string, function UserFunction, opts ...FuntionOption) {
	env.BindGlobal(name, MakeUserFunction(name, function, opts...))
}

func (env *Environment) AddNamedFunction(name string, function NamedUserFunction, opts ...FuntionOption) {
	env.BindGlobal(name, MakeUserFunction(name, function(name), opts...))
}

func (env *Environment) AddInstruction(name string, instr UserInstruction) {
	env.userInstr[name] = instr
}

func (env *Environment) AddMacro(name string, function UserFunction, opts ...FuntionOption) {
	sym := env.MakeSymbol(name)
	env.macros.Add(sym, MakeUserFunction(name, function, opts...))
}

func (env *Environment) AddFuzzyMacro(name string, function UserFunction, opts ...FuntionOption) {
	sym := env.MakeSymbol(name)
	opts = append(opts, withNameRegexp(name))
	env.macros.Add(sym, MakeUserFunction(name, function, opts...))
}

func (env *Environment) ImportEval() error {
	env.AddNamedFunction("source-file", GetSourceFileFunction, WithDoc(QueryBuiltinDoc("source-file")))
	env.AddNamedFunction("eval", GetEvalFunction, WithDoc(QueryBuiltinDoc("eval")))
	return nil
}

func (env *Environment) DumpFunctionByName(w io.Writer, name string) error {
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
	DumpFunction(w, fun)
	return nil
}

func DumpFunction(w io.Writer, fun Function) {
	for i, instr := range fun {
		fmt.Fprintf(w, "[%02d]\t%s\n", i, instr.InstrString())
	}
}

func (env *Environment) DumpEnvironment(w io.Writer) {
	fmt.Fprintln(w, "Instructions:")
	if !env.curfunc.user {
		DumpFunction(w, env.curfunc.fun)
	}
	fmt.Fprintln(w, "Stack:")
	env.datastack.PrintStack(w)
	fmt.Fprintf(w, "PC: %d\n", env.pc)
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
	env.scopestack.Clear()
	env.addrstack.tos = -1
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

func (env *Environment) FindMacro(name string) (*SexpFunction, bool) {
	sym := env.MakeSymbol(name)
	return env.macros.Find(sym)
}

func (env *Environment) ApplyByName(fun string, args []Sexp) (Sexp, error) {
	f, ok := env.FindObject(fun)
	if !ok {
		return SexpNull, fmt.Errorf("function %s not found", fun)
	}
	fn, ok := f.(*SexpFunction)
	if !ok {
		return SexpNull, fmt.Errorf("%s is not a function", InspectType(f))
	}
	return env.Apply(fn, args)
}

func (env *Environment) Apply(fun *SexpFunction, args []Sexp) (Sexp, error) {
	if fun.user {
		return fun.userfun(env, args)
	}

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
		var buf bytes.Buffer
		env.DumpEnvironment(&buf)
		return SexpNull, errors.New(buf.String())
	}

	return env.datastack.PopExpr()
}

func (env *Environment) GlobalFunctions() []string {
	ret := env.scopestack.GlobalFuntions()
	ret = append(ret, env.macros.Names()...)
	ret = append(ret,
		"and", "or", "cond",
		"quote",
		"def", "fn", "defn", "set!",
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

func (env *Environment) MakeScriptFunction(script string) (*SexpFunction, error) {
	templ := `#(begin %s)`
	fnstr := fmt.Sprintf(templ, script)
	expr, err := env.EvalString(fnstr)
	if err != nil {
		return nil, err
	}
	return expr.(*SexpFunction), nil
}

func (env *Environment) OverrideFunction(name string, f OverrideFunction, opts ...FuntionOption) error {
	obj, ok := env.FindObject(name)
	if !ok {
		return fmt.Errorf("function `%s` not found", name)
	}
	if !IsFunction(obj) {
		return fmt.Errorf("`%s` is not a function", name)
	}
	fn := obj.(*SexpFunction).Clone()
	fn.name = name
	nopts := []FuntionOption{WithDoc(fn.Doc())}
	env.AddFunction(name, f(fn), append(nopts, opts...)...)
	return nil
}

type nextSymbol struct{ counter int64 }

func (g *nextSymbol) Incr() int64 {
	return atomic.AddInt64(&g.counter, 1)
}

func (g *nextSymbol) Get() int64 {
	return atomic.LoadInt64(&g.counter)
}

func (g *nextSymbol) Clone() *nextSymbol {
	return &nextSymbol{counter: g.counter}
}
