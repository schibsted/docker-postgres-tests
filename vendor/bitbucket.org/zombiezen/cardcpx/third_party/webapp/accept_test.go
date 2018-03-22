package webapp

import (
	"reflect"
	"testing"
)

func TestAcceptHeader(t *testing.T) {
	type QualityCheck struct {
		Type    string
		Params  map[string][]string
		Quality float64
	}

	tests := []struct {
		Accept string
		Checks []QualityCheck
	}{
		{
			"text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5",
			[]QualityCheck{
				{"text/html", map[string][]string{"level": {"1"}}, 1.0},
				{"text/html", map[string][]string{"level": {"1"}}, 1.0},
				{"text/html", map[string][]string{}, 0.7},
				{"text/plain", map[string][]string{}, 0.3},
				{"image/jpeg", map[string][]string{}, 0.5},
				{"text/html", map[string][]string{"level": {"2"}}, 0.4},
				{"text/html", map[string][]string{"level": {"3"}}, 0.7},
			},
		},
	}
	for _, test := range tests {
		h, err := ParseAcceptHeader(test.Accept)
		if err != nil {
			t.Errorf("ParseAcceptHeader(%q) error: %v", test.Accept, err)
			continue
		}
		for _, check := range test.Checks {
			q := h.Quality(check.Type, check.Params)
			if q != check.Quality {
				t.Errorf("Accept: %s\n%v = %.3f, want %.3f", test.Accept, &MediaRange{Range: check.Type, Quality: 1.0, Params: check.Params}, q, check.Quality)
			}
		}
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		Accept      string
		Expect      AcceptHeader
		ExpectError bool
	}{
		{"", AcceptHeader{}, false},
		{"foo/)bar", AcceptHeader{MediaRange{"foo/", 1.0, nil}}, true},
		{
			`text/html; q=1`,
			AcceptHeader{
				{"text/html", 1.0, map[string][]string{}},
			},
			false,
		},
		{
			`text/html; q=0.001`,
			AcceptHeader{
				{"text/html", 0.001, map[string][]string{}},
			},
			false,
		},
		{
			`text/html; q=0`,
			AcceptHeader{
				{"text/html", 1.0, map[string][]string{}},
			},
			true,
		},
		{
			`text/html; q=1.5`,
			AcceptHeader{
				{"text/html", 1.0, map[string][]string{}},
			},
			true,
		},
		{
			"audio/*; q=0.2, audio/basic",
			AcceptHeader{
				{"audio/*", 0.2, map[string][]string{}},
				{"audio/basic", 1.0, map[string][]string{}},
			},
			false,
		},
		{
			`text/html; charset="utf-8"`,
			AcceptHeader{
				{"text/html", 1.0, map[string][]string{"charset": {"utf-8"}}},
			},
			false,
		},
		{
			`text/html; charset="utf-8"; charset="utf 8"; charset="utf\"8"`,
			AcceptHeader{
				{"text/html", 1.0, map[string][]string{"charset": {"utf-8", "utf 8", "utf\"8"}}},
			},
			false,
		},
		{
			"text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c",
			AcceptHeader{
				{"text/plain", 0.5, map[string][]string{}},
				{"text/html", 1.0, map[string][]string{}},
				{"text/x-dvi", 0.8, map[string][]string{}},
				{"text/x-c", 1.0, map[string][]string{}},
			},
			false,
		},
		{
			"text/*, text/html, text/html;level=1, */*",
			AcceptHeader{
				{"text/*", 1.0, map[string][]string{}},
				{"text/html", 1.0, map[string][]string{}},
				{"text/html", 1.0, map[string][]string{"level": {"1"}}},
				{"*/*", 1.0, map[string][]string{}},
			},
			false,
		},
		{
			"text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5",
			AcceptHeader{
				{"text/*", 0.3, map[string][]string{}},
				{"text/html", 0.7, map[string][]string{}},
				{"text/html", 1.0, map[string][]string{"level": {"1"}}},
				{"text/html", 0.4, map[string][]string{"level": {"2"}}},
				{"*/*", 0.5, map[string][]string{}},
			},
			false,
		},
	}

	for _, test := range tests {
		mr, err := ParseAcceptHeader(test.Accept)
		if err != nil && !test.ExpectError {
			t.Errorf("ParseAcceptHeader(%q) error: %v", test.Accept, err)
		} else if err == nil && test.ExpectError {
			t.Errorf("ParseAcceptHeader(%q) error = nil", test.Accept)
		}
		if !reflect.DeepEqual(mr, test.Expect) {
			t.Errorf("ParseAcceptHeader(%q) = %v; want %v", test.Accept, mr, test.Expect)
		}
	}
}

