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
	var addComma bool
	for elem := a; elem.head != SexpNull && elem.head != nil; {
		if addComma {
			buffer.WriteByte(',')
		} else {
			addComma = true
		}
		data, err := Marshal(elem.head)
		if err != nil {
			return nil, err
		}
		buffer.Write(data)
		var ok bool
		if elem, ok = elem.tail.(*SexpPair); !ok {
			break
		}
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

func (a *SexpHash) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte('{')
	keys := a.KeyOrder
	var addComma bool
	for _, key := range keys {
		val, _ := a.HashGet(key)
		keyj, err := Marshal(key)
		if err != nil {
			return nil, err
		}
		valj, err := Marshal(val)
		if err != nil {
			return nil, err
		}
		if addComma {
			buffer.WriteByte(',')
		} else {
			addComma = true
		}
		buffer.Write(keyj)
		buffer.WriteByte(':')
		buffer.Write(valj)
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

func stdMarshal(t interface{}) ([]byte, error) {
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
