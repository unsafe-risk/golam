package golam

//type Req struct {
//	*http.Request
//}

//func (r *Req) readCookiesFromHeader(h http.Header) {
//	lines := h["Cookie"]
//	if len(lines) == 0 {
//		r.cookies = []*http.Cookie{}
//		return
//	}
//
//	cookies := make([]*http.Cookie, 0, len(lines)+strings.Count(lines[0], ";"))
//	cookieTable := make(map[string][]*http.Cookie)
//	for _, line := range lines {
//		line = textproto.TrimString(line)
//
//		var part string
//		for len(line) > 0 { // continue since we have rest
//			if splitIndex := strings.Index(line, ";"); splitIndex > 0 {
//				part, line = line[:splitIndex], line[splitIndex+1:]
//			} else {
//				part, line = line, ""
//			}
//			part = textproto.TrimString(part)
//			if len(part) == 0 {
//				continue
//			}
//			name, val := part, ""
//			if j := strings.Index(part, "="); j >= 0 {
//				name, val = name[:j], name[j+1:]
//			}
//			if !isCookieNameValid(name) {
//				continue
//			}
//
//			val, ok := parseCookieValue(val, true)
//			if !ok {
//				continue
//			}
//			cookie := &http.Cookie{Name: name, Value: val}
//			cookies = append(cookies, cookie)
//			cookieTable[name] = append(cookieTable[name], cookie)
//		}
//	}
//
//	r.cookies = cookies
//	r.cookieTable = cookieTable
//}

//func (r *Req) readCookiesFromStrings(values []string) {
//	cookies := make([]*http.Cookie, 0, len(values))
//	cookieTable := make(map[string][]*http.Cookie)
//	for _, part := range values {
//		part = textproto.TrimString(part)
//		if len(part) == 0 {
//			continue
//		}
//		name, val := part, ""
//		if j := strings.Index(part, "="); j >= 0 {
//			name, val = name[:j], name[j+1:]
//		}
//		if !isCookieNameValid(name) {
//			continue
//		}
//
//		val, ok := parseCookieValue(val, true)
//		if !ok {
//			continue
//		}
//		cookie := &http.Cookie{Name: name, Value: val}
//		cookies = append(cookies, cookie)
//		cookieTable[name] = append(cookieTable[name], cookie)
//	}
//
//	r.cookies = cookies
//	r.cookieTable = cookieTable
//}

//var isTokenTable = [127]bool{
//	'!':  true,
//	'#':  true,
//	'$':  true,
//	'%':  true,
//	'&':  true,
//	'\'': true,
//	'*':  true,
//	'+':  true,
//	'-':  true,
//	'.':  true,
//	'0':  true,
//	'1':  true,
//	'2':  true,
//	'3':  true,
//	'4':  true,
//	'5':  true,
//	'6':  true,
//	'7':  true,
//	'8':  true,
//	'9':  true,
//	'A':  true,
//	'B':  true,
//	'C':  true,
//	'D':  true,
//	'E':  true,
//	'F':  true,
//	'G':  true,
//	'H':  true,
//	'I':  true,
//	'J':  true,
//	'K':  true,
//	'L':  true,
//	'M':  true,
//	'N':  true,
//	'O':  true,
//	'P':  true,
//	'Q':  true,
//	'R':  true,
//	'S':  true,
//	'T':  true,
//	'U':  true,
//	'W':  true,
//	'V':  true,
//	'X':  true,
//	'Y':  true,
//	'Z':  true,
//	'^':  true,
//	'_':  true,
//	'`':  true,
//	'a':  true,
//	'b':  true,
//	'c':  true,
//	'd':  true,
//	'e':  true,
//	'f':  true,
//	'g':  true,
//	'h':  true,
//	'i':  true,
//	'j':  true,
//	'k':  true,
//	'l':  true,
//	'm':  true,
//	'n':  true,
//	'o':  true,
//	'p':  true,
//	'q':  true,
//	'r':  true,
//	's':  true,
//	't':  true,
//	'u':  true,
//	'v':  true,
//	'w':  true,
//	'x':  true,
//	'y':  true,
//	'z':  true,
//	'|':  true,
//	'~':  true,
//}

//func isTokenRune(r rune) bool {
//	i := int(r)
//	return i < len(isTokenTable) && isTokenTable[i]
//}
//
//func isNotToken(r rune) bool {
//	return !isTokenRune(r)
//}
//
//func isCookieNameValid(raw string) bool {
//	if raw == "" {
//		return false
//	}
//	return strings.IndexFunc(raw, isNotToken) < 0
//}
//
//func validCookieValueByte(b byte) bool {
//	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
//}
//
//func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
//	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
//		raw = raw[1 : len(raw)-1]
//	}
//	for i := 0; i < len(raw); i++ {
//		if !validCookieValueByte(raw[i]) {
//			return "", false
//		}
//	}
//	return raw, true
//}
