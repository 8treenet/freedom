package middleware

import (
	"fmt"
	"path"
	"runtime"
	"sync"

	"github.com/8treenet/freedom"

	"github.com/kataras/golog"
)

const (
	fatalLevel = 1
	// errorLevel will print only errors.
	errorLevel = 2
	// warnLevel will print errors and warnings.
	warnLevel = 3
	// infoLevel will print errors, warnings and infos.
	infoLevel = 4
	// debugLevel will print on any level, fatals, errors, warnings, infos and debug logs.
	debugLevel = 5
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
	logger.callerWithLevel = 0
	return logger
}

// Logger The implementation of Logger..
type Logger struct {
	traceID         string
	traceName       string
	callerWithLevel int
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
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(fatalLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Fatal(v...)
}

// Fatalf will `os.Exit(1)` no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(fatalLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Fatalf(format, v...)
}

// Error will print only when freedomLogger's Level is error, warn, info or debug.
func (l *Logger) Error(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(errorLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Error(v...)
}

// Errorf will print only when freedomLogger's Level is error, warn, info or debug.
func (l *Logger) Errorf(format string, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(errorLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Errorf(format, v...)
}

// Warn will print when freedomLogger's Level is warn, info or debug.
func (l *Logger) Warn(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(warnLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Warn(v...)
}

// Warnf will print when freedomLogger's Level is warn, info or debug.
func (l *Logger) Warnf(format string, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(warnLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Warnf(format, v...)
}

// Info will print when freedomLogger's Level is info or debug.
func (l *Logger) Info(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(infoLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Info(v...)
}

// Infof will print when freedomLogger's Level is info or debug.
func (l *Logger) Infof(format string, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(infoLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Infof(format, v...)
}

// Debug will print when freedomLogger's Level is debug.
func (l *Logger) Debug(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(debugLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Debug(v...)
}

// Debugf will print when freedomLogger's Level is debug.
func (l *Logger) Debugf(format string, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	if callerField := l.getCallerField(debugLevel); callerField != nil {
		v = append(v, callerField)
	}
	freedom.Logger().Debugf(format, v...)
}

// SetCallerLevel .
func (l *Logger) SetCallerLevel(level golog.Level) {
	index := 1
	switch level {
	case golog.DebugLevel:
		index = debugLevel
	case golog.InfoLevel:
		index = infoLevel
	case golog.WarnLevel:
		index = warnLevel
	case golog.ErrorLevel:
		index = errorLevel
	case golog.FatalLevel:
		index = fatalLevel
	}
	l.callerWithLevel |= (1 << index)
}

// traceField
func (l *Logger) traceField() golog.Fields {
	return golog.Fields{l.traceName: l.traceID}
}

// callerField
func (l *Logger) getCallerField(level int) golog.Fields {
	if (l.callerWithLevel >> level & 1) == 0 {
		return nil
	}

	_, file, line, _ := runtime.Caller(2)
	return golog.Fields{"caller": fmt.Sprintf("%s:%d", path.Base(file), line)}
}
