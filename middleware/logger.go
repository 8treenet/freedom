package middleware

import (
	"fmt"
	"path"
	"runtime"
	"sync"

	"github.com/8treenet/freedom"

	"github.com/kataras/golog"
)

var loggerPool sync.Pool

func init() {
	loggerPool = sync.Pool{
		New: func() interface{} {
			return &Logger{}
		},
	}
}

// NewLogger Create a Loger.
// The name and value of the incoming Trace.
// Requests should have their own Loger.
func NewLogger(traceName, traceID string) *Logger {
	logger := loggerPool.New().(*Logger)
	logger.traceID = traceID
	logger.traceName = traceName
	return logger
}

// Logger The implementation of Logger..
type Logger struct {
	traceID   string
	traceName string
}

// Print prints a log message without levels and colors.
func (l *Logger) Print(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	freedom.Logger().Print(v...)
}

// Printf formats according to a format specifier and writes to `Printer#Output` without levels and colors.
func (l *Logger) Printf(format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	freedom.Logger().Printf(format, args...)
}

// Println prints a log message without levels and colors.
// It adds a new line at the end, it overrides the `NewLine` option.
func (l *Logger) Println(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	freedom.Logger().Println(v...)
}

// Log prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *Logger) Log(level golog.Level, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	freedom.Logger().Log(level, v...)
}

// Logf prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *Logger) Logf(level golog.Level, format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	freedom.Logger().Logf(level, format, args...)
}

// Fatal `os.Exit(1)` exit no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *Logger) Fatal(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	freedom.Logger().Fatal(v...)
}

// Fatalf will `os.Exit(1)` no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	freedom.Logger().Fatalf(format, args...)
}

// Error will print only when freedomLogger's Level is error, warn, info or debug.
func (l *Logger) Error(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	freedom.Logger().Error(v...)
}

// Errorf will print only when freedomLogger's Level is error, warn, info or debug.
func (l *Logger) Errorf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	freedom.Logger().Errorf(format, args...)
}

// Warn will print when freedomLogger's Level is warn, info or debug.
func (l *Logger) Warn(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	freedom.Logger().Warn(v...)
}

// Warnf will print when freedomLogger's Level is warn, info or debug.
func (l *Logger) Warnf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	freedom.Logger().Warnf(format, args...)
}

// Info will print when freedomLogger's Level is info or debug.
func (l *Logger) Info(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	freedom.Logger().Info(v...)
}

// Infof will print when freedomLogger's Level is info or debug.
func (l *Logger) Infof(format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	freedom.Logger().Infof(format, args...)
}

// Debug will print when freedomLogger's Level is debug.
func (l *Logger) Debug(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	freedom.Logger().Debug(v...)
}

// Debugf will print when freedomLogger's Level is debug.
func (l *Logger) Debugf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	freedom.Logger().Debugf(format, args...)
}

// traceField
func (l *Logger) traceField() golog.Fields {
	return golog.Fields{l.traceName: l.traceID}
}

// callerField
func (l *Logger) callerField() golog.Fields {
	_, file, line, _ := runtime.Caller(2)
	return golog.Fields{"caller": fmt.Sprintf("%s:%d", path.Base(file), line)}
}
