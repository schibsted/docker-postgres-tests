package webapp

import (
	"testing"
)

func TestTokenChars(t *testing.T) {
	for c := rune(0); c < 0x10ffff; c++ {
		expected := c < 127 && c > 31 &&
			c != '(' && c != ')' && c != '<' && c != '>' && c != '@' &&
			c != ',' && c != ';' && c != ':' && c != '\\' && c != '"' &&
			c != '/' && c != '[' && c != ']' && c != '?' && c != '=' &&
			c != '{' && c != '}' && c != ' ' && c != '\t'
		if found := isTokenChar(c); found != expected {
			t.Errorf("%q found=%t, want %t", c, found, expected)
		}
	}
}

func TestQuoteHTTP(t *testing.T) {
	tests := []struct {
		Value  string
		Quoted string
	}{
		{"", `""`},
		{"a", "a"},
		{"abc", "abc"},
		{"Hello, World!", `"Hello, World!"`},
		{`C:\`, `"C:\\"`},
		{`"foo"`, `"\"foo\""`},
	}

	for _, test := range tests {
		q := quoteHTTP(test.Value)
		if q != test.Quoted {
			t.Errorf("quoteHTTP(%q) = %q; want %q", test.Value, q, test.Quoted)
		}
	}
}
