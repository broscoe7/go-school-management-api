package router

import (
	"net/http"

	"github.com/broscoe7/go-school-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()
	teachersRouter(mux)
	studentsRouter(mux)
	execsRouter(mux)

	mux.HandleFunc("GET /", handlers.RootHandler)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}
