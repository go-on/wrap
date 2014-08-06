package wrap

import (
	"fmt"
	"net/http/httptest"

	"net/http"
)

// Wrapper can wrap a http.Handler with another one
type Wrapper interface {
	// Wrap wraps the next `http.Handler` of the stack and returns a wrapping `http.Handler`.
	// If it does not call `next.ServeHTTP`, nobody will.
	Wrap(next http.Handler) http.Handler
}

// New returns the wrapping http.Handler that returned from calling
// the Wrap method on the first given wrapper that received the returned
// handler of returning from the second wrappers Wrap method and so on.
//
// The last wrapper begins the loop receiving the NoOp handler.
//
// When the ServeHTTP method of the returned handler is called each wrapping
// handler may call is next until the NoOp handler is run.
//
// Or some wrapper decides not to call next.ServeHTTP.
//
// If DEBUG is set, each handler is wrapped with a Debug struct that calls DEBUGGER.Debug before
// running the handler.
func New(wrapper ...Wrapper) (h http.Handler) {
	if DEBUG {
		return _debug(wrapper...)
	}
	h = NoOp
	for i := len(wrapper) - 1; i >= 0; i-- {
		h = wrapper[i].Wrap(h)
	}
	return
}

// WrapperFunc is an adapter for a function that acts as Wrapper
type WrapperFunc func(http.Handler) http.Handler

// Wrap makes the WrapperFunc fulfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(next http.Handler) http.Handler { return wf(next) }

// NoOp is a http.Handler doing nothing
var NoOp = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

// ContextInjecter injects itself as Contexter into a middleware stack via
// its Wrapper interface
type ContextInjecter interface {

	// Contexter interface must be implemented on a pointer receiver of the struct
	Contexter

	// Wrapper interface might be implemented on the struct itself
	Wrapper
}

// a type that can't be supported by the Contexter (because it is not exported)
type contextUnsupported int

func validatecontextInjecterUnsupportedSetter(inject ContextInjecter) (panicked bool, correctError bool, correctType bool) {

	defer func() {
		if p := recover(); p != nil {
			// fmt.Printf("p is %T\n", p)
			panicked = true
			unspp, ok := p.(*ErrUnsupportedContextSetter)
			if ok {
				correctError = true
				var cu = contextUnsupported(0)
				if fmt.Sprintf("%T", unspp.Type) == fmt.Sprintf("%T", &cu) {
					correctType = true
				}
			}
		}
	}()

	rec := httptest.NewRecorder()
	var next http.HandlerFunc
	next = func(rw http.ResponseWriter, req *http.Request) {
		var unsupported contextUnsupported = 4
		rw.(Contexter).SetContext(&unsupported)
	}
	inject.Wrap(next).ServeHTTP(rec, nil)
	return
}

func validatecontextInjecterUnsupportedGetter(inject ContextInjecter) (panicked bool, correctError bool, correctType bool) {

	defer func() {
		if p := recover(); p != nil {
			// fmt.Printf("p is %T\n", p)
			panicked = true
			unspp, ok := p.(*ErrUnsupportedContextGetter)
			if ok {
				correctError = true
				var cu = contextUnsupported(0)
				if fmt.Sprintf("%T", unspp.Type) == fmt.Sprintf("%T", &cu) {
					correctType = true
				}
			}
		}
	}()

	rec := httptest.NewRecorder()
	var next http.HandlerFunc
	next = func(rw http.ResponseWriter, req *http.Request) {
		var unsupported contextUnsupported
		rw.(Contexter).Context(&unsupported)
	}
	inject.Wrap(next).ServeHTTP(rec, nil)
	return
}

func validateContextInjecterSupportsResponseWriter(inject ContextInjecter) {
	rec := httptest.NewRecorder()
	var nextCalled bool
	var next http.HandlerFunc
	next = func(rw http.ResponseWriter, req *http.Request) {
		nextCalled = true
		ctx := rw.(Contexter)
		var rw2 http.ResponseWriter
		if !ctx.Context(&rw2) {
			panic(fmt.Sprintf("%T.Context() does not support *http.ResponseWriter", ctx))
		}

		rec2, ok := rw2.(*httptest.ResponseRecorder)
		if !ok {
			panic(fmt.Sprintf("%T.Context() does not return the wrapped *http.ResponseWriter", ctx))
		}

		if rec2 != rec {
			panic(fmt.Sprintf("%T.Context() does not return the wrapped *http.ResponseWriter", ctx))
		}
	}
	inject.Wrap(next).ServeHTTP(rec, nil)
	if !nextCalled {
		panic(fmt.Sprintf("%T.Wrap() does not call the next http.Handler", inject))
	}
}

// ValidateContextInjecter panics if inject does not inject a Contexter that supports
// http.ResponseWriter,otherwise it returns true, so you may use it in var declarations
// that are executed before the init functions
func ValidateContextInjecter(inject ContextInjecter) bool {
	validateContextInjecterSupportsResponseWriter(inject)
	panicked, correctErr, correctType := validatecontextInjecterUnsupportedGetter(inject)
	if !panicked {
		panic(fmt.Sprintf("%T.Context() does not panic for unknown types", inject))
	}
	if !correctErr {
		panic(fmt.Sprintf("%T.Context() panic does not panic with *ErrUnsupportedContextGetter", inject))
	}
	if !correctType {
		panic(fmt.Sprintf("%T.Context() panic does set *ErrUnsupportedContextGetter with correct type", inject))
	}
	panicked, correctErr, correctType = validatecontextInjecterUnsupportedSetter(inject)
	if !panicked {
		panic(fmt.Sprintf("%T.SetContext() does not panic for unknown types", inject))
	}
	if !correctErr {
		panic(fmt.Sprintf("%T.SetContext() panic does not panic with *ErrUnsupportedContextSetter", inject))
	}
	if !correctType {
		panic(fmt.Sprintf("%T.SetContext() panic does set *ErrUnsupportedContextSetter with correct type", inject))
	}
	return true
}

// ContextWrapper is a Wrapper that uses some kind of context
// It has a Validate() method that panics if the contexts does not
// support the types that the Wrapper wants to store/retrieve inside the context
type ContextWrapper interface {
	Wrapper

	// ValidateContext should panic if the given Contexter does not support the required
	// types
	ValidateContext(Contexter)
}

// ValidateWrapperContexts validates the given Contexter against all of the
// given wrappers that implement the ContextWrapper interface.
// If every middleware that requires context implements the ContextWrapper
// interface and is passed to this function, then any missing support for a context type
// needed by a Wrapper would be uncovered. If then this function is called early it
// would save many headaches.
func ValidateWrapperContexts(ctx Contexter, wrapper ...Wrapper) {
	for _, wr := range wrapper {
		val, ok := wr.(ContextWrapper)
		if ok {
			val.ValidateContext(ctx)
		}
	}
}

// Stack creates a stack of middlewares with a context that is injected via inject.
// After validating the ContextInjecter it adds it at first middleware into the stack
// and returns the stack built by New
// This has the effect that the context is injected into the middleware chain at the beginning
// and every middleware may type assert the ResponseWriter to a Contexter in order to get and
// set context.
// Stack panics if inject is not valid.
// Stack should only be called once per application and must not be embedded into other stacks
func Stack(inject ContextInjecter, wrapper ...Wrapper) (h http.Handler) {
	ValidateContextInjecter(inject)
	st := []Wrapper{inject}
	st = append(st, wrapper...)
	return New(st...)
}
