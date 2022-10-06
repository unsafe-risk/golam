package golam

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
	HeaderAccept                          = "Accept"
	HeaderAcceptEncoding                  = "Accept-Encoding"
	HeaderAllow                           = "Allow"
	HeaderAuthorization                   = "Authorization"
	HeaderContentDisposition              = "Content-Disposition"
	HeaderContentEncoding                 = "Content-Encoding"
	HeaderContentLength                   = "Content-Length"
	HeaderContentType                     = "Content-Type"
	HeaderCookie                          = "Cookie"
	HeaderSetCookie                       = "Set-Cookie"
	HeaderIfModifiedSince                 = "If-Modified-Since"
	HeaderLastModified                    = "Last-Modified"
	HeaderLocation                        = "Location"
	HeaderRetryAfter                      = "Retry-After"
	HeaderUpgrade                         = "Upgrade"
	HeaderVary                            = "Vary"
	HeaderWWWAuthenticate                 = "WWW-Authenticate"
	HeaderXForwardedFor                   = "X-Forwarded-For"
	HeaderXForwardedProto                 = "X-Forwarded-Proto"
	HeaderXForwardedProtocol              = "X-Forwarded-Protocol"
	HeaderXForwardedSsl                   = "X-Forwarded-Ssl"
	HeaderXUrlScheme                      = "X-Url-Scheme"
	HeaderXHTTPMethodOverride             = "X-HTTP-Method-Override"
	HeaderXRealIP                         = "X-Real-Ip"
	HeaderXRequestID                      = "X-Request-Id"
	HeaderXCorrelationID                  = "X-Correlation-Id"
	HeaderXRequestedWith                  = "X-Requested-With"
	HeaderServer                          = "Server"
	HeaderOrigin                          = "Origin"
	HeaderHost                            = "Host"
	HeaderCacheControl                    = "Cache-Control"
	HeaderConnection                      = "Connection"
	HeaderAccessControlRequestMethod      = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders     = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin        = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods       = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders       = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials   = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders      = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge             = "Access-Control-Max-Age"
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
)

var (
	notBinaryTable = map[string]bool{
		MIMEApplicationJSON:            true,
		MIMEApplicationJSONCharsetUTF8: true,
		MIMEApplicationXML:             true,
		MIMEApplicationXMLCharsetUTF8:  true,
		MIMETextHTML:                   true,
		MIMETextHTMLCharsetUTF8:        true,
		MIMETextPlain:                  true,
		MIMETextPlainCharsetUTF8:       true,
	}
)
