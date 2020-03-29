package requests

import "net/http"

type Middleware interface {
	Next()
	Stop()
	GetRequest() *http.Request
	GetRespone() (*http.Response, error)
	getStop() bool
}

var middlewares []Handler

type Handler func(Middleware)

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
		if request.getStop() {
			return
		}
	}
}
