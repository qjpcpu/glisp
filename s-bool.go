package glisp

type SexpBool bool

func (b SexpBool) SexpString() string {
	if b {
		return "true"
	}
	return "false"
}
