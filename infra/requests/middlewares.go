package requests

import "net/http"

// Middleware .
type Middleware interface {
	Next()
	Stop(...error)
	GetRequest() *http.Request
	GetRespone() (*http.Response, error)
	IsStopped() bool
}

var middlewares []Handler

// Handler .
type Handler func(Middleware)

// UseMiddleware .
func UseMiddleware(handle ...Handler) {
	middlewares = append(middlewares, handle...)
}

func handle(request Middleware) {
	if len(middlewares) == 0 {
		request.Next()
		return
	}
	for i := 0; i < len(middlewares); i++ {
		middlewares[i](request)
		if request.IsStopped() {
			return
		}
	}
}
