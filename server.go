package fastapi

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Runmode uint8

const (
	DebugMode Runmode = iota
	ProductMode
)

type Server struct {
	handlers   []HandleFunc
	getRouter  map[string][]HandleFunc
	postRouter map[string][]HandleFunc
	anyRouter  map[string][]HandleFunc
	Catch      func(ctx *Context, err interface{})
}

func New() *Server {
	return &Server{
		handlers:   make([]HandleFunc, 0),
		getRouter:  make(map[string][]HandleFunc),
		postRouter: make(map[string][]HandleFunc),
		anyRouter:  make(map[string][]HandleFunc),
	}
}

func (this *Server) Use(handles ...HandleFunc) {
	this.handlers = append(this.handlers, handles...)
}

func (this *Server) prepare(handlers ...HandleFunc) []HandleFunc {
	var h = make([]HandleFunc, 0)
	h = append(h, this.handlers...)
	h = append(h, handlers...)
	return h
}

func (this *Server) Group(prefix string, handlers ...HandleFunc) *Group {
	if len(handlers) == 0 {
		handlers = make([]HandleFunc, 0)
	}

	return &Group{
		server:   this,
		prefix:   prefix,
		handlers: this.prepare(handlers...),
	}
}

func (this *Server) GET(path string, handles ...HandleFunc) {
	this.getRouter[path] = this.prepare(handles...)
}

func (this *Server) POST(path string, handles ...HandleFunc) {
	this.postRouter[path] = this.prepare(handles...)
}

func (this *Server) ANY(path string, handles ...HandleFunc) {
	this.anyRouter[path] = this.prepare(handles...)
}

func (this *Server) Run(addr string) error {
	if this.Catch == nil {
		this.Catch = defaultCatcher
	}
	this.fprintRouters()

	var mode = "debug"
	if globalMode == ProductMode {
		mode = "product"
	}
	fmt.Printf("FastAPI server is listening on %s in %s mode.\n", addr, mode)
	return http.ListenAndServe(addr, this)
}

func (this *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var ctx = newContext(req, res)
	defer func() {
		if err := recover(); err != nil {
			this.Catch(ctx, err)
		}
	}()

	req.URL.Path = strings.TrimSpace(req.URL.Path)
	var handlers []HandleFunc
	var exist bool

	switch req.Method {
	case "GET":
		handlers, exist = this.getRouter[req.URL.Path]
	case "POST":
		handlers, exist = this.postRouter[req.URL.Path]
	}

	if !exist {
		handlers, exist = this.anyRouter[req.URL.Path]
	}

	if !exist {
		ctx.Write(404, []byte("handler not exist"))
	}

	for _, fn := range handlers {
		fn(ctx)
		if !ctx.next {
			break
		}
	}
}

func (this *Server) fprintRouters() {
	var m = make([]struct {
		Path   string
		Method string
		Fn     string
	}, 0)
	for k, fns := range this.getRouter {
		var n = len(fns)
		if n == 0 {
			continue
		}

		fn := runtime.FuncForPC(reflect.ValueOf(fns[n-1]).Pointer()).Name()
		m = append(m, struct {
			Path   string
			Method string
			Fn     string
		}{Path: k, Method: "GET", Fn: fn})
	}
	for k, fns := range this.postRouter {
		var n = len(fns)
		if n == 0 {
			continue
		}

		fn := runtime.FuncForPC(reflect.ValueOf(fns[n-1]).Pointer()).Name()
		m = append(m, struct {
			Path   string
			Method string
			Fn     string
		}{Path: k, Method: "POST", Fn: fn})
	}
	for k, fns := range this.anyRouter {
		var n = len(fns)
		if n == 0 {
			continue
		}

		fn := runtime.FuncForPC(reflect.ValueOf(fns[n-1]).Pointer()).Name()
		m = append(m, struct {
			Path   string
			Method string
			Fn     string
		}{Path: k, Method: "GET", Fn: fn}, struct {
			Path   string
			Method string
			Fn     string
		}{Path: k, Method: "POST", Fn: fn})
	}

	var length = len(m)
	for i := 0; i < length-1; i++ {
		for j := i + 1; j < length; j++ {
			if m[i].Path > m[j].Path {
				m[i], m[j] = m[j], m[i]
			}
		}
	}

	var maxLength = 0
	for i := 0; i < length; i++ {
		if len(m[i].Path) > maxLength {
			maxLength = len(m[i].Path)
		}
	}
	for _, item := range m {
		fmt.Printf(
			"%s %s -> %s\n",
			appendSpace("["+item.Method+"]", 6),
			appendSpace(item.Path, maxLength),
			item.Fn,
		)
	}
}
