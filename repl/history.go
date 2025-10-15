package repl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qjpcpu/glisp"
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

func (h *History) Truncate() {
	if h.file != nil {
		h.file.Close()
	}
	os.Remove(h.filename())
	h.setup()
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
	os.WriteFile(h.filename(), buf.Bytes(), 0755)
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

func exportHistory(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.WrongNumberArguments(name, args.Len(), 1)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf("filename should be string but got %v", args.Get(0).SexpString())
		}
		var buf bytes.Buffer
		for _, line := range history.Get() {
			buf.WriteString(line + "\n")
		}
		file := string(args.Get(0).(glisp.SexpStr))
		if strings.HasPrefix(file, "~") {
			home, _ := os.UserHomeDir()
			file = filepath.Join(home, strings.TrimPrefix(file, "~"))
		}
		os.WriteFile(file, buf.Bytes(), 0755)
		fmt.Printf("save history to %s\n", args.Get(0).(glisp.SexpStr))
		return glisp.SexpNull, nil
	}
}

func clearHistory(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 0 {
			return glisp.WrongNumberArguments(name, args.Len(), 0)
		}
		history.Truncate()
		fmt.Println("history truncated.")
		return glisp.SexpNull, nil
	}
}
