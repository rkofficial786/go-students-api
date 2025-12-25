package router

import (
	"net/http"
	"school-api/internal/api/handlers"
)

func RegisterExecRoutes(mux *http.ServeMux) {

	// Collection routes
	mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
	mux.HandleFunc("POST /execs", handlers.AddExecHandler)

	// Single exec routes
	mux.HandleFunc("GET /execs/{id}", handlers.GetExecByIdHandler)
	mux.HandleFunc("PUT /execs/{id}", handlers.UpdateExecHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteExecHandler)

	// Password & auth routes
	mux.HandleFunc("POST /execs/{id}/updatePassword", handlers.UpdatePasswordHandler)
	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	mux.HandleFunc("POST /execs/forgotPassword", handlers.ForgotPasswordHandler)
	mux.HandleFunc("POST /execs/reset/password/{resetcode}", handlers.ResetPasswordHandler)
}
