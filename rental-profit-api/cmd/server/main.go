package main

import (
	"log"
	"net/http"
)

func main() {

	log.Println("Starting rental profit maximization server on port 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Rental profit maximization server started on port 8080")
}