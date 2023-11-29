package main

import (
	"os"

	"github.com/qjpcpu/glisp/repl"
)

func main() {
	switch len(os.Args) {
	case 0:
	case 1:
		repl.Run()
	case 2:
		if os.Args[1] == "-i" {
			/* glisp -i */
			repl.Run()
		} else {
			/* glisp FILE */
			file := os.Args[1]
			os.Args = os.Args[1:]
			repl.RunScript(file, false)
		}
	default:
		if os.Args[1] == "-i" {
			/* glisp -i FILE args... */
			file := os.Args[2]
			os.Args = os.Args[2:]
			repl.RunScript(file, true)
		} else {
			/* glisp FILE args... */
			file := os.Args[1]
			os.Args = os.Args[1:]
			repl.RunScript(file, false)
		}
	}
}
