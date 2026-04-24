# FE-199: DailyCareNoteForm の複合違反（undefined variant・デザイントークン・スタッフ名ハードコード）

## 概要

入院管理の「経過記録」フォーム（`DailyCareNoteForm`）に以下の3つの問題が存在する:
1. `variant="primary"` — shadcn Button に存在しないバリアントを指定しておりスタイルが未適用
2. `border-gray-100`, `text-gray-500`, `text-gray-400` — デザイントークンでなく Tailwind 直接色指定
3. `staff: "スタッフ"` — 記録送信時のスタッフ名がハードコードされており、実際のログインユーザーが設定されない

## 再現手順

1. 入院管理 → 任意の入院患者を開く
2. 「デイリーレコード」タブ → 経過記録フォームに文字を入力
3. 「記録」ボタンを確認 → **黒色（デフォルト）で表示される**（青の primary スタイル未適用）
4. 記録を送信 → サーバーに送られるスタッフ名が **"スタッフ"** になる

## 期待する動作

- 「記録」ボタンは青（`C.bgAccent`）で表示される
- 色は設計トークンで管理される
- 送信時のスタッフ名はログインユーザーの名前が自動設定される

## 現状コード

### `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx`

**問題1: undefined variant（行57）**
```tsx
<Button
  size="sm"
  onClick={handleSubmit}
  disabled={!note.trim()}
  variant="primary"   // ← shadcn Button に "primary" バリアントは存在しない
  className={`gap-2 shadow-sm ${H_STYLES.button.action}`}
>
  <Send className={H_STYLES.button.icon} />
  記録
</Button>
```

shadcn Button の定義済みバリアント: `default`, `destructive`, `outline`, `secondary`, `ghost`, `link` のみ。
`variant="primary"` は無視され、スタイルが適用されない（デフォルト外観になる）。

**問題2: デザイントークン未使用（行49, 51, 52）**
```tsx
className={`... placeholder:text-gray-400 ...`}  // 行49
<div className="flex justify-between items-center border-t border-gray-100 pt-1.5 mt-0.5">  // 行51
<span className={`... font-medium text-gray-500`}>  // 行52
```

**問題3: スタッフ名ハードコード（行37）**
```tsx
onSave({
  time: format(new Date(), "HH:mm"),
  type: "other",
  status: "completed",
  value: "経過記録",
  notes: note,
  staff: "スタッフ"   // ← ハードコード。ログインユーザー名を使うべき
});
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:57` | undefined variant | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:49,51,52` | Tailwind 直接色指定 | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx:37` | スタッフ名ハードコード | 要修正 |

## 修正方針

### 1. variant="primary" の削除 + 適切なスタイル付与（行57）
```tsx
// Before
variant="primary"
className={`gap-2 shadow-sm ${H_STYLES.button.action}`}

// After
className={`gap-2 shadow-sm ${H_STYLES.button.action} ${C.bgAccent} ${C.bgAccentHover} text-white`}
```

### 2. デザイントークンへの置換（行49, 51, 52）
```tsx
// Before: placeholder:text-gray-400
// After:  placeholder:${C.text40} または placeholder:text-[color:var(--text-40)]

// Before: border-gray-100
// After:  ${C.borderLight}

// Before: text-gray-500
// After:  ${C.text60}
```

### 3. スタッフ名をログインユーザーから取得（行28-41）
```tsx
// useMe() hookからログインユーザー名を取得
import { useMe } from "@/features/auth";

export function DailyCareNoteForm({ onSave }: DailyCareNoteFormProps) {
  const { data: me } = useMe();
  // ...
  onSave({
    // ...
    staff: me?.name ?? "",   // ログインユーザー名
  });
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> Tailwind の `gray-*` 直接指定は禁止。

### プロジェクト内参照実装
- `frontend/src/features/hospitalization/styles.ts` — `H_STYLES` は `C` を参照しているが、直接 Tailwind を混在させている点に注意

## 優先度
**Medium** — 「記録」ボタンのスタイル崩れ + スタッフ名が "スタッフ" 固定でデータ品質が低下。

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareNoteForm.tsx` — 要修正
- `frontend/src/lib/design-tokens.ts` — 置換先トークン定義
- `frontend/src/features/auth/index.ts` — useMe hook
