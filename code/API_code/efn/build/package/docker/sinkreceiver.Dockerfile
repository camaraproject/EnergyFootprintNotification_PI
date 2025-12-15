# syntax=docker/dockerfile:1

# -------- builder --------
FROM golang:1.24-alpine AS builder
WORKDIR /src

# Optional: CA certs for copying to runtime image
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Cache modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/sinkreceiver ./cmd/sinkreceiver

# -------- runtime --------
FROM alpine:3.20
WORKDIR /app

# non-root user
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot \
    && apk add --no-cache ca-certificates curl \
    && update-ca-certificates

COPY --from=builder /out/sinkreceiver /app/sinkreceiver

# expose default sinkreceiver port (adjust if needed)
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/app/sinkreceiver"]
