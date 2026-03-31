# 動物病院管理システム デザインシステム

このドキュメントは、動物病院管理システムのUIデザインにおけるトーン&マナー（トンマナ）規約を定義します。本システムは、**Notionライトモード**のデザインシステムをベースに、クリーンでミニマル、かつ機能的なインターフェースを実現しています。

## デザイン思想

> **ミニマリズムと機能性の完璧なバランス**

- **Single Source of Truth**: すべてのスタイル定数は `frontend/src/lib/design-tokens.ts` で管理されます。
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
  --primary: #038B94; /* メインカラー（ティール）に変更 */
  --primary-foreground: #ffffff;
  --secondary: #F1F0EE;
  --muted: #F1F0EE;
  --muted-foreground: #787774;
  --accent: #EBEAE8;
  --destructive: #EB5757;
  --border: rgba(55, 53, 47, 0.09);
  --input: rgba(55, 53, 47, 0.16);
  --input-background: rgba(242, 241, 238, 0.6);
  --switch-background: #E3E2E0;
  --font-weight-medium: 500;
  --font-weight-normal: 400;
  --ring: rgba(55, 53, 47, 0.16);
}
```

#### タブレットファーストスケーリング (`@theme inline`)

iPad縦向き（1024×1366）/横向き（1366×1024）を想定し、Tailwind v4の `@theme inline` でベースサイズを拡大しています。

```css
@theme inline {
  --text-xs: 0.8125rem;          /* 13px (default 12px) → +8% */
  --text-xs--line-height: 1.125rem;
  --text-sm: 0.9375rem;          /* 15px (default 14px) → +7% */
  --text-sm--line-height: 1.375rem;
  --spacing: 0.275rem;           /* +10% base spacing (default 0.25rem) */
}
```

> これにより `text-xs` / `text-sm` / spacing ユーティリティが全体で10%前後拡大され、タブレットでの可読性・タッチ操作性が向上します。

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
| **アクセントブルー (Notion Blue)** | `#2383E2` | Notionブルー (`focus-visible:ring-[#2383E2]`) |
| **デストラクティブ (Danger)** | `#C0392B` | WCAG AA準拠 (コントラスト比 7.1:1)。以前の `#EB5757` から変更。 |
| **ブランドカラー (Veterinary Teal)** | `#038B94` | 病院のメインカラー。 |
| **メインティール (Teal)** | `#038B94` | システムのメインカラー。ヘッダーや主要な装飾に使用 |
| **プライマリボタン (PrimaryButton)** | `#2383E2` | Notionブルー。`STYLE.btnPrimary` / `STYLE.btnAccent` で使用。影なし（フラット） |
| **破壊的アクション (インライン)** | `#E03E3E` | Notionレッド。削除ボタン、必須マーカー等に直接Tailwindクラスで使用 |
| **破壊的アクション (CSS変数)** | `#EB5757` | `--destructive` CSS変数。shadcn/ui の destructive variant で使用 |
| **破壊的ホバー背景** | `#E03E3E/10` | 薄い赤背景 |

#### シフト管理カラー（Notion風バッジカラー）

| シフトタイプ | 背景色 | テキスト色 | 説明 |
|---|---|---|---|
| **通常勤務 (full)** | `#DDEDEA` | `#0F7B6C` | Notionグリーン |
| **午前のみ (morning)** | `#D3E5EF` | `#183B56` | Notionブルー |
| **午後のみ (afternoon)** | `#EBE2F5` | `#5B2D8E` | Notionパープル |
| **休み (off)** | `#EBECED` | `#9B9A97` | Notionグレー |
| **有給 (paid_leave)** | `#FDECC8` | `#7B5B29` | Notionオレンジ |

> シフトタイプ別カラーは `/features/shifts/types/index.ts` の `SHIFT_TYPE_COLOR_MAP` で一元管理。ロールバッジも同パレット（医師=グリーン、スタッフ=ブルー、トリマー=パープル）で統一。

#### 予約カレンダーカラー（Figma準拠パステルパレット）

**デザイン方針**: Figmaデザインから抽出した明るく柔らかいパステルカラーを採用。暗い色は極力使用せず、目に優しく美しいトーンで統一。

| 診療種別 | 背景色 | テキスト色 | ボーダー色 | ドット色（凡例） | 説明 |
|---|---|---|---|---|---|
| **診療** | `#dbeafe` | `#5b8def` | `#93c5fd` | `#5b8def` | 明るい青（Figma pale blue） |
| **検診** | `#dcfce7` | `#16a34a` | `#86efac` | `#16a34a` | 明るい緑（Figma pale green） |
| **手術** | `#ffe2e2` | `#f87171` | `#fca5a5` | `#f87171` | 明るいコーラルレッド（Figma pale red） |
| **ワクチン** | `#f3e8ff` | `#a855f7` | `#c084fc` | `#a855f7` | 明るいラベンダー紫（Figma pale purple） |
| **入院** | `#cefafe` | `#0891b2` | `#67e8f9` | `#0891b2` | 明るいターコイズ（Figma pale cyan） |
| **トリミング** | `#ffedd4` | `#f97316` | `#fdba74` | `#f97316` | 明るいピーチオレンジ（Figma pale orange） |

**実装場所**:
- `/lib/design-tokens.ts`: `bgFigmaBlue` / `textFigmaBlue` / `borderFigmaBlue` / `dotFigmaBlue` などのTailwindクラス
- `/features/master/constants/service-type-colors.ts`: `SERVICE_TYPE_COLOR_MAP` でマスタ連動
- 予約カレンダー（`/features/reservations/routes/ReservationManagement.tsx`）で`useServiceTypeColorMap`フックにより動的適用
- カラーは`ServiceTypeMaster.color`フィールドで管理され、マスタ設定から変更可能

**使用例**:
```tsx
// 診療種別マスタの色設定
const serviceType = {
  name: "診療",
  color: "blue", // → bgFigmaBlue + textFigmaBlue + borderFigmaBlue
};

// カレンダーレジェンド
<span className={`w-2.5 h-2.5 rounded-full ${C.dotFigmaBlue}`} />
<span className={`text-xs ${C.text60}`}>診療</span>
```

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

### 注意事項
- Tailwind の `text-*` (フォントサイズ), `font-*` (フォントウェイト), `leading-*` (行間) クラスは、ユーザーから明示的な変更指示がない限り追加しない
- `/styles/globals.css` の `@layer base` でデフォルトのタイポグラフィが定義済み

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

### シャドウ使用ポリシー

本システムでは **Notionライクなフラットデザイン** を基本とし、shadow 使用を以下の3クラスに限定します。

