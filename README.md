wrap
====

Package wrap creates fast and flexible middleware for http.Handlers.

[![Build Status](https://secure.travis-ci.org/go-on/wrap.png)](http://travis-ci.org/go-on/wrap)

[![GoDoc](https://godoc.org/github.com/go-on/wrap?status.png)](http://godoc.org/github.com/go-on/wrap)

Status
------
100% test coverage.
This package is considered complete, stable and ready for production.

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


