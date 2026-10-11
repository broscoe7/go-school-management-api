package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/broscoe7/go-school-api/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}
		jwtSecret := os.Getenv("JWT_SECRET")

		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method, %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			utils.ErrorHandler(err, "")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if !parsedToken.Valid {
			http.Error(w, "Invalid login token", http.StatusUnauthorized)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid login token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), utils.ContextKey("role"), claims["role"])
		ctx = context.WithValue(ctx, utils.ContextKey("userID"), claims["uid"])
		ctx = context.WithValue(ctx, utils.ContextKey("username"), claims["user"])
		ctx = context.WithValue(ctx, utils.ContextKey("expiresAt"), claims["exp"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
