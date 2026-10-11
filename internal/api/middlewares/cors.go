package middlewares

import (
	"net/http"
	"slices"
)

var allowedOrigins = []string{
	"https://my-origin-url.com",
	"https://localhost:3000",
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if !slices.Contains(allowedOrigins, origin) {
			http.Error(w, "Not allowed by CORS", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		// says that the browser can include credentials (e.g. cookies, authentication headers, client-side SSL certificates) in cross-origin requests
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// says how long (in seconds) the browser can cache the preflight response (with the CORS conditions); improves performance by reducing OPTIONS requests
		w.Header().Set("Access-Control-Max-Age", "3600")

		// handle a preflight request
		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
