package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/kataras/iris/v12/context"

	"github.com/ryanuber/columnize"
)

// NewRequestLogger .
func NewRequestLogger(traceIDName string, body bool) func(context.Context) {
	loggerConf := DefaultConfig()
	loggerConf.IP = false
	loggerConf.Query = false
	if body {
		loggerConf.Query = true
	}
	loggerConf.MessageContextKeys = append(loggerConf.MessageContextKeys, "logger_message", "response")
	loggerConf.MessageHeaderKeys = append(loggerConf.MessageHeaderKeys, traceIDName)
	loggerConf.TraceName = traceIDName
	return NewRequest(loggerConf)
}

type requestLoggerMiddleware struct {
	config      loggerConfig
	traceIDName string
}

func NewRequest(cfg ...loggerConfig) context.Handler {
	c := DefaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c.buildSkipper()
	l := &requestLoggerMiddleware{config: c}

	return l.ServeHTTP
}

// Serve serves the middleware
func (l *requestLoggerMiddleware) ServeHTTP(ctx context.Context) {
	// skip logs and serve the main request immediately
	if l.config.skip != nil {
		if l.config.skip(ctx) {
			ctx.Next()
			return
		}
	}

	// all except latency to string
	var status, ip, method, path string
	var latency time.Duration
	var startTime, endTime time.Time
	startTime = time.Now()
	reqBodyBys, _ := ioutil.ReadAll(ctx.Request().Body)
	ctx.Request().Body.Close() //  must close
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBys))

	work := freedom.ToWorker(ctx)
	freelog := newFreedomLogger(l.config.TraceName, work.Bus().Get(l.config.TraceName))
	work.Store().Set("logger_impl", freelog)
	ctx.Next()

	if !work.IsDeferRecycle() {
		loggerPool.Put(freelog)
	}

	// no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	if l.config.Status {
		status = strconv.Itoa(ctx.GetStatusCode())
	}

	if l.config.IP {
		ip = ctx.RemoteAddr()
	}

	if l.config.Method {
		method = ctx.Method()
	}

	if l.config.Path {
		if l.config.Query {
			path = ctx.Request().URL.RequestURI()
		} else {
			path = ctx.Path()
		}
	}

	var message interface{}
	var headerMessage interface{}
	if headerKeys := l.config.MessageHeaderKeys; len(headerKeys) > 0 {
		bus := freedom.ToWorker(ctx).Bus()
		for _, key := range headerKeys {
			msg := bus.Get(key)
			if msg == "" {
				continue
			}
			msg = key + ":" + msg
			if headerMessage == nil {
				headerMessage = msg
			} else {
				headerMessage = fmt.Sprintf(" %v %v", headerMessage, msg)
			}
		}
	}

	if l.config.Query {
		cl := ctx.GetContentLength()
		if cl < 512 && len(reqBodyBys) < 512 {
			msg := string(reqBodyBys)
			msg = strings.Replace(msg, "\n", "", -1)
			msg = strings.Replace(msg, " ", "", -1)
			if msg != "" {
				msg = "request:" + msg
				if message == nil {
					message = msg
				} else {
					message = fmt.Sprintf(" %v %v", message, msg)
				}
			}
		}
	}

	if ctxKeys := l.config.MessageContextKeys; len(ctxKeys) > 0 {
		for _, key := range ctxKeys {
			msg := ctx.Values().Get(key)
			if msg == nil {
				continue
			}
			msg = key + ":" + fmt.Sprint(msg)
			if message == nil {
				message = msg
			} else {
				message = fmt.Sprintf(" %v %v", message, msg)
			}
		}
	}

	// print the logs
	if logFunc := l.config.LogFunc; logFunc != nil {
		logFunc(endTime, latency, status, ip, method, path, headerMessage, message)
		return
	} else if logFuncCtx := l.config.LogFuncCtx; logFuncCtx != nil {
		logFuncCtx(ctx, latency)
		return
	}

	if l.config.Columns {
		endTimeFormatted := endTime.Format("2006/01/02 - 15:04:05")
		output := Columnize(endTimeFormatted, latency, status, ip, method, path, headerMessage, message)
		ctx.Application().Logger().Printer.Output.Write([]byte(output))
		return
	}
	// no new line, the framework's logger is responsible how to render each log.
	line := fmt.Sprintf("%v %4v %s %s %s", status, latency, ip, method, path)
	if headerMessage != nil {
		line += fmt.Sprintf(" %v", headerMessage)
	}
	if message != nil {
		line += fmt.Sprintf(" %v", message)
	}

	ctx.Application().Logger().Info(line)
}

// Columnize formats the given arguments as columns and returns the formatted output,
// note that it appends a new line to the end.
func Columnize(nowFormatted string, latency time.Duration, status, ip, method, path string, message interface{}, headerMessage interface{}) string {
	titles := "Time | Status | Latency | IP | Method | Path"
	line := fmt.Sprintf("%s | %v | %4v | %s | %s | %s", nowFormatted, status, latency, ip, method, path)
	if message != nil {
		titles += " | Message"
		line += fmt.Sprintf(" | %v", message)
	}

	if headerMessage != nil {
		titles += " | HeaderMessage"
		line += fmt.Sprintf(" | %v", headerMessage)
	}

	outputC := []string{
		titles,
		line,
	}
	output := columnize.SimpleFormat(outputC) + "\n"
	return output
}
