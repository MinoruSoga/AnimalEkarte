# 動物病院管理システム デザインシステム

このドキュメントは、動物病院管理システムのUIデザインにおけるトーン&マナー（トンマナ）規約を定義します。本システムは、**Notionライトモード**のデザインシステムをベースに、クリーンでミニマル、かつ機能的なインターフェースを実現しています。

## デザイン思想

> **ミニマリズムと機能性の完璧なバランス**

- 長時間の使用でも目に優しいカラーパレット
- 一貫性のある視覚体験
- 直感的なインタラクション
- 情報密度と可読性のバランス

---

## カラーパレット

### 基本カラーシステム

本システムは、Notionの公式カラーパレットを採用しています。

#### 背景色

| 用途 | カラーコード | Tailwind Class | 説明 |
|------|------------|----------------|------|
| **メインコンテンツ背景** | `#FFFFFF` | `bg-white` | 純白。最大限の可読性を確保 |
| **サイドバー・画面背景** | `#F7F6F3` | `bg-[#F7F6F3]` | 暖かみのあるオフホワイト |
| **カード背景** | `#FFFFFF` | `bg-white` | 純白 |
| **テーブルヘッダー** | `#F7F6F3` | `bg-[#F7F6F3]` | 視覚的階層の作成 |
| **カードホバー** | `rgba(55, 53, 47, 0.06)` | `hover:bg-[rgba(55,53,47,0.06)]` | 極薄グレーオーバーレイ |
| **アイコン背景 (デフォルト)** | `#F7F6F3` | `bg-[#F7F6F3]` | マスタカード等のアイコン背景 |
| **アイコン背景 (ホバー)** | `#37352F` | `group-hover:bg-[#37352F]` | ホバー時に反転 |

#### CSS変数 (`/styles/globals.css`)

```css
:root {
  --background: #ffffff;
  --foreground: #37352F;
  --primary: #37352F;
  --primary-foreground: #ffffff;
  --secondary: #F1F0EE;
  --muted: #F1F0EE;
  --muted-foreground: #787774;
  --accent: #EBEAE8;
  --destructive: #EB5757;
  --border: rgba(55, 53, 47, 0.09);
  --input: rgba(55, 53, 47, 0.16);
  --ring: rgba(55, 53, 47, 0.16);
}
```

#### テキストカラー

Tailwindの任意値構文（`text-[#37352F]/60`など）を使用して透明度を制御します。

| 用途 | カラーコード | Tailwind Class | 説明 |
|------|------------|----------------|------|
| **プライマリテキスト** | `#37352F` | `text-[#37352F]` | 深みのあるチャコールグレー |
| **セカンダリテキスト** | `#9B9A97` | `text-[#9B9A97]` | グレーテキスト |
| **ラベル・補助テキスト** | `rgba(55, 53, 47, 0.6)` | `text-[#37352F]/60` | フォームラベル、補足情報 |
| **プレースホルダー** | `rgba(55, 53, 47, 0.4)` | `text-[#37352F]/40` | 入力例、件数表示 |
| **強調テキスト（やや薄）** | `rgba(55, 53, 47, 0.8)` | `text-[#37352F]/80` | 文脈上の強調 |
| **アイコンテキスト（薄）** | `rgba(55, 53, 47, 0.7)` | `text-[#37352F]/70` | アイコンデフォルト色 |
| **ごく薄テキスト** | `rgba(55, 53, 47, 0.3)` | `text-[#37352F]/30` | ChevronRight等の装飾的要素 |

#### ボーダーカラー

| 用途 | カラーコード | Tailwind Class | 説明 |
|------|------------|----------------|------|
| **標準ボーダー** | `rgba(55, 53, 47, 0.16)` | `border-[rgba(55,53,47,0.16)]` | カード、入力フィールド |
| **薄いボーダー** | `rgba(55, 53, 47, 0.09)` | `border-[rgba(55,53,47,0.09)]` | テーブル行区切り、フォームアクション区切り |
| **極薄ボーダー** | `rgba(55, 53, 47, 0.06)` | `border-[rgba(55,53,47,0.06)]` | 微細な区切り |
| **ホバーボーダー** | `rgba(55, 53, 47, 0.3)` | `hover:border-[rgba(55,53,47,0.3)]` | カードホバー時 |

#### アクセントカラー

| 用途 | カラーコード | Tailwind Class / 説明 |
|------|------------|------|
| **選択・フォーカス** | `#2EAADC` | Notionブルー (`focus-visible:ring-[#2EAADC]`) |
| **プライマリボタン** | `#37352F` | Notionブラック |
| **プライマリボタンホバー** | `#37352F/90` | 90%不透明度 |
| **破壊的アクション (インライン)** | `#E03E3E` | Notionレッド。削除ボタン、必須マーカー等に直接Tailwindクラスで使用 |
| **破壊的アクション (CSS変数)** | `#EB5757` | `--destructive` CSS変数。shadcn/ui の destructive variant で使用 |
| **破壊的ホバー背景** | `#E03E3E/10` | 薄い赤背景 |

