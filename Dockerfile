FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/upload-ms/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8081
CMD ["./main"]
