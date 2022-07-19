package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

var history []string

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func getLine(prefix string) (string, error) {
	line := prompt.Input(prefix, completer,
		prompt.OptionHistory(history),
		prompt.OptionPrefixTextColor(prompt.Yellow),
	)
	return line, nil
}

func isBalanced(str string) bool {
	parens := 0
	squares := 0

	for _, c := range str {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			squares++
		case ']':
			squares--
		}
	}

	return parens == 0 && squares == 0
}

func getExpression() (string, error) {
	line, err := getLine("> ")
	if err != nil {
		return "", err
	}
	for !isBalanced(line) {
		nextline, err := getLine(">> ")
		if err != nil {
			return "", err
		}
		line += "\n" + nextline
	}
	history = append(history, strings.ReplaceAll(line, "\n", " "))
	return line, nil
}

func processDumpCommand(env *glisp.Environment, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment()
	} else {
		err := env.DumpFunctionByName(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func repl(env *glisp.Environment) {
	for {
		line, err := getExpression()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		parts := strings.Split(line, " ")
		if strings.TrimSpace(line) == "" {
			continue
		}

		if parts[0] == "quit" {
			break
		}

		if parts[0] == "dump" {
			processDumpCommand(env, parts[1:])
			continue
		}

		expr, err := env.EvalString(line)
		if err != nil {
			fmt.Print(env.GetStackTrace(err))
			env.Clear()
			continue
		}

		if expr != glisp.SexpNull {
			fmt.Println(expr.SexpString())
		}
	}
}

func runScript(env *glisp.Environment, fname string) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	_, err = env.Run()
	if err != nil {
		fmt.Print(env.GetStackTrace(err))
		os.Exit(-1)
	}
}

func main() {
	env := glisp.New()
	env.ImportEval()
	extensions.ImportRandom(env)
	extensions.ImportTime(env)
	extensions.ImportChannels(env)
	extensions.ImportCoroutines(env)
	extensions.ImportRegex(env)
	extensions.ImportBase64(env)
	extensions.ImportCoreUtils(env)
	extensions.ImportJSON(env)
	extensions.ImportString(env)

	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		runScript(env, args[0])
	} else {
		repl(env)
	}
}
