# =====================================================================
# Stage 1: Build Go Backend Binary
# =====================================================================
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git build-base

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cbom-server ./cmd/server

# =====================================================================
# Stage 2: Minimal Runtime Container
# =====================================================================
FROM alpine:3.19

WORKDIR /app

# Install git (required for secret history scans) & ca-certificates for TLS handshakes
RUN apk add --no-cache ca-certificates git tzdata

# Copy compiled binary and detection rules
COPY --from=builder /build/cbom-server /app/cbom-server
COPY --from=builder /build/rules.yaml /app/rules.yaml

# Expose backend REST API port
EXPOSE 8080

# Run API server
ENTRYPOINT ["/app/cbom-server"]
