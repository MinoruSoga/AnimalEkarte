# BUG-309: MasterLink — CATEGORY_PATH_MAP ハードコード16箇所 + paths.ts 欠落エントリ

## 概要

`frontend/src/components/shared/MasterLink.tsx:24-41` の `CATEGORY_PATH_MAP` 定数に16のURL文字列がハードコードされていた。さらに `config/paths.ts` に8エントリが欠落していた。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/components/shared/MasterLink.tsx` | line 24-41 |
| `frontend/src/config/paths.ts` | vaccine/examination/trimmingCourse/trimmingOption/consultation/procedure/diagnosisCategory/diagnosisName が欠落 |

## 修正内容

### paths.ts に追加
- `settings.vaccine`
- `settings.examination`
- `settings.trimmingCourse`
- `settings.trimmingOption`
- `settings.consultation`
- `settings.procedure`
- `settings.diagnosisCategory`
- `settings.diagnosisName`

### MasterLink.tsx
`CATEGORY_PATH_MAP` の全16エントリを `paths.settings.*getHref()` 経由に変更。

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止

## ステータス

✅ 修正済み
