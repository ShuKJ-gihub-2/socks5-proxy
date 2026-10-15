package main

import (
	"fmt"
	"net/http"
	"os"
	"://github.com"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	// 世界標準のプロキシエンジンを起動
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// すべてのトンネル通信（HTTPS）をそのまま直通させる
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	fmt.Printf("Goproxy Engine started on port %s\n", port)
	if err := http.ListenAndServe(":"+port, proxy); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
