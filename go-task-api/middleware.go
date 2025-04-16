package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go-task-api/handlers"

	"github.com/golang-jwt/jwt/v5"
)

// LoggingMiddleware unchanged
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, duration)
	})
}

// AuthMiddleware verifies JWT
func AuthMiddleware(h *handlers.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				h.SendError(w, "Unauthorized - missing or invalid token", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return handlers.GetJWTSecret(), nil
			})
			if err != nil || !token.Valid {
				h.SendError(w, "Unauthorized - invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				h.SendError(w, "Unauthorized - invalid claims", http.StatusUnauthorized)
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				h.SendError(w, "Unauthorized - invalid user ID", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", int(userID))
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
