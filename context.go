package fastapi

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/json-iterator/go"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"reflect"
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

func (this *Context) Bind(v interface{}) error {
	contentType, _, err := mime.ParseMediaType(this.Request.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	if contentType == ContextType.JSON {
		body, err := ioutil.ReadAll(this.Request.Body)
		if err != nil {
			return err
		}

		if err := jsoniter.Unmarshal(body, v); err != nil {
			return err
		}
	} else if contentType == ContextType.Form {
		if err := this.Request.ParseForm(); err != nil {
			return err
		}
		this.bindForm(this.Request.Form, reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
	} else {
		return errors.New("unknown content type")
	}

	this.setDefault(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
	err = validate.Struct(v)
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

func (this *Context) BindQuery(v interface{}) error {
	this.bindForm(this.Request.URL.Query(), reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
	this.setDefault(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem())
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

func (this *Context) setDefault(typs reflect.Type, values reflect.Value) {
	for i := 0; i < values.NumField(); i++ {
		t := typs.Field(i)
		v := values.Field(i)

		if t.Name[0] >= 'a' && t.Name[0] <= 'z' {
			continue
		}

		var kind = t.Type.Kind()
		if kind == reflect.Struct {
			this.setDefault(t.Type, v)
			continue
		} else if kind.String() == "ptr" {
			item := v.Interface()
			this.setDefault(reflect.TypeOf(item).Elem(), reflect.ValueOf(item).Elem())
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
			v.SetInt(ToInt(val))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(uint64(ToInt(val)))
		}
	}
}

func (this *Context) bindForm(query url.Values, typs reflect.Type, values reflect.Value) {
	for i := 0; i < values.NumField(); i++ {
		t := typs.Field(i)
		v := values.Field(i)

		if t.Name[0] >= 'a' && t.Name[0] <= 'z' {
			continue
		}

		var kind = t.Type.Kind()
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

func (this *Context) Next() {
	this.next = true
}

func (this *Context) Abort() {
	this.next = false
}
