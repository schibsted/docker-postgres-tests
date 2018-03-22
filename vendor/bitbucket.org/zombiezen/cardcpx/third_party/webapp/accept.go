package webapp

import (
	"strconv"
	"strings"
)

// An AcceptHeader represents a set of media ranges as sent in the Accept header
// of an HTTP request.
//
// http://tools.ietf.org/html/rfc2616#section-14.1
type AcceptHeader []MediaRange

func (h AcceptHeader) String() string {
	parts := make([]string, len(h))
	for i := range h {
		mr := &h[i]
		parts[i] = mr.String()
	}
	return strings.Join(parts, ",")
}

// Quality returns the quality of a content type based on the media ranges in h.
func (h AcceptHeader) Quality(contentType string, params map[string][]string) float64 {
	results := make(mediaRangeMatches, 0, len(h))
	for i := range h {
		mr := &h[i]
		if m := mr.match(contentType, params); m.Valid {
			results = append(results, m)
		}
	}
	if len(results) == 0 {
		return 0.0
	}

	// find most specific
	i := 0
	for j := 1; j < results.Len(); j++ {
		if results.Less(j, i) {
			i = j
		}
	}
	return results[i].MediaRange.Quality
}

// ParseAcceptHeader parses an Accept header of an HTTP request.  The media
// ranges are unsorted.
func ParseAcceptHeader(accept string) (AcceptHeader, error) {
	ranges := make(AcceptHeader, 0)
	p := httpParser{r: []rune(accept)}
	p.space()
	for !p.eof() {
		if len(ranges) > 0 {
			if !p.consume(",") {
				return ranges, &parseError{Expected: "','", Found: p.peek(), EOF: p.eof()}
			}
			p.space()
		}

		r, err := parseMediaRange(&p)
		if err != nil {
			if r != "" {
				ranges = append(ranges, MediaRange{Range: r, Quality: 1.0})
			}
			return ranges, err
		}
		quality, params, err := parseAcceptParams(&p)
		ranges = append(ranges, MediaRange{Range: r, Quality: quality, Params: params})
		if err != nil {
			return ranges, err
		}

	}
	return ranges, nil
}

func parseMediaRange(p *httpParser) (string, error) {
	const sep = "/"
	type_ := p.token()
	if len(type_) == 0 {
		return "", &parseError{Expected: "token", Found: p.peek(), EOF: p.eof()}
	}
	if !p.consume(sep) {
		return string(type_), &parseError{Expected: "'" + sep + "'", In: "media-range", Found: p.peek(), EOF: p.eof()}
	}
	subtype := p.token()
	if len(subtype) == 0 {
		return string(type_[:len(type_)+len(sep)]), &parseError{Expected: "subtype", In: "media-range", Found: p.peek(), EOF: p.eof()}
	}
	return string(type_[:len(type_)+len(sep)+len(subtype)]), nil
}

func parseAcceptParams(p *httpParser) (float64, map[string][]string, error) {
	const qualityParam = "q"

	quality, params := 1.0, make(map[string][]string)
	p.space()
	for p.consume(";") {
		p.space()
		key := string(p.token())
		p.space()
		if !p.consume("=") {
			return quality, params, &parseError{Expected: "'='", In: "accept-params", Found: p.peek(), EOF: p.eof()}
		}
		p.space()
		var value string
		if s, err := p.quotedString(); err != nil {
			return quality, params, err
		} else if s != nil {
			value = string(s)
		} else {
			value = string(p.token())
		}
		p.space()

		if key == qualityParam {
			// check for qvalue
			if q, err := strconv.ParseFloat(value, 64); err != nil {
				return quality, params, &qvalueError{value, err}
			} else if q <= 0 || q > 1 {
				return quality, params, &qvalueError{value, errQValueRange}
			} else {
				quality = q
			}
		} else {
			params[key] = append(params[key], value)
		}
	}
	return quality, params, nil
}

// A MediaRange represents a set of MIME types as sent in the Accept header of
// an HTTP request.
type MediaRange struct {
	Range   string
	Quality float64
	Params  map[string][]string
}

// Match reports whether the range applies to a content type.
func (mr *MediaRange) Match(contentType string, params map[string][]string) bool {
	return mr.match(contentType, params).Valid
}

type mediaRangeMatch struct {
	MediaRange *MediaRange
	Valid      bool
	Type       int
	Subtype    int
	Params     int
}

type mediaRangeMatches []mediaRangeMatch

func (m mediaRangeMatches) Len() int      { return len(m) }
func (m mediaRangeMatches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m mediaRangeMatches) Less(i, j int) bool {
	mi, mj := &m[i], &m[j]
	switch {
	case !mi.Valid && !mj.Valid:
		return false
	case !mi.Valid && mj.Valid:
		return false
	case mi.Valid && !mj.Valid:
		return true
	}
	if mi.Params != mj.Params {
		return mi.Params > mj.Params
	}
	if mi.Subtype != mj.Subtype {
		return mi.Subtype > mj.Subtype
	}
	return mi.Type > mj.Type
}

func (mr *MediaRange) match(contentType string, params map[string][]string) mediaRangeMatch {
	mrType, mrSubtype := splitContentType(mr.Range)
	ctType, ctSubtype := splitContentType(contentType)
	match := mediaRangeMatch{MediaRange: mr}

	if !(mrSubtype == "*" || mrSubtype == ctSubtype) || !(mrType == "*" || mrType == ctType) {
		return match
	}
	if mrType != "*" {
		match.Type++
	}
	if mrSubtype != "*" {
		match.Subtype++
	}

	for k, v1 := range mr.Params {
		v2, ok := params[k]
		if !ok {
			return match
		}
		if len(v1) != len(v2) {
			return match
		}
		for i := range v1 {
			if v1[i] != v2[i] {
				return match
			}
		}
		match.Params++
	}
	match.Valid = true
	return match
}

func splitContentType(s string) (string, string) {
	i := strings.IndexRune(s, '/')
	if i == -1 {
		return "", ""
	}
	return s[:i], s[i+1:]
}

func (mr *MediaRange) String() string {
	parts := make([]string, 0, len(mr.Params)+1)
	parts = append(parts, mr.Range)
	if mr.Quality != 1.0 {
		parts = append(parts, "q="+strconv.FormatFloat(mr.Quality, 'f', 3, 64))
	}
	for k, vs := range mr.Params {
		for _, v := range vs {
			parts = append(parts, k+"="+quoteHTTP(v))
		}
	}
	return strings.Join(parts, ";")
}
