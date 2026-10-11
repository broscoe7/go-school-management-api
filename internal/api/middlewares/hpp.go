package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	Whitelist                   []string // list of accepted parameters
}

func Hpp(options HPPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, options.CheckBodyOnlyForContentType) {
				// filter body parameters
				filterBodyParams(r, options.Whitelist)
			}
			if options.CheckQuery && r.URL.Query() != nil {
				// filter query parameters
				filterQueryParams(r, options.Whitelist)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range r.PostForm {
		if len(v) > 1 {
			r.PostForm.Set(k, v[0]) // first value
			// r.PostForm.Set(k, v[len(v)-1]) // last value
		}
		if !slices.Contains(whitelist, k) {
			delete(r.PostForm, k)
		}
	}
}

func filterQueryParams(r *http.Request, whitelist []string) {
	queries := r.URL.Query()
	for k, v := range queries {
		if len(v) > 1 {
			queries.Set(k, v[0])
		}
		if !slices.Contains(whitelist, k) {
			queries.Del(k)
		}
	}
	r.URL.RawQuery = queries.Encode()
}
