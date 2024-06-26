package glisp

import (
	"errors"
	"fmt"
	"regexp"
)

type Generator struct {
	env          *Environment
	funcname     string
	tail         bool
	scopes       int
	instructions []Instruction
}

type Loop struct {
	stmtname       SexpSymbol
	loopStart      int
	loopLen        int
	breakOffset    int // i.e. relative to loopStart
	continueOffset int // i.e. relative to loopStart
}

func (loop *Loop) IsStackElem() {}

func NewGenerator(env *Environment) *Generator {
	gen := new(Generator)
	gen.env = env
	gen.instructions = make([]Instruction, 0)
	// tail marks whether or not we are in the tail position
	gen.tail = false
	// scopes is the number of extra (non-function) scopes we've created
	gen.scopes = 0
	return gen
}

func (gen *Generator) AddInstructions(instr []Instruction) {
	gen.instructions = append(gen.instructions, instr...)
}

func (gen *Generator) AddInstruction(instr Instruction) {
	gen.instructions = append(gen.instructions, instr)
}

func (gen *Generator) GenerateBegin(expressions []Sexp) error {
	size := len(expressions)
	oldtail := gen.tail
	gen.tail = false
	if size == 0 {
		return errors.New("No expressions found")
	}
	for _, expr := range expressions[:size-1] {
		err := gen.Generate(expr)
		if err != nil {
			return err
		}
		// insert pops after all but the last instruction
		// that way the stack remains clean
		gen.AddInstruction(PopInstr(0))
	}
	gen.tail = oldtail
	return gen.Generate(expressions[size-1])
}

func buildSexpFun(env *Environment, name string, funcargs SexpArray,
	funcbody []Sexp) (*SexpFunction, error) {
	gen := NewGenerator(env)
	gen.tail = true

	if len(name) == 0 {
		gen.funcname = env.GenSymbol("__anon").name
	} else {
		gen.funcname = name
	}

	argsyms := make([]SexpSymbol, len(funcargs))

	for i, expr := range funcargs {
		switch t := expr.(type) {
		case SexpSymbol:
			argsyms[i] = t
		default:
			return MissingFunction,
				errors.New("function argument must be symbol")
		}
	}

	varargs := false
	nargs := len(funcargs)

	if len(argsyms) >= 2 && argsyms[len(argsyms)-2].name == "&" {
		argsyms[len(argsyms)-2] = argsyms[len(argsyms)-1]
		argsyms = argsyms[0 : len(argsyms)-1]
		varargs = true
		nargs = len(argsyms) - 1
	}

	for i := len(argsyms) - 1; i >= 0; i-- {
		gen.AddInstruction(PutInstr{sym: argsyms[i]})
	}

	var doc string
	if len(funcbody) > 1 && IsString(funcbody[0]) {
		doc = string(funcbody[0].(SexpStr))
		funcbody = funcbody[1:]
	}

	err := gen.GenerateBegin(funcbody)
	if err != nil {
		return MissingFunction, err
	}
	gen.AddInstruction(ReturnInstr{})

	newfunc := Function(gen.instructions)
	return MakeFunction(gen.funcname, nargs, varargs, newfunc, WithDoc(doc)), nil
}

func (gen *Generator) GenerateFn(args []Sexp) error {
	if len(args) < 2 {
		return errors.New("malformed function definition")
	}

	var funcargs SexpArray

	switch expr := args[0].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	funcbody := args[1:]
	sfun, err := buildSexpFun(gen.env, "", funcargs, funcbody)
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstrClosure{sfun})

	return nil
}

func (gen *Generator) GenerateDef(args []Sexp, isSet bool) error {
	if len(args) != 2 {
		return errors.New("Wrong number of arguments to def")
	}

	var sym SexpSymbol
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	default:
		return errors.New("Definition name must by symbol")
	}

	gen.tail = false
	err := gen.Generate(args[1])
	if err != nil {
		return err
	}
	gen.AddInstruction(PutInstr{sym: sym, isSet: isSet})
	gen.AddInstruction(PushInstr{SexpNull})
	return nil
}

