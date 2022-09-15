package repl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const maxHistoryLines = 10000

func GetHistory() *History {
	h := &History{}
	h.setup()
	return h
}

type History struct {
	file *os.File
	List []string
}

func (h *History) Get() []string {
	return h.List
}

func (h *History) Append(v string) {
	if v != `` {
		h.List = append(h.List, v)
		if h.file != nil {
			h.file.WriteString(h.encodeLine(v) + "\n")
		}
	}
}

func (h *History) setup() {
	h.List = h.readAll()
	h.rotate()
	ofile, err := os.OpenFile(h.filename(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return
	}
	h.file = ofile
}

func (h *History) filename() string {
	return filepath.Join(os.TempDir(), `.glisp_history`)
}

func (h *History) rotate() {
	if len(h.List) > maxHistoryLines {
		h.List = h.List[len(h.List)-maxHistoryLines:]
	}
	buf := new(bytes.Buffer)
	for _, item := range h.List {
		buf.WriteString(h.encodeLine(item) + "\n")
	}
	ioutil.WriteFile(h.filename(), buf.Bytes(), 0755)
}

func (h *History) readAll() (ret []string) {
	file, err := os.Open(h.filename())
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if bs, err := base64.StdEncoding.DecodeString(line); err == nil {
			if len(bs) > 0 {
				ret = append(ret, string(bs))
			}
		}
	}
	return
}

func (h *History) encodeLine(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
