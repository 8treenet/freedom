package requests

import (
	"context"
	"net/http"
)

// Middleware The interface definition of the middleware.
type Middleware interface {
	//Go on.
	Next()
	Stop(...error)
	// Returns the standard request object.
	GetRequest() *http.Request
	// Returns carry information for HTTP requests.
	GetRespone() *Response
	// Returns http package data.
	GetResponeBody() []byte
	IsStopped() bool
	// Is HTTP/2 Cleartext Request.
	IsH2C() bool
	Context() context.Context
	WithContextFromMiddleware(context.Context)
	// Turn on link tracking for HTTP connections.
	EnableTraceFromMiddleware()
	// Set up client.
	SetClientFromMiddleware(Client)
}

var middlewares []Handler

// Handler The definition of the middleware processor.
type Handler func(Middleware)

// InstallMiddleware Install the middleware processor.
func InstallMiddleware(handle ...Handler) {
	middlewares = append(middlewares, handle...)
}
