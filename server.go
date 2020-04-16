package fastapi

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Runmode uint8

const (
	DebugMode Runmode = iota
	ProductMode
)

type Server struct {
	handlers   []HandlerFunc
	getRouter  map[string][]HandlerFunc
	postRouter map[string][]HandlerFunc
	anyRouter  map[string][]HandlerFunc
	catch      func(ctx *Context, err interface{})
}

func New() *Server {
	return &Server{
		handlers:   make([]HandlerFunc, 0),
		getRouter:  make(map[string][]HandlerFunc),
		postRouter: make(map[string][]HandlerFunc),
		anyRouter:  make(map[string][]HandlerFunc),
	}
}

func (this *Server) SetCatch(fn func(ctx *Context, err interface{})) {
	this.catch = fn
}

// global middleware
func (this *Server) Use(handles ...HandlerFunc) {
	for _, handle := range handles {
		name := runtime.FuncForPC(reflect.ValueOf(handle).Pointer()).Name()
		if !strings.Contains(name, "github.com/lxzan/fastapi.Logger") {
			this.handlers = append(this.handlers, handle)
		}
	}
}

func (this *Server) prepare(handlers ...HandlerFunc) []HandlerFunc {
	var h = make([]HandlerFunc, 0)
	h = append(h, handlers...)
	return h
}

func (this *Server) Group(prefix string, handlers ...HandlerFunc) *Group {
	if len(handlers) == 0 {
		handlers = make([]HandlerFunc, 0)
	}

	return &Group{
		server:   this,
		prefix:   prefix,
		handlers: handlers,
	}
}

func (this *Server) GET(path string, handles ...HandlerFunc) {
	this.getRouter[path] = this.prepare(handles...)
}

func (this *Server) POST(path string, handles ...HandlerFunc) {
	this.postRouter[path] = this.prepare(handles...)
}

func (this *Server) ANY(path string, handles ...HandlerFunc) {
	this.anyRouter[path] = this.prepare(handles...)
}

func (this *Server) Run(addr string) error {
	if this.catch == nil {
		this.catch = defaultCatcher
	}
	this.fprintRouters()

	var mode = "debug"
	if globalMode == ProductMode {
		mode = "product"
	}
	logger.Info().Msgf("FastAPI server is listening on %s in %s mode.", addr, mode)
	return http.ListenAndServe(addr, this)
}

func (this *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	t0 := time.Now().UnixNano()
	defer func() {
		if useLogger {
			t1 := time.Now().UnixNano()
			cost := fmt.Sprintf("%dms", (t1-t0)/1000000)
			logger.Info().Str("cost", cost).Msgf("%s %s", req.Method, req.URL.Path)
		}
	}()

	var ctx = newContext(req, res)
	defer func() {
		if err := recover(); err != nil {
			this.catch(ctx, err)
		}
	}()

	for _, fn := range this.handlers {
		fn(ctx)
		if !ctx.next {
			return
		}
	}

	req.URL.Path = strings.TrimSpace(req.URL.Path)
	var handlers []HandlerFunc
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
