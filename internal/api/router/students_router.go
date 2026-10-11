package router

import (
	"net/http"

	"github.com/broscoe7/go-school-api/internal/api/handlers"
)

func studentsRouter(mux *http.ServeMux) {
	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetStudentHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentsHandler)
	mux.HandleFunc("PUT /students/{id}", handlers.UpdateStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchStudentHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteStudentHandler)
}
