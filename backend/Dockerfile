# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/server /app/server
RUN mkdir -p /app/logs && chown -R app:app /app

USER app
EXPOSE 8080
ENV APP_ENV=prod \
    APP_PORT=8080 \
    LOG_DIR=/app/logs \
    TZ=Asia/Shanghai

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/server"]
