# 本番用 Dockerfile（マルチステージビルド）
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api

# 軽量な実行イメージ
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
