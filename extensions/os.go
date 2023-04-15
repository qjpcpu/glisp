package extensions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"os/exec"

	"github.com/qjpcpu/glisp"
)

func ImportOS(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedFunction("os/read-file", GetReadFile)
	env.AddNamedFunction("os/write-file", GetWriteFile)
	env.AddNamedFunction("os/file-exist?", GetExistFile)
	env.AddNamedFunction("os/read-dir", ReadDir)
	env.AddNamedFunction("os/remove-file", GetRemoveFile)
	env.AddNamedFunction("os/exec", ExecCommand(&CommandOptions{AssertSuccess: false}))
	env.AddNamedFunction("os/exec!", ExecCommand(&CommandOptions{AssertSuccess: true}))
	env.AddNamedFunction("os/run", RunCommand)
	env.AddNamedFunction("os/env", Getenv)
	env.AddNamedFunction("os/setenv", Setenv)
	env.AddNamedFunction("os/mkdir", Mkdir)
	//mustLoadScript(env.Environment, "os")
	return nil
}

type CommandOptions struct {
	Stdout, Stderr io.Writer
	AssertSuccess  bool
}

func ExecCommand(opts *CommandOptions) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) != 1 {
				return glisp.WrongNumberArguments(name, len(args), 1)
			}
			switch args[0].(type) {
			case glisp.SexpStr:
				args[0], _ = glisp.MakeHash([]glisp.Sexp{glisp.SexpStr("cmd"), args[0]})
			case *glisp.SexpHash:
			default:
				return glisp.SexpNull, fmt.Errorf("argument of command must be string/hash but got %v", glisp.InspectType(args[0]))
			}

			hash := args[0].(*glisp.SexpHash)
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
			if stderr := getHashWriter(hash, "stderr"); stderr != nil {
				cmd.Stderr = stderr
			} else if opts.Stderr != nil {
				cmd.Stderr = opts.Stderr
			} else {
				cmd.Stderr = &errBuf
			}
			err := cmd.Run()
			if opts.AssertSuccess {
				if err != nil {
					return glisp.SexpNull, err
				}
				return glisp.SexpStr(chomp(buf.Bytes())), nil
			}
			if err != nil {
				return glisp.Cons(glisp.NewSexpInt(cmd.ProcessState.ExitCode()), glisp.SexpStr(chomp(errBuf.Bytes()))), nil
			}
			return glisp.Cons(glisp.NewSexpInt(cmd.ProcessState.ExitCode()), glisp.SexpStr(chomp(buf.Bytes()))), nil
		}
	}
}

func GetReadFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		filename := replaceHomeDirSymbol(string(str))
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return glisp.SexpNull, err
		}
		return glisp.NewSexpBytes(data), nil
	}
}

func Mkdir(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		filename := replaceHomeDirSymbol(string(str))
		os.MkdirAll(filename, 0755)
		return glisp.SexpNull, nil
	}
}

func GetWriteFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		filename := replaceHomeDirSymbol(string(str))
		if _, err := os.Stat(filepath.Dir(filename)); err != nil && os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(filename), 0755)
		}
		switch data := args[1].(type) {
		case glisp.SexpStr:
			return glisp.SexpNull, ioutil.WriteFile(filename, []byte(data), 0644)
		case glisp.SexpBytes:
			return glisp.SexpNull, ioutil.WriteFile(filename, data.Bytes(), 0644)
		default:
			return glisp.SexpNull, fmt.Errorf("%s expect write string/bytes to file", name)
		}
	}
}

func ReadDir(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 && len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 1, 2)
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		var listType int
		if len(args) == 2 {
			sy, ok := args[1].(glisp.SexpSymbol)
			if !ok {
				return glisp.SexpNull, fmt.Errorf(`second argument of %s should be symbol but got %v`, name, glisp.InspectType(args[1]))
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
		fs, err := ioutil.ReadDir(dir)
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
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
		}
		filename := replaceHomeDirSymbol(string(str))
		return glisp.SexpNull, os.RemoveAll(filename)
	}
}

func GetExistFile(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, fmt.Errorf(`%s expect 1 argument but got %v`, name, len(args))
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
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
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) == 0 {
			return glisp.SexpNull, errors.New("no arguments")
		}
		if !glisp.IsString(args[0]) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args[0]))
		}
		return glisp.SexpStr(os.Getenv(string(args[0].(glisp.SexpStr)))), nil
	}
}

func Setenv(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !glisp.IsString(args[0]) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args[0]))
		}
		if !glisp.IsString(args[1]) {
			return glisp.SexpNull, fmt.Errorf("env variable should be string but got %v", glisp.InspectType(args[1]))
		}
		name := string(args[0].(glisp.SexpStr))
		if name == `` {
			return glisp.SexpNull, errors.New("env variable name can't be empty")
		}
		os.Setenv(name, string(args[1].(glisp.SexpStr)))
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
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("no command arguments")
		}
		if !glisp.IsString(args[0]) {
			return glisp.SexpNull, errors.New("cmd must be string but got " + glisp.InspectType(args[0]))
		}
		cmd := exec.Command("bash", "-c", string(args[0].(glisp.SexpStr)))
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
	v, ok := val.(*SexpWriter)
	if !ok {
		return nil
	}
	return v.w
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
