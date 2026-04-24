# CODE-QUALITY-204: payment_method_master 全レイヤー品質修正

## 概要

`payment_method_master` は他マスタと比較して複数の規約違反・実装欠落がある。
Service のバリデーション不足、Repository の RowsAffected チェック欠落、
インターフェース設計の不整合など、全レイヤーで修正が必要。

## 優先度

MEDIUM（複数の軽微問題の集合）

## 影響ファイル

| ファイル | 問題数 |
|---------|--------|
| `backend/internal/service/payment_method_master_service.go` | 3件 |
| `backend/internal/repository/payment_method_master_repository.go` | 2件 |
| `backend/internal/handler/payment_method_master_handler.go` | 1件 |

---

## 問題一覧

### [Service] 1. validateRequiredName 未呼び出し（Create）

```go
// 現状 — バリデーションなし
func (s *paymentMethodMasterService) Create(...) {
    m := &model.PaymentMethodMaster{
        Name: input.Name,   // 空文字が通過する
        ...
    }
```

**修正方針**:
```go
func (s *paymentMethodMasterService) Create(...) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    ...
}
```

---

### [Service] 2. validateOptionalName 未呼び出し（Update）

```go
// 現状 — Name の空文字チェックなし
func (s *paymentMethodMasterService) Update(...) {
    if input == nil { ... }
    fields := buildPaymentMethodUpdateFields(input)
    ...
```

**修正方針**:
```go
func (s *paymentMethodMasterService) Update(...) {
    if input == nil { ... }
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildPaymentMethodUpdateFields(input)
    ...
}
```

---

### [Service] 3. buildPaymentMethodUpdateFields のカラム名が文字列リテラル直書き

```go
// 現状 — リテラル直書き
func buildPaymentMethodUpdateFields(input *UpdatePaymentMethodInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name    // リテラル
    }
    if input.DisplayOrder != nil {
        fields["display_order"] = *input.DisplayOrder  // リテラル
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive  // リテラル
    }
    return fields
}
```

**修正方針**: 他マスタと同様に定数を定義する。
```go
const (
    colPaymentMethodName         = "name"
    colPaymentMethodDisplayOrder = "display_order"
    colPaymentMethodIsActive     = "is_active"
)
```

---

### [Repository] 4. UpdateFields で RowsAffected チェックなし

```go
// 現状 — RowsAffected を確認せず FindByID に委ねる
if err := r.db.WithContext(ctx).
    Model(&model.PaymentMethodMaster{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).
    Updates(fields).Error; err != nil {
    return nil, apperrors.FromGORM(err, "payment_method", ...)
}
return r.FindByID(ctx, clinicID, id)  // 存在しない場合も FindByID の Not Found に依存
```

**修正方針**:
```go
result := r.db.WithContext(ctx).
    Model(&model.PaymentMethodMaster{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).
    Updates(fields)
if result.Error != nil {
    return nil, apperrors.FromGORM(result.Error, "payment_method", fmt.Sprintf("%d", id))
}
if result.RowsAffected == 0 {
    return nil, apperrors.WrapNotFound("payment_method", fmt.Sprintf("%d", id))
}
return r.FindByID(ctx, clinicID, id)
```

---

### [Repository] 5. CountUsageByID の命名が規約と不統一

他マスタ: `CountUsageByExamTypeID`, `CountUsageByAnimalSpeciesID` 等（エンティティ名を含む）
payment_method: `CountUsageByID`（エンティティ名なし・曖昧）

**修正方針**: インターフェース・実装・呼び出し箇所を一括でリネーム。
```go
// 変更前
CountUsageByID(ctx context.Context, clinicID, id uint64) (int64, error)
// 変更後
CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error)
```

---

### [Handler] 6. GetByID エンドポイント欠如 / FindByID のインターフェース公開の不整合

Repository に `FindByID` が存在するが Service インターフェースに `GetByID` がなく、
Handler にも `GET /:id` エンドポイントがない。

**対応方針**: 以下のどちらかを選択する。
- **A（推奨）**: Service に `GetByID` を追加し、Handler に `GET /payment-methods/:id` を追加。他マスタとの API 対称性を確保。
- **B**: `FindByID` を Repository インターフェースから削除し、`UpdateFields`/`Delete` 内部のプライベートメソッド的な位置づけにする（仕様上 GET 単件が不要の場合）。

---

## 規約参照

- `.claude/CLAUDE.md`: Service バリデーションパターン
- `.claude/rules/naming-conventions.md`: メソッド名命名規則

## テスト

- `Create` に空文字名を渡した場合に 400 が返ることを確認
- `Update` に `name: ""` を渡した場合に 400 が返ることを確認
- 存在しない ID を `Update` した場合に 404 が返ることを確認
