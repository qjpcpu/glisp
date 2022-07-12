package glisp

import (
	"encoding/hex"
)

type SexpBytes struct {
	bytes []byte
}

func NewSexpBytes(b []byte) SexpBytes {
	return SexpBytes{bytes: b}
}

func NewSexpBytesByHex(hexstr string) (SexpBytes, error) {
	bs, err := hex.DecodeString(hexstr)
	if err != nil {
		return SexpBytes{}, err
	}
	return SexpBytes{bytes: bs}, nil
}

func (bs SexpBytes) SexpString() string {
	return `0B` + hex.EncodeToString(bs.bytes)
}

func (bs SexpBytes) Bytes() []byte {
	return bs.bytes
}
