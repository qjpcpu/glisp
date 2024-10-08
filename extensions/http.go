package extensions

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
			return processHTTP(name, withRespStatus, newHttpReq(), env, args)
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
	URLs                  []string
	MultiURL              bool
	Header                http.Header
	Data                  io.Reader
	Timeout               time.Duration
	IncludeHeaderInOutput bool
	Method                string
	Outfile               string
	IgnoreErr             bool
	Proxy                 SexpDialer
	ProxyURL              *url.URL
}

func newHttpReq() request {
	return request{Header: make(http.Header), Timeout: 15 * time.Second}
}

func (r request) copy() request {
	r2 := r
	r2.Header = make(http.Header)
	for k := range r.Header {
		r2.Header.Set(k, r.Header.Get(k))
	}
	return r2
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
	"-x": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			switch dialer := val.(type) {
			case SexpDialer:
				req.Proxy = dialer
			case glisp.SexpStr:
				urlstr := string(dialer)
				if !strings.HasPrefix(urlstr, "http") {
					urlstr = "http://" + urlstr
				}
				purl, err := url.Parse(urlstr)
				if err != nil {
					return nil, err
				}
				req.ProxyURL = purl
			default:
				return nil, fmt.Errorf("-x need dialer but got %v", querySexpType(env, val))
			}
			return req, nil
		},
	},
	"-ignore-error": {
		needValue: false,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			req.IgnoreErr = true
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
	"-o": {
		needValue: true,
		decorator: func(env *glisp.Environment, req *request, val glisp.Sexp) (*request, error) {
			if glisp.IsString(val) {
				req.Outfile = string(val.(glisp.SexpStr))
			}
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

func prepareHTTPReq(name string, hreq *request, env *glisp.Environment, args []glisp.Sexp) (bool, error) {
	/* parse user options */
	var functions []func(*request) (*request, error)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if option, ok := _httpIsOption(arg); ok {
			name, _ := readSymOrStr(arg)
			if option.needValue {
				if i+1 >= len(args) {
					return false, fmt.Errorf("%s need an argument but got nothing", name)
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
			fmtURL := func(a glisp.Sexp) string {
				str := string(a.(glisp.SexpStr))
				if !strings.HasPrefix(str, "http") {
					str = "http://" + str
				}
				return str
			}
			switch val := arg.(type) {
			case glisp.SexpStr:
				functions = append(functions, func(req *request) (*request, error) {
					req.URLs = []string{fmtURL(arg)}
					return req, nil
				})
			case glisp.SexpArray:
				functions = append(functions, func(req *request) (*request, error) {
					for _, item := range val {
						req.URLs = append(req.URLs, fmtURL(item))
					}
					req.MultiURL = true
					return req, nil
				})
			default:
				return false, fmt.Errorf("unknown option %v(%v)", arg.SexpString(), querySexpType(env, arg))
			}

		}
	}

	/* decorate request by user options */
	for _, fn := range functions {
		var err error
		if hreq, err = fn(hreq); err != nil {
			return false, fmt.Errorf("%s build request fail %v", name, err)
		}
	}

	envProxy := os.Getenv("HTTP_PROXY")
	if envProxy == "" {
		envProxy = os.Getenv("http_proxy")
	}
	if envProxy != "" {
		if !strings.HasPrefix(envProxy, "http") {
			envProxy = "http://" + envProxy
		}
		if u, err := url.Parse(envProxy); err == nil {
			hreq.ProxyURL = u
		}
	}

	return len(hreq.URLs) != 0, nil
}

func evalHTTP(name string, hreq request, env *glisp.Environment, withRespStatus bool) (glisp.Sexp, error) {
	if !hreq.MultiURL {
		return evalSingleHTTP(name, hreq, hreq.URLs[0], env, withRespStatus)
	}
	ret := make([]glisp.Sexp, len(hreq.URLs))
	var ferr error
	wg := new(sync.WaitGroup)
	for i := range hreq.URLs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res, err := evalSingleHTTP(name, hreq, hreq.URLs[idx], env, withRespStatus)
			if err != nil {
				ferr = err
			} else {
				ret[idx] = res
			}
		}(i)
	}
	wg.Wait()
	if ferr != nil {
		return glisp.SexpNull, ferr
	}
	lb := glisp.NewListBuilder()
	for _, item := range ret {
		lb.Add(item)
	}
	return lb.Get(), nil
}

func evalSingleHTTP(name string, hreq request, urlstr string, env *glisp.Environment, withRespStatus bool) (glisp.Sexp, error) {
	/* pick method */
	method := strings.ToUpper(strings.TrimPrefix(name, "http/"))
	if name == `http/curl` {
		method = `GET`
		if hreq.Method != "" {
			method = hreq.Method
		}
	}
	/* build http request */
	req, err := http.NewRequest(method, urlstr, hreq.Data)
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
	var cli HttpClient
	if hreq.Proxy != nil {
		key := fmt.Sprintf("transport-with-proxy:%p", hreq.Proxy)
		cli = &http.Client{Timeout: hreq.Timeout, Transport: getTransport(key, func(tr *http.Transport) {
			tr.DialContext = hreq.Proxy
		})}
	} else {
		key := "transport:default"
		var proxy func(*http.Request) (*url.URL, error)
		if hreq.ProxyURL != nil {
			key = fmt.Sprintf("transport:%s", hreq.ProxyURL)
			proxy = func(*http.Request) (*url.URL, error) { return hreq.ProxyURL, nil }
		}
		cli = &http.Client{Timeout: hreq.Timeout, Transport: getTransport(key, func(tr *http.Transport) {
			tr.Proxy = proxy
		})}
	}
	if hreq.Verbose {
		cli = newDebugHttpClient(cli, os.Stderr)
	}
	resp, err := cli.Do(req)
	if err != nil {
		if hreq.IgnoreErr {
			errBytes := []byte(fmt.Sprintf("[GLISP_HTTP_ERROR]%v", err.Error()))
			if withRespStatus {
				return glisp.Cons(glisp.NewSexpInt(http.StatusBadGateway), glisp.NewSexpBytes(errBytes)), nil
			} else {
				return glisp.NewSexpBytes(errBytes), nil
			}
		}
		return glisp.SexpNull, fmt.Errorf("%s %v fail %v", req.Method, req.URL.String(), err)
	}

	/* parse response */
	defer resp.Body.Close()
	var bs []byte
	if hreq.Outfile != "" {
		os.MkdirAll(filepath.Dir(hreq.Outfile), 0755)
		file, err := os.OpenFile(hreq.Outfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return glisp.SexpNull, err
		}
		defer file.Close()
		io.Copy(file, resp.Body)
	} else {
		bs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return glisp.SexpNull, err
		}
	}

	var responseBody glisp.Sexp = glisp.NewSexpBytes(bs)
	if hreq.IncludeHeaderInOutput {
		var kvs []glisp.Sexp
		kvs = append(kvs, glisp.SexpStr("Status"), glisp.SexpStr(resp.Status))
		kvs = append(kvs, glisp.SexpStr("StatusCode"), glisp.SexpStr(strconv.FormatInt(int64(resp.StatusCode), 10)))
		for key := range resp.Header {
			kvs = append(kvs, glisp.SexpStr(key), glisp.SexpStr(resp.Header.Get(key)))
		}
		header, _ := glisp.MakeHash(kvs)
		responseBody = glisp.Cons(header, responseBody)
	}

	/* return cons cell for curl */
	if withRespStatus {
		return glisp.Cons(glisp.NewSexpInt(resp.StatusCode), responseBody), nil
	}
	return responseBody, nil

}

func processHTTP(name string, withRespStatus bool, req request, env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) < 1 {
		return glisp.WrongNumberArguments(name, len(args), 1, glisp.Many)
	}
	if evalNow, err := prepareHTTPReq(name, &req, env, args); err != nil {
		return glisp.SexpNull, err
	} else if !evalNow {
		return glisp.MakeUserFunction(env.GenSymbol().Name(), func(env0 *glisp.Environment, args0 []glisp.Sexp) (glisp.Sexp, error) {
			return processHTTP(name, withRespStatus, req.copy(), env0, args0)
		}), nil
	}
	return evalHTTP(name, req, env, withRespStatus)
}

type SexpDialer func(ctx context.Context, network, addr string) (net.Conn, error)

func (sd SexpDialer) SexpString() string { return sd.TypeName() }

func (sd SexpDialer) TypeName() string {
	return "func(ctx context.Context, network string, addr string) (net.Conn, error)"
}

func MakeDialer(dialer func(context.Context, string, string) (net.Conn, error)) SexpDialer {
	return SexpDialer(dialer)
}

var (
	transportPool = make(map[string]*http.Transport)
	transportMu   sync.RWMutex
)

func getTransport(key string, mws ...func(*http.Transport)) *http.Transport {
	transportMu.RLock()
	if val, ok := transportPool[key]; ok {
		transportMu.RUnlock()
		return val
	}
	transportMu.RUnlock()
	transportMu.Lock()
	defer transportMu.Unlock()
	if val, ok := transportPool[key]; ok {
		return val
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	for _, fn := range mws {
		fn(tr)
	}
	transportPool[key] = tr
	return tr
}
