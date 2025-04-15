package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload" // Auto-load .env file
	_ "github.com/lib/pq"                 // PostgreSQL driver
)

var db *sql.DB // Global DB handle so it is accessible to all handlers

func main() {
	PORT := ":9000"

	// Load database credentials from env
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Validate env vars
	if dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("DB_USER, DB_PASSWORD, and DB_NAME environment variables are required")
	}
	if dbHost == "" {
		dbHost = "localhost" // Default
	}
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}
	if dbSSLMode == "" {
		dbSSLMode = "disable" // Default for local dev
	}

	// Connect to PostgreSQL
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", dbUser, dbPassword, dbName, dbHost, dbPort, dbSSLMode)

	var err error
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	defer db.Close() // close it when main exit

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)

	// Public routes
	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(AuthMiddleware)
	protected.HandleFunc("/users", getUsers).Methods("GET")
	protected.HandleFunc("/tasks", getTasks).Methods("GET")
	protected.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	protected.HandleFunc("/tasks", createTask).Methods("POST")
	protected.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	protected.HandleFunc("/tasks/{id}/toggle", toggleTaskCompletion).Methods("PATCH")
	protected.HandleFunc("/tasks", deleteAllTasks).Methods("DELETE")

	log.Printf("Starting server on port %v...", PORT)
	err = http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
