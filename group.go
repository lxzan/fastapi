package fastapi

type Group struct {
	server   *Server
	prefix   string
	handlers []HandlerFunc
}

func (g *Group) prepare(path string, handlers ...HandlerFunc) (p string, h []HandlerFunc) {
	if path == "/" {
		path = ""
	}
	p = g.prefix + path
	h = make([]HandlerFunc, 0)
	h = append(h, g.handlers...)
	h = append(h, handlers...)
	return
}

func (g *Group) GET(path string, handlers ...HandlerFunc) {
	p, h := g.prepare(path, handlers...)
	g.server.getRouter[p] = h
}

func (g *Group) POST(path string, handlers ...HandlerFunc) {
	p, h := g.prepare(path, handlers...)
	g.server.postRouter[p] = h
}

func (g *Group) ANY(path string, handlers ...HandlerFunc) {
	p, h := g.prepare(path, handlers...)
	g.server.anyRouter[p] = h
}

func (g *Group) Group(prefix string, handlers ...HandlerFunc) *Group {
	if len(handlers) == 0 {
		handlers = make([]HandlerFunc, 0)
	}
	if prefix == "/" {
		prefix = ""
	}
	var myHandlers = make([]HandlerFunc, 0)
	myHandlers = append(myHandlers, g.handlers...)
	myHandlers = append(myHandlers, handlers...)

	return &Group{
		server:   g.server,
		prefix:   g.prefix + prefix,
		handlers: myHandlers,
	}
}
