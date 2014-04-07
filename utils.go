package wrap

import "net/http"

// WrapperFunc is a function that acts as Wrapper
type WrapperFunc func(http.Handler) http.Handler

// Wrap makes the WrapperFunc fullfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(next http.Handler) http.Handler { return wf(next) }

// Handler returns a Wrapper for a http.Handler.
// The returned Wrapper simply runs the given handler and ignores the
// next handler in the stack.
func Handler(h http.Handler) Wrapper {
	return ServeHandlerFunc(
		func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(rw, req)
		},
	)
}

// HandlerFunc is like Handler but for a function with the type signature of http.HandlerFunc
func HandlerFunc(fn func(http.ResponseWriter, *http.Request)) Wrapper {
	return ServeHandlerFunc(
		func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
			fn(rw, req)
		},
	)
}

// ServeHandler has a ServeHandle method
type ServeHandler interface {
	// ServeHandle serves the given request with the aid of the given handler
	ServeHandle(next http.Handler, rw http.ResponseWriter, req *http.Request)
}

// ServeWrapper returns a Wrapper for a ServeHandler
func ServeWrapper(sh ServeHandler) Wrapper {
	fn := func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
		sh.ServeHandle(next, rw, req)
	}
	return ServeHandlerFunc(fn)
}

// ServeHandle creates a http.Handler by using the given ServeHandler
func ServeHandle(sh ServeHandler, next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		sh.ServeHandle(next, rw, req)
	}
	return http.HandlerFunc(fn)
}

// ServeHandleFunc is like ServeHandle but for a function with the signature of ServeHandlerFunc
func ServeHandleFunc(fn func(http.Handler, http.ResponseWriter, *http.Request), next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fn(next, rw, req)
		},
	)
}

// ServeHandlerFunc is a Wrapper that is a function handling the request with the aid of the given handler
type ServeHandlerFunc func(next http.Handler, rw http.ResponseWriter, req *http.Request)

// Wrap makes the ServeHandlerFunc fullfill the Wrapper interface by calling itself.
func (f ServeHandlerFunc) Wrap(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		f(next, rw, req)
	}
	return http.HandlerFunc(fn)
}