---

## タイポグラフィ

### フォントサイズ階層

Tailwindのデフォルトタイポグラフィを使用しますが、特定のコンテキストでは以下のサイズを遵守してください。

| 要素 | サイズ | Tailwind Class | 用途 |
|------|-------|----------------|------|
| **カード内タイトル** | 20px | `text-xl` | ペット名、飼主名などの主要情報 |
| **本文** | 14px | `text-sm` | 一般的なテキスト、フォーム入力値、テーブルセル |
| **補足・ラベル** | 12px | `text-xs` | フィールドラベル、日付、メタデータ |

### 行間

- **主要情報**: `leading-tight` を使用して引き締まった印象を与えます。

---

## 余白パターン (Spacing)

### ページレベル

| コンテキスト | パターン | 説明 |
|---|---|---|
| **リストページ** | `gap-4` | カード間・セクション間の標準間隔 |
| **medical-records タブコンテンツ** | `gap-3 pb-16` | タブ内の間隔 + 下部余白 |
| **フォーム全体** | `space-y-4` | フォーム内のフィールド間隔 |
| **モーダル** | `space-y-4` | モーダル内のセクション間隔 |
| **カードパディング** | `p-4` or `p-6` | カード内のパディング |

### フォーム内部

| コンテキスト | パターン | 説明 |
|---|---|---|
| **ラベル ↔ 入力間** | `space-y-1.5` | Label と Input の間隔 |
| **フォームアクション** | `pt-4 border-t border-[rgba(55,53,47,0.09)]` | 保存/キャンセルボタンの区切り |
| **グリッドフォーム** | `grid grid-cols-1 md:grid-cols-2 gap-4` | 2カラムフォームレイアウト |

### マスタ設定カードグリッド

| コンテキスト | パターン | 説明 |
|---|---|---|
| **カードグリッド** | `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3` | 設定カード一覧 |
| **セクション間** | `mb-6` | セクション見出し間の余白 |
| **最終セクション** | `mb-6 pb-16` | 最終セクションの下部余白 |

### 禁止パターン

- **`h-[calc(100vh-...)]`**: 使用禁止。`pb-16` 等の固定パディングで代替
- **`overflow-y-scroll`**: `overflow-y-auto` を使用

---

## コンポーネントスタイリング

### カード

```tsx
<div className="bg-white rounded-lg p-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)] border border-[rgba(55,53,47,0.16)]">
  {/* カードコンテンツ */}
</div>
```

**仕様:**
- 背景: `#FFFFFF`
- ボーダー: `rgba(55, 53, 47, 0.16)`
- 角丸: `8px` (`rounded-lg`)
- パディング: `16px` (`p-4`) or `24px` (`p-6`)
- シャドウ: `0 1px 3px rgba(0,0,0,0.04)`

### ホバー付きカード (マスタ設定等)

```tsx
<button className="w-full text-left bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)] hover:shadow-md hover:border-[rgba(55,53,47,0.3)] transition-all group cursor-pointer">
  {/* カードコンテンツ */}
</button>
```

### 入力フィールド

```tsx
<Input
  className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
  placeholder="例: 入力してください"
/>
```

**仕様:**
- 高さ: `40px` (`h-10`)
- 背景: `#FFFFFF` (明示的に指定)
- テキスト色: `#37352F`
- ボーダー: `rgba(55, 53, 47, 0.16)`
- フォーカスリング: `#2EAADC`
- 角丸: `6px` (`rounded-md`)

### ボタン

#### PrimaryButton (共有コンポーネント)

```tsx
import { PrimaryButton } from "../components/shared/PrimaryButton";

<PrimaryButton onClick={handleSave} className="gap-2">
  <Save className="size-4" />
  保存
</PrimaryButton>
```

**実装:**
```tsx
// bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm shadow-sm border-transparent
```

#### 破壊的ボタン

```tsx
<Button
  variant="ghost"
  size="sm"
  className="h-10 text-sm text-[#E03E3E] hover:bg-[#E03E3E]/10 hover:text-[#E03E3E]"
>
  <Trash2 className="mr-1.5 size-4" />
  削除
</Button>
```

#### アウトラインボタン

```tsx
<Button
  variant="outline"
  size="sm"
  className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
>
  キャンセル
</Button>
```

### テーブル (DataTable)

