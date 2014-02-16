package wrap

import (
	"fmt"
	"net/http"
	"testing"
)

type write string

func (w write) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
}

func (w write) ServeHandle(inner http.Handler, wr http.ResponseWriter, req *http.Request) {
	w.ServeHTTP(wr, req)
	inner.ServeHTTP(wr, req)
}

func (w write) Wrap(inner http.Handler) http.Handler {
	return ServeHandle(w, inner)
}

func TestWrap(t *testing.T) {
	tests := map[string]http.Handler{
		"abc": New(
			write("a"),
			write("b"),
			write("c"),
		),
		"ab": New(
			write("a"),
			Handler(write("b")),
			write("c"),
		),
	}

	for body, h := range tests {
		rec, req := newTestRequest("GET", "/")
		h.ServeHTTP(rec, req)
		assertResponse(t, rec, body, 200)
	}
}
