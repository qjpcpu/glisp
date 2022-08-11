package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

var (
	history  []string
	keywords []string
)

func registKeywords(vm *glisp.Environment) {
	for _, fn := range vm.GlobalFunctions() {
		if len(fn) > 1 {
			keywords = append(keywords, fn)
		}
	}
}

func findWordBackward(line string) (string, int) {
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == ' ' || line[i] == '\t' || line[i] == '\n' {
			return line[i+1:], i + 1
		}
	}
	return line, 0
}

func getLine(prefix string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	for _, kw := range history {
		line.AppendHistory(kw)
	}

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range keywords {
			n = "(" + n
			word, idx := findWordBackward(line)
			if strings.HasPrefix(n, word) {
				c = append(c, line[0:idx]+n)
			}
		}
		return
	})

	if sentence, err := line.Prompt(prefix); err == nil {
		line.AppendHistory(sentence)
		history = append(history, sentence)
		return sentence, nil
	} else {
		return "", err
	}
}

func readLine(waitMore bool) (string, error) {
	prefix := "> "
	if waitMore {
		prefix = ">> "
	}
	line, err := getLine(prefix)
	if err != nil {
		return "", err
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
	stremRepl := NewStreamRepl(env)
	var waitMore bool
	var pendingCount int
	for {
		line, err := readLine(waitMore)
		if err != nil {
			stremRepl.Stop()
			os.Exit(-1)
		}
		stremRepl.Write(line + "\n")
		pendingCount += len(strings.TrimSpace(line))

		select {
		case <-time.After(time.Millisecond * 100):
			waitMore = pendingCount > 0
		case ret := <-stremRepl.Out():
			waitMore = false
			pendingCount = 0
			expr, err := ret.Ret, ret.Err
			if err != nil {
				fmt.Print(env.GetStackTrace(err))
				env.Clear()
				break
			}

			if expr != glisp.SexpNull {
				fmt.Println(expr.SexpString())
			}
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

func fmtScript(env *glisp.Environment, fname string) {
	expressions, err := env.ParseFile(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println(glisp.FormatPretty(expressions))
}

var (
	fmtFile string
)

func main() {
	flag.StringVar(&fmtFile, "f", "", "format file")

	env := glisp.New()
	env.ImportEval()
	extensions.ImportRandom(env)
	extensions.ImportMathUtils(env)
	extensions.ImportTime(env)
	extensions.ImportChannels(env)
	extensions.ImportCoroutines(env)
	extensions.ImportRegex(env)
	extensions.ImportBase64(env)
	extensions.ImportCoreUtils(env)
	extensions.ImportJSON(env)
	extensions.ImportString(env)
	extensions.ImportIO(env)
	extensions.ImportContainerUtils(env)

	flag.Parse()

	registKeywords(env)

	if fmtFile != "" {
		fmtScript(env, fmtFile)
		return
	}

	if args := os.Args; len(args) > 1 {
		runScript(env, args[1])
	} else {
		repl(env)
	}
}
