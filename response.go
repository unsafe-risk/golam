package golam

import (
	"errors"
	"net/http"
)

func NewResponse(adapter ResponseAdapter) *Response {
	return &Response{
		adapter:   adapter,
		Committed: false,
	}
}

type Response struct {
	adapter   ResponseAdapter
	Committed bool
}

func (r *Response) Header() http.Header {
	return r.adapter.Header()
}

func (r *Response) Write(b []byte) (int, error) {
	return r.adapter.Write(b)
}

func (r *Response) WriteHeader(statusCode int) {
	r.adapter.WriteHeader(statusCode)
}

func (r *Response) SetCookie(cookie *http.Cookie) {
	r.adapter.SetCookie(cookie)
}

func (r *Response) ResponseWriter() http.ResponseWriter {
	return r.adapter
}

func (r *Response) Commit() error {
	if r.Committed {
		return errors.New("already committed")
	}

	defer func() {
		r.Committed = true
	}()
	return r.adapter.Commit()
}

type ResponseAdapter interface {
	http.ResponseWriter
	SetCookie(cookie *http.Cookie)
	Commit() error
}
