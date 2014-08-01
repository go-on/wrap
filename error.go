package wrap

// ErrBodyFlushedBeforeCode is the error returned if a body flushed to an underlying response writer
// before the status code has been flushed. It should help to sort out errors in middleware that uses
// responsewriter wrappers from this package.
type ErrBodyFlushedBeforeCode struct{}

// Error returns the error message
func (e ErrBodyFlushedBeforeCode) Error() string {
	return "body flushed before code"
}

// ErrCodeFlushedBeforeHeaders is the error returned if a status code flushed to an underlying response writer
// before the headers have been flushed. It should help to sort out errors in middleware that uses
// responsewriter wrappers from this package.
type ErrCodeFlushedBeforeHeaders struct{}

// Error returns the error message
func (e ErrCodeFlushedBeforeHeaders) Error() string {
	return "code flushed before headers"
}
