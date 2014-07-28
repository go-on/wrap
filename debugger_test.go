package wrap_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-on/wrap"

	"github.com/go-on/wrap-contrib/wraps"
)

type write string

func (w write) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
}

func (w write) ServeHandle(next http.Handler, wr http.ResponseWriter, req *http.Request) {
	w.ServeHTTP(wr, req)
	next.ServeHTTP(wr, req)
}

func (w write) Wrap(next http.Handler) http.Handler {
	return wrap.ServeHandle(w, next)
}

func x(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("x"))
}

func TestDebug1(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	wrap.NewLogDebugger(&buf, log.Lshortfile)
	wrap.DEBUG = true

	wrap.New(
		wraps.Before(write("one")),
		wrap.Handler(write("two")),
	).ServeHTTP(rec, req)

	wrap.DEBUG = false

	splitted := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(splitted) != 3 {
		t.Errorf("expected 3 lines, got %d", len(splitted))
	}

	prefix := "[go-on/wrap debugger]"

	if !strings.HasPrefix(splitted[0], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[0], prefix)
	}

	suffix := "GET / wraps.BeforeFunc as Wrapper"
	if !strings.HasSuffix(splitted[0], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[0], suffix)
	}

	if !strings.HasPrefix(splitted[1], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[1], prefix)
	}

	suffix = "GET / wrap.ServeHandlerFunc as Wrapper"
	if !strings.HasSuffix(splitted[1], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[1], suffix)
	}

	if !strings.HasPrefix(splitted[2], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[2], prefix)
	}

	suffix = "GET / wrap_test.write as http.Handler"
	if !strings.HasSuffix(splitted[2], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[2], suffix)
	}

}

func TestDebug2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	wrap.NewLogDebugger(&buf, log.Lshortfile)
	wrap.DEBUG = true

	wrap.New(
		wraps.Before(write("one")),
		wrap.HandlerFunc(x),
	).ServeHTTP(rec, req)

	wrap.DEBUG = false

	splitted := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(splitted) != 3 {
		t.Errorf("expected 3 lines, got %d", len(splitted))
	}

	prefix := "[go-on/wrap debugger]"

	if !strings.HasPrefix(splitted[0], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[0], prefix)
	}

	suffix := "GET / wraps.BeforeFunc as Wrapper"
	if !strings.HasSuffix(splitted[0], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[0], suffix)
	}

	if !strings.HasPrefix(splitted[1], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[1], prefix)
	}

	suffix = "GET / wrap.ServeHandlerFunc as Wrapper"
	if !strings.HasSuffix(splitted[1], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[1], suffix)
	}

	if !strings.HasPrefix(splitted[2], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[2], prefix)
	}

	suffix = "GET / func(http.ResponseWriter, *http.Request) as http.HandlerFunc"
	if !strings.HasSuffix(splitted[2], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[2], suffix)
	}

}
