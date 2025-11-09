package main

import (
	"fmt"
	"strings"

	"net/http"
)

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	City string `json:"city"`
}

func teachersHandlers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println(r.URL.Path, "paths")
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userId := strings.TrimSuffix(path, "/")
		fmt.Println(userId, "userId")

		queryParams := r.URL.Query() 
		sortBy:=queryParams.Get("sortBy")
		fmt.Println(sortBy, "sortBy")
		if userId != "" {
			fmt.Fprintf(w, "Hello, World! Get Teacher with IDs: %s\n", userId)
			return
		}
		fmt.Fprintln(w, "Hello, World! Get Teachers")
		return
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World! new")
	})

	http.HandleFunc("/teachers/", teachersHandlers)

	fmt.Println("server is running on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
