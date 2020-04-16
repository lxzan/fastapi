package fastapi

type Group struct {
	server   *Server
	prefix   string
	handlers []HandlerFunc
}

func (this *Group) prepare(path string, handlers ...HandlerFunc) (p string, h []HandlerFunc) {
	if path == "/" {
		path = ""
	}
	p = this.prefix + path
	h = make([]HandlerFunc, 0)
	h = append(h, this.handlers...)
	h = append(h, handlers...)
	return
}

func (this *Group) GET(path string, handlers ...HandlerFunc) {
	p, h := this.prepare(path, handlers...)
	this.server.getRouter[p] = h
}

func (this *Group) POST(path string, handlers ...HandlerFunc) {
	p, h := this.prepare(path, handlers...)
	this.server.postRouter[p] = h
}

func (this *Group) ANY(path string, handlers ...HandlerFunc) {
	p, h := this.prepare(path, handlers...)
	this.server.anyRouter[p] = h
}

func (this *Group) Group(prefix string, handlers ...HandlerFunc) *Group {
	if len(handlers) == 0 {
		handlers = make([]HandlerFunc, 0)
	}
	if prefix == "/" {
		prefix = ""
	}
	var myHandlers = make([]HandlerFunc, 0)
	myHandlers = append(myHandlers, this.handlers...)
	myHandlers = append(myHandlers, handlers...)

	return &Group{
		server:   this.server,
		prefix:   this.prefix + prefix,
		handlers: myHandlers,
	}
}
