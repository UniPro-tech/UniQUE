# check=error=true
# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine3.23 AS builder
WORKDIR /app

COPY . .

WORKDIR /app/src

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    go mod download -x

RUN ../scripts/_build.sh

FROM alpine:3.24.1
WORKDIR /root/

COPY --from=builder /app/src/main .
RUN chmod 755 ./main

CMD ["/root/main"]

# OCI Metadata

LABEL org.opencontainers.image.url="https://github.com/UniPro-tech/UniQUE"
LABEL org.opencontainers.image.source="https://github.com/UniPro-tech/UniQUE"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.title="UniQUE Discord Bot"
LABEL org.opencontainers.image.description="UniQUE is a Discord bot that provides various features such as music playback, moderation tools, and fun commands. It is designed to enhance the Discord experience for users and server administrators."
LABEL org.opencontainers.image.vendor="All-Japan Digital Creative Club UniProject <info@uniproject.jp>"
