package golam

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

func wrapMiddleware(handler HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
