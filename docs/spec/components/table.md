# Table

> 業務 data-table のプリミティブ。配置: `frontend/src/components/ui/table.tsx`
> Storybook: `table.stories.tsx` / 関連監査: C18（cell override 禁止）・C19（row onClick 禁止）

## Anatomy

- `Table` → `TableHeader` → `TableRow` → `TableHead`
- `Table` → `TableBody` → `TableRow` → `TableCell`
- `TableFooter`

## 状態・規則

| 規則 | 内容 |
|------|------|
| cell typography | TableHead/TableCell が正本。呼び出し側で `text-*`/`font-*`/`px-*` を上書きしない（C18） |
| row hover | 行側の `STYLE.tableRow` / `PALETTE.tableRowHover` が正本 |
| row onClick | **禁止**（C19）— 行内の明示的アクション要素を使う |

## アクセシビリティ要件

- ヘッダセルは `TableHead`（th 相当）を必ず使う — scope=col の意味論。
- ソート可能列は `SortableDataTableRow` 等の既存ユーティリティ経路を使う（C19 の対象にもなる）。
- 空状態・読み込み状態は TableBody 内の1行 colSpan メッセージか、テーブル外に出す。

## 使い方（Do / Don't）

- Do: dense な業務一覧には Table primitive を使い、見た目の調整は STYLE/BADGE トークンで。
- Don't: セルに raw `bg-*`・`text-*`・padding 任意値を書かない（C18 で CI fail）。
- Don't: 行全体をクリック可能にしない（誤操作・キーボード操作不可のため C19 禁止）。

## 既知のギャップ

- C18 allowlist に print 帳票・owner-report 密度テーブル等の例外が残存（`C18_RAW_CELL_ALLOWLIST` 参照）。
- sticky header / 横スクロールの統一パターンは feature 側に散在 — 要標準化。
