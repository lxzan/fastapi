package fastapi

import (
	"fmt"
	"github.com/json-iterator/go"
	"net/http"
)

var defaultCatcher = func(ctx *Context, err interface{}) {
	myError, ok := err.(*Error)
	if ok {
		ctx.JSON(400, myError)
	} else {
		var msg = fmt.Sprintf("known exception: %v", err)
		ctx.Write(400, []byte(msg))
	}
}

var ContextType = struct {
	JSON string
}{
	JSON: "application/json",
}

type HandleFunc func(ctx *Context)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	next     bool
	Storage  Any
}

func (this *Context) Write(code int, body []byte) error {
	this.Response.WriteHeader(code)
	_, err := this.Response.Write(body)
	return err
}

func (this *Context) JSON(code int, v interface{}) error {
	this.Response.Header().Set("Content-Type", ContextType.JSON)
	body, err := jsoniter.Marshal(v)
	if err != nil {
		return err
	}
	return this.Write(code, body)
}

func (this *Context) Next() {
	this.next = true
}

func (this *Context) Abort() {
	this.next = false
}
