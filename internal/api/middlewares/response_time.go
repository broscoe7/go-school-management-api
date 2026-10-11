package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// create a custom ResponseWriter to capture the status code
		wrappedWriter := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		// calculate duration
		duration := time.Since(start)
		// log request details
		fmt.Printf("Method: %s, URL: %s, Status: %d, Duration: %v\n", r.Method, r.URL, wrappedWriter.status, duration.String())
	})
}

// the struct and method below overide the WriteHeader method of the http.ResponseWriter interface.
// Anytime w.Write (which calls WriteHeader behind the scenes) or w.WriteHeader is called, this will be used instead, and the status code will be captured.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
