---
description: Docker開発規約（マルチステージビルド、セキュリティ、ベストプラクティス）
alwaysApply: false
globs: ["**/Dockerfile*", "docker-compose.yml", "**/.dockerignore"]
---

# Docker Rules

Docker Compose 開発環境標準ルール。

## 核心ルール

### 1. マルチステージビルド（必須）

```dockerfile
# ❌ 非効率: イメージサイズ 1.2GB
FROM golang:1.25
COPY . .
RUN go build -o app ./cmd/api
EXPOSE 8080
CMD ["./app"]

# ✅ 最適化: イメージサイズ 50MB
# ビルドステージ
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

# 実行ステージ
FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/app .
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser
EXPOSE 8080
CMD ["./app"]
```

### 2. レイヤーキャッシング最適化

```dockerfile
# ❌ 非効率: コード変更で全層再ビルド
FROM golang:1.25-alpine
COPY . .                 # ← コード（最も変更頻度が高い）
RUN go mod download
RUN go build -o app ./cmd/api

# ✅ 最適化: 変更頻度が低い層を先に
FROM golang:1.25-alpine
COPY go.mod go.sum ./   # ← 依存（変更頻度が低い）
RUN go mod download
COPY . .                # ← コード（最後に配置）
RUN go build -o app ./cmd/api
```

### 3. .dockerignore ファイル

```
# backend/.dockerignore
.git
.gitignore
README.md
Makefile
*.local
.DS_Store
__pycache__
*.pyc
node_modules
dist
build
coverage
.env.local

# frontend/.dockerignore
.git
.gitignore
README.md
node_modules
dist
build
.env.local
.DS_Store
cypress
coverage
```

### 4. Non-root ユーザー（セキュリティ）

```dockerfile
# ✅ Dockerfile
FROM alpine:3.18

# ユーザー作成
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder --chown=appuser:appuser /app/app .

# non-root で実行
USER appuser

EXPOSE 8080
CMD ["./app"]
```

### 5. Frontend Nginx 最適化

```dockerfile
# ビルドステージ
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# 実行ステージ
FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 6. Docker Compose コマンド（プロジェクト標準）

```bash
# ✅ Make コマンド使用
make up          # コンテナ起動
make down        # コンテナ停止
make logs        # ログ表示
make clean       # キャッシュクリア＆再ビルド

# ✅ コンテナ内コマンド実行（npm/go はローカル禁止）
docker compose exec frontend npm run build
docker compose exec backend go test ./... -v
docker compose exec backend golangci-lint run ./...

# ❌ ローカル実行禁止
npm run build
go test ./...
```

### 7. セキュリティスキャン

```bash
# trivy でセキュリティスキャン
docker build -t ekarte-backend:latest .
trivy image ekarte-backend:latest

# 脆弱性ない状態を維持
# Critical 0, High 0
```

## チェックリスト

- [ ] マルチステージビルド使用
- [ ] レイヤー順序最適化（go.mod/sum → コード）
- [ ] .dockerignore で不要ファイル除外
- [ ] Non-root ユーザー実装
- [ ] イメージサイズ < 200MB
- [ ] セキュリティスキャン (trivy) で Critical 0
- [ ] npm/go はローカル実行禁止（Docker経由のみ）

## パフォーマンス目標

```
Build Time:    < 2分
Image Size:    Backend < 50MB, Frontend < 100MB
Startup Time:  < 5s
Security:      Vulnerability Critical/High 0
```
