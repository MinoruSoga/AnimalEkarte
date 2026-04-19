# BUG-357: マイグレーションに冗長なインデックスが 6 件存在

## 概要

`001_init.sql` に、複合インデックスまたは UNIQUE インデックスの先頭カラムと同じ単独カラムインデックスが存在する。
PostgreSQL の B-tree は複合インデックスの先頭カラムだけで等値検索をカバーできるため、単独インデックスは冗長。
書き込み時の I/O オーバーヘッドとディスク消費が不要に増加する。

## 冗長インデックス一覧

| 冗長なインデックス | カバーされるインデックス | テーブル |
|---|---|---|
| `idx_billings_medical_record_id(medical_record_id)` | `idx_billings_medical_record_id_unique(medical_record_id) WHERE NOT NULL` | billings |
| `idx_permission_groups_clinic(clinic_id)` | `uk_permission_groups(clinic_id, name)` | permission_groups |
| `idx_medical_records_clinic_id(clinic_id)` | `idx_medical_records_clinic_record_no(clinic_id, record_no)` | medical_records |
| `idx_estimates_clinic_id(clinic_id)` | `idx_estimates_clinic_estimate_no(clinic_id, estimate_no)` | estimates |
| `idx_daily_records_hospitalization_id(hospitalization_id)` | `idx_daily_records_hosp_date(hospitalization_id, date)` | daily_records |
| `idx_shift_templates_clinic(clinic_id)` | `uk_shift_templates_clinic_name(clinic_id, name)` | shift_templates |

## 根拠

PostgreSQL の B-tree インデックスは左端プレフィックスマッチをサポートする。
`CREATE INDEX ON t(A, B)` は `WHERE A = ?` の等値検索に使用可能。
このプロジェクトでは `clinic_id` は常に等値フィルタ（テナント分離）のため、
`(clinic_id)` 単独インデックスは `(clinic_id, xxx)` 複合インデックスで完全にカバーされる。

## 修正内容

`001_init.sql` から以下の 6 行を削除:

```sql
-- 削除対象
CREATE INDEX idx_billings_medical_record_id ON billings(medical_record_id);
CREATE INDEX idx_permission_groups_clinic ON permission_groups(clinic_id);
CREATE INDEX idx_medical_records_clinic_id ON medical_records(clinic_id);
CREATE INDEX idx_estimates_clinic_id ON estimates(clinic_id);
CREATE INDEX idx_daily_records_hospitalization_id ON daily_records(hospitalization_id);
CREATE INDEX idx_shift_templates_clinic ON shift_templates(clinic_id);
```

## 優先度

**LOW** — 機能影響なし。書き込みパフォーマンスとストレージの微小な改善のみ。
