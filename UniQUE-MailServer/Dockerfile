FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

# cache Go modules
COPY src/go.mod src/go.sum ./
RUN go mod download

# copy Go module and source
COPY src .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags "-s -w" -o /app/server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /app/server /app/server

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]