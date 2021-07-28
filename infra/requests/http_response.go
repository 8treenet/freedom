package requests

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
)

// Response is a return carry information for HTTP requests.
type Response struct {
	stdResponse   *http.Response
	Error         error
	HTTP11        bool
	ContentType   string
	Status        string // e.g. "200 OK"
	StatusCode    int    // e.g. 200
	Proto         string // e.g. "HTTP/1.0"
	ProtoMajor    int    // e.g. 1
	ProtoMinor    int    // e.g. 0
	Header        http.Header
	ContentLength int64
	Uncompressed  bool
	traceInfo     HTTPTraceInfo
	cookies       []*http.Cookie
}

// Clone Copy a new one.
func (res *Response) Clone() *Response {
	return &Response{
		stdResponse:   res.stdResponse,
		Error:         res.Error,
		HTTP11:        res.HTTP11,
		ContentType:   res.ContentType,
		Status:        res.Status,
		StatusCode:    res.StatusCode,
		Proto:         res.Proto,
		ProtoMajor:    res.ProtoMajor,
		ProtoMinor:    res.ProtoMinor,
		Header:        res.Header.Clone(),
		ContentLength: res.ContentLength,
		Uncompressed:  res.Uncompressed,
		traceInfo:     res.traceInfo,
	}
}

// TraceInfo Link information for the connection.
func (res *Response) TraceInfo() HTTPTraceInfo {
	return res.traceInfo
}

// ProtoAtLeast reports whether the HTTP protocol used
// in the response is at least major.minor.
func (res *Response) ProtoAtLeast(major, minor int) bool {
	return res.ProtoMajor > major ||
		res.ProtoMajor == major && res.ProtoMinor >= minor
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (res *Response) Cookies() []*http.Cookie {
	if cap(res.cookies) > 0 {
		return res.cookies
	}

	res.cookies = make([]*http.Cookie, 0, 1)
	res.cookies = append(res.cookies, readSetCookies(res.Header)...)
	return res.cookies
}

// Cookie returns cookie's value by its name
// returns empty string if nothing was found.
func (res *Response) Cookie(name string) *http.Cookie {
	for _, cookie := range res.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// readSetCookies parses all "Set-Cookie" values from
// the header h and returns the successfully parsed Cookies.
func readSetCookies(h http.Header) []*http.Cookie {
	cookieCount := len(h["Set-Cookie"])
	if cookieCount == 0 {
		return []*http.Cookie{}
	}
	cookies := make([]*http.Cookie, 0, cookieCount)
	for _, line := range h["Set-Cookie"] {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		parts[0] = strings.TrimSpace(parts[0])
		j := strings.Index(parts[0], "=")
		if j < 0 {
			continue
		}
		name, value := parts[0][:j], parts[0][j+1:]
		if !isCookieNameValid(name) {
			continue
		}
		value, ok := parseCookieValue(value, true)
		if !ok {
			continue
		}
		c := &http.Cookie{
			Name:  name,
			Value: value,
			Raw:   line,
		}
		for i := 1; i < len(parts); i++ {
			parts[i] = strings.TrimSpace(parts[i])
			if len(parts[i]) == 0 {
				continue
			}

			attr, val := parts[i], ""
			if j := strings.Index(attr, "="); j >= 0 {
				attr, val = attr[:j], attr[j+1:]
			}
			lowerAttr := strings.ToLower(attr)
			val, ok = parseCookieValue(val, false)
			if !ok {
				c.Unparsed = append(c.Unparsed, parts[i])
				continue
			}
			switch lowerAttr {
			case "samesite":
				lowerVal := strings.ToLower(val)
				switch lowerVal {
				case "lax":
					c.SameSite = http.SameSiteLaxMode
				case "strict":
					c.SameSite = http.SameSiteStrictMode
				case "none":
					c.SameSite = http.SameSiteNoneMode
				default:
					c.SameSite = http.SameSiteDefaultMode
				}
				continue
			case "secure":
				c.Secure = true
				continue
			case "httponly":
				c.HttpOnly = true
				continue
			case "domain":
				c.Domain = val
				continue
			case "max-age":
				secs, err := strconv.Atoi(val)
				if err != nil || secs != 0 && val[0] == '0' {
					break
				}
				if secs <= 0 {
					secs = -1
				}
				c.MaxAge = secs
				continue
			case "expires":
				c.RawExpires = val
				exptime, err := time.Parse(time.RFC1123, val)
				if err != nil {
					exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", val)
					if err != nil {
						c.Expires = time.Time{}
						break
					}
				}
				c.Expires = exptime.UTC()
				continue
			case "path":
				c.Path = val
				continue
			}
			c.Unparsed = append(c.Unparsed, parts[i])
		}
		cookies = append(cookies, c)
	}
	return cookies
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
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

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}
