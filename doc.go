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
  - freely mix middleware with and without context (same interface)
  - easy to create adapters / wrappers for 3rd party middleware

Wrappers can be found at http://godoc.org/github.com/go-on/wrap-contrib/wraps.

A (mountable) router that plays fine with wrappers can be found at http://godoc.org/github.com/go-on/router.


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

Then there are functions to validate implementations of Contexter (ValidateContextInjecter) and to validate them against
wrappers that store and retrieve the context data (ValidateWrapperContexts).

An complete example for shared contexts can be found in the file example_context_test.go.

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

    // make sure it conforms to the Wrapper interface
    var _ wrap.Wrapper = MyMiddleware{}

    // implement the wrapper interface
    func (m MyMiddleware) Wrap( next http.Handler) http.Handler {
	     var f http.HandlerFunc
	     f = func (rw http.ResponseWriter, req *http.Request) {
	        // do stuff

	        // at some point you might want to run the next handler
	        // if not, your middleware ends the stack chain
	        next.ServeHTTP(rw, req)
	     }
	     return f
    }

If you need to run the next handler in order to inspect what it did,
replace the response writer with a Peek (see NewPeek) or if you need
full access to the written body with a Buffer.

How to use middleware

To form a middleware stack, simply use New() and pass it the middlewares.
They get the request top-down. There are some adapters to let for example
a http.Handler be a middleware (Wrapper) that does not call the next handler
and stops the chain.

      stack := wrap.New(
          MyMiddleware{},
          OtherMiddleware{},
          wrap.Handler(aHandler),
      )

      // stack is now a http.Handler

How to write a middleware that uses per request context

To use per request context a custom type is needed that carries the context data and
the user is expected to create and inject a Contexter supporting this type.

Here is a template for a middleware that you could use to write middleware
that wants to use / share context.

    // MyMiddleware expects the ResponseWriter to implement wrap.Contexter and
    // to support storing and retrieving the MyContextData type.
    type MyMiddleware struct {
       // add your options
    }

    // define whatever type you like, but define a type
    // for each kind of context data you will want to store/retrieve
    type MyContextData string

    // make sure it conforms to the ContextWrapper interface
    var _ wrap.ContextWrapper = MyMiddleware{}

    // implements ContextWrapper; panics if Contexter does not support
    // the needed type
    func (m MyMiddleware) ValidateContext( ctx wrap.Contexter ) {
      var m MyContextData
      // try the getter and setter, they will panic if they don't support the type
      ctx.Context(&m); ctx.SetContext(&m)
      // do this for every type you need
    }

    // implement the wrapper interface
    func (m MyMiddleware) Wrap( next http.Handler) http.Handler {
       var f http.HandlerFunc
       f = func (rw http.ResponseWriter, req *http.Request) {

          ctx := rw.(wrap.Contexter)
          m := MyContextData("Hello World")
          ctx.SetContext(&m) // always pass the pointer

          var n MyContextData
          ctx.Context(&n)

          // n now is MyContextData("Hello World")

          ... do stuff
          next.ServeHTTP(rw, req)
       }
       return f
    }

How to use middleware that uses per request context

For context sharing the user has to implement the Contexter interface in a way that supports
all types the used middlewares expect.

Here is a template for an implementation of the Contexter interface

    type MyContext struct {
      http.ResponseWriter // you always need this
      myContextData *myMiddleware.MyContextData // a property for each supported type
    }

    // make sure it is a valid context, i.e. http.ResponseWriter is supported by Context
    // method, the correct types are returned and the panic types are correct
    var _ = wrap.ValidateContextInjecter(&MyContext{})

    // retrieves the value of the type to which ctxPtr is a pointer to
    func (c *MyContext) Context(ctxPtr interface{}) (found bool) {
      found = true // save work
      switch ty := ctxPtr.(type) {
      // always support http.ResponseWriter in Context method
      case *http.ResponseWriter:
        *ty = c.ResponseWriter
      // add such a case for each supported type
      case *myMiddleware.MyContextData:
        if c.myContextData == nil {
          return false
        }
        *ty = *c.myContextData
      default:
        // always panic with wrap.ErrUnsupportedContextGetter in Context method on default
        panic(&wrap.ErrUnsupportedContextGetter{ctxPtr})
      }
      return
    }

    // sets the context of the given type
    func (c *MyContext) SetContext(ctxPtr interface{}) {
      switch ty := ctxPtr.(type) {
      case *myMiddleware.MyContextData:
        c.myContextData = ty
      default:
        // always panic with wrap.ErrUnsupportedContextSetter in SetContext method on default
        panic(&wrap.ErrUnsupportedContextSetter{ctxPtr})
      }
    }

    // Wrap implements the wrap.Wrapper interface by wrapping a ResponseWriter inside a new
    // &MyContext and injecting it into the middleware chain.
    func (c MyContext) Wrap(next http.Handler) http.Handler {
      var f http.HandlerFunc
      f = func(rw http.ResponseWriter, req *http.Request) {
        next.ServeHTTP(&MyContext{ResponseWriter: rw}, req)
      }
      return f
    }

