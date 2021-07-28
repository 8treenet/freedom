package middleware

import (
	"math/rand"
	"strings"
	"time"

	"github.com/8treenet/freedom"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/v12/context"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewTrace The default HTTP Trace.
func NewTrace(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		bus := freedom.ToWorker(ctx).Bus()
		traceID := bus.Get(traceIDName)
		for {
			if traceID != "" {
				break
			}

			var e error
			if traceID, e = GenerateTraceID(); e != nil {
				break
			}
			bus.Add(traceIDName, traceID)
			break
		}
		ctx.Next()
	}
}

// GenerateTraceID Build a unique ID.
func GenerateTraceID() (string, error) {
	uuidv1, e := uuid.NewV1()
	if e != nil {
		return "", e
	}
	return strings.ReplaceAll(uuidv1.String(), "-", ""), nil
}
