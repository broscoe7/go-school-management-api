package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// wrap http.ResponseWriter in a Gzip Writer
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w = &gzipResponseWriter{w, gz}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter wraps the http.ResponseWriter to send gzipped responses
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

// provides a custom implementation for the ResponseWriter interface
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}
