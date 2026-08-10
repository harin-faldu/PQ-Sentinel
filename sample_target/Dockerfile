# Sample Dockerfile for QuantumVault CBOM Container Scanning

FROM ubuntu:22.04

# Install legacy OpenSSL and LibSSL packages
RUN apt-get update && apt-get install -y \
    openssl \
    libssl-dev \
    curl \
    ca-certificates

WORKDIR /app
COPY . .

CMD ["node", "auth_service.js"]
