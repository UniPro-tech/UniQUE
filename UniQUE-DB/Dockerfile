FROM debian:bookworm-slim

ENV version=v4.15.2
ENV os=linux
ENV arch=amd64

RUN apt-get update \
	&& apt-get install -y --no-install-recommends curl ca-certificates default-mysql-client git \
	&& rm -rf /var/lib/apt/lists/*

# Download and extract migrate binary in a temp dir, then move into /usr/local/bin
RUN curl -fSL https://github.com/golang-migrate/migrate/releases/download/$version/migrate.$os-$arch.tar.gz \
	| tar -xz -C /tmp \
	&& mv /tmp/migrate /usr/local/bin/migrate

COPY entrypoint.sh /usr/local/bin/migrate-entrypoint.sh
RUN chmod +x /usr/local/bin/migrate-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/migrate-entrypoint.sh"]
CMD ["up"]
