package wrap

// panic("body written before code")

type BodyFlushedBeforeCode struct{}

func (e BodyFlushedBeforeCode) Error() string {
	return "body flushed before code"
}

type CodeFlushedBeforeHeaders struct{}

func (e CodeFlushedBeforeHeaders) Error() string {
	return "code flushed before headers"
}
