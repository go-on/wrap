// +build go1.1

package wrap

import (
	"net/http"
	"testing"
)

func TestWrapMethod(t *testing.T) {
	tests := map[string]http.Handler{
		"ABC": New(
			WrapperFunc(write("A").Wrap),
			ServeWrapper(write("B")),
			Handler(ServeHandleFunc(write("C").ServeHandle, NoOp)),
		),
		"not found": New(
			HandlerFunc(write("not found").ServeHTTP),
			write("c"),
		),
	}

	for body, h := range tests {
		rec, req := newTestRequest("GET", "/")
		h.ServeHTTP(rec, req)
		assertResponse(t, rec, body, 200)
	}
}
