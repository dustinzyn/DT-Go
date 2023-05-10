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

	// IP displays request's remote address (bool).
	// Defaults to true.
	IP bool

	// Method displays the http method (bool).
	// Defaults to true.
	Method bool

	// Status displays status code (bool).
	// Defaults to true.
	Status bool

	// Path displays the request path (bool).
	//
	// Defaults to true.
	Path bool

	// Query will append the URL Query to the Path.
	// Path should be true too.
	// Defaults to true.
	Query bool

	// MessageContextKeys if not empty,
	// the middleware will try to fetch
	// the contents with `ctx.Values().Get(MessageContextKey)`
	// and if available then these contents will be
	// appended as part of the logs (with `%v`, in order to be able to set a struct too),
	// if Columns field was set to true then
	// a new column will be added named 'Message'.
	//
	// Defaults to empty.
	MessageContextKeys []string

	// MessageHeaderKeys if not empty,
	// the middleware will try to fetch
	// the contents with `ctx.Values().Get(MessageHeaderKey)`
	// and if available then these contents will be
	// appended as part of the logs (with `%v`, in order to be able to set a struct too),
	// if Columns field was set to true then
	// a new column will be added named 'HeaderMessage'.
	//
	// Defaults to empty.
	MessageHeaderKeys []string

	// RequestRawBody displays request body (bool)
	// Defaults to true
	RequestRawBody       bool

	// RequestRawBodyMaxLen displays request body'length
	// Defaults to 512
	RequestRawBodyMaxLen int64

	// Title displays title, Defaults to [ACCESS]
	// When x-request-id is configured, the value of x-request-id in the request header is displayed
	Title                string

	traceName            string

	// Columns will display the logs as a formatted columns-rows text (bool).
	// If custom `LogFunc` has been provided then this field is useless and users should
	// use the `Columinize` function of the logger to get the output result as columns.
	//
	// Defaults to false.
	Columns bool
}

// DefaultLoggerConfig returns a default config
// that have all boolean fields to true except `Columns`,
// all strings are empty,
// LogFunc and Skippers to nil as well.
func DefaultLoggerConfig() *RequestLoggerConfig {
	return &RequestLoggerConfig{
		IP:                   true,
		Query:                true,
		RequestRawBody:       true,
		RequestRawBodyMaxLen: 512,
		// MessageContextKeys:   []string{"response"},
		Title:                "[ACCESS]",
	}
}
