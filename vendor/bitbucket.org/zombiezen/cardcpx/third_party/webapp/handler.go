package webapp

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

// URLError records an error and the URL that caused it.
type URLError struct {
	URL *url.URL
	Err error
}

func (e *URLError) Error() string {
	return e.URL.String() + ": " + e.Err.Error()
}

// NotFound is an HTTP 404 error.
var NotFound = errors.New("not found")

// IsNotFound reports whether an error is a NotFound error.
func IsNotFound(e error) bool {
	if e == NotFound {
		return true
	}
	if e, ok := e.(*URLError); ok && e.Err == NotFound {
		return true
	}
	return false
}

// A ResponseBuffer is a ResponseWriter that stores the data written to it in memory.  The zero value is an empty response.
type ResponseBuffer struct {
	bytes.Buffer
	code   int
	header http.Header
	sent   http.Header
}

// StatusCode returns the status code sent with WriteHeader or 0 if WriteHeader has not been called.
func (br *ResponseBuffer) StatusCode() int {
	return br.code
}

// Size returns the number of bytes in the buffer.
func (br *ResponseBuffer) Size() int64 {
	return int64(br.Len())
}

// HeaderSent returns the headers that were sent when WriteHeader was called or nil if WriteHeader has not been called.
func (br *ResponseBuffer) HeaderSent() http.Header {
	return br.sent
}

// Copy sends br's data to another ResponseWriter.
func (br *ResponseBuffer) Copy(w http.ResponseWriter) error {
	for k, v := range br.sent {
		w.Header()[k] = v
	}
	w.WriteHeader(br.code)
	_, err := io.Copy(w, br)
	return err
}

func (br *ResponseBuffer) Header() http.Header {
	if br.header == nil {
		br.header = make(http.Header)
	}
	return br.header
}

func (br *ResponseBuffer) WriteHeader(code int) {
	if br.sent == nil {
		br.code = code
		br.sent = make(http.Header, len(br.header))
		for k, v := range br.header {
			v2 := make([]string, len(v))
			copy(v2, v)
			br.sent[k] = v2
		}
	}
}

func (br *ResponseBuffer) Write(p []byte) (int, error) {
	br.WriteHeader(http.StatusOK)
	return br.Buffer.Write(p)
}

// ResponseStats is a ResponseWriter that records statistics about a response.
type ResponseStats struct {
	w    http.ResponseWriter
	code int
	size int64
}

// NewResponseStats returns a new ResponseStats that writes to w.
func NewResponseStats(w http.ResponseWriter) *ResponseStats {
	return &ResponseStats{w: w}
}

// StatusCode returns the status code sent with WriteHeader or 0 if WriteHeader has not been called.
func (r *ResponseStats) StatusCode() int {
	return r.code
}

// Size returns the number of bytes written to the underlying ResponseWriter.
func (r *ResponseStats) Size() int64 {
	return r.size
}

func (r *ResponseStats) Header() http.Header {
	return r.w.Header()
}

func (r *ResponseStats) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
	r.code = statusCode
}

func (r *ResponseStats) Write(p []byte) (n int, err error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	n, err = r.w.Write(p)
	r.size += int64(n)
	return
}

// Logger logs all HTTP requests sent to an http.Handler.
type Logger struct {
	http.Handler
}

func (l Logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method, path, remote := req.Method, req.URL.Path, req.RemoteAddr
	stats := NewResponseStats(w)
	l.Handler.ServeHTTP(stats, req)
	log.Printf("%s %s %d %d %s", method, path, stats.StatusCode(), stats.Size(), remote)
}
