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

To help sharing per request context there is a Context interface that must be implemented by
the ResponseWriter. That can easily be done be providing a middleware that injects a context
that wraps the current ResponseWriter and implements the Context interface.

An example can be found in the file example_context_test.go.

Furthermore this package provides some ResponseWriter wrappers that respect the possibility that
the inner ResponseWriter implements the Context interface and that help with development of middleware.

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
forget to implement the Context interface).

Finally EscapeHTML provides a response writer wrapper that allows on the fly
html escaping of the bytes written to the wrapped response writer.

*/
package wrap
