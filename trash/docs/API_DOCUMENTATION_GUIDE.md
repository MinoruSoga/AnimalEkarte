# API ドキュメント統合ガイド

> **バージョン**: 2.0  
> **最終更新**: 2026-04-23  
> **ステータス**: ✅ TIER 3 — API ドキュメント統合フェーズ完了

---

## 📋 概要

本ドキュメントは、Animal Ekarte API の仕様書を保守・更新するためのガイドです。

### ドキュメント構成

| ファイル | 形式 | 用途 | 対象者 |
|---------|------|------|--------|
| **API_SPEC.md** | Markdown | マニュアル仕様書（可読性重視） | 開発者・API ユーザー |
| **openapi.yaml** | OpenAPI 3.0 YAML | 機械解析可能な仕様定義 | API ツール・クライアント生成 |
| **swagger-ui (localhost:8081)** | Web UI | ブラウザベースの対話的ドキュメント | 開発者 |
| **redoc (localhost:8082)** | Web UI | API リファレンス表示 | ドキュメント閲覧 |

---

## 🚀 使用方法

### 1. Swagger UI での閲覧

#### ローカル環境での起動

```bash
# Swagger UI + Redoc を起動
docker compose -f docker-compose.swagger.yml up -d

# または個別起動
docker compose -f docker-compose.swagger.yml up swagger-ui
docker compose -f docker-compose.swagger.yml up redoc
```

**アクセス URL**:
- Swagger UI: `http://localhost:8081`
- Redoc: `http://localhost:8082`

#### 停止

```bash
docker compose -f docker-compose.swagger.yml down
```

---

### 2. API_SPEC.md での読書

Markdown 形式で、git リポジトリやドキュメント生成ツール（Mkdocs, Hugo, Netlify）に統合可能です。

```bash
# ローカル表示
cat docs/API_SPEC.md
```

---

### 3. OpenAPI YAML の利用

#### Redocly CLI での検証

```bash
# インストール
pnpm install -g @redocly/cli

# 仕様検証
redocly lint docs/openapi.yaml

# HTML 生成
redocly build-docs docs/openapi.yaml -o api-docs.html
```

#### Swagger Codegen でのクライアント生成

```bash
# Java クライアント生成
swagger-codegen generate \
  -i docs/openapi.yaml \
  -l java \
  -o generated/java-client

# TypeScript/JavaScript クライアント生成
swagger-codegen generate \
  -i docs/openapi.yaml \
  -l typescript-fetch \
  -o generated/ts-client
```

---

## 📝 ドキュメント更新方法

### パターン 1: 新しいエンドポイント追加

#### Step 1: OpenAPI YAML に追加

```yaml
paths:
  /api/v1/new-resource:
    post:
      tags:
        - New Resource
      summary: 新規リソース作成
      operationId: createNewResource
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NewResourceInput'
      responses:
        '201':
          description: 作成成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/NewResource'
        '403':
          $ref: '#/components/responses/ForbiddenError'
```

#### Step 2: コンポーネント・スキーマを定義

```yaml
components:
  schemas:
    NewResource:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        created_at:
          type: string
          format: date-time

    NewResourceInput:
      type: object
      required:
        - name
      properties:
        name:
          type: string
```

#### Step 3: API_SPEC.md にも手動記載

```markdown
### 新規リソース (New Resources)

#### 新規リソース作成

POST /api/v1/new-resource
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "name": "sample"
}

**レスポンス (201 Created)**:
...
```

---

### パターン 2: 既存エンドポイントのリファクタリング

例: `/api/v1/owners/:id` のレスポンス構造を変更

#### Step 1: OpenAPI YAML の該当スキーマを更新

```yaml
components:
  schemas:
    Owner:
      type: object
      properties:
        id:
          type: integer
        owner_name:
          type: string
        # 新フィールド追加
        created_at:
          type: string
          format: date-time
```

