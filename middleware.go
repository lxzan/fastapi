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
	"time"
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
			header.Set("Access-Control-Allow-Origin", opt.AllowOrigin)
			header.Set("Access-Control-Allow-Methods", strings.Join(opt.AllowMethods, ","))
			header.Set("Access-Control-Allow-Headers", strings.Join(opt.AllowHeaders, ","))
			header.Set("Access-Control-Max-Age", strconv.Itoa(opt.MaxAge))
			ctx.Response.WriteHeader(204)
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
		contentType, _, err := mime.ParseMediaType(ctx.Request.Header.Get("Content-Type"))
		if err != nil {
			ctx.Storage.Set("body", body)
			return
		}

		ctx.ContentType = contentType
		if contentType == ContentType.Form {
			err := ctx.Request.ParseForm()
			if err == nil {
				body = []byte(ctx.Request.Form.Encode())
			}
			ctx.Storage.Set("body", body)
		} else {
			var buf = bytes.NewBufferString("")
			_, err := io.Copy(buf, ctx.Request.Body)
			if err == nil {
				body = buf.Bytes()
			}
			ctx.Storage.Set("body", body)
		}
	}
}

// limit concurrent access speed
func Limit(n int64) HandlerFunc {
	return func(ctx *Context) {
		for {
			var running = accessMap.Get(ctx.Request.URL.Path)
			if running > n {
				time.Sleep(10 * time.Millisecond)
			} else {
				break
			}
		}
	}
}
