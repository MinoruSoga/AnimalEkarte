---
name: docker-optimization
description: Docker イメージ最適化（マルチステージ、レイヤーキャッシング、サイズ削減）
---

# Docker Optimization

イメージサイズ削減、ビルド時間高速化、セキュリティ向上を実現します。

## 最適化戦略

### 1. マルチステージビルド

#### ❌ 非効率
```dockerfile
FROM golang:1.25
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/api
EXPOSE 8080
CMD ["./app"]

# 結果: イメージサイズ 1.2GB (Go SDK含む)
```

#### ✅ 最適化
```dockerfile
# ビルドステージ
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

# 実行ステージ
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]

# 結果: イメージサイズ 50MB (Go SDK除外)
```

### 2. レイヤーキャッシング最適化

#### レイヤー順序が重要
```dockerfile
# ❌ 非効率: コード変更で全層再ビルド
FROM golang:1.25-alpine
COPY . .                 # コード = 最も変更頻度が高い → 最後に配置
RUN go mod download
RUN go build -o app ./cmd/api

# ✅ 最適化: 変更頻度が低い層を先に配置
FROM golang:1.25-alpine
COPY go.mod go.sum ./   # 依存 = 変更頻度が低い
RUN go mod download
COPY . .                # コード = 最後
RUN go build -o app ./cmd/api
```

### 3. イメージサイズ削減

```dockerfile
FROM alpine:3.21

# 不要なファイル除外
RUN apk add --no-cache ca-certificates tzdata

# キャッシュクリア
RUN rm -rf /var/cache/apk/*

COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
```

### 4. セキュリティ最適化

```dockerfile
FROM alpine:3.21

# non-root ユーザー作成
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder --chown=appuser:appuser /app/app .

# non-root で実行
USER appuser

EXPOSE 8080
CMD ["./app"]
```

### 5. React Frontend 最適化

> 本プロジェクトのフロントエンド本番は Vercel デプロイ（Nginx イメージは使用しない）。以下は汎用パターン。

```dockerfile
# ビルドステージ
FROM node:24-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
# ※ --prod は devDependencies(vite/tsc) を除外するため後続の pnpm build が失敗する

COPY . .
RUN pnpm build

# 実行ステージ (Nginx)
FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

# 結果: 開発時 1GB → 本番時 100MB
```

## チェックリスト

- [ ] マルチステージビルド使用
- [ ] レイヤー順序最適化（依存 → コード）
- [ ] 不要ファイル除外（.dockerignore）
- [ ] キャッシュクリア実装
- [ ] Non-root ユーザー使用
- [ ] イメージサイズ目標: Backend < 50MB / Frontend < 100MB（パフォーマンス目標と同一値）
- [ ] セキュリティスキャン (trivy)

## パフォーマンス目標

```
Build Time:    < 2分
Image Size:    Backend < 50MB, Frontend < 100MB
Startup Time:  < 5s
Security:      Vulnerability 0 (Critical)
```

## 関連スキル

- `ci-cd-automation` - CI パイプライン最適化
- `deployment` - 本番環境デプロイ