#### Step 2: API_SPEC.md の該当セクションを更新

```markdown
#### 飼主詳細取得

**レスポンス (200 OK)**:
{
  "id": 1,
  "owner_name": "山田 太郎",
  "created_at": "2026-04-23T10:00:00Z"  # 新フィールド
}
```

#### Step 3: バージョン管理（git）

```bash
git add docs/openapi.yaml docs/API_SPEC.md
git commit -m "docs: API スキーマ更新 — Owner に created_at フィールド追加"
```

---

## 🔍 ドキュメント検証

### 構文検証

```bash
# OpenAPI YAML の検証（npm依存）
pnpm install -g swagger-cli
swagger-cli validate docs/openapi.yaml
```

### タイポ・リンク確認

```bash
# リンク切れ検査
pnpm install -g markdown-link-check
markdown-link-check docs/API_SPEC.md
```

---

## 🔗 Go ハンドラーへの Swag コメント統合（将来計画）

現在、API_SPEC.md と openapi.yaml はマニュアル管理されています。  
将来的には、Go ハンドラーに swag コメントを追加して自動生成を導入可能：

```go
// @Summary 飼主詳細取得
// @Tags Owners
// @Security ApiKeyAuth
// @Param id path int true "Owner ID"
// @Success 200 {object} handler.OwnerResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/owners/{id} [get]
func (h *Handler) GetOwner(c *gin.Context) {
  ...
}
```

その場合：

```bash
# Go 環境で swag インストール
go install github.com/swaggo/swag/cmd/swag@latest

# OpenAPI YAML 自動生成
swag init -g cmd/api/main.go --output docs --generalInfo cmd/api/main.go
```

---

## 📊 API 変更の影響範囲

### 新規エンドポイント追加

| ファイル | 影響度 |
|---------|--------|
| openapi.yaml | ✅ **必須** |
| API_SPEC.md | ✅ **推奨** |
| handler/*.go | ✅ **既実装** |
| frontend/ | 🔄 クライアント生成（オプション） |

### 既存エンドポイント変更

| 項目 | 影響度 |
|------|--------|
| リクエストボディ追加 | ✅ openapi.yaml 更新 |
| レスポンス フィールド変更 | ✅ openapi.yaml + API_SPEC.md 更新 |
| エラーレスポンス追加 | ✅ openapi.yaml responses セクション更新 |
| パラメータ削除 | ⚠️ **破壊的変更** — major version bump 検討 |

---

## 📚 関連リソース

### 外部ドキュメント

- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.3)
- [Swagger UI Documentation](https://swagger.io/tools/swagger-ui/)
- [Redoc Documentation](https://redoc.ly/)

### プロジェクト内ドキュメント

- [バックエンド開発ガイド](./architecture.md)
- [エラーハンドリング](../.claude/refs/error-handling.md)
- [API 設計](../.claude/refs/api.md)

### CLI ツール

- [Swagger Codegen](https://swagger.io/tools/swagger-codegen/)
- [Redocly CLI](https://redocly.com/docs/cli/)
- [swagger-cli](https://github.com/swagger-api/swagger-cli)

---

## ✅ チェックリスト

新しいエンドポイント追加時：

- [ ] Go ハンドラー実装 (`backend/internal/handler/`)
- [ ] ルート登録 (`handler.go RegisterRoutes`)
- [ ] ユニットテスト (`*_test.go`)
- [ ] openapi.yaml に paths 追加
- [ ] openapi.yaml に components/schemas 追加
- [ ] API_SPEC.md に説明追加
- [ ] Swagger UI (`localhost:8081`) で確認
- [ ] git commit（`docs: API 仕様書更新`）

---

**維持コスト**: ドキュメント自動生成化によって、ハンドラー変更時の openapi.yaml 同期を自動化できます。  
現在はマニュアル管理で、更新負荷は低く抑えられています（384 ハンドラーの自動コメント化は不要）。
