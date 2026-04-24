# FE-233: SidePeek・HistoryFilterPanel のアクセシビリティ違反

## 概要

`frontend/src/components/shared/SidePeek/` および `HistoryFilterPanel/` 配下の
複数コンポーネントでアクセシビリティ規約違反が存在する。
アイコンのみのボタンに `aria-label` がなく、input と label の関連付けが欠けている。

## 違反箇所

### `SidePeek/SidePeekToolbar.tsx`

| 行 | 問題 |
|----|------|
| 21-27 | Trash2 アイコンのみのボタンに `aria-label` なし → スクリーンリーダーで機能不明 |

```tsx
// Before
<button onClick={onDelete}><Trash2 /></button>

// After
<button aria-label="削除" onClick={onDelete}><Trash2 /></button>
```

### `SidePeek/StatusToggleButton.tsx`

| 行 | 問題 |
|----|------|
| 14-20 | アイコンのみのボタンに `aria-label` なし |

```tsx
// After
<button aria-label="ステータスを切り替え" ...>
```

### `SidePeek/PropertyInput.tsx`

| 行 | 問題 |
|----|------|
| 20-26 | `<input>` に `id` なし → `<label>` と関連付けできない |

### `SidePeek/MoneyInput.tsx`

| 行 | 問題 |
|----|------|
| 17-24 | `<input>` に `id` なし → `PropertyRow` の視覚ラベルとの関連付けが欠落 |

### `HistoryFilterPanel/HistoryFilterPanel.tsx`

| 行 | 問題 |
|----|------|
| 69-74 | `<input>` に `id` なし、`<Label>` に `htmlFor` なし |
| 82-95 | ソート用 `<Select>` に `aria-label` または関連 `<Label>` なし |

## 修正方針

```tsx
// HistoryFilterPanel — Before
<Label>検索</Label>
<input type="text" ... />

// After
<Label htmlFor="history-search">検索</Label>
<input id="history-search" type="text" ... />
```

## 準拠すべきプロジェクト規約

### `.claude/rules/accessibility-rules.md`
> フォームラベル `htmlFor` で input と関連付けること（必須）
> アイコンのみボタンは `aria-label` で機能説明すること

### FE-205（既存関連チケット）
本チケットは FE-205 でカバーされていない shared/SidePeek コンポーネントを対象とする。

## 優先度
**Medium** — スクリーンリーダー利用者がアイコンボタンの目的を把握できない。
`SidePeekToolbar` の削除ボタンは誤操作リスクもある。

## 関連ファイル
- `frontend/src/components/shared/SidePeek/SidePeekToolbar.tsx`
- `frontend/src/components/shared/SidePeek/StatusToggleButton.tsx`
- `frontend/src/components/shared/SidePeek/PropertyInput.tsx`
- `frontend/src/components/shared/SidePeek/MoneyInput.tsx`
- `frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx`
