# [master] hook 命名規則違反: `useList*` → `useGet*` に統一せよ

## 優先度
中

## 種別
命名規則 / コード品質

## ステータス
status: closed
closed_at: 2026-03-16

## 対象ファイル
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`
- `frontend/src/features/master/routes/TrimmingSettings.tsx`
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`
- `frontend/src/features/master/routes/StaffSettings.tsx`

## 問題

プロジェクト規約では API query hook の命名は `useGet` + エンティティ（例: `useGetOwners`）と定義されているが、
マスタ関連の hook が `useList*` という動詞を使っている。

| ファイル | 現状 | 修正後 |
|---------|------|--------|
| DiagnosisSettings | `useListDiagnosisCategories` | `useGetDiagnosisCategories` |
| DiagnosisSettings | `useListDiagnosisNames` | `useGetDiagnosisNames` |
| TrimmingSettings | `useListTrimmingCourses` | `useGetTrimmingCourses` |
| TrimmingSettings | `useListTrimmingOptions` | `useGetTrimmingOptions` |
| ServiceTypeSettings | `useListServiceTypes` | `useGetServiceTypes` |
| StaffSettings | `useListStaffs` | `useGetStaffs` |

## 修正方針

該当 API ファイル（`diagnosis.ts`、`trimming.ts`、`service-types.ts`、`staffs.ts`）内の hook 名を変更し、
呼び出し元（routes ファイル）のインポート名も更新する。

**旧名は削除してよい**（alias export で後方互換を保つ必要なし。同一 feature 内でのみ使用）。
