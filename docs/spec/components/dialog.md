# Dialog

> Radix Dialog ベースのモーダル。配置: `frontend/src/components/ui/dialog.tsx`
> Storybook: `dialog.stories.tsx`

## Anatomy

- `Dialog`（Root）/ `DialogTrigger` / `DialogPortal`
- `DialogOverlay`（backdrop）+ `DialogContent`（panel）
- `DialogHeader` / `DialogTitle` / `DialogDescription` / `DialogFooter`

## Props / Variants

| prop | 型 | 既定 | 説明 |
|------|----|----|------|
| `DialogContent.onInteractOutside` | handler | — | EMR-64 focus restore と併用（外側クリック制御） |
| `DialogContent.ref` | `React.Ref` | — | React 19 の prop-ref |

## 状態

| 状態 | 実装 | 備考 |
|------|------|------|
| open/closed | ○ | `data-[state]` アニメーション（fade + zoom-95） |
| focus | ○ | EMR-64: open 時にフォーカスされた要素を記録し、controlled dialog でも close 時に復元（body 落ちしない） |
| overlay | ○ | `bg-black/50` + `fixed inset-0` |

## アクセシビリティ要件

- `DialogTitle` は必須（Radix が警告）。視覚的に出したくない場合でも内容は提供する。
- `DialogDescription` で目的を説明する（aria-describedby 接続）。
- Esc / overlay click で close — Radix 既定。確認必須の操作は `AlertDialog` を使う。

## 使い方（Do / Don't）

- Do: 破壊的確認は `AlertDialog`、情報表示・フォームは `Dialog`。
- Don't: `DialogContent` の `className` で `fixed`/`translate` を上書きしない（センタリングが壊れる）。

## 既知のギャップ

- モバイル fullscreen バリアントなし（sm:max-w-lg 固定幅のみ）。
- スクロールロック・focus 復元は実装済み（EMR-64）。
