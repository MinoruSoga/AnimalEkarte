# リクエスト〜レスポンス データフロー

Owner CRUD を例にした、HTTPリクエストからレスポンスまでの全層の処理フロー。

---

## 層の責務マップ

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────────┐
│ Middleware (middleware/)                     │
│  JWT検証・claims抽出 → gin.Context に格納   │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│ Handler (handler/)                          │
│  パラメータ抽出・バインド・型変換           │
│  Service呼び出し → HTTPレスポンス書き込み   │
└─────────────────────────────────────────────┘
    │  *service.CreateOwnerInput (DTO pointer)
    ▼
┌─────────────────────────────────────────────┐
│ Service (service/)                          │
│  ビジネスバリデーション・DTO→Model変換      │
│  slogによる構造化ログ                        │
└─────────────────────────────────────────────┘
    │  *model.Owner, []model.Pet
    ▼
┌─────────────────────────────────────────────┐
│ Repository (repository/)                    │
│  GORM操作・DBエラーのセンチネルエラー変換   │
└─────────────────────────────────────────────┘
    │  SQL
    ▼
┌─────────────────────────────────────────────┐
│ PostgreSQL                                  │
└─────────────────────────────────────────────┘
```

### 各層が「やらないこと」

| 層 | やらないこと |
|---|---|
| Middleware | ビジネスロジック、DB操作 |
| Handler | バリデーション（型チェック以外）、SQL、slog |
| Service | HTTPの概念（ステータスコード等）、DB操作 |
| Repository | ビジネスルール、slog、HTTP |

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
  │  GET /api/v1/owners?page=1&limit=20&search=山田
  │  Cookie: auth_token=<JWT>
  ▼

[Middleware: Auth]
  1. Cookie "auth_token" を読む（なければ Authorization Bearer にフォールバック）
  2. JWT を検証（HMAC署名確認・有効期限確認）
  3. claims を gin.Context に格納
       c.Set("user_id",   "42")
       c.Set("clinic_id", "1")
       c.Set("user_type", "staff")  // "system_admin" | "clinic_admin" | "staff"
  4. c.Next() で次のハンドラへ
     ⚠️ BUG-061: DB の account_status を確認しない（停止アカウントが通過する）
     ⚠️ BUG-063: deleted_at IS NULL を確認しない（論理削除済みユーザーが通過する）

[Handler: ListOwners]
  1. extractClinicID(c)
       → c.Get("clinic_id") → "1" → ParseUint → clinicID = 1
  2. parsePagination(c)
       → page=1, limit=20 (範囲: 1≤page, 1≤limit≤100)
  3. search = c.Query("search") → "山田"
  4. h.svc.Owner.List(ctx, clinicID=1, page=1, limit=20, search="山田")

[Service: ownerService.List]
  1. s.repo.FindAll(ctx, clinicID, page, limit, search) をそのまま委譲
     ※ 一覧取得はビジネスロジックなし

[Repository: ownerRepository.FindAll]
  1. buildBase() でベースクエリ構築
       SELECT * FROM owners WHERE clinic_id = 1
  2. search != "" なので追加条件
       AND (owner_name ILIKE '%山田%' OR phone ILIKE '%山田%' OR email ILIKE '%山田%')
  3. COUNT(*) で total 取得
  4. Preload("Pets").Preload("Pets.AnimalSpecies").Preload("Pets.Insurance")
     OFFSET 0 LIMIT 20 ORDER BY created_at DESC で実データ取得
  5. ([]model.Owner, total=2, nil) を返す

[Handler: ListOwners — 続き]
  5. err == nil なので
  6. c.JSON(200, PaginatedResponse{Data: owners, Total: 2, Page: 1, Limit: 20})

Response: 200 OK
{
  "data": [ { "id": 1, "owner_name": "山田 太郎", "pets": [...] }, ... ],
  "total": 2,
  "page":  1,
  "limit": 20
}
```

