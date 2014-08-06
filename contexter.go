package wrap

import (
	"bufio"
	"net"
	"net/http"
)

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
	// It should always support *http.ResponseWriter and set it to the underlying
	// http.ResponseWriter in order to allow middleware to type assert to Flusher et. al.
	Context(ctxPtr interface{}) (found bool)

	// SetContext saves the given context pointer via type switch
	SetContext(ctxPtr interface{})
}

// ReclaimResponseWriter is a helper that expects the given ResponseWriter to either be
// the original ResponseWriter or a Contexter which supports getting the original
// response writer via *http.ResponseWriter. In either case it returns the underlying
// response writer
func ReclaimResponseWriter(rw http.ResponseWriter) http.ResponseWriter {
	ctx, ok := rw.(Contexter)
	if !ok {
		return rw
	}
	var w http.ResponseWriter
	ctx.Context(&w)
	return w
}

// Flush is a helper that flushes the buffer in the  underlying response writer if it is a http.Flusher.
// The http.ResponseWriter might also be a Contexter if it allows the retrieval of the underlying
// ResponseWriter. Ok returns if the underlying ResponseWriter was a http.Flusher
func Flush(rw http.ResponseWriter) (ok bool) {
	w := ReclaimResponseWriter(rw)
	if fl, is := w.(http.Flusher); is {
		fl.Flush()
		return true
	}
	return false
}

// CloseNotify is the same for http.CloseNotifier as Flush is for http.Flusher
// ok tells if it was a CloseNotifier
func CloseNotify(rw http.ResponseWriter) (ch <-chan bool, ok bool) {
	w := ReclaimResponseWriter(rw)
	if cl, is := w.(http.CloseNotifier); is {
		ch = cl.CloseNotify()
		ok = true
		return
	}
	return
}

// Hijack is the same for http.Hijacker as Flush is for http.Flusher
// ok tells if it was a Hijacker
func Hijack(rw http.ResponseWriter) (c net.Conn, brw *bufio.ReadWriter, err error, ok bool) {
	w := ReclaimResponseWriter(rw)
	if hj, is := w.(http.Hijacker); is {
		c, brw, err = hj.Hijack()
		ok = true
		return
	}
	return
}
