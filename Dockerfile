FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

COPY --from=builder /app/api .

RUN mkdir -p /root/data /root/files

EXPOSE 8081

ENV GIN_MODE=release
ENV DB_PATH=/root/data/sounds.db
ENV PORT=8081

CMD ["./api"]
