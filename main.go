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

	// RenderのHTTPS(443)から届く通信を、そのままSOCKS5のパケットとして処理するサーバー
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ChromebookがSOCKSとして接続してきた通信を乗っ取る（Hijack）
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

			// 横取りした通信をそのままSOCKS5プロトコルとして処理する
			handleSOCKS5(clientConn)
		}),
	}

	fmt.Printf("Hybrid SOCKS5 Server started on port %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleSOCKS5(c net.Conn) {
	buf := make([]byte, 257)
	// 1. 認証リクエストの読み込み
	if _, err := io.ReadFull(c, buf[:2]); err != nil {
		return
	}
	// SOCKS4(0x04)が来てもSOCKS5(0x05)が来ても両方受け付けるハイブリッド仕様
	version := buf[0]
	if version != 0x05 && version != 0x04 {
		return
	}

	if version == 0x05 {
		// SOCKS5の処理
		nmethods := int(buf[1])
		if _, err := io.ReadFull(c, buf[:nmethods]); err != nil {
			return
		}
		c.Write([]byte{0x05, 0x00}) // 認証なしで許可

		// 2. コマンドリクエストの読み込み
		if _, err := io.ReadFull(c, buf[:4]); err != nil {
			return
		}
		if buf[1] != 0x01 { // CONNECTコマンド以外は終了
			return
		}

		var addr string
		switch buf[3] { // アドレスタイプ
		case 0x01: // IPv4
			if _, err := io.ReadFull(c, buf[:4]); err != nil {
				return
			}
			addr = net.IP(buf[:4]).String()
		case 0x03: // ドメイン名（これが重要！リモート名前解決）
			if _, err := io.ReadFull(c, buf[:1]); err != nil {
				return
			}
			length := int(buf[0])
			if _, err := io.ReadFull(c, buf[:length]); err != nil {
				return
			}
			addr = string(buf[:length])
		case 0x04: // IPv6
			if _, err := io.ReadFull(c, buf[:16]); err != nil {
				return
			}
			addr = net.IP(buf[:16]).String()
		default:
			return
		}

		if _, err := io.ReadFull(c, buf[:2]); err != nil {
			return
		}
		port := binary.BigEndian.Uint16(buf[:2])
		dest := fmt.Sprintf("%s:%d", addr, port)

		// ターゲットへの接続（Render側で名前解決を実行）
		remote, err := net.Dial("tcp", dest)
		if err != nil {
			c.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			return
		}
		defer remote.Close()
		c.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		// パケットの双方向転送
		relay(c, remote)

	} else if version == 0x04 {
		// SOCKS4の処理（念のための互換性モード）
		if buf[1] != 0x01 { // CONNECT
			return
		}
		if _, err := io.ReadFull(c, buf[:6]); err != nil {
			return
		}
		port := binary.BigEndian.Uint16(buf[:2])
		addr := net.IP(buf[2:6]).String()

		// USERID（ヌル文字まで読み飛ばし）
		for {
			b := make([]byte, 1)
			if _, err := c.Read(b); err != nil || b[0] == 0x00 {
				break
			}
		}

		remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
		if err != nil {
			c.Write([]byte{0x00, 0x5b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			return
		}
		defer remote.Close()
		c.Write([]byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		relay(c, remote)
	}
}

func relay(client, remote net.Conn) {
	errc := make(chan error, 2)
	go func() {
		_, err := io.Copy(remote, client)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(client, remote)
		errc <- err
	}()
	<-errc
}
