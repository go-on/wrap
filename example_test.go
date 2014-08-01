package wrap

import (
	"fmt"
	"net/http"
)

/*
This example illustrates 3 ways to write and use middleware.

1. use a http.Handler as wrapper (needs adapter in the middleware stack)
2. use a ServeHTTPNext method (nice to write but needs adapter in the middleware stack)
3. use a Wrapper (needs no adapter)

For sharing context, look at example_context_test.go.
*/

type print1 string

func (p print1) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(p)
}

func (p print1) ServeHTTPNext(next http.Handler, wr http.ResponseWriter, req *http.Request) {
	fmt.Print(p)
	next.ServeHTTP(wr, req)
}

type print2 string

func (p print2) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		fmt.Print(p)
		next.ServeHTTP(wr, req)
	}
	return f
}

func Example() {
	h := New(
		NextHandler(print1("ready...")), // make use of ServeHTTPNext method
		print2("steady..."),             // print2 directly fulfills Wrapper interface
		Handler(print1("go!")),          // make use of ServeHTTP method, this stopps the chain
		// if there should be a handler after this, you will need the Before wrapper from go-on/wrap-contrib/wraps
	)
	r, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(nil, r)

	// Output:
	// ready...steady...go!
	//
}
