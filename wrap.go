package wrap

import "net/http"

// Wrapper can wrap a http.Handler with another one
type Wrapper interface {
	// Wrap wraps the next `http.Handler` of the stack and returns the previous `http.Handler`
	// If `next.ServeHTTP` is not called, the next `http.Handler` won't be used
	Wrap(next http.Handler) (previous http.Handler)
}

// NoOp is a http.Handler that does nothing
var NoOp = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

// New returns a http.Handler that wraps the given wrappers.
// When it serves the request the first given wrapper
// serves the request and may let the second wrapper (its "next" wrapper) serve.
// The second wrapper may let the third wrapper serve and so on.
// The last wrapper has as "next" wrapper the not exported NoOp handler that does nothing.
// If DEBUG is set, each http.Handler is wrapped with a Debug struct that calls DEBUGGER.Debug before
// running the actual http.Handler.
func New(wrapper ...Wrapper) (h http.Handler) {
	if DEBUG {
		return _debug(wrapper...)
	}
	h = NoOp
	for i := len(wrapper) - 1; i >= 0; i-- {
		h = wrapper[i].Wrap(h)
	}
	return
}

// WrapperFunc is an adapter for a function that acts as Wrapper
type WrapperFunc func(http.Handler) http.Handler

// Wrap makes the WrapperFunc fulfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(next http.Handler) http.Handler { return wf(next) }
