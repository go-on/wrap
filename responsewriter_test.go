package wrap

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/go-on/wrap-contrib.v2/helper"
)

/*
type write string

func (w write) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, string(w))
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

func writeHeader(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("a", "b")
}

func writeCode(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(407)
}

func writeCodeCreated(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(201)
}

func TestPeek(t *testing.T) {

	ck := NewPeek(nil, nil)

	writeHeader(ck, nil)

	if !ck.HasChanged() {
		t.Errorf("should have changed, but has not")
	}

	if ck.Header().Get("a") != "b" {
		t.Errorf("header a should be b, but is: %#v", ck.Header().Get("a"))
	}
}

func TestPeekFlushMissing1(t *testing.T) {
	ckB := NewPeek(nil, nil)
	ck := NewPeek(ckB, nil)

	writeHeader(ck, nil)
	writeCode(ck, nil)

	ck.FlushMissing()

	if !ckB.HasChanged() {
		t.Errorf("should have changed, but has not")
	}

	if ckB.Header().Get("a") != "b" {
		t.Errorf("header a should be b, but is: %#v", ckB.Header().Get("a"))
	}

	if ckB.Code != 407 {
		t.Errorf("code should be 407, but is: %d", ckB.Code)
	}
}

func TestPeekFlushMissing2(t *testing.T) {
	ckB := NewPeek(nil, nil)
	ck := NewPeek(ckB, nil)

	writeCode(ck, nil)
	ck.FlushCode()

	if ckB.Code != 407 {
		t.Errorf("code should be 407, but is: %d", ckB.Code)
	}

	ckB.changed = false
	ck.FlushMissing()

	if ckB.HasChanged() {
		t.Errorf("should not have changed, but has")
	}

	if ckB.Code != 407 {
		t.Errorf("code should be 407, but is: %d", ckB.Code)
	}
}

func TestCheckResponseCode(t *testing.T) {

	ck := NewPeek(nil, nil)

	writeCode(ck, nil)

	if !ck.HasChanged() {
		t.Errorf("should have changed, but has not")
	}

	if ck.Code != 407 {
		t.Errorf("code should be 407, but is: %v", ck.Code)
	}
}

func TestCheckResponseIsOk1(t *testing.T) {
	ck := NewPeek(nil, nil)
	NoOp(ck, nil)

	if !ck.IsOk() {
		t.Errorf("should be ok when doing nothing, but is not")
	}
}

func TestCheckResponseIsOk2(t *testing.T) {
	ck := NewPeek(nil, nil)
	writeCodeCreated(ck, nil)

	if !ck.IsOk() {
		t.Errorf("should be ok with code 201, but is not")
	}
}

func TestCheckResponseIsOk3(t *testing.T) {
	ck := NewPeek(nil, nil)
	writeCode(ck, nil)

	if ck.IsOk() {
		t.Errorf("should not be ok with code 407, but is")
	}
}

func TestFlushCode(t *testing.T) {
	ckB := NewPeek(nil, nil)
	ckA := NewPeek(ckB, nil)

	writeCode(ckA, nil)

	ckA.FlushCode()

	if !ckB.HasChanged() {
		t.Errorf("should have changed, but has not")
	}

	if ckB.Code != 407 {
		t.Errorf("code should be 407, but is: %v", ckB.Code)
	}

	// don't write a second time
	ckB.Code = 0
	ckA.FlushCode()

	if ckB.Code != 0 {
		t.Errorf("code should be 0, but is: %v", ckB.Code)
	}
}

func TestCheckFlushHeaders1(t *testing.T) {
	ckB := NewPeek(nil, nil)
	ckA := NewPeek(ckB, nil)

	writeHeader(ckA, nil)

	ckA.FlushHeaders()

	if !ckB.HasChanged() {
		t.Errorf("should have changed, but has not")
	}

	if ckB.Header().Get("a") != "b" {
		t.Errorf("header a should be b, but is: %#v", ckB.Header().Get("a"))
	}

	// don't write a second time
	ckB.Header().Set("a", "")
	ckA.FlushHeaders()
	if ckB.Header().Get("a") != "" {
		t.Errorf(`header a should be "", but is: %#v`, ckB.Header().Get("a"))
	}
}

/*
func TestCheckFlushHeaders2(t *testing.T) {
	ckB := NewPeek(nil, nil)
	ckA := NewPeek(ckB, nil)

	writeHeader(ckA, nil)
	ckA.FlushCode()

	defer func() {
		if recover() == nil {
			t.Errorf("should panic if code is written before headers, but does not")
		}
	}()

	ckA.FlushHeaders()

}
*/

