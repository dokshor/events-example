# Dockerfile
FROM golang:1.24.4-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o events ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/events /app/events

ENV PORT=8080
EXPOSE 8080

CMD ["/app/events"]