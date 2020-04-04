package fastapi

import (
	"net/http"
)

type Server struct {
	getRouter  map[string][]HandleFunc
	postRouter map[string][]HandleFunc
	anyRouter  map[string][]HandleFunc
	Catch      func(ctx *Context, err interface{})
}

func New() *Server {
	return &Server{
		getRouter:  make(map[string][]HandleFunc),
		postRouter: make(map[string][]HandleFunc),
		anyRouter:  make(map[string][]HandleFunc),
	}
}

func (this *Server) GET(path string, handles ...HandleFunc) {
	this.getRouter[path] = handles
}

func (this *Server) Run(addr string) error {
	if this.Catch == nil {
		this.Catch = defaultCatcher
	}

	return http.ListenAndServe(addr, this)
}

func (this *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var ctx = &Context{
		Request:  req,
		Response: res,
		Storage:  Any{},
		next:     true,
	}

	defer func() {
		if err := recover(); err != nil {
			this.Catch(ctx, err)
		}
	}()

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
