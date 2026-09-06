# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY *.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o go-sscp .

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/go-sscp /app/go-sscp

# Set default environment variables
ENV SSCP_SOURCE="_SYSTEM:SYS@iris1:1972@USER:/iris1/TESTA.txt"
ENV SSCP_TARGET="_SYSTEM:SYS@iris2:1972@USER:/iris2/copyTESTA.txt"
ENV SSCP_INTERVAL="5s"
ENV SSCP_MODE="out-of-band"

ENTRYPOINT ["/app/go-sscp"]
