package extensions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"os/exec"

	"github.com/qjpcpu/glisp"
)

func ImportOS(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("os/read-file", GetReadFile)
	env.AddNamedFunction("os/open-file", GetOpenFile)
	env.AddNamedFunction("os/write-file", GetWriteFile)
	env.AddNamedFunction("os/file-exist?", GetExistFile)
	env.AddNamedFunction("os/read-dir", ReadDir)
	env.AddNamedFunction("os/remove-file", GetRemoveFile)
	env.AddNamedFunction("os/which", LookupPath)
	env.AddNamedFunction("os/exec", ExecCommand(&CommandOptions{AssertSuccess: false}))
	env.AddNamedFunction("os/exec!", ExecCommand(&CommandOptions{AssertSuccess: true}))
	env.AddNamedFunction("os/run", RunCommand)
	env.AddNamedFunction("os/env", Getenv)
	env.AddNamedFunction("os/setenv", Setenv)
	env.AddNamedFunction("os/mkdir", Mkdir)
	env.AddNamedFunction("os/args", GetOSArgs)
	//mustLoadScript(env.Environment, "os")
	return nil
}

type CommandOptions struct {
	Stdout, Stderr io.Writer
	AssertSuccess  bool
}

func ExecCommand(opts *CommandOptions) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
			if args.Len() != 1 {
				return glisp.WrongNumberArguments(name, args.Len(), 1)
			}
			var hash *glisp.SexpHash
			switch val := args.Get(0).(type) {
			case glisp.SexpStr:
				hash, _ = glisp.MakeHash([]glisp.Sexp{glisp.SexpStr("cmd"), args.Get(0)})
			case *glisp.SexpHash:
				hash = val
			default:
				return glisp.SexpNull, fmt.Errorf("argument of command must be string/hash but got %v", glisp.InspectType(args.Get(0)))
			}

			/* command */
			cmdstr := getHashStr(hash, "cmd")
			if cmdstr == "" {
				return glisp.SexpNull, errors.New("no cmd found")
			}
			cmd := exec.Command("bash", "-c", cmdstr)
			/* workding directory */
			if cwd := getHashStr(hash, "cwd"); cwd != "" {
				cmd.Dir = replaceHomeDirSymbol(cwd)
			}
			/* env */
			if env := getHashStrList(hash, "env"); len(env) > 0 {
				cmd.Env = mergeCurrentEnv(env)
			}
			var buf, errBuf bytes.Buffer
			/* stdout */
			if stdout := getHashWriter(hash, "stdout"); stdout != nil {
				cmd.Stdout = stdout
			} else if opts.Stdout != nil {
				cmd.Stdout = opts.Stdout
			} else {
				cmd.Stdout = &buf
			}
			/* stderr */
			cmd.Stderr = &errBuf
			combineErr := func(err0 error) error {
				if err0 == nil {
					return nil
				}
				err0 = fmt.Errorf("%w %s", err0, string(chomp(errBuf.Bytes())))
				if stderr := getHashWriter(hash, "stderr"); stderr != nil {
					stderr.Write(chomp(errBuf.Bytes()))
				} else if opts.Stderr != nil {
					opts.Stderr.Write(chomp(errBuf.Bytes()))
				}
				return err0
			}
			err := cmd.Run()
			if opts.AssertSuccess {
				if err != nil {
					return glisp.SexpNull, combineErr(err)
				}
				return glisp.SexpStr(chomp(buf.Bytes())), nil
			}
			if err != nil {
				return glisp.Cons(glisp.NewSexpInt(cmd.ProcessState.ExitCode()), glisp.SexpStr(combineErr(err).Error())), nil
			}
			return glisp.Cons(glisp.NewSexpInt(cmd.ProcessState.ExitCode()), glisp.SexpStr(chomp(buf.Bytes()))), nil
		}
	}
}

func GetReadFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		filename := replaceHomeDirSymbol(string(str))
		data, err := os.ReadFile(filename)
		if err != nil {
			return glisp.SexpNull, err
		}
		return glisp.NewSexpBytes(data), nil
	}
}

func Mkdir(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		filename := replaceHomeDirSymbol(string(str))
		os.MkdirAll(filename, 0755)
		return glisp.SexpNull, nil
	}
}

func GetWriteFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		filename := replaceHomeDirSymbol(string(str))
		if _, err := os.Stat(filepath.Dir(filename)); err != nil && os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(filename), 0755)
		}
		switch data := args.Get(1).(type) {
		case glisp.SexpStr:
			return glisp.SexpNull, os.WriteFile(filename, []byte(data), 0644)
		case glisp.SexpBytes:
			return glisp.SexpNull, os.WriteFile(filename, data.Bytes(), 0644)
		default:
			return glisp.SexpNull, fmt.Errorf("%s expect write string/bytes to file", name)
		}
	}
}

