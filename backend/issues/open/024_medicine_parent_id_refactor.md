---
status: closed
closed_at: 2026-03-15
---

# Medicine: drug_category 廃止 → parent_id による親子構造への移行

## 背景

現行の `medicines.drug_category`（varchar）はカテゴリ名を文字列で保持しており、カテゴリ自体が DB レコードとして存在しない。そのため「空カテゴリの永続化」「カテゴリの並び替え」「カテゴリ自身のプロパティ管理」が構造的に不可能。

**新仕様**: カテゴリ自体も `medicines` レコードとして表現する。価格なし medicine = カテゴリヘッダー、price あり medicine = 子アイテム。`parent_id` で親子関係を管理する。

## データモデル（移行後）

| ケース | parent_id | price | 意味 |
|--------|-----------|-------|------|
| カテゴリ medicine | NULL | NULL | カテゴリヘッダー行として描画 |
| 子 medicine | 親の id | 設定あり | カテゴリ配下のアイテム行 |
| カテゴリなし medicine | NULL | 設定あり | グループ外フラット行 |

## 変更内容

### 1. DB スキーマ（`backend/migrations/001_init.sql`）

`medicines` テーブル:
- `drug_category varchar(100)` を削除
- `parent_id BIGINT NULL REFERENCES medicines(id) ON DELETE SET NULL` を追加

```sql
-- 変更前
drug_category    varchar(100),

-- 変更後
parent_id        bigint REFERENCES medicines(id) ON DELETE SET NULL,
```

インデックス追加:
```sql
CREATE INDEX idx_medicines_parent_id ON medicines(parent_id);
```

### 2. Go モデル（`backend/internal/model/medicine.go`）

```go
// 変更前
DrugCategory    *string       `gorm:"column:drug_category"   json:"drug_category,omitempty"`

// 変更後
ParentID        *uint64       `gorm:"column:parent_id"       json:"parent_id,omitempty"`
Parent          *Medicine     `gorm:"foreignKey:ParentID"    json:"parent,omitempty"`
```

### 3. ハンドラリクエスト型（`backend/internal/handler/medicine_request.go`）

```go
// createMedicineRequest
type createMedicineRequest struct {
    Name            string        `json:"name"             binding:"required"`
    ParentID        *uint64       `json:"parent_id"`       // drug_category → parent_id
    Price           *float64      `json:"price"`
    IsActive        *bool         `json:"is_active"`
    Description     *string       `json:"description"`
    DosageForm      *string       `json:"dosage_form"`
    MedicineUnit    *string       `json:"medicine_unit"`
    DefaultQuantity *int          `json:"default_quantity"`
    SortOrder       *int          `json:"sort_order"`
}

// updateMedicineRequest（PATCH: ポインタ型で全フィールド optional）
type updateMedicineRequest struct {
    Name            *string       `json:"name"`
    ParentID        *uint64       `json:"parent_id"`       // drug_category → parent_id
    // ... 他フィールドは現行維持
}
```

### 4. サービス層（`backend/internal/service/medicine_service.go`）

`CreateStaffInput` / `UpdateStaffInput` 相当の medicine 用 DTO に `ParentID *uint64` を追加、`DrugCategory` を削除。

`buildMedicineUpdateFields()` で `parent_id` の更新を処理:
```go
// parent_id は明示的に null セットが必要なため、リクエストに含まれている場合のみ更新
// （PATCH で "parent_id": null を送ると ungrouped に移動する）
if req.ParentID != nil {
    fields["parent_id"] = *req.ParentID
}
// parent_id を null にしたい場合は別フラグが必要（要検討）
```

> **注意**: `parent_id` を null にクリアする操作（子 → ungrouped 移動）は `*uint64` では表現できない。
> 別途 `clear_parent_id bool` フラグを request に追加するか、ゼロ値を「クリア」として扱う設計を検討すること。

### 5. FindAll クエリ（`backend/internal/repository/medicine_repository.go`）

フラットリストを返す（フロントエンドが `parent_id` でグルーピング）。`drug_category` の `ORDER BY` を削除。

```go
// 変更前
Order("sort_order ASC, name ASC")

// 変更後（同じ。ソート順は変更なし）
Order("sort_order ASC, name ASC")
```

### 6. シードデータ（`backend/migrations/002_seed_master.sql`）

カテゴリ medicine レコード（price なし）を先に INSERT し、子 medicine レコードが `parent_id` で参照する形に書き換え。

```sql
-- カテゴリ medicine（price なし = カテゴリヘッダー）
INSERT INTO medicines (id, clinic_id, name, is_active, sort_order) VALUES
    (100, 3, '抗生剤',      true, 1),
    (101, 3, 'ステロイド',  true, 2),
    (102, 3, '消炎剤',      true, 3),
    ...

-- 子 medicine（parent_id あり）
INSERT INTO medicines (id, clinic_id, name, price, parent_id, ...) VALUES
    (1,  3, 'アモキシシリン 50mg',  500, 100, ...),
    (2,  3, 'メトロニダゾール 250mg', 600, 100, ...),
    (3,  3, 'プレドニゾロン 5mg',   400, 101, ...),
    ...
```

### 7. codegen 実行

モデル変更後に `make codegen` を実行し `frontend/src/types/generated/models.ts` を再生成する。

## 完了条件

- [ ] `001_init.sql`: `drug_category` 削除、`parent_id` 追加、インデックス追加
- [ ] `model/medicine.go`: `DrugCategory` → `ParentID *uint64` + `Parent *Medicine` relation
- [ ] `medicine_request.go`: create/update request を `parent_id` ベースに更新
- [ ] `medicine_service.go`: DTO + `buildMedicineUpdateFields` を更新
- [ ] `medicine_repository.go`: `drug_category` 参照箇所を削除
- [ ] `002_seed_master.sql`: カテゴリ medicine レコード + 子 medicine の `parent_id` 参照に書き換え
- [ ] `make codegen` 実行 → `models.ts` 再生成
- [ ] `docker compose exec backend go build ./...` が通ること
- [ ] DB リセット（`make reset`）後に seed データが正常投入されること

## 影響範囲

- フロントエンド側も `drug_category` → `parent_id` への対応が必要（別途フロントイシューで管理）
- `treatments`, `care_plan_items` 等で `medicine_id` 参照しているテーブルへの影響なし（medicine レコード削除なし）
