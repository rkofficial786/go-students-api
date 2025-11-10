package main

import (
	"encoding/json"
	"fmt"
	mw "school-api/internal/api/middlewares"
	"slices"
	"strings"

	"time"

	"net/http"
)

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	City string `json:"city"`
}

type Teacher struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Subject   string `json:"subject"`
	Class     string `json:"class"`
}

var teachers = []Teacher{
	{ID: "1", FirstName: "John", LastName: "Doe", Subject: "Math", Class: "10A"},
	{ID: "2", FirstName: "Jane", LastName: "Smith", Subject: "Science", Class: "10B"},
	{ID: "3", FirstName: "Emily", LastName: "Johnson", Subject: "History", Class: "10C"},
	{ID: "4", FirstName: "Michael", LastName: "Brown", Subject: "English", Class: "10D"},
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	fmt.Println(idStr)

	if idStr != "" {
		response := struct {
			Status string    `json:"status"`
			Count  int       `json:"count"`
			Data   []Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teachers),
			Data:   teachers,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	teacher := slices.DeleteFunc(teachers, func(t Teacher) bool {
		return t.ID != idStr

	})

	if len(teacher)>0{
		response := struct {
			Status string    `json:"status"`
			Count  int       `json:"count"`
			Data   []Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teacher),
			Data:   teacher,
		}
	}
   

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func teachersHandlers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		getTeachersHandler(w, r)
	case http.MethodPost:
		fmt.Fprintln(w, "Hello, World! Post Teachers")
		return
	case http.MethodDelete:
		fmt.Fprintln(w, "Hello, World! Delete Teachers")
		return
	case http.MethodPut:
		fmt.Fprintln(w, "Hello, World! Put Teachers")
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	port := ":5173"

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World! new")
	})
	mux.HandleFunc("/teachers/", teachersHandlers)

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
	err := http.ListenAndServe(port, handler)
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
