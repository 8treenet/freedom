package middleware

import (
	"github.com/kataras/iris/v12/context"
)

// The skipperFunc signature, used to serve the main request without logs.
// See `Configuration` too.
type skipperFunc func(ctx context.Context) bool

// RequestLoggerConfig contains the options for the logger middleware
// can be optionally be passed to the `New`.
type RequestLoggerConfig struct {
	IP                 bool
	Query              bool
	MessageContextKeys []string
	MessageHeaderKeys  []string
	RequestRawBody     bool
	Title              string
	// Status displays status code (bool).
	//
	// Defaults to true.
	// Status bool
	// IP displays request's remote address (bool).
	//
	// Defaults to true.

	// Method displays the http method (bool).
	//
	// Defaults to true.
	//Method bool
	// Path displays the request path (bool).
	//
	// Defaults to true.
	//Path bool

	// Query will append the URL Query to the Path.
	// Path should be true too.
	//
	// Defaults to true.

	// Columns will display the logs as a formatted columns-rows text (bool).
	// If custom `LogFunc` has been provided then this field is useless and users should
	// use the `Columinize` function of the logger to get the output result as columns.
	//
	// Defaults to false.
	//Columns bool

	// MessageContextKeys if not empty,
	// the middleware will try to fetch
	// the contents with `ctx.Values().Get(MessageContextKey)`
	// and if available then these contents will be
	// appended as part of the logs (with `%v`, in order to be able to set a struct too),
	// if Columns field was set to true then
	// a new column will be added named 'Message'.
	//
	// Defaults to empty.

	// MessageHeaderKeys if not empty,
	// the middleware will try to fetch
	// the contents with `ctx.Values().Get(MessageHeaderKey)`
	// and if available then these contents will be
	// appended as part of the logs (with `%v`, in order to be able to set a struct too),
	// if Columns field was set to true then
	// a new column will be added named 'HeaderMessage'.
	//
	// Defaults to empty.
	traceName string
}

// DefaultConfig returns a default config
// that have all boolean fields to true except `Columns`,
// all strings are empty,
// LogFunc and Skippers to nil as well.
func DefaultConfig() *RequestLoggerConfig {
	return &RequestLoggerConfig{
		IP:                 false,
		Query:              true,
		RequestRawBody:     true,
		MessageContextKeys: []string{"response"},
		Title:              "[access]",
	}
}

// AddSkipper adds a skipper to the configuration.
// func (c *LoggerConfig) AddSkipper(sk skipperFunc) {
// 	c.Skippers = append(c.Skippers, sk)
// 	c.buildSkipper()
// }

// func (c *LoggerConfig) buildSkipper() {
// 	if len(c.Skippers) == 0 {
// 		return
// 	}
// 	skippersLocked := c.Skippers[0:]
// 	c.skip = func(ctx context.Context) bool {
// 		for _, s := range skippersLocked {
// 			if s(ctx) {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }
