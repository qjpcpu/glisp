package main

import (
	"errors"
	"io"

	"unicode/utf8"

	"github.com/qjpcpu/glisp"
)

type Result struct {
	Ret glisp.Sexp
	Err error
}

type StreamRepl struct {
	env    *glisp.Environment
	input  chan rune
	output chan *Result
	stopc  chan struct{}
}

func NewStreamRepl(env *glisp.Environment) *StreamRepl {
	sr := &StreamRepl{
		env:    env,
		input:  make(chan rune, 1024),
		output: make(chan *Result, 10),
		stopc:  make(chan struct{}, 1),
	}
	go sr.start()
	return sr
}

func (sr *StreamRepl) start() {
START:
	sr.drainInput()
	lexer := glisp.NewLexerFromStream(sr)
	for {
		select {
		case <-sr.stopc:
			return
		default:
		}
		expr, err := sr.replOnce(lexer)
		if err != nil {
			sr.output <- &Result{Err: err}
			sr.env.Clear()
			goto START
		}
		sr.output <- &Result{Ret: expr}
	}
}

func (sr *StreamRepl) replOnce(lexer *glisp.Lexer) (glisp.Sexp, error) {
	expr, err := glisp.ParseExpression(glisp.NewParser(lexer, sr.env))
	if err != nil {
		return glisp.SexpNull, errors.New(sr.env.GetStackTrace(err))
	}
	if err = sr.env.LoadExpressions([]glisp.Sexp{expr}); err != nil {
		return glisp.SexpNull, errors.New(sr.env.GetStackTrace(err))
	}
	ret, err := sr.env.Run()
	if err != nil {
		return glisp.SexpNull, errors.New(sr.env.GetStackTrace(err))
	}
	return ret, nil
}

func (sr *StreamRepl) drainInput() {
	for {
		select {
		case <-sr.input:
		default:
			return
		}
	}
}

func (sr *StreamRepl) Write(str string) {
	bs := []rune(str)
	for _, b := range bs {
		sr.input <- b
	}
}

func (sr *StreamRepl) ReadRune() (r rune, size int, err error) {
	select {
	case r = <-sr.input:
		size = utf8.RuneLen(r)
	case <-sr.stopc:
		err = io.EOF
		return
	}
	return
}

func (sr *StreamRepl) Stop() {
	close(sr.stopc)
}

func (sr *StreamRepl) Out() <-chan *Result {
	return sr.output
}
