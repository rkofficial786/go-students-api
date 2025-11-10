package middlewares

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
return  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	

 w.Header().Set("X-DNS-Prefetch-Control", "off")
 w.Header().Set("X-Frame-Options", "SAMEORIGIN")
 w.Header().Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")
 w.Header().Set("X-Download-Options", "noopen")
 w.Header().Set("X-Content-Type-Options", "nosniff")
 w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
 w.Header().Set("Referrer-Policy", "no-referrer")
 w.Header().Set("X-XSS-Protection", "0") 
 

next.ServeHTTP(w, r)


})
}