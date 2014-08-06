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
//
// A Contexter is expected to be implemented in tandem with the creation of the
// outer application wide middelware stack. So the user of middleware with implement
// the Contexter even if it is used by middlewares to store middleware specific context
// data. There must only be one Contexter per application stack and it should be injected
// into the outer most stack at the beginning.
//
// Injecting a Contexter into the middleware stack means adding a Wrapper (may be the Contexter itself)
// that does pass the Contexter wrapping the original ResponseWriter as ResponseWriter to the ServeHTTP
// method of the next http.Handler in the stack.
//
// Doing so at the very beginning has the effect that all following http.Handler in the stack receive a
// ResponseWriter that implements Contexter. That allows all of them to set and get context data freely.
// The only condition is that the type of the context data must be supported by the Contexter.
// That is a hidden dependency (some middleware depends on the Contexter to support its context data type)
// and t should be covered by unit tests. However it gives great freedom when organizing and ordering middleware
// as the middleware in between does not have to know about a Contexter at all (if it does not need context) and
// especially not about a certain context type being supported if does not care about that type.
// It simply passes the ResponseWriter that happens to have context down the middleware stack chain.
//
// When implementing a Contexter it is important to support the *http.ResponseWriter type by the Context()
// method, because it allows middleware to retrieve the original response writer and type assert it to
// some interfaces of the standard http library, e.g. http.Flusher. To make this easier there are some
// helper functions like Flush(), CloseNotify() and Hijack() - all making use of the more general ReclaimResponseWriter().
//
// Template for the implementation of a Contexter
//
//     type context struct {
//       http.ResponseWriter
//       err error
//     }
//
//     // make sure to fulfill the Contexter interface
//     var _ wrap.Contexter = &context{}

//     // Context receives a pointer to a value that is already stored inside the context.
//     // Values are distiguished by their type.
//     // Context sets the value of the given pointer to the value of the same type
//     // that is stored inside of the context.
//     // A pointer type that is not supported results in a panic.
//     // Context returns if ctxPtr will be nil after return
//     // Context must support *http.ResponseWriter
//     func (c *context) Context(ctxPtr interface{}) (found bool) {
//       found = true
//       switch ty := ctxPtr.(type) {
//       case *http.ResponseWriter:
//         *ty = c.ResponseWriter
//       case *error:
//         if c.err == nil {
//           return false
//         }
//         *ty = c.err
//       default:
//         panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
//       }
//       return
//     }
//
//     // SetContext receives a pointer to a value that will be stored inside the context.
//     // Values are distiguished by their type, that means that SetContext replaces
//     // and stored value of the same type.
//     // A pointer type that is not supported results in a panic.
//     func (c *context) SetContext(ctxPtr interface{}) {
//       switch ty := ctxPtr.(type) {
//       case *error:
//         c.err = *ty
//       default:
//         panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
//       }
//     }
//
//     // Wrap implements the wrap.Wrapper interface.
//     //
//     // When the request is served, the response writer is wrapped by a
//     // new *context which is passed to the next handlers ServeHTTP method.
//     func (c context) Wrap(next http.Handler) http.Handler {
//       var f http.HandlerFunc
//       f = func(rw http.ResponseWriter, req *http.Request) {
//         next.ServeHTTP(&context{ResponseWriter: rw}, req)
//       }
//       return f
//     }
//
// While it looks like much work it is in fact not, since this is only written once per application and
// the effort to support a new context data type is low.
//
// Lets have a look at the effort to support a new context type (here: error):
//
// 1. Add a new field to your context (1 LOC)
//
//     // type context struct {
//     //  ...
//         err error
//     // }
//
// 2. Add a new case to retrieve the context data (5 LOC)
//
//     // func (c *context) Context(ctxPtr interface{}) (found bool) {
//     //  ...
//     //  switch ty := ctxPtr.(type) {
//     //  ...
//         case *error:
//           if c.err == nil {
//             return false
//           }
//           *ty = c.err
//     //  }
//     //  ...
//     // }
//
// 3. Add a new case to set the context data (3 LOC)
//
//     // func (c *context) SetContext(ctxPtr interface{}) {
//     //  ...
//     //  switch ty := ctxPtr.(type) {
//     //  ...
//         case *error:
//           c.err = *ty
//         }
//     //  ...
//     // }
//
// So 9 simple LOC to support a new context type is not that bad.
//
// A middleware using the error context could look like this:
//
//     var StopIfError wrap.NextHandlerFunc = func (next http.Handler, rw http.ResponseWriter, req *http.Request) {
//    	 var err error
// 	     if rw.(wrap.Contexter).Context(&err) {
//      		rw.WriteHeader(500)
//
//      		if os.Getenv("DEVELOPMENT") != "" {
// 		     	  fmt.Fprintf(rw, "Error:\n\n%s", err.Error())
// 		     	  return
// 		      }
//
//          fmt.Fprint(rw, "Internal server error")
// 		      return
// 	     }
//       next.ServeHTTP(rw,req)
//     }
//
//
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