func (gen *Generator) GenerateDefn(args []Sexp) error {
	if len(args) < 3 {
		return errors.New("Wrong number of arguments to defn")
	}

	var funcargs SexpArray
	switch expr := args[1].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	var sym SexpSymbol
	var dynName bool
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	case *SexpPair:
		if IsList(expr) {
			if err := gen.GenerateCall(expr); err != nil {
				return err
			}
			dynName = true
		} else {
			return errors.New("Definition name must by symbol")
		}
	default:
		return errors.New("Definition name must by symbol")
	}

	sfun, err := buildSexpFun(gen.env, sym.name, funcargs, args[2:])
	if err != nil {
		return err
	}

	if !dynName {
		gen.AddInstruction(PushInstr{sfun})
		gen.AddInstruction(PutInstr{sym: sym})
		gen.AddInstruction(PushInstr{SexpNull})
	} else {
		gen.AddInstruction(PushInstr{sfun})
		gen.AddInstruction(BindDynFunInstr{})
		gen.AddInstruction(PushInstr{SexpNull})
	}

	return nil
}

func (gen *Generator) GenerateDefmac(args []Sexp) error {
	if len(args) < 3 {
		return errors.New("Wrong number of arguments to defmac")
	}

	var funcargs SexpArray
	switch expr := args[1].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	var sym SexpSymbol
	var regName *regexp.Regexp
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	case SexpStr:
		if r, err := regexp.Compile(string(expr)); err != nil {
			return err
		} else {
			regName = r
		}
		sym = gen.env.GenSymbol("__anon")
	default:
		return errors.New("Definition name must by symbol")
	}

	sfun, err := buildSexpFun(gen.env, sym.name, funcargs, args[2:])
	if err != nil {
		return err
	}
	if regName != nil {
		sfun.nameRegexp = regName
	}

	gen.env.macros.Add(sym, sfun)
	gen.AddInstruction(PushInstr{SexpNull})

	return nil
}

func (gen *Generator) GenerateMacexpand(args []Sexp) error {
	if len(args) != 1 {
		return WrongGeneratorNumberArguments("macexpand", len(args), 1)
	}

	var list *SexpPair
	var islist bool
	var ismacrocall bool

	switch t := args[0].(type) {
	case *SexpPair:
		if IsList(t.tail) {
			list = t
			islist = true
		}
	default:
		islist = false
	}

	var macro *SexpFunction
	if islist {
		switch t := list.head.(type) {
		case SexpSymbol:
			macro, ismacrocall = gen.env.macros.Find(t)
		default:
			ismacrocall = false
		}
	}

	if !ismacrocall {
		gen.AddInstruction(PushInstr{args[0]})
		return nil
	}

	macargs, err := ListToArray(list.tail)
	if err != nil {
		return err
	}

	env := gen.env.Duplicate()
	expr, err := env.Apply(macro, prependCallName(macro, list.head.(SexpSymbol), macargs))
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstr{expr})
	return nil
}

func (gen *Generator) GenerateShortCircuit(or bool, args []Sexp) error {
	size := len(args)

	subgen := NewGenerator(gen.env)
	subgen.scopes = gen.scopes
	subgen.tail = gen.tail
	subgen.funcname = gen.funcname
	subgen.Generate(args[size-1])
	instructions := subgen.instructions

	for i := size - 2; i >= 0; i-- {
		subgen = NewGenerator(gen.env)
		subgen.Generate(args[i])
		subgen.AddInstruction(DupInstr(0))
		subgen.AddInstruction(BranchInstr{or, len(instructions) + 2})
		subgen.AddInstruction(PopInstr(0))
		instructions = append(subgen.instructions, instructions...)
	}
	gen.AddInstructions(instructions)

	return nil
}

func (gen *Generator) GenerateCond(args []Sexp) error {
	if len(args)%2 == 0 {
		return errors.New("missing default case")
	}

	subgen := NewGenerator(gen.env)
	subgen.tail = gen.tail
	subgen.scopes = gen.scopes
	subgen.funcname = gen.funcname
	err := subgen.Generate(args[len(args)-1])
	if err != nil {
		return err
	}
	instructions := subgen.instructions

	for i := len(args)/2 - 1; i >= 0; i-- {
		subgen.Reset()
		err := subgen.Generate(args[2*i])
		if err != nil {
			return err
		}
		pred_code := subgen.instructions

		subgen.Reset()
		subgen.tail = gen.tail
		subgen.scopes = gen.scopes
		subgen.funcname = gen.funcname
		err = subgen.Generate(args[2*i+1])
		if err != nil {
			return err
		}
		body_code := subgen.instructions

		subgen.Reset()
		subgen.AddInstructions(pred_code)
		subgen.AddInstruction(BranchInstr{false, len(body_code) + 2})
		subgen.AddInstructions(body_code)
		subgen.AddInstruction(JumpInstr{len(instructions) + 1})
		subgen.AddInstructions(instructions)
		instructions = subgen.instructions
	}

	gen.AddInstructions(instructions)
	return nil
}

