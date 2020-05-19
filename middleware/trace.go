package middleware

import (
	"math/rand"
	"strings"
	"time"

	"github.com/8treenet/freedom"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/v12/context"
)

// NewTrace .
func NewTrace(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		bus := freedom.PickRuntime(ctx).Bus()
		traceID := bus.Get(traceIDName)
		for {
			if traceID != "" {
				break
			}
			uuidv1, e := uuid.NewV1()
			if e != nil {
				break
			}
			traceID = strings.ReplaceAll(uuidv1.String(), "-", "")
			bus.Add(traceIDName, traceID)
			break
		}
		ctx.Next()
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
