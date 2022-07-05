package golam

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// MIME types
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationXML             = "application/xml"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextHTML                   = "text/html"
	MIMETextHTMLCharsetUTF8        = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; " + charsetUTF8
)

const (
	charsetUTF8 = "charset=UTF-8"
)

const (
	HeaderContentType = "Content-Type"
)

const (
	headerXForwardedFor      = "x-forwarded-for"
	headerXForwardedProto    = "x-forwarded-proto"
	headerXForwardedProtocol = "x-forwarded-protocol"
	headerXForwardedSsl      = "x-forwarded-ssl"
	headerXUrlScheme         = "x-url-scheme"
	headerXRealIP            = "x-real-ip"
	headerContentLength      = "content-length"
	headerHost               = "host"
)

func New() (g *Golam) {
	g = &Golam{
		isLambdaRuntime: isLambdaRuntime(),
		NotFoundHandler: DefaultNotFound,
	}

	if g.isLambdaRuntime {
		g.router = newLambdaRouter()
		g.start = func() error {
			lambda.Start(g.handleAPIGatewayToLambdaV2)
			return nil
		}
	} else {
		g.router = newLocalRouter()
		g.LocalHandler = &defaultHttpHandler{golam: g}
		g.start = func() error {
			return http.ListenAndServe(g.LocalAddr, g.LocalHandler)
		}
	}
	return
}

var _ http.Handler = (*defaultHttpHandler)(nil)

type (
	HandlerFunc func(c Context) error

	defaultHttpHandler struct {
		golam *Golam
	}

	Golam struct {
		LocalAddr    string
		LocalHandler http.Handler

		NotFoundHandler HandlerFunc

		// TODO
		//Logger log.Logger

		isLambdaRuntime bool
		start           func() error
		preMiddleware   []MiddlewareFunc
		middleware      []MiddlewareFunc

		router Router
	}
)

func (d *defaultHttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	method := request.Method

	ctxReq := &Req{
		Method:        request.Method,
		ProtoMajor:    request.ProtoMajor,
		ProtoMinor:    request.ProtoMinor,
		ReqHeader:     make(map[string]string),
		ContentLength: request.ContentLength,
		Host:          request.Host,
		RemoteAddr:    request.RemoteAddr,
		ctx:           request.Context(),
	}

	var b bytes.Buffer
	if written, err := io.Copy(&b, request.Body); err != nil {
		// TODO io 에러
	} else if written != request.ContentLength {
		// TODO 잘못된 컨텐츠 사이즈
	}

	reqHeader := request.Header
	ctxReq.readHeader(reqHeader)
	ctxReq.readCookiesFromHeader(reqHeader)
	ctxReq.Body = b.String()

	reqPath := request.URL.Path

	ctxImpl := &contextImpl{
		request: ctxReq,
		response: &Resp{
			RespHeader: make(http.Header),
		},
		path:              reqPath,
		query:             request.URL.Query(),
		golam:             d.golam,
		primalRequestHTTP: request,
	}

	r := d.golam.router.FindRoute(reqPath)
	if r != nil {
		var params *[]routeParam
		handler := r.handlers.getHandler(method)
		if handler != nil {
			ctxImpl.handler = handler
			params = r.params.getParamsInfo(method)
		} else {
			ctxImpl.handler = r.handlers.getHandler("")
			params = r.params.getParamsInfo("")
		}

		if params != nil && len(*params) > 0 {
			ctxImpl.pathParams = make(PathParams)
			parts := strings.Split(reqPath, "/")
			for _, p := range *params {
				if p.index >= len(parts) {
					// error?
					continue
				}

				if p.isGreedy {
					ctxImpl.pathParams[p.key] = PathParam{
						Key:   p.key,
						Value: strings.Join(parts[p.index:], "/"),
					}
					break
				}

				ctxImpl.pathParams[p.key] = PathParam{
					Key:   p.key,
					Value: parts[p.index],
				}
			}

		}

	}

	if ctxImpl.handler == nil {
		ctxImpl.handler = d.golam.NotFoundHandler
	} else {
		ctxImpl.handler = wrapMiddleware(ctxImpl.handler, append(d.golam.preMiddleware, d.golam.middleware...)...)
	}

	// TODO error handle
	switch err := ctxImpl.handler(ctxImpl); err {
	case nil:
		resp := ctxImpl.Response()
		resp.Committed = true

		for k, values := range resp.RespHeader {
			for _, val := range values {
				writer.Header().Add(k, val)
			}
		}

		for _, cookie := range resp.Cookies {
			http.SetCookie(writer, cookie)
		}

		writer.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(writer, &resp.BodyBuffer)
	default:
		// TODO error handler
	}
}

func (g *Golam) StartWithLocalAddr(localAddr string) error {
	g.LocalAddr = localAddr
	return g.Start()
}

