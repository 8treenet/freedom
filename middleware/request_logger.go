package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/8treenet/freedom"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12/context"
)

// NewRequestLogger .
func NewRequestLogger(traceIDName string, loggerConf ...*RequestLoggerConfig) func(context.Context) {
	l := DefaultLoggerConfig()
	if len(loggerConf) > 0 {
		l = loggerConf[0]
	}
	l.traceName = traceIDName
	return NewRequest(l)
}

type requestLoggerMiddleware struct {
	config *RequestLoggerConfig
}

// NewRequest .
func NewRequest(cfg *RequestLoggerConfig) context.Handler {
	l := &requestLoggerMiddleware{config: cfg}
	return l.ServeHTTP
}

// Serve serves the middleware
func (l *requestLoggerMiddleware) ServeHTTP(ctx context.Context) {
	// all except latency to string
	var status, method, path string
	var latency time.Duration
	var startTime, endTime time.Time
	startTime = time.Now()
	var reqBodyBys []byte
	if l.config.RequestRawBody {
		reqBodyBys, _ = ioutil.ReadAll(ctx.Request().Body)
		ctx.Request().Body.Close() //  must close
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBys))
	}

	work := freedom.ToWorker(ctx)
	freelog := NewLogger(l.config.traceName, work.Bus().Get(l.config.traceName))
	work.SetLogger(freelog)

	rawQuery := ctx.Request().URL.Query()
	ctx.Next()

	if !work.IsDeferRecycle() {
		loggerPool.Put(freelog)
	}

	// no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	status = strconv.Itoa(ctx.GetStatusCode())

	method = ctx.Method()
	path = ctx.Path()

	fieldsMessage := golog.Fields{}
	if l.config.IP {
		fieldsMessage["ip"] = ctx.RemoteAddr()
	}

	if headerKeys := l.config.MessageHeaderKeys; len(headerKeys) > 0 {
		header := ctx.Request().Header
		for _, key := range headerKeys {
			header.Get(key)
			msg := header.Get(key)
			if msg == "" {
				continue
			}
			fieldsMessage[key] = msg
		}
	}
	bus := freedom.ToWorker(ctx).Bus()
	traceInfo := bus.Get(l.config.traceName)
	if traceInfo != "" {
		fieldsMessage[l.config.traceName] = traceInfo
	}

	if l.config.RequestRawBody && len(reqBodyBys) > 0 {
		reqBodyBys = reqBodyBys[0:l.config.RequestRawBodyMaxLen]
		msg := string(reqBodyBys)
		if msg != "" {
			fieldsMessage["request"] = msg
		}
	}

	if ctxKeys := l.config.MessageContextKeys; len(ctxKeys) > 0 {
		for _, key := range ctxKeys {
			msg := ctx.Values().Get(key)
			if msg == nil {
				continue
			}
			fieldsMessage[key] = fmt.Sprint(msg)
		}
	}

	fieldsMessage["status"] = status
	fieldsMessage["latency"] = fmt.Sprint(latency)
	fieldsMessage["method"] = method
	fieldsMessage["path"] = path
	if len(rawQuery) > 0 && l.config.Query {
		fieldsMessage["query"] = rawQuery.Encode()
	}

	ctx.Application().Logger().Info(l.config.Title, fieldsMessage)
}