func TestMediaRange_match(t *testing.T) {
	tests := []struct {
		Range  string
		Params map[string][]string

		ContentType   string
		ContentParams map[string][]string

		Match mediaRangeMatch
	}{
		{
			"text/html", map[string][]string{},
			"text/html", map[string][]string{},
			mediaRangeMatch{nil, true, 1, 1, 0},
		},
		{
			"text/html", map[string][]string{},
			"text/plain", map[string][]string{},
			mediaRangeMatch{nil, false, 0, 0, 0},
		},
		{
			"text/*", map[string][]string{},
			"image/jpeg", map[string][]string{},
			mediaRangeMatch{nil, false, 0, 0, 0},
		},
		{
			"text/*", map[string][]string{},
			"text/plain", map[string][]string{},
			mediaRangeMatch{nil, true, 1, 0, 0},
		},
		{
			"*/*", map[string][]string{},
			"image/jpeg", map[string][]string{},
			mediaRangeMatch{nil, true, 0, 0, 0},
		},
		{
			"text/html", map[string][]string{"level": {"1"}},
			"text/html", map[string][]string{"level": {"1"}},
			mediaRangeMatch{nil, true, 1, 1, 1},
		},
		{
			"text/html", map[string][]string{"level": {"1"}},
			"text/html", map[string][]string{"level": {"2"}},
			mediaRangeMatch{nil, false, 1, 1, 0},
		},
		{
			"text/html", map[string][]string{"level": {"1"}},
			"text/html", map[string][]string{},
			mediaRangeMatch{nil, false, 1, 1, 0},
		},
		{
			"text/html", map[string][]string{},
			"text/html", map[string][]string{"level": {"1"}},
			mediaRangeMatch{nil, true, 1, 1, 0},
		},
		{
			"text/html", map[string][]string{"level": {"1"}},
			"text/html", map[string][]string{"level": {"1"}, "foo": {"bar"}},
			mediaRangeMatch{nil, true, 1, 1, 1},
		},
		{
			"text/html", map[string][]string{"level": {"1"}, "charset": {"utf-8"}},
			"text/html", map[string][]string{"level": {"1"}, "charset": {"utf-8"}, "foo": {"bar"}},
			mediaRangeMatch{nil, true, 1, 1, 2},
		},
	}
	for _, test := range tests {
		mr := MediaRange{Range: test.Range, Params: test.Params}
		test.Match.MediaRange = &mr
		match := mr.match(test.ContentType, test.ContentParams)
		if match != test.Match {
			t.Errorf("{Range:%v Params:%v}.match(%v, %v) = %v; want %v", test.Range, test.Params, test.ContentType, test.ContentParams, match, test.Match)
		}
	}
}

func TestMediaRangeMatchLess(t *testing.T) {
	tests := []struct {
		A, B mediaRangeMatch
		Less bool
	}{
		{mediaRangeMatch{}, mediaRangeMatch{}, false},
		{mediaRangeMatch{Valid: true}, mediaRangeMatch{}, true},
		{mediaRangeMatch{}, mediaRangeMatch{Valid: true}, false},
		{mediaRangeMatch{nil, true, 0, 0, 0}, mediaRangeMatch{nil, true, 0, 0, 0}, false},
		{mediaRangeMatch{nil, true, 1, 0, 0}, mediaRangeMatch{nil, true, 0, 0, 0}, true},
		{mediaRangeMatch{nil, true, 0, 0, 0}, mediaRangeMatch{nil, true, 1, 0, 0}, false},
		{mediaRangeMatch{nil, true, 1, 1, 0}, mediaRangeMatch{nil, true, 0, 0, 0}, true},
		{mediaRangeMatch{nil, true, 0, 0, 0}, mediaRangeMatch{nil, true, 1, 1, 0}, false},
		{mediaRangeMatch{nil, true, 0, 0, 1}, mediaRangeMatch{nil, true, 0, 0, 0}, true},
		{mediaRangeMatch{nil, true, 0, 0, 0}, mediaRangeMatch{nil, true, 0, 0, 1}, false},
		{mediaRangeMatch{nil, true, 1, 1, 1}, mediaRangeMatch{nil, true, 0, 0, 0}, true},
		{mediaRangeMatch{nil, true, 0, 0, 0}, mediaRangeMatch{nil, true, 1, 1, 1}, false},
	}

	matches := make(mediaRangeMatches, 2)
	for _, test := range tests {
		matches[0] = test.A
		matches[1] = test.B
		result := matches.Less(0, 1)
		if result != test.Less {
			t.Errorf("%v < %v = %t; want %t", test.A, test.B, result, test.Less)
		}
	}
}
