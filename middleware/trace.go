package middleware

import (
	"math/rand"
	"strings"
	"time"

	"github.com/8treenet/freedom"

	"github.com/8treenet/iris/v12/context"
	"github.com/google/uuid"
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
	uuidv1 := uuid.New()
	return strings.ReplaceAll(uuidv1.String(), "-", ""), nil
}
