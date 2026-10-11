package router

import (
	"net/http"

	"github.com/broscoe7/go-school-api/internal/api/handlers"
)

func execsRouter(mux *http.ServeMux) {
	mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
	mux.HandleFunc("GET /execs/{id}", handlers.GetExecHandler)
	mux.HandleFunc("POST /execs", handlers.AddExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PatchExecHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteExecHandler)
	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	mux.HandleFunc("POST /execs/forgot-password", handlers.ForgotPasswordHandler)
	mux.HandleFunc("POST /execs/password/reset/{resetcode}", handlers.ResetPasswordHandler)
	mux.HandleFunc("POST /execs/{id}/update-password", handlers.UpdatePasswordHandler)
}
