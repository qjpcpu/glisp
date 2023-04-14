package repl

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/glisp/extensions"
)

var (
	history = GetHistory()
)

func getKeywords(vm *glisp.Environment) []KeyWord {
	var keywords []KeyWord
	for _, fn := range vm.GlobalFunctions() {
		if len(fn) > 1 && !strings.Contains(fn, "__") && !strings.Contains(fn, "/_") {
			sg := KeyWord{Word: fn}
			if expr, ok := vm.FindObject(fn); ok && glisp.IsFunction(expr) {
				sg.Desc = expr.(*glisp.SexpFunction).Doc()
			} else if mac, ok := vm.FindMacro(fn); ok {
				sg.Desc = mac.Doc()
			} else {
				sg.Desc = glisp.QueryBuiltinDoc(fn)
			}
			keywords = append(keywords, sg)
		}
	}
	sort.SliceStable(keywords, func(i, j int) bool {
		return keywords[i].Word < keywords[j].Word
	})
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

func getLine(liner LinerProducer, prefix string, keywords []KeyWord) (string, error) {
	line := liner()
	defer line.Close()
	result, err := line.Prompt(prefix, PromptOption{History: history.List, Keywords: keywords})
	if result != "" {
		history.Append(result)
	}
	return result, err
}

func readLine(liner LinerProducer, keywords []KeyWord, waitMore bool) (string, error) {
	prefix := "> "
	if waitMore {
		prefix = ">> "
	}
	line, err := getLine(liner, prefix, keywords)
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

func repl(liner LinerProducer, env *glisp.Environment) {
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
		line, err := readLine(liner, getKeywords(env), waitMore)
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

func runScript(env *Repl, fname string) {
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

func NewRepl() *Repl {
	env := glisp.New()
	if err := extensions.ImportAll(env); err != nil {
		panic(err)
	}
	env.AddNamedFunction("export-history", exportHistory, glisp.WithDoc(`(export-history FILE)`))
	env.AddNamedFunction("clear-history", clearHistory, glisp.WithDoc(`(clear-history)`))
	return &Repl{Environment: env, liner: Default()}
}

type Repl struct {
	*glisp.Environment
	liner LinerProducer
}

func SetLiner(l LinerProducer) ReplOption { return func(r *Repl) { r.liner = l } }

type ReplOption func(*Repl)

func RunScript(file string, interactive bool, opts ...ReplOption) {
	env := NewRepl()
	for _, fn := range opts {
		fn(env)
	}
	runScript(env, file)
	if interactive {
		repl(env.liner, env.Environment)
	}
}

func CompileScript(file string, opts ...ReplOption) error {
	env := NewRepl()
	for _, fn := range opts {
		fn(env)
	}
	_, err := env.ParseFile(file)
	return err
}

func Run(opts ...ReplOption) {
	env := NewRepl()
	for _, fn := range opts {
		fn(env)
	}
	repl(env.liner, env.Environment)
}
