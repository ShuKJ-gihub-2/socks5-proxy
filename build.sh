#!/bin/bash
go mod init my-socks5 || true
go get ://github.com
go tidy
go build -o main main.go
