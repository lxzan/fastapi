package fastapi

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
)

var defaultCatcher = func(ctx *Context, err interface{}) {
	if err1, ok := err.(*Error); ok {
		ctx.JSON(400, err1)
		return
	}

	if globalMode == DebugMode {
		buf := make([]byte, 2048)
		n := runtime.Stack(buf, false)
		stackInfo := fmt.Sprintf("%s", buf[:n])
		logger.Error().Msg("Runtime Error")
		println(stackInfo)
	}

	ctx.JSON(500, NewError(Internal, "internal error"))
}

type HandlerFunc func(ctx *Context)

func newContext(req *http.Request, res http.ResponseWriter) *Context {
	return &Context{
		next:     true,
		Request:  req,
		Response: res,
		Storage:  Any{},
	}
}

type Context struct {
	next        bool
	Request     *http.Request
	Response    http.ResponseWriter
	Storage     Any
	ContentType string
}

func (c *Context) Write(code int, body []byte) error {
	c.Response.WriteHeader(code)
	_, err := c.Response.Write(body)
	return err
}

func (c *Context) JSON(code int, v interface{}) error {
	c.Response.Header().Set("Content-Type", ContentType.JSON)
	body, err := jsoniter.Marshal(v)
	if err != nil {
		return err
	}
	return c.Write(code, body)
}

func (c *Context) Bind(v interface{}) error {
	if c.Request.Method == "POST" {
		if c.ContentType == ContentType.JSON {
			body := c.GetBody()
			if len(body) == 0 {
				body = []byte("{}")
			}
			if err := jsoniter.Unmarshal(body, v); err != nil {
				return err
			}
		} else if c.ContentType == ContentType.Form {
			c.bindForm(c.Request.Form, reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
		} else {
			return errors.New("unknown content type")
		}
	} else if c.Request.Method == "GET" {
		c.bindForm(c.Request.URL.Query(), reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
	} else {
		return errors.New("unsupport http method")
	}

	c.setDefault(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
	err := validate.Struct(v)
	if err != nil {
		errs := err.(validator.ValidationErrors).Translate(trans)
		for k, v := range errs {
			return &TransError{
				Message: v,
				Field:   k,
			}
		}
	}

	return nil
}

func (c *Context) setDefault(typs reflect.Type, values reflect.Value) {
	for i := 0; i < values.NumField(); i++ {
		t := typs.Field(i)
		v := values.Field(i)

		if t.Name[0] >= 'a' && t.Name[0] <= 'z' {
			continue
		}

		var kind = t.Type.Kind()
		if kind == reflect.Struct {
			c.setDefault(t.Type, v)
			continue
		} else if kind.String() == "ptr" {
			item := v.Interface()
			c.setDefault(reflect.TypeOf(item).Elem(), reflect.ValueOf(item).Elem())
			continue
		}

		val := t.Tag.Get("default")
		if val == "" {
			continue
		}
		switch kind {
		case reflect.Bool:
			if v.Bool() == false && val == "true" {
				v.SetBool(true)
			}
		case reflect.String:
			if v.String() == "" {
				v.SetString(val)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() == 0 {
				v.SetInt(ToInt(val))
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() == 0 {
				v.SetUint(uint64(ToInt(val)))
			}
		}
	}
}

func (c *Context) bindForm(query url.Values, typs reflect.Type, values reflect.Value) {
	for i := 0; i < values.NumField(); i++ {
		t := typs.Field(i)
		v := values.Field(i)

		if t.Name[0] >= 'a' && t.Name[0] <= 'z' {
			continue
		}

		var kind = t.Type.Kind()
		if kind == reflect.Struct {
			c.bindForm(query, t.Type, v)
			continue
		} else if kind == reflect.Ptr {
			item := v.Interface()
			c.bindForm(query, reflect.TypeOf(item).Elem(), reflect.ValueOf(item).Elem())
			continue
		}

		tag := t.Tag.Get("json")
		if tag == "" {
			tag = t.Name
		}
		var vals []string
		var ok bool
		if kind == reflect.Slice {
			vals, ok = query[tag+"[]"]
		} else {
			vals, ok = query[tag]
		}
		if !ok || len(vals) == 0 {
			continue
		}

		var val = vals[0]
		switch kind {
		case reflect.String:
			v.SetString(val)
		case reflect.Bool:
			if val == "true" {
				v.SetBool(true)
			} else if val == "false" {
				v.SetBool(false)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(ToInt(val))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(uint64(ToInt(val)))
		case reflect.Slice:
			var vi = v.Interface()
			if arr, success := vi.([]int64); success {
				arr = make([]int64, 0)
				for i, _ := range vals {
					arr = append(arr, ToInt(vals[i]))
				}
				v.Set(reflect.ValueOf(arr))
				continue
			}

			if arr, success := vi.([]string); success {
				arr = make([]string, 0)
				for i, _ := range vals {
					arr = append(arr, vals[i])
				}
				v.Set(reflect.ValueOf(arr))
				continue
			}
		}
	}
}

func (c *Context) Next() {
	c.next = true
}

func (c *Context) Abort() {
	c.next = false
}

func (c *Context) ClientIP() string {
	var ip = c.Request.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return ""
	}
	if host == "::1" {
		return "127.0.0.1"
	}
	return host
}

func (c *Context) GetBody() []byte {
	v, _ := c.Storage.Get("body")
	return v.([]byte)
}
