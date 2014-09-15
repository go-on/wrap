wrap
====

Package wrap creates a fast and flexible middleware stack for http.Handlers.

[![Build Status](https://drone.io/github.com/go-on/wrap/status.png)](https://drone.io/github.com/go-on/wrap/latest) [![Coverage Status](https://img.shields.io/coveralls/go-on/wrap.svg)](https://coveralls.io/r/go-on/wrap?branch=master)  [![GoDoc](https://godoc.org/github.com/go-on/wrap?status.png)](http://godoc.org/github.com/go-on/wrap) [![Total views](https://sourcegraph.com/api/repos/github.com/go-on/wrap/counters/views.png)](https://sourcegraph.com/github.com/go-on/wrap)


Features
--------

- **small**; core is only 13 LOC
- based on **http.Handler interface**; integrates fine with net/http
- middleware **stacks are http.Handlers too** and may be embedded
- has a solution for **per request context sharing**
- freely **mix middleware with and without context** (same interface)
- has **debugging helper**
- **low memory footprint**
- **fast**
- easy to create adapters / wrappers for 3rd party middleware

How does it work
----------------

`wrap.New(w ...Wrapper)` creates a stack of middlewares. `Wrapper` is defined as

    type Wrapper interface {
        Wrap(next http.Handler) (previous http.Handler)
    }

Each wrapper wraps the the `http.Handler` that comes further down
the middleware stack and returns a `http.Handler` that handles the
request previously.

Examples
--------

See `example_test.go` for a simple example without context and `example_context_test.go` for an example with context sharing.

Also look into the repository of blessed middleware [github.com/go-on/wrap-contrib/wraps](https://godoc.org/github.com/go-on/wrap-contrib/wraps).

Middleware
----------

more examples and middleware and can be found at [github.com/go-on/wrap-contrib](https://godoc.org/github.com/go-on/wrap-contrib/wraps) 

Router
------

A router that is also tested but may change, can be found at [github.com/go-on/router](https://godoc.org/github.com/go-on/router)

Benchmarks (go 1.3)
-------------------

    // The overhead of n writes to http.ResponseWriter via n wrappers
    // vs n writes in a loop within a single http.Handler

    BenchmarkServing2Simple     1000000 1067    ns/op   1,00x
    BenchmarkServing2Wrappers   1000000 1121    ns/op   1,05x

    BenchmarkServing50Simple    100000  26041   ns/op   1,00x
    BenchmarkServing50Wrappers  100000  27053   ns/op   1,04x

    BenchmarkServing100Simple   50000   52074   ns/op   1,00x
    BenchmarkServing100Wrappers 50000   53450   ns/op   1,03x



Credits
-------

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.
