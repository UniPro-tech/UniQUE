#!/bin/sh

if [ ! -f /app/keys/private/rsa_private.pem ] || [ ! -f /app/keys/public/rsa_public.pem ]; then \
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /app/keys/private/rsa_private.pem && \
    openssl rsa -pubout -in /app/keys/private/rsa_private.pem -out /app/keys/public/rsa_public.pem && \
    chmod 600 /app/keys/private/rsa_private.pem && chmod 644 /app/keys/public/rsa_public.pem; \
  fi

/app/cmd/server