package wrap

import "net/http"

// Wrapper can wrap a http.Handler with another one
type Wrapper interface {
	// Wrap wraps the next `http.Handler` of the stack and returns a wrapping `http.Handler`.
	// If it does not call `next.ServeHTTP`, nobody will.
	Wrap(next http.Handler) http.Handler
}

// New returns the wrapping http.Handler that returned from calling
// the Wrap method on the first given wrapper that received the returned
// handler of returning from the second wrappers Wrap method and so on.
//
// The last wrapper begins the loop receiving the NoOp handler.
//
// When the ServeHTTP method of the returned handler is called each wrapping
// handler may call is next until the NoOp handler is run.
//
// Or some wrapper decides not to call next.ServeHTTP.
//
// If DEBUG is set, each handler is wrapped with a Debug struct that calls DEBUGGER.Debug before
// running the handler.
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

// NoOp is a http.Handler doing nothing
var NoOp = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
