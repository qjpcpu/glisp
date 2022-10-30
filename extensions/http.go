package extensions

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/qjpcpu/glisp"
)

func ImportHTTP(vm *glisp.Environment) error {
	env := autoAddDoc(vm)
	env.AddNamedMacro("http/get", DoHTTPMacro(false))
	env.AddNamedFunction("http/get", DoHTTP(false))

	env.AddNamedMacro("http/post", DoHTTPMacro(false))
	env.AddNamedFunction("http/post", DoHTTP(false))

	env.AddNamedMacro("http/put", DoHTTPMacro(false))
	env.AddNamedFunction("http/put", DoHTTP(false))

	env.AddNamedMacro("http/patch", DoHTTPMacro(false))
	env.AddNamedFunction("http/patch", DoHTTP(false))

	env.AddNamedMacro("http/delete", DoHTTPMacro(false))
	env.AddNamedFunction("http/delete", DoHTTP(false))

	env.AddNamedMacro("http/options", DoHTTPMacro(false))
	env.AddNamedFunction("http/options", DoHTTP(false))

	env.AddNamedMacro("http/head", DoHTTPMacro(false))
	env.AddNamedFunction("http/head", DoHTTP(false))

	env.AddNamedMacro("http/curl", DoHTTPMacro(true))
	env.AddNamedFunction("http/curl", DoHTTP(true))
	return nil
}

func DoHTTPMacro(withRespStatus bool) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		realFn := glisp.MakeUserFunction(name, DoHTTP(withRespStatus)(name))
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			for i := 0; i < len(args); i++ {
				arg := args[i]
				if option, ok := _httpIsOption(arg); ok {
					args[i] = glisp.MakeList([]glisp.Sexp{
						env.MakeSymbol("quote"),
						arg,
					})
					if option.needValue {
						i++
					}
				}
			}
			lb := glisp.NewListBuilder()
			lb.Add(realFn)
			for i := range args {
				lb.Add(args[i])
			}
			return lb.Get(), nil
		}
	}
}

/* (http/get|post|put|patch|delete OPTIONS URL) */
func DoHTTP(withRespStatus bool) glisp.NamedUserFunction {
	return func(name string) glisp.UserFunction {
		return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
			if len(args) < 1 {
				return glisp.WrongNumberArguments(name, len(args), 1, glisp.Many)
			}

			/* parse user options */
			hreq := newHttpReq()
			var functions []func(*request) (*request, error)
			for i := 0; i < len(args); i++ {
				arg := args[i]
				if option, ok := _httpIsOption(arg); ok {
					name, _ := readSymOrStr(arg)
					if option.needValue {
						if i+1 >= len(args) {
							return glisp.SexpNull, fmt.Errorf("%s need an argument but got nothing", name)
						}
						val := args[i+1]
						functions = append(functions, func(req *request) (*request, error) {
							return option.decorator(env, req, val)
						})
						if name == `-H` {
							functions[0], functions[len(functions)-1] = functions[len(functions)-1], functions[0]
						}
						i++
					} else {
						functions = append(functions, func(req *request) (*request, error) {
							return option.decorator(env, req, nil)
						})
					}
				} else {
					if !glisp.IsString(arg) {
						return glisp.SexpNull, fmt.Errorf("unknown option %v(%v)", arg.SexpString(), querySexpType(env, arg))
					}
					functions = append(functions, func(req *request) (*request, error) {
						req.URL = string(arg.(glisp.SexpStr))
						if !strings.HasPrefix(req.URL, "http") {
							req.URL = "http://" + req.URL
						}
						return req, nil
					})
				}
			}

			/* decorate request by user options */
			for _, fn := range functions {
				var err error
				if hreq, err = fn(hreq); err != nil {
					return glisp.SexpNull, fmt.Errorf("%s build request fail %v", name, err)
				}
			}

			/* pick method */
			method := strings.ToUpper(strings.TrimPrefix(name, "http/"))
			if name == `http/curl` {
				method = `GET`
				if hreq.Method != "" {
					method = hreq.Method
				}
			}

			/* build http request */
			req, err := http.NewRequest(method, hreq.URL, hreq.Data)
			if err != nil {
				return glisp.SexpNull, fmt.Errorf("%s build request fail %v", name, err)
			}

			/* populate headers */
			for k, vals := range hreq.Header {
				for _, val := range vals {
					req.Header.Add(k, val)
				}
			}

			/* perform http request */
			var cli HttpClient = &http.Client{Timeout: hreq.Timeout, Transport: httpTransport}
			if hreq.Verbose {
				cli = newDebugHttpClient(cli, os.Stderr)
			}
			resp, err := cli.Do(req)
			if err != nil {
				return glisp.SexpNull, fmt.Errorf("%s %v fail %v", req.Method, req.URL.String(), err)
			}

			/* parse response */
			defer resp.Body.Close()
			bs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return glisp.SexpNull, err
			}

			if hreq.IncludeHeaderInOutput {
				buf := new(bytes.Buffer)
				buf.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
				for key := range resp.Header {
					buf.WriteString(fmt.Sprintf("%s: %s\n", key, resp.Header.Get(key)))
				}
				buf.WriteByte('\n')
				buf.Write(bs)
				bs = buf.Bytes()
			}

			/* return cons cell for curl */
			if withRespStatus {
				return glisp.Cons(glisp.NewSexpInt(resp.StatusCode), glisp.NewSexpBytes(bs)), nil
			}
			return glisp.NewSexpBytes(bs), nil
		}
	}
}

func _httpIsOption(expr glisp.Sexp) (httpOption, bool) {
	if glisp.IsSymbol(expr) {
		opt, ok := availableHttpOptions[expr.(glisp.SexpSymbol).Name()]
		return opt, ok
	}
	return httpOption{}, false
}