---

### GET /api/v1/owners/:id — 単体取得

```
Client
  │  GET /api/v1/owners/10
  ▼

[Middleware: Auth] — 同上

[Handler: GetOwner]
  1. extractClinicID(c) → clinicID = 1
  2. strconv.ParseUint(c.Param("id")) → id = 10
     ParseUint 失敗時: 400 {"error": "invalid id"}
  3. h.svc.Owner.GetByID(ctx, clinicID=1, id=10)

[Service: ownerService.GetByID]
  1. s.repo.FindByID(ctx, clinicID, id) を委譲

[Repository: ownerRepository.FindByID]
  1. SELECT * FROM owners WHERE id=10 AND clinic_id=1 LIMIT 1
     + Preload("Pets"), Preload("Pets.AnimalSpecies"), Preload("Pets.Insurance")
  2. gorm.ErrRecordNotFound → apperrors.WrapNotFound("owner", "10")
                                   ↓
                              &AppError{Err: ErrNotFound}
  3. 見つかった場合: *model.Owner を返す

[Handler: GetOwner — 続き]
  4. err != nil → RespondError(c, err)
       errors.Is(err, ErrNotFound) → true
         AppError.Message = "owner with id 10 not found"
         → 404 {"error": "owner with id 10 not found"}
  4. err == nil → c.JSON(200, owner)

Response (成功): 200 OK  { "id": 10, "owner_name": "...", "pets": [...] }
Response (失敗): 404     { "error": "owner with id 10 not found" }
```

---

### POST /api/v1/owners — 新規作成（ペット同時登録）

```
Client
  │  POST /api/v1/owners
  │  Body: {
  │    "owner_name": "林 文昭",
  │    "email": "hayashi@example.com",
  │    "discount_rate": 10,
  │    "membership_type": "member",
  │    "pets": [
  │      { "name": "ポチ", "animal_species_id": 1, "gender": "male" }
  │    ]
  │  }
  ▼

[Middleware: Auth] — 同上

[Handler: CreateOwner]
  1. extractClinicID(c) → clinicID = 1
  2. var input service.CreateOwnerInput
     c.ShouldBindJSON(&input)
       NG例: owner_name 欠落 → validator.ValidationErrors
             parseBindError(err) → "owner_name is required"
             → 400 {"error": "owner_name is required"}
       NG例: discount_rate=150 → "discount_rate must be at most 100"
             → 400 {"error": "discount_rate must be at most 100"}
  3. h.svc.Owner.CreateWithPets(ctx, clinicID=1, &input)

[Service: ownerService.CreateWithPets]
  1. validateDiscountRate(input.DiscountRate)
       0 <= rate <= 100 でなければ apperrors.WrapInvalidInput(...)
  2. validateMembershipType(input.MembershipType)
       {"non_member","member","deceased","transferred"} 以外 → WrapInvalidInput
  3. for i, p := range input.Pets
       validatePetGender(p.Gender) → {"male","female","unknown",""}以外 → エラー
  4. DTO → Model 変換
       owner := &model.Owner{
           ClinicID: clinicID, OwnerName: "林 文昭",
           Email: "hayashi@example.com", DiscountRate: 10,
           MembershipType: "member",
       }
       pets := []model.Pet{
           {Name: "ポチ", AnimalSpeciesID: 1, Gender: "male"},
       }
  5. s.repo.CreateWithPets(ctx, owner, pets)
  6. 成功時: slog.InfoContext(ctx, "owner created with pets",
                 slog.Uint64("owner_id", owner.ID),
                 slog.Uint64("clinic_id", clinicID),
                 slog.Int("pets_count", 1))
  7. owner を返す（Pets フィールドにDB挿入済みのペットが含まれる）

[Repository: ownerRepository.CreateWithPets]
  1. db.Transaction(func(tx) error {
       a. tx.Create(owner)
            重複メール → WrapAlreadyExists("owner", "email already registered")
       b. for i := range pets {
            pets[i].OwnerID  = owner.ID   // サーバー側でセット（クライアントから受け取らない）
            pets[i].ClinicID = owner.ClinicID
            tx.Create(&pets[i])
            tx.Preload("AnimalSpecies").First(&created, pets[i].ID)
          }
       c. owner.Pets = createdPets
       d. return nil → コミット / error → ロールバック
     })
  2. Transaction エラー → apperrors.Wrap(err, "create owner with pets")

[Handler: CreateOwner — 続き]
  4. err != nil → RespondError(c, err)
       ErrAlreadyExists → 409 {"error": "owner 'email already registered' already exists"}
       ErrInvalidInput  → 400 {"error": "..."}
  4. err == nil → c.JSON(201, owner)

Response (成功): 201 Created  { "id": 5, "owner_name": "林 文昭", "pets": [...] }
Response (重複): 409          { "error": "owner 'email already registered' already exists" }
```

