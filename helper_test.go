package wrap

// Make a testing request

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		fmt.Printf("could not make request %s (%s): %s\n", path, method, err.Error())
	}
	recorder := httptest.NewRecorder()

	return recorder, request
}

func assertResponse(t *testing.T, rec *httptest.ResponseRecorder, body string, code int) {
	trimmed := strings.TrimSpace(string(rec.Body.Bytes()))
	if trimmed != body {
		t.Errorf("body should be %#v but is %#v", body, trimmed)
	}

	if rec.Code != code {
		t.Errorf("status code should be %d but is %d", code, rec.Code)
	}
}

type noHTTPWriter struct{}

func (noHTTPWriter) Write([]byte) (i int, err error) {
	return
}

func (noHTTPWriter) Header() (h http.Header) {
	return
}

func (noHTTPWriter) WriteHeader(i int) {
}

type writeString string

func (w writeString) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
}

func (w writeString) ServeHTTPNext(next http.Handler, wr http.ResponseWriter, req *http.Request) {
	w.ServeHTTP(wr, req)
	next.ServeHTTP(wr, req)
}

func (w writeString) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(wr http.ResponseWriter, req *http.Request) {
		w.ServeHTTP(wr, req)
		next.ServeHTTP(wr, req)
	}
	return f
}

type write string

func (w write) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(wr, string(w))
}

func (w write) ServeHTTPNext(next http.Handler, wr http.ResponseWriter, req *http.Request) {
	w.ServeHTTP(wr, req)
	next.ServeHTTP(wr, req)
}

func (w write) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(wr http.ResponseWriter, req *http.Request) {
		w.ServeHTTP(wr, req)
		next.ServeHTTP(wr, req)
	}
	return f
}

type writeStop string

func (w writeStop) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
}

func (w writeStop) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(wr http.ResponseWriter, req *http.Request) {
		w.ServeHTTP(wr, req)
	}
	return f
}
