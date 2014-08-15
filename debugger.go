package wrap

import (
	"io"
	"log"
	"net/http"
	"os"
)

var (
	asHandler         = "http.Handler"
	asHandlerFunc     = "http.HandlerFunc"
	asNextHandler     = "NextHandler"
	asNextHandlerFunc = "NextHandlerFunc"
	asWrapper         = "Wrapper"
)

type logDebugger struct {
	*log.Logger
}

func (l *logDebugger) Debug(req *http.Request, obj interface{}, role string) {
	l.Printf("%s %s %T as %s", req.Method, req.URL.Path, obj, role)
}

// NewLogDebugger sets the DEBUGGER  to a logger that logs to the given io.Writer.
// Flag is a flag from the log standard library that is passed to log.New
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
// It defaults to a logging debugger that logs to os.Stdout
var DEBUGGER = Debugger(&logDebugger{log.New(os.Stdout, "[go-on/wrap debugger]", log.LstdFlags)})

// DEBUG indicates if any stack should be debugged. Set it before any call to New.
var DEBUG = false

// SetDebug provides a way to set DEBUG=true in a var declaration, like
//
//   var _ = wrap.SetDebug()
//
// This is an easy way to ensure DEBUG is set to true before the init functions run
func SetDebug() bool {
	DEBUG = true
	return DEBUG
}

// debug is an internal type
type debug struct {
	Object interface{}
	Role   string
	http.Handler
}

func (d *debug) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	DEBUGGER.Debug(req, d.Object, d.Role)
	d.Handler.ServeHTTP(rw, req)
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