---

### PATCH /api/v1/owners/:id — 部分更新

```
Client
  │  PATCH /api/v1/owners/10
  │  Body: { "is_dangerous": false, "discount_rate": 5.0 }
  │  ※ 送信したフィールドだけ更新（未送信フィールドは変更なし）
  ▼

[Middleware: Auth] — 同上

[Handler: UpdateOwner]
  1. extractClinicID(c) → clinicID = 1
  2. ParseUint(c.Param("id")) → id = 10
  3. var input service.UpdateOwnerInput
     c.ShouldBindJSON(&input)
       全フィールドがポインタ型のため:
         未送信 → nil（「変更しない」の意味）
         送信済 → 非nil（「この値に変更する」の意味）
       ※ is_dangerous: false でも nil にならない（ゼロ値問題を回避）
  4. h.svc.Owner.Update(ctx, clinicID=1, id=10, &input)

[Service: ownerService.Update]
  1. input.DiscountRate != nil → validateDiscountRate(5.0) → OK
  2. input.MembershipType == nil → スキップ
  3. buildOwnerUpdateFields(input)
       map[string]any{
           "is_dangerous":  false,   // *bool → bool（ゼロ値でも明示的に格納）
           "discount_rate": 5.0,
       }
       ※ len(fields) == 0 なら WrapInvalidInput("at least one field must be provided")
  4. s.repo.Update(ctx, clinicID, id, fields)
  5. 成功時: slog.InfoContext(ctx, "owner updated", ...)
  6. s.repo.FindByID(ctx, clinicID, id) → 更新後のDB最新状態を返す

[Repository: ownerRepository.Update]
  1. db.Model(&Owner{}).Where("id=10 AND clinic_id=1").Updates(fields)
       ※ map[string]any を渡すことでGORMのゼロ値スキップを回避
          struct渡しだと false/0/"" がスキップされるバグが発生する
  2. RowsAffected == 0 → WrapNotFound("owner", "10")
  3. result.Error != nil → Wrap(err, "update owner")

[Repository: ownerRepository.FindByID] — Update後に呼ばれる
  1. SELECT + Preload で最新状態を取得して返す

[Handler: UpdateOwner — 続き]
  5. err != nil → RespondError(c, err)
       ErrNotFound    → 404 {"error": "owner with id 10 not found"}
       ErrInvalidInput → 400 {"error": "at least one field must be provided"}
  5. err == nil → c.JSON(200, owner)

Response (成功): 200 OK  { "id": 10, "is_dangerous": false, "discount_rate": 5.0, ... }
Response (未存在): 404   { "error": "owner with id 10 not found" }
Response (空body): 400   { "error": "at least one field must be provided" }
```

---

### DELETE /api/v1/owners/:id — 削除

