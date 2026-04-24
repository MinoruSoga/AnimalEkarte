# FE-206: 入院管理 デイリーレコード系コンポーネントでスタッフ名をハードコード

## 概要

入院管理のデイリーレコード関連コンポーネント4箇所で、記録時のスタッフ名フィールドが
`"スタッフ"` や `"担当医"` にハードコードされている。
本番では送信時のスタッフ名は認証済みユーザー（`useMe()` hook）から取得すべきである。
ハードコード値のまま保存されると、誰が記録したか特定できなくなる。

## 現状コード

### `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:37`
```ts
onSave({
  time: format(new Date(), "HH:mm"),
  type: "other",
  status: "completed",
  value: "経過記録",
  notes: note,
  staff: "スタッフ",  // ← ハードコード
});
```

### `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:52`
```ts
setForm(prev => ({ 
  ...prev, 
  staff: "スタッフ"   // ← ハードコード
}));
```

### `frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx:58`
```ts
setForm(prev => ({ 
  ...prev, 
  staff: "担当医"    // ← ハードコード
}));
```

### `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx:50`
```ts
setForm(prev => ({ 
  ...prev, 
  staff: "担当医"    // ← ハードコード
}));
```

**注**: `DailyCareNoteForm.tsx` は FE-199 でも対応対象に含まれているが、ここでは他3箇所も含めて一括対応を推奨。

## 期待する動作

記録送信時に `useMe()` で取得したログインユーザーの名前がスタッフ名として設定される。

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx` | 37 | 要修正（FE-199 と重複） |
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx` | 52 | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx` | 58 | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx` | 50 | 要修正 |

## 修正方針

各コンポーネントで `useMe()` hook を使ってログインユーザー名を取得する。

```tsx
// 各コンポーネントに追加
import { useMe } from "@/features/auth";

export function DailyCareLogDialog({ ... }) {
  const { data: me } = useMe();

  // form 初期化時
  setForm(prev => ({ 
    ...prev, 
    staff: me?.name ?? ""   // ログインユーザー名
  }));
```

**注意**: `useMe()` の `data` は非同期なので、ダイアログが開いた時点で `me` が undefined の場合を考慮して `??""` でフォールバックを設定すること。

## 参照実装

`frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:190` では
すでに `me?.displayName` を使って正しくログインユーザー名を設定している参照実装がある。

```tsx
// HospitalizationForm.tsx:190 — 正しい実装
doctorName: me?.displayName ?? "",
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — コーディング姿勢
> ハードコードされた値の使用禁止（use env vars or constants）

## 優先度
**High** — 本番環境で全ての入院デイリーレコードのスタッフ名が「スタッフ」または「担当医」になる。
監査証跡として機能せず、データ品質が損なわれる。

## 関連チケット
- FE-199: DailyCareNoteForm の複合違反（`DailyCareNoteForm.tsx:37` はここでも対応対象）

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:52`
- `frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx:58`
- `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx:50`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:37`
- `frontend/src/features/auth/index.ts` — useMe hook
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:190` — 参照実装
