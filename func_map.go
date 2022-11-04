package glisp

type FuncMap struct {
	funcs map[int]*SexpFunction
	fuzzy []*SexpFunction
}

func NewFuncMap() *FuncMap {
	return &FuncMap{funcs: make(map[int]*SexpFunction)}
}

func (fm *FuncMap) Add(sym SexpSymbol, f *SexpFunction) {
	if f.nameRegexp == nil {
		fm.funcs[sym.number] = f
		return
	}
	for i, v := range fm.fuzzy {
		if v.nameRegexp.String() == f.nameRegexp.String() {
			fm.fuzzy[i] = f
			return
		}
	}
	fm.fuzzy = append(fm.fuzzy, f)
}

func (fm *FuncMap) Names() (ret []string) {
	for _, f := range fm.funcs {
		ret = append(ret, f.name)
	}
	return
}

func (fm *FuncMap) Find(sym SexpSymbol) (*SexpFunction, bool) {
	if v, ok := fm.funcs[sym.number]; ok {
		return v, true
	}
	for _, f := range fm.fuzzy {
		if f.nameRegexp.MatchString(sym.name) {
			return f, true
		}
	}
	return nil, false
}

func (fm *FuncMap) Clone() *FuncMap {
	funcs := make(map[int]*SexpFunction)
	for k, v := range fm.funcs {
		funcs[k] = v
	}
	fuzzy := make([]*SexpFunction, len(fm.fuzzy))
	copy(fuzzy, fm.fuzzy)
	return &FuncMap{funcs: funcs, fuzzy: fuzzy}
}
