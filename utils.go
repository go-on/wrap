package wrap

import "net/http"

// WrapperFunc is a function that acts as Wrapper
type WrapperFunc func(http.Handler) http.Handler

// Wrap makes the WrapperFunc fullfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(in http.Handler) http.Handler { return wf(in) }

// Handler returns a Wrapper for a http.Handler.
// The returned Wrapper simply runs the given handler and ignores the
// inner handler in the stack.
func Handler(h http.Handler) Wrapper {
	return ServeHandlerFunc(
		func(inner http.Handler, rw http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(rw, req)
		},
	)
}

// HandlerFunc serves the same purpose as Handler but for a function of the type
// signature as http.HandlerFunc
func HandlerFunc(fn func(http.ResponseWriter, *http.Request)) Wrapper {
	return ServeHandlerFunc(
		func(inner http.Handler, rw http.ResponseWriter, req *http.Request) {
			fn(rw, req)
		},
	)
}

// ServeHandler can serve the given request with the aid of the given handler
type ServeHandler interface {
	// ServeHandler serves the given request with the aid of the given handler
	ServeHandle(inner http.Handler, rw http.ResponseWriter, req *http.Request)
}

// ServeWrapper returns a Wrapper for a ServeHandler
func ServeWrapper(wh ServeHandler) Wrapper {
	fn := func(inner http.Handler, rw http.ResponseWriter, req *http.Request) {
		wh.ServeHandle(inner, rw, req)
	}
	return ServeHandlerFunc(fn)
}

// ServeHandle creates a http.Handler by using the given ServeHandler
func ServeHandle(wh ServeHandler, inner http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			wh.ServeHandle(inner, rw, req)
		},
	)
}

// ServeHandleFunc serves the same purpose as ServeHandle but for a function of the type
// signature as ServeHandlerFunc
func ServeHandleFunc(fn func(http.Handler, http.ResponseWriter, *http.Request), inner http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fn(inner, rw, req)
		},
	)
}

// ServeHandlerFunc is a function that handles the given request with the aid of the given handler
// and is a Wrapper
type ServeHandlerFunc func(inner http.Handler, rw http.ResponseWriter, req *http.Request)

// Wrap makes the ServeHandlerFunc fullfill the Wrapper interface by calling itself.
func (f ServeHandlerFunc) Wrap(inner http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			f(inner, rw, req)
		},
	)
}