At any time there must be only one Contexter in the whole middleware stack and its the best
to let it be the first middleware. Then you don't have to worry if its there or not (the Stack function
might help you).

The corresponding middleware stack would look like this

      // first check if the Contexter supports all context types needed by the middlewares
      // this uses the ValidateContext() method of the middlewares that uses context.
      // It panics on errors.
      wrap.ValidateWrapperContexts(&MyContext{}, MyMiddleware{}, OtherMiddleware{})

      stack := wrap.New(
          MyContext{}, // injects the &MyContext{} wrapper, should be done at the beginning
          MyMiddleware{},
          OtherMiddleware{},
          wrap.Handler(aHandler),
      )

      // stack is now a http.Handler

If your application / handler also uses context data, it is a good idea to implement it as
ContextWrapper as if it were a middleware and pass it ValidateWrapperContexts(). So
if your context is wrong, you will get nice panics before your server even starts.
And this is always in sync with your app / middleware.

If for some reason the original ResponseWriter is needed (to type assert it to a http.Flusher
for example), it may be reclaimed with the help of ReclaimResponseWriter().

You might want to look at existing middlewares to get some ideas:
http://godoc.org/github.com/go-on/wrap-contrib/wraps

FAQ

1. Should the context not better be an argument to a middleware function, to make this
dependency visible in documentation and tools?

Answer: A unified interface gives great freedom when organizing and ordering middleware as the
middleware in between does not have to know about a certain context type if does not care about it.
It simply passes the ResponseWriter that happens to have context down the middleware stack chain.

On the other hand, exposing the context type by making it a parameter has the effect that every middleware
will need to pass this parameter to the next one if the next one needs it. Every middleware is then
tightly coupled to the next one and reordering or mixing of middleware from different sources
becomes impossible.

However with the ContextHandler interface and the ValidateWrapperContexts function we have a way
to guarantee that the requirements are met. And this solution is also type safe.

2. A ResponseWriter is an interface, because it may implement other interfaces from the http libary,
e.g. http.Flusher. If it is wrapped that underlying implementation is not accessible anymore

Answer: When the Contexter is validated, it is checked, that the Context method supports
http.ResponseWriter as well (and that it returns the underlying ResponseWriter). Since only one
Contexter may be used within a stack, it is always possible to ask the Contexter for the underlying
ResponseWriter. This is what helper functions like ReclaimResponseWriter(), Flush(), CloseNotify()
and Hijack() do.

3. Why is the recommended way to use the Contexter interface to make a type assertion from
the ResponseWriter to the Contexter interface without error handling?

Answer: Middleware stacks should be created before the server is handling requests and then not change
anymore. And you should run tests. Then the assertion will blow up and that is correct because
there is no reasonable way to handle such an error. It means that either you have no context in your
stack or you inject the context to late or the context does not handle the kind of type the middleware
expects. In each case you should fix it early and the panic forces you to do.
Use the DEBUG flag to see what's going on.

4. What happens if my context is wrapped inside another context or response writer?

Answer: You should only have one Contexter per application and inject it as first wrapper into your
middleware stack. All context specific data belongs there. Having multiple Contexter in a stack is considered a bug.

The good news is that you now can be sure to be able to access all of your context data
everywhere inside your stack.

Never should a context wrap another one. There also should be no need for another context wrapping
ResponseWriter because every type can be saved by a Contexter with few code.

All response writer wrappers of this package implement the Contexter interface and every response
writer you use should.

5. Why isn't there any default context object? Why do I have to write it on my own?

Answer: To write your own context and context injection has several benefits:

  - no magic or reflection necessary
  - all your context data is managed at one place
  - context data may be generated/updated based on other context data
  - your context management is independant from the middleware

6. Why is the context data accessed and stored via type switch and not by string keys?

Answer: Type based context allows a very simple implementation that namespaces across packages needs
no extra memory allocation and is, well, type safe. If you need to store multiple context
data of the same type, simple defined an alias type for each key.

7. Is there an example how to integrate with 3rd party middleware libraries that expect context?

Answer: Yes, have a look at http://godoc.org/github.com/go-on/wrap-contrib/third-party.

8. What about spawning goroutines / how does it relate to code.google.com/p/go.net/context?

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
