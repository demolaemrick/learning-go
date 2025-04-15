package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context key type to avoid collisions
type contextKey string

const userIDKey contextKey = "user_id"

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
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			sendError(w, "Unauthorized - missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			sendError(w, "Unauthorized - invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			sendError(w, "Unauthorized - invalid claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(float64) // JSON numbers are float64
		if !ok {
			sendError(w, "Unauthorized - invalid user ID", http.StatusUnauthorized)
			return
		}

		// Store user_id in context
		ctx := context.WithValue(r.Context(), userIDKey, int(userID))
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
