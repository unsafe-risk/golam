package golam

import (
	"net/http"
)

func newLambdaRouter() Router {
	return &lambdaRouter{
		routeTable: make(map[string]*route),
	}
}

var _ Router = (*lambdaRouter)(nil)

type lambdaRouter struct {
	routeTable map[string]*route
}

func (lr *lambdaRouter) FindRoute(path string) *route {
	return lr.routeTable[path]
}

func (lr *lambdaRouter) AddRoute(method string, path string, handler HandlerFunc) {
	r := lr.routeTable[path]
	if r == nil {
		r = &route{
			path: path,
		}
		lr.routeTable[path] = r
	}

	method = replaceMethodWildcardToBlank(method)
	r.params.setParamsInfo(method, findParamKeys(path))
	r.handlers.setHandler(method, handler)
}

func (lr *lambdaRouter) DelRoute(method string, path string) {
	r := lr.routeTable[path]
	if r == nil {
		return
	}

	method = replaceMethodWildcardToBlank(method)
	r.handlers.delHandler(method)
	r.params.delParamsInfo(method)
	if r.handlers.countHandler() == 0 && r.params.countParamsInfo() == 0 {
		delete(lr.routeTable, path)
	}
}

func (lr *lambdaRouter) Any(path string, handler HandlerFunc) {
	lr.AddRoute("", path, handler)
}

func (lr *lambdaRouter) GET(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodGet, path, handler)
}

func (lr *lambdaRouter) HEAD(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodHead, path, handler)
}

func (lr *lambdaRouter) POST(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPost, path, handler)
}

func (lr *lambdaRouter) PUT(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPut, path, handler)
}

func (lr *lambdaRouter) PATCH(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPatch, path, handler)
}

func (lr *lambdaRouter) DELETE(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodDelete, path, handler)
}

func (lr *lambdaRouter) CONNECT(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodConnect, path, handler)
}

func (lr *lambdaRouter) OPTIONS(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodOptions, path, handler)
}

func (lr *lambdaRouter) TRACE(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodTrace, path, handler)
}
