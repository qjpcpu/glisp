package repl

import (
	"strings"

	"github.com/peterh/liner"
)

type KeyWord struct {
	Word string
	Desc string
}

type PromptOption struct {
	History  []string
	Keywords []KeyWord
}

type Liner interface {
	Prompt(prefix string, o PromptOption) (string, error)
	Close() error
}

type defaultLiner struct {
	line *liner.State
}

func (line *defaultLiner) Close() error { return line.line.Close() }

func (self *defaultLiner) Prompt(prefix string, opts PromptOption) (string, error) {
	for _, kw := range opts.History {
		self.line.AppendHistory(kw)
	}
	self.line.SetCompleter(func(line string) (c []string) {
		word, idx := findWordBackward(line)
		prependLParen := strings.HasPrefix(word, "(") || line == ""
		for _, w := range opts.Keywords {
			n := w.Word
			if prependLParen {
				n = "(" + w.Word
			}
			if strings.HasPrefix(n, word) {
				c = append(c, line[0:idx]+n)
			}
		}
		return
	})
	return self.line.Prompt(prefix)
}

type LinerProducer func() Liner

func Default() func() Liner {
	return func() Liner {
		line := liner.NewLiner()
		line.SetCtrlCAborts(true)
		line.SetMultiLineMode(true)
		return &defaultLiner{line: line}
	}
}
