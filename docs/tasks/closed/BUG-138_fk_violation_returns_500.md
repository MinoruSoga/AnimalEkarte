# BUG-138: 外部キー違反・整数オーバーフローが 500 Internal Server Error を返す

## 概要
存在しない外部キー ID や整数オーバーフロー値でリソースを作成すると、
500 Internal Server Error が返される。DB の FK 制約違反や型エラーが
Service/Repository 層で適切にハンドリングされず、汎化エラーとして 500 が返る。

ユーザーには 400 Bad Request（入力値エラー）または 409 Conflict（参照先不存在）で返すべき。

## 脆弱性分類
- **CWE-755**: Improper Handling of Exceptional Conditions
- **影響**: セキュリティ実害なし（BUG-129 修正によりエラーメッセージは `internal server error` に汎化済み）。
  ただし 500 エラーはサーバー側の予期しない状態を示し、ログにスタックトレースが出力される。

## 再現手順と結果

| テスト | リクエスト | 期待 | 実際 |
|--------|----------|------|------|
| 存在しない occupation_id | `POST /masters/staffs {"name":"t","occupation_id":999999}` | 400 | **500** |
| 存在しない FK (reservation) | `POST /reservations {"pet_id":999999,"owner_id":999999,...}` | 400 | **500** |
| 整数オーバーフロー | `POST /masters/staffs {"name":"t","sort_order":99999999999999}` | 400 | **500** |

## エラーレスポンス
```json
{"error": "internal server error"}
```
メッセージは汎化されている（BUG-129 修正効果）が、HTTP ステータスコードが不適切。

## 現状コード（推定）

### Repository 層
```go
func (r *StaffRepository) Create(ctx context.Context, staff *model.Staff) error {
    return r.db.WithContext(ctx).Create(staff).Error
    // ❌ GORM の FK 違反エラーが raw error のまま返る
    // apperrors.FromGORM() で変換されていない可能性
}
```

### 期待する処理
```go
func (r *StaffRepository) Create(ctx context.Context, staff *model.Staff) error {
    if err := r.db.WithContext(ctx).Create(staff).Error; err != nil {
        return apperrors.FromGORM(err, "staff", "")
        // FK 違反 → apperrors.ErrInvalidInput or ErrConflict → 400 or 409
    }
    return nil
}
```

## 修正方針

### 1. `apperrors.FromGORM` に FK 違反ハンドリングを追加

```go
func FromGORM(err error, resource string, id string) error {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return fmt.Errorf("not found: %w", ErrNotFound)
    }
    // FK 違反
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
        return fmt.Errorf("invalid input: %w", ErrInvalidInput)
    }
    // 数値範囲エラー
    if errors.As(err, &pgErr) && pgErr.Code == "22003" { // numeric_value_out_of_range
        return fmt.Errorf("invalid input: %w", ErrInvalidInput)
    }
    return fmt.Errorf("internal error: %w", err)
}
```

### 2. Handler の RespondError で 400 にマッピング

`ErrInvalidInput` → 400 Bad Request（既存のマッピング）

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換

FK 違反も `FromGORM` で適切に変換すべき。

### `.claude/rules/error-handling.md`
> HTTP Status マッピング: 400 Bad Request = 入力値エラー

FK 違反は「ユーザーが無効な参照 ID を指定した」入力エラーであり、400 が正しい。

### `.claude/rules/api.md`
> "Return consistent error response format" / "Use proper HTTP status codes"

500 は予期しないサーバーエラーにのみ使用すべき。FK 違反は予期可能なエラー。

## 優先度
**Medium** — 500 エラーはサーバー監視で誤報を生む。エラーメッセージの漏洩はないが、
ステータスコードの修正とエラーハンドリングの一貫性のために対応すべき。

## 関連チケット
- BUG-129（修正済み）: エラーメッセージ漏洩
- BUG-135: バリデーションエラーフィールド名漏洩

## 関連ファイル
- `backend/internal/errors/errors.go` — FromGORM（FK 違反ハンドリング追加）
- `backend/internal/repository/staff_repository.go` — Create（例）
- `backend/internal/repository/reservation_repository.go` — Create（例）
- `backend/internal/handler/response.go` — RespondError のステータスマッピング
