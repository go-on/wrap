package wrap

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

// userIP represents the IP address of a http.Request
type userIP net.IP

// context implements Contexter, providing a userIP and a error
type context struct {
	http.ResponseWriter
	userIP userIP
	err    error
}

// make sure to fulfill the Contexter interface
var _ Contexter = &context{}

// context is an implementation for the Contexter interface.
//
// It receives a pointer to a value that is already stored inside the context.
// Values are distiguished by their type.
// Context sets the value of the given pointer to the value of the same type
// that is stored inside of the context.
// A pointer type that is not supported results in a panic.
// *http.ResponseWriter should always be supported in order to get the underlying ResponseWriter
// Context returns if the pointer is no nil pointer when returning.
func (c *context) Context(ctxPtr interface{}) (found bool) {
	found = true // save work
	switch ty := ctxPtr.(type) {
	case *http.ResponseWriter:
		*ty = c.ResponseWriter
	case *userIP:
		if c.userIP == nil {
			return false
		}
		*ty = c.userIP
	case *error:
		if c.err == nil {
			return false
		}
		*ty = c.err
	default:
		panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
	}
	return
}

// SetContext is an implementation for the Contexter interface.
//
// It receives a pointer to a value that will be stored inside the context.
// Values are distiguished by their type, that means that SetContext replaces
// and stored value of the same type.
// A pointer type that is not supported results in a panic.
func (c *context) SetContext(ctxPtr interface{}) {
	switch ty := ctxPtr.(type) {
	case *http.ResponseWriter:
		c.ResponseWriter = *ty
	case *userIP:
		c.userIP = *ty
	case *error:
		c.err = *ty
	default:
		panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
	}
}

// Wrap implements the wrap.Wrapper interface.
//
// When the request is served, the response writer is wrapped by a
// new *context which is passed to the next handlers ServeHTTP method.
func (c context) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(&context{ResponseWriter: rw}, req)
	}
	return f
}

// setUserIP is a middleware that requires a context supporting the userIP and the error type
type setUserIP struct{}

func (setUserIP) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		ip, err := ipfromRequest(req)
		if err != nil {
			rw.(Contexter).SetContext(&err)
		} else {
			uIP := userIP(ip)
			rw.(Contexter).SetContext(&uIP)
		}
		next.ServeHTTP(rw, req)
	}
	return f
}

// ipfromRequest extracts the user IP address from req, if present.
// taken from http://blog.golang.org/context/userip/userip.go (FromRequest)
func ipfromRequest(req *http.Request) (net.IP, error) {
	s := strings.SplitN(req.RemoteAddr, ":", 2)
	userIP := net.ParseIP(s[0])
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return userIP, nil
}

// handleError is a middleware for handling errors.
// it requires a context supporting the error type.
type handleError struct{}

// Wrap implements the wrap.Wrapper interface and checks for an error context.
// If it finds one, the status 500 is set and the error is written to the response writer.
// If no error is inside the context, the next handler is called.
func (handleError) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		var err error
		rw.(Contexter).Context(&err)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		next.ServeHTTP(rw, req)
	}
	return f
}

// app gets the userIP and writes it to the responsewriter. it requires  a context supporting the userIP
type app struct{}

// Wrap implements the wrap.Wrapper interface and writes a userIP from a context to the response writer
func (app) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		var uIP userIP
		rw.(Contexter).Context(&uIP)
		fmt.Fprint(rw, net.IP(uIP).String())
	}
	return f
}

func ExampleContext() {
	h := New(
		context{}, // context must always be the first one
		setUserIP{},
		handleError{},
		app{},
	)
	rec := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "garbage"
	h.ServeHTTP(rec, r)
	fmt.Println(rec.Body.String())

	rec.Body.Reset()
	r.RemoteAddr = "127.0.0.1:45643"
	h.ServeHTTP(rec, r)
	fmt.Println(rec.Body.String())

	// Output:
	// userip: "garbage" is not IP:port
	// 127.0.0.1
	//
}
