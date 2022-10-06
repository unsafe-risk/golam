package golam

import (
	"bytes"
	"net/http"
)

var _ http.ResponseWriter = (*Response)(nil)

type Response struct {
	header   http.Header
	cookies  []*http.Cookie
	isBinary bool

	StatusCode int
	BodyBuffer bytes.Buffer
	Committed  bool
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Write(b []byte) (int, error) {
	return r.BodyBuffer.Write(b)
}

func (r *Response) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}

func (r *Response) SetCookie(cookie *http.Cookie) {
	r.cookies = append(r.cookies, cookie)
}