```tsx
import { DataTable } from "../components/shared/DataTable";
import { DataTableRow } from "../components/shared/DataTableRow";

<DataTable
  columns={columns}
  data={items}
  emptyMessage="データが見つかりません"
  renderRow={(item) => (
    <DataTableRow key={item.id} onClick={() => handleClick(item)}>
      <TableCell className="text-sm text-[#37352F] py-2">{item.name}</TableCell>
    </DataTableRow>
  )}
/>
```

**デザイントークン (`/lib/design-tokens.ts`):**
```typescript
export const TABLE_STYLES = {
  row: "border-b-[rgba(55,53,47,0.09)] hover:bg-[#F7F6F3]/50 transition-colors cursor-pointer h-12",
  actionButton: "h-10 w-10 text-[#37352F]/60 hover:text-[#37352F]",
  cell: "text-sm text-[#37352F] py-2",
  cellMono: "font-mono text-sm text-[#37352F] py-2",
};
```

**仕様:**
- コンテナボーダー: `rgba(55, 53, 47, 0.16)`
- ヘッダー背景: `#F7F6F3`
- ヘッダーテキスト: `text-[#37352F]/60`
- 行ホバー色: `#F7F6F3/50`
- 行高さ: `h-12`
- 行間ボーダー: `rgba(55, 53, 47, 0.09)` (薄いボーダー)

### ステータスバッジ

```tsx
import { StatusBadge } from "../components/shared/StatusBadge";

<StatusBadge colorClass={getMasterStatusColor(item.status)}>
  {getMasterStatusLabel(item.status)}
</StatusBadge>
```

### ページレイアウト (PageLayout)

```tsx
import { PageLayout } from "../components/shared/PageLayout";

<PageLayout
  title="マスタ設定"
  icon={<Settings className="size-5 text-[#37352F]" />}
  onBack={() => navigate("/settings")}
  headerAction={<PrimaryButton>新規登録</PrimaryButton>}
  maxWidth="max-w-full"
>
  {/* ページコンテンツ */}
</PageLayout>
```

**仕様:**
- 外枠背景: `#F7F6F3`
- コンテンツは `flex-1 overflow-y-auto`
- maxWidth デフォルト: `max-w-[1440px]`

### 検索バー

```tsx
import { SearchFilterBar } from "../components/shared/SearchFilterBar";

<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  placeholder="コード、名称で検索..."
  count={filteredItems.length}
/>
```

---

## アイコン

**lucide-react** を使用します。アイコンコンポーネントは `LucideIcon` 型で管理し、各コンテキストで異なるクラスを適用します。

| サイズ | Tailwind Class | 用途 |
|--------|---------------|------|
| **標準** | `size-4` (16px) | ボタン内アイコン、検索アイコン |
| **ヘッダー** | `size-5` (20px) | ページタイトルアイコン |
| **小** | `size-3` (12px) | 補足的なアイコン |

### アイコンの色指定パターン

| コンテキスト | クラス | 説明 |
|---|---|---|
| **ページヘッダー** | `text-[#37352F]` | ページタイトル横のアイコン |
| **カード内 (デフォルト)** | `text-[#37352F]/70` | カード内のアイコン |
| **カード内 (ホバー)** | `group-hover:text-white` | ホバー時に反転 |
| **テーブルアクション** | `text-[#37352F]/60` → `hover:text-[#37352F]` | 操作ボタン |
| **装飾的** | `text-[#37352F]/30` → `group-hover:text-[#37352F]/60` | ChevronRight等 |

---

## スクロールバー

カスタムスクロールバーを透明トラックで統一しています。

```css
/* globals.css */
scrollbar-color: rgba(55, 53, 47, 0.2) transparent;
scrollbar-width: thin;

/* WebKit */
::-webkit-scrollbar-thumb {
  background: rgba(55, 53, 47, 0.2);
  border-radius: 9999px;
}
::-webkit-scrollbar-thumb:hover {
  background: rgba(55, 53, 47, 0.35);
}
::-webkit-scrollbar-track {
  background: transparent;
}
```

---

## アクセシビリティ

- **コントラスト**: テキストには `#37352F` (100% or 80%) または `rgba(55, 53, 47, 0.6)` を使用し、可読性を確保
- **フォーカス**: すべてのインタラクティブ要素（入力、ボタン）には `focus-visible:ring` を適用
- **キーボードナビゲーション**: shadcn/ui (Radix UI) ベースにより、ダイアログ・ドロップダウン等はキーボード操作対応済み

---

## まとめ

このデザインシステムは、**Notionライトモード**の美学を忠実に再現することを目的としています。特に以下の3点に注意を払い、一貫性のある実装を行ってください。

1. **ボーダーの色**: `rgba(55,53,47,0.16)` (標準) / `rgba(55,53,47,0.09)` (薄い) の使い分け
2. **テキストの透明度**: `#37352F` の不透明度バリエーションで階層を表現
3. **余白の統一**: 本ドキュメントの余白パターン表に準拠
