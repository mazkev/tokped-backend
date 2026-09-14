# Production Dockerfile - Tokopedia Backend Monolith
FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api

# Minimal runtime
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]