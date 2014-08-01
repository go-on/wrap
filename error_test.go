package wrap

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
)

func errorMustBe(err interface{}, class interface{}) string {
	classTy := reflect.TypeOf(class)
	if err == nil {
		return fmt.Sprintf("error must be of type %s but is nil", classTy)
	}

	errTy := reflect.TypeOf(err)
	if errTy.String() != classTy.String() {
		return fmt.Sprintf("error must be of type %s but is of type %s", classTy, errTy)
	}
	return ""
}

func TestBodyFlushedBeforeCode(t *testing.T) {

	rec := httptest.NewRecorder()
	ckA := NewPeek(rec, func(rwp *Peek) bool {
		return true
	})

	write("hu").ServeHTTP(ckA, nil)

	writeCode(ckA, nil)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrBodyFlushedBeforeCode{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrBodyFlushedBeforeCode)
		_ = err.Error()
	}()

	ckA.FlushCode()

}

func TestBodyFlushedBeforeHeaders(t *testing.T) {

	rec := httptest.NewRecorder()
	ckA := NewPeek(rec, func(rwp *Peek) bool {
		return true
	})

	write("hu").ServeHTTP(ckA, nil)

	writeHeader(ckA, nil)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrBodyFlushedBeforeCode{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrBodyFlushedBeforeCode)
		_ = err.Error()
	}()

	ckA.FlushHeaders()

}

func TestCodeFlushedBeforeHeaders(t *testing.T) {

	rec := httptest.NewRecorder()
	ckA := NewPeek(rec, func(rwp *Peek) bool {
		return true
	})

	writeCode(ckA, nil)
	ckA.FlushCode()
	writeHeader(ckA, nil)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrCodeFlushedBeforeHeaders{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrCodeFlushedBeforeHeaders)
		_ = err.Error()
	}()

	ckA.FlushHeaders()

}
