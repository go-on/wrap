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
