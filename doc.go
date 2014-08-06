// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package wrap creates a fast and flexible middleware stack for http.Handlers.

Each middleware is a wrapper for another middleware and implements the
Wrapper interface.

Features

  - small; core is only 13 LOC
  - based on http.Handler interface; nicely integrates with net/http
  - middleware stacks are http.Handlers too and may be embedded
  - has a solution for per request context sharing
  - has debugging helper
  - low memory footprint
  - fast
  - no dependency apart from the standard library

Wrappers can be found at http://godoc.org/github.com/go-on/wrap-contrib/wraps.

A (mountable) router that plays fine with wrappers can be found at http://godoc.org/github.com/go-on/router.

Status

100% test coverage.
This package is considered complete and the API is stable.

Benchmarks (Go 1.3)

	// The overhead of n writes to http.ResponseWriter via n wrappers
	// vs n writes in a loop within a single http.Handler

  BenchmarkServing2Simple     1000000 1067    ns/op   1,00x
  BenchmarkServing2Wrappers   1000000 1121    ns/op   1,05x

  BenchmarkServing50Simple    100000  26041   ns/op   1,00x
  BenchmarkServing50Wrappers  100000  27053   ns/op   1,04x

  BenchmarkServing100Simple   50000   52074   ns/op   1,00x
  BenchmarkServing100Wrappers 50000   53450   ns/op   1,03x

Credits

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.


Content of the package

The core of this package is the New function that constructs a stack of middlewares that implement
the Wrapper interface.

If the global DEBUG flag is set before calling New then each middleware call will result in
calling the Debug method of the global DEBUGGER (defaults to a logger).

To help constructing middleware there are some adapters like WrapperFunc, Handler, HandlerFunc,
NextHandler and NextHandlerFunc each of them adapting to the Wrapper interface.

To help sharing per request context there is a Contexter interface that must be implemented by
the ResponseWriter. That can easily be done be providing a middleware that injects a context
that wraps the current ResponseWriter and implements the Contexter interface. It must a least support
the extraction of the wrapped ResponseWriter.

An example can be found in the file example_context_test.go.

Furthermore this package provides some ResponseWriter wrappers that also implement Contexter and
that help with development of middleware.

These are Buffer, Peek and EscapeHTML.

Buffer is a simple buffer. A middleware may pass it to the next handlers ServeHTTP method as a
drop in replacement for the response writer. After the ServeHTTP method is run the middleware may
examine what has been written to the Buffer and decide what to write to the "original" ResponseWriter
(that may well be another buffer passed from another middleware).

The downside is the body being written two times and the complete caching of the
body in the memory which will be inacceptable for large bodies.

Therefor Peek is an alternative response writer wrapper that only caching headers and status code
but allowing to intercept calls of the Write method. All middleware without the need to read
the whole response body should use Peek or provide their own ResponseWriter wrapper (then do not
forget to implement the Contexter interface).

Finally EscapeHTML provides a response writer wrapper that allows on the fly
html escaping of the bytes written to the wrapped response writer.


How to write a middleware

It is pretty easy to write your custom middleware. You should start with a new struct
type - that allows you to add options as fields later on.

Then you could use the following template to implement the Wrapper interface

    type MyMiddleware struct {
	     // add your options
    }

    func (m MyMiddleware) Wrap( next http.Handler) http.Handler {
	     var f http.HandlerFunc
	     f = func (rw http.ResponseWriter, req *http.Request) {
	        // here is where your magic happens

	        // at some point you might want to run the next handler
	        // if not, your middleware ends the stack chain
	        next.ServeHTTP(rw http.ResponseWriter, req *http.Request)
	     }
	     return f
    }

If you need to run the next handler in order to inspect what it did,
replace the response writer with a Peek (see NewPeek) or if you need
full access to the written body with a Buffer.

To use per request context a custom type is needed that carries the context data and
the user is expected to create and inject a Contexter supporting this type.
See the documentation of Contexter for more information.

