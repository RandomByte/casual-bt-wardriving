#!/bin/bash
GOOS=linux GOARCH=mipsle go build -ldflags "-s -w" -compiler gc -o out/casual-bt-wardriving casual-bt-wardriving.go