# CODE-QUALITY-007: Repository 層実装パターン統一

## 概要

マスタ系 Repository 層に複数の実装パターン不統一がある。  
`clinicScope` の未使用、`billing.deleted_at IS NULL` 欠落による論理削除済みレコードの誤カウント、
`tzdata` 欠落による panic リスクなど。

## 優先度

MEDIUM（`merchandise_item_repository.go` の項目は HIGH）

## 影響ファイル

| ファイル | 問題 | 重大度 |
|---------|-----|------|
| `backend/internal/repository/merchandise_item_repository.go` | L91: billings.deleted_at IS NULL 欠落 | **HIGH** |
| `backend/internal/repository/reservation_type_unavailable_time_repository.go` | L59: clinicScope 非使用 | MEDIUM |
| `backend/internal/repository/reservation_type_occupation_repository.go` | L16: tzdata 欠落 / L62: isUniqueConstraintErr なし | MEDIUM |
| `backend/internal/repository/payment_method_master_repository.go` | L72-78: UpdateFields の実装が非統一 | MEDIUM |
| `backend/internal/repository/animal_species_repository.go` | グローバルマスタ設計の文書化不足 | MEDIUM |

---

## 問題一覧

### [HIGH] 1. `merchandise_item_repository.go:91` — `billings.deleted_at IS NULL` 欠落

`CountUsageByMerchandiseItemID` の `billing_items` JOIN において `billings.deleted_at IS NULL` が欠落している。  
`estimate_items` の JOIN には `estimates.deleted_at IS NULL` が付いているが、`billing_items` 側のみ欠落しており非対称。

**問題**: 論理削除済みの請求（billing）から参照されている物販品を「使用中」と誤カウントし、
未使用の物販品マスタを削除できなくなる。

**修正方針**:
```go
// 修正前（欠落）
Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ?", clinicID).

// 修正後
Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
```

---

### 2. `reservation_type_unavailable_time_repository.go:59` — `clinicScope` 非使用

```go
// 現状: 手動 WHERE 条件
Where("clinic_id = ? AND id = ?", clinicID, id).Delete(...)

// 他の全リポジトリ
Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(...)
```

**修正方針**: `clinicScope` を使用するパターンに統一。

---

### 3. `reservation_type_occupation_repository.go:16` — `tzdata` 欠落による panic リスク

```go
// パッケージレベルで time.LoadLocation を即時実行
var jstLoc = func() *time.Location {
    loc, err := time.LoadLocation("Asia/Tokyo")
    if err != nil {
        panic(err)  // Alpine イメージなど tzdata がない環境で panic
    }
    return loc
}()
```

`time.LoadLocation` は OS の tzdata ファイルに依存する。  
Alpine Linux（本番 Docker イメージ）には tzdata が含まれないため起動時 panic が発生するリスクがある。

**修正方針**: `import _ "time/tzdata"` を追加し、Go 標準ライブラリの tzdata を埋め込む。
```go
import _ "time/tzdata"  // Alpine 等 tzdata 未インストール環境での panic を防ぐ
```

---

### 4. `reservation_type_occupation_repository.go:62` — `Create` の `isUniqueConstraintErr` チェックなし

`reservation_type_occupations` テーブルは `(clinic_id, reservation_type_id, occupation_id)` のユニーク制約を持つが、
`Create` で重複時のエラーメッセージが他の全 Create（`WrapConflict` で日本語メッセージ）と統一されていない。

**修正方針**:
```go
if err := r.db.WithContext(ctx).Create(rto).Error; err != nil {
    if isUniqueConstraintErr(err) {
        return apperrors.WrapConflict("同じ職種が既に登録されています")
    }
    return apperrors.FromGORM(err, "reservation_type_occupation", "")
}
```

---

### 5. `payment_method_master_repository.go:72-78` — `UpdateFields` の実装が非統一

他の全リポジトリは `UpdateFields` の末尾で `return r.FindByID(ctx, clinicID, id)` を呼ぶが、
このリポジトリのみインライン実装になっている。

**修正方針**:
```go
func (r *paymentMethodMasterRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
    if err := r.db.WithContext(ctx).
        Model(&model.PaymentMethodMaster{}).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        Updates(fields).Error; err != nil {
        return nil, apperrors.FromGORM(err, "payment_method", fmt.Sprintf("%d", id))
    }
    return r.FindByID(ctx, clinicID, id)  // 統一パターン
}
```

---

### 6. `animal_species_repository.go` — グローバルマスタ設計の文書化不足

`AnimalSpeciesRepository` の全メソッドが `clinic_id` を持たない「システム共通マスタ」だが、
他のリポジトリと異なる設計方針について説明がない。コードを読む人が「clinic_id 取得漏れ」と誤解する可能性がある。

**修正方針**: インターフェース定義にコメントを追加。
```go
// AnimalSpeciesRepository は動物種マスタのデータアクセス層。
// 動物種はシステム全体で共有されるグローバルマスタであり、clinic_id を持たない。
type AnimalSpeciesRepository interface {
```

---

## 規約参照

- `.claude/rules/error-handling.md`: Repository は `apperrors.FromGORM()` でエラー変換
- `.claude/rules/database-design.md`: マルチテナント設計（clinic_id 必須）
- `.claude/rules/docker-rules.md`: Docker Alpine イメージの制約

## テスト

- `merchandise_item` の削除チェックで、削除済み billing に紐づく物販品が「未使用」と判定されることを確認
- `reservation_type_occupation` の重複 Create で 409 Conflict が返ることを確認
- Docker Alpine 環境（または tzdata なし環境）で起動 panic が発生しないことを確認
