package golam

import (
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"strconv"
)

func getSchemeFromHeader(header http.Header) string {
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

func getHeaderFromAPIGatewayV2HTTPRequest(request *events.APIGatewayV2HTTPRequest) (header http.Header) {
	header = make(http.Header)
	for k, v := range request.Headers {
		header.Set(k, v)
	}
	return
}

const (
	lambdaHeaderContentLength = "content-length"
)

func getContentLengthFromAPIGatewayV2HTTPRequest(request *events.APIGatewayV2HTTPRequest) (l int64) {
	if request == nil {
		return
	}

	l, _ = strconv.ParseInt(request.Headers[lambdaHeaderContentLength], 10, 0)
	return
}
