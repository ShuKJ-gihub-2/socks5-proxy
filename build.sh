#!/bin/bash
go get github.com
go mod tidy
go build -o main main.go