Then inside your f function type assert the response writer to a wrap.Contexter
and use the SetContext and Context methods to store and retrieve your context data.
Always pass a pointer of the context object to these methods.

Don't forget to document that your middleware expects the response writer to
implement the Contexter interface and to support your context type.

You might want to look at existing middlewares to get some ideas:
http://godoc.org/github.com/go-on/wrap-contrib/wraps

How to write a Contexter

A Contexter is expected to be implemented in tandem with the creation of the
outer application wide middelware stack. If a middleware requires specific contextual
data to be passed along some request it defines a type for it that carries those data
and expects the user to implement a Contexter supporting this type that is passed to it
as http.ResponseWriter.

There must only be one Contexter per application stack and it should be injected
into the outer most stack at the beginning.

Injecting a Contexter into the middleware stack means adding a Wrapper (may be the Contexter itself)
that does pass the Contexter wrapping the original ResponseWriter as ResponseWriter to the ServeHTTP
method of the next http.Handler in the stack.

Doing so at the very beginning has the effect that all following http.Handler in the stack receive a
ResponseWriter that implements Contexter. That allows all of them to set and get context data freely.
The only condition is that the type of the context data must be supported by the Contexter.
That is a hidden dependency (some middleware depends on the Contexter to support its context data type)
and should be covered by unit tests.

However it gives great freedom when organizing and ordering middleware as the middleware in between
does not have to know about a certain context type if does not care about it.

It simply passes the ResponseWriter that happens to have context down the middleware stack chain.

On the other hand, exposing the context type by making it a parameter has the effect that every middleware
will need to pass this parameter to the next one if the next one needs it. That requirement however
may change making context sharing and reusability of middlewares from different sources virtually impossible.

When implementing a Contexter it is important to support the *http.ResponseWriter type by the Context()
method, because it allows middleware to retrieve the original response writer and type assert it to
some interfaces of the standard http library, e.g. http.Flusher. To make this easier there are some
helper functions like Flush(), CloseNotify() and Hijack() - all making use of the more general ReclaimResponseWriter().

Template for the implementation of a Contexter

    type context struct {
      http.ResponseWriter
      err error
    }

    // make sure to fulfill the Contexter interface
    var _ wrap.Contexter = &context{}

    // Context receives a pointer to a value that is already stored inside the context.
    // Values are distiguished by their type.
    // Context sets the value of the given pointer to the value of the same type
    // that is stored inside of the context.
    // A pointer type that is not supported results in a panic.
    // Context returns if ctxPtr will be nil after return
    // Context must support *http.ResponseWriter
    func (c *context) Context(ctxPtr interface{}) (found bool) {
      found = true
      switch ty := ctxPtr.(type) {
      case *http.ResponseWriter:
        *ty = c.ResponseWriter
      case *error:
        if c.err == nil {
          return false
        }
        *ty = c.err
      default:
        panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
      }
      return
    }

    // SetContext receives a pointer to a value that will be stored inside the context.
    // Values are distiguished by their type, that means that SetContext replaces
    // and stored value of the same type.
    // A pointer type that is not supported results in a panic.
    func (c *context) SetContext(ctxPtr interface{}) {
      switch ty := ctxPtr.(type) {
      case *error:
        c.err = *ty
      default:
        panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
      }
    }

    // Wrap implements the wrap.Wrapper interface and injects the Contexter to the
    // middleware stack.
    //
    // When the request is served, the response writer is wrapped by a
    // new *context which is passed to the next handlers ServeHTTP method.
    func (c context) Wrap(next http.Handler) http.Handler {
      var f http.HandlerFunc
      f = func(rw http.ResponseWriter, req *http.Request) {
        next.ServeHTTP(&context{ResponseWriter: rw}, req)
      }
      return f
    }

While it looks like much work it is in fact not, since this is only written once per application and
the effort to support a new context data type is low.

