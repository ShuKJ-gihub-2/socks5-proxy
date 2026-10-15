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

	// 世界中のエンジニアが検証に使う、超強力なプロキシエンジンを起動
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// すべてのCONNECT通信（HTTPS中継）をノーチェックで直通させる
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	fmt.Printf("Goproxy Engine started on port %s\n", port)
	if err := http.ListenAndServe(":"+port, proxy); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
