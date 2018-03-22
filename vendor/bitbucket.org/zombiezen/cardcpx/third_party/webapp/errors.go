package webapp

import (
	"errors"
	"strconv"
)

// A MultiError is returned by operations that have errors on particular elements.
// This is functionally identical to appengine.MultiError.
type MultiError []error

func (e MultiError) Error() string {
	msg, n := "", 0
	for _, err := range e {
		if err != nil {
			if n == 0 {
				msg = err.Error()
			}
			n++
		}
	}
	switch n {
	case 0:
		return "0 errors"
	case 1:
		return msg
	case 2:
		return msg + " (and 1 other error)"
	}

	s := []byte(msg)
	s = append(s, []byte(" (and ")...)
	s = strconv.AppendInt(s, int64(n-1), 10)
	s = append(s, []byte(" other errors)")...)
	return string(s)
}

type parseError struct {
	Expected string
	In       string
	Found    rune
	EOF      bool
}

func (e *parseError) Error() string {
	var s []byte
	s = append(s, []byte("expected ")...)
	s = append(s, []byte(e.Expected)...)
	if e.In != "" {
		s = append(s, []byte(" in ")...)
		s = append(s, []byte(e.In)...)
	}
	s = append(s, []byte(", found ")...)
	if !e.EOF {
		s = strconv.AppendQuoteRune(s, e.Found)
	} else {
		s = append(s, []byte("EOF")...)
	}
	return string(s)
}

type qvalueError struct {
	QValue string
	Err    error
}

func (e *qvalueError) Error() string {
	return "qvalue " + strconv.Quote(e.QValue) + ": " + e.Err.Error()
}

var errQValueRange = errors.New("must be >0.0 and <=1.0")
