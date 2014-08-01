package wrap

import "net/http"

// Context is a http.ResponseWriter that can set and get contexts. It allows
// plain http.Handlers to share per request context data without global state.
//
// Implementations of Context should be structs wrapping a ResponseWriter.
type Context interface {

	// since implementations of Context should be a wrapper around a responsewriter
	// they must implement the http.ResponseWriter interface.
	http.ResponseWriter

	// Context lets the given pointer point to the saved context of the same type
	Context(ctxPtr interface{})

	// SetContext saves the given context pointer via type switch
	SetContext(ctxPtr interface{})
}
