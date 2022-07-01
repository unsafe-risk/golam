package golam

import (
	"bytes"
	"net/http"
)

var _ http.ResponseWriter = (*Resp)(nil)

type Resp struct {
	StatusCode int
	RespHeader http.Header
	Cookies    []*http.Cookie
	BodyBuffer bytes.Buffer
	Committed  bool
}

func (r *Resp) Header() http.Header {
	return r.RespHeader
}

func (r *Resp) Write(b []byte) (int, error) {
	return r.BodyBuffer.Write(b)
}

func (r *Resp) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}

func (r *Resp) SetCookie(cookie *http.Cookie) {
	r.Cookies = append(r.Cookies, cookie)
}
