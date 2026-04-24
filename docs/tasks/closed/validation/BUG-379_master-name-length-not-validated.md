# BUG-379: マスタ名称の文字数制限が未実装 (Backend/Frontend 両方)

**作成日**: 2026-04-15
**Status**: OPEN
**Priority**: MEDIUM (データ健全性・UI 破綻リスク)
**Affects**: 全マスタ (animal-species, treatment-items, medicines, cages, reservation-types, 等)

## 概要

`/api/v1/masters/animal-species` へ `name="あ" × 300文字` で POST → **201 Created**、DB に 300 文字保存成功。他マスタも同様と推定 (未全件検証)。

## 再現手順

```bash
curl -X POST http://localhost:8080/api/v1/masters/animal-species \
  -H 'x-clinic-id: 3' -b cookies.txt \
  -d '{"name":"'$(printf 'あ%.0s' {1..300})'"}'
# 201 Created
```

または UI 経由: `/settings/animal-species` → 新規登録 → 名称に 300 文字 → 保存 → 201 + 一覧に 300 文字の行が表示され UI 崩れ。

## 期待動作

255 文字超で **400 Bad Request** + メッセージ「名称は 255 文字以内で入力してください」。Frontend は `maxLength=255` で入力段階から制限。

## 修正方針

### Backend

1. **DB スキーマ確認**: `\d animal_species` で name カラムが `text` 型か `varchar(N)` かを確認。text なら VARCHAR(255) or CHECK 制約追加 (migration)
2. **Service 層バリデータ**: `backend/internal/service/validators.go` に `validateMasterName(name string) error` を追加し、各 master service の Create/Update で呼び出す
   - 空文字チェック (既存か要確認)
   - 255 文字以内 (UTF-8 rune count)
   - trim 後の長さで判定

### Frontend

1. 全マスタのサイドパネル入力に `maxLength={255}` 属性を追加
   - 共通化: `SidePeekTitleInput` コンポーネントにデフォルト `maxLength={255}` を設定
2. 400 エラー時は `handle-api-error.ts` で既に toast 表示されているため追加実装不要

## 影響範囲

- 全マスタテーブル (animal_species, breeds, consultations, procedures, medicines, cages, hospitalization_plans, trimming_courses, trimming_options, insurances, occupations, reservation_types, reservation_type_groups, diagnosis_categories, diagnosis_names, chief_complaint_categories, chief_complaint_names, exam_types, checkup_types, job_titles, permission_groups, merchandise_items, staffs, inquiry_templates 等)

## 確認事項

- [ ] 既に 256+ 文字のレコードが DB に存在するか (SELECT count(*) WHERE length(name) > 255)
- [ ] 他のマスタ (medicines/insurances 等) の name カラム型を grep で一覧化
- [ ] Unicode 絵文字・サロゲートペアでの文字数カウント挙動を統一 (rune vs byte)
