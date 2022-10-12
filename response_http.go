package golam

import "net/http"

func NewResponseHTTPAdapter(writer http.ResponseWriter) ResponseAdapter {
	return &responseHTTPAdapter{
		writer: writer,
	}
}

var _ ResponseAdapter = (*responseHTTPAdapter)(nil)

type responseHTTPAdapter struct {
	writer http.ResponseWriter
}

func (w *responseHTTPAdapter) Header() http.Header {
	return w.writer.Header()
}

func (w *responseHTTPAdapter) Write(bytes []byte) (int, error) {
	return w.writer.Write(bytes)
}

func (w *responseHTTPAdapter) WriteHeader(statusCode int) {
	w.writer.WriteHeader(statusCode)
}

func (w *responseHTTPAdapter) SetCookie(cookie *http.Cookie) {
	http.SetCookie(w.writer, cookie)
}

func (w *responseHTTPAdapter) Commit() error {
	return nil
}
