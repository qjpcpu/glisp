package main

import (
	"flag"
	"github.com/qjpcpu/glisp/repl"
	"os"
)

var (
	fmtFile string
)

func main() {
	flag.StringVar(&fmtFile, "f", "", "format file")
	flag.Parse()

	if fmtFile != "" {
		repl.FormatScript(fmtFile)
		return
	}

	if args := os.Args; len(args) > 1 {
		repl.RunScript(args[1])
	} else {
		repl.Run()
	}
}
