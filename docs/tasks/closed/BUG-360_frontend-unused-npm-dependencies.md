# BUG-360: 未使用 pnpm 依存パッケージ 11件を削除

## 概要

`depcheck` により、`package.json` の `dependencies` に含まれるが、`frontend/src/` 内のどのファイルからも import されていないパッケージを 11件検出。shadcn/ui の `components/ui/` にも対応する UI コンポーネントファイルが存在しないことを確認済み。

## 優先度

**MEDIUM** — `pnpm install` 時間の短縮 + `node_modules` サイズ削減 + lock ファイルの保守負担軽減。

## 対象パッケージ（11件）

| パッケージ | 分類 |
|-----------|------|
| `@radix-ui/react-aspect-ratio` | shadcn/ui 未使用コンポーネントの依存 |
| `@radix-ui/react-collapsible` | 同上 |
| `@radix-ui/react-context-menu` | 同上 |
| `@radix-ui/react-hover-card` | 同上 |
| `@radix-ui/react-menubar` | 同上 |
| `@radix-ui/react-slider` | 同上 |
| `embla-carousel-react` | Carousel 未使用 |
| `input-otp` | OTP 入力未使用 |
| `react-resizable-panels` | Resizable Panels 未使用 |
| `vaul` | Drawer 未使用 |
| `zustand` | Store 未使用（`stores/` ディレクトリ不在） |

## 検出方法

`depcheck` + 手動 grep 検証（2026-04-14）
