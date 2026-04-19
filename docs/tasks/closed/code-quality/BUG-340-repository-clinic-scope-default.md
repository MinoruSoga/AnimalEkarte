# BUG-340: Repository の clinic_id フィルタが手動運用で漏れやすい

## 優先度
HIGH

## 背景

現在、全 repository メソッドが `WHERE clinic_id = ?` を**手動で**記述する運用になっている。
BUG-339 の監査で、JOIN 先テーブル（`medical_records` 等）の `clinic_id` / `deleted_at` フィルタが
複数箇所で抜けていたことが判明し修正した。

この構造は「書き忘れ = 別クリニックのデータリーク」に直結するため、デフォルト安全な設計に改める必要がある。

## 現状の問題

```go
// 現在: 各メソッドが手動フィルタ（漏れやすい）
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
    var owner model.Owner
    err := r.db.WithContext(ctx).
        Where("owners.clinic_id = ? AND owners.id = ?", clinicID, id).  // 手動
        First(&owner).Error
    ...
}
```

### リスク
1. **clinic_id フィルタ漏れ** → 別クリニックのデータが取得・更新・削除される（データリーク）
2. **JOIN 先の soft delete フィルタ漏れ** → 論理削除済みレコード経由でデータが漏れる
3. **新規 repository 実装時に同じミスを繰り返す**

## 対応方針

### 1. `clinicScope` ヘルパーの標準化

```go
// backend/internal/repository/base.go（新規作成）

// clinicScope は全クエリに clinic_id フィルタをデフォルト適用するスコープ。
// JOIN 先テーブルには適用されないため、JOIN 条件側での明示指定は引き続き必要。
func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("clinic_id = ?", clinicID)
    }
}

// softDeletedJoin は医療記録等を JOIN する際の標準条件テンプレート。
// テーブル名を引数に取り、clinic_id + deleted_at IS NULL を AND 条件として返す。
const joinMedicalRecords = "JOIN medical_records ON medical_records.id = %s.medical_record_id" +
    " AND medical_records.deleted_at IS NULL"
```

### 2. 全 repository での `clinicScope` 使用を統一

```go
// ✅ 改善後
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
    var owner model.Owner
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        First(&owner).Error
    ...
}
```

### 3. JOIN 先フィルタのコードレビューチェックリスト追加

`CLAUDE.md` のバックエンド規約に以下を追記：

```
JOIN を含む repository メソッドは必ず以下を確認：
- [ ] JOIN 先テーブルの clinic_id フィルタが JOIN 条件に含まれているか
- [ ] JOIN 先テーブルの deleted_at IS NULL が JOIN 条件に含まれているか
```

## 対象ファイル

### `clinicScope` 導入が有効な repository（主要なもの）
- `owner_repository.go`
- `pet_repository.go`
- `medical_record_repository.go`
- `appointment_repository.go`
- `billing_repository.go`
- `staff_repository.go`
- その他全 repository（`backend/internal/repository/*.go`）

### JOIN フィルタの確認が必要なもの（BUG-339 で修正済みだが将来的に漏れやすい）
- `medical_records` を JOIN する全 repository
- `clinic_id` を直接持たずに親テーブル経由でテナント判定するテーブル

## 受け入れ条件

- [ ] `backend/internal/repository/base.go` に `clinicScope` ヘルパーを実装
- [ ] 全 repository で `clinicScope` を使用（手動 `WHERE clinic_id = ?` を排除）
- [ ] `backend/CLAUDE.md` に JOIN フィルタチェックリストを追記
- [ ] `golangci-lint run ./...` が通過
- [ ] `go test ./...` が通過

## 関連

- BUG-339: backend Go 規約監査（soft delete JOIN フィルタ漏れを発見・修正）
- `.claude/rules/database-design.md`: マルチテナント設計規約

## 作成日
2026-04-13
