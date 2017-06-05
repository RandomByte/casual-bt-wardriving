#!/bin/bash
GOOS=linux GOARCH=mipsle go build -ldflags "-s -w" -compiler gc -o out/onion-bt-wardriving onion-bt-wardriving.go
