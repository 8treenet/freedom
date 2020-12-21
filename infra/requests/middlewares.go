package requests

import (
	"context"
	"net/http"
)

// Middleware .
type Middleware interface {
	Next()
	Stop(...error)
	GetRequest() *http.Request
	GetRespone() *Response
	GetResponeBody() []byte
	IsStopped() bool
	//Is HTTP/2 Cleartext Request
	IsH2C() bool
	Context() context.Context
	WithContextFromMiddleware(context.Context)
	EnableTraceFromMiddleware()
	SetClientFromMiddleware(Client)
}

var middlewares []Handler

// Handler .
type Handler func(Middleware)

// InstallMiddleware .
func InstallMiddleware(handle ...Handler) {
	middlewares = append(middlewares, handle...)
}
