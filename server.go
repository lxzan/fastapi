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

func (r Runmode) String() string {
	if r == ProductMode {
		return "product"
	} else {
		return "debug"
	}
}

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
	s := &Server{
		handlers:   make([]HandlerFunc, 0),
		getRouter:  make(map[string][]HandlerFunc),
		postRouter: make(map[string][]HandlerFunc),
		anyRouter:  make(map[string][]HandlerFunc),
	}
	s.Use(bodyParser())
	return s
}

func (s *Server) SetCatch(fn func(ctx *Context, err interface{})) {
	s.catch = fn
}

// global middleware
func (s *Server) Use(handles ...HandlerFunc) {
	for _, handle := range handles {
		name := runtime.FuncForPC(reflect.ValueOf(handle).Pointer()).Name()
		if !strings.Contains(name, "github.com/lxzan/fastapi.Logger") {
			s.handlers = append(s.handlers, handle)
		}
	}
}

func (s *Server) prepare(handlers ...HandlerFunc) []HandlerFunc {
	var h = make([]HandlerFunc, 0)
	h = append(h, handlers...)
	return h
}

func (s *Server) Group(prefix string, handlers ...HandlerFunc) *Group {
	if len(handlers) == 0 {
		handlers = make([]HandlerFunc, 0)
	}

	return &Group{
		server:   s,
		prefix:   prefix,
		handlers: handlers,
	}
}

func (s *Server) GET(path string, handles ...HandlerFunc) {
	s.getRouter[path] = s.prepare(handles...)
}

func (s *Server) POST(path string, handles ...HandlerFunc) {
	s.postRouter[path] = s.prepare(handles...)
}

func (s *Server) ANY(path string, handles ...HandlerFunc) {
	s.anyRouter[path] = s.prepare(handles...)
}

func (s *Server) Run(addr string) error {
	if s.catch == nil {
		s.catch = defaultCatcher
	}
	s.fprintRouters()

	logger.Info().Msgf("FastAPI server is listening on %s in %s mode.", addr, globalMode.String())
	return http.ListenAndServe(addr, s)
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	accessMap.Add(req.URL.Path)
	t0 := time.Now().UnixNano()
	defer func() {
		accessMap.Sub(req.URL.Path)
		if useLogger {
			t1 := time.Now().UnixNano()
			cost := fmt.Sprintf("%dms", (t1-t0)/1000000)
			logger.Info().Str("cost", cost).Msgf("%s %s", req.Method, req.URL.Path)
		}
	}()

	var ctx = newContext(req, res)
	defer func() {
		if err := recover(); err != nil {
			s.catch(ctx, err)
		}
	}()

	for _, fn := range s.handlers {
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
		handlers, exist = s.getRouter[req.URL.Path]
	case "POST":
		handlers, exist = s.postRouter[req.URL.Path]
	}

	if !exist {
		handlers, exist = s.anyRouter[req.URL.Path]
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

func (s *Server) fprintRouters() {
	var m = make([]struct {
		Path   string
		Method string
		Fn     string
	}, 0)
	for k, fns := range s.getRouter {
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
	for k, fns := range s.postRouter {
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
	for k, fns := range s.anyRouter {
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
