package repl

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

var (
	history = GetHistory()
)

func getKeywords(vm *glisp.Environment) []string {
	var keywords []string
	for _, fn := range vm.GlobalFunctions() {
		if len(fn) > 1 && !strings.Contains(fn, "__") && !strings.Contains(fn, "/_") {
			keywords = append(keywords, fn)
		}
	}
	sort.Strings(keywords)
	return keywords
}

// find word backward until space or tab or newline
func findWordBackward(line string) (string, int) {
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == ' ' || line[i] == '\t' || line[i] == '\n' {
			return line[i+1:], i + 1
		}
	}
	return line, 0
}

func getLine(prefix string, keywords []string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	for _, kw := range history.Get() {
		line.AppendHistory(kw)
	}

	line.SetCompleter(func(line string) (c []string) {
		word, idx := findWordBackward(line)
		prependLParen := strings.HasPrefix(word, "(") || line == ""
		for _, n := range keywords {
			if prependLParen {
				n = "(" + n
			}
			if strings.HasPrefix(n, word) {
				c = append(c, line[0:idx]+n)
			}
		}
		return
	})

	if sentence, err := line.Prompt(prefix); err == nil {
		line.AppendHistory(sentence)
		history.Append(sentence)
		return sentence, nil
	} else {
		return "", err
	}
}

func readLine(keywords []string, waitMore bool) (string, error) {
	prefix := "> "
	if waitMore {
		prefix = ">> "
	}
	line, err := getLine(prefix, keywords)
	if err != nil {
		return "", err
	}
	return line, nil
}

func processDumpCommand(env *glisp.Environment, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment(os.Stderr)
	} else {
		err := env.DumpFunctionByName(os.Stderr, args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func repl(env *glisp.Environment) {
	stremRepl := NewStreamRepl(env)
	var waitMore bool
	var pendingCount int
	handleOutput := func(ret *Result) {
		waitMore = false
		pendingCount = 0
		expr, err := ret.Ret, ret.Err
		if err != nil {
			fmt.Println(ret.Err)
			return
		}

		if expr != glisp.SexpNull {
			fmt.Println(expr.SexpString())
		}
		env.BindGlobal(`$?`, expr)
	}

	for {
		line, err := readLine(getKeywords(env), waitMore)
		if err != nil {
			stremRepl.Stop()
			os.Exit(-1)
		}
		stremRepl.Write(line + "\n")
		pendingCount += len(strings.TrimSpace(line))

	WAIT_OUTPUT:
		select {
		case <-time.After(time.Millisecond * 100):
			waitMore = pendingCount > 0
			if stremRepl.IsRunning() {
				goto WAIT_OUTPUT
			}
		case ret := <-stremRepl.Out():
			handleOutput(ret)
		}
		select {
		case ret := <-stremRepl.Out():
			handleOutput(ret)
		default:
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

func newEnv() *glisp.Environment {
	env := glisp.New()
	modules := []func(*glisp.Environment) error{
		func(e *glisp.Environment) error { return e.ImportEval() },
		extensions.ImportCoreUtils,
		extensions.ImportRandom,
		extensions.ImportMathUtils,
		extensions.ImportTime,
		extensions.ImportChannels,
		extensions.ImportCoroutines,
		extensions.ImportRegex,
		extensions.ImportBase64,
		extensions.ImportJSON,
		extensions.ImportString,
		extensions.ImportOS,
		extensions.ImportHTTP,
	}
	for _, f := range modules {
		if err := f(env); err != nil {
			panic(err)
		}
	}
	env.AddNamedFunction("export-history", exportHistory, glisp.WithDoc(`(export-history FILE)`))
	env.AddNamedFunction("clear-history", clearHistory, glisp.WithDoc(`(clear-history)`))
	return env
}

type EnvOption func(*glisp.Environment)

func RunScript(file string, interactive bool, opts ...EnvOption) {
	env := newEnv()
	for _, fn := range opts {
		fn(env)
	}
	runScript(env, file)
	if interactive {
		repl(env)
	}
}

func CompileScript(file string, opts ...EnvOption) error {
	env := newEnv()
	for _, fn := range opts {
		fn(env)
	}
	_, err := env.ParseFile(file)
	return err
}

func Run(opts ...EnvOption) {
	env := newEnv()
	for _, fn := range opts {
		fn(env)
	}
	repl(env)
}
