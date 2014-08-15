package wrap

import (
	"fmt"
)

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

// ErrUnsupportedContextSetter is the error returned if the context type is not supported by the SetContext()
// method of a Contexter
type ErrUnsupportedContextSetter struct {
	Type interface{}
}

func (e *ErrUnsupportedContextSetter) Error() string {
	return fmt.Sprintf("setting the context type %T is not supported by the Contexter", e.Type)
}

// ErrUnsupportedContextGetter is the error returned if the context type is not supported by the Context()
// method of a Contexter
type ErrUnsupportedContextGetter struct {
	Type interface{}
}

func (e *ErrUnsupportedContextGetter) Error() string {
	return fmt.Sprintf("getting the context type %T is not supported by the Contexter", e.Type)
}