func (gen *Generator) GenerateQuote(args []Sexp) error {
	for _, expr := range args {
		gen.AddInstruction(PushInstr{expr})
	}
	return nil
}

func (gen *Generator) GenerateLet(name string, args []Sexp) error {
	if len(args) < 2 {
		return errors.New("malformed let statement")
	}

	switch expr := args[0].(type) {
	case SexpArray:
		return gen.generateLetArray(name, expr, args[1:])
	case *SexpPair:
		return gen.generateLetList(name, expr, args[1:])
	default:
		return errors.New("let bindings must be in array")
	}
}

func (gen *Generator) generateLetArray(name string, bindings SexpArray, args []Sexp) error {
	lstatements := make([]SexpSymbol, 0)
	rstatements := make([]Sexp, 0)

	if len(bindings)%2 != 0 {
		return errors.New("uneven let binding list")
	}

	for i := 0; i < len(bindings)/2; i++ {
		switch t := bindings[2*i].(type) {
		case SexpSymbol:
			lstatements = append(lstatements, t)
		default:
			return errors.New("cannot bind to non-symbol")
		}
		rstatements = append(rstatements, bindings[2*i+1])
	}

	gen.AddInstruction(AddScopeInstr(0))
	gen.scopes++

	if name == "let*" {
		for i, rs := range rstatements {
			err := gen.Generate(rs)
			if err != nil {
				return err
			}
			gen.AddInstruction(PutInstr{sym: lstatements[i]})
		}
	} else if name == "let" {
		for _, rs := range rstatements {
			err := gen.Generate(rs)
			if err != nil {
				return err
			}
		}
		for i := len(lstatements) - 1; i >= 0; i-- {
			gen.AddInstruction(PutInstr{sym: lstatements[i]})
		}
	}
	err := gen.GenerateBegin(args)
	if err != nil {
		return err
	}
	gen.AddInstruction(RemoveScopeInstr(0))
	gen.scopes--

	return nil
}

func (gen *Generator) generateLetList(name string, bindings *SexpPair, args []Sexp) error {
	gen.AddInstruction(AddScopeInstr(0))
	gen.scopes++

	if err := gen.Generate(bindings); err != nil {
		return err
	}
	gen.AddInstruction(BindlistInstr{})

	err := gen.GenerateBegin(args)
	if err != nil {
		return err
	}
	gen.AddInstruction(RemoveScopeInstr(0))
	gen.scopes--

	return nil
}

func (gen *Generator) GenerateAssert(args []Sexp) error {
	if len(args) != 1 && len(args) != 2 {
		return WrongGeneratorNumberArguments("assert", len(args), 1, 2)
	}
	err := gen.Generate(args[0])
	if err != nil {
		return err
	}

	if len(args) == 1 {
		reterrmsg := fmt.Sprintf("Assertion failed: %s\n",
			args[0].SexpString())
		gen.AddInstruction(BranchInstr{true, 2})
		gen.AddInstruction(ReturnInstr{err: errors.New(reterrmsg)})
		gen.AddInstruction(PushInstr{SexpNull})
		return nil
	}

	subgen := NewGenerator(gen.env)
	subgen.scopes = gen.scopes
	subgen.tail = gen.tail
	subgen.funcname = gen.funcname
	subgen.Generate(args[1])
	instructions := subgen.instructions

	gen.AddInstruction(BranchInstr{true, 2 + len(instructions)})
	gen.AddInstructions(instructions)
	gen.AddInstruction(ReturnInstr{dynamicErr: true})
	gen.AddInstruction(PushInstr{SexpNull})
	return nil
}

