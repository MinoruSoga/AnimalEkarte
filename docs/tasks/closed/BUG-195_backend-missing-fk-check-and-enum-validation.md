# BUG-195: バックエンド FK 依存チェック欠如（inventory_service）・enum バリデーション欠如（pet_service / estimate_service）

## 概要

3 つのバックエンドサービスで重要なバリデーションが欠落している:
1. `inventory_service.go` — 削除時に診察記録で使用中の在庫アイテムを FK チェックなしで削除できる
2. `estimate_service.go` — 見積作成時に `Subtotal`/`TaxTotal`/`TotalAmount` に負の値を受け入れる
3. `pet_service.go` — ペット更新時に `Gender`/`Status`/`AcquisitionType`/`DangerLevel` の enum 値バリデーションなし

`.claude/CLAUDE.md` のマスタ削除 FK 依存チェック MANDATORY ルールに違反している。

## 再現手順

### 問題1: inventory_service FK チェック欠如
1. 診察記録（treatment）で使用中の在庫アイテムを作成
2. `DELETE /api/inventory/:id` でそのアイテムを削除
3. **結果**: 409 Conflict ではなく 200 OK で削除される
4. 診察記録の在庫参照が孤立する

### 問題2: estimate_service 負の金額
1. `POST /api/estimates` で以下を送信:
   ```json
   { "subtotal": -1000, "tax_total": -100, "total_amount": -1100 }
   ```
2. **結果**: 400 ではなく 200 で作成される

### 問題3: pet_service enum バリデーション
1. `PUT /api/pets/:id` で無効な enum 値を送信:
   ```json
   { "gender": "INVALID_GENDER", "status": "UNKNOWN" }
   ```
2. **結果**: 400 ではなく 200 で更新される（DB に無効値が保存される可能性）

## 現状コード

### `backend/internal/service/inventory_service.go` — Delete メソッド
```go
// ❌ FK 依存チェックなし
func (s *InventoryService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)  // FK チェックなしで即削除
}
```

### `backend/internal/service/estimate_service.go:70-73` — Create メソッド
```go
// ❌ 負の金額バリデーションなし
func (s *EstimateService) Create(ctx context.Context, input CreateEstimateInput) (*Estimate, error) {
    // input.Subtotal, input.TaxTotal, input.TotalAmount の正値チェックなし
    return s.repo.Create(ctx, &Estimate{
        Subtotal:    input.Subtotal,    // 負値を受け入れる
        TaxTotal:    input.TaxTotal,
        TotalAmount: input.TotalAmount,
    })
}
```

### `backend/internal/service/pet_service.go` — Update メソッド
```go
// ❌ enum バリデーションなし
func (s *PetService) Update(ctx context.Context, id uint64, input UpdatePetInput) (*Pet, error) {
    // input.Gender, input.Status, input.AcquisitionType, input.DangerLevel の検証なし
    fields := buildPetUpdateFields(input)
    return s.repo.UpdateFields(ctx, id, fields)
}
```

### 比較: 正しい実装（参照実装）
```go
// ✅ MANDATORY: マスタ削除前の FK 依存チェック（CLAUDE.md より）
func (s *InventoryService) Delete(ctx context.Context, id uint64) error {
    count, err := s.repo.CountUsageByInventoryID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check inventory usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この在庫アイテムは診察記録で使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}

// ✅ 金額バリデーション
func validateEstimateAmounts(input CreateEstimateInput) error {
    if input.Subtotal < 0 || input.TaxTotal < 0 || input.TotalAmount < 0 {
        return apperrors.WrapInvalidInput("金額は 0 以上である必要があります")
    }
    return nil
}

// ✅ enum バリデーション
var validGenders = map[string]bool{"male": true, "female": true, "unknown": true}

func validatePetGender(gender string) error {
    if gender != "" && !validGenders[gender] {
        return apperrors.WrapInvalidInput(fmt.Sprintf("無効な性別値: %s", gender))
    }
    return nil
}
```

## 影響範囲

| 対象ファイル | 問題 | リスク | 状態 |
|---|---|---|---|
| `backend/internal/service/inventory_service.go` | Delete に FK チェックなし | 参照孤立・データ不整合 | 未修正 |
| `backend/internal/service/estimate_service.go` | Create に負の金額チェックなし | 不正な見積データ | 未修正 |
| `backend/internal/service/pet_service.go` | Update に enum バリデーションなし | 無効な enum 値が DB に保存 | 未修正 |

## 修正方針

### 1. `inventory_service.go` — FK 依存チェック追加

Repository に `CountUsageByInventoryID` を追加し、Service の Delete で呼び出す:

```go
// repository/inventory_repository.go
func (r *InventoryRepository) CountUsageByInventoryID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&TreatmentItem{}).
        Where("inventory_id = ?", id).
        Count(&count).Error
    return count, err
}

// service/inventory_service.go
func (s *InventoryService) Delete(ctx context.Context, id uint64) error {
    count, err := s.repo.CountUsageByInventoryID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check inventory usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この在庫アイテムは診察記録で使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

### 2. `estimate_service.go:70-73` — 負の金額バリデーション

```go
func (s *EstimateService) Create(ctx context.Context, input CreateEstimateInput) (*Estimate, error) {
    if input.Subtotal < 0 || input.TaxTotal < 0 || input.TotalAmount < 0 {
        return nil, apperrors.WrapInvalidInput("金額は 0 以上である必要があります")
    }
    return s.repo.Create(ctx, &Estimate{...})
}
```

### 3. `pet_service.go` — enum バリデーション

`internal/service/validators.go`（既存）に enum バリデーション関数を追加:

```go
var validPetGenders = map[string]bool{"male": true, "female": true, "unknown": true}
var validPetStatuses = map[string]bool{"active": true, "deceased": true}
// ... 他の enum

func validatePetUpdate(input UpdatePetInput) error {
    if input.Gender != nil && !validPetGenders[*input.Gender] {
        return apperrors.WrapInvalidInput(fmt.Sprintf("無効な性別: %s", *input.Gender))
    }
    // ... 他フィールドのチェック
    return nil
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — マスタ削除の FK 依存チェック (MANDATORY)
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。

### `.claude/rules/go-language.md` — エラーハンドリング
> Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。

## 優先度
**High** — inventory FK チェック欠如はデータ整合性問題（診察記録が参照先不明になる）。estimate の負の金額は業務データの不整合を引き起こす。次回リリースまでに対応が必要。

## 関連チケット
- BUG-176: マスタ Repository の clinic_id フィルタ欠如（同カテゴリ）
- BUG-177: マスタ Repository の Delete に clinic_id フィルタ欠如（同カテゴリ）

## 関連ファイル
- `backend/internal/service/inventory_service.go`
- `backend/internal/service/estimate_service.go`
- `backend/internal/service/pet_service.go`
- `backend/internal/repository/inventory_repository.go`
- `backend/internal/service/validators.go`
- `backend/internal/apperrors/errors.go`
