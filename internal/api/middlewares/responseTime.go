package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTime(next http.Handler) http.Handler{

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()


		wrappedWriter:= &responseWriter{
			ResponseWriter: w,status: http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter,r)

		duration:=time.Since(start)


		fmt.Printf("Method: %s, URL: %s, Status: %d, Response Time: %v\n", r.Method, r.URL.Path, wrappedWriter.status, duration)


	})
}



type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) writeHeader(code int){
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}