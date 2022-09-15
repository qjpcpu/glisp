package main

import (
	"flag"
	"github.com/qjpcpu/glisp/repl"
)

var (
	scriptFile  string
	interactive bool
)

func main() {
	flag.StringVar(&scriptFile, "f", "", "script file")
	flag.BoolVar(&interactive, "i", false, "enter repl mode")
	flag.Parse()

	if scriptFile != `` {
		repl.RunScript(scriptFile, interactive)
	} else {
		repl.Run()
	}
}
