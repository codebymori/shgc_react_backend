# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

# Stage 2: Run
FROM alpine:latest
# Ubah /root/ jadi /app supaya sinkron dengan volume di docker-compose
WORKDIR /app
# Ambil binary ke folder /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]