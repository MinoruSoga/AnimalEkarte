---
status: closed
closed_at: 2026-03-15
---

# 診療項目マスタ 5テーブル: parent_id による親子構造追加

## 背景

`medicines` の `parent_id` 導入（issue 024）と同仕様で、診療項目マスタページ（`TreatmentItemsSettings.tsx`）に含まれる以下 5 テーブルにも `parent_id` を追加する。

**対象テーブル:**

| タブ | テーブル | Goモデル |
|------|---------|---------|
| 診察 | `consultations` | `model.Consultation` |
| 検査 | `exam_types` | `model.ExaminationType` |
| 処置 | `procedures` | `model.Procedure` |
| 予防接種 | `vaccines` | `model.Vaccine` |
| 定期健診 | `checkup_types` | `model.CheckupType` |

## データモデル（medicines と共通仕様）

| ケース | parent_id | price | 意味 |
|--------|-----------|-------|------|
| カテゴリレコード | NULL | NULL | カテゴリヘッダー行として描画 |
| 子アイテム | 親の id | 設定あり | カテゴリ配下のアイテム行 |
| カテゴリなしアイテム | NULL | 設定あり | グループ外フラット行 |

## 変更内容

### 1. DB スキーマ（`backend/migrations/001_init.sql`）

以下の 5 テーブルそれぞれに同じカラムとインデックスを追加:

```sql
-- consultations
ALTER TABLE consultations ADD COLUMN parent_id bigint REFERENCES consultations(id) ON DELETE SET NULL;
CREATE INDEX idx_consultations_parent_id ON consultations(parent_id);

-- exam_types
ALTER TABLE exam_types ADD COLUMN parent_id bigint REFERENCES exam_types(id) ON DELETE SET NULL;
CREATE INDEX idx_exam_types_parent_id ON exam_types(parent_id);

-- procedures
ALTER TABLE procedures ADD COLUMN parent_id bigint REFERENCES procedures(id) ON DELETE SET NULL;
CREATE INDEX idx_procedures_parent_id ON procedures(parent_id);

-- vaccines
ALTER TABLE vaccines ADD COLUMN parent_id bigint REFERENCES vaccines(id) ON DELETE SET NULL;
CREATE INDEX idx_vaccines_parent_id ON vaccines(parent_id);

-- checkup_types
ALTER TABLE checkup_types ADD COLUMN parent_id bigint REFERENCES checkup_types(id) ON DELETE SET NULL;
CREATE INDEX idx_checkup_types_parent_id ON checkup_types(parent_id);
```

> 001_init.sql を直接編集すること（CREATE TABLE 定義内に `parent_id` カラムを追加）。

### 2. Go モデル（各モデルファイルに共通追加）

```go
// consultation.go / procedure.go / vaccine.go / checkup_type.go / examination_type.go
ParentID *uint64 `gorm:"column:parent_id" json:"parent_id,omitempty"`
```

自己参照リレーションは medicine 同様不要（フロントが parent_id でグルーピング）。

### 3. ハンドラリクエスト型（各 `*_request.go`）

```go
// create request（各エンティティ）
ParentID *uint64 `json:"parent_id"`

// update request（PATCH: ポインタ型 optional）
ParentID *uint64 `json:"parent_id"`
```

### 4. サービス層（各 `*_service.go`）

`buildXxxUpdateFields()` に `parent_id` を追加:

```go
if req.ParentID != nil {
    fields["parent_id"] = *req.ParentID
}
```

> `parent_id` を null にクリアする操作（子 → ungrouped 移動）は issue 024 と同じ設計判断に従うこと。

### 5. シードデータ（`backend/migrations/002_seed_master.sql`）

各テーブルのシードをカテゴリレコード先行 INSERT + 子レコードが `parent_id` 参照する形に書き換え。

```sql
-- 例: consultations
INSERT INTO consultations (id, clinic_id, name, is_active, sort_order) VALUES
    (100, 3, '一般診察',   true, 1),  -- カテゴリ（price なし）
    (101, 3, '専門診察',   true, 2),  -- カテゴリ
    ...

INSERT INTO consultations (id, clinic_id, name, price, parent_id, ...) VALUES
    (1, 3, '初診料', 3000, 100, ...),
    (2, 3, '再診料', 1500, 100, ...),
    ...
```

各テーブルのカテゴリ設計はフロントエンド実装担当者と調整すること。

### 6. codegen 実行

全モデル変更後に `make codegen` を実行し `frontend/src/types/generated/models.ts` を再生成する。

## 完了条件

- [ ] `001_init.sql`: 5テーブルに `parent_id` カラム追加、インデックス追加
- [ ] `model/consultation.go`: `ParentID *uint64` 追加
- [ ] `model/examination_type.go`: `ParentID *uint64` 追加
- [ ] `model/procedure.go`: `ParentID *uint64` 追加
- [ ] `model/vaccine.go`: `ParentID *uint64` 追加
- [ ] `model/checkup_type.go`: `ParentID *uint64` 追加
- [ ] 各 `*_request.go`: create/update request に `parent_id` 追加
- [ ] 各 `*_service.go`: `buildXxxUpdateFields` に `parent_id` 追加
- [ ] `002_seed_master.sql`: 各テーブルのシードをカテゴリ + 子構造に書き換え
- [ ] `make codegen` 実行 → `models.ts` 再生成
- [ ] `docker compose exec backend go build ./...` が通ること
- [ ] DB リセット（`make reset`）後に seed データが正常投入されること

## 関連イシュー

- issue 024: medicines の parent_id 導入（同仕様・先行実装）
- フロントエンド側の UI 対応は別途フロントイシューで管理
