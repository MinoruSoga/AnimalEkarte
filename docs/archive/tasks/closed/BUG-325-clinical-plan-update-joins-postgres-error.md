# BUG-325: 診察所見・治療方針の保存が常に400エラーになる

**Status**: FIXED  
**Priority**: High  
**Discovery**: 機能テスト Section 4.3 (2026-04-12)

## 概要

カルテ編集の「診察/治療プラン」タブで「保存」ボタンをクリックすると、`PATCH /v1/medical-records/:id/clinical-plan` が常に HTTP 400 を返し「入力値が正しくありません」トーストが表示される。

身体検査所見・診断詳細・治療方針のいずれかに文字列（空文字含む）を含むリクエストで再現する。

## 再現手順

1. `/medical-records/21` を開き「診察/治療プラン」タブに切り替える
2. 身体検査所見など任意フィールドを入力（空欄でも発生）
3. 「保存」ボタンをクリック
4. **結果**: "入力値が正しくありません" トースト / 400 Bad Request
5. **期待**: "保存しました" トースト / 200 OK

## 根本原因

`backend/internal/repository/clinical_plan_repository.go:52-57` で GORM の `Joins()` と `Updates(map)` を組み合わせていた。

GORM は `Joins` + `Updates` 時、PostgreSQL 向けに `FROM` 句を生成せず `WHERE` 句だけにテーブル名を残す不正な SQL を生成する:

```sql
-- 生成された不正 SQL
UPDATE "clinical_plans" SET "physical_exam"='test', "updated_at"='...'
WHERE clinical_plans.id = 21 AND medical_records.clinic_id = 3
-- → ERROR: missing FROM-clause entry for table "medical_records" (SQLSTATE 42P01)
```

`null` を送ると fields マップが空になり service が early return するため、GORM クエリ自体が実行されず 200 になっていた（バグの隠蔽）。

## 修正箇所

`backend/internal/repository/clinical_plan_repository.go:52-57`

```go
// 修正前（バグあり）
result := r.db.WithContext(ctx).
    Model(&model.ClinicalPlan{}).
    Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id").
    Where("clinical_plans.id = ? AND medical_records.clinic_id = ?", id, clinicID).
    Updates(fields)

// 修正後（サブクエリで tenant isolation を維持）
result := r.db.WithContext(ctx).
    Model(&model.ClinicalPlan{}).
    Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
    Updates(fields)
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/clinical_plan_repository.go:52` | Update メソッド | ✅ 修正済み |
| `PATCH /v1/medical-records/:id/clinical-plan` | 診察所見・診断・治療方針の更新 | ✅ 動作確認済み |

## 準拠すべきプロジェクト規約

### `.claude/rules/database-design.md` — マルチテナント設計

> WHERE 句は clinic_id から開始。外部テーブルの JOIN が必要な場合はサブクエリを使用

GORM の `Joins()` は SELECT 専用。UPDATE で別テーブルの clinic_id を検証する場合はサブクエリが必要。

### `.claude/rules/go-language.md` — GORM PATCH

> repository/owner_repository.go で `UpdateFields` に `map[string]any` を使用するパターンを参照

## テスト確認事項

- [x] 身体検査所見を入力して保存 → 200 / トーストなし（成功）
- [x] 空欄で保存 → 200 / トーストなし（成功）
- [x] `{"physical_exam":"test"}` 単体 PATCH → 200

## 優先度

**High** — カルテの診察所見・治療方針が一切保存できない機能不全バグ
