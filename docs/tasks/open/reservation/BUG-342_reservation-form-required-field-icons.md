# BUG-342: 予約登録フォームの必須項目に必須アイコンを表示する

## 概要
予約登録フォームには患者・日付・予約区分・時間など必須フィールドがあるが、フィールドラベルに必須を示すアイコン（`*` や `必須` バッジ）が一切表示されていない。ユーザーは送信して初めてエラーに気づく設計になっており、UX として不完全である。

## 再現手順
1. `admin@example.com` / `password` でログイン
2. 予約カレンダー → 「予約登録」ボタンをクリック
3. フォーム内の各フィールドラベルを確認
4. **結果**: 「患者」「日付」「予約区分」「開始時刻」「終了時刻」のラベルに必須表示がない

## 期待する動作
- 必須フィールドのラベル横に視覚的な必須インジケータ（例: 赤い `*` やバッジ）を表示する
- 任意項目には表示しない（担当者・メモ等）

## 現状コード

### `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:136-172`
```tsx
// 必須チェックは実装済みだがラベルに表示なし
if (selectedPets.length === 0) errors.patient = "患者を選択してください";
if (!formData.start)           errors.date   = "日付を選択してください";
if (!formData.type)            errors.type   = "予約区分を選択してください";
if (formData.end <= formData.start) errors.time = "終了時刻は開始時刻より後に設定してください";
```

### `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:42-50`
```tsx
// FieldLabel コンポーネント: trailing オプションはあるが required prop なし
function FieldLabel({ label, trailing }: { label: string; trailing?: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between mb-1">
      <label className="text-sm font-medium">{label}</label>
      {trailing ? trailing : null}
    </div>
  );
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// owners/routes 等で使われているパターン
<label className="text-sm font-medium">
  患者
  <span style={{ color: C.DANGER }} className="ml-1">*</span>
</label>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:42-50` | `FieldLabel` に `required?: boolean` prop を追加 | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:96-114` | 日付フィールドの FieldLabel | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:191-205` | 予約区分フィールドの FieldLabel | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:264-287` | 担当者フィールド（任意 → 表示不要） | 確認不要 |

## 修正方針

### 1. `FieldLabel` に `required` prop を追加
`frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:42-50`
```tsx
import { C } from '@/lib/design-tokens';

function FieldLabel({ label, required, trailing }: {
  label: string;
  required?: boolean;
  trailing?: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between mb-1">
      <label className="text-sm font-medium">
        {label}
        {required ? (
          <span style={{ color: C.DANGER }} className="ml-1" aria-hidden="true">*</span>
        ) : null}
      </label>
      {trailing ? trailing : null}
    </div>
  );
}
```

### 2. 必須フィールドの FieldLabel に `required` を付与
患者・日付・予約区分・開始時刻・終了時刻の FieldLabel 呼び出し箇所で `required` を追加する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Conditional Render
> 必ず `? (...) : null`（`&&` 禁止）

`required` prop の条件レンダリングは `{required ? (...) : null}` で実装する。

### `.claude/rules/accessibility-rules.md` — フォームラベル（必須）
> セマンティック HTML でフォームラベルと input を関連付け

`aria-hidden="true"` を付与し、スクリーンリーダー向けには別途 `aria-required="true"` を input 側に設定する。

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `C`, `STYLE` 定数を使用。Hexカラー禁止。

`*` の色には `C.DANGER` を使用する（直接 `red` や `#ef4444` 禁止）。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx` — 必須フィールドのラベル表示パターン（参照）

## 優先度
**Medium** — ユーザーが「どれが必須か」を事前に把握できない UX 上の問題。実害は軽微だが修正コストも低い。

## 関連チケット
- BUG-341: 予約区分コンボボックス変更（同一フォーム対象）

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:42-50` — FieldLabel コンポーネント
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:136-172` — バリデーションロジック（必須フィールド定義の参照元）
- `frontend/src/lib/design-tokens.ts` — `C.DANGER` カラー定数
