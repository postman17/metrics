package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w           http.ResponseWriter
	zw          *gzip.Writer // Будет инициализирован только при необходимости
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.initGzipIfNeeded()
	}

	if c.zw != nil {
		return c.zw.Write(b)
	}
	return c.w.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.initGzipIfNeeded()
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) initGzipIfNeeded() {
	c.wroteHeader = true

	cType := c.w.Header().Get("Content-Type")

	if strings.Contains(cType, "application/json") || strings.Contains(cType, "text/html") {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.zw = gzip.NewWriter(c.w)
	}
}

func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GZIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
