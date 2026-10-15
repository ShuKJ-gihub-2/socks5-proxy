package main

import (
	"log"
	"net/http"
	"os"
	"github.com/armon/go-socks5"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conf := &socks5.Config{}
	socksServer, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on port %s...", port)
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	if err := socksServer.ListenAndServe("tcp", "0.0.0.0:1080"); err != nil {
		log.Fatal(err)
	}
}
