package general

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/kataras/iris/context"
)

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

// newRecover
func newRecover() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				rt := ctx.Values().Get(runtimeKey).(*appRuntime)
				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from path('%s')\n", ctx.Path())
				logMessage += fmt.Sprintf("Panic: %s\n", err)
				logMessage += fmt.Sprintf("At stack: %s\n", string(panicTrace()))
				rt.Logger().Warn(logMessage)

				ctx.StatusCode(500)
				ctx.StopExecution()
			}
		}()
		ctx.Next()
	}
}
