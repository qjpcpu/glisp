package glisp

import (
	"bytes"
	_ "embed"
	"strings"
)

var (
	//go:embed documentation.txt
	documentations string
	docMap         = make(map[string]string)
)

func QueryBuiltinDoc(funcName string) string {
	return docMap[funcName]
}

func init() {
	docMap = ParseDoc(documentations)
	documentations = ``
}

func ParseDoc(docstr string) map[string]string {
	dmap := make(map[string]string)
	lines := strings.Split(docstr, "\n")
	var foundHeader bool
	var header string
	docBuf := bytes.Buffer{}
	for i := 0; i < len(lines); i++ {
		if !foundHeader {
			header, foundHeader = getDocHeader(lines[i])
		} else {
			if h, ok := getDocHeader(lines[i]); ok {
				dmap[header] = strings.TrimSpace(docBuf.String())
				header = h
				docBuf.Reset()
			} else {
				docBuf.WriteString(lines[i] + "\n")
			}
		}
	}
	if header != `` && docBuf.Len() > 0 {
		dmap[header] = strings.TrimSpace(docBuf.String())
	}
	return dmap
}

func getDocHeader(line string) (string, bool) {
	lineBytes := []byte(strings.TrimSpace(line))
	var start, end int
	for i := 0; i < len(lineBytes); i++ {
		if lineBytes[i] != '=' {
			start = i
			break
		}
	}
	for i := len(lineBytes) - 1; i > start; i-- {
		if lineBytes[i] != '=' {
			end = i
			break
		}
	}
	if start >= 5 && end <= len(lineBytes)-5 && start < end {
		str := strings.TrimSpace(string(lineBytes[start : end+1]))
		return str, str != ``
	}
	return "", false
}
