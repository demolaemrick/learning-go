package main

import (
	"go-microservice/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	l := log.New(os.Stdout, "product-api", log.LstdFlags)
	helloHandler := handlers.NewHello(l)

	serveMux := http.NewServeMux()
	serveMux.Handle("/", helloHandler)

	log.Fatal(http.ListenAndServe(":8080", serveMux))
}
