# リクエスト〜レスポンス データフロー

Owner CRUD を例にした、HTTPリクエストからレスポンスまでの全層の処理フロー。
層の責務概要は [architecture.md](./architecture.md) を参照。

---

## トレーサビリティとロギング

本システムは、商用グレードの運用監視を実現するため、すべてのリクエストを一意の ID で追跡しています。

### Request ID の伝播フロー

1.  **生成**: `middleware.RequestID()` がリクエスト受信時に一意の UUID を生成。
2.  **格納**: `gin.Context` に `request_id` としてセット。
3.  **レスポンスヘッダー**: `X-Request-ID` ヘッダーとしてクライアントに返却。
4.  **ログ出力**: `middleware.RequestLoggingMiddleware()` および Service 層の `slog` 出力において、常に `request_id` フィールドが含まれる。

### 構造化ログの実装方針

- **コンテキストの保持**: すべての Service/Repository メソッドは `context.Context` を第一引数に受け取ります。
- **slog の活用**: `slog.InfoContext(ctx, "message", ...)` を使用することで、ログ基盤（Datadog/CloudWatch等）でリクエスト単位のフィルタリングが可能になります。

---

## CRUD 別フロー

### GET /api/v1/owners — 一覧取得

```
Client
  │  GET /api/v1/owners?page=1&per_page=20&search=山田
  │  Cookie: access_token=<JWT>
  ▼

[Middleware: Auth]
  1. Cookie "access_token" を読む（Authorization ヘッダーも対応）
  2. JWT を検証（HMAC署名確認・有効期限確認）
  3. claims を gin.Context に格納 (user_id, clinic_id, user_type)
  4. account_status / deleted_at の有効性を DB で再検証（BUG-061/063対応）
  5. c.Next() で次のハンドラへ

[Handler: ListOwners]
  1. extractClinicID(c) → clinicID 取得 (JWT claims 由来)
  2. parsePagination(c) → page, limit 取得 (limit と per_page の両方をサポート)
  3. h.svc.Owner.List(ctx, clinicID, page, limit, search)

[Service: ownerService.List]
  1. s.repo.FindAll(ctx, clinicID, page, limit, search) を呼び出し

[Repository: ownerRepository.FindAll]
  1. buildBase() でベースクエリ構築: 
       SELECT * FROM owners 
       WHERE clinic_id = 1 AND deleted_at IS NULL
  2. search != "" なので追加条件:
       AND (name ILIKE '%山田%' OR phone ILIKE '%山田%' OR email ILIKE '%山田%')
  3. COUNT(*) で total 取得
  4. Preload("Pets") 等の実データ取得
  5. ([]model.Owner, total, nil) を返す

[Handler: ListOwners — 続き]
  4. RespondError(c, err) または 
     c.JSON(200, PaginatedResponse{Data: owners, Total: total, Page: page, Limit: limit})
```

---

### POST /api/v1/owners — 新規作成（ペット同時登録）

```
Client
  │  POST /api/v1/owners
  │  Body: { "name": "林 文昭", "pets": [...] }
  ▼

[Middleware: Auth] — 同上

[Handler: CreateOwner]
  1. extractClinicID(c) → clinicID 取得
  2. c.ShouldBindJSON(&req)
       失敗時: RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
  3. h.svc.Owner.CreateWithPets(ctx, clinicID, &input)

[Service: ownerService.CreateWithPets]
  1. 業務バリデーション (validators.go)
  2. DTO → model.Owner 変換
  3. s.repo.CreateWithPets(ctx, owner, pets)

[Repository: ownerRepository.CreateWithPets]
  1. db.Transaction(func(tx) error {
       a. tx.Create(owner)
       b. for pets { tx.Create(pet) }
       return nil
     })

[Handler: CreateOwner — 続き]
  4. RespondError(c, err) または c.JSON(201, toOwnerResponse(owner))
```

---

## エラーハンドリング全体図

### エラーレスポンス形式

`RespondError(c, err)` は一律以下の JSON 形式でレスポンスを返却します。

```json
{
  "error": "エラーメッセージ（ユーザー向け）"
}
```

- **4xx (Client Error)**: `WrapInvalidInput` (400), `WrapNotFound` (404), `WrapConflict` (409) 等、具体的な理由を返却。
- **5xx (Server Error)**: 予期せぬエラー。セキュリティのため詳細は露出せず、一律 `"internal server error"` を返却。サーバー側の `slog` に詳細を記録。
- **入力バリデーション**: `ShouldBindJSON` 失敗時、`camelToSnake` により **snake_case のフィールド名**を含めたメッセージを返却。

### RespondError のエラーマッピング

| センチネルエラー (apperrors) | HTTPステータス |
|---|---|
| `ErrNotFound` | 404 |
| `ErrInvalidInput` | 400 |
| `ErrAlreadyExists` | 409 |
| `ErrConflict` | 409 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| その他 | 500 |

---

## 3. ページネーションと検索のフロー

`GET /api/v1/owners?page=1&per_page=20&search=山田` を例にする。

1.  **Handler**:
    - `parsePagination(c)` を呼び出し。
    - **エイリアス対応**: `limit` パラメータがない場合、`per_page` も自動的に参照。
    - デフォルト値: `page=1`, `limit=20` (Max 100)。
2.  **Service**:
    - `List(ctx, clinicID, page, limit, search)` を実行。
3.  **Repository**:
    - `page`, `limit` から SQL の `OFFSET`, `LIMIT` を計算。
    - `COUNT(*)` と `SELECT *` を同一コンテキスト内で実行し、総件数とリストを返却。

---

## マルチテナント分離の徹底

全エンドポイントで `clinic_id` によるテナント分離を徹底している。

```
JWT claims (clinic_id) → extractClinicID() → Service → Repository
                                                         ↓
                                       WHERE clinic_id = ? AND deleted_at IS NULL
```

**セキュリティ原則**:
- クライアントからの `clinic_id` 指定は一切信用せず、常に JWT 内の情報を正とする。
- 他クリニックのリソースへのアクセス（ID指定等）が発生した場合は、存在を推測させないため `403` ではなく `404` を返却する。
