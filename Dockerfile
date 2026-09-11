# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/tokped-server ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Jakarta

COPY --from=builder /app/tokped-server /app/tokped-server

EXPOSE 8080

CMD ["/app/tokped-server"]
