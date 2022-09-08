package glisp

type SexpFunction struct {
	name       string
	user       bool
	nargs      int
	varargs    bool
	fun        Function
	userfun    UserFunction
	closeScope *Stack
	doc        string
}

func (sf *SexpFunction) SexpString() string {
	return "fn [" + sf.name + "]"
}

func (sf *SexpFunction) Clone() *SexpFunction {
	cp := *sf
	return &cp
}

func (sf *SexpFunction) Doc() string { return sf.doc }
