# [master] テーブル行の高さを +20px にする（全マスタページ）

## 優先度
低

## 種別
UIスタイル変更

## 対象ファイル
- `frontend/src/lib/design-tokens.ts`（`STYLE.tableRow`）

---

## 変更内容

マスタページのテーブル行高さを現在の `h-12`（48px）から `h-16`（64px）に変更する。

### 修正箇所

```diff
// frontend/src/lib/design-tokens.ts
  tableRow:
-   `border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors cursor-pointer h-12`,
+   `border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors cursor-pointer h-16`,
```

---

## 影響範囲

`STYLE.tableRow` は `TABLE_STYLES.row` 経由で `DataTableRow` に適用されるため、
**design-tokens.ts の1行変更で全マスタページに一括適用**される。

| コンポーネント | 伝播経路 |
|---------------|---------|
| `DataTableRow` | `TABLE_STYLES.row` = `STYLE.tableRow` |
| `SortableDataTableRow` | `DataTableRow` 経由 |
| 全マスタページ | `DataTableRow` / `SortableDataTableRow` 使用 |

---

## 確認事項

- `tableCell` の `py-2.5`（10px top/bottom）は維持したまま、行高さ増加分は縦方向の余白として吸収されることを確認する
- `tableHeaderRow`（`h-11` = 44px）は変更対象外（ヘッダ行はそのまま）