func (gen *Generator) GenerateInclude(args []Sexp) error {
	if len(args) < 1 {
		return WrongGeneratorNumberArguments("include", len(args), 1)
	}

	var err error
	var exps []Sexp

	var sourceItem func(item Sexp) error

	sourceItem = func(item Sexp) error {
		switch t := item.(type) {
		case SexpArray:
			for _, v := range t {
				if err := sourceItem(v); err != nil {
					return err
				}
			}
		case *SexpPair:
			expr := item
			for expr != SexpNull {
				list := expr.(*SexpPair)
				if err := sourceItem(list.head); err != nil {
					return err
				}
				expr = list.tail
			}
		case SexpStr:
			exps, err = gen.env.ParseFile(string(t))
			if err != nil {
				return err
			}

			err = gen.GenerateBegin(exps)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("include: Expected `string`, `list`, `array` given %v", InspectType(item))
		}

		return nil
	}

	for _, v := range args {
		err = sourceItem(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gen *Generator) GenerateCallBySymbol(sym SexpSymbol, args []Sexp) error {
	switch sym.name {
	case "and":
		return gen.GenerateShortCircuit(false, args)
	case "or":
		return gen.GenerateShortCircuit(true, args)
	case "cond":
		return gen.GenerateCond(args)
	case "quote":
		return gen.GenerateQuote(args)
	case "def":
		return gen.GenerateDef(args, false)
	case "set!":
		return gen.GenerateDef(args, true)
	case "fn":
		return gen.GenerateFn(args)
	case "defn":
		return gen.GenerateDefn(args)
	case "begin":
		return gen.GenerateBegin(args)
	case "let":
		return gen.GenerateLet("let", args)
	case "let*":
		return gen.GenerateLet("let*", args)
	case "assert":
		return gen.GenerateAssert(args)
	case "defmac":
		return gen.GenerateDefmac(args)
	case "macexpand":
		return gen.GenerateMacexpand(args)
	case "syntax-quote":
		return gen.GenerateSyntaxQuote(args)
	case "include":
		return gen.GenerateInclude(args)
	case "sharp-quote":
		return gen.GenerateSharpQuote(args)
	}

	macro, found := gen.env.macros.Find(sym)
	if found {
		// calling Apply on the current environment will screw up
		// the stack, creating a duplicate environment is safer
		env := gen.env.Duplicate()
		expr, err := env.Apply(macro, prependCallName(macro, sym, args))
		if err != nil {
			return err
		}
		return gen.Generate(expr)
	}

	oldtail := gen.tail
	gen.tail = false
	err := gen.GenerateAll(args)
	if err != nil {
		return err
	}
	if oldtail && sym.name == gen.funcname {
		// to do a tail call
		// pop off all the extra scopes
		// then jump to beginning of function
		for i := 0; i < gen.scopes; i++ {
			gen.AddInstruction(RemoveScopeInstr(0))
		}
		gen.AddInstruction(PrepareCallInstr{sym, len(args)})
		gen.AddInstruction(GotoInstr{0})
	} else {
		gen.AddInstruction(CallInstr{sym, len(args)})
	}
	gen.tail = oldtail
	return nil
}

func (gen *Generator) GenerateDispatch(fun Sexp, args []Sexp) error {
	gen.GenerateAll(args)
	gen.Generate(fun)
	gen.AddInstruction(DispatchInstr{len(args)})
	return nil
}

func (gen *Generator) GenerateCall(expr *SexpPair) error {
	arr, _ := ListToArray(expr.tail)
	switch head := expr.head.(type) {
	case SexpSymbol:
		return gen.GenerateCallBySymbol(head, arr)
	}
	return gen.GenerateDispatch(expr.head, arr)
}

func (gen *Generator) GenerateArray(arr SexpArray) error {
	err := gen.GenerateAll(arr)
	if err != nil {
		return err
	}
	gen.AddInstruction(CallInstr{gen.env.MakeSymbol("array"), len(arr)})
	return nil
}

func (gen *Generator) Generate(expr Sexp) error {
	switch e := expr.(type) {
	case SexpSymbol:
		gen.AddInstruction(GetInstr{e})
		return nil
	case *SexpPair:
		if IsList(e) {
			err := gen.GenerateCall(e)
			if err != nil {
				return fmt.Errorf("Error generating %s: %v",
					expr.SexpString(), err)
			}
			return nil
		} else {
			gen.AddInstruction(PushInstr{expr})
		}
	case SexpArray:
		return gen.GenerateArray(e)
	default:
		gen.AddInstruction(PushInstr{expr})
		return nil
	}
	return nil
}

func (gen *Generator) GenerateAll(expressions []Sexp) error {
	for _, expr := range expressions {
		err := gen.Generate(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) Reset() {
	gen.instructions = make([]Instruction, 0)
	gen.tail = false
	gen.scopes = 0
}

// side-effect (or main effect) has to be pushing an expression on the top of
// the datastack that represents the expanded and substituted expression
func (gen *Generator) GenerateSyntaxQuote(args []Sexp) error {

	if len(args) != 1 {
		return errors.New("syntax-quote takes exactly one argument")
	}
	arg := args[0]

	// need to handle arrays, since they can have unquotes
	// in them too.
	switch arg.(type) {
	case SexpArray:
		gen.generateSyntaxQuoteArray(arg)
		return nil
	case *SexpPair:
		if !IsList(arg) {
			break
		}
		gen.generateSyntaxQuoteList(arg)
		return nil
	case *SexpHash:
		gen.generateSyntaxQuoteHash(arg)
		return nil
	}
	gen.AddInstruction(PushInstr{arg})
	return nil
}

func (gen *Generator) generateSyntaxQuoteList(arg Sexp) error {

	switch a := arg.(type) {
	case *SexpPair:
		//good, required here
	default:
		return fmt.Errorf("arg to generateSyntaxQuoteList() must be list; got %v", InspectType(a))
	}

	// things that need unquoting end up as
	// (unquote mysym)
	// i.e. a pair
	// list of length 2 exactly, with first atom
	// being "unquote" and second being the symbol
	// to substitute.
	quotebody, _ := ListToArray(arg)

	if len(quotebody) == 2 {
		var issymbol bool
		var sym SexpSymbol
		switch t := quotebody[0].(type) {
		case SexpSymbol:
			sym = t
			issymbol = true
		default:
			issymbol = false
		}
		if issymbol {
			if sym.name == "unquote" {
				gen.Generate(quotebody[1])
				return nil
			} else if sym.name == "unquote-splicing" {
				gen.Generate(quotebody[1])
				gen.AddInstruction(ExplodeInstr(0))
				return nil
			}
		}
	}

	gen.AddInstruction(PushInstr{SexpMarker})

	for _, expr := range quotebody {
		gen.GenerateSyntaxQuote([]Sexp{expr})
	}

	gen.AddInstruction(SquashInstr(0))

	return nil
}

func (gen *Generator) generateSyntaxQuoteArray(arg Sexp) error {

	var arr SexpArray
	switch a := arg.(type) {
	case SexpArray:
		//good, required here
		arr = a
	default:
		return fmt.Errorf("arg to generateSyntaxQuoteArray() must be an array; got %v", InspectType(a))
	}

	gen.AddInstruction(PushInstr{SexpMarker})
	for _, expr := range arr {
		gen.AddInstruction(PushInstr{SexpMarker})
		gen.GenerateSyntaxQuote([]Sexp{expr})
		gen.AddInstruction(SquashInstr(0))
		gen.AddInstruction(ExplodeInstr(0))
	}
	gen.AddInstruction(VectorizeInstr(0))
	return nil
}

func (gen *Generator) generateSyntaxQuoteHash(arg Sexp) error {

	var hash *SexpHash
	switch a := arg.(type) {
	case *SexpHash:
		//good, required here
		hash = a
	default:
		return fmt.Errorf("arg to generateSyntaxQuoteHash() must be a hash; got %v", InspectType(a))
	}
	n, err := HashCountKeys(hash)
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstr{SexpMarker})
	for i := 0; i < n; i++ {
		// must reverse order here to preserve order on rebuild
		key := hash.KeyOrder[(n-i)-1]
		val, err := hash.HashGet(key)
		if err != nil {
			return err
		}
		// value first, since value comes second on rebuild
		gen.AddInstruction(PushInstr{SexpMarker})
		gen.GenerateSyntaxQuote([]Sexp{val})
		gen.AddInstruction(SquashInstr(0))
		gen.AddInstruction(ExplodeInstr(0))

		gen.AddInstruction(PushInstr{SexpMarker})
		gen.GenerateSyntaxQuote([]Sexp{key})
		gen.AddInstruction(SquashInstr(0))
		gen.AddInstruction(ExplodeInstr(0))
	}
	gen.AddInstruction(HashizeInstr{
		HashLen: n,
	})
	return nil
}

func (gen *Generator) GenerateSharpQuote(args []Sexp) error {
	if len(args) != 1 {
		return errors.New("sharp-quote takes exactly one argument")
	}
	arg := args[0]

	switch expr := arg.(type) {
	case SexpSymbol:
		gen.AddInstruction(GetInstr{sym: expr})
		return nil
	case *SexpPair:
		if err := gen.Generate(arg); err != nil {
			return err
		}
		gen.AddInstruction(RefSymInstr{})
		return nil
	default:
		return fmt.Errorf("sharp-quote resolve fail, unexpected s-expr %s", arg.SexpString())
	}
}

func prependCallName(macro *SexpFunction, sym SexpSymbol, args []Sexp) []Sexp {
	if macro.nameRegexp != nil {
		return append([]Sexp{SexpStr(sym.Name())}, args...)
	}
	return args
}
