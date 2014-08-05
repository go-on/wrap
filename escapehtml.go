package wrap

import (
	"net/http"
	"unicode/utf8"
)

var (
	//similar to http://golang.org/src/pkg/html/escape.go
	ampOrig = []byte(`&`)[0]
	ampRepl = []byte(`&amp;`)

	sgQuoteOrig = []byte(`'`)[0]
	sgQuoteRepl = []byte(`&#39;`)

	dblQuoteOrig = []byte(`"`)[0]
	dblQuoteRepl = []byte(`&#34;`)

	ltQuoteOrig = []byte(`<`)[0]
	ltQuoteRepl = []byte(`&lt;`)

	gtQuoteOrig = []byte(`>`)[0]
	gtQuoteRepl = []byte(`&gt;`)
)

// EscapeHTML wraps an http.ResponseWriter in order to override
// its Write method so that it escape html special chars while writing
type EscapeHTML struct {
	http.ResponseWriter
}

// make sure to fulfill the Contexter interface
var _ Contexter = &EscapeHTML{}

// Context gets the Context of the underlying response writer. It panics if the underlying response writer
// does no implement Contexter
func (e *EscapeHTML) Context(ctxPtr interface{}) bool {
	return e.ResponseWriter.(Contexter).Context(ctxPtr)
}

// SetContext sets the Context of the underlying response writer. It panics if the underlying response writer
// does no implement Contexter
func (e *EscapeHTML) SetContext(ctxPtr interface{}) {
	e.ResponseWriter.(Contexter).SetContext(ctxPtr)
}

// Write writes to the inner *http.ResponseWriter escaping html special chars on the fly
// Since there is nothing useful to do with the number of bytes written returned from
// the inner responsewriter, the returned int is always 0. Since there is nothing useful to do
// in case of a failed write to the response writer, writing errors are silently dropped.
// the method is modelled after EscapeText from encoding/xml
func (e *EscapeHTML) Write(b []byte) (num int, err error) {
	var esc []byte
	n := len(b)
	last := 0

	for i := 0; i < n; {
		r, width := utf8.DecodeRune(b[i:])
		i += width
		switch r {
		case '&':
			esc = ampRepl
		case '\'':
			esc = sgQuoteRepl
		case '"':
			esc = dblQuoteRepl
		case '<':
			esc = ltQuoteRepl
		case '>':
			esc = gtQuoteRepl
		default:
			continue
		}

		e.ResponseWriter.Write(b[last : i-width])
		e.ResponseWriter.Write(esc)
		last = i
	}

	e.ResponseWriter.Write(b[last:])
	return
}
