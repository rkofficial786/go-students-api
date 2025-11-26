package main

import (
	"fmt"
	"school-api/internal/api/handlers"
	mw "school-api/internal/api/middlewares"
	"school-api/internal/repositeries/sqlconnect"
	"time"

	"net/http"

	"github.com/joho/godotenv"
)

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	City string `json:"city"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
		return
	}

	_, err = sqlconnect.ConnectDB()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	port := ":5173"

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World! new")
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	
	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherByIdHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("DELETE /teachers/bulk", handlers.DeleteMupltipleTeachersHandler)

	rl := mw.NewRateLimiter(400, time.Minute)
	handler := applyMiddlewares(
		mux,
		mw.Cors,
		rl.Middleware,
		mw.Compression,
		mw.ResponseTime,
		mw.SecurityHeaders,
	)

	fmt.Println("server is running on port", port)
	err = http.ListenAndServe(port, handler)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}

type Middleware func(http.Handler) http.Handler

func applyMiddlewares(h http.Handler, middlewares ...Middleware) http.Handler {

	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
