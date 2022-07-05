package golam

import (
	"net/http"
	"strings"
)

type Router interface {
	FindRoute(path string) *route

	AddRoute(method string, path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	DelRoute(method string, path string)

	Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	CONNECT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	TRACE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
}

type routeParam struct {
	index    int
	key      string
	isGreedy bool
}

type route struct {
	path     string
	params   paramMethods
	handlers handlerMethods
}

type paramMethods struct {
	anyMethod     *[]routeParam
	getMethod     *[]routeParam
	headMethod    *[]routeParam
	postMethod    *[]routeParam
	putMethod     *[]routeParam
	patchMethod   *[]routeParam
	deleteMethod  *[]routeParam
	connectMethod *[]routeParam
	optionsMethod *[]routeParam
	traceMethod   *[]routeParam
	others        map[string]*[]routeParam
}

func (h *paramMethods) getParamsInfo(method string) *[]routeParam {
	switch method {
	case "":
		return h.anyMethod
	case http.MethodGet:
		return h.getMethod
	case http.MethodHead:
		return h.headMethod
	case http.MethodPost:
		return h.postMethod
	case http.MethodPut:
		return h.putMethod
	case http.MethodPatch:
		return h.patchMethod
	case http.MethodDelete:
		return h.deleteMethod
	case http.MethodConnect:
		return h.connectMethod
	case http.MethodOptions:
		return h.optionsMethod
	case http.MethodTrace:
		return h.traceMethod
	default:
		return h.others[method]
	}
}

func (h *paramMethods) setParamsInfo(method string, params []routeParam) {
	switch method {
	case "":
		h.anyMethod = &params
	case http.MethodGet:
		h.getMethod = &params
	case http.MethodHead:
		h.headMethod = &params
	case http.MethodPost:
		h.postMethod = &params
	case http.MethodPut:
		h.putMethod = &params
	case http.MethodPatch:
		h.patchMethod = &params
	case http.MethodDelete:
		h.deleteMethod = &params
	case http.MethodConnect:
		h.connectMethod = &params
	case http.MethodOptions:
		h.optionsMethod = &params
	case http.MethodTrace:
		h.traceMethod = &params
	default:
		h.others[method] = &params
	}
}

func (h *paramMethods) delParamsInfo(method string) {
	switch method {
	case "":
		h.anyMethod = nil
	case http.MethodGet:
		h.getMethod = nil
	case http.MethodHead:
		h.headMethod = nil
	case http.MethodPost:
		h.postMethod = nil
	case http.MethodPut:
		h.putMethod = nil
	case http.MethodPatch:
		h.patchMethod = nil
	case http.MethodDelete:
		h.deleteMethod = nil
	case http.MethodConnect:
		h.connectMethod = nil
	case http.MethodOptions:
		h.optionsMethod = nil
	case http.MethodTrace:
		h.traceMethod = nil
	default:
		delete(h.others, method)
	}
}

func (h *paramMethods) countParamsInfo() int {
	return existsParamsInfoCount(
		h.anyMethod,
		h.getMethod,
		h.headMethod,
		h.postMethod,
		h.putMethod,
		h.patchMethod,
		h.deleteMethod,
		h.connectMethod,
		h.optionsMethod,
		h.traceMethod,
	) + len(h.others)
}

func existsParamsInfoCount(handlers ...*[]routeParam) (c int) {
	for i := range handlers {
		if handlers[i] != nil {
			c++
		}
	}

	return
}

type handlerMethods struct {
	anyMethod     HandlerFunc
	getMethod     HandlerFunc
	headMethod    HandlerFunc
	postMethod    HandlerFunc
	putMethod     HandlerFunc
	patchMethod   HandlerFunc
	deleteMethod  HandlerFunc
	connectMethod HandlerFunc
	optionsMethod HandlerFunc
	traceMethod   HandlerFunc
	others        map[string]HandlerFunc
}

func (h *handlerMethods) getHandler(method string) (handler HandlerFunc) {
	switch method {
	case "":
		return h.anyMethod
	case http.MethodGet:
		return h.getMethod
	case http.MethodHead:
		return h.headMethod
	case http.MethodPost:
		return h.postMethod
	case http.MethodPut:
		return h.putMethod
	case http.MethodPatch:
		return h.patchMethod
	case http.MethodDelete:
		return h.deleteMethod
	case http.MethodConnect:
		return h.connectMethod
	case http.MethodOptions:
		return h.optionsMethod
	case http.MethodTrace:
		return h.traceMethod
	default:
		return h.others[method]
	}
}

func (h *handlerMethods) setHandler(method string, handler HandlerFunc) {
	switch method {
	case "":
		h.anyMethod = handler
	case http.MethodGet:
		h.getMethod = handler
	case http.MethodHead:
		h.headMethod = handler
	case http.MethodPost:
		h.postMethod = handler
	case http.MethodPut:
		h.putMethod = handler
	case http.MethodPatch:
		h.patchMethod = handler
	case http.MethodDelete:
		h.deleteMethod = handler
	case http.MethodConnect:
		h.connectMethod = handler
	case http.MethodOptions:
		h.optionsMethod = handler
	case http.MethodTrace:
		h.traceMethod = handler
	default:
		h.others[method] = handler
	}
}

func (h *handlerMethods) delHandler(method string) {
	switch method {
	case "":
		h.anyMethod = nil
	case http.MethodGet:
		h.getMethod = nil
	case http.MethodHead:
		h.headMethod = nil
	case http.MethodPost:
		h.postMethod = nil
	case http.MethodPut:
		h.putMethod = nil
	case http.MethodPatch:
		h.patchMethod = nil
	case http.MethodDelete:
		h.deleteMethod = nil
	case http.MethodConnect:
		h.connectMethod = nil
	case http.MethodOptions:
		h.optionsMethod = nil
	case http.MethodTrace:
		h.traceMethod = nil
	default:
		delete(h.others, method)
	}
}

func (h *handlerMethods) countHandler() int {
	return handlerNotNilCount(
		h.anyMethod,
		h.getMethod,
		h.headMethod,
		h.postMethod,
		h.putMethod,
		h.patchMethod,
		h.deleteMethod,
		h.connectMethod,
		h.optionsMethod,
		h.traceMethod,
	) + len(h.others)
}

func handlerNotNilCount(handlers ...HandlerFunc) (c int) {
	for i := range handlers {
		if handlers[i] != nil {
			c++
		}
	}

	return
}

func findParamKeys(path string) (res []routeParam) {
	parts := strings.Split(path, "/")
	for i, v := range parts {
		startMarkerIdx := strings.Index(v, "{")
		endMarkerIdx := strings.Index(v, "}")

		if startMarkerIdx == -1 || endMarkerIdx == -1 {
			continue
		}

		v = v[startMarkerIdx+1 : endMarkerIdx]
		rp := routeParam{
			index:    i,
			isGreedy: v[len(v)-1] == '+',
		}

		if rp.isGreedy {
			rp.key = v[:len(v)-1]
		}

		res = append(res, rp)
	}

	return
}

func replaceMethodWildcardToBlank(method string) string {
	if method == "*" {
		return ""
	}

	return method
}

func replaceRootToEmpty(path string) string {
	if path == "/" {
		return ""
	}

	return path
}
