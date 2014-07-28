package wrap

import (
	"bytes"
	"io"
	"net/http"
	"unicode/utf8"
)

// RWContext is a http.ResponseWriter that can set and get contexts
type RWContext interface {
	http.ResponseWriter

	// Context lets the given ctxPtr point to the saved context of the same type
	Context(ctxPtr interface{})

	// SetContext saves the given context pointer via type switch
	SetContext(ctxPtr interface{})
}

type RWPeek struct {
	http.ResponseWriter
	Code           int
	changed        bool
	header         http.Header
	writeForbidden bool
	isChecked      bool
	codeWritten    bool
	headersWritten bool
	bodyWritten    bool
	// proceed should return true if the data should be written to the inner ResponseWriter
	// otherwise false
	// Proceed may check the Code and headers that have been set and to the RWPeek
	// and may also decide to transfer them to the inner ResponseWriter or set them directly on
	// the ResponseWriter. Proceed can be sure to be invoked before the first write to http.ResponseWriter
	proceed func(*RWPeek) bool
}

func NewRWPeek(rw http.ResponseWriter, proceed func(*RWPeek) bool) *RWPeek {
	return &RWPeek{ResponseWriter: rw, proceed: proceed, header: make(http.Header)}
}

// FlushMissing ensures that the Headers and Code are written to the
// underlying ResponseWriter if they are not written yet (and nothing has been written to the body)
func (f *RWPeek) FlushMissing() {
	if f.bodyWritten || f.codeWritten {
		return
	}
	f.FlushHeaders()
	f.FlushCode()
}

func (f *RWPeek) Context(ctxPtr interface{}) {
	f.ResponseWriter.(RWContext).Context(ctxPtr)
}

func (f *RWPeek) SetContext(ctxPtr interface{}) {
	f.ResponseWriter.(RWContext).SetContext(ctxPtr)
}

// Header returns the http.Header
func (f *RWPeek) Header() http.Header {
	f.changed = true
	return f.header
}

// WriteHeader writes the status code
func (f *RWPeek) WriteHeader(i int) {
	f.changed = true
	f.Code = i
}

// IsOk returns true if the returned status code is
// not set or in the 2xx range
func (f *RWPeek) IsOk() bool {
	if f.Code == 0 {
		return true
	}
	if f.Code >= 200 && f.Code < 300 {
		return true
	}
	return false
}

// Write only writes if writing is allowed
func (f *RWPeek) Write(b []byte) (int, error) {
	if f.proceed != nil {
		if !f.isChecked {
			f.writeForbidden = !f.proceed(f)
			f.isChecked = true
		}
	}
	if f.writeForbidden {
		return 0, io.EOF
	}
	f.bodyWritten = true
	f.changed = true
	return f.ResponseWriter.Write(b)
}

// Reset set the RWPeek to the defaults
func (f *RWPeek) Reset() {
	f.Code = 0
	f.header = make(http.Header)
	f.changed = false
	f.writeForbidden = false
	f.isChecked = false
	f.codeWritten = false
	f.headersWritten = false
	f.bodyWritten = false
}

func (f *RWPeek) HasChanged() bool {
	return f.changed
}

// FlushCodeTo writes the status code to the underlying responsewriter if it was set
func (f *RWPeek) FlushCode() {
	if f.codeWritten {
		return
	}
	if f.bodyWritten {
		panic(BodyFlushedBeforeCode{})
		return
	}
	if f.Code != 0 {
		f.ResponseWriter.WriteHeader(f.Code)
		f.codeWritten = true
	}
}

// FlushHeaders adds the headers to the underlying ResponseWriter
func (f *RWPeek) FlushHeaders() {
	if f.headersWritten {
		return
	}
	if f.codeWritten {
		panic(CodeFlushedBeforeHeaders{})
	}
	if f.bodyWritten {
		panic(BodyFlushedBeforeCode{})
	}
	header := f.ResponseWriter.Header()
	for k, v := range f.header {
		header.Del(k)
		for _, val := range v {
			header.Add(k, val)
		}
	}
	f.headersWritten = true
}

// RWBuffer is a ResponseWriter wrapper that may be used as buffer.
type RWBuffer struct {
	http.ResponseWriter
	Buffer  bytes.Buffer
	Code    int
	changed bool
	header  http.Header
}

// Header returns the http.Header
func (f *RWBuffer) Header() http.Header {
	f.changed = true
	return f.header
}

// WriteHeader writes the status code
func (f *RWBuffer) WriteHeader(i int) { f.changed = true; f.Code = i }

// Write writes to the buffer
func (f *RWBuffer) Write(b []byte) (int, error) {
	f.changed = true
	return f.Buffer.Write(b)
}

// Reset set the RWBuffer to the defaults
func (f *RWBuffer) Reset() {
	f.Buffer.Reset()
	f.Code = 0
	f.changed = false
	f.header = make(http.Header)
}

// WriteTo writes header, body and status code to the underlying ResponseWriter, if something changed
func (f *RWBuffer) FlushAll() {
	if f.HasChanged() {
		f.FlushHeaders()
		f.FlushCode()
		f.ResponseWriter.Write(f.Buffer.Bytes())
	}
}

// Body returns the body as slice of bytes
func (f *RWBuffer) Body() []byte {
	return f.Buffer.Bytes()
}

// BodyString returns the body as string
func (f *RWBuffer) BodyString() string {
	return f.Buffer.String()
}

// HasChanged returns true if something has been written to the RWBuffer
func (f *RWBuffer) HasChanged() bool { return f.changed }

// IsOk returns true if the returned status code is
// not set or in the 2xx range
func (f *RWBuffer) IsOk() bool {
	if f.Code == 0 {
		return true
	}
	if f.Code >= 200 && f.Code < 300 {
		return true
	}
	return false
}

// FlushCodeTo writes the status code to the underlying responsewriter if it was set
func (f *RWBuffer) FlushCode() {
	if f.Code != 0 {
		f.ResponseWriter.WriteHeader(f.Code)
		// w.WriteHeader(f.Code)
	}
}

// FlushHeadersTo adds the headers to the underlying ResponseWriter
func (f *RWBuffer) FlushHeaders() {
	header := f.ResponseWriter.Header()
	for k, v := range f.header {
		header.Del(k)
		for _, val := range v {
			header.Add(k, val)
		}
	}
}

// NewRWBuffer creates a new RWBuffer
func NewRWBuffer(w http.ResponseWriter) (f *RWBuffer) {
	f = &RWBuffer{}
	f.ResponseWriter = w
	f.header = make(http.Header)
	return
}

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

// RWEscapeHTML wraps an http.ResponseWriter in order to override
// its Write method so that it escape html special chars while writing
type RWEscapeHTML struct {
	http.ResponseWriter
}

func (f *RWEscapeHTML) Context(ctxPtr interface{}) {
	f.ResponseWriter.(RWContext).Context(ctxPtr)
}

func (f *RWEscapeHTML) SetContext(ctxPtr interface{}) {
	f.ResponseWriter.(RWContext).SetContext(ctxPtr)
}

// Write writes to the inner *http.ResponseWriter escaping html special chars on the fly
// Since there is nothing useful to do with the number of bytes written returned from
// the inner responsewriter, the returned int is always 0. Since there is nothing useful to do
// in case of a failed write to the response writer, writing errors are silently dropped.
// the method is modelled after EscapeText from encoding/xml
func (rw *RWEscapeHTML) Write(b []byte) (num int, err error) {
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

		rw.ResponseWriter.Write(b[last : i-width])
		rw.ResponseWriter.Write(esc)
		last = i
	}

	rw.ResponseWriter.Write(b[last:])
	return
}
