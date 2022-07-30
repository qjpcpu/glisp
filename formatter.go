package glisp

import (
	"bytes"
	"strings"
)

func FormatCompact(expressions []Sexp) string {
	buffer := new(bytes.Buffer)
	for _, expr := range expressions {
		buffer.WriteString(expr.SexpString())
		buffer.WriteByte('\n')
	}
	return buffer.String()
}

func FormatPretty(expressions []Sexp) string {
	buffer := new(bytes.Buffer)
	for _, expr := range expressions {
		formatExpression(buffer, expr, 0, 0)
	}
	return buffer.String()
}

func formatExpression(buffer *bytes.Buffer, expr Sexp, indent, lastIndent int) {
	pair, ok := expr.(*SexpPair)
	if !ok {
		buffer.WriteString(expr.SexpString())
		return
	}
	pre := buffer.Bytes()
	if len(pre) > 0 && pre[len(pre)-1] != '(' {
		if lastIndent > 0 && lastIndent+1 < indent {
			indent = lastIndent + 1
		}
		buffer.WriteString("\n" + strings.Repeat(" ", indent*4))
		lastIndent = indent
	}
	buffer.WriteString("(")

	for {
		switch pair.tail.(type) {
		case *SexpPair:
			formatExpression(buffer, pair.Head(), indent+1, lastIndent)
			buffer.WriteString(" ")
			pair = pair.Tail().(*SexpPair)
			continue
		}
		break
	}

	formatExpression(buffer, pair.Head(), indent+1, lastIndent)

	if pair.Tail() == SexpNull {
		buffer.WriteString(")")
	} else {
		buffer.WriteString(" . ")
		formatExpression(buffer, pair.Tail(), indent+1, lastIndent)
		buffer.WriteString(")")
	}
}
