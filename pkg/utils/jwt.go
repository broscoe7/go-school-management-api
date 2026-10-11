package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SignToken(userID int, username, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiresIn := os.Getenv("JWT_EXPIRES_IN")
	claims := jwt.MapClaims{
		"uid":  userID,
		"user": username,
		"role": role,
	}

	if jwtExpiresIn != "" {
		duration, err := time.ParseDuration(jwtExpiresIn)
		if err != nil {
			return "", ErrorHandler(err, "Internal error")
		}
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrorHandler(err, "Internal error")
	}
	return token, nil
}
