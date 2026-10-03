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

## 閉じるボタン（X）の占有領域

`DialogContent` は右上に組み込みの閉じるボタンを absolute 配置する（`dialog.tsx`）。

- 占有領域: `top-4 right-4` + `min-h-11 min-w-11`（44px）→ **コンテンツ右上 上端60px × 右端60px** がボタン領域。
- `SheetContent` も同様に `top-1.5 right-1.5` + 44px（右上 約50px）。
- `AlertDialogContent` には閉じるボタンがない（本ルール対象外）。

ヘッダーやコンテンツ最上段に右寄せ要素（バッジ・ボタン・全幅入力）を置く場合、その右端を右端60px より内側に収めること。収めないと要素が X の下に潜り、クリック吸収・視覚衝突が起きる（実例: 受付詳細のステータスバッジ、PetEditModal の「飼主変更」ボタン、CommandDialog の検索入力行）。

目安となる padding（コンテンツ右端 = X 領域の内側 60px より左に置く）:

| コンテキスト | ヘッダー側の右 padding | ダイアログ端からの合計クリアランス |
|---|---|---|
| `DialogContent` に `p-0`＋自前 `px-6` ヘッダー | `pr-16` | 64px |
| `DialogContent` 既定 `p-6` 内のヘッダー | `pr-12` | 72px（24+48） |
| `CommandDialog` の全幅入力行 | input wrapper に `pr-16` | 64px |
| `SheetContent` ヘッダー | `pr-16` | 64px超（既定 `p-4` 込み） |

## 使い方（Do / Don't）

- Do: 破壊的確認は `AlertDialog`、情報表示・フォームは `Dialog`。
- Do: ヘッダー右端に要素を置くときは上記「閉じるボタンの占有領域」の padding で X 領域を避ける。
- Don't: `DialogContent` の `className` で `fixed`/`translate` を上書きしない（センタリングが壊れる）。
- Don't: コンテンツ最上段の右端60px領域に `justify-between` の右要素・全幅入力などを重ねない。

## 既知のギャップ

- モバイル fullscreen バリアントなし（sm:max-w-lg 固定幅のみ）。
- スクロールロック・focus 復元は実装済み（EMR-64）。
