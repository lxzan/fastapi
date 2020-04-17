package fastapi

import (
	"bytes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"mime"
	"os"
	"strconv"
	"strings"
)

var (
	logger    *zerolog.Logger
	useLogger = false
)

func setLogger() {
	var out = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}
	L := log.Logger.Output(out).With().Logger()
	logger = &L
}

func Logger() HandlerFunc {
	useLogger = true
	return func(ctx *Context) {}
}

type CorsOption struct {
	AllowOrigin  string
	AllowMethods []string
	AllowHeaders []string
	MaxAge       int
}

func CORS(opt *CorsOption) HandlerFunc {
	if opt == nil {
		opt = &CorsOption{}
	}
	if opt.AllowOrigin == "" {
		opt.AllowOrigin = "*"
	}
	if len(opt.AllowMethods) == 0 {
		opt.AllowMethods = []string{"GET", "POST"}
	}
	if opt.MaxAge == 0 {
		opt.MaxAge = 3600
	}

	return func(ctx *Context) {
		header := ctx.Response.Header()
		if ctx.Request.Method == "OPTIONS" {
			ctx.Response.WriteHeader(204)
			header.Set("Access-Control-Allow-Origin", opt.AllowOrigin)
			header.Set("Access-Control-Allow-Methods", strings.Join(opt.AllowMethods, ","))
			header.Set("Access-Control-Allow-Headers", strings.Join(opt.AllowHeaders, ","))
			header.Set("Access-Control-Max-Age", strconv.Itoa(opt.MaxAge))
			ctx.Abort()
		} else {
			header.Set("Access-Control-Allow-Origin", opt.AllowOrigin)
			ctx.Next()
		}
	}
}

func bodyParser() HandlerFunc {
	return func(ctx *Context) {
		var body = []byte("")
		defer ctx.Storage.Set("body", body)

		contentType, _, err := mime.ParseMediaType(ctx.Request.Header.Get("Content-Type"))
		if err != nil {
			return
		}

		ctx.ContentType = contentType
		if contentType == ContentType.Form {
			err := ctx.Request.ParseForm()
			if err == nil {
				body = []byte(ctx.Request.Form.Encode())
			}
		} else {
			var buf = bytes.NewBufferString("")
			_, err := io.Copy(buf, ctx.Request.Body)
			if err == nil {
				body = buf.Bytes()
			}
		}
	}
}
