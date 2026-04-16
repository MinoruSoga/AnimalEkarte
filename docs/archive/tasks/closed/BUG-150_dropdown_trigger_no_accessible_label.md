# BUG-150: 操作メニューのドロップダウンボタンにアクセシブルラベルがない

## 概要
飼主一覧・マスタ一覧等のテーブル行にある操作メニュー（`...` アイコン）の
ドロップダウントリガーボタンに `aria-label` がない。
スクリーンリーダーで「ボタン」とだけ読み上げられ、何のボタンか判別できない。

## 脆弱性分類
- **WCAG 2.1 AA 4.1.2**: Name, Role, Value（アクセシビリティ）
- **影響**: スクリーンリーダーユーザーが操作メニューを識別できない

## ブラウザテスト結果
飼主一覧ページで **24個** のラベルなしボタンを検出。
すべて `data-slot="dropdown-menu-trigger"` のアイコンボタン。

## 現状 HTML
```html
<button data-slot="dropdown-menu-trigger" class="...">
  <!-- SVG icon only, no text, no aria-label -->
</button>
```

## 期待する HTML
```html
<button data-slot="dropdown-menu-trigger" aria-label="田中 花子の操作メニュー" class="...">
  <!-- SVG icon -->
</button>
```

または汎用的に:
```html
<button data-slot="dropdown-menu-trigger" aria-label="操作" class="...">
```

## 修正方針

`RowActionDropdown` コンポーネントに `aria-label` を追加:

```typescript
<DropdownMenuTrigger asChild>
  <Button variant="ghost" size="icon" aria-label={`${rowLabel}の操作メニュー`}>
    <MoreHorizontal />
  </Button>
</DropdownMenuTrigger>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/accessibility-rules.md`
> **ARIA 属性（適切に）**: `aria-label` で機能説明
> **キーボード操作対応**: Tab フォーカスで操作可能

## 優先度
**Low** — アクセシビリティ改善。機能的な影響なし。

## 関連ファイル
- `frontend/src/components/shared/RowActionDropdown/` — 操作メニューコンポーネント
- `frontend/src/features/master/components/MasterListPage.tsx` — マスタ一覧の操作カラム
