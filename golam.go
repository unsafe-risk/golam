package golam

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
		response: NewResponse(
			NewResponseHTTPAdapter(writer),
		),
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
		err = ctxImpl.Response().Commit()
		if err != nil {
			panic(err) // unreachable code
		}
	default:
		// TODO error handler
	}
}

func (d *defaultLambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	var lReq events.APIGatewayV2HTTPRequest
	err := json.Unmarshal(payload, &lReq)
	if err != nil {
		//TODO: logging
		return nil, err
	}

	req, err := newHTTPRequestFromAPIGatewayV2HTTPRequest(ctx, &lReq)
	if err != nil {
		//TODO: logging
		return nil, err
	}

	var response events.APIGatewayV2HTTPResponse
	ctxImpl := &contextImpl{
		request: req,
		response: NewResponse(
			NewResponseLambdaAdapter(&response),
		),
		path:                lReq.RequestContext.HTTP.Path,
		golam:               d.golam,
		primalRequestLambda: &lReq,
	}

	ctxImpl.query, _ = url.ParseQuery(lReq.RawQueryString)

	pathKey := lReq.RouteKey[strings.Index(lReq.RouteKey, " ")+1:]
	r := d.golam.Router().FindRoute(pathKey)
	if r != nil {
		handler := r.handlers.getHandler(req.Method)
		if handler != nil {
			ctxImpl.handler = handler
		} else {
			ctxImpl.handler = r.handlers.getHandler("")
		}

		if len(lReq.PathParameters) > 0 {
			ctxImpl.pathParams = make(PathParams)
			for k, v := range lReq.PathParameters {
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

	switch err = ctxImpl.handler(ctxImpl); err {
	case nil:
		err = ctxImpl.Response().Commit()
		if err != nil {
			//TODO: logging
			return nil, err
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

func newHTTPRequestFromAPIGatewayV2HTTPRequest(ctx context.Context, from *events.APIGatewayV2HTTPRequest) (req *http.Request, err error) {
	rawProtocol := from.RequestContext.HTTP.Protocol
	protoDiv := strings.Index(rawProtocol, "/")
	if protoDiv == -1 {
		return
	}

	proto := rawProtocol[protoDiv+1:]
	protoDiv = strings.Index(proto, ".")
	if protoDiv == -1 {
		return
	}

	major, err := strconv.Atoi(proto[:protoDiv])
	if err != nil {
		return
	}
	minor, err := strconv.Atoi(proto[protoDiv+1:])
	if err != nil {
		return
	}

	var body []byte
	if from.IsBase64Encoded {
		body, err = base64.StdEncoding.DecodeString(from.Body)
		if err != nil {
			return
		}
	} else {
		body = []byte(from.Body)
	}

	header := getHeaderFromAPIGatewayV2HTTPRequest(from)

	var rawURL strings.Builder
	rawURL.WriteString(getSchemeFromHeader(header))
	rawURL.WriteString("://")
	rawURL.WriteString(header.Get(HeaderHost))
	rawURL.WriteString(from.RawPath)
	if len(from.RawQueryString) > 0 {
		rawURL.WriteRune('?')
		rawURL.WriteString(from.RawQueryString)
	}

	req, err = http.NewRequestWithContext(ctx, from.RequestContext.HTTP.Method, rawURL.String(), bytes.NewReader(body))
	if err != nil {
		return
	}

	req.Proto = rawProtocol
	req.ProtoMajor = major
	req.ProtoMinor = minor
	req.RemoteAddr = from.RequestContext.HTTP.SourceIP
	req.Header = header
	return
}
