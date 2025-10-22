package glisp

import (
	"bytes"
	"encoding/json"
)

func Marshal(a Sexp) ([]byte, error) {
	switch expr := a.(type) {
	case SexpInt:
		return expr.MarshalJSON()
	case SexpBool:
		return expr.MarshalJSON()
	case SexpSentinel:
		return expr.MarshalJSON()
	case SexpFloat:
		return expr.MarshalJSON()
	case SexpArray:
		return expr.MarshalJSON()
	case SexpChar:
		return expr.MarshalJSON()
	case *SexpFunction:
		return expr.MarshalJSON()
	case *SexpHash:
		return expr.MarshalJSON()
	case *SexpPair:
		return expr.MarshalJSON()
	case SexpStr:
		return expr.MarshalJSON()
	case SexpSymbol:
		return expr.MarshalJSON()
	case SexpBytes:
		return expr.MarshalJSON()
	}
	if m, ok := a.(json.Marshaler); ok {
		return m.MarshalJSON()
	}
	return stdMarshal(a)
}

func (a SexpInt) MarshalJSON() ([]byte, error) {
	return a.v.MarshalText()
}

func (a *SexpFunction) MarshalJSON() ([]byte, error) {
	return stdMarshal(a.SexpString())
}

func (a SexpBytes) MarshalJSON() ([]byte, error) {
	return stdMarshal(a.bytes)
}

func (a SexpBool) MarshalJSON() ([]byte, error) {
	if bool(a) {
		return []byte(`true`), nil
	}
	return []byte(`false`), nil
}

func (a SexpStr) MarshalJSON() ([]byte, error) {
	return stdMarshal(string(a))
}

func (a SexpSymbol) MarshalJSON() ([]byte, error) {
	return stdMarshal(a.Name())
}

func (a SexpFloat) MarshalJSON() ([]byte, error) {
	if a.rawStr != "" {
		return []byte(a.rawStr), nil
	}
	return a.v.MarshalText()
}

func (a SexpChar) MarshalJSON() ([]byte, error) {
	return stdMarshal(rune(a))
}

func (a SexpSentinel) MarshalJSON() ([]byte, error) {
	return []byte(`null`), nil
}

func (a *SexpPair) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte('[')
	pair := a
	for {
		switch pair.tail.(type) {
		case *SexpPair:
			data, err := Marshal(pair.head)
			if err != nil {
				return nil, err
			}
			buffer.Write(data)
			buffer.WriteByte(',')
			pair = pair.tail.(*SexpPair)
			continue
		}
		break
	}
	data, err := Marshal(pair.head)
	if err != nil {
		return nil, err
	}
	buffer.Write(data)
	if pair.tail != SexpNull {
		data, err := Marshal(pair.tail)
		if err != nil {
			return nil, err
		}
		buffer.WriteByte(',')
		buffer.Write(data)
	}

	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

func (a *SexpHash) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte('{')
	var addComma bool
	var err error
	a.Visit(func(key Sexp, val Sexp) bool {
		keyj, err0 := Marshal(key)
		if err0 != nil {
			err = err0
			return false
		}
		valj, err0 := Marshal(val)
		if err0 != nil {
			err = err0
			return false
		}
		if addComma {
			buffer.WriteByte(',')
		} else {
			addComma = true
		}
		buffer.Write(keyj)
		buffer.WriteByte(':')
		buffer.Write(valj)
		return true
	})
	if err != nil {
		return nil, err
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func (a SexpArray) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte('[')
	for i, expr := range a {
		data, err := Marshal(expr)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			buffer.WriteByte(',')
		}
		buffer.Write(data)
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

func stdMarshal(t any) ([]byte, error) {
	if t == nil {
		return nil, nil
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(t); err != nil {
		return nil, err
	}
	ret := buffer.Bytes()
	// golang's encoder would always append a '\n', so we should drop it
	if size := len(ret); size > 0 && ret[size-1] == '\n' {
		ret = ret[:size-1]
	}
	return ret, nil
}
