package wrap

import (
	"fmt"
	"net/http"
	. "launchpad.net/gocheck"
)

func mkRequestResponse() (w http.ResponseWriter, r *http.Request) {
	r, _ = http.NewRequest("GET", "/", nil)
	w = noHTTPWriter{}
	return
}

func mkWrap(num int) http.Handler {
	wrappers := make([]Wrapper, num)

	for i := 0; i < num; i++ {
		wrappers[i] = write("")
	}
	return New(wrappers...)
}

type times int

func (w times) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	n := int(w)
	wri := write("")
	for i := 0; i < n; i++ {
		fmt.Fprint(wr, string(wri))
	}
}

func (s *wrapSuite) BenchmarkWrapping(c *C) {
	c.StopTimer()
	wrappers := make([]Wrapper, c.N)

	for i := 0; i < c.N; i++ {
		wrappers[i] = write("")
	}
	c.StartTimer()
	New(wrappers...)
}

func (s *wrapSuite) BenchmarkServing100Simple(c *C) {
	c.StopTimer()
	h := times(100)
	wr, req := mkRequestResponse()
	c.StartTimer()

	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func (s *wrapSuite) BenchmarkServing100Wrappers(c *C) {
	c.StopTimer()
	h := mkWrap(100)
	wr, req := mkRequestResponse()
	c.StartTimer()
	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func (s *wrapSuite) BenchmarkServing50Wrappers(c *C) {
	c.StopTimer()
	h := mkWrap(50)
	wr, req := mkRequestResponse()
	c.StartTimer()
	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func (s *wrapSuite) BenchmarkServing50Simple(c *C) {
	c.StopTimer()
	h := times(50)
	wr, req := mkRequestResponse()
	c.StartTimer()
	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func (s *wrapSuite) BenchmarkServing2Wrappers(c *C) {
	c.StopTimer()
	h := mkWrap(2)
	wr, req := mkRequestResponse()
	c.StartTimer()
	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func (s *wrapSuite) BenchmarkServing2Simple(c *C) {
	c.StopTimer()
	h := times(2)
	wr, req := mkRequestResponse()
	c.StartTimer()
	for i := 0; i < c.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

/*
func (w *wrapSuite) BenchmarkTestWrap(c *C) {
	h := http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprintln(rw, "hiho")
		},
	)

	r := New(h)

	// fmt.Println("creating wrapper")
	for i := 0; i < 2000; i++ {
		r.Wrap(mw(fmt.Sprintf("hu%d", i)))
	}

	// fmt.Println("serving wrappers")
	for i := 0; i < 20000; i++ {
		rw, req := newTestRequest("GET", "/xyz")
		r.ServeHTTP(rw, req)
		c.Assert(rw.Code, Equals, 200)
		assertResponse(c, rw, "hiho", 200)
	}

	// fmt.Println("served")

}

func (w *wrapSuite) BenchmarkTestNested(c *C) {
	h := http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprintln(rw, "hiho")
		},
	)

	r := New(h)

	var final Racker

	// fmt.Println("creating nested")
	for i := 0; i < 1000; i++ {
		final = New(r)
		r = final
	}
	// fmt.Println("serving nested")
	// serving nested racker seems slow
	for i := 0; i < 20000; i++ {
		rw, req := newTestRequest("GET", "/xyz")
		final.ServeHTTP(rw, req)
		c.Assert(rw.Code, Equals, 200)
		assertResponse(c, rw, "hiho", 200)
	}
	// fmt.Println("served")
}

func (w *wrapSuite) TestWrapFunc(c *C) {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "huho")
	}

	wr := func(inner http.Handler, rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "before-")
		inner.ServeHTTP(rw, req)
	}

	r := NewFunc(fn, ServeHandlerFunc(wr))

	rw, req := newTestRequest("GET", "/xyz")
	r.ServeHTTP(rw, req)
	c.Assert(rw.Code, Equals, 200)
	assertResponse(c, rw, "before-huho", 200)
}

type beforestring string

func (b beforestring) ServeHandle(inner http.Handler, rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, string(b))
	inner.ServeHTTP(rw, req)
}

func (w *wrapSuite) TestServeWrapper(c *C) {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "zwei")
	}

	r := NewFunc(fn, ServeWrapper(beforestring("eins-")))

	rw, req := newTestRequest("GET", "/xyz")
	r.ServeHTTP(rw, req)
	c.Assert(rw.Code, Equals, 200)
	assertResponse(c, rw, "eins-zwei", 200)
}
*/
