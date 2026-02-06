package main

import (
	"log"

	"github.com/demolaemrick/learning-go/blog-api/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
