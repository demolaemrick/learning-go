package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB // Global DB handle so it is accessible to all handlers

func main() {
	PORT := ":9000"
	var err error

	// Connect to PostgreSQL
	connStr := "user=taskuser password=secret dbname=tasks sslmode=disable"
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
	router.Use(LoggingMiddlware)

	//Define routes
	router.HandleFunc("/tasks", getTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	router.HandleFunc("/tasks", createTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	router.HandleFunc("/tasks/{id}/toggle", toggleTaskCompletion).Methods("PATCH")
	router.HandleFunc("/tasks", deleteAllTasks).Methods("DELETE")

	log.Printf("Starting server on port %v...", PORT)

	err = http.ListenAndServe(PORT, router)

	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
