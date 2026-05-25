# ============================================================
# Stage 1: Build the Go binary
# ============================================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer if go.mod unchanged)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# ============================================================
# Stage 2: Minimal runtime image
# Non-root user reduces blast radius if the container is exploited.
# No build tools, no source code in the final image.
# ============================================================
FROM alpine:3.19

RUN adduser -D -s /bin/sh appuser

WORKDIR /app
COPY --from=builder /app/server .

USER appuser
EXPOSE 8080
ENTRYPOINT ["./server"]
