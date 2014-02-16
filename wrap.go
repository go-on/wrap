package wrap

import "net/http"

// Wrapper can wrap a http.Handler with another one
type Wrapper interface {
	// Wrap wraps an inner http.Handler with a new http.Handler that
	// is returned. The inner handler might be used in the scope of a
	// returned http.HandlerFunc.
	Wrap(inner http.Handler) (outer http.Handler)
}

// noop is a http.Handler that does nothing
var noop = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

// New returns a http.Handler that runs a stack of the given wrappers.
// When the handler serves the request the first wrapper
// serves the request and may let the second wrapper (its "inner" wrapper) serve.
// The second wrapper may let the third wrapper serve and so on.
// The last wrapper has as "inner" wrapper the not exported noop handler that does nothing.
func New(wrapper ...Wrapper) (h http.Handler) {
	h = noop
	for i := len(wrapper) - 1; i >= 0; i-- {
		h = wrapper[i].Wrap(h)
	}
	return
}
