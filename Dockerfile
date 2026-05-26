# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /subscription-api ./cmd/api/main.go

# Stage 2: Final minimal image
FROM alpine:latest
WORKDIR /
COPY --from=builder /subscription-api /subscription-api
EXPOSE 8080
CMD ["/subscription-api"]