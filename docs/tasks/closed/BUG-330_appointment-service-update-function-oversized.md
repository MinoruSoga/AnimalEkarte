# BUG-330: appointment_service.go の Update/checkSlotConflict 関数が規約上限(50行)を超過

## 概要
`appointment_service.go` の `Update()` 関数（92行）と `checkSlotConflict()` 関数（65行）がプロジェクト規約の50行上限を超過している。また `auth_handler.go` の `Login()` 関数（171行）も同様に超過している。関数が大きすぎるとテスト困難・可読性低下・責務の混在が発生する。

## 再現手順
1. `backend/internal/service/appointment_service.go` を開く
2. `Update()` 関数 (行167-259 = 92行)、`checkSlotConflict()` (行100-165 = 65行) を確認
3. `backend/internal/handler/auth_handler.go` の `Login()` 関数 (行99-270 = 171行) を確認
4. **結果**: 全て50行を超過

## 期待する動作
- 各関数は50行以内
- 競合チェックロジックは `resolveConflictCheckParams()` 等の補助関数に抽出
- `Login()` は認証・トークン発行・クリニック取得・監査ログを別関数に分割

## 現状コード

### `backend/internal/service/appointment_service.go:167` (92行)
```go
func (s *reservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Appointment, error) {
    // ... 競合チェックパラメータ解決 + トランザクション = 92行
}
```

### `backend/internal/service/appointment_service.go:100` (65行)
```go
func checkSlotConflict(ctx context.Context, tx *gorm.DB, clinicID uint64, doctorID *uint64, startTime, endTime time.Time, excludeID *uint64) error {
    // ... 医師指定/未指定の2パスチェック = 65行
}
```

### `backend/internal/handler/auth_handler.go:99` (171行)
```go
func (h *Handler) Login(c *gin.Context) {
    // 認証 + スタッフ取得 + クリニック取得 + JWT生成 + Cookie設定 + 監査ログ = 171行
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/service/appointment_service.go:167` | `Update()` 92行 | 要修正 |
| `backend/internal/service/appointment_service.go:100` | `checkSlotConflict()` 65行 | 要修正 |
| `backend/internal/handler/auth_handler.go:99` | `Login()` 171行 | 要修正 |

## 修正方針

### 1. `checkSlotConflict` を doctor 指定/未指定で分割 — `backend/internal/service/appointment_service.go:100`
```go
func checkSlotConflict(ctx context.Context, tx *gorm.DB, clinicID uint64, doctorID *uint64, startTime, endTime time.Time, excludeID *uint64) error {
    if doctorID != nil {
        return checkDoctorSlotConflict(ctx, tx, clinicID, *doctorID, startTime, endTime, excludeID)
    }
    return checkCapacitySlotConflict(ctx, tx, clinicID, startTime, endTime, excludeID)
}
```

### 2. `Update` のトランザクション内ロジックを `executeReservationUpdateTx` に抽出 — `backend/internal/service/appointment_service.go:167`
```go
func (s *reservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Appointment, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput("input must not be nil")
    }
    fields := buildReservationUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput("at least one field must be provided")
    }
    needsConflictCheck := input.StartTime != nil || input.EndTime != nil || input.DoctorID != nil
    if !needsConflictCheck {
        return s.updateFieldsDirect(ctx, clinicID, id, fields)
    }
    return s.updateWithConflictCheck(ctx, clinicID, id, fields, input)
}
```

### 3. `Login` の認証ロジックを `buildLoginResponse` 補助関数に抽出 — `backend/internal/handler/auth_handler.go:99`
責務分割:
- `verifyCredentials(ctx, input)` — account/staff 取得 + パスワード検証
- `issueTokenCookies(c, staff, account)` — JWT 生成 + Cookie 設定
- 監査ログ呼び出しは `Login()` に残す

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — コード品質
> **大きすぎる関数**: 50行超

### `.claude/rules/go-language.md` — パッケージ1ファイル
> `パッケージ1ファイル < 500行`（関数単位では50行）

### プロジェクト内参照実装
- `backend/internal/service/cage_service.go:68-83` — 20行以内の Delete 関数（依存チェック→削除→ログ）

## 優先度
**Medium** — 機能的なバグはないが、テスタビリティ・可読性・今後の変更コストに影響する

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/appointment_service.go:100-165` — `checkSlotConflict()`
- `backend/internal/service/appointment_service.go:167-259` — `Update()`
- `backend/internal/handler/auth_handler.go:99-270` — `Login()`
