package golam

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	lambdaHeaderContentLength   = "content-length"
	lambdaHeaderXForwardedProto = "x-forwarded-proto"
	lambdaHeaderHost            = "host"
)

func New() (g *Golam) {
	g = &Golam{
		isLambdaRuntime: isLambdaRuntime(),
		NotFoundHandler: DefaultNotFound,
	}

	if g.isLambdaRuntime {
		g.router = newLambdaRouter()
		g.LambdaHandler = &defaultLambdaHandler{golam: g}
		g.start = func() error {
			lambda.Start(g.LambdaHandler)
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

var (
	_ http.Handler   = (*defaultHttpHandler)(nil)
	_ lambda.Handler = (*defaultLambdaHandler)(nil)
)

type (
	defaultHttpHandler struct {
		golam *Golam
	}

	defaultLambdaHandler struct {
		golam *Golam
	}

	HandlerFunc func(c Context) error

	Golam struct {
		LocalAddr    string
		LocalHandler http.Handler

		LambdaHandler lambda.Handler

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
	reqPath := request.URL.Path

	ctxImpl := &contextImpl{
		request: request,
		response: &Response{
			header: make(http.Header),
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

		for k, values := range resp.header {
			for _, val := range values {
				writer.Header().Add(k, val)
			}
		}

		for _, cookie := range resp.cookies {
			http.SetCookie(writer, cookie)
		}

		writer.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(writer, &resp.BodyBuffer)
	default:
		// TODO error handler
	}
}

func (d *defaultLambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	var request events.APIGatewayV2HTTPRequest
	err := json.Unmarshal(payload, &request)
	if err != nil {
		//TODO: logging
		return nil, err
	}

	method := request.RequestContext.HTTP.Method
	var rawURL strings.Builder
	rawURL.WriteString(request.Headers[lambdaHeaderXForwardedProto])
	rawURL.WriteString("://")
	rawURL.WriteString(request.Headers[lambdaHeaderHost])
	rawURL.WriteString(request.RawPath)
	if len(request.RawQueryString) > 0 {
		rawURL.WriteString("?")
		rawURL.WriteString(request.RawQueryString)
	}

	var body []byte
	if request.IsBase64Encoded {
		body, err = base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			//TODO: logging
			return nil, err
		}
	} else {
		body = []byte(request.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL.String(), bytes.NewReader(body))
	if err != nil {
		//TODO: logging
		return nil, err
	}

	req.ProtoMajor, req.ProtoMinor = getProtoVersion(&request)
	req.RemoteAddr = request.RequestContext.HTTP.SourceIP
	for k, v := range request.Headers {
		if req.Header.Get(k) == v {
			continue
		}

		req.Header.Set(k, v)
	}

	ctxImpl := &contextImpl{
		request: req,
		response: &Response{
			header: make(http.Header),
		},
		path:                request.RequestContext.HTTP.Path,
		golam:               d.golam,
		primalRequestLambda: &request,
	}

	ctxImpl.query, _ = url.ParseQuery(request.RawQueryString)

	pathKey := request.RouteKey[strings.Index(request.RouteKey, " ")+1:]
	r := d.golam.Router().FindRoute(pathKey)
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
		ctxImpl.handler = d.golam.NotFoundHandler
	} else {
		ctxImpl.handler = wrapMiddleware(ctxImpl.handler, append(d.golam.preMiddleware, d.golam.middleware...)...)
	}

	var response events.APIGatewayV2HTTPResponse
	switch err = ctxImpl.handler(ctxImpl); err {
	case nil:
		resp := ctxImpl.Response()
		resp.Committed = true

		response = events.APIGatewayV2HTTPResponse{
			StatusCode:        resp.StatusCode,
			Headers:           make(map[string]string),
			MultiValueHeaders: make(map[string][]string),
			IsBase64Encoded:   resp.isBinary,
			Body:              resp.BodyBuffer.String(),
		}

		if response.IsBase64Encoded {
			response.Body = base64.StdEncoding.EncodeToString(resp.BodyBuffer.Bytes())
		} else {
			response.Body = resp.BodyBuffer.String()
		}

		for k, v := range resp.header {
			if len(v) > 1 {
				response.MultiValueHeaders[k] = v
			} else {
				response.Headers[k] = v[0]
			}
		}

		response.Cookies = make([]string, 0, len(resp.cookies))
		for i := range resp.cookies {
			response.Cookies = append(response.Cookies, resp.cookies[i].String())
		}

	default:
		// TODO error handler
	}

	return json.Marshal(response)
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

	l, _ = strconv.ParseInt(request.Headers[lambdaHeaderContentLength], 10, 0)
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
