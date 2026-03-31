# [master] テーブルのフォントサイズを text-base に、font-family を Noto Sans JP のみに変更（全マスタページ）

## 優先度
低

## 種別
UIスタイル変更

## 対象ファイル
- `frontend/src/lib/design-tokens.ts`（tableCell 系）
- `frontend/src/styles/globals.css`（font-family）

---

## 変更内容

### 1. フォントサイズ: `text-sm` → `text-base`

テーブルセルおよびヘッダーセルのフォントサイズを `text-sm`（15px）から `text-base`（16px）に変更する。

```diff
// frontend/src/lib/design-tokens.ts
  tableHeaderCell:
-   `text-xs font-medium ${C.text70} h-11`,
+   `text-xs font-medium ${C.text70} h-11`,   // ヘッダーは text-xs のまま（変更なし）

  tableCell:
-   `text-sm ${C.text} py-2.5`,
+   `text-base ${C.text} py-2.5`,

  tableCellMono:
-   `font-mono text-sm ${C.text} py-2.5`,
+   `font-mono text-base ${C.text} py-2.5`,

  tableCellMuted:
-   `text-sm ${C.text70} py-2.5`,
+   `text-base ${C.text70} py-2.5`,
```

> `tableHeaderCell` は `text-xs`（13px）のままとする。

---

### 2. font-family: Inter を除去し Noto Sans JP のみに

現状の `'Inter', 'Noto Sans JP', sans-serif` から Inter を取り除き、Noto Sans JP を単独指定にする。
Noto Sans JP はラテン文字・日本語文字を両方カバーするため、Inter は不要。

```diff
// frontend/src/styles/globals.css

 @theme inline {
-  --font-sans: 'Inter', 'Noto Sans JP', sans-serif;
+  --font-sans: 'Noto Sans JP', sans-serif;
   ...
 }

 @layer base {
   html {
-    font-family: 'Inter', 'Noto Sans JP', sans-serif;
+    font-family: 'Noto Sans JP', sans-serif;
   }

   body {
-    font-family: 'Inter', 'Noto Sans JP', sans-serif;
+    font-family: 'Noto Sans JP', sans-serif;
   }
 }

 /* Override Tailwind preflight font-family */
 html, body {
-  font-family: 'Inter', 'Noto Sans JP', sans-serif !important;
+  font-family: 'Noto Sans JP', sans-serif !important;
 }
```

---

## 影響範囲

| 変更 | 影響箇所 |
|------|---------|
| `text-base` | `DataTableRow` / `SortableDataTableRow` 経由で全マスタページのテーブルセルに適用 |
| font-family | globals.css はアプリ全体に適用されるため、テーブルだけでなく全ページに反映される |

> font-family の変更はアプリ全体への影響となるが、Noto Sans JP はラテン文字も含むため視覚的な差異は軽微。
