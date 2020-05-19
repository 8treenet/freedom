package middleware

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/8treenet/freedom"
	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"

	"github.com/kataras/golog"
)

var loggerPool sync.Pool

func init() {
	loggerPool = sync.Pool{
		New: func() interface{} {
			return &freedomLogger{}
		},
	}
}

// NewRuntimeLogger .
func NewRuntimeLogger(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		freelog := newFreedomLogger(ctx, traceIDName)
		ctx.Values().Set("logger_impl", freelog)
		ctx.Next()
		loggerPool.Put(freelog)
	}
}

func newFreedomLogger(ctx iris.Context, traceIDName string) *freedomLogger {
	logger := loggerPool.New().(*freedomLogger)
	logger.ctx = ctx
	logger.traceIDName = traceIDName
	return logger
}

type freedomLogger struct {
	ctx         iris.Context
	traceIDName string
}

// Print prints a log message without levels and colors.
func (l *freedomLogger) Print(v ...interface{}) {
	l.ctx.Application().Logger().Print(l.format(v...))
}

// Printf formats according to a format specifier and writes to `Printer#Output` without levels and colors.
func (l *freedomLogger) Printf(format string, args ...interface{}) {
	l.Print(fmt.Sprintf(format, args...))
}

// Println prints a log message without levels and colors.
// It adds a new line at the end, it overrides the `NewLine` option.
func (l *freedomLogger) Println(v ...interface{}) {
	l.ctx.Application().Logger().Println(l.format(v...))
}

// Log prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *freedomLogger) Log(level golog.Level, v ...interface{}) {
	l.ctx.Application().Logger().Log(level, l.format(v...))
}

// Logf prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *freedomLogger) Logf(level golog.Level, format string, args ...interface{}) {
	l.Log(level, fmt.Sprintf(format, args...))
}

// Fatal `os.Exit(1)` exit no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *freedomLogger) Fatal(v ...interface{}) {
	fileLine := fileLine()
	l.ctx.Application().Logger().Fatal(l.format(v...), " ", fileLine)
}

// Fatalf will `os.Exit(1)` no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *freedomLogger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg)
}

// Error will print only when freedomLogger's Level is error, warn, info or debug.
func (l *freedomLogger) Error(v ...interface{}) {
	fileLine := fileLine()
	l.ctx.Application().Logger().Error(l.format(v...), " ", fileLine)
}

// Errorf will print only when freedomLogger's Level is error, warn, info or debug.
func (l *freedomLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
}

// Warn will print when freedomLogger's Level is warn, info or debug.
func (l *freedomLogger) Warn(v ...interface{}) {
	l.ctx.Application().Logger().Warn(l.format(v...))
}

// Warnf will print when freedomLogger's Level is warn, info or debug.
func (l *freedomLogger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Warn(msg)
}

// Info will print when freedomLogger's Level is info or debug.
func (l *freedomLogger) Info(v ...interface{}) {
	l.ctx.Application().Logger().Info(l.format(v...))
}

// Infof will print when freedomLogger's Level is info or debug.
func (l *freedomLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg)
}

// Debug will print when freedomLogger's Level is debug.
func (l *freedomLogger) Debug(v ...interface{}) {
	l.ctx.Application().Logger().Debug(l.format(v...))
}

// Debugf will print when freedomLogger's Level is debug.
func (l *freedomLogger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg)
}

// format
func (l *freedomLogger) format(v ...interface{}) string {
	var list []string
	traceID := freedom.PickRuntime(l.ctx).Bus().Get(l.traceIDName)

	if traceID != "" {
		traceIDStr := fmt.Sprint(traceID)
		if traceIDStr != "" && l.traceIDName != "" {
			list = append(list, l.traceIDName+":"+traceIDStr)
		}
	}

	for _, i := range v {
		list = append(list, fmt.Sprint(i))
	}

	return strings.Join(list, " ")
}

func fileLine() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s %d", file, line)
}
