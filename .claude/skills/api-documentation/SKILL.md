---
name: api-documentation
description: API ドキュメント生成（OpenAPI/Swagger、手動メンテナンス）
---

# API Documentation

OpenAPI (Swagger) 仕様による API ドキュメント化。

## ドキュメント方針

本プロジェクトは **手動 OpenAPI 仕様管理** を採用：
- `backend/docs/api.yaml` - 手動で定義
- 閲覧: `backend/docs/api.yaml` を任意の OpenAPI ビューア（VSCode 拡張等）で開く（backend に Swagger UI 配信ルートは存在しない）
- コード内 swag タグは廃止（ドキュメント専用）

## api.yaml 構造

```yaml
openapi: 3.0.0
info:
  title: Animal Ekarte API
  version: 1.0.0
  description: 動物病院向け電子カルテシステムAPI

servers:
  - url: http://localhost:8080/api/v1
    description: Development（compose backend。FE は :3003）

paths:
  /owners:
    get:
      summary: オーナー一覧取得
      operationId: listOwners
      # clinic コンテキストは HttpOnly Cookie ベース（clinic_id クエリパラメータは不要）
      parameters:
        - name: page
          in: query
          schema: { type: integer, default: 1 }
        - name: limit
          in: query
          schema: { type: integer, default: 10 }
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  owners: { $ref: '#/components/schemas/Owner' }
                  total: { type: integer }

    post:
      summary: オーナー作成
      operationId: createOwner
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateOwnerRequest' }
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Owner' }
        '400':
          $ref: '#/components/responses/ValidationError'
        '409':
          $ref: '#/components/responses/ConflictError'

  /owners/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: integer }
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Owner' }
        '404':
          $ref: '#/components/responses/NotFoundError'

components:
  schemas:
    Owner:
      type: object
      required: [id, clinic_id, name, email]
      properties:
        id: { type: integer }
        clinic_id: { type: integer }
        name: { type: string }
        email: { type: string, format: email }
        phone: { type: string, pattern: '^\+?\d{10,}$' }
        address: { type: string }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
        pets:
          type: array
          items: { $ref: '#/components/schemas/Pet' }

    CreateOwnerRequest:
      type: object
      required: [name, email]
      properties:
        name: { type: string, minLength: 1, maxLength: 100 }
        email: { type: string, format: email }
        phone: { type: string }
        address: { type: string }

  responses:
    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            type: object
            properties:
              code: { type: string, example: 'INVALID_INPUT' }
              message: { type: string }
              details:
                type: array
                items:
                  type: object
                  properties:
                    field: { type: string }
                    error: { type: string }

    NotFoundError:
      description: Not found
      content:
        application/json:
          schema:
            type: object
            properties:
              code: { type: string, example: 'NOT_FOUND' }
              message: { type: string }

    ConflictError:
      description: Conflict (e.g., duplicate email)
      content:
        application/json:
          schema:
            type: object
            properties:
              code: { type: string, example: 'CONFLICT' }
              message: { type: string }
```

## エンドポイント設計パターン

### RESTful CRUD
```yaml
POST   /owners              # Create
GET    /owners              # List (クエリフィルタ)
GET    /owners/{id}         # Get one
PATCH  /owners/{id}         # Update (部分更新)
DELETE /owners/{id}         # Delete (論理削除)
```

### リレーション

```yaml
GET    /owners/{owner_id}/pets           # Owner の Pet 一覧
POST   /owners/{owner_id}/pets           # Pet 作成 (owner_id 自動)
GET    /owners/{owner_id}/pets/{pet_id}  # 特定 Pet 取得
```

### フィルタ・ソート・ページネーション

```yaml
GET /owners?name=太&page=2&limit=20&sort=-created_at

Parameters:
  # clinic コンテキストは HttpOnly Cookie ベース（クエリパラメータでは渡さない）
  - name (optional) - 部分検索
  - page (default: 1)
  - limit (default: 10, max: 100)
  - sort (例: created_at, -updated_at)

Response:
  {
    "data": [...],
    "total": 150,
    "page": 2,
    "limit": 20
  }
```

## エラーレスポンス

```yaml
400 Bad Request:
  {
    "code": "INVALID_INPUT",
    "message": "Validation failed",
    "details": [
      { "field": "email", "error": "invalid format" },
      { "field": "phone", "error": "must be E.164 format" }
    ]
  }

401 Unauthorized:
  { "code": "UNAUTHORIZED", "message": "Missing or invalid token" }

403 Forbidden:
  { "code": "FORBIDDEN", "message": "Insufficient permissions" }

404 Not Found:
  { "code": "NOT_FOUND", "message": "Owner not found" }

409 Conflict:
  { "code": "CONFLICT", "message": "Email already registered" }

500 Internal Server Error:
  { "code": "INTERNAL_ERROR", "message": "Internal server error" }
```

## セキュリティ

主メカニズムは **HttpOnly Cookie**（`access_token` 15分 / `refresh_token` 7日）。
`Authorization: Bearer <jwt>` ヘッダは後方互換フォールバックとしてのみサポート
（api.yaml 冒頭の認証説明を正とする）。

```yaml
components:
  securitySchemes:
    # Bearer は後方互換フォールバック。実運用は HttpOnly Cookie（axios withCredentials: true）
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - BearerAuth: []
```

## ドキュメント保守

### 手動更新ワークフロー

1. **エンドポイント変更時**
   - Go ハンドラを実装
   - `api.yaml` のパスを更新
   - スキーマを修正
   - OpenAPI ビューア（VSCode 拡張等）で仕様を確認

2. **定期監査**
   - 月 1 回、実装と api.yaml の同期確認
   - deprecated エンドポイントの削除予定を通知

3. **バージョニング**
   - 破壊的変更 → API v2 追加
   - 後方互換性維持 → v1 継続

## チェックリスト

- [ ] api.yaml が最新状態
- [ ] すべてのエンドポイントがドキュメント化
- [ ] パラメータ・レスポンススキーマが完全
- [ ] エラーコードが明示的
- [ ] セキュリティスキーム定義完了
- [ ] OpenAPI ビューアで仕様を確認
- [ ] 月 1 回の同期確認

## 関連スキル

- `golang-gin-api` - REST API 実装パターン
- `database-indexing` - クエリパラメータの最適化