func TestCheckReset(t *testing.T) {

	ck := NewPeek(nil, nil)

	writeHeader(ck, nil)
	writeCode(ck, nil)
	ck.Reset()

	if ck.HasChanged() {
		t.Errorf("should not have changed, but has")
	}

	if ck.Header().Get("a") != "" {
		t.Errorf(`header a should be "", but is: %#v`, ck.Header().Get("a"))
	}
}

/*
type write string

func (w write) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, write)
}
*/

func TestCheckWrite1(t *testing.T) {
	rec := httptest.NewRecorder()
	ck := NewPeek(rec, nil)
	write("hiho").ServeHTTP(ck, nil)

	if rec.Body.String() != "hiho" {
		t.Errorf(`body a should be "hiho", but is: %#v`, rec.Body.String())
	}
}

func TestCheckWrite2(t *testing.T) {
	rec := httptest.NewRecorder()
	ck := NewPeek(rec, func(c *Peek) bool {
		return true
	})
	write("hiho").ServeHTTP(ck, nil)

	if rec.Body.String() != "hiho" {
		t.Errorf(`body a should be "hiho", but is: %#v`, rec.Body.String())
	}
}

func TestCheckWrite3(t *testing.T) {
	rec := httptest.NewRecorder()
	ck := NewPeek(rec, func(c *Peek) bool {
		return false
	})
	write("hiho").ServeHTTP(ck, nil)

	if rec.Body.String() != "" {
		t.Errorf(`body a should be "", but is: %#v`, rec.Body.String())
	}
}

type ctx struct {
	http.ResponseWriter
	context string
}

func (c *ctx) SetContext(context interface{}) {
	c.context = context.(string)
}

func (c *ctx) Context(context interface{}) bool {
	//*context = *c.context
	ctx := context.(*string)
	*ctx = c.context
	return true
}

func contextSetter(rw http.ResponseWriter, req *http.Request) {
	var hello string
	rw.(Contexter).Context(&hello)
	rw.(Contexter).SetContext(hello + "world")
}

func TestCheckContext(t *testing.T) {
	c := &ctx{context: "hello "}
	ck := NewPeek(c, nil)

	contextSetter(ck, nil)

	if c.context != "hello world" {
		t.Errorf(`context should be "hello world", but is: %#v`, c.context)
	}
}

func TestEscapeHTML(t *testing.T) {
	c := &ctx{context: "hello "}
	esc := &EscapeHTML{c}

	contextSetter(esc, nil)

	if c.context != "hello world" {
		t.Errorf(`context should be "hello world", but is: %#v`, c.context)
	}
}

func TestBufferContext(t *testing.T) {
	c := &ctx{context: "hello "}
	buf := NewBuffer(c)
	// esc := &Buffer{c}

	contextSetter(buf, nil)

	if c.context != "hello world" {
		t.Errorf(`context should be "hello world", but is: %#v`, c.context)
	}
}

func TestResponseBufferWriteTo(t *testing.T) {
	rec, req := helper.NewTestRequest("GET", "/")
	buf := NewBuffer(rec)
	write("hi").ServeHTTP(buf, req)
	buf.FlushAll()
	err := helper.AssertResponse(rec, "hi", 200)
	if err != nil {
		t.Error(err)
	}
}

func TestResponseBufferReset(t *testing.T) {
	buf := NewBuffer(nil)
	_, req := helper.NewTestRequest("GET", "/")
	write("hi").ServeHTTP(buf, req)

	buf.Reset()
	if buf.Code != 0 {
		t.Errorf("wrong code, expecting 0, got %d", buf.Code)
	}

	if len(buf.header) != 0 {
		t.Errorf("header is not empty")
	}

	if buf.BodyString() != "" {
		t.Errorf("body is not empty")
	}

	if buf.HasChanged() {
		t.Errorf("HasChanged should return false")
	}

}

