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

#### 予約カレンダーカラー（動的カラーシステム）

**デザイン方針**: 各予約区分（またはグループ）に設定されたベースカラーから、背景色（透過率10%）、ボーダー色（透過率30%）、テキスト色（ベースカラーそのまま）を動的に生成します。これにより、マスタ設定で自由な色を選択しつつ、一貫したパステル調の視覚効果を維持します。

**実装場所**:
- `/features/master/hooks/use-reservation-type-color-map.ts`: `useReservationTypeColorMap` フックによる動的スタイル生成
- `/lib/design-tokens.ts`: `PALETTE.pickerDefaultBlue` (デフォルト色)
- 予約カレンダー（`/features/reservations/routes/ReservationManagement.tsx`）およびダッシュボードで使用

**色の解決ルール**:
1. **グループ色優先**: 予約区分が属する `ReservationTypeGroup.color` が設定されていればそれを使用。
2. **区分個別色**: グループ未設定（またはグループに色がない）場合は `ReservationType.color` を使用。
3. **デフォルト**: いずれも未設定の場合は `PALETTE.grayMedium` を使用。

**使用例**:
```tsx
const { getColor } = useReservationTypeColorMap();
const color = getColor(appointment.reservationType.name);

// 予約カードのスタイル
<div style={color.style}>
  {appointment.petName}
</div>

// 凡例のドット
<span style={color.dotStyle} className="w-2.5 h-2.5 rounded-full" />
```

---

## 実装詳細

タイポグラフィ、余白、コンポーネントスタイリング、フォーム、アイコン、アクセシビリティ等の実装ルールは以下を参照:

- **Single Source of Truth**: `frontend/src/lib/design-tokens.ts`（`C`, `STYLE`, `PALETTE` 定数）
- **コーディング規約**: `.claude/rules/typescript-react.md`（Design Tokens セクション）
- **アクセシビリティ**: `.claude/rules/accessibility-rules.md`

---

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
