package main

import (
	"flag"
	"fmt"

	"github.com/qjpcpu/glisp/repl"
)

var (
	scriptFile  string
	compileFile string
	interactive bool
)

func main() {
	flag.StringVar(&scriptFile, "f", "", "script file")
	flag.BoolVar(&interactive, "i", false, "enter repl mode")
	flag.StringVar(&compileFile, "c", "", "compile file")
	flag.Parse()

	if scriptFile != `` {
		repl.RunScript(scriptFile, interactive)
	} else if compileFile != `` {
		if err := repl.CompileScript(compileFile); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Compile passed.")
		}
	} else {
		repl.Run()
	}
}
