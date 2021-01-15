package freedom

import (
	"github.com/8treenet/freedom/internal"
)

// WorkerFromCtx extracts Worker from a Context.
func WorkerFromCtx(ctx Context) Worker {
	if result, ok := ctx.Values().Get(internal.WorkerKey).(Worker); ok {
		return result
	}
	return nil
}

// TODO(coco): deprecated
// ToWorker proxy a call to WorkerFromCtx.
func ToWorker(ctx Context) Worker {
	return WorkerFromCtx(ctx)
}
