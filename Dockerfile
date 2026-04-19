FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY app ./app
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/robocall ./app

FROM debian:bookworm-slim

WORKDIR /app

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates \
      tzdata; \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/robocall /app/robocall
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
COPY templates /app/templates
COPY static /app/static
COPY app/config_sample.json /app/config_sample.json

RUN chmod 755 /app/robocall /app/docker-entrypoint.sh

EXPOSE 8080 443

ENTRYPOINT ["/app/docker-entrypoint.sh"]
