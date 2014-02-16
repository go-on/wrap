package wrap

import (
	"fmt"
	"net/http"
)

type print string

func (p print) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(p)
}

func (p print) ServeHandle(inner http.Handler, wr http.ResponseWriter, req *http.Request) {
	fmt.Print(p)
	inner.ServeHTTP(wr, req)
}

func Example() {
	h := New(
		ServeWrapper(print("ready...")),
		ServeWrapper(print("steady...")),
		Handler(print("go!")),
	)
	r, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(nil, r)

	// Output:
	// ready...steady...go!
	//
}
