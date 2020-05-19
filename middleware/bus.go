package middleware

import (
	"strings"

	"github.com/8treenet/freedom"
)

// NewLimiter .
func NewLimiter() func(freedom.Runtime) {
	return func(run freedom.Runtime) {
		bus := run.Bus()
		for key := range bus.Header {
			if strings.Index(key, "x-") == 0 || strings.Index(key, "X-") == 0 {
				continue
			}
			bus.Del(key)
		}
	}
}
