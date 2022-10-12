package golam

import (
	"bytes"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"io/ioutil"
	"net/http"
)

var _ ResponseAdapter = (*responseLambdaAdapter)(nil)

func NewResponseLambdaAdapter(response *events.APIGatewayV2HTTPResponse) ResponseAdapter {
	return &responseLambdaAdapter{
		header:   make(http.Header),
		response: response,
	}
}

type responseLambdaAdapter struct {
	header   http.Header
	buffer   bytes.Buffer
	response *events.APIGatewayV2HTTPResponse
}

func (w *responseLambdaAdapter) Header() http.Header {
	return w.header
}

func (w *responseLambdaAdapter) Write(bytes []byte) (int, error) {
	return w.buffer.Write(bytes)
}

func (w *responseLambdaAdapter) WriteHeader(statusCode int) {
	w.response.StatusCode = statusCode
}

func (w *responseLambdaAdapter) SetCookie(cookie *http.Cookie) {
	v := cookie.String()
	if len(v) == 0 {
		return
	}

	w.setCookies(v)
}

func (w *responseLambdaAdapter) setCookies(v ...string) {
	w.response.Cookies = append(w.response.Cookies, v...)
}

func (w *responseLambdaAdapter) Commit() error {
	return w.commit()
}

func (w *responseLambdaAdapter) commit() error {
	if w.response.StatusCode == 0 {
		w.response.StatusCode = http.StatusOK
	}

	body, err := ioutil.ReadAll(&w.buffer)
	if err != nil {
		return err
	}

	w.response.IsBase64Encoded = w.isBinary()
	if w.response.IsBase64Encoded {
		w.response.Body = base64.StdEncoding.EncodeToString(body)
	} else {
		w.response.Body = string(body)
	}

	w.response.Headers = make(map[string]string)
	w.response.MultiValueHeaders = make(map[string][]string)
	
	for k, v := range w.header {
		// known-headers
		switch k {
		case HeaderSetCookie:
			w.setCookies(v...)
			continue
		}

		if len(v) == 1 {
			w.response.Headers[k] = v[0]
		} else {
			w.response.MultiValueHeaders[k] = v
		}
	}

	return nil
}

func (w *responseLambdaAdapter) isBinary() bool {
	contentType := w.header.Get(HeaderContentType)
	if len(contentType) == 0 {
		return false
	}

	return !notBinaryTable[contentType]
}
