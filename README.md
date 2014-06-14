wrap
====

Package wrap creates a fast and flexible middleware stack for http.Handlers.

[![Build Status](https://secure.travis-ci.org/go-on/wrap.png)](http://travis-ci.org/go-on/wrap) [![GoDoc](https://godoc.org/github.com/go-on/wrap?status.png)](http://godoc.org/github.com/go-on/wrap) [![Coverage Status](https://img.shields.io/coveralls/go-on/wrap.svg)](https://coveralls.io/r/go-on/wrap?branch=master)

Features

- small; core is only 13 LOC
- based on http.Handler interface; integrates fine with net/http
- middleware stacks are http.Handlers too and may be embedded
- low memory footprint
- fast

How does it work
----------------

A nice introduction into this library is [on my blog](http://metakeule.github.io/article/wrap-go-middlware-framework.html).

`wrap.New(w ...Wrapper)` creates a stack of middlewares. `Wrapper` is defined as

    type Wrapper interface {
        Wrap(next http.Handler) (previous http.Handler)
    }

Each wrapper wraps the the `http.Handler` that comes further down
the middleware stack and returns a `http.Handler` that handles the
request previously.

Example
-------

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/go-on/wrap"
)

type print string

func (p print) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
    fmt.Println(p)
}

// ServeHandle prints the string and calls the next handler in the chain
func (p print) ServeHandle(next http.Handler, wr http.ResponseWriter, req *http.Request) {
    fmt.Print(p)
    next.ServeHTTP(wr, req)
}

func main() {

    // creates a chain of Wrappers
    h := wrap.New(

        // uses print.ServeHandle
        wrap.ServeWrapper(print("ready...")),

        // uses print.ServeHandle
        wrap.ServeWrapper(print("steady...")),
        
        // uses print.ServeHTTP
        wrap.Handler(print("go!")),
    )

    r, _ := http.NewRequest("GET", "/", nil)
    h.ServeHTTP(nil, r)

    // Output:
    // ready...steady...go!
    //
}
```

Status
------
This package is considered feature complete, stable and ready for production.

It will not change (apart from documentation improvements), since all further
development will be middleware in the go-on/wrap-contrib repository.

Middleware
----------

more examples and middleware and can be found at [github.com/go-on/wrap-contrib](https://github.com/go-on/wrap-contrib) 

Router
------

A router that is also tested but may change, can be found at [github.com/go-on/router](https://github.com/go-on/router)

Benchmarks
----------

    // The overhead of n writes to http.ResponseWriter via n wrappers
    // vs n writes in a loop within a single http.Handler

    BenchmarkServing2Simple     5000000   718 ns/op 1.00x
    BenchmarkServing2Wrappers   2000000   824 ns/op 1.14x

    BenchmarkServing50Simple     100000 17466 ns/op 1.00x
    BenchmarkServing50Wrappers   100000 23984 ns/op 1.37x

    BenchmarkServing100Simple     50000 33686 ns/op 1.00x
    BenchmarkServing100Wrappers   50000 46676 ns/op 1.39x


Credits
-------

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.

