package main

import (
	"log"
	"net/http"
	"os"

	"github.com/armon/go-socks5"
	"://github.com"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	// Renderが指定するポート（デフォルトは8080）を取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 認証用トークン（環境変数から取得）
	secretToken := os.Getenv("PROXY_TOKEN")
	if secretToken == "" {
		log.Fatal("PROXY_TOKEN environment variable is required")
	}

	// SOCKS5サーバーの初期化
	conf := &socks5.Config{}
	socksServer, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	// WebSocketエンドポイントのハンドラー
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		// 簡易的なトークン認証
		token := r.URL.Query().Get("token")
		if token != secretToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// WebSocketへアップグレード
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer ws.Close()

		// WebSocket接続をネット標準のConnインターフェースにラップしてSOCKSへ流す
		// ※実用にはWS/TCP相互変換をサポートするライブラリやラッパーを噛ませるか、
		// 既存のOSS（shadowsocks+v2ray-plugin、gost等）の利用が一般的です。
	})

	// ヘルスチェック用
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Response) {
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
