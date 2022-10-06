package golam

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultIndent = "\t"
)

type Context interface {
	Ctx() context.Context

	Set(key interface{}, val interface{})

	Get(key interface{}) interface{}

	SetRequest(r *http.Request)

	Request() *http.Request

	RequestBody() io.ReadCloser

	RequestBodyString() string

	RequestBodyBytes() []byte

	Response() *Response

	Scheme() string

	RealIP() string

	Path() string

	SetPath(p string)

	PathParams() PathParams

	SetPathParams(pathParams PathParams)

	QueryParams() url.Values

	SetQueryParams(query url.Values)

	Cookie(name string) (*http.Cookie, error)

	SetCookie(cookie *http.Cookie)

	Cookies() []*http.Cookie

	NoContent(status int) error

	HTML(status int, html string) error

	HTMLBytes(status int, b []byte) error

	String(status int, s string) error

	JSON(status int, i interface{}) error

	JSONPretty(status int, i interface{}) error

	JSONWithIndent(status int, i interface{}, indent string) error

	JSONBytes(status int, b []byte) error

	XML(status int, i interface{}) error

	XMLPretty(status int, i interface{}) error

	XMLWithIndent(status int, i interface{}, indent string) error

	XMLBytes(status int, b []byte) error

	ResultStream(status int, contentType string, reader io.Reader) error

	Result(status int, contentType string, b []byte) error

	Write(b []byte) (int, error)

	// TODO
	//Logger() log.Logger

	// TODO
	//SetLogger(l log.Logger)

	Golam() *Golam

	PrimalRequest() interface{}

	PrimalRequestLambda() *events.APIGatewayV2HTTPRequest

	PrimalRequestHTTP() *http.Request
}

var _ Context = (*contextImpl)(nil)

type contextImpl struct {
	ctx                 context.Context
	request             *http.Request
	requestBodyBytes    []byte
	response            *Response
	path                string
	pathParams          PathParams
	query               url.Values
	handler             HandlerFunc
	golam               *Golam
	primalRequestLambda *events.APIGatewayV2HTTPRequest
	primalRequestHTTP   *http.Request

	// TODO
	//logger   log.Logger

}

func (c *contextImpl) Ctx() context.Context {
	if c.ctx == nil {
		c.ctx = c.Request().Context()
	}
	return c.ctx
}

func (c *contextImpl) Set(key interface{}, val interface{}) {
	c.ctx = context.WithValue(c.Ctx(), key, val)
}

func (c *contextImpl) Get(key interface{}) interface{} {
	return c.Ctx().Value(key)
}

func (c *contextImpl) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *contextImpl) SetRequest(r *http.Request) {
	c.request = r
}

func (c *contextImpl) Request() *http.Request {
	return c.request
}

func (c *contextImpl) RequestBody() io.ReadCloser {
	return c.Request().Body
}

func (c *contextImpl) RequestBodyString() string {
	return string(c.RequestBodyBytes())
}

func (c *contextImpl) RequestBodyBytes() []byte {
	if c.requestBodyBytes != nil {
		return c.requestBodyBytes
	}

	var buf bytes.Buffer
	tee := io.TeeReader(c.request.Body, &buf)

	var err error
	c.requestBodyBytes, err = io.ReadAll(tee)
	if err != nil {
		//TODO: logging
		return []byte{}
	}

	c.request.Body = io.NopCloser(&buf)
	return c.requestBodyBytes
}

func (c *contextImpl) Response() *Response {
	return c.response
}

func (c *contextImpl) Scheme() string {
	header := c.request.Header
	if scheme := header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (c *contextImpl) RealIP() string {
	header := c.request.Header
	if ip := header.Get(HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			return strings.TrimSpace(ip[:i])
		}
		return ip
	}
	if ip := header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.request.RemoteAddr)
	return ra
}

func (c *contextImpl) Path() string {
	return c.path
}

func (c *contextImpl) SetPath(p string) {
	c.path = p
}

