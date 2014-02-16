wrap
====

Package wrap creates a fast and flexible middleware stack for http.Handlers.

[![Build Status](https://secure.travis-ci.org/go-on/wrap.png)](http://travis-ci.org/go-on/wrap)

[![GoDoc](https://godoc.org/github.com/go-on/wrap?status.png)](http://godoc.org/github.com/go-on/wrap)

Features

- small, the core is only 28 LOC including comments
- integrates fine with net/http
- middleware stacks may be used in middleware because they are http.Handlers
- low memory footprint
- fast

Status
------
100% test coverage.
This package is considered complete, stable and ready for production.

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


Examples
--------

more examples, middleware and router can be found [here](https://github.com/go-on/wrap-contrib) 

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

func (p print) ServeHandle(inner http.Handler, wr http.ResponseWriter, req *http.Request) {
    fmt.Print(p)
    inner.ServeHTTP(wr, req)
}

func main() {
    h := wrap.New(
        wrap.ServeWrapper(print("ready...")),
        wrap.ServeWrapper(print("steady...")),
        wrap.Handler(print("go!")),
    )
    r, _ := http.NewRequest("GET", "/", nil)
    h.ServeHTTP(nil, r)

    // Output:
    // ready...steady...go!
    //
}
```


Credits
-------

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.

