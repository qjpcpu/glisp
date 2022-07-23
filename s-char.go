package glisp

import (
	"strconv"
	"strings"
)

type SexpChar rune

func (c SexpChar) SexpString() string {
	switch int32(c) {
	case 39:
		/* char is ' */
		return `#\'`
	case '"':
		return `#\"`
	case '(':
		return `#\(`
	case ')':
		return `#\)`
	case '[':
		return `#\[`
	case ']':
		return `#\]`
	case '#':
		return `#\#`
	case '~':
		return `#\~`
	case '`':
		return "#\\`"
	case '{':
		return `#\{`
	case '}':
		return `#\}`
	case ';':
		return `#\;`
	case '\\':
		return `#\\`
	}
	return "#" + strings.Trim(strconv.QuoteRune(rune(c)), "'")
}
