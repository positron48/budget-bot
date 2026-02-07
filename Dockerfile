FROM golang:1.23.1-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags withgrpc -o /out/budget-bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget

RUN addgroup -g 1001 -S budgetbot && adduser -u 1001 -S budgetbot -G budgetbot
WORKDIR /app

COPY --from=builder /out/budget-bot /app/budget-bot
COPY --from=builder /src/migrations /app/migrations

RUN mkdir -p /app/data /app/logs /app/configs && chown -R budgetbot:budgetbot /app
USER budgetbot

EXPOSE 8088 9090
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8088/healthz || exit 1

ENTRYPOINT ["/app/budget-bot"]
