package wrap

import (
	"fmt"
	"net/http"
	"testing"
)

func mkRequestResponse() (w http.ResponseWriter, r *http.Request) {
	r, _ = http.NewRequest("GET", "/", nil)
	w = noHTTPWriter{}
	return
}

func mkWrap(num int) http.Handler {
	wrappers := make([]Wrapper, num)

	for i := 0; i < num; i++ {
		wrappers[i] = writeString("")
	}
	return New(wrappers...)
}

type times int

func (w times) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	n := int(w)
	wri := writeString("")
	for i := 0; i < n; i++ {
		fmt.Fprint(wr, string(wri))
	}
}

func BenchmarkWrapping(b *testing.B) {
	b.StopTimer()
	wrappers := make([]Wrapper, b.N)

	for i := 0; i < b.N; i++ {
		wrappers[i] = writeString("")
	}
	b.StartTimer()
	New(wrappers...)
}

func BenchmarkServing100Simple(b *testing.B) {
	b.StopTimer()
	h := times(100)
	wr, req := mkRequestResponse()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func BenchmarkServing100Wrappers(b *testing.B) {
	b.StopTimer()
	h := mkWrap(100)
	wr, req := mkRequestResponse()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func BenchmarkServing50Wrappers(b *testing.B) {
	b.StopTimer()
	h := mkWrap(50)
	wr, req := mkRequestResponse()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func BenchmarkServing50Simple(b *testing.B) {
	b.StopTimer()
	h := times(50)
	wr, req := mkRequestResponse()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func BenchmarkServing2Wrappers(b *testing.B) {
	b.StopTimer()
	h := mkWrap(2)
	wr, req := mkRequestResponse()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}

func BenchmarkServing2Simple(b *testing.B) {
	b.StopTimer()
	h := times(2)
	wr, req := mkRequestResponse()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(wr, req)
	}
}
