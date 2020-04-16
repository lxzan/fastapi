package fastapi

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
)

var (
	logger    *zerolog.Logger
	useLogger = false
)

func init() {
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