func (c *contextImpl) PathParams() PathParams {
	return c.pathParams
}

func (c *contextImpl) SetPathParams(pathParams PathParams) {
	c.pathParams = pathParams
}

func (c *contextImpl) QueryParams() url.Values {
	return c.query
}

func (c *contextImpl) SetQueryParams(query url.Values) {
	c.query = query
}

func (c *contextImpl) Cookie(name string) (*http.Cookie, error) {
	return c.Request().Cookie(name)
}

func (c *contextImpl) SetCookie(cookie *http.Cookie) {
	c.Response().SetCookie(cookie)
}

func (c *contextImpl) Cookies() []*http.Cookie {
	return c.Request().Cookies()
}

func (c *contextImpl) NoContent(status int) error {
	c.Response().WriteHeader(status)
	return nil
}

func (c *contextImpl) HTML(status int, html string) error {
	return c.HTMLBytes(status, []byte(html))
}

func (c *contextImpl) HTMLBytes(status int, b []byte) error {
	return c.Result(status, MIMETextHTMLCharsetUTF8, b)
}

func (c *contextImpl) String(status int, s string) error {
	return c.Result(status, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *contextImpl) json(status int, i interface{}, indent string) error {
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(status)
	enc := json.NewEncoder(c)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (c *contextImpl) JSON(status int, i interface{}) error {
	// TODO
	// indent := ""
	// if debugOn {
	// 	indent = defaultIndent
	// }
	return c.json(status, i, "")
}

func (c *contextImpl) JSONPretty(status int, i interface{}) error {
	return c.JSONWithIndent(status, i, defaultIndent)
}

func (c *contextImpl) JSONWithIndent(status int, i interface{}, indent string) error {
	return c.json(status, i, indent)
}

func (c *contextImpl) JSONBytes(status int, b []byte) (err error) {
	return c.Result(status, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *contextImpl) xml(status int, i interface{}, indent string) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.Response().WriteHeader(status)
	enc := xml.NewEncoder(c)
	if indent != "" {
		enc.Indent("", indent)
	}
	if _, err = c.Write([]byte(xml.Header)); err != nil {
		return
	}
	return enc.Encode(i)
}

func (c *contextImpl) XML(status int, i interface{}) error {
	// TODO
	// indent := ""
	// if debugOn {
	// 	indent = defaultIndent
	// }
	return c.xml(status, i, "")
}

func (c *contextImpl) XMLPretty(status int, i interface{}) error {
	return c.XMLWithIndent(status, i, defaultIndent)
}

func (c *contextImpl) XMLWithIndent(status int, i interface{}, indent string) error {
	return c.xml(status, i, indent)
}

func (c *contextImpl) XMLBytes(status int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.Response().WriteHeader(status)
	if _, err = c.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.Write(b)
	return
}

func (c *contextImpl) ResultStream(status int, contentType string, reader io.Reader) (err error) {
	c.writeContentType(contentType)
	if len(contentType) > 0 {
		c.Response().isBinary = !notBinaryTable[contentType]
	}
	c.Response().WriteHeader(status)
	_, err = io.Copy(c.Response(), reader)
	return
}

func (c *contextImpl) Result(status int, contentType string, b []byte) (err error) {
	return c.ResultStream(status, contentType, bytes.NewReader(b))
}

func (c *contextImpl) Write(b []byte) (int, error) {
	return c.Response().Write(b)
}

func (c *contextImpl) Golam() *Golam {
	return c.golam
}

func (c *contextImpl) PrimalRequest() interface{} {
	if isLambdaRuntime() {
		return c.PrimalRequestLambda()
	}

	return c.PrimalRequestHTTP()
}

func (c *contextImpl) PrimalRequestLambda() *events.APIGatewayV2HTTPRequest {
	return c.primalRequestLambda
}

func (c *contextImpl) PrimalRequestHTTP() *http.Request {
	return c.primalRequestHTTP
}
