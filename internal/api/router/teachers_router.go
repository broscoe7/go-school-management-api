package router

import (
	"net/http"

	"github.com/broscoe7/go-school-api/internal/api/handlers"
)

func teachersRouter(mux *http.ServeMux) {
	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeachersHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHandler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchTeacherHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetStudentsByTeacherID)
	mux.HandleFunc("GET /teachers/{id}/student-count", handlers.GetStudentsCountByTeacherID)
}