type request struct {
	Verbose               bool
	URL                   string
	Header                http.Header
	Data                  io.Reader
	Timeout               time.Duration
	IncludeHeaderInOutput bool
	Method                string
}

func newHttpReq() *request {
	return &request{Header: make(http.Header), Timeout: 15 * time.Second}
}

type requestDecorator func(*glisp.Environment, *request, glisp.Sexp) (*request, error)

type httpOption struct {
	decorator requestDecorator
	needValue bool
}

var availableHttpOptions = map[string]httpOption{
	"-H": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			if !glisp.IsString(val) {
				return nil, fmt.Errorf("-H option value must be a string but got %v", querySexpType(env, val))
			}
			expr := string(val.(glisp.SexpStr))
			if !strings.Contains(expr, ":") {
				return nil, fmt.Errorf("bad format %s, -H option value must like header:value", expr)
			}
			arr := strings.SplitN(expr, ":", 2)
			req.Header.Add(strings.TrimSpace(arr[0]), strings.TrimSpace(arr[1]))
			return req, nil
		},
	},
	"-i": {
		needValue: false,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			req.IncludeHeaderInOutput = true
			return req, nil
		},
	},
	"-timeout": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			// such as "300ms", "-1.5h" or "2h45m".
			// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
			switch expr := val.(type) {
			case glisp.SexpInt:
				req.Timeout = time.Second * time.Duration(expr.ToInt())
			case glisp.SexpStr:
				dur, err := time.ParseDuration(string(expr))
				if err != nil {
					return nil, fmt.Errorf("bad -timeout %v", err)
				}
				req.Timeout = dur
			default:
				return nil, fmt.Errorf(`-timeout value should be integer or duration string such as  "1ns", "1us" (or "1µs"), "1ms", "1s", "1m", "1h"`)
			}
			return req, nil
		},
	},
	"-X": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			if !glisp.IsString(val) {
				return nil, fmt.Errorf("-X Method need string but got %v", querySexpType(env, val))
			}
			req.Method = strings.ToUpper(string(val.(glisp.SexpStr)))
			return req, nil
		},
	},
	"-v": {
		needValue: false,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			req.Verbose = true
			return req, nil
		},
	},
	"-d": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			if val == glisp.SexpNull {
				return req, nil
			}
			switch expr := val.(type) {
			case glisp.SexpStr:
				req.Data = bytes.NewBufferString(string(expr))
			case glisp.SexpBytes:
				req.Data = bytes.NewBuffer(expr.Bytes())
			case glisp.SexpArray:
				req.Header.Set("Content-Type", "application/json")
				data, err := glisp.Marshal(expr)
				if err != nil {
					return nil, err
				}
				req.Data = bytes.NewBuffer(data)
			case *glisp.SexpPair:
				req.Header.Set("Content-Type", "application/json")
				data, err := glisp.Marshal(expr)
				if err != nil {
					return nil, err
				}
				req.Data = bytes.NewBuffer(data)
			case *glisp.SexpHash:
				if strings.Contains(req.Header.Get("Content-Type"), `form`) {
					val := make(url.Values)
					expr.Foreach(func(k glisp.Sexp, v glisp.Sexp) bool {
						val.Add(_httpToFormValue(k), _httpToFormValue(v))
						return true
					})
					req.Data = bytes.NewBufferString(val.Encode())
				} else {
					req.Header.Set("Content-Type", "application/json")
					data, err := glisp.Marshal(expr)
					if err != nil {
						return nil, err
					}
					req.Data = bytes.NewBuffer(data)
				}
			case glisp.SexpInt, glisp.SexpFloat, glisp.SexpBool:
				req.Data = bytes.NewBufferString(val.SexpString())
			default:
				return nil, fmt.Errorf("bad value of -d: %s", val.SexpString())
			}
			return req, nil
		},
	},
}

func querySexpType(env *glisp.Environment, val glisp.Sexp) string {
	t, _ := glisp.GetTypeFunction("type")(env, []glisp.Sexp{val})
	return t.SexpString()
}

func _httpToFormValue(expr glisp.Sexp) string {
	if glisp.IsString(expr) {
		return string(expr.(glisp.SexpStr))
	}
	return expr.SexpString()
}

var httpTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
}

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func newDebugHttpClient(c HttpClient, w io.Writer) HttpClient {
	return &httpDebugClient{writer: w, cli: c}
}

type httpDebugClient struct {
	writer io.Writer
	cli    HttpClient
}

func (c *httpDebugClient) Do(req *http.Request) (*http.Response, error) {
	var payload []byte
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		payload = data
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}
	now := time.Now()
	fmt.Fprintf(c.writer, "%s %s\n", req.Method, req.URL.String())
	for k := range req.Header {
		fmt.Fprintf(c.writer, "%s: %s\n", k, req.Header.Get(k))
	}
	fmt.Fprintln(c.writer)
	if len(payload) > 0 {
		fmt.Fprintf(c.writer, "%s\n", string(payload))
		fmt.Fprintln(c.writer)
	}
	res, err := c.cli.Do(req)
	if err != nil {
		fmt.Fprintln(c.writer, err.Error())
		return res, err
	}
	fmt.Fprintf(c.writer, "%s cost:%v\n", res.Status, time.Since(now))
	for k := range res.Header {
		fmt.Fprintf(c.writer, "%s: %s\n", k, res.Header.Get(k))
	}
	respBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	fmt.Fprintln(c.writer)
	fmt.Fprintf(c.writer, "%s\n", string(respBytes))
	res.Body = io.NopCloser(bytes.NewBuffer(respBytes))
	return res, err
}