func TestResponseBufferWriteToStatus(t *testing.T) {
	rec, req := helper.NewTestRequest("GET", "/")
	buf := NewBuffer(rec)
	http.NotFoundHandler().ServeHTTP(buf, req)
	buf.FlushAll()
	err := helper.AssertResponse(rec, "404 page not found", 404)
	if err != nil {
		t.Error(err)
	}

	if buf.IsOk() {
		t.Error("buf is ok, but should be not")
	}
}

func TestResponseBufferChanged(t *testing.T) {
	buf2 := NewBuffer(nil)
	buf1 := NewBuffer(buf2)
	_, req := helper.NewTestRequest("GET", "/")
	write("hi").ServeHTTP(buf1, req)
	buf1.FlushAll()

	if buf1.BodyString() != "hi" {
		t.Errorf("body string of buf1 should be \"hi\" but is :%#v", buf1.BodyString())
	}

	if buf2.BodyString() != "hi" {
		t.Errorf("body string of buf2 should be \"hi\" but is :%#v", buf2.BodyString())
	}

	if string(buf1.Body()) != "hi" {
		t.Errorf("body of buf1 should be \"hi\" but is :%#v", string(buf1.Body()))
	}

	if string(buf2.Body()) != "hi" {
		t.Errorf("body of buf2 should be \"hi\" but is :%#v", string(buf2.Body()))
	}

	if buf1.Code != 0 {
		t.Errorf("Code of buf1 should be %d but is :%d", 0, buf1.Code)
	}

	if buf2.Code != 0 {
		t.Errorf("Code of buf2 should be %d but is :%d", 0, buf2.Code)
	}

	ctype1 := buf1.Header().Get("Content-Type")
	if ctype1 != "text/plain" {
		t.Errorf("Content-Type of buf1 should be %#v but is: %#v", "text/plain", ctype1)
	}

	ctype2 := buf2.Header().Get("Content-Type")
	if ctype2 != "text/plain" {
		t.Errorf("Content-Type of buf2 should be %#v but is: %#v", "text/plain", ctype2)
	}

	if !buf1.HasChanged() {
		t.Error("buf1 should be changed, but is not")
	}

	if !buf2.HasChanged() {
		t.Error("buf2 should be changed, but is not")
	}

	if !buf1.IsOk() {
		t.Error("buf1 should be ok, but is not")
	}

	if !buf2.IsOk() {
		t.Error("buf2 should be ok, but is not")
	}
}

func TestResponseBufferNotChanged(t *testing.T) {
	buf1 := NewBuffer(nil)
	buf2 := NewBuffer(nil)
	_, req := helper.NewTestRequest("GET", "/")
	NoOp(buf1, req)
	buf1.FlushAll()

	if buf1.HasChanged() {
		t.Error("buf1 is changed, but should not be")
	}

	if buf2.HasChanged() {
		t.Error("buf2 is changed, but should not be")
	}

	if !buf1.IsOk() {
		t.Error("buf1 should be ok, but is not")
	}

	if !buf2.IsOk() {
		t.Error("buf2 should be ok, but is not")
	}
}

func TestResponseBufferStatusCreate(t *testing.T) {
	writeCreate := func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(201)
	}

	buf := NewBuffer(nil)
	_, req := helper.NewTestRequest("GET", "/")
	writeCreate(buf, req)

	if buf.Code != 201 {
		t.Errorf("Code of buf should be %d but is :%d", 201, buf.Code)
	}

	if !buf.IsOk() {
		t.Error("buf should be ok, but is not")
	}
}

func TestEscapeHTMLResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	esc := &EscapeHTML{rec}
	esc.Write([]byte(`abc<d>"e'f&g`))

	expected := `abc&lt;d&gt;&#34;e&#39;f&amp;g`
	got := rec.Body.String()
	if expected != got {
		t.Errorf("expected: %#v, got %#v", expected, got)
	}
}

type flushingRW struct {
	http.ResponseWriter
	flushed bool
}

func (f *flushingRW) Flush() {
	f.flushed = true
}

func TestReclaimResponseWriter(t *testing.T) {
	rw1 := &flushingRW{}
	rw2 := &context{ResponseWriter: rw1}

	res1 := ReclaimResponseWriter(rw1)

	if _, ok := res1.(*flushingRW); !ok {
		t.Errorf("must return the given response writer if that is not a Contexter, but does not")
	}

	res2 := ReclaimResponseWriter(rw2)

	if _, ok := res2.(*flushingRW); !ok {
		t.Errorf("must return the wrapped response writer if given a Contexter, but does not")
	}
}

