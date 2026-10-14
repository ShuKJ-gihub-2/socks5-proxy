package main

import (
	"log"
	"os"
	"://github.com"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1080"
	}
	server := socks5.NewServer()
	log.Printf("SOCKS5 server starting on port %s", port)
	if err := server.ListenAndServe("tcp", ":"+port); err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
}