| クラス | 許可用途 | 具体例 |
|---|---|---|
| **`shadow-lg`** | floating overlay（最前面の浮動要素）| コンテキストメニュー・ポップオーバー（`QuickActionMenu`・`CarePlanPreviewPopover`・`ShiftEditPopover`）|
| **`shadow-md`** | absolute 配置の浮動ボタン | 画像削除ボタン（`TrimmingForm` スタイル/完成画像オーバーレイ）|
| **`shadow-sm`** | micro decoration のみ | 今日日付バッジ（`MonthView`）・ステータスドット（`WeekView`）・選択中状態ボタン（`PatientSelectionTable`）|

**禁止パターン（DESIGN_SYSTEM 違反）:**
- カード・フォームパネル・テーブルコンテナへの `shadow-sm` 適用 → `border ${C.borderMedium}` のみで深度表現
- ページヘッダーボタン（削除・保存）への `shadow-lg` 付与 → shadcn/ui `Button` のデフォルトスタイルに委ねる
- `hover:shadow-md` によるホバー昇格 → `hover:ring-1 hover:ring-[#37352F]/20`（`C.ringPrimary20`）で代替

**ボタンシャドウポリシー:**
- **塗りつぶしボタン** (`btnPrimary` / `btnAccent` / `btnDanger`): `shadow-none`（Notionのfilled buttonはフラット）
- **アウトラインボタン** (`btnOutline` / button.tsx `variant="outline"`): `shadow-[var(--notion-shadow-btn)]`（微細な陰影で立体感を付与）
- **ゴースト/セカンダリボタン**: シャドウなし（ホバー背景色のみ）

### カード

```tsx
<div className="bg-white rounded-lg p-4 border border-[rgba(55,53,47,0.16)]">
  {/* カードコンテンツ */}
</div>
```

**仕様:**
- 背景: `#FFFFFF`
- ボーダー: `rgba(55, 53, 47, 0.16)`
- 角丸: `8px` (`rounded-lg`)
- パディング: `16px` (`p-4`) or `24px` (`p-6`)
- シャドウ: **なし**（border のみで深度を表現）

### ホバー付きカード (マスタ設定等)

```tsx
<button className="w-full text-left bg-white rounded-lg border border-[rgba(55,53,47,0.16)] p-4 hover:ring-1 hover:ring-[#37352F]/20 hover:border-[rgba(55,53,47,0.3)] transition-all group cursor-pointer">
  {/* カードコンテンツ */}
</button>
```

### 検索可能なセレクトボックス (Combobox)

全セレクトボックス（プルダウン）は、選択肢が多い場合や一貫したUXを提供するため、`Popover` と `Command` を組み合わせた **コンボボックス化（Combobox）** を標準とします。標準の `Select` コンポーネントは原則使用しません。

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
- フォーカスリング: `#2383E2`
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
// bg-[#2383E2] hover:bg-[#1B6EC2] text-white h-10 text-sm shadow-none border-transparent rounded-[4px]
// STYLE.btnPrimary — Notionアクセントブルー、影なしフラット
```

#### 破壊的ボタン

```tsx
<Button
  variant="ghost"
  size="sm"
  className="h-10 text-sm text-[#E03E3E] hover:bg-[#E03E3E]/10 hover:text-[#E03E3E]"
>
  <Trash2 className="size-4" />
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
  row: "border-b border-[rgba(55,53,47,0.09)] hover:bg-[#F7F6F3]/50 transition-colors cursor-pointer h-12",
  actionButton: "h-10 w-10 text-[#37352F]/60 hover:text-[#37352F]",
  cell: "text-sm text-[#37352F] py-2.5",
  cellMono: "font-mono text-sm text-[#37352F] py-2.5",
};
```

**仕様:**
- コンテナボーダー: `rgba(55, 53, 47, 0.16)`
- ヘッダー背景: `#F7F6F3`
- ヘッダーテキスト: `text-[#37352F]/70`
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
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";

<NotionFilter
  properties={FILTER_PROPERTIES}
  activeFilters={activeFilters}
  onFilterChange={setActiveFilters}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="コード、名称で検索..."
  count={filteredItems.length}
/>
```

---

## フォーム保護パターン

### NavigationBlocker

フォームの未保存変更がある状態でのページ離脱を防止します。

```tsx
import { NavigationBlocker } from "../components/shared/NavigationBlocker";

<NavigationBlocker
  when={isDirty}
  title="変更が保存されていません"
  description="保存されていない変更があります。ページを離れますか？"