func TestFlush(t *testing.T) {
	rw1 := &flushingRW{}

	ok := Flush(rw1)

	if !rw1.flushed {
		t.Errorf("did not flush a http.Flusher")
	}

	if !ok {
		t.Errorf("did not report the flush of a http.Flusher")
	}

	rw1 = &flushingRW{}
	rw2 := &context{ResponseWriter: rw1}

	ok = Flush(rw2)

	if !rw1.flushed {
		t.Errorf("did not flush a http.Flusher wrapped inside a Contexter")
	}

	if !ok {
		t.Errorf("did not report the flush of a http.Flusher wrapped inside a Contexter")
	}

	rw1 = &flushingRW{}
	rw2 = &context{ResponseWriter: &context{ResponseWriter: rw1}}

	ok = Flush(rw2)

	// not allowed nesting of Contexter (supporting it would be inefficient and lead to complicated and confusing coding style)
	if rw1.flushed {
		t.Errorf("must not flush a http.Flusher wrapped inside a Contexter that is inside another Contexter")
	}

	if ok {
		t.Errorf("must not report the flush of a http.Flusher wrapped inside a Contexter that is inside another Contexter")
	}

}

type hijackerRW struct {
	http.ResponseWriter
	hijacked bool
}

func (h *hijackerRW) Hijack() (c net.Conn, brw *bufio.ReadWriter, err error) {
	h.hijacked = true
	return
}

func TestHijack(t *testing.T) {
	rw1 := &hijackerRW{}

	_, _, _, ok := Hijack(rw1)

	if !rw1.hijacked {
		t.Errorf("did not hijack a http.Hijacker")
	}

	if !ok {
		t.Errorf("did not report the hijack of a http.Hijacker")
	}

	rw1 = &hijackerRW{}
	rw2 := &context{ResponseWriter: rw1}

	_, _, _, ok = Hijack(rw2)

	if !rw1.hijacked {
		t.Errorf("did not hijack a http.Hijacker wrapped inside a Contexter")
	}

	if !ok {
		t.Errorf("did not report the hijack of a http.Hijacker wrapped inside a Contexter")
	}

	rw1 = &hijackerRW{}
	rw2 = &context{ResponseWriter: &context{ResponseWriter: rw1}}

	_, _, _, ok = Hijack(rw2)

	// not allowed nesting of Contexter (supporting it would be inefficient and lead to complicated and confusing coding style)
	if rw1.hijacked {
		t.Errorf("must not hijack a http.Hijacker wrapped inside a Contexter that is inside another Contexter")
	}

	if ok {
		t.Errorf("must not report the hijack of a http.Hijacker wrapped inside a Contexter that is inside another Contexter")
	}

}

type closeNotifyRW struct {
	http.ResponseWriter
	closeNotified bool
}

func (c *closeNotifyRW) CloseNotify() (ch <-chan bool) {
	c.closeNotified = true
	return
}

func TestCloseNotifier(t *testing.T) {
	rw1 := &closeNotifyRW{}

	_, ok := CloseNotify(rw1)

	if !rw1.closeNotified {
		t.Errorf("did not closenotify a http.CloseNotifier")
	}

	if !ok {
		t.Errorf("did not report the closenotify of a http.CloseNotifier")
	}

	rw1 = &closeNotifyRW{}
	rw2 := &context{ResponseWriter: rw1}

	_, ok = CloseNotify(rw2)

	if !rw1.closeNotified {
		t.Errorf("did not closenotify a http.CloseNotifier wrapped inside a Contexter")
	}

	if !ok {
		t.Errorf("did not report the closenotify of a http.CloseNotifier wrapped inside a Contexter")
	}

	rw1 = &closeNotifyRW{}
	rw2 = &context{ResponseWriter: &context{ResponseWriter: rw1}}

	_, ok = CloseNotify(rw2)

	// not allowed nesting of Contexter (supporting it would be inefficient and lead to complicated and confusing coding style)
	if rw1.closeNotified {
		t.Errorf("must not closenotify a http.CloseNotifier wrapped inside a Contexter that is inside another Contexter")
	}

	if ok {
		t.Errorf("must not report the closenotify of a http.CloseNotifier wrapped inside a Contexter that is inside another Contexter")
	}

}
