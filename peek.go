package wrap

import (
	"io"
	"net/http"
)

// Peek is a ResponseWriter wrapper that intercepts the writing of the body, allowing to check headers and
// status code that has been set to prevent the body writing and to write a modified body.
//
// Peek is more efficient than Buffer, since it does not write to a buffer first before determining
// if the body will be flushed to the underlying response writer.
type Peek struct {
	// the cached status code
	Code int

	// the underlying response writer
	http.ResponseWriter

	changed        bool
	header         http.Header
	writeForbidden bool
	isChecked      bool
	codeWritten    bool
	headersWritten bool
	bodyWritten    bool
	// proceed should return true if the data should be written to the inner ResponseWriter
	// otherwise false
	// Proceed may check the Code and headers that have been set and to the Peek
	// and may also decide to transfer them to the inner ResponseWriter or set them directly on
	// the ResponseWriter. Proceed can be sure to be invoked before the first write to http.ResponseWriter
	proceed func(*Peek) bool
}

// NewPeek creates a new Peek for the given response writer using the given proceed function.
//
// The proceed function is called when the Write method is run for the first time.
// It receives the Peek and may check the cached headers and the cached status code.
//
// If the cached headers and the cached status code should be flushed to the underlying
// response writer, the proceed function must do so (e.g. by calling FlushMissing). This also allows to write other headers
// status codes or not write them at all.
//
// If the proceed function returns true, the body will be written to the underlying response write.
// That also holds for all following calls of Write when proceed is not run anymore.
//
// To write some other body or no body at all, proceed must return false,
// Then after the Write method has been run the Peek might be checked again and the underlying
// ResponseWriter that is exposed by Peek might be used to write a custom body.
//
// However if the http.Handler that receives the Peek does not write to the body, proceed will not
// be called at all.
//
// To ensure that any cached headers and status code will be flushed, the FlushMissing
// method can be called after the serving http.Handler is run.
//
// If proceed is nil, Write behaves as if proceed would have returned true.
func NewPeek(rw http.ResponseWriter, proceed func(*Peek) bool) *Peek {
	return &Peek{ResponseWriter: rw, proceed: proceed, header: make(http.Header)}
}

// FlushMissing ensures that the Headers and Code are written to the
// underlying ResponseWriter if they are not written yet (and nothing has been written to the body)
func (p *Peek) FlushMissing() {
	if p.bodyWritten || p.codeWritten {
		return
	}
	p.FlushHeaders()
	p.FlushCode()
}

// Context gets the Context of the underlying response writer. It panics if the underlying response writer
// does no implement Context
func (p *Peek) Context(ctxPtr interface{}) bool {
	return p.ResponseWriter.(Contexter).Context(ctxPtr)
}

// SetContext sets the Context of the underlying response writer. It panics if the underlying response writer
// does no implement Context
func (p *Peek) SetContext(ctxPtr interface{}) {
	p.ResponseWriter.(Contexter).SetContext(ctxPtr)
}

// Header returns the cached http.Header, tracking the call as change
func (p *Peek) Header() http.Header {
	p.changed = true
	return p.header
}

// WriteHeader writes the cached status code, tracking the call as change
func (p *Peek) WriteHeader(i int) {
	p.changed = true
	p.Code = i
}

// IsOk returns true if the returned status code is
// not set or in the 2xx range
func (p *Peek) IsOk() bool {
	if p.Code == 0 {
		return true
	}
	if p.Code >= 200 && p.Code < 300 {
		return true
	}
	return false
}

// Write writes to the underlying response writer, if the proceed function
// returns true. Otherwise it returns 0, io.EOF.
// If the data is written, the call is tracked as change.
//
// The proceed function is only called the first time, Write has been called.
// If proceed is nil, it behaves as if proceed would have returned true.
//
// See NewPeek for more informations about the usage of the proceed function.
func (p *Peek) Write(b []byte) (int, error) {
	if p.proceed != nil {
		if !p.isChecked {
			p.writeForbidden = !p.proceed(p)
			p.isChecked = true
		}
	}
	if p.writeForbidden {
		return 0, io.EOF
	}
	p.bodyWritten = true
	p.changed = true
	return p.ResponseWriter.Write(b)
}

// Reset set the Peek to the defaults, so it will act as if it was freshly initialized.
func (p *Peek) Reset() {
	p.Code = 0
	p.header = make(http.Header)
	p.changed = false
	p.writeForbidden = false
	p.isChecked = false
	p.codeWritten = false
	p.headersWritten = false
	p.bodyWritten = false
}

// HasChanged returns true if Header or WriteHeader method have been called or if
// Write has been called and did write to the underlying response writer.
func (p *Peek) HasChanged() bool {
	return p.changed
}

// FlushCode writes the status code to the underlying responsewriter if it was set
func (p *Peek) FlushCode() {

	if p.codeWritten {
		return
	}

	if p.bodyWritten {
		panic(ErrBodyFlushedBeforeCode{})
	}

	if p.Code != 0 {
		p.ResponseWriter.WriteHeader(p.Code)
		p.codeWritten = true
	}

}

// FlushHeaders adds the headers to the underlying ResponseWriter, removing them from Peek
func (p *Peek) FlushHeaders() {
	if p.headersWritten {
		return
	}
	if p.codeWritten {
		panic(ErrCodeFlushedBeforeHeaders{})
	}
	if p.bodyWritten {
		panic(ErrBodyFlushedBeforeCode{})
	}
	header := p.ResponseWriter.Header()
	for k, v := range p.header {
		header.Del(k)
		for _, val := range v {
			header.Add(k, val)
		}
	}
	p.headersWritten = true
}
