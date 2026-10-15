package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "Hijack not supported", http.StatusInternalServerError)
				return
			}
			clientConn, _, err := hijacker.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer clientConn.Close()

			// SOCKS4専用の処理を実行
			handleSOCKS4(clientConn)
		}),
	}

	fmt.Printf("SOCKS4/4a Dedicated Server started on port %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleSOCKS4(c net.Conn) {
	buf := make([]byte, 8)
	
	// 1. SOCKS4ヘッダー（8バイト）の読み込み
	// [VN(1) | CD(1) | DSTPORT(2) | DSTIP(4)]
	if _, err := io.ReadFull(c, buf); err != nil {
		return
	}

	// バージョンが0x04（SOCKS4）でなければ即切断
	if buf[0] != 0x04 {
		return
	}

	// コマンドが0x01（CONNECT）でなければ即切断
	if buf[1] != 0x01 {
		return
	}

	port := binary.BigEndian.Uint16(buf[2:4])
	ip := net.IP(buf[4:8])
	var host string

	// USERIDの読み飛ばし（ヌル文字 0x00 が来るまで読む）
	userIdBuf := make([]byte, 1)
	for {
		if _, err := c.Read(userIdBuf); err != nil || userIdBuf[0] == 0x00 {
			break
		}
	}

	// SOCKS4aの対応（IPが 0.0.0.x の場合、後ろにドメイン名が続く仕様）
	if ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] != 0 {
		domainBuf := make([]byte, 1)
		var domainBytes []byte
		for {
			if _, err := c.Read(domainBuf); err != nil || domainBuf[0] == 0x00 {
				break
			}
			domainBytes = append(domainBytes, domainBuf[0])
		}
		host = string(domainBytes) // リモート名前解決用のドメイン名を取得
	} else {
		host = ip.String() // 通常のIPアドレス
	}

	dest := fmt.Sprintf("%s:%d", host, port)

	// 2. 宛先サーバーへ接続（Render側から直通で繋ぎにいきます）
	remote, err := net.Dial("tcp", dest)
	if err != nil {
		// 接続失敗のレスポンス（拒否: 0x5b）
		c.Write([]byte{0x00, 0x5b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}
	defer remote.Close()

	// 接続成功のレスポンス（許可: 0x5a）
	c.Write([]byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// 3. データの中継（双方向コピー）
	errc := make(chan error, 2)
	go func() {
		_, err := io.Copy(remote, c)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(c, remote)
		errc <- err
	}()
	<-errc
}
