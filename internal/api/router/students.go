package router

import (
	"net/http"
	"school-api/internal/api/handlers"
)

func RegisterStudentsRoutes (mux *http.ServeMux){

	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetStudentByIdHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PUT /students/{id}", handlers.UpdateStudentHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteStudentHandler)
	mux.HandleFunc("GET /students/teachers", handlers.GetStudentOfTeachers)

}