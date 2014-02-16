package wrap

// Make a testing request

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	. "launchpad.net/gocheck"
)

//
// gocheck: hook into "go test"
//
func Test(t *testing.T) { TestingT(t) }

func newTestRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		fmt.Printf("could not make request %s (%s): %s\n", path, method, err.Error())
	}
	recorder := httptest.NewRecorder()

	return recorder, request
}

func assertResponse(c *C, rec *httptest.ResponseRecorder, body string, code int) {
	c.Assert(strings.TrimSpace(string(rec.Body.Bytes())), Equals, body)
	c.Assert(rec.Code, Equals, code)
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
