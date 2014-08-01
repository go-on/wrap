package wrap

import (
	"net/http"
)

var (
	asHandler         = "http.Handler"
	asHandlerFunc     = "http.HandlerFunc"
	asNextHandler     = "NextHandler"
	asNextHandlerFunc = "NextHandlerFunc"
)

// Handler returns a Wrapper for a http.Handler.
// The returned Wrapper simply runs the given handler and ignores the
// next handler in the stack.
func Handler(h http.Handler) Wrapper {
	var nf NextHandlerFunc

	if DEBUG {
		nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
			(&debug{Object: h, Role: asHandler, Handler: h}).ServeHTTP(rw, req)
		}
		return nf
	}

	nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) { h.ServeHTTP(rw, req) }
	return nf
}

// HandlerFunc is like Handler but for a function with the type signature of http.HandlerFunc
func HandlerFunc(fn func(http.ResponseWriter, *http.Request)) Wrapper {
	var nf NextHandlerFunc

	if DEBUG {
		nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
			(&debug{Object: fn, Role: asHandlerFunc, Handler: http.HandlerFunc(fn)}).ServeHTTP(rw, req)
		}
		return nf
	}

	nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) { fn(rw, req) }
	return nf
}

// NextHandler returns a Wrapper for an interface with a ServeHTTPNext method
func NextHandler(sh interface {
	ServeHTTPNext(next http.Handler, rw http.ResponseWriter, req *http.Request)
}) Wrapper {
	var nf NextHandlerFunc

	if DEBUG {
		nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) {
			var f http.HandlerFunc
			f = func(rw http.ResponseWriter, req *http.Request) { sh.ServeHTTPNext(next, rw, req) }
			(&debug{Object: sh, Role: asNextHandler, Handler: f}).ServeHTTP(rw, req)
		}
		return nf
	}

	nf = func(next http.Handler, rw http.ResponseWriter, req *http.Request) { sh.ServeHTTPNext(next, rw, req) }
	return nf
}

// NextHandlerFunc is a Wrapper that is a function handling the request with the aid of the given handler
type NextHandlerFunc func(next http.Handler, rw http.ResponseWriter, req *http.Request)

// Wrap makes the ServeHandlerFunc fullfill the Wrapper interface by calling itself.
func (f NextHandlerFunc) Wrap(next http.Handler) http.Handler {
	var fn http.HandlerFunc

	if DEBUG {
		fn = func(rw http.ResponseWriter, req *http.Request) { f(next, rw, req) }
		return (&debug{Object: f, Role: asNextHandlerFunc, Handler: fn})
	}

	fn = func(rw http.ResponseWriter, req *http.Request) { f(next, rw, req) }
	return fn
}