```
Client
  │  DELETE /api/v1/owners/10
  ▼

[Middleware: Auth] — 同上

[Handler: DeleteOwner]
  1. extractClinicID(c) → clinicID = 1
  2. ParseUint(c.Param("id")) → id = 10
  3. h.svc.Owner.Delete(ctx, clinicID=1, id=10)

[Service: ownerService.Delete]
  1. s.repo.Delete(ctx, clinicID, id) を委譲
     ※ 削除にビジネスロジックなし（将来的な関連チェック等はここに追加）

[Repository: ownerRepository.Delete]
  1. DELETE FROM owners WHERE id=10 AND clinic_id=1
     （GORMがdeleted_atを持つ場合はソフトデリート）
  2. RowsAffected == 0 → WrapNotFound("owner", "10")

[Handler: DeleteOwner — 続き]
  4. err != nil → RespondError(c, err)
       ErrNotFound → 404 {"error": "owner with id 10 not found"}
  4. err == nil → c.Status(204)

Response (成功): 204 No Content  (body なし)
Response (未存在): 404           { "error": "owner with id 10 not found" }
```

---

## エラーハンドリング全体図

```
エラー発生箇所          センチネルエラーへの変換         HTTPレスポンス
─────────────────────────────────────────────────────────────────────
Repository
  gorm.ErrRecordNotFound → WrapNotFound()         → ErrNotFound
  unique constraint err  → WrapAlreadyExists()    → ErrAlreadyExists
  その他DBエラー         → Wrap(err, "message")   → ErrInternal(500)

Service
  バリデーション失敗     → WrapInvalidInput()     → ErrInvalidInput
  空フィールドPATCH      → WrapInvalidInput()     → ErrInvalidInput

Handler
  JWT検証失敗           → (middleware が直接返す)  → 401
  ParseUint失敗         → (直接返す)              → 400
  ShouldBindJSON失敗    → parseBindError()         → 400
  その他(5xx)           → slog.ErrorContext() 出力 → 500
```

### RespondError のエラーマッピング

| センチネルエラー | errors.Is | HTTPステータス |
|---|---|---|
| `ErrNotFound` | ✓ | 404 |
| `ErrInvalidInput` | ✓ | 400 |
| `ErrAlreadyExists` | ✓ | 409 |
| `ErrUnauthorized` | ✓ | 401 |
| `ErrForbidden` | ✓ | 403 |
| その他 | — | 500 (内部エラー詳細は非公開) |

`AppError` は `Unwrap()` で sentinel を保持するため、`errors.Is()` がラップチェーンを辿って正しくマッチする。

---

## PATCH のゼロ値問題とその解決

**問題**: GORMは `struct` を渡すと `false` / `0` / `""` を「未変更」と判断してスキップする。

```go
// ❌ is_dangerous: false がスキップされる
db.Updates(&model.Owner{IsDangerous: false})

// ✅ map なら全キーが明示的なので false もセットされる
db.Updates(map[string]any{"is_dangerous": false})
```

**実装パターン**:
1. `UpdateOwnerInput` の全フィールドをポインタ型 (`*bool`, `*float64`, ...) にする
2. `buildOwnerUpdateFields()` で `nil` チェックし、非nil のみ `map[string]any` に追加
3. Repository に `map[string]any` を渡す

```
リクエストBody       UpdateOwnerInput           map[string]any
{ is_dangerous:    → IsDangerous: *bool(false) → {"is_dangerous": false}
  false }            DiscountRate: nil(省略)       ← discount_rate はキーなし
```

---

## マルチテナント分離

全エンドポイントで `clinic_id` によるテナント分離を徹底している。

```
JWT claims → c.Set("clinic_id", ...) → extractClinicID() → Service → Repository
                                                                        ↓
                                              WHERE clinic_id = ? が全クエリに付く
```

`extractClinicID()` が失敗（JWTに clinic_id なし等）した場合はハンドラが即 `return` するため、
**Repository に clinic_id なしのクエリが到達することはない**。
