package wrap

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// "gopkg.in/go-on/wrap.v2"
)

func TestDebug(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	NewLogDebugger(&buf, log.Lshortfile)
	SetDebug()

	New(
		NextHandler(write("one")),
		writeStop("two"),
	).ServeHTTP(rec, req)

	DEBUG = false

	splitted := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(splitted) != 4 {
		t.Errorf("expected 4 lines, got %d", len(splitted))
	}

	prefix := "[go-on/wrap debugger]"

	if !strings.HasPrefix(splitted[0], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[0], prefix)
	}

	suffix := "GET / wrap.NextHandlerFunc as Wrapper"
	if !strings.HasSuffix(splitted[0], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[0], suffix)
	}

	if !strings.HasPrefix(splitted[1], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[1], prefix)
	}

	suffix = "GET / wrap.NextHandlerFunc as NextHandlerFunc"
	if !strings.HasSuffix(splitted[1], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[1], suffix)
	}

	if !strings.HasPrefix(splitted[2], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[2], prefix)
	}

	suffix = "GET / wrap.write as NextHandler"
	if !strings.HasSuffix(splitted[2], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[2], suffix)
	}

	if !strings.HasPrefix(splitted[3], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[3], prefix)
	}

	suffix = "GET / wrap.writeStop as Wrapper"
	if !strings.HasSuffix(splitted[3], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[3], suffix)
	}

}

/*
type write string

func (w write) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
}

func (w write) ServeHandle(next http.Handler, wr http.ResponseWriter, req *http.Request) {
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
*/

func x(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("x"))
}

func TestDebug1(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	NewLogDebugger(&buf, log.Lshortfile)
	DEBUG = true

	New(
		// wraps.Before(write("one")),
		write("one"),
		Handler(write("two")),
	).ServeHTTP(rec, req)

	DEBUG = false

	splitted := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(splitted) != 4 {
		t.Errorf("expected 4 lines, got %d", len(splitted))
	}

	prefix := "[go-on/wrap debugger]"

	if !strings.HasPrefix(splitted[0], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[0], prefix)
	}
	suffix := "GET / wrap.write as Wrapper"
	if !strings.HasSuffix(splitted[0], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[0], suffix)
	}

	if !strings.HasPrefix(splitted[1], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[1], prefix)
	}

	suffix = "GET / wrap.NextHandlerFunc as Wrapper"
	if !strings.HasSuffix(splitted[1], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[1], suffix)
	}

	if !strings.HasPrefix(splitted[2], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[2], prefix)
	}

	suffix = "GET / wrap.NextHandlerFunc as NextHandlerFunc"
	if !strings.HasSuffix(splitted[2], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[2], suffix)
	}

	if !strings.HasPrefix(splitted[3], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[3], prefix)
	}

	suffix = "GET / wrap.write as http.Handler"
	if !strings.HasSuffix(splitted[3], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[3], suffix)
	}

}

func TestDebug2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	NewLogDebugger(&buf, log.Lshortfile)
	DEBUG = true

	New(
		//wraps.Before(write("one")),
		write("one"),
		HandlerFunc(x),
	).ServeHTTP(rec, req)

	DEBUG = false

	splitted := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(splitted) != 4 {
		t.Errorf("expected 4 lines, got %d", len(splitted))
	}

	prefix := "[go-on/wrap debugger]"

	if !strings.HasPrefix(splitted[0], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[0], prefix)
	}
	suffix := "GET / wrap.write as Wrapper"
	if !strings.HasSuffix(splitted[0], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[0], suffix)
	}

	if !strings.HasPrefix(splitted[1], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[1], prefix)
	}

	suffix = "GET / wrap.NextHandlerFunc as Wrapper"
	if !strings.HasSuffix(splitted[1], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[1], suffix)
	}

	if !strings.HasPrefix(splitted[2], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[2], prefix)
	}

	suffix = "GET / wrap.NextHandlerFunc as NextHandlerFunc"
	if !strings.HasSuffix(splitted[2], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[2], suffix)
	}

	if !strings.HasPrefix(splitted[3], prefix) {
		t.Errorf("%#v should start with %#v but does not", splitted[3], prefix)
	}

	suffix = "GET / func(http.ResponseWriter, *http.Request) as http.HandlerFunc"
	if !strings.HasSuffix(splitted[3], suffix) {
		t.Errorf("%#v should end with %#v but does not", splitted[3], suffix)
	}
}
