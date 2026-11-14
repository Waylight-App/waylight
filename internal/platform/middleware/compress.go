package middleware

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
)

// minCompressBytes: smaller answers are sent as they are (gzip would
// barely help and costs a little CPU).
const minCompressBytes = 1024

var gzipPool = sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed); return w }}

// Compress gzips JSON answers of the API for clients that accept it (#318).
// Static files are pre-compressed by the build and served by the SPA
// handler; streams (Flush) keep working.
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") || r.Method == http.MethodHead || !acceptsGzip(r) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		cw := &compressWriter{ResponseWriter: w, code: http.StatusOK}
		defer cw.close()
		next.ServeHTTP(cw, r)
	})
}

// acceptsGzip reports whether the client lists gzip with a non-zero weight.
func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		coding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(coding), "gzip") {
			continue
		}
		q := strings.ReplaceAll(strings.ToLower(params), " ", "")
		return q != "q=0" && q != "q=0.0" && q != "q=0.00" && q != "q=0.000"
	}
	return false
}

// compressWriter buffers the start of an answer and decides on gzip once
// it knows the content type and that the answer is big enough.
type compressWriter struct {
	http.ResponseWriter
	code        int
	wroteHeader bool // WriteHeader was called by the handler
	decided     bool
	gz          *gzip.Writer
	buf         []byte
}

func (c *compressWriter) WriteHeader(code int) {
	if c.decided || c.wroteHeader {
		return
	}
	c.code, c.wroteHeader = code, true
	if code < 200 || code == http.StatusNoContent || code == http.StatusNotModified {
		c.decide(false)
	}
}

func (c *compressWriter) compressible() bool {
	h := c.Header()
	return strings.HasPrefix(h.Get("Content-Type"), "application/json") && h.Get("Content-Encoding") == ""
}

// decide sends the header, with gzip or without, and the buffered bytes.
func (c *compressWriter) decide(gz bool) {
	if c.decided {
		return
	}
	c.decided = true
	if gz {
		c.Header().Del("Content-Length")
		c.Header().Set("Content-Encoding", "gzip")
		c.gz = gzipPool.Get().(*gzip.Writer)
		c.gz.Reset(c.ResponseWriter)
	}
	c.ResponseWriter.WriteHeader(c.code)
	if len(c.buf) > 0 {
		buf := c.buf
		c.buf = nil
		_, _ = c.write(buf)
	}
}

func (c *compressWriter) write(p []byte) (int, error) {
	if c.gz != nil {
		return c.gz.Write(p)
	}
	return c.ResponseWriter.Write(p)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.decided {
		return c.write(p)
	}
	if !c.compressible() {
		c.decide(false)
		return c.write(p)
	}
	c.buf = append(c.buf, p...)
	if len(c.buf) >= minCompressBytes {
		c.decide(true)
	}
	return len(p), nil
}

// Flush sends what is buffered (streams such as AI summaries).
func (c *compressWriter) Flush() {
	if !c.decided {
		c.decide(len(c.buf) >= minCompressBytes && c.compressible())
	}
	if c.gz != nil {
		_ = c.gz.Flush()
	}
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack passes through for handlers that need the raw connection.
func (c *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := c.ResponseWriter.(http.Hijacker); ok && !c.decided {
		c.decided = true
		return h.Hijack()
	}
	return nil, nil, errors.New("compress: hijack not supported")
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (c *compressWriter) Unwrap() http.ResponseWriter { return c.ResponseWriter }

func (c *compressWriter) close() {
	if !c.decided {
		c.decide(false) // small answer: sent as it is
	}
	if c.gz != nil {
		_ = c.gz.Close()
		c.gz.Reset(nil)
		gzipPool.Put(c.gz)
		c.gz = nil
	}
}
