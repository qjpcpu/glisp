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
	env.AddNamedFunction("os/exec", ExecCommand(nil, nil))
	env.AddNamedFunction("os/run", RunCommand)
	env.AddNamedFunction("os/env", Getenv)
	env.AddNamedFunction("os/setenv", Setenv)
	mustLoadScript(env.Environment, "os")
	return nil
}

func ExecCommand(stdout, stderr io.Writer) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) == 0 {
				return glisp.SexpNull, errors.New("no command arguments")
			}
			var arguments []string
			for _, arg := range args {
				switch expr := arg.(type) {
				case glisp.SexpStr:
					arguments = append(arguments, string(expr))
				case glisp.SexpInt, glisp.SexpFloat, glisp.SexpBool, glisp.SexpChar, glisp.SexpSymbol:
					arguments = append(arguments, arg.SexpString())
				case glisp.SexpBytes:
					arguments = append(arguments, string(expr.Bytes()))
				default:
					return glisp.SexpNull, fmt.Errorf("argument of command must be string but got %v", glisp.InspectType(arg))
				}
			}
			var buf, errBuf bytes.Buffer
			cmd := exec.Command("bash", "-c", strings.Join(arguments, " "))
			if stdout != nil {
				cmd.Stdout = stdout
			} else {
				cmd.Stdout = &buf
			}
			if stderr != nil {
				cmd.Stderr = stderr
			} else {
				cmd.Stderr = &errBuf
			}
			err := cmd.Run()
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
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		str, ok := args[0].(glisp.SexpStr)
		if !ok {
			return glisp.SexpNull, fmt.Errorf(`%s argument should be string`, name)
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
			files = append(files, glisp.SexpStr(f.Name()))
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
