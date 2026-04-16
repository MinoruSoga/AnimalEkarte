# BUG-163: MedicalRecordForm — ペットデータローディング中に空白ページが表示される

## 概要

`frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` で、ペットデータ取得中および
ペット未選択時に `return null` を返している。ユーザーには空白ページが表示されるため、
ローディング中なのかエラーなのかを判断できない。

## 再現手順

1. 診療カルテ新規作成フォームを開く（`/medical-records/new` または対応 URL）
2. ネットワーク速度を低速にシミュレート（DevTools → Network → Slow 3G）
3. ページをリロード
4. **結果**: ペットデータ取得完了まで画面が完全に白紙になる

## 期待する動作

- ローディング中は `<LoadingFallback />` を表示する
- ペット未選択時はエラーまたはガイダンスメッセージを表示する（空白ではなく）

## 現状コード

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:257付近`
```tsx
// Before: return null でサイレント空白
if (isPetLoading) {
  return null;  // ❌ ユーザーに何も表示されない
}
if (!selectedPet) {
  return null;  // ❌ エラーか正常かわからない
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:196-197
// ✅ LoadingFallback / ErrorFallback を使用
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `MedicalRecordForm.tsx` | isPetLoading 中の null return | 未修正 |
| `MedicalRecordForm.tsx` | selectedPet === null 時の null return | 未修正 |

## 修正方針

### 1. ローディング中のフォールバック表示 — `MedicalRecordForm.tsx:257`
```tsx
import { LoadingFallback } from "@/components/shared/DataStates/LoadingFallback";
import { ErrorFallback } from "@/components/shared/DataStates/ErrorFallback";

// After
if (isPetLoading) return <LoadingFallback />;
if (!selectedPet) return <ErrorFallback message="ペット情報が見つかりません" />;
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
プロジェクト内参照実装 `features/medical-records/routes/MedicalRecords.tsx` では
`LoadingFallback` / `ErrorFallback` を正しく使用している。同 feature 内で実装が不統一。

### プロジェクト内参照実装
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx:196-197` — 正しいパターン

## 優先度
**Medium** — 機能的障害ではないが、ネットワーク遅延環境でユーザーが混乱する。

## 関連チケット
- FE-247: 受付カンバンの初期ローディングスケルトン欠如（同種問題）

## 関連ファイル
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:257-263`
- `frontend/src/components/shared/DataStates/LoadingFallback.tsx`
- `frontend/src/components/shared/DataStates/ErrorFallback.tsx`
