package webapp

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Common HTTP headers
const (
	HeaderAccept             = "Accept"
	HeaderAllow              = "Allow"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentEncoding    = "Content-Encoding"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
	HeaderLocation           = "Location"
)

// Content types
const (
	HTMLType = "text/html; charset=utf-8"
	JSONType = "application/json; charset=utf-8"
)

// MethodNotAllowed replies to a request with an HTTP 405 method not allowed error.
func MethodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set(HeaderAllow, strings.Join(methods, ", "))
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// ContentLength sets the Content-Length header to size.
func ContentLength(h http.Header, size int64) {
	h.Set(HeaderContentLength, strconv.FormatInt(size, 10))
}

// Attachment sets the Content-Disposition header to an attachment with the given file name.
func Attachment(h http.Header, base, ext string) {
	h.Set(HeaderContentDisposition, "attachment; filename="+base+"."+ext)
}

// JSONResponse writes a JSON value to w, setting the Content-Type.
func JSONResponse(w http.ResponseWriter, v interface{}) error {
	w.Header().Set(HeaderContentType, JSONType)
	return json.NewEncoder(w).Encode(v)
}

func quoteHTTP(s string) string {
	if s == "" {
		return `""`
	}
	isToken := true
	for _, r := range s {
		if !isTokenChar(r) {
			isToken = false
			break
		}
	}
	if isToken {
		return s
	}
	sb := make([]byte, 0, len(s)+2)
	sb = append(sb, '"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\', '"':
			sb = append(sb, '\\', c)
		default:
			sb = append(sb, c)
		}
	}
	sb = append(sb, '"')
	return string(sb)
}

// httpParser is a rune-based parser that has rules for common HTTP productions.
type httpParser struct {
	r []rune
}

func (p *httpParser) eof() bool {
	return len(p.r) == 0
}

func (p *httpParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.r[0]
}

func (p *httpParser) consume(literal string) bool {
	if len(p.r) < len(literal) {
		return false
	}
	if len(literal) == 1 {
		// common case
		ok := p.r[0] == rune(literal[0])
		if ok {
			p.r = p.r[1:]
		}
		return ok
	}

	i := 0
	for _, c := range literal {
		if p.r[i] != c {
			return false
		}
		i++
	}
	p.r = p.r[i:]
	return true
}

func (p *httpParser) run(f func(rune) bool) []rune {
	var i int
	for i = 0; i < len(p.r); i++ {
		if !f(p.r[i]) {
			break
		}
	}
	run := p.r[:i]
	p.r = p.r[i:]
	return run
}

const tokenChars = "!#$%&'*+-.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ^_`abcdefghijklmnopqrstuvwxyz|~"

func isTokenChar(r rune) bool {
	return strings.IndexRune(tokenChars, r) != -1
}

func (p *httpParser) token() []rune {
	var i int
	for i = 0; i < len(p.r); i++ {
		if !isTokenChar(p.r[i]) {
			break
		}
	}
	run := p.r[:i]
	p.r = p.r[i:]
	return run
}

func (p *httpParser) quotedString() ([]rune, error) {
	if len(p.r) == 0 || p.r[0] != '"' {
		return nil, nil
	}
	s := make([]rune, 0, len(p.r)-1)
	var i int
	for i = 1; i < len(p.r); i++ {
		if ru := p.r[i]; ru == '"' {
			p.r = p.r[i+1:]
			return s, nil
		} else if ru == '\\' {
			i++
			if i < len(p.r) {
				s = append(s, p.r[i])
			} else {
				p.r = p.r[i:]
				return s, io.ErrUnexpectedEOF
			}
		} else {
			s = append(s, ru)
		}
	}
	p.r = p.r[i:]
	return s, io.ErrUnexpectedEOF
}

func (p *httpParser) space() []rune {
	var i int
	for i = 0; i < len(p.r); i++ {
		if ru := p.r[i]; ru != ' ' && ru != '\t' {
			break
		}
	}
	run := p.r[:i]
	p.r = p.r[i:]
	return run
}
