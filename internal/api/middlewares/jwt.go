package middlewares

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var jwtSecret = os.Getenv("JWT_SECRET")

		var hmacSampleSecret = []byte(jwtSecret)

		cookie, err := r.Cookie("Bearer")

		if err != nil {
			http.Error(w, "Unauthorized: token missing", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			// Ensure token method is HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return hmacSampleSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {

			http.Error(w, "Unauthorized: Unable to extract", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "role", claims["role"])

		ctx = context.WithValue(ctx, "userId", claims["uid"])
		ctx = context.WithValue(ctx, "username", claims["user"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
