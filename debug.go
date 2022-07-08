package glisp

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

var (
	openDebug = os.Getenv(`LISP_DEBUG`) == `1`
	logger    = log.New(os.Stdout, ``, log.Ltime)
)

func debugln(vv ...interface{}) {
	if openDebug {
		vv = append([]interface{}{debugFileLine()}, vv...)
		logger.Println(vv...)
	}
}

func debug(v string, args ...interface{}) {
	if openDebug {
		logger.Printf(debugFileLine()+" "+v+"\n", args...)
	}
}

func debugFileLine() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", path.Base(file), line)
}
