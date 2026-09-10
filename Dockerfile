# Stage 1: Build the binary
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o orderhub orderhub.go

# Stage 2: Final lightweight runner
FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/orderhub /app/orderhub
COPY --from=builder /app/etc /app/etc

EXPOSE 8888
CMD ["./orderhub", "-f", "etc/orderhub-api.yaml"]