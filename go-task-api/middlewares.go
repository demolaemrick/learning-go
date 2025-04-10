package main

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)

		duration := time.Since(start)

		log.Printf("Completed %s %s in %v", r.Method, r.URL, duration)
	})
}
