package wrap

import "net/http"

// Contexter is a http.ResponseWriter that can set and get contexts. It allows
// plain http.Handlers to share per request context data without global state.
//
// Implementations of Context should be structs wrapping a ResponseWriter.
type Contexter interface {

	// since implementations of Context should be a wrapper around a responsewriter
	// they must implement the http.ResponseWriter interface.
	http.ResponseWriter

	// Context lets the given pointer point to the saved context of the same type
	// Returns if it has found something
	Context(ctxPtr interface{}) (found bool)

	// SetContext saves the given context pointer via type switch
	SetContext(ctxPtr interface{})
}
