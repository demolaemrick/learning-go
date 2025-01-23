package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("Hello world")
		d, err := io.ReadAll(r.Body)
		if err != nil {
			// rw.WriteHeader(http.StatusBadRequest)
			// rw.Write([]byte("Oops"))
			http.Error(rw, "Oops", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(rw, "Hello %s\n", d)
	})

	http.HandleFunc("/goodbye", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Goodbye world")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
