# TASK-009: 外部キーカラムのインデックス欠落 — vaccinations / exams / checkups

## 概要

`vaccinations.vaccine_id`, `exams.exam_type_id`, `checkups.checkup_type_id` の外部キーカラムにインデックスが存在しない。JOIN やフィルタ時に Seq Scan が発生し、データ増加に伴うパフォーマンス劣化のリスクがある。

## 優先度

MEDIUM（パフォーマンス）

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/migrations/001_init.sql` | vaccinations, exams, checkups テーブルのインデックス定義部 |

## 規約違反

`.claude/rules/database-design.md`:
> 外部キーカラムにはインデックス必須

## 修正方針

```sql
-- 001_init.sql のインデックス定義セクションに追加
CREATE INDEX idx_vaccinations_vaccine_id ON vaccinations(vaccine_id);
CREATE INDEX idx_exams_exam_type_id ON exams(exam_type_id);
CREATE INDEX idx_checkups_checkup_type_id ON checkups(checkup_type_id);
```

追加で、`owners` と `pets` の clinic_id 先頭複合インデックスも欠落している:

```sql
CREATE INDEX idx_owners_clinic_id_pk ON owners(clinic_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pets_clinic_owner ON pets(clinic_id, owner_id) WHERE deleted_at IS NULL;
```

## 注意

リリース前のため `001_init.sql` を直接編集してよい（incremental migration 不要）。

## テスト

```sql
-- 追加後に EXPLAIN ANALYZE で Index Scan を確認
EXPLAIN ANALYZE
  SELECT * FROM vaccinations WHERE vaccine_id = 1;
-- → Index Scan が使われていること
```
