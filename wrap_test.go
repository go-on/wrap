// +build go1.1

package wrap

import (
	"net/http"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := map[string]http.Handler{
		"abc": New(
			write("a"),
			write("b"),
			HandlerFunc(write("c").ServeHTTP),
		),
		"ab": New(
			WrapperFunc(write("a").Wrap),
			writeStop("b"),
			write("c"),
		),
	}

	for body, h := range tests {
		rec, req := newTestRequest("GET", "/")
		h.ServeHTTP(rec, req)
		assertResponse(t, rec, body, 200)
	}
}
