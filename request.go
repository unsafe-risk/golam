package golam

import (
	"context"
	"net"
	"net/http"
	"net/textproto"
	"strings"
)

type Req struct {
	Method string

	ProtoMajor int // 1
	ProtoMinor int // 0

	ReqHeader map[string]string

	Body string

	ContentLength int64

	Host string

	RemoteAddr string

	cookies     []*http.Cookie
	cookieTable map[string][]*http.Cookie

	ctx context.Context
}

func (r *Req) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Req) WithContext(ctx context.Context) *Req {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Req)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

func (r *Req) Clone(ctx context.Context) *Req {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Req)
	*r2 = *r
	r2.ctx = ctx

	if len(r.ReqHeader) > 0 {
		r2.ReqHeader = make(map[string]string)
		for k, v := range r.ReqHeader {
			r2.ReqHeader[k] = v
		}
	}

	return r2
}

func (r *Req) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major ||
		r.ProtoMajor == major && r.ProtoMinor >= minor
}

func (r *Req) readHeader(h http.Header) {
	for k := range h {
		r.ReqHeader[strings.ToLower(k)] = h.Get(k)
	}
}

func (r *Req) UserAgent() string {
	return r.ReqHeader["user-agent"]
}

func (r *Req) readCookiesFromHeader(h http.Header) {
	lines := h["Cookie"]
	if len(lines) == 0 {
		r.cookies = []*http.Cookie{}
		return
	}

	cookies := make([]*http.Cookie, 0, len(lines)+strings.Count(lines[0], ";"))
	cookieTable := make(map[string][]*http.Cookie)
	for _, line := range lines {
		line = textproto.TrimString(line)

		var part string
		for len(line) > 0 { // continue since we have rest
			if splitIndex := strings.Index(line, ";"); splitIndex > 0 {
				part, line = line[:splitIndex], line[splitIndex+1:]
			} else {
				part, line = line, ""
			}
			part = textproto.TrimString(part)
			if len(part) == 0 {
				continue
			}
			name, val := part, ""
			if j := strings.Index(part, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if !isCookieNameValid(name) {
				continue
			}

			val, ok := parseCookieValue(val, true)
			if !ok {
				continue
			}
			cookie := &http.Cookie{Name: name, Value: val}
			cookies = append(cookies, cookie)
			cookieTable[name] = append(cookieTable[name], cookie)
		}
	}

	r.cookies = cookies
	r.cookieTable = cookieTable
}

func (r *Req) readCookiesFromStrings(values []string) {
	cookies := make([]*http.Cookie, 0, len(values))
	cookieTable := make(map[string][]*http.Cookie)
	for _, part := range values {
		part = textproto.TrimString(part)
		if len(part) == 0 {
			continue
		}
		name, val := part, ""
		if j := strings.Index(part, "="); j >= 0 {
			name, val = name[:j], name[j+1:]
		}
		if !isCookieNameValid(name) {
			continue
		}

		val, ok := parseCookieValue(val, true)
		if !ok {
			continue
		}
		cookie := &http.Cookie{Name: name, Value: val}
		cookies = append(cookies, cookie)
		cookieTable[name] = append(cookieTable[name], cookie)
	}

	r.cookies = cookies
	r.cookieTable = cookieTable
}

func (r *Req) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *Req) Cookie(name string) (*http.Cookie, error) {
	for _, c := range r.cookieTable[name] {
		return c, nil
	}

	return nil, http.ErrNoCookie
}

func (r *Req) Referer() string {
	return r.ReqHeader["referer"]
}

func (r *Req) Scheme() string {
	if scheme := r.ReqHeader[headerXForwardedProto]; scheme != "" {
		return scheme
	}
	if scheme := r.ReqHeader[headerXForwardedProtocol]; scheme != "" {
		return scheme
	}
	if ssl := r.ReqHeader[headerXForwardedSsl]; ssl == "on" {
		return "https"
	}
	if scheme := r.ReqHeader[headerXUrlScheme]; scheme != "" {
		return scheme
	}
	return "http"
}

func (r *Req) RealIP() string {
	if ip := r.ReqHeader[headerXForwardedFor]; ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			return strings.TrimSpace(ip[:i])
		}
		return ip
	}
	if ip := r.ReqHeader[headerXRealIP]; ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}

var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}

func isTokenRune(r rune) bool {
	i := int(r)
	return i < len(isTokenTable) && isTokenTable[i]
}

func isNotToken(r rune) bool {
	return !isTokenRune(r)
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}
