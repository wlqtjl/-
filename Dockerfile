# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/wozai ./cmd/wozai

# ---- Runtime Stage (极小镜像，~15MB) ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S wozai && adduser -S wozai -G wozai

COPY --from=builder /app/wozai /usr/local/bin/wozai

# Go 运行时内存控制 (适合 700MB VPS)
ENV GOMEMLIMIT=50MiB
ENV GOGC=50

USER wozai
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=10s \
  CMD wget -q --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["wozai"]
