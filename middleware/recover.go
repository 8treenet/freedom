package middleware

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/kataras/golog"
	"github.com/kataras/iris/v12/context"

	"github.com/8treenet/freedom"
)

// NewRecover
func NewRecover() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				rt := freedom.ToWorker(ctx)
				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from path('%s')\n", ctx.Path())
				logMessage += fmt.Sprintf("Panic: %s\n", err)
				logMessage += fmt.Sprintf("At stack: %s\n", string(panicTrace()))
				rt.Logger().Log(golog.ErrorLevel, logMessage)

				ctx.StatusCode(500)
				ctx.StopExecution()
			}
		}()
		ctx.Next()
	}
}

func panicTrace() []byte {
	def := []byte("An unknown error")
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, 2<<10) //2KB
	length := runtime.Stack(stack, false)
	start := bytes.Index(stack, s)
	if start == -1 || start > len(stack) || length > len(stack) {
		return def
	}
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	if start >= len(stack) {
		return def
	}
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end > len(stack) {
		return def
	}
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end > len(stack) {
		return def
	}
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}
