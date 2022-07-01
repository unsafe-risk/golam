package golam

import (
	"net/http"
	"strings"
)

func newLocalRouter() Router {
	return &localRouter{
		root: routerNode{
			path:     "",
			children: make(map[string]*routerNode),
		},
	}
}

var _ Router = (*localRouter)(nil)

type routerNode struct {
	path           string
	route          *route
	greedyWildcard *routerNode
	wildcard       *routerNode
	children       map[string]*routerNode
}

type localRouter struct {
	root routerNode
}

func (lr *localRouter) FindRoute(path string) *route {
	path = replaceRootToEmpty(path)
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return lr.root.route
	}

	cur := &lr.root
	parts = parts[1:]
	greedyRoute := cur.greedyWildcard
	for _, k := range parts {
		next := cur.children[k]
		if next == nil {
			next = cur.wildcard
		}

		if next == nil {
			if greedyRoute != nil {
				return greedyRoute.route
			}
			return nil
		}

		if next.greedyWildcard != nil {
			greedyRoute = next.greedyWildcard
		}

		cur = next
	}

	return cur.route
}

func (lr *localRouter) AddRoute(method string, path string, handler HandlerFunc) {
	path = replaceRootToEmpty(path)
	parts := strings.Split(path, "/")
	targetRoute := &lr.root.route

	curPath := "/"

	var pathParams []routeParam
	if len(parts) > 1 {
		cur := &lr.root
		parts = parts[1:]

		for i, p := range parts {
			curPath += p + "/"

			startMarkerIdx := strings.Index(p, "{")
			endMarkerIdx := strings.Index(p, "}")

			isWildcard := startMarkerIdx != -1 && endMarkerIdx != -1
			isGreedyWildcard := isWildcard && p[endMarkerIdx-1] == '+'
			if isGreedyWildcard {
				if i != len(parts)-1 {
					panic("{greedy path variable+} must write suffix")
				}

				if cur.greedyWildcard != nil {
					panic("already exists greedy path")
				}

				pathParams = append(pathParams, routeParam{
					index:    i + 1,
					key:      p[startMarkerIdx+1 : endMarkerIdx-1],
					isGreedy: true,
				})

				cur.greedyWildcard = &routerNode{
					path: curPath[:len(curPath)-1],
				}

				cur = cur.greedyWildcard
				break
			}

			var next *routerNode
			if isWildcard {
				pathParams = append(pathParams, routeParam{
					index: i + 1,
					key:   p[startMarkerIdx+1 : endMarkerIdx],
				})
				next = cur.wildcard
			} else {
				next = cur.children[p]
			}

			if next == nil {
				next = &routerNode{
					path:     curPath[:len(curPath)-1],
					children: make(map[string]*routerNode),
				}

				if isWildcard {
					cur.wildcard = next
				} else {
					cur.children[p] = next
				}
			}

			cur = next
		}

		targetRoute = &cur.route
	}

	if *targetRoute == nil {
		*targetRoute = &route{
			path: curPath[:len(curPath)-1],
		}
	}

	method = replaceMethodWildcardToBlank(method)
	(*targetRoute).params.setParamsInfo(method, pathParams)
	(*targetRoute).handlers.setHandler(method, handler)
}

func (lr *localRouter) DelRoute(method string, path string) {
	//TODO implement me
	panic("implement me")

	//path = replaceRootToEmpty(path)
	//parts := strings.Split(path, "/")
	//targetRoute := &lr.root.route
	//
	//curPath := "/"
	//
	////depth := make()
	//if len(parts) > 1 {
	//	cur := &lr.root
	//	parts = parts[1:]
	//
	//	for i, p := range parts {
	//		curPath += p + "/"
	//
	//		startMarkerIdx := strings.Index(p, "{")
	//		endMarkerIdx := strings.Index(p, "}")
	//
	//		isWildcard := startMarkerIdx != -1 && endMarkerIdx != -1
	//		isGreedyWildcard := isWildcard && p[endMarkerIdx-1] == '+'
	//		if isGreedyWildcard {
	//			if i != len(parts)-1 {
	//				panic("{greedy path variable+} must write suffix")
	//			}
	//
	//			cur = cur.greedyWildcard
	//			break
	//		}
	//
	//		var next *routerNode
	//		if isWildcard {
	//			next = cur.wildcard
	//		} else {
	//			next = cur.children[p]
	//		}
	//
	//		if next == nil {
	//			break
	//		}
	//
	//		cur = next
	//	}
	//
	//	targetRoute = &cur.route
	//}
}

func (lr *localRouter) Any(path string, handler HandlerFunc) {
	lr.AddRoute("", path, handler)
}

func (lr *localRouter) GET(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodGet, path, handler)
}

func (lr *localRouter) HEAD(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodHead, path, handler)
}

func (lr *localRouter) POST(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPost, path, handler)
}

func (lr *localRouter) PUT(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPut, path, handler)
}

func (lr *localRouter) PATCH(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodPatch, path, handler)
}

func (lr *localRouter) DELETE(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodDelete, path, handler)
}

func (lr *localRouter) CONNECT(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodConnect, path, handler)
}

func (lr *localRouter) OPTIONS(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodOptions, path, handler)
}

func (lr *localRouter) TRACE(path string, handler HandlerFunc) {
	lr.AddRoute(http.MethodTrace, path, handler)
}
