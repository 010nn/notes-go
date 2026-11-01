# --- Stage 1: Build ---
FROM golang:1.27.1-alpine3.24 AS builder

WORKDIR /build

# Copy dependencies for caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binaries (app and migration)
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o migration ./cmd/migration


# --- Stage 2: App Image ---
FROM alpine:latest AS app

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary & static from the builder stage
COPY --from=builder /build/app .
COPY --from=builder /build/static/html ./static/html

# Run the application by default
CMD ["./app"]


# --- Stage 3: Migration Image ---
FROM alpine:latest AS migration

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary & migrations folder.
COPY --from=builder /build/migration .
COPY --from=builder /build/migrations ./migrations

CMD ["./migration"]
