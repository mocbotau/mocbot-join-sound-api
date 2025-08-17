FROM golang:1.25 AS builder

RUN apt-get update && apt-get install -y gcc libc-dev libsqlite3-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o api ./cmd/api

FROM debian:bookworm-slim AS release

RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

RUN useradd -m -u 10001 -s /bin/bash appuser

WORKDIR /app

COPY --from=builder /app/api .

RUN mkdir -p data/sounds && chown -R appuser:appuser /app && chown appuser:appuser /app/data && chown appuser:appuser /app/data/sounds

USER appuser

EXPOSE 8081

CMD ["./api"]
