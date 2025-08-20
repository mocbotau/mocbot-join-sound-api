FROM golang:1.25.0-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_enabled=0 go build -o api ./cmd/api

FROM alpine:latest AS release

RUN apk add --no-cache ca-certificates
RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /app/api .

RUN mkdir -p data/sounds && chown -R appuser:appuser /app && chown -R appuser:appuser /app/data && chown -R appuser:appuser /app/data/sounds

USER appuser

EXPOSE 8081

CMD ["./api"]
