# BUG-218: PetSelectionResultsTable の死亡ペット表示に text-gray-400 ハードコード

## 概要

`PetSelectionResultsTable.tsx:82` で死亡ペットの選択ボタンに `text-gray-400` をハードコードしている。デザイントークン規約違反。

## 再現手順

1. ペット選択UI（`ReservationFormModal` 等）を開く
2. 死亡ステータスのペットが表示される行を確認
3. 選択ボタンが `text-gray-400` で薄く表示される
4. **結果**: `text-gray-400`（Tailwind ハードコード）が使われる

## 期待する動作

- 無効化状態のテキスト色はデザイントークン（`C.textStatusGray` または `C.text40`）を使用すること

## 現状コード

### `frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx:82`
```tsx
className={`h-11 gap-1 ${isDeceased
  ? "text-gray-400"
  : `${C.bgAccent} ${C.bgAccentHover} text-white`
} text-sm px-4`}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// features/medical-records 等で使われる muted テキストの正しいパターン
<span className={C.text40}>...</span>
// または
<span className={C.textStatusGray}>...</span>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `components/shared/PetSelection/PetSelectionResultsTable.tsx:82` | text-gray-400 | 未修正 |

## 修正方針

### `frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx:82`

```tsx
// Before
? "text-gray-400"

// After
? C.textStatusGray
```

`C.textStatusGray` = `"text-[#9B9A97]"` — 無効化・グレーアウト用の標準トークン

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/typescript-react.md` §4 デザイントークン
> スタイリングには必ず `src/lib/design-tokens.ts` の定数を使用する。Hexカラーの直接指定は厳禁。

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用（ハードコード禁止）

### プロジェクト内参照実装
- `frontend/src/lib/design-tokens.ts:334` — `textStatusGray: "text-[#9B9A97]"` 存在確認済み

## 優先度
**Low** — 機能的問題なし。視覚的差異も軽微。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx:82`
- `frontend/src/lib/design-tokens.ts:334`
