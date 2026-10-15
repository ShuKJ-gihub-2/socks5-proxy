package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"github.com/armon/go-socks5"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	conf := &socks5.Config{}
	socksServer, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				w.WriteHeader(http.StatusOK)
				hijacker, ok := w.(http.Hijacker)
				if !ok {
					return
				}
				clientConn, _, err := hijacker.Hijack()
				if err != nil {
					return
				}
				defer clientConn.Close()

				// ここでHTTPの通信を強引にSOCKS5の生TCP通信に変換します
				socksConn, err := net.Dial("tcp", "127.0.0.1:1080")
				if err != nil {
					return
				}
				defer socksConn.Close()

				go io.Copy(socksConn, clientConn)
				io.Copy(clientConn, socksConn)
				return
			}

			// Renderのヘルスチェック（生存確認）用
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}),
	}

	log.Printf("Server starting on port %s...", port)
	go func() {
		if err := socksServer.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
