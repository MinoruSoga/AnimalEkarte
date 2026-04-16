# BUG-176: マスタ系 Repository の FindByID/FindAll に clinic_id フィルタ欠如（マルチテナント違反）

## 概要

20 以上のマスタ系 Repository で `FindByID` / `FindAll` 実行時に `clinic_id` フィルタが存在しない。`id = ?` のみで検索しているため、**テナントAのユーザーがテナントBのマスタデータを参照できる可能性**がある。マルチテナント設計規約の根幹に関わる重大な違反。

## 脆弱性分類
- **CWE-284**: Improper Access Control
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 悪意あるユーザーが他テナントのマスタデータ（診断名、処置内容、薬剤、ケージ情報等）を参照できる可能性

## 再現手順

1. テナント A で作成されたマスタデータの ID を確認（例: 診察種別 ID = 1）
2. テナント B のアカウントで GET `/v1/checkup-types/1` を実行
3. **結果**: テナントBのトークンでテナントAのデータが返される（clinic_id による隔離がない）

## 期待する動作

```go
// ✅ 正しい: clinic_id を含めた検索
err := r.db.WithContext(ctx).
    First(&checkupType, "id = ? AND clinic_id = ?", id, clinicID).Error
```

## 現状コード

### `backend/internal/repository/checkup_type_repository.go:43`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&checkupType, "id = ?", id).Error
```

### `backend/internal/repository/exam_type_repository.go:43`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).Preload("Items").First(&exType, "id = ?", id).Error
```

### `backend/internal/repository/procedure_repository.go:40`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&procedure, "id = ?", id).Error
```

### `backend/internal/repository/vaccine_repository.go:44`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&vaccine, "id = ?", id).Error
```

### `backend/internal/repository/cage_repository.go:43`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&cage, "id = ?", id).Error
```

### `backend/internal/repository/insurance_repository.go:41`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&insurance, "id = ?", id).Error
```

### `backend/internal/repository/occupation_repository.go:46`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&occupation, "id = ?", id).Error
```

### `backend/internal/repository/consultation_repository.go:43`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&consultation, "id = ?", id).Error
```

### `backend/internal/repository/chief_complaint_category_repository.go:44`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
```

### `backend/internal/repository/inquiry_template_repository.go:44`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
```

### `backend/internal/repository/billing_item_repository.go:33`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
```

### `backend/internal/repository/trimming_master_repository.go:42,143`
```go
// ❌ clinic_id なし（TrimmingCourse / TrimmingOption）
err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error
err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
```

### `backend/internal/repository/staff_repository.go:62`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).Preload("Account").Preload("Occupation").First(&staff, "id = ?", id).Error
```

### `backend/internal/repository/hospitalization_plan_repository.go:43`
```go
// ❌ clinic_id なし
err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
```

### FindAll に clinic_id フィルタ欠如

| ファイルパス | 行番号 | 問題 |
|---|---|---|
| `checkup_type_repository.go` | 34 | `.Find(&checkupTypes)` — clinic_id フィルタなし |
| `exam_type_repository.go` | 34 | `.Find(&exTypes)` — clinic_id フィルタなし |
| `procedure_repository.go` | 32 | `.Find(&procedures)` — clinic_id フィルタなし |

### 比較: 正しい実装（参照実装）
```go
// ✅ owner_repository.go — clinic_id を含めた正しい実装
err := r.db.WithContext(ctx).
    First(&owner, "id = ? AND clinic_id = ?", id, clinicID).Error
```

## 影響範囲

| 対象 Repository | 違反メソッド | リスク |
|---|---|---|
| `checkup_type_repository.go` | FindByID, FindAll | 他テナントのデータ参照 |
| `exam_type_repository.go` | FindByID, FindAll | 他テナントのデータ参照 |
| `procedure_repository.go` | FindByID, FindAll | 他テナントのデータ参照 |
| `vaccine_repository.go` | FindByID | 他テナントのデータ参照 |
| `cage_repository.go` | FindByID | 他テナントのデータ参照 |
| `insurance_repository.go` | FindByID | 他テナントのデータ参照 |
| `occupation_repository.go` | FindByID | 他テナントのデータ参照 |
| `consultation_repository.go` | FindByID | 他テナントのデータ参照 |
| `chief_complaint_category_repository.go` | FindByID | 他テナントのデータ参照 |
| `inquiry_template_repository.go` | FindByID | 他テナントのデータ参照 |
| `billing_item_repository.go` | FindByID | 他テナントのデータ参照 |
| `trimming_master_repository.go` | FindByID (×2) | 他テナントのデータ参照 |
| `staff_repository.go` | FindByID | 他テナントのデータ参照 |
| `hospitalization_plan_repository.go` | FindByID | 他テナントのデータ参照 |

## 修正方針

### 全 Repository の FindByID シグネチャに clinicID を追加

```go
// Before
func (r *CheckupTypeRepository) FindByID(ctx context.Context, id uint64) (*model.CheckupType, error) {
    var checkupType model.CheckupType
    err := r.db.WithContext(ctx).First(&checkupType, "id = ?", id).Error

// After
func (r *CheckupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
    var checkupType model.CheckupType
    err := r.db.WithContext(ctx).
        First(&checkupType, "id = ? AND clinic_id = ?", id, clinicID).Error
```

### Interface / Service 層も同様に clinicID を伝播

```go
// service/checkup_type_service.go
func (s *CheckupTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
    return s.repo.FindByID(ctx, clinicID, id)
}

// handler/checkup_type_handler.go
clinicID := getClinicID(c)  // JWTクレームから取得
result, err := h.service.GetByID(c.Request.Context(), clinicID, id)
```

### FindAll の clinic_id 追加

```go
// Before
err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error

// After
err := r.db.WithContext(ctx).
    Where("clinic_id = ?", clinicID).
    Order("sort_order ASC, name ASC").
    Find(&checkupTypes).Error
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md` — マルチテナント設計
> **❌ 危険: clinic_id なしのクエリ（データリーク可能性）**
> `SELECT * FROM owners WHERE id = 1;`
>
> **✅ 安全: 常に clinic_id を条件に含める**
> `SELECT * FROM owners WHERE clinic_id = $1 AND id = $2;`

### プロジェクト内参照実装
- `backend/internal/repository/owner_repository.go` — `First(&owner, "id = ? AND clinic_id = ?", id, clinicID)` ✅

## 優先度
**High** — 現時点では単一クリニックの運用だが、マルチテナント対応が本来の設計目標であり、テナント分離不足はデータリークに直結する。次回リリースまでに対応が必要。

## 関連チケット
- BUG-177: clinic_id なし Delete（削除操作の同様の違反）

## 関連ファイル
- `backend/internal/repository/` 配下の全マスタ系 repository ファイル
- `.claude/rules/database-design.md`
