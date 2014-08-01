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

Wrappers can be found at github.com/go-on/wrap-contrib.

A (mountable) router that plays fine with wrappers can be found at github.com/go-on/router.

Status

100% test coverage.
This package is considered complete and the API is stable.

Benchmarks (Go 1.3):

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
a calling the Debug method of the global DEBUGGER (defaults to a logger).

To help constructing middleware there are some adapters like WrapperFunc, Handler, HandlerFunc,
NextHandler and NextHandlerFunc each of them adapting to the Wrapper interface.

To help sharing per request context there is a Contexter interface that must be implemented by
the ResponseWriter. That can easily be done be providing a middleware that injects a context
that wraps the current ResponseWriter and implements the Contexter interface.

An example can be found in the file example_context_test.go.

Furthermore this package provides some ResponseWriter wrappers that respect the possibility that
the inner ResponseWriter implements the Contexter interface and that help with development of middleware.

These are Buffer, Peek and EscapeHTML.

Buffer is a simple buffer. A middleware may pass it to the next handlers ServeHTTP method as a
drop in replacement for the response writer.

After the ServeHTTP method is run the middleware may examine what has been written to the Buffer and
decide what to write to the "real" ResponseWriter (that may well be another buffer passed from another
middleware).

The disadvantage of the Buffer is that the body of the response is written two times and that the complete
body must be cached in the memory which will be inacceptable for large bodies.

Therefor Peek is an alternative response writer wrapper that only caches the headers and the
status code but allows interception of the Write method. All middleware that don't need to read
the whole response body should use Peek or provide their own ResponseWriter wrapper (then do not
forget to implement the Contexter interface).

Finally EscapeHTML provides a response writer wrapper that allows on the fly
html escaping of the bytes written to the wrapped response writer.

FAQ

1. Why is the recommended way to use the Contexter interface to make a type assertion from
the ResponseWriter to the Contexter interface without error handling?

Answer: You should not create middleware stacks on the fly but beforehand. And you should run tests.
Then the assertion will blow up and that is correct because there is no reasonable way to handle such an
error. It means that either you have no context in your stack or you inject the context to late or
the context does not handle the kind of type the middleware expects. In each case you should fix it and
the panic forces you to do. Use the DEBUG flag to see what's going on.

2. What happens if another response writer wrapper wraps my context or another context?

Answer: You should only have one context per application and inject it as first wrapper into your
middleware stack. All context specific data belongs there.

That also has the benefit that you can be sure to be able to access all of your context data
everywhere inside your stack.

Never should a context wrap another one. There also should be no need for another context wrapping
ResponseWriter because every type can be saved by a Contexter.

All response writer wrappers of this package fulfill the Contexter interface and every response
writer you use should.

3. Why isn't there any default context object? Why do I have to write it on my own?

Answer: To write your own context and context injection has several benefits:

  - no magic or reflection necessary
  - all your context data is managed at one place
  - context data may be generated/updated based on other context data
  - your context management is independant from the middleware
  - you have no dependency issues (apart from supporting types the middleware packages expect)

4. Why is the context data managed by type and by string keys?

Answer: Type based context allows a very simple implementation that namespaces across packages
and needs no memory allocation and is, well, type safe. If you need to store multiple context
data of the same type, simple defined an alias type for each key.

5. What about spawning goroutines / how does it relate to code.google.com/p/go.net/context?

If you need your request to be handled by different goroutines and middlewares you might use
your Contexter to store and provide access to a code.google.com/p/go.net/context Context just
like any other context data. The cool thing is that you can write you middleware like before
and extract your Context everywhere you need it. And if you have middleware in between
that does not know about about Context or other context data you don't have build wrappers around that
have Context as parameter. Instead the ResponseWriter that happens to provide you a Context will be passed down the
chain bythe middleware.

"At Google, we require that Go programmers pass a Context parameter as the first argument to every
function on the call path between incoming and outgoing requests."

(Sameer Ajmani, http://blog.golang.org/context)

This is not neccessary anymore. And it is not neccessary for any type of contextual data.

*/
package wrap