/>
```

**実装方式:**
- `useBlocker` を呼ぶ内部コンポーネント `NavigationBlockerDialog` を分離
- `when` が `true` のときだけマウント（初期ロード時の React Router 警告を回避）

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

### テーブル等の操作アイコン統一
各種アイテムの行における「操作（アクション）」ボタンのアイコンは、原則として **`MoreHorizontal`** (3点リーダー) を使用し、サイズは **`size-5`** で統一します。

### UIからの内部ID表示の削除
システム内部で利用する識別子（UUID等の内部ID）は、デバッグ目的を除き、UI上には表示しません。ユーザーにとって意味のある「患者番号」「カルテ番号」「予約番号」などの業務コードのみを表示してください。

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

## ドラッグ&ドロップ インタラクション

マスタ設定テーブルでは、HTML5ネイティブD&D APIを使用したNotionライクな並び替えを実装しています。

### ドラッグハンドル

| 要素 | スタイル | 説明 |
|------|---------|------|
| **アイコン** | `GripVertical` (`size-3.5`) | 6点ドットアイコン |
| **通常時** | `text-[#37352F]/20` | 控えめに表示 |
| **ホバー時** | `hover:text-[#37352F]/50` | 操作可能であることを示唆 |
| **カーソル** | `cursor-grab` / `active:cursor-grabbing` | ドラッグ可能状態を明示 |

### カスタムドラッグレビュー（`setDragImage`）

Notionライクなピル型ゴーストを`document.createElement`で動的生成し、`setDragImage`で適用します。

```
┌─────────────────────────┐
│  ⁞  項目名テキスト       │
└─────────────────────────┘
```

| プロパティ | 値 | 説明 |
|-----------|-----|------|
| **背景** | `white` | 白背景 |
| **ボーダー** | `1px solid rgba(55,53,47,0.16)` | Notion標準ボーダー |
| **角丸** | `6px` | ピル型 |
| **影** | `0 4px 12px rgba(0,0,0,0.12), 0 0 0 1px rgba(0,0,0,0.04)` | 浮遊感のあるシャドウ |
| **フォント** | `13px`, `font-weight: 500` | Notionのテキストスタイル |
| **グリップアイコン** | Unicode `⁝` (`\u2807`), `color: rgba(55,53,47,0.3)` | ドラッグ中であることを視覚化 |
| **配置** | `position: fixed; top: -1000px` | 画外に配置（DragImage用） |
| **クリーンアップ** | `dragEnd` 時に `remove()` | DOM汚染防止 |

### ドロップ位置インジケータ（`box-shadow`方式）

ドロップ可能位置を行の`box-shadow`で示します。背景色変更ではなく`box-shadow`を使うことで、既存のホバースタイルと干渉しません。

| 位置 | box-shadow | 用途 |
|------|-----------|------|
| **前に挿入** | `inset 0 2px 0 0 #2383E2` | 行の上端に青いライン |
| **後に挿入** | `inset 0 -2px 0 0 #2383E2` | 行の下端に青いライン |
| **子項目化（ツリーのみ）** | `inset 0 0 0 2px #2383E2` | 行全体を青い枠で囲む |

インジケータカラー `#2383E2` はNotionのリンク/選択色に準拠しています。

### ドロップ位置判定（マウスY座標比率）

| テーブル種別 | ゾーン分割 | 説明 |
|-------------|-----------|------|
| **ツリーテーブル** | 上端25% = before / 中央50% = on / 下端25% = after | 3段階判定（reparent対応） |
| **フラットテーブル** | 上半分50% = before / 下半分50% = after | 2段階判定（reorderのみ） |

### ドラッグ中の行スタイル

| 状態 | スタイル | 説明 |
|------|---------|------|
| **ドラッグ元** | `opacity-40` | 移動中の項目を半透明化 |
| **通常行** | 変化なし | ドラッグ操作に影響されない |

### ホバー自動展開（ツリーテーブルのみ）

折りたたまれた親ノードの中央ゾーン（"on" 位置）に **600ms** ホバーし続けると、自動的に展開されます。

| パラメータ | 値 | 説明 |
|-----------|-----|------|
| **遅延時間** | `600ms` | 意図しない展開を防止するディレイ |
| **トリガー条件** | `pos === "on"` かつ `hasChildren` かつ `!isExpanded` | 折りたたまれた子持ちノードのみ |
| **キャンセル** | マウスが離れる or ゾーンが"on"以外に変化 | `clearTimeout`で即キャンセル |

### 安全制約

| 制約 | 実装 | 説明 |
|------|------|------|
| **自己ドロップ防止** | `srcId === targetId` チェック | 自分自身へのドロップを無視 |
| **子孫ドロップ防止** | `collectDescendantIds()` | 循環参照を防止（ツリーのみ） |
| **sortOrder永続化** | `bulkUpdate` で兄弟全体を一括更新 | 整合性のある順序管理 |

---

## アクセシビリティ

- **コントラスト**: テキストには `#37352F` (100% or 80%) または `rgba(55, 53, 47, 0.6)` を使用し、可読性を確保
- **フォーカス**: すべてのインタラクティブ要素（入力、ボタン）には `focus-visible:ring` を適用
- **キーボードナビゲーション**: shadcn/ui (Radix UI) ベースにより、ダイアログ・ドロップダウン等はキーボード操作対応済み
- **タッチターゲット**: iPad縦向き（1024×1366）/横向き（1366×1024）を想定したタブレットファーストUI
  - ボタン・入力フィールド: 最小 `h-10`（40px）、アイコンボタン: `size-9` 以上
  - ナビゲーションボタン等の小型ボタン: `after:absolute after:-inset-2` でヒットエリア拡張
  - テーブル行: `h-12`（48px）
  - 閉じるボタン: `size-8` + `after:absolute after:-inset-2`
- **フォーカストラップ**: `useFocusTrap` フックでモーダル/Popover内のTab循環・Escape閉じ・フォーカス復帰を一元管理。`ShiftEditPopover`、マスタ選択モーダル等で使用
- **D&Dアクセシビリティ**: ドラッグハンドルに `role="button"`, `tabIndex={0}`, `aria-roledescription`, `aria-label` を付与。`aria-grabbed` / `aria-dropeffect="move"` で状態をスクリーンリーダーに通知。`Alt+ArrowUp/Down` でキーボードのみの並び替え、`Alt+ArrowLeft/Right` でツリー階層変更（昇格/降格、循環参照防止付き）に対応。フォーカスリングは `focus-visible:ring-[1.5px] focus-visible:ring-[#2383E2]`。`useAnnounce` で全D&D操作結果を `aria-live="polite"` リージョンにリアルタイム通知
- **フォームバリデーション**: `FormFieldError` に `role="alert"` + `aria-live="assertive"` + 一意の `id` を付与。関連する入力要素にエラー時のみ `aria-describedby` を条件適用。`PatientInfoCard.staffAriaDescribedBy` / `MasterSelectTrigger.ariaDescribedBy` prop で分離コンポーネント間のエラー接続をサポート
- **インタラクティブ要素**: `MasterSelectTrigger`（選択済み/未選択）・`MasterSelectModal` アイテムリストはすべて `<button>` 要素を使用し、キーボードのみでの操作を保証
- **ライブリージョン**: `LiveAnnouncerProvider` + `useAnnounce` フックによるスクリーンリーダー向けリアルタイム通知基盤（D&D操作・フォーム送信結果等）
- **トグルボタン**: `aria-pressed` でオン/オフ状態を通知。ボタングループには `role="group"` + `aria-label` を付与
- **モーション制御 (`useReducedMotion`)**: `prefers-reduced-motion: reduce` を検知する共有フック。Motionアニメーションを使用する全コンポーネントで必ず適用し、ユーザー設定に応じてアニメーションを抑制する（WCAG 2.3.3準拠）
  - **ポリシー**: Motionの `duration`・`scale`・`opacity`・`shadow` 等の視覚効果を `reduced ? { duration: 0 } : normalTransition` パターンで切り替え。レイアウトシフト（要素の追加/削除に伴う高さ変化）は reduced 時も維持可
  - **適用箇所**: `TreatmentTable`（行フェードイン/ハイライトフラッシュ/エグジット）、`MasterSidePeek`（サイドパネル開閉duration→0）、`WeekView.AppointmentCard`（ドラッグ時scale/opacity/shadow効果抑制）
  - **新規Motion追加時**: `useReducedMotion()` の呼び出しと分岐を必須とし、`SPECIFICATION.md` の共有フック一覧に適用箇所を追記すること

---

## Notion風チェックボックス (`NotionCheckbox`)

shadcn/ui のデフォルト `Checkbox` ではなく、Notion の実際のチェックボックスUIを再現した `NotionCheckbox` 共有コンポーネントを使用します。

### ビジュアル仕様

| 状態 | スタイル | 説明 |
|------|---------|------|
| **未チェック** | `bg-white`, `border-[1.5px] border-[rgba(55,53,47,0.25)]`, `rounded-[3px]` | 白背景、薄グレーボーダー |
| **未チェック + ホバー** | `bg-[rgba(35,131,226,0.08)]`, `border-[rgba(35,131,226,0.4)]` | Notionブルーのティント |
| **チェック済** | `bg-[#2383E2]`, `border-[#2383E2]` | Notionアクセントブルー塗りつぶし |
| **チェックマーク** | 白色SVG (`strokeWidth: 1.8`, `strokeLinecap: round`) | Notion風の細いチェックマーク |
| **フォーカス** | `ring-2 ring-[#2383E2]/40 ring-offset-1` | アクセントブルーのフォーカスリング |
| **無効** | `opacity-40 cursor-not-allowed` | 操作不可状態 |

### サイズ

- デフォルト: `size-[18px]`（Notionと同等）
- `className` prop でサイズカスタマイズ可

### アクセシビリティ

- `role="checkbox"` + `aria-checked` + `aria-disabled`
- `tabIndex={0}` によるキーボードフォーカス
- `Space` / `Enter` キーでトグル
- `stopPropagation` によるラッパー要素との二重トグル防止

### 使用パターン

```tsx
// 単独使用（TreatmentTable等）
<NotionCheckbox
  checked={isChecked}
  onCheckedChange={handleChange}
  aria-label="項目を完了にする"
/>

// ラベル付き使用（Dashboard・TrimmingForm等）
// ※ <label htmlFor> はdiv[role=checkbox]に効かないため、
//    ラッパーdivのonClickで代替する
<div
  className="flex items-center gap-2 cursor-pointer select-none"
  onClick={handleToggle}
  role="none"
>
  <NotionCheckbox checked={isChecked} onCheckedChange={handleChange} />
  <span className="text-sm">ラベルテキスト</span>
</div>
```

---

## まとめ

このデザインシステムは、**Notionライトモード**の美学を忠実に再現することを目的としています。特に以下の3点に注意を払い、一貫性のある実装を行ってください。

1. **ボーダーの色**: `rgba(55,53,47,0.16)` (標準) / `rgba(55,53,47,0.09)` (薄い) の使い分け
2. **テキストの透明度**: `#37352F` の不透明度バリエーションで階層を表現
3. **余白の統一**: 本ドキュメントの余白パターン表に準拠

---

## デザイントークンアーキテクチャ (`/lib/design-tokens.ts`)

全カラー値・レイアウト定数・コンポジットTailwindクラスプリセットは `/lib/design-tokens.ts` に一元管理されています。コンポーネントはこのファイルからインポートし、ハードコーディングされたhex値やpxサイズの直接記述を避けます。

### 5つのエクスポートカテゴリ

#### 1. `PALETTE` — 生カラー値

`style` prop、Canvas描画、または生の文字列が必要な箇所で使用します。

| キー | 値 | 用途 |
|------|-----|------|
| `primary` | `#37352F` | Notionプライマリ（テキスト、アイコン、フィル） |
| `bgMain` | `#F7F6F3` | ページ/セクション背景 |
| `bgSubtle` | `#FAFAF8` | やや薄い背景（日時カード等） |
| `bgActive` | `#EAE9E5` | サイドバーアクティブ行 |
| `white` | `#ffffff` | 白 |
| `borderLight` | `rgba(55,53,47,0.09)` | ボーダー（薄い） |
| `borderMedium` | `rgba(55,53,47,0.16)` | ボーダー（中間） |
| `hoverLight` | `rgba(55,53,47,0.04)` | ホバーオーバーレイ（薄い） |
| `hoverMedium` | `rgba(55,53,47,0.08)` | ホバーオーバーレイ（中間） |
| `hoverStrong` | `rgba(55,53,47,0.35)` | スクロールバーホバー |
| `accent` | `#2383E2` | Notionブルーアクセント |
| `accentHover` | `#1B6EC2` | アクセントホバー |
| `accentLight` | `#D3E5EF` | アクセント薄い背景（ステータスピル） |
| `accentDark` | `#183B56` | アクセント濃いテキスト |
| `danger` | `#EB5757` | 破壊的/危険（CSS変数 `--destructive`） |
| `notionRed` | `#E03E3E` | 必須マーカー、バリデーション |
| `redIcon` | `#EA3323` | アラートサークルアイコン |
| `muted` | `#787774` | ミュートテキスト |
| `mutedBg` | `#F1F0EE` | ミュート背景 |
| `statusGreenText` | `#0F7B6C` | ステータス意味色（グリーンテキスト） |
| `statusGreenBg` | `#DDEDEA` | ステータス意味色（グリーン背景） |
| `grayMedium` | `#9B9A97` | グレー中間色（statusGrayText/dotGray の正規名） |
| `statusGrayBg` | `#EBECED` | ステータス意味色（グレー背景） |
| `grayLight` | `#E3E2E0` | グレー薄色（bgInactive/grayTagBg の正規名） |
| `costBlueBg` / `costBlueText` | `#E3F2FD` / `#1565C0` | コスト集計ブルー（Material palette） |
| `costGreenBg` / `costGreenText` | `#E8F5E9` / `#2E7D32` | コスト集計グリーン（Material palette） |
| `successGreen` / `successGreenHover` | `#10B981` / `#059669` | 成功/会計リンク（Emerald green） |
| `noticeBg` / `noticeBorder` / `noticeText` | `#FDECC8` / `#F2DBA7` / `#C29243` | Notion Yellow（注意・アラート） |
| `notionOrange` / `notionOrangeLight` | `#D9730D` / `#FAEBDD` | Notion Orange（値引き・財務） |
| `statusPurpleText` / `statusPurpleBg` | `#6940A5` / `#EEE0F7` | Notion Purple（診療中ステータス） |
| `brownBg` / `brownText` / `brownBorder` | `#EEE0DA` / `#64473A` / `#DFD0C8` | Notion Brown（入院予約カード） |
| `pinkBg` / `pinkText` / `pinkBorder` | `#F5E0E9` / `#AD1A72` / `#ECCBDA` | Notion Pink（ホテル予約カード） |
| `warningBg` / `warningIcon` / `warningText` | `#FFF3CD` / `#B58105` / `#856404` | 警告（イエロー） |
| `dot*` (10色) | `#529CCA` 等 | 診療内容凡例ドットカラー |

> 全エントリは `/lib/design-tokens.ts` の `PALETTE` オブジェクトを参照。

#### 2. `C` — Tailwindカラークラストークン

単一目的のクラストークンで、`className` 文字列の組み立てに使用します。

```tsx
import { C } from "../lib/design-tokens";
// 使用例: className={`${C.text} ${C.bgPage}`}
```

| カテゴリ | 例 | 説明 |
|---|---|---|
| **テキスト** | `C.text`, `C.text80`, `C.text70`, `C.text60`, `C.text50`, `C.text40`, `C.text35`, `C.text30`, `C.text25`, `C.text20`, `C.text15` | プライマリ〜極薄テキスト（11段階） |
| **テキスト（特殊）** | `C.textPlaceholder`, `C.textPlaceholderFaint` | プレースホルダー（30%, 15%） |
| **背景** | `C.bgPage`, `C.bgPageHalf`, `C.bgPage30`, `C.bgWhite`, `C.bgSubtle`, `C.bgActive`, `C.bgHover`, `C.bgHoverMd`, `C.bgPrimary`, `C.bgPrimary10`, `C.bgPrimary5` | ページ、カード、ホバー、プライマリ背景 |
| **ボーダー** | `C.borderLight`, `C.borderMediumLight`, `C.borderMedium`, `C.borderDivider`, `C.borderPrimary`, `C.borderPrimary20`, `C.borderPrimary10`, `C.borderLPrimary` | 4段階のボーダー + プライマリ系 |
| **ディバイダー** | `C.divideDivider`, `C.divideDividerFaint` | テーブル/リスト行間のdivide色 |
| **アクセント** | `C.accent`, `C.bgAccent`, `C.bgAccentHover`, `C.bgAccentLight`, `C.bgAccent5`, `C.bgAccent8`, `C.textAccentDark`, `C.borderAccent`, `C.ringAccent`, `C.ringAccent40`, `C.hoverBgAccent5`, `C.focusRingAccent` | フォーカス、選択、リンク |
| **破壊的** | `C.danger`, `C.bgDanger`, `C.hoverTextDanger`, `C.hoverBgDanger5` | 削除、エラー |
| **Notionレッド** | `C.textRequired`, `C.textRedIcon` | 必須マーカー、アラートアイコン |
| **ステータス** | `C.textStatusGreen`, `C.bgStatusGreen`, `C.borderStatusGreen`, `C.hoverBgStatusGreen`, `C.textStatusGray`, `C.bgStatusGray`, `C.borderStatusGray`, `C.hoverBgStatusGray`, `C.bgInactive` | 意味的ステータス色（グリーン/グレー/無効） |
| **コスト** | `C.bgCostBlue`, `C.textCostBlue`, `C.bgCostGreen`, `C.textCostGreen` | コスト集計カラー |
| **警告** | `C.bgWarning50`, `C.textWarningIcon`, `C.textWarning` | 警告カラー（イエロー系） |
| **成功** | `C.textSuccess`, `C.bgSuccess`, `C.bgSuccessHover`, `C.borderSuccess30`, `C.hoverBgSuccess10` | 会計リンク等（Emerald green） |
| **注意** | `C.textNotice`, `C.bgNotice`, `C.borderNotice`, `C.borderNotice50`, `C.bgNotice40` | Notion Yellow（アラート・依頼中） |
| **Notionレッド（バッジ）** | `C.textNotionRed`, `C.bgRedLight`, `C.borderRed30`, `C.bgNotionRed`, `C.borderRedBadge` | 赤系バッジ・ステータス |
| **Notionオレンジ** | `C.bgDiscount`, `C.bgDiscountHover`, `C.textDiscount`, `C.bgDiscountLight`, `C.borderDiscount20` | 値引き・財務オレンジ |
| **Notionパープル** | `C.textStatusPurple`, `C.bgStatusPurple`, `C.bgStatusPurpleDot`, `C.borderStatusPurple`, `C.hoverBgStatusPurple`, `C.hoverBgStatusPurpleDark` | 診療中ステータス |
| **Brown/Pink** | `C.bgBrown`, `C.textBrown`, `C.borderBrown`, `C.bgPink`, `C.textPink`, `C.borderPink` | 予約カード（入院/ホテル） |
| **ミュート（バッジ）** | `C.bgMutedBadge`, `C.textMuted`, `C.borderMuted` | デフォルトバッジ |
| **ダッシュボードカンバン** | `C.bgAccentLight50`, `C.textAccentDark60`, `C.hoverBgAccentBadge40` 等（16トークン） | カンバンカラムの透明度バリエーション |
| **ドットカラー** | `C.dotBlue`, `C.dotGreen`, `C.dotRed`, `C.dotOrange`, `C.dotYellow`, `C.dotPurple`, `C.dotPink`, `C.dotBrown`, `C.dotGray`, `C.dotDefault` | 診療内容凡例ドット |
| **ホバー** | `C.hoverBgPage`, `C.hoverBgPageHalf`, `C.hoverBgLight`, `C.hoverBgMedium`, `C.hoverBgPrimary10`, `C.hoverText`, `C.hoverText60`, `C.hoverBorderPrimary30`, `C.hoverBorderPrimary40`, `C.hoverBorderMedium`, `C.hoverBgSubtle` | ホバー状態 |
| **フォーカス** | `C.focusBgLight`, `C.focusRingAccent`, `C.focusRingMedium` | フォーカス状態 |
| **リング** | `C.ringPrimary50`, `C.ringPrimary40`, `C.ringPrimary20` | リング色 |

> 合計約190トークン。全エントリは `/lib/design-tokens.ts` の `C` オブジェクトを参照。

#### 2b. `BADGE` — バッジカラーコンボ

再利用可能な「bg + text + border」クラス文字列で、ステータスバッジ/ピルに使用します。`C` トークンを組み合わせた合成プリセットです。

```tsx
import { BADGE } from "../lib/design-tokens";
// 使用例: className={`${BADGE.blue} ${STYLE.badge}`}
```

| キー | 用途 |
|------|------|
| `BADGE.blue` | 作成中、入院中、受付済、検査中、予約(trimming)、スタッフ |
| `BADGE.gray` | 確定済、退院済、cancelled、inactive |
| `BADGE.green` | 完了、会計済、sufficient、active |
| `BADGE.red` | 入院(type)、waiting(acct)、out_of_stock、手術 |
| `BADGE.purple` | ホテル(type)、treatment、トリマー、予防接種 |
| `BADGE.orange` | 進行中(trimming)、food |
| `BADGE.yellow` | 依頼中、pending、low stock、定期健診 |
| `BADGE.brown` / `BADGE.pink` | 入院予約カード / ホテル予約カード |
| `BADGE.muted` | デフォルト/フォールバック |
| `BADGE.*NoBorder` | ケアプラン用（ボーダーなし5色） |
| `BADGE.*Hover` | ペットステータス用（ホバー付き） |

> バッジカラーは「カルテ＝青(accent)、会計＝緑(statusGreen)、入院＝紫(statusPurple)」で統一。

#### 3. `LAYOUT` — レイアウト定数

アニメーションターゲット、style props、幅/高さ制約のTailwindクラス文字列を提供します。

| キー | 値 | 用途 |
|------|-----|------|
| `sidebar.expanded/collapsed` | `w-[220px]` / `w-[56px]` | サイドバー幅 |
| `sidebar.expandedPx/collapsedPx` | `220` / `56` | アニメーション用px値 |
| `sidePeek.width` | `w-[520px]` | サイドピーク（マスタ編集）幅 |
| `propertyRow.minH` | `min-h-[40px]` | Notionプロパティ行 |
| `propertyRow.labelW` | `w-[140px]` | ラベル列幅 |
| `header.h` | `h-12` | フォームヘッダー高さ |
| `touch.md` | `h-10` | タッチターゲット（プライマリ） |
| `touch.sm` | `h-10` | タッチターゲット（セカンダリ） |
| `touch.row` | `h-12` | テーブル行高さ |
| `touch.tableHead` | `h-10` | テーブルヘッダー高さ |
| `touch.iconBtn` | `size-9` | ツールバーアイコンボタン |
| `touch.badge` | `h-8` | ステータスバッジ高さ |
| `pageTitle.fontSize/fontWeight` | `30px` / `700` | ページタイトルスタイル |
| `pageIcon.size` | `size-[38px]` | Notionページアイコンコンテナ |
| `pageIcon.innerIcon` | `size-[20px]` | ページアイコン内部サイズ |
| `modal.sm` / `md` / `lg` / `xl` / `full` | `sm:max-w-[480px]` 〜 `w-[98%]` | モーダルサイズ（5段階） |

#### 4. `STYLE` — コンポジットスタイルプリセット

再利用可能なclassNameストリングで、頻出するUIパターンを集約します。

| プリセット | 用途 |
|---|---|
| `STYLE.page` / `STYLE.pageContent` / `STYLE.sectionDivider` | ページ全体のレイアウト |
| `STYLE.formHeader` / `STYLE.formHeaderTitle` / `STYLE.formHeaderDesc` | フォームヘッダー（sticky） |
| `STYLE.btnPrimary` / `STYLE.btnGhost` / `STYLE.btnAccent` / `STYLE.btnDanger` / `STYLE.btnOutline` | ボタンバリアント（5種） |
| `STYLE.tableContainer` / `STYLE.tableHeaderRow` / `STYLE.tableHeaderCell` / `STYLE.tableRow` / `STYLE.tableCell` / `STYLE.tableCellMono` / `STYLE.tableCellMuted` / `STYLE.tableEmpty` / `STYLE.tableActionBtn` | テーブルスタイル |
| `STYLE.searchInput` / `STYLE.searchIcon` / `STYLE.searchCount` | 検索フィルタバー |
| `STYLE.paginationBtn` / `STYLE.paginationBtnActive` / `STYLE.paginationInfo` | ページネーション |
| `STYLE.sidebarContainer` / `STYLE.sidebarHeader` / `STYLE.sidebarItemActive` / `STYLE.sidebarItemIdle` / `STYLE.sidebarToggle` | サイドバー |
| `STYLE.propertyRow` / `STYLE.propertyLabel` / `STYLE.propertyInput` | Notionプロパティ行 |
| `STYLE.sidePeekPanel` / `STYLE.sidePeekToolbar` / `STYLE.sidePeekToolbarBtn` / `STYLE.sidePeekBody` / `STYLE.sidePeekFooter` / `STYLE.sidePeekCancelBtn` / `STYLE.sidePeekSaveBtn` | サイドピーク（マスタ編集） |
| `STYLE.pageIcon` | Notionページアイコン |
| `STYLE.selectCompact` | サイドピーク用コンパクトSelect |
| `STYLE.sectionLabel` | セクション見出し（uppercase） |
| `STYLE.formLabel` / `STYLE.formInput` / `STYLE.formCard` | フォームコントロール |
| `STYLE.badge` | ステータスバッジ |
| `STYLE.dropdownShadow` | Notionスタイルドロップダウンシャドウ |
| `STYLE.confirmPrimary` | 確認ダイアログプライマリボタン |
| `STYLE.settingsRow` / `STYLE.settingsRowIcon` | マスタ設定インデックス行 |
| `STYLE.inlineAddBtn` | インライン追加行 |
| `STYLE.formInputLight` | フォームInput（薄いボーダー + ホバー） |

> 合計約53プリセット。全エントリは `/lib/design-tokens.ts` の `STYLE` オブジェクトを参照。

### Notion風プロパティ行パターン

全Notion風フォーム（`MasterItemEditForm`・`ClinicSettings`・マスタセクション）で使用される共通UIパターンです。`/components/shared/NotionPropertyRow.tsx` に3つの共通コンポーネントが定義されています。

#### NotionPropertyRow

ラベル＋値のキー・バリュー行。ホバー時に薄いハイライト背景が表示されます。

```tsx
import { NotionPropertyRow } from "../components/shared/NotionPropertyRow";

<NotionPropertyRow label="病院名" required>
  <input className={STYLE.propertyInput} {...register("name")} placeholder="空" />
</NotionPropertyRow>
```

| Prop | 型 | デフォルト | 説明 |
|---|---|---|---|
| `label` | `string` | — | 140px幅の左ラベル |
| `children` | `ReactNode` | — | 右側の入力要素 |
| `required` | `boolean` | `false` | 必須マーカー（`*`）表示 |
| `align` | `"start" \| "center"` | `"center"` | 垂直方向の配置。テキストエリアやSelectは `"start"` |

**スタイルトークン:**
- 行: `STYLE.propertyRow` — `flex gap-2 py-2 px-2 -mx-2 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] min-h-[40px]`
- ラベル: `STYLE.propertyLabel` — `w-[140px] shrink-0 text-sm text-[rgba(55,53,47,0.65)]`
- 入力: `STYLE.propertyInput` — 透過背景・ホバー/フォーカスで薄いハイライト
- セレクト: `STYLE.selectCompact` — ボーダーなし・コンパクトHeight

#### NotionSectionLabel

セクション見出し（薄字・小文字）。

```tsx
<NotionSectionLabel>所在地</NotionSectionLabel>
```

#### NotionSectionDivider

セクション間の薄罫線。`className` propでマージンカスタマイズ可能（デフォルト `my-3`）。

```tsx
<NotionSectionDivider />
<NotionSectionDivider className="my-4" />
```

#### SectionPropertyRow（マスタセクション用 re-export）

`/features/master/components/sections/SectionWrapper.tsx` から export される `SectionPropertyRow` は `NotionPropertyRow` の re-export エイリアスです。マスタの10個のセクションコンポーネントは既存の import パスを維持しつつ、内部では共通コンポーネントを使用しています。

#### 適用範囲

| フォーム種別 | UIパターン | 使用コンポーネント |
|---|---|---|
| **MasterItemEditForm** (サイドピーク) | Notionプロパティ行 | `NotionPropertyRow` |
| **ClinicSettings** (病院情報設定) | Notionプロパティ行 + セクションラベル/ディバイダー | `NotionPropertyRow` + `NotionSectionLabel` + `NotionSectionDivider` |
| **マスタセクション** (10個) | Notionプロパティ行 | `SectionPropertyRow` (= `NotionPropertyRow`) |
| **OwnerForm / PetEditModal** | グリッドベースLabel+Input | `Label` + `Input` (従来パターン) |
| **TrimmingForm / InventoryForm** | カード/グリッドベースLabel+Input | `Label` + `Input` / `STYLE.formLabel` + `STYLE.formInput` |
| **ReservationFormFields** | カスタムFieldLabel | `FieldLabel` (ローカル) |
| **ExaminationForm / VaccinationForm** | カード内Label+Input | `Label` + `Input` (医療記録インライン) |
| **HospitalizationBasicInfo** | コンパクトカード | `Label` + `Input` (入院コンパクト) |

> **設計方針**: Notionページスタイルのフォーム（サイドピーク・設定画面）は `NotionPropertyRow` パターンを使用。グリッドレイアウトやカード内のインラインフォームは従来の `Label` + `Input` パターンを維持。

#### 5. `TABLE_STYLES` — 後方互換エクスポート

既存コンシューマ向けの後方互換エクスポートで、内部的には `STYLE` からの再エクスポートです。

```typescript
export const TABLE_STYLES = {
  row:          STYLE.tableRow,
  actionButton: STYLE.tableActionBtn,
  cell:         STYLE.tableCell,
  cellMono:     STYLE.tableCellMono,
};
```

---

## ネイティブSVGスパークライン (`MiniSparkline`)

`DashboardSummaryWidget` で使用される軽量なスパークラインチャートは、Rechartsではなく**ネイティブSVG要素**で描画されています。これにより、バンドルサイズの削減とレンダリングパフォーマンスの向上を実現しています。

### 描画仕様

| 要素 | 実装 | 説明 |
|------|------|------|
| **折れ線** | `<polyline>` | `stroke` のみ（`fill="none"`）。データポイントを直線で結ぶ |
| **アクティブドット** | `<circle>` | ホバー中のデータポイントにのみ表示。CSSトランジションで出現 |
| **ツールチップ** | `<foreignObject>` または CSS positioned `<div>` | ホバーポイントの上部にデータ値を表示 |
| **グラデーション塗りつぶし** | `<linearGradient>` + `<polygon>` | 折れ線の下を薄いグラデーションで塗りつぶし（オプション） |

### カラー

- **上昇トレンド**: `#0F7B6C`（`PALETTE.statusGreenText`）
- **下降トレンド**: `#E03E3E`（`PALETTE.notionRed`）
- **ニュートラル**: `#787774`（`PALETTE.muted`）

### 適用方針

- `recharts` は `/features/` 配下で一切使用しない
- `/components/ui/chart.tsx` 内にshadcn/uiのRecharts統合が残存するが、どのfeatureからも参照されていない
- 新規チャートを追加する場合は、ネイティブSVGまたはCanvasで実装すること

---

## カスタムドラッグプレビュー (`CageDragPreview`)

入院管理ボード（`HospitalizationBoard`）でケージ間のペット移動を行う際に、`react-dnd` の `useDragLayer` を使用してカスタムドラッグプレビューオーバーレイを描画します。

### ビジュアル仕様

| 属性 | 値 | 説明 |
|------|-----|------|
| **位置** | `position: fixed` | ビューポート固定でカーソルに追従 |
| **不透明度** | `opacity: 0.85` | 半透明でドラッグ元の視認性を維持 |
| **影** | `shadow-lg` | 浮遊感を演出 |
| **変形** | `rotate(-2deg)` | 若干の傾きでドラッグ中であることを視覚的に強調 |
| **z-index** | `z-50` | 全要素の最前面に表示 |
| **ポインターイベント** | `pointer-events: none` | ドラッグ中のホバー判定に干渉しない |

### `useReducedMotion` 連携

`prefers-reduced-motion: reduce` が有効な場合、`rotate` および `scale` 効果を無効化し、不透明度変更のみに簡略化します。

---

## ログインページ (`/login`)

サイドバー・レイアウトの外側にレンダリングされるフルスクリーンログインページ。認証済みの場合はダッシュボード (`/`) にリダイレクトします。

### レイアウト

| 属性 | 値 | 説明 |
|------|-----|------|
| **コンテナ** | `min-h-screen flex items-center justify-center` + `C.bgPage` | 背景 `#F7F6F3`、垂直水平中央 |
| **フォーム幅** | `max-w-[400px]` | コンパクトなログインカード |
| **パディング** | `p-4` | モバイル対応のアウターパディング |

### ヘッダー

| 要素 | スタイル | 説明 |
|------|---------|------|
| **アイコンコンテナ** | `size-14 rounded-xl` + `C.bgPrimary` + `text-white` | `#37352F` 背景の丸角アイコン |
| **アイコン** | `LogIn` (`size-7`) | lucide-react の LogIn アイコン |
| **タイトル** | `text-2xl` + `C.text` + `fontWeight: 700` | 「ログイン」 |
| **サブテキスト** | `text-sm` + `C.text50` | 「動物病院管理システムにログイン」 |

### フォームフィールド

| フィールド | 仕様 | 説明 |
|---|---|---|
| **メールアドレス** | `STYLE.formLabel` + `STYLE.formInput` + `type="email"` + `autoComplete="email"` | 標準フォーム入力 |
| **パスワード** | `STYLE.formLabel` + `STYLE.formInput` + `type="password"` / `type="text"` | パスワード表示トグル付き |
| **パスワード表示トグル** | `Eye` / `EyeOff` アイコン (`size-4`) + 右端配置 (`absolute right-2`) | `aria-label` でアクセシビリティ対応 |
| **送信ボタン** | `STYLE.btnPrimary` + `w-full` | 「ログイン」/「ログイン中...」（disabled） |

### バリデーション・エラー

| 仕様 | 詳細 |
|------|------|
| **入力検証** | メールアドレス・パスワード空欄チェック（フロントエンド） |
| **認証エラー** | API から返されるエラーメッセージを `FormFieldError` で表示 |
| **`aria-invalid`** | エラー時に `true` を設定 |
| **`aria-describedby`** | エラー時のみ `"login-error"` を条件適用 |
| **`role="alert"`** | `FormFieldError` 内蔵（`aria-live="assertive"`） |

### デモアカウントパネル

ログインフォームの下に表示される開発用アカウント一覧。

| 要素 | スタイル | 説明 |
|------|---------|------|
| **区切り線** | `h-px` + `C.bgPrimary10` | 左右に伸びる薄い罫線 |
| **ラベル** | `text-xs` + `C.text40` | 「デモアカウント（パスワード: password）」 |
| **アカウント行** | `C.hoverBgLight` + `rounded-md` + `px-3 py-2` | ホバーで薄いハイライト |
| **表示名** | `text-sm` + `C.text` | ユーザー名 |
| **ロール** | `text-xs` + `C.text50` | `USER_TYPE_LABELS` / `JOB_TITLE_LABELS` から取得 |
| **メール** | `text-xs` + `C.text30` | ホバーで `C.text50` に変化 |

クリックでメールアドレスとパスワード（`"password"`）を自動入力します。

### モックアカウント一覧

| メール | 表示名 | ユーザー種別 | 職種 | 所属クリニック |
|--------|--------|------------|------|--------------|
| `admin@example.com` | 田中 太郎 | `clinic_admin` | 医師 | 八王子院 + 新宿院 |
| `vet@example.com` | 山田 花子 | `staff` | 医師 | 八王子院 |
| `nurse@example.com` | 佐藤 美咲 | `staff` | 看護師 | 八王子院 |
| `reception@example.com` | 鈴木 一郎 | `staff` | 受付 | 八王子院 |
| `trimmer@example.com` | 高橋 さくら | `staff` | トリマー | 八王子院 |
| `system@example.com` | 本部 管理者 | `system_admin` | — | 八王子院 + 新宿院 |

### 関連コンポーネント

| コンポーネント | パス | 役割 |
|---|---|---|
| **Login** (ルートページ) | `/features/auth/routes/Login.tsx` | 認証状態チェック + リダイレクト + `LoginForm` レンダリング |
| **LoginForm** | `/features/auth/components/LoginForm.tsx` | フォームUI + バリデーション + デモアカウントパネル |
| **ClinicSwitcher** | `/features/auth/components/ClinicSwitcher.tsx` | サイドバーヘッダーのクリニック切替ドロップダウン |
| **ProtectedRoute** | `/features/auth/components/ProtectedRoute.tsx` | ルートレベル認証 + 権限ガード |
| **PermissionGate** | `/features/auth/components/PermissionGate.tsx` | コンポーネントレベル権限ゲート |

---

## サイドバー認証連携

### ユーザー情報フッター

サイドバー下部に認証ユーザー情報とログアウトボタンを表示します。

| 要素 | スタイル | 説明 |
|------|---------|------|
| **アバター** | `size-8 rounded-full` + `C.bgActive` | ユーザーアイコン（`User` lucide） |
| **表示名** | `text-sm` + `C.text` + `truncate` | `user.displayName` |
| **ロール** | `text-xs` + `C.text40` + `truncate` | ユーザー種別 / 職種 |
| **ログアウト** | `LogOut` アイコン + ghost button | `aria-label="ログアウト"` |
| **折りたたみ時** | ログアウトボタンのみ中央配置 | 名前・ロールは非表示 |

### メニュー権限フィルタリング

サイドバーのメニュー項目は `SidebarMenuEntry.requiredPermissions` に基づきフィルタリングされます。

| 条件 | 表示 |
|------|------|
| `system_admin` / `clinic_admin` | 全メニュー表示 |
| `requiredPermissions` 未設定 | 常時表示（ダッシュボード、シフト管理） |
| `requiredPermissions` 設定あり | `hasAnyPermission()` で判定 |

### クリニック切替

複数クリニックに所属するユーザーのみ `ClinicSwitcher` ドロップダウンを表示します。

| 条件 | UI |
|------|-----|
| **単一クリニック** | クリニック名 + ブランチ名（静的テキスト） |
| **複数クリニック** | ブランチ名 + `ChevronDown` アイコン → ドロップダウンメニュー |
| **現在のクリニック** | `Check` アイコンで示す |

---

## 意図的保持カラー（Notionトークン非適用箇所）

本システムはすべての汎用色を `/lib/design-tokens.ts` の `C.*` / `PALETTE.*` トークンに統一していますが、下記の2コンポーネントは**医療・時間帯という意味的文脈**に基づき、Tailwind標準色クラスをそのまま使用します。これらは不統一ではなく、仕様書明示の意図的保持対象です。

> **レビュー指針**: `C.*` への置換提案が発生した場合は「意図的保持」として却下してください。
> 今後同様の医療慣習色・時間帯テーマ色を追加する場合は、このセクションに明記してください。

### VitalInputDialog（バイタルサイン医療慣習色）

**ファイル**: `features/medical-records/components/VitalInputDialog.tsx`

医療現場では体温=オレンジ・心拍=レッド・呼吸=ブルー・体重=グリーンという配色慣習が国際的に定着しており、スタッフの視覚的識別を支援するために保持します。

| バイタルサイン | 入力ラベルアイコン | 前回値テキスト | 履歴バッジ |
|---|---|---|---|
| **体温 (℃)** | `text-orange-500` | `text-orange-600/70` | `bg-orange-50 text-orange-700 border-orange-200` |
| **心拍数 (bpm)** | `text-red-500` | `text-red-600/70` | `bg-red-50 text-red-700 border-red-200` |
| **呼吸数 (回/分)** | `text-blue-500` | `text-blue-600/70` | `bg-blue-50 text-blue-700 border-blue-200` |
| **体重 (kg)** | `text-green-500` | `text-green-600/70` | `bg-green-50 text-green-700 border-green-200` |

**TrendIcon（値の増減トレンド表示）**:

| 方向 | クラス | 説明 |
|------|--------|------|
| **増加** | `text-red-400` | 体温・心拍数の上昇は生理的危険方向 |
| **減少** | `text-blue-400` | 冷却・低下トレンドを寒色で視覚化 |

> **設計根拠**: 体温上昇→発熱リスクのため赤系、呼吸数・心拍数減少→ショックリスクのため青系という医療慣習に準拠。Notionトークンのグリーン/グレーで代替すると意味が破壊される。

### DailyRecordSection（時帯テーマ色）

**ファイル**: `features/hospitalization/components/DailyRecord/DailyRecordSection.tsx`

入院管理の日課記録において、朝・昼・夜の時間帯を直感的に識別するための色分けです。

| 時間帯 | アイコン | テーマ色クラス | 根拠 |
|--------|---------|--------------|------|
| **朝 (morning)** | `Sun` | `text-orange-600` | 朝日・日の出の暖色 |
| **昼 (noon)** | `Coffee` | `text-yellow-600` | 昼の光・明るさを表す黄色 |
| **夜 (night)** | `Moon` | `text-indigo-600` | 夜空・月明かりを表すインディゴ |

> **設計根拠**: 時間帯に対応する自然な色であり、Notionのインディゴ/オレンジ/イエロートークンは存在しないため、Tailwindの意味的色名クラスで対応。入院スタッフが1日3サイクルの投薬・食事・ケアを素早く識別するための重要な視覚補助。

---
