package golam

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"net/url"
)

const (
	defaultIndent = "\t"
)

type Context interface {
	Ctx() context.Context

	Request() *Req

	Response() *Resp

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
	request             *Req
	response            *Resp
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
	return c.Request().Context()
}

func (c *contextImpl) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *contextImpl) Request() *Req {
	return c.request
}

func (c *contextImpl) Response() *Resp {
	return c.response
}

func (c *contextImpl) Scheme() string {
	return c.Request().Scheme()
}

func (c *contextImpl) RealIP() string {
	return c.Request().RealIP()
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

func (c contextImpl) HTML(status int, html string) error {
	return c.HTMLBytes(status, []byte(html))
}

func (c contextImpl) HTMLBytes(status int, b []byte) error {
	return c.Result(status, MIMETextHTMLCharsetUTF8, b)
}

func (c contextImpl) String(status int, s string) error {
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

func (c contextImpl) JSONPretty(status int, i interface{}) error {
	return c.JSONWithIndent(status, i, defaultIndent)
}

func (c *contextImpl) JSONWithIndent(status int, i interface{}, indent string) error {
	return c.json(status, i, indent)
}

func (c contextImpl) JSONBytes(status int, b []byte) (err error) {
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

func (c *contextImpl) Result(status int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.Response().WriteHeader(status)
	_, err = c.Write(b)
	return
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
