package router

import (
	"net/http"
	"school-api/internal/api/handlers"
)

func RegisterTeachersRoutes (mux *http.ServeMux){

	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherByIdHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("DELETE /teachers/bulk", handlers.DeleteMupltipleTeachersHandler)

}