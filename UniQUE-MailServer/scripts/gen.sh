#!/bin/sh

export DB_DSN="root:rootpass123@tcp(localhost:3306)/devdb?parseTime=true"

go run cmd/gen/main.go
