package wrap

import (
	"bytes"
	"net/http"
)

// Buffer is a ResponseWriter wrapper that may be used as buffer.
type Buffer struct {

	// ResponseWriter is the underlying response writer that is wrapped by Buffer
	http.ResponseWriter

	// Buffer is the underlying io.Writer that buffers the response body
	Buffer bytes.Buffer

	// Code is the cached status code
	Code int

	// changed tracks if anything has been set on the responsewriter. Also reads from the header
	// are seen as changes
	changed bool

	// header is the cached header
	header http.Header
}

// make sure to fulfill the Contexter interface
var _ Contexter = &Buffer{}

// NewBuffer creates a new Buffer by wrapping the given response writer.
func NewBuffer(w http.ResponseWriter) (bf *Buffer) {
	bf = &Buffer{}
	bf.ResponseWriter = w
	bf.header = make(http.Header)
	return
}

// Context gets the context of the underlying response writer. It panics if the underlying response writer
// does no implement Contexter
func (bf *Buffer) Context(ctxPtr interface{}) bool {
	return bf.ResponseWriter.(Contexter).Context(ctxPtr)
}

// SetContext sets the Context of the underlying response writer. It panics if the underlying response writer
// does no implement Contexter
func (bf *Buffer) SetContext(ctxPtr interface{}) {
	bf.ResponseWriter.(Contexter).SetContext(ctxPtr)
}

// Header returns the cached http.Header and tracks this call as change
func (bf *Buffer) Header() http.Header {
	bf.changed = true
	return bf.header
}

// WriteHeader writes the cached status code and tracks this call as change
func (bf *Buffer) WriteHeader(i int) {
	bf.changed = true
	bf.Code = i
}

// Write writes to the underlying buffer and tracks this call as change
func (bf *Buffer) Write(b []byte) (int, error) {
	bf.changed = true
	return bf.Buffer.Write(b)
}

// Reset set the Buffer to the defaults
func (bf *Buffer) Reset() {
	bf.Buffer.Reset()
	bf.Code = 0
	bf.changed = false
	bf.header = make(http.Header)
}

// FlushAll flushes headers, status code and body to the underlying ResponseWriter, if something changed
func (bf *Buffer) FlushAll() {
	if bf.HasChanged() {
		bf.FlushHeaders()
		bf.FlushCode()
		bf.ResponseWriter.Write(bf.Buffer.Bytes())
	}
}

// Body returns the bytes of the underlying buffer (that is meant to be the body of the response)
func (bf *Buffer) Body() []byte {
	return bf.Buffer.Bytes()
}

// BodyString returns the string of the underlying buffer (that is meant to be the body of the response)
func (bf *Buffer) BodyString() string {
	return bf.Buffer.String()
}

// HasChanged returns true if Header, WriteHeader or Write has been called
func (bf *Buffer) HasChanged() bool {
	return bf.changed
}

// IsOk returns true if the cached status code is not set or in the 2xx range.
func (bf *Buffer) IsOk() bool {
	if bf.Code == 0 {
		return true
	}
	if bf.Code >= 200 && bf.Code < 300 {
		return true
	}
	return false
}

// FlushCode flushes the status code to the underlying responsewriter if it was set.
func (bf *Buffer) FlushCode() {
	if bf.Code != 0 {
		bf.ResponseWriter.WriteHeader(bf.Code)
	}
}

// FlushHeaders adds the headers to the underlying ResponseWriter, removing them from Buffer.
func (bf *Buffer) FlushHeaders() {
	header := bf.ResponseWriter.Header()
	for k, v := range bf.header {
		header.Del(k)
		for _, val := range v {
			header.Add(k, val)
		}
	}
}
