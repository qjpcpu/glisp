package glisp

func IsArray(expr Sexp) bool {
	switch expr.(type) {
	case SexpArray:
		return true
	}
	return false
}

func IsList(expr Sexp) bool {
	if expr == SexpNull {
		return true
	}
	switch list := expr.(type) {
	case *SexpPair:
		return IsList(list.tail)
	}
	return false
}

func IsFloat(expr Sexp) bool {
	switch expr.(type) {
	case SexpFloat:
		return true
	}
	return false
}

func IsInt(expr Sexp) bool {
	switch expr.(type) {
	case SexpInt:
		return true
	}
	return false
}

func IsString(expr Sexp) bool {
	switch expr.(type) {
	case SexpStr:
		return true
	}
	return false
}

func IsChar(expr Sexp) bool {
	switch expr.(type) {
	case SexpChar:
		return true
	}
	return false
}

func IsNumber(expr Sexp) bool {
	switch expr.(type) {
	case SexpFloat:
		return true
	case SexpInt:
		return true
	case SexpChar:
		return true
	}
	return false
}

func IsSymbol(expr Sexp) bool {
	switch expr.(type) {
	case SexpSymbol:
		return true
	}
	return false
}

func IsBool(expr Sexp) bool {
	switch expr.(type) {
	case SexpBool:
		return true
	}
	return false
}

func IsHash(expr Sexp) bool {
	switch expr.(type) {
	case *SexpHash:
		return true
	}
	return false
}

func IsBytes(expr Sexp) bool {
	switch expr.(type) {
	case SexpBytes:
		return true
	}
	return false
}

func IsFunction(expr Sexp) bool {
	switch expr.(type) {
	case *SexpFunction:
		return true
	}
	return false
}

func IsZero(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpInt:
		return e.IsZero()
	case SexpChar:
		return int(e) == 0
	case SexpFloat:
		return e.Cmp(NewSexpFloat(0)) == 0
	}
	return false
}

func IsEmpty(expr Sexp) bool {
	if expr == SexpNull {
		return true
	}

	switch e := expr.(type) {
	case SexpArray:
		return len(e) == 0
	case *SexpHash:
		return HashIsEmpty(e)
	case SexpStr:
		return len(e) == 0
	case SexpBytes:
		return len(e.bytes) == 0
	}

	return false
}

func isComparable(v Sexp) bool {
	_, ok := v.(Comparable)
	return ok
}

func IsTruthy(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpBool:
		return bool(e)
	case SexpSentinel:
		return e != SexpNull
	}
	return true
}
