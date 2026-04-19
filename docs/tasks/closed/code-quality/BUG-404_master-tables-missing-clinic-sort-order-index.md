# BUG-404: マスタテーブルの一覧取得で使用する (clinic_id, sort_order) 複合インデックスが欠落

## 概要
マスタ一覧取得クエリは `ORDER BY sort_order ASC` を使用しているが、`(clinic_id, sort_order)` の複合インデックスが存在するのは `merchandise_items` のみ。他の全マスタテーブルには `clinic_id` 単独インデックスしかなく、一覧取得時に `clinic_id` でフィルタ後に全件ソートが発生する。データ量増加時にパフォーマンス低下が予想される。

## 再現手順
```sql
EXPLAIN ANALYZE
SELECT * FROM vaccines WHERE clinic_id = 1 AND deleted_at IS NULL ORDER BY sort_order ASC;
-- → Seq Scan or Index Scan (clinic_id only) + Sort になる可能性
```

## 現状コード

### `backend/migrations/001_init.sql:1403`（唯一の sort_order インデックス）
```sql
CREATE INDEX idx_merchandise_items_sort ON merchandise_items(clinic_id, sort_order);
-- ↑ merchandise_items のみに存在。他マスタには欠落
```

### 他マスタのインデックス（例）
```sql
CREATE INDEX idx_vaccines_clinic_id ON vaccines(clinic_id);               -- sort_order なし
CREATE INDEX idx_medicines_clinic_id ON medicines(clinic_id);             -- sort_order なし
CREATE INDEX idx_exam_types_clinic_id ON exam_types(clinic_id);           -- sort_order なし
CREATE INDEX idx_procedures_clinic_id ON procedures(clinic_id);           -- sort_order なし
CREATE INDEX idx_cages_clinic_id ON cages(clinic_id);                     -- sort_order なし
```

## 影響範囲

以下のテーブルすべてに `(clinic_id, sort_order)` 複合インデックスの追加が必要：

| テーブル | 現在のインデックス | 追加すべきインデックス |
|---------|----------------|-------------------|
| `vaccines` | `idx_vaccines_clinic_id` | `idx_vaccines_clinic_sort` |
| `medicines` | `idx_medicines_clinic_id` | `idx_medicines_clinic_sort` |
| `exam_types` | `idx_exam_types_clinic_id` | `idx_exam_types_clinic_sort` |
| `procedures` | `idx_procedures_clinic_id` | `idx_procedures_clinic_sort` |
| `cages` | `idx_cages_clinic_id` | `idx_cages_clinic_sort` |
| `checkup_types` | `idx_checkup_types_clinic_id` | `idx_checkup_types_clinic_sort` |
| `chief_complaint_types` | — | `idx_chief_complaints_clinic_sort` |
| `diagnosis_types` | — | `idx_diagnosis_types_clinic_sort` |
| `diagnosis_names` | — | `idx_diagnosis_names_clinic_sort` |
| `trimming_courses` | — | `idx_trimming_courses_clinic_sort` |
| `trimming_options` | — | `idx_trimming_options_clinic_sort` |
| `insurances` | — | `idx_insurances_clinic_sort` |
| `occupations` | — | `idx_occupations_clinic_sort` |
| `reservation_types` | — | `idx_reservation_types_clinic_sort` |
| `reservation_type_groups` | — | `idx_reservation_type_groups_clinic_sort` |

## 修正方針

### `backend/migrations/001_init.sql` にインデックス追加
```sql
-- マスタ一覧取得最適化 (clinic_id, sort_order) 複合インデックス
CREATE INDEX idx_vaccines_clinic_sort          ON vaccines(clinic_id, sort_order)          WHERE deleted_at IS NULL;
CREATE INDEX idx_medicines_clinic_sort         ON medicines(clinic_id, sort_order)         WHERE deleted_at IS NULL;
CREATE INDEX idx_exam_types_clinic_sort        ON exam_types(clinic_id, sort_order)        WHERE deleted_at IS NULL;
CREATE INDEX idx_procedures_clinic_sort        ON procedures(clinic_id, sort_order)        WHERE deleted_at IS NULL;
CREATE INDEX idx_cages_clinic_sort             ON cages(clinic_id, sort_order)             WHERE deleted_at IS NULL;
CREATE INDEX idx_checkup_types_clinic_sort     ON checkup_types(clinic_id, sort_order)     WHERE deleted_at IS NULL;
CREATE INDEX idx_chief_complaints_clinic_sort  ON chief_complaint_types(clinic_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_types_clinic_sort   ON diagnosis_types(clinic_id, sort_order)   WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_names_clinic_sort   ON diagnosis_names(clinic_id, sort_order)   WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_courses_clinic_sort  ON trimming_courses(clinic_id, sort_order)  WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_options_clinic_sort  ON trimming_options(clinic_id, sort_order)  WHERE deleted_at IS NULL;
CREATE INDEX idx_insurances_clinic_sort        ON insurances(clinic_id, sort_order)        WHERE deleted_at IS NULL;
CREATE INDEX idx_occupations_clinic_sort       ON occupations(clinic_id, sort_order)       WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_types_clinic_sort ON reservation_types(clinic_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_type_groups_sort  ON reservation_type_groups(clinic_id, sort_order) WHERE deleted_at IS NULL;
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md`
> `WHERE clinic_id = X ORDER BY created_at DESC LIMIT 10` のクエリには `(clinic_id, created_at DESC)` インデックスが必要。同様に sort_order でのソートには `(clinic_id, sort_order)` が必要。

## 優先度
**Medium** — 現在のデータ量では問題が顕在化しにくいが、各クリニックのマスタデータが数百〜数千件になると一覧取得レスポンスが劣化する。運用初期に入れておくべき最適化。

## 関連チケット
なし

## 関連ファイル
- `backend/migrations/001_init.sql` — 修正対象（インデックス追加）
