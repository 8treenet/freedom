package middleware

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/8treenet/freedom/general"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/context"
)

// NewTrace .
func NewTrace(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		bus := general.GetBus(ctx)
		traceID, ok := bus.Get(traceIDName)
		for {
			if ok || traceID != nil {
				fmt.Println("NewTrace2")
				break
			}
			uuidv1, e := uuid.NewV1()
			if e != nil {
				break
			}
			traceID = strings.ReplaceAll(uuidv1.String(), "-", "")
			break
		}

		ctx.Values().Set(traceIDName, traceID)
		bus.Add(traceIDName, traceID)
		ctx.Next()
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
