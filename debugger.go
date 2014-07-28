package wrap

import (
	"io"
	"log"
	"net/http"
	"os"
)

type logDebugger struct {
	*log.Logger
}

func (l *logDebugger) Debug(req *http.Request, obj interface{}, role string) {
	l.Printf("%s %s %T as %s", req.Method, req.URL.Path, obj, role)
}

func NewLogDebugger(out io.Writer, flag int) {
	DEBUGGER = &logDebugger{log.New(out, "[go-on/wrap debugger]", flag)}
}

// Debugger has a Debug method to debug middleware stacks
type Debugger interface {
	// Debug receives the current request, the object that wraps and the role in which
	// the object acts. Role is a string representing the interface in which obj
	// is used, e.g. "Wrapper", "http.Handler" and so on
	Debug(req *http.Request, obj interface{}, role string)
}

// DEBUGGER is the Debugger used for debugging middleware stacks.
// It defaults to a logging debugger
var DEBUGGER Debugger = &logDebugger{log.New(os.Stdout, "[go-on/wrap debugger]", log.LstdFlags)}

// DEBUG indicates if any stack should be debugged. Set it before any call to New.
var DEBUG = false

type debug struct {
	obj  interface{}
	role string
	h    http.Handler
}

var asHandler = "http.Handler"
var asHandlerFunc = "http.HandlerFunc"
var asWrapper = "Wrapper"

func (d *debug) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	DEBUGGER.Debug(req, d.obj, d.role)
	d.h.ServeHTTP(rw, req)
}

// _debug is like New() but wraps each http.Handler with a debug struct that calls DEBUGGER.Debug before
// running the actual http.Handler.
func _debug(wrapper ...Wrapper) (h http.Handler) {
	h = NoOp
	for i := len(wrapper) - 1; i >= 0; i-- {
		h = &debug{wrapper[i], asWrapper, wrapper[i].Wrap(h)}
	}
	return
}
