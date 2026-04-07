## Multi-stage Dockerfile
##  - builder: build the Go server binary
##  - runtime: minimal Alpine with openssl and mysql client, runs `scripts/run.sh`

FROM golang:1.26.1-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Cache Go modules
COPY ./src/go.mod ./src/go.sum ./src/
RUN cd src && go mod download

COPY . .

RUN export COMMIT=$(git rev-parse --short HEAD)
RUN export BRANCH=$(git branch --show-current)


# Build the Go server (static, no cgo)
RUN cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

FROM alpine:3.23 AS runtime
RUN apk add --no-cache openssl mysql-client ca-certificates

WORKDIR /app
RUN mkdir -p /app/keys/private /app/keys/public /app/unique_api

# Copy built server and scripts
COPY --from=builder /app/server /app/cmd/server
COPY --from=builder /src/scripts /app/scripts

EXPOSE 8080

RUN chmod +x /app/scripts/run.sh

CMD ["/app/scripts/run.sh"]