Lets have a look at the effort to support a new context type (here: error):

1. Add a new field to your context (1 LOC)

    // type context struct {
    //  ...
        err error
    // }

2. Add a new case to retrieve the context data (5 LOC)

    // func (c *context) Context(ctxPtr interface{}) (found bool) {
    //  ...
    //  switch ty := ctxPtr.(type) {
    //  ...
        case *error:
          if c.err == nil {
            return false
          }
          *ty = c.err
    //  }
    //  ...
    // }

3. Add a new case to set the context data (3 LOC)

    // func (c *context) SetContext(ctxPtr interface{}) {
    //  ...
    //  switch ty := ctxPtr.(type) {
    //  ...
        case *error:
          c.err = *ty
        }
    //  ...
    // }

So 9 simple LOC to support a new context type is not that bad.

A middleware using the error context could look like this:

    var StopIfError wrap.NextHandlerFunc = func (next http.Handler, rw http.ResponseWriter, req *http.Request) {
      var err error
      if rw.(wrap.Contexter).Context(&err) {
         rw.WriteHeader(500)

         if os.Getenv("DEVELOPMENT") != "" {
           fmt.Fprintf(rw, "Error:\n\n%s", err.Error())
           return
         }

         fmt.Fprint(rw, "Internal server error")
         return
      }
      next.ServeHTTP(rw,req)
    }

FAQ

1. Why is the recommended way to use the Contexter interface to make a type assertion from
the ResponseWriter to the Contexter interface without error handling?

Answer: Middleware stacks should be created before the server is handling requests and then not change
anymore. And you should run tests. Then the assertion will blow up and that is correct because
there is no reasonable way to handle such an error. It means that either you have no context in your
stack or you inject the context to late or the context does not handle the kind of type the middleware
expects. In each case you should fix it early and the panic forces you to do.
Use the DEBUG flag to see what's going on.

2. What happens if my context is wrapped inside another context or response writer?

Answer: You should only have one Contexter per application and inject it as first wrapper into your
middleware stack. All context specific data belongs there. Having multiple Contexter in a stack is considered a bug.

The good news is that you now can be sure to be able to access all of your context data
everywhere inside your stack.

Never should a context wrap another one. There also should be no need for another context wrapping
ResponseWriter because every type can be saved by a Contexter with few code.

All response writer wrappers of this package fulfill the Contexter interface and every response
writer you use should.

3. Why isn't there any default context object? Why do I have to write it on my own?

Answer: To write your own context and context injection has several benefits:

  - no magic or reflection necessary
  - all your context data is managed at one place
  - context data may be generated/updated based on other context data
  - your context management is independant from the middleware

4. Why is the context data accessed and stored via type switch and not by string keys?

Answer: Type based context allows a very simple implementation that namespaces across packages needs
no extra memory allocation and is, well, type safe. If you need to store multiple context
data of the same type, simple defined an alias type for each key.

5. Is there an example how to integrate with 3rd party middleware libraries that expect context?

Answer: Yes, have a look at http://godoc.org/github.com/go-on/wrap-contrib/third-party.

6. What about spawning goroutines / how does it relate to code.google.com/p/go.net/context?

If you need your request to be handled by different goroutines and middlewares you might use
your Contexter to store and provide access to a code.google.com/p/go.net/context Context just
like any other context data. The good news is that you can write you middleware like before
and extract your Context everywhere you need it. And if you have middleware in between
that does not know about about Context or other context data you don't have build wrappers around that
have Context as parameter. Instead the ResponseWriter that happens to provide you a Context will be passed down the
chain by the middleware.

"At Google, we require that Go programmers pass a Context parameter as the first argument to every
function on the call path between incoming and outgoing requests."

(Sameer Ajmani, http://blog.golang.org/context)

This is not neccessary anymore. And it is not neccessary for any type of contextual data because
that does not have to be in the type signature anymore.

*/
package wrap
