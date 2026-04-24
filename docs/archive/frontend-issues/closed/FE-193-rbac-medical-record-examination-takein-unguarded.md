# FE-193: カルテ検査タブ — ExaminationFilter「検査取り込み」ボタンに RBAC なし（FE-156 修正漏れ）

## 概要

`/medical-records/:id` の **検査タブ**（`MedicalRecordExamination` / `ExaminationFilter`）で
「検査取り込み」ボタンに `canCreate` ガードがなく、
`canCreate=false` ユーザーでもボタンが表示・操作可能になっている。

FE-156 では `MedicalRecordExamination.tsx` は「GET のみのため問題なし」と記載されていたが、
「検査取り込み」ボタン（POST 操作）が `ExaminationFilter.tsx` に含まれており、
これが未確認・未修正だった。

## 再現手順

1. `canCreate=false` ユーザーでログイン
2. `/medical-records/22` 等カルテ詳細へ移動
3. 「検査」タブを開く

## 確認された問題

| 問題 | 状態 |
|------|------|
| 「検査取り込み」ボタン表示・操作可能 | ❌ `disabled: false` |

スクリーンショット確認済み：右上に青い「検査取り込み」ボタンが表示されている。

## 根本原因

```tsx
// ExaminationFilter.tsx — canCreate/usePermission なし
// (grep: canEdit/canCreate/usePermission/disabled → No matches found)

<Button onClick={onImport}>
  検査取り込み  {/* canCreate=false でも常に表示 ❌ */}
</Button>
```

## 修正方針

```tsx
// ExaminationFilter.tsx
const { canCreate } = usePermission("medical-records");

// または props で受け取る
interface ExaminationFilterProps {
  // ...
  canCreate?: boolean;
}

// ボタンを条件付きレンダー
{canCreate ? (
  <Button onClick={onImport}>
    <FileText />
    検査取り込み
  </Button>
) : null}
```

## 優先度

**MEDIUM** — 閲覧のみユーザーが検査データ取り込みを試みることができる。
API 側で 403 は返るが、UI 上は操作可能な状態。

## 関連ファイル

- `frontend/src/features/medical-records/components/ExaminationFilter.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx`
- 発見日: 2026-04-08
- 関連: FE-156（本チケットの対象外として誤認・修正漏れ）