func (g *Golam) Start() error {
	if g.start == nil {
		return errors.New("must setup start")
	}

	return g.start()
}

func (g *Golam) Router() Router {
	return g.router
}

func (g *Golam) handleAPIGatewayToLambdaV2(ctx context.Context, request events.APIGatewayV2HTTPRequest) (response events.APIGatewayV2HTTPResponse, err error) {

	method := request.RequestContext.HTTP.Method

	ctxReq := &Req{
		Method:        method,
		ReqHeader:     request.Headers,
		Body:          request.Body,
		ContentLength: getContentLength(&request),
		Host:          request.Headers[headerHost],
		RemoteAddr:    request.RequestContext.HTTP.SourceIP,
		ctx:           ctx,
	}

	ctxReq.ProtoMajor, ctxReq.ProtoMinor = getProtoVersion(&request)
	ctxReq.readCookiesFromStrings(request.Cookies)

	ctxImpl := &contextImpl{
		request: ctxReq,
		response: &Resp{
			RespHeader: make(http.Header),
		},
		path:                request.RequestContext.HTTP.Path,
		golam:               g,
		primalRequestLambda: &request,
	}

	ctxImpl.query, _ = url.ParseQuery(request.RawQueryString)

	pathKey := request.RouteKey[strings.Index(request.RouteKey, " ")+1:]
	r := g.Router().FindRoute(pathKey)
	if r != nil {
		handler := r.handlers.getHandler(method)
		if handler != nil {
			ctxImpl.handler = handler
		} else {
			ctxImpl.handler = r.handlers.getHandler("")
		}

		if len(request.PathParameters) > 0 {
			ctxImpl.pathParams = make(PathParams)
			for k, v := range request.PathParameters {
				ctxImpl.pathParams[k] = PathParam{
					Key:   k,
					Value: v,
				}
			}
		}

	}

	if ctxImpl.handler == nil {
		ctxImpl.handler = g.NotFoundHandler
	} else {
		ctxImpl.handler = wrapMiddleware(ctxImpl.handler, append(g.preMiddleware, g.middleware...)...)
	}

	// TODO error handle
	switch err = ctxImpl.handler(ctxImpl); err {
	case nil:
		resp := ctxImpl.Response()
		resp.Committed = true

		response = events.APIGatewayV2HTTPResponse{
			StatusCode:        resp.StatusCode,
			Headers:           make(map[string]string),
			MultiValueHeaders: make(map[string][]string),
			Body:              resp.BodyBuffer.String(),
		}

		for k, v := range resp.RespHeader {
			if len(v) > 1 {
				response.MultiValueHeaders[k] = v
			} else {
				response.Headers[k] = v[0]
			}
		}

		response.Cookies = make([]string, 0, len(resp.Cookies))
		for i := range resp.Cookies {
			response.Cookies = append(response.Cookies, resp.Cookies[i].String())
		}

	default:
		// TODO error handler
	}

	return
}

func getProtoVersion(request *events.APIGatewayV2HTTPRequest) (major, minor int) {
	if request == nil {
		return
	}
	rawProtocol := request.RequestContext.HTTP.Protocol
	div := strings.Index(rawProtocol, "/") + 1
	if div == -1 {
		return
	}

	proto := rawProtocol[div:]
	div = strings.Index(proto, ".")
	if div == -1 {
		return
	}

	major, _ = strconv.Atoi(proto[:div])
	minor, _ = strconv.Atoi(proto[div+1:])
	return
}

func getContentLength(request *events.APIGatewayV2HTTPRequest) (l int64) {
	if request == nil {
		return
	}

	l, _ = strconv.ParseInt(request.Headers[headerContentLength], 10, 0)
	return
}

func (g *Golam) Pre(middleware ...MiddlewareFunc) {
	g.preMiddleware = append(g.preMiddleware, middleware...)
}

func (g *Golam) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

func (g *Golam) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().Any(path, handler, middleware...)
}

func (g *Golam) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().GET(path, handler, middleware...)
}

func (g *Golam) HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().HEAD(path, handler, middleware...)
}

func (g *Golam) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().POST(path, handler, middleware...)
}

func (g *Golam) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().PUT(path, handler, middleware...)
}

func (g *Golam) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().PATCH(path, handler, middleware...)
}

func (g *Golam) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().DELETE(path, handler, middleware...)
}

func (g *Golam) CONNECT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().CONNECT(path, handler, middleware...)
}

func (g *Golam) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().OPTIONS(path, handler, middleware...)
}

func (g *Golam) TRACE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Router().TRACE(path, handler, middleware...)
}

const defaultNotFoundResponseData = "{\"message\":\"Not Found\"}"

func DefaultNotFound(ctx Context) error {
	return ctx.JSONBytes(http.StatusNotFound, []byte(defaultNotFoundResponseData))
}
