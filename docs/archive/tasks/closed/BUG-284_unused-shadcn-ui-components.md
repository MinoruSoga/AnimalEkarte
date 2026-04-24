# BUG-284: 未使用 shadcn/ui コンポーネント 12件 — バンドルに含まれる dead code

## 概要
`components/ui/` 配下に shadcn/ui のセットアップ時に追加されたが、プロジェクト全体で一度もimportされていないコンポーネントが12件存在する。Tree-shaking で除外されるためバンドルサイズへの実影響は軽微だが、コードベースのノイズ削減のため削除する。

## 再現手順
以下のコマンドで各ファイルが0件であることを確認:
```bash
# 例: aspect-ratio.tsx
grep -r "ui/aspect-ratio" frontend/src --include="*.tsx" --include="*.ts" \
  | grep -v "components/ui/aspect-ratio.tsx"
# → 0件

# 全件一括確認スクリプト
for f in aspect-ratio carousel context-menu hover-card input-otp menubar \
          breadcrumb chart collapsible drawer resizable slider; do
  count=$(grep -r "ui/$f" frontend/src --include="*.tsx" --include="*.ts" \
    | grep -v "components/ui/$f" | wc -l | tr -d ' ')
  echo "$f: $count usages"
done
# 全て 0 usages
```

## 期待する動作
- 実際に使用されているコンポーネントのみが `components/ui/` に存在する
- または、将来使用予定のものはドキュメントで明示されている

## 現状コード

### 未使用コンポーネント一覧（12件）
```
frontend/src/components/ui/
├── aspect-ratio.tsx    ← 未使用
├── breadcrumb.tsx      ← 未使用
├── carousel.tsx        ← 未使用
├── chart.tsx           ← 未使用（recharts への依存を持つ大型コンポーネント）
├── collapsible.tsx     ← 未使用
├── context-menu.tsx    ← 未使用
├── drawer.tsx          ← 未使用（vaul ライブラリへの依存あり）
├── hover-card.tsx      ← 未使用
├── input-otp.tsx       ← 未使用（input-otp ライブラリへの依存あり）
├── menubar.tsx         ← 未使用
├── resizable.tsx       ← 未使用（react-resizable-panels への依存あり）
└── slider.tsx          ← 未使用
```

**注意**: `toggle.tsx` は `toggle-group.tsx` から内部参照されているため除外。`alert-dialog.tsx` は使用中のため除外。

## 影響範囲

| 対象ファイル | 詳細 | 状態 |
|------------|------|------|
| `components/ui/aspect-ratio.tsx` | 未使用 | 削除対象 |
| `components/ui/breadcrumb.tsx` | 未使用 | 削除対象 |
| `components/ui/carousel.tsx` | 未使用 | 削除対象 |
| `components/ui/chart.tsx` | 未使用（recharts依存） | 削除対象 |
| `components/ui/collapsible.tsx` | 未使用 | 削除対象 |
| `components/ui/context-menu.tsx` | 未使用 | 削除対象 |
| `components/ui/drawer.tsx` | 未使用（vaul依存） | 削除対象 |
| `components/ui/hover-card.tsx` | 未使用 | 削除対象 |
| `components/ui/input-otp.tsx` | 未使用（input-otp依存） | 削除対象 |
| `components/ui/menubar.tsx` | 未使用 | 削除対象 |
| `components/ui/resizable.tsx` | 未使用（react-resizable-panels依存） | 削除対象 |
| `components/ui/slider.tsx` | 未使用 | 削除対象 |

## 修正方針

### 1. 削除前確認（各コンポーネントの使用状況を最終確認）
```bash
for f in aspect-ratio breadcrumb carousel chart collapsible context-menu \
          drawer hover-card input-otp menubar resizable slider; do
  count=$(grep -r "ui/$f" frontend/src --include="*.tsx" --include="*.ts" \
    | grep -v "components/ui/$f" | wc -l | tr -d ' ')
  echo "$f: $count usages"
done
```
全て 0 であることを確認後に削除する。

### 2. ファイル削除
```bash
cd frontend/src/components/ui/
rm aspect-ratio.tsx breadcrumb.tsx carousel.tsx chart.tsx collapsible.tsx \
   context-menu.tsx drawer.tsx hover-card.tsx input-otp.tsx menubar.tsx \
   resizable.tsx slider.tsx
```

### 3. 不要な pnpm 依存のチェック（オプション）
削除後、以下のパッケージが他で使われていない場合は `package.json` から削除を検討:
- `recharts`（chart.tsx のみで使用の場合）
- `vaul`（drawer.tsx のみで使用の場合）
- `input-otp`（input-otp.tsx のみで使用の場合）
- `react-resizable-panels`（resizable.tsx のみで使用の場合）

```bash
grep -r "recharts\|vaul\|input-otp\|react-resizable-panels" frontend/src \
  --include="*.tsx" --include="*.ts" | grep -v "components/ui/"
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/performance-rules.md` — Bundle Size
> Bundle Size: < 200KB (JS)

未使用コンポーネントは Tree-shaking で除外されるが、依存ライブラリ（recharts, vaul等）が package.json に残ると pnpm install 時にダウンロードされる。CI/CD のビルド時間にも影響。

### プロジェクト内参照実装
- `components/ui/button.tsx`, `components/ui/input.tsx` — 実際に使用されているコンポーネントの例

## 優先度
**Low** — Tree-shaking により実際のバンドルサイズへの影響は限定的。コードベースの可読性向上が主目的。

## 関連チケット
- BUG-280: 安全削除対象コンポーネント（DateRangePicker等）

## 関連ファイル
- `frontend/src/components/ui/` — 削除対象ファイルのディレクトリ
- `frontend/package.json` — 不要依存の削除確認先
