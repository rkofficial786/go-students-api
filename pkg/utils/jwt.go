package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SignToken(userId int, username, role string) (string, error) {
	id := strconv.Itoa(userId)
	jwtSecret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"uid":  id,
		"user": username,
		"role": role,
		"exp":  jwt.NewNumericDate(time.Now().Add(20 * time.Second)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		fmt.Println("JWT error", err)
		return "", nil
	}

	return signedToken, nil

}
