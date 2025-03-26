package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	PORT := ":9000"
	router := mux.NewRouter()

	//Define routes
	router.HandleFunc("/tasks", getTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	router.HandleFunc("/tasks", createTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	router.HandleFunc("/tasks/{id}/toggle", toggleTaskCompletion).Methods("PATCH")

	log.Printf("Starting server on port %v...", PORT)

	err := http.ListenAndServe(PORT, router)

	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