func ReadDir(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 && args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 1, 2)
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		var listType int
		if args.Len() == 2 {
			sy, ok := args.Get(1).(glisp.SexpSymbol)
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`second argument of %s should be symbol but got %v`, name, glisp.InspectType(args.Get(1)))
			}
			switch sy.Name() {
			case "file":
				listType = 1
			case "dir":
				listType = 2
			default:
				listType = 3
			}
		}
		dir := replaceHomeDirSymbol(string(str))
		fs, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return glisp.SexpNull, nil
			}
			return glisp.SexpNull, err
		}
		var files glisp.SexpArray
		for _, f := range fs {
			if listType == 0 || (listType == 1 && !f.IsDir()) || (listType == 2 && f.IsDir()) {
				files = append(files, glisp.SexpStr(f.Name()))
			}
		}
		return files, nil
	}
}

func GetRemoveFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		filename := replaceHomeDirSymbol(string(str))
		return glisp.SexpNull, os.RemoveAll(filename)
	}
}

func GetOpenFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() > 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 0/1 argument but got %v`, name, args.Len())
		}
		var file *os.File
		var err error
		if args.Len() == 1 {
			str, ok := args.Get(0).(glisp.SexpStr)
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
			}
			filename := replaceHomeDirSymbol(string(str))
			file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

		} else {
			file, err = os.CreateTemp(os.TempDir(), "glisp")
		}
		if err != nil {
			return glisp.SexpNull, err
		}
		return NewIO(file), nil
	}
}

func GetExistFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		filename := replaceHomeDirSymbol(string(str))
		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			return glisp.SexpBool(false), nil
		}
		return glisp.SexpBool(true), nil
	}
}

func replaceHomeDirSymbol(file string) string {
	if strings.HasPrefix(file, `~`) {
		if dir, err := os.UserHomeDir(); err == nil {
			file = filepath.Join(dir, strings.TrimPrefix(file, `~`))
		}
	}
	return file
}

func Getenv(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() == 0 {
			return glisp.SexpNull, errors.New("no arguments")
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args.Get(0)))
		}
		return glisp.SexpStr(os.Getenv(string(args.Get(0).(glisp.SexpStr)))), nil
	}
}

func Setenv(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 2 {
			return glisp.WrongNumberArguments(name, args.Len(), 2)
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args.Get(0)))
		}
		if !glisp.IsString(args.Get(1)) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args.Get(1)))
		}
		name := string(args.Get(0).(glisp.SexpStr))
		if name == `` {
			return glisp.SexpNull, errors.New("env variable name can't be empty")
		}
		os.Setenv(name, string(args.Get(1).(glisp.SexpStr)))
		return glisp.SexpNull, nil
	}
}

func chomp(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}

func RunCommand(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, errors.New("no command arguments")
		}
		if !glisp.IsString(args.Get(0)) {
			return glisp.SexpNull, errors.New("cmd must be string but got " + glisp.InspectType(args.Get(0)))
		}
		cmd := exec.Command("bash", "-c", string(args.Get(0).(glisp.SexpStr)))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return glisp.SexpNull, nil
	}
}

func getHashStr(hash *glisp.SexpHash, key string) string {
	val, err := hash.HashGet(glisp.SexpStr(key))
	if err != nil {
		return ""
	}
	str, ok := val.(glisp.SexpStr)
	if !ok {
		return ""
	}
	return string(str)
}

func getHashWriter(hash *glisp.SexpHash, key string) io.Writer {
	val, err := hash.HashGet(glisp.SexpStr(key))
	if err != nil {
		return nil
	}
	v, ok := val.(*SexpIO)
	if !ok {
		return nil
	}
	return v
}

func mergeCurrentEnv(env []string) []string {
	kv := make(map[string]string)
	for _, item := range os.Environ() {
		if arr := strings.SplitN(item, "=", 2); len(arr) == 2 {
			kv[arr[0]] = arr[1]
		}
	}
	for _, item := range env {
		if arr := strings.SplitN(item, "=", 2); len(arr) == 2 {
			kv[arr[0]] = arr[1]
		}
	}
	var ret []string
	for k, v := range kv {
		ret = append(ret, fmt.Sprintf("%v=%v", k, v))
	}
	return ret
}

func getHashStrList(hash *glisp.SexpHash, key string) []string {
	val, err := hash.HashGet(glisp.SexpStr(key))
	if err != nil {
		return nil
	}
	str, ok := val.(glisp.SexpArray)
	if !ok {
		return nil
	}
	var ret []string
	for _, item := range str {
		if glisp.IsString(item) {
			if s, ok := item.(glisp.SexpStr); ok && s != "" {
				ret = append(ret, string(s))
			}
		}
	}
	return ret
}

func GetOSArgs(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		var ret glisp.SexpArray
		for _, str := range os.Args {
			ret = append(ret, glisp.SexpStr(str))
		}
		return ret, nil
	}
}

func LookupPath(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args glisp.Args) (glisp.Sexp, error) {
		if args.Len() != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, args.Len())
		}
		str, ok := args.Get(0).(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string but got %v`, name, glisp.InspectType(args.Get(0)))
		}
		name := replaceHomeDirSymbol(string(str))
		p, err := exec.LookPath(name)
		if err != nil || p == "" {
			return glisp.SexpNull, nil
		}
		return glisp.SexpStr(p), nil
	}
}
