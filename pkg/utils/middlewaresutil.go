package utils

import "net/http"

type Middleware func(http.Handler) http.Handler

// the last middleware will be the first one applied to incoming requests
func ApplyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
