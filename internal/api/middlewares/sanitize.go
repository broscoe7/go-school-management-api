package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/broscoe7/go-school-api/pkg/utils"
	"github.com/microcosm-cc/bluemonday"
)

func XSSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sanitize the URL path
		sanitizedPath, err := clean(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// sanitize query parameters
		params := r.URL.Query()
		sanitizedQueries := make(map[string][]string)
		for k, v := range params {
			sanitizedKey, err := clean(k)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var sanitizedValues []string
			for _, s := range v {
				cleanValue, err := clean(s)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				sanitizedValues = append(sanitizedValues, cleanValue.(string))
			}
			sanitizedQueries[sanitizedKey.(string)] = sanitizedValues
		}

		// overwrite path and queries with the cleaned values
		r.URL.Path = sanitizedPath.(string)
		r.URL.RawQuery = url.Values(sanitizedQueries).Encode()

		// sanitize request body
		if r.Header.Get("Content-Type") != "application/json" && r.Header.Get("Content-Type") != "" {
			log.Printf("Received request with unsupported Content-Type: %s. Expected application/json.", r.Header.Get("Content-Type"))
			http.Error(w, "Unsupported content type. Please use application/json.", http.StatusUnsupportedMediaType)
			return
		}
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, utils.ErrorHandler(err, "Error handling request body").Error(), http.StatusBadRequest)
				return
			}
			bodyString := strings.TrimSpace(string(bodyBytes))

			// overwrite request body
			r.Body = io.NopCloser(bytes.NewReader([]byte(bodyString)))

			if len(bodyString) > 0 {
				var inputData any
				err := json.NewDecoder(bytes.NewReader([]byte(bodyString))).Decode(&inputData)
				if err != nil {
					http.Error(w, utils.ErrorHandler(err, "Invalid JSON body").Error(), http.StatusBadRequest)
					return
				}
				sanitizedData, err := clean(inputData)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				sanitizedBody, err := json.Marshal(sanitizedData)
				if err != nil {
					http.Error(w, utils.ErrorHandler(err, "Error sanitizing body").Error(), http.StatusBadRequest)
					return
				}
				// overwrite request body one more time
				r.Body = io.NopCloser(bytes.NewReader([]byte(sanitizedBody)))
			}

		}
		next.ServeHTTP(w, r)
	})
}

// Clean sanitizes input data to prevent XSS attacks
func clean(data any) (any, error) {
	switch d := data.(type) {
	case map[string]any:
		for k, v := range d {
			d[k] = sanitizeValue(v)
		}
		return d, nil
	case []any:
		for i, v := range d {
			d[i] = sanitizeValue(v)
		}
		return d, nil
	case string:
		return sanitizeString(d), nil
	default:
		return nil, utils.ErrorHandler(fmt.Errorf("unsupported data type: %T", data), fmt.Sprintf("unsupported data type: %T", data))
	}
}

// we use recursion to handle nested data types
func sanitizeValue(value any) any {
	switch d := value.(type) {
	case string:
		return sanitizeString(d)
	case map[string]any:
		for k, v := range d {
			d[k] = sanitizeValue(v)
		}
		return d
	case []any:
		for i, v := range d {
			d[i] = sanitizeValue(v)
		}
		return d
	default:
		return d
	}
}

func sanitizeString(value string) string {
	return bluemonday.UGCPolicy().Sanitize(value)
}
