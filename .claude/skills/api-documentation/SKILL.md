---
name: api-documentation
description: API ドキュメント生成（OpenAPI/Swagger、手動メンテナンス）
---

# API Documentation

OpenAPI (Swagger) 仕様による API ドキュメント化。

## ドキュメント方針

本プロジェクトは **手動 OpenAPI 仕様管理** を採用：
- `backend/docs/api.yaml` - 手動で定義
- Swagger UI: `http://localhost:8080/swagger/index.html`
- コード内 swag タグは廃止（ドキュメント専用）

## api.yaml 構造

```yaml
openapi: 3.0.0
info:
  title: Animal Ekarte API
  version: 1.0.0
  description: 動物病院向け電子カルテシステムAPI

servers:
  - url: http://localhost:8080
    description: Development

paths:
  /api/owners:
    get:
      summary: オーナー一覧取得
      operationId: listOwners
      parameters:
        - name: clinic_id
          in: query
          required: true
          schema: { type: integer }
        - name: page
          in: query
          schema: { type: integer, default: 1 }
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

  /api/owners/{id}:
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
POST   /api/owners              # Create
GET    /api/owners              # List (クエリフィルタ)
GET    /api/owners/{id}         # Get one
PATCH  /api/owners/{id}         # Update (部分更新)
DELETE /api/owners/{id}         # Delete (論理削除)
```

### リレーション

```yaml
GET    /api/owners/{owner_id}/pets           # Owner の Pet 一覧
POST   /api/owners/{owner_id}/pets           # Pet 作成 (owner_id 自動)
GET    /api/owners/{owner_id}/pets/{pet_id}  # 特定 Pet 取得
```

### フィルタ・ソート・ページネーション

```yaml
GET /api/owners?clinic_id=1&name=太&page=2&size=20&sort=-created_at

Parameters:
  - clinic_id (required) - マルチテナント
  - name (optional) - 部分検索
  - page (default: 1)
  - size (default: 10, max: 100)
  - sort (例: created_at, -updated_at)

Response:
  {
    "data": [...],
    "pagination": {
      "page": 2,
      "size": 20,
      "total": 150,
      "pages": 8
    }
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

## セキュリティヘッダー

```yaml
components:
  securitySchemes:
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
   - Swagger UI で動作確認

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
- [ ] Swagger UI で動作確認
- [ ] 月 1 回の同期確認

## 関連スキル

- `golang-gin-api` - REST API 実装パターン
- `database-indexing` - クエリパラメータの最適化
