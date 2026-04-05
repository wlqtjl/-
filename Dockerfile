# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/wozai ./cmd/wozai

# ---- Runtime Stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S wozai && adduser -S wozai -G wozai

COPY --from=builder /app/wozai /usr/local/bin/wozai

USER wozai
EXPOSE 8080

ENTRYPOINT ["wozai"]
