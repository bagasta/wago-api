FROM golang:1.24-bookworm AS builder
WORKDIR /src

# Pre-fetch modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire workspace
COPY . .

# Build a static binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o whatsapp-api cmd/api/main.go

FROM debian:12-slim
WORKDIR /app

# Create non-root user
RUN groupadd -r app && useradd -r -g app app
ENV HOME=/app
RUN mkdir -p /home/app && chown app:app /home/app

# Copy binary and runtime assets
COPY --from=builder /src/whatsapp-api .
COPY config/config.docker.yaml /app/config/config.yaml
COPY docs /app/docs

# Fiber will bind to this port (override with SERVER_PORT env)
EXPOSE 9300

USER app
CMD ["./whatsapp-api"]
