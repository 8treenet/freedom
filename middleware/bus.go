package middleware

import (
	"strings"

	"github.com/8treenet/freedom"
)

// NewBusFilter .
func NewBusFilter() func(freedom.Worker) {
	return func(run freedom.Worker) {
		bus := run.Bus()
		for key := range bus.Header {
			if strings.Index(key, "x-") == 0 || strings.Index(key, "X-") == 0 {
				continue
			}
			bus.Del(key)
		}
	}
}
