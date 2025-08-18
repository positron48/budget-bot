FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bot ./cmd/bot

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /srv
COPY --from=builder /bot /srv/bot
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
ENV SERVER_ADDRESS=:8088
EXPOSE 8088
ENTRYPOINT ["/srv/bot"]


