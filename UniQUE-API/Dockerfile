FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Cache Go modules
COPY ./src/go.mod ./src/go.sum ./src/
RUN cd src && go mod download

COPY . .

# Build arguments for embedding version metadata
ARG VERSION=dev
ARG COMMIT=none
ARG BRANCH=none

ENV CGO_ENABLED=0
RUN cd src && go build -ldflags "-s -w -X github.com/UniPro-tech/UniQUE-API/internal/config.Version=${VERSION} -X github.com/UniPro-tech/UniQUE-API/internal/config.GitCommit=${COMMIT} -X github.com/UniPro-tech/UniQUE-API/internal/config.GitBranch=${BRANCH}" ./cmd/server/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates
WORKDIR /root/

COPY --from=builder /app/src/main ./server

ENV GIN_MODE=release
EXPOSE 8080
CMD ["./server"]