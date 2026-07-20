# UI デザインシステム 規約 (Design System)

> **目的**: [DESIGN.md](../../DESIGN.md)（ルート — Notion Analysis / 意匠言語）を Animal Ekarte の実装に落とし込むための規約を定義する。
> **読者**: フロントエンド実装者。
> **タイミング**: UI 実装・レビュー時。
> **最新更新**: 2026-07-21（**FE10: DESIGN.md 字義遵守リブランド** — 製品上書き（teal 等）を撤回し brand=`#0075DE` へ反転。唯一の存続例外は §2.4 臨床 semantic 色）

### SSOT 優先順位

| 層 | ファイル | 役割 |
|---|---|---|
| 1 | [DESIGN.md](../../DESIGN.md) | 意匠言語の**正本**。トークン値・タイポ/レイアウト/コンポーネント定義に**字義で従う**（FE10 決裁・2026-07-21 曽我）。 |
| 2 | **本書** (`docs/spec/design-system.md`) | DESIGN.md の実装規約化。逸脱は §2.4 臨床 semantic 色（唯一の例外）のみ。 |
| 3 | `frontend/src/lib/design-tokens.ts` | 実行時 SSOT（`C` / `STYLE` / `PALETTE` / `LAYOUT`）。構造色は `#0075DE` に同期済み — §2.6 参照。 |

> **Animal Ekarte**: Notion ライクな体験を臨床現場へ。構造色は DESIGN.md `{colors.primary}` 字義値 **`#0075DE`（Notion blue）**。旧製品上書き `#038B94`（teal）は FE10 で撤回・legacy 化（audit C1 が再導入を禁止）。

---

## 1. Overview / デザイン思想

本システムの UI は、**「臨床現場での集中力を妨げないクリーンな情報設計」**を最優先としている。DESIGN.md の言葉を借りれば、「良い光の下の整理されたデスク」— 暖色の紙のようなキャンバス（`{colors.canvas-soft}`）の上に、抑制されたモノクロ＋**単一構造色（brand `#0075DE`・Notion blue）** の UI が乗り、装飾は必要な場所（スティッカーパレット・意味的カラー）にのみ許可される。

### 核心となる原則

- **ミニマリズム**: 装飾を削ぎ落とし、データそのものを主役にする。構造色は `{colors.primary}`（`#0075DE`）一色のみとし、二番目の構造アクセントを導入しない。
- **直感的な階層**: Notion スタイルの「プロパティ」と「カード」による情報の整理。カードは白 surface + hairline、ページ canvas は暖色オフホワイト。
- **臨床安全の可視化**: 警告や異常値のみを、戦略的に配色された「意味的カラー」で強調する。危険バッジ・死亡グレーアウト・RBAC 非活性表示など、臨床安全 UI は本デザイン変更で退行させない。

### Key Characteristics（DESIGN.md 字義準拠）

- 暖色 paper-soft canvas `{colors.canvas-soft}` — 純白ページではなく document-like な落ち着き
- 近黒 `{colors.ink}` の Inter 系タイポ。display サイズではネガティブ tracking
- **単一構造アクセント** — `{colors.primary}` = **`#0075DE`（Notion blue）** — Primary CTA・リンク・active/focus のみ
- 装飾専用マルチカラースティッカーパレット（purple / pink / orange / teal / green / sky）— 構造には使わない
- Primary CTA は pill `{rounded.full}`、ユーティリティボタンは `{rounded.md}`（8px）— 意図的な対比
- hairline `#E6E6E6` + 極薄レイヤードシャドウによる elevation（重い drop-shadow 禁止）
- 深 indigo `{colors.secondary}` の hero-band はマーケ向けパターン。AE アプリ本体では臨床 UI に合わせて限定的に使用

---

## 2. Colors / カラーパレット

> トークン名・色値とも [DESIGN.md](../../DESIGN.md) `colors:` フロントマターが正（FE10 字義遵守）。

### 2.1 Brand & Accent（構造色）

| トークン | 値（DESIGN.md 字義） | 用途 |
|---|---|---|
| `{colors.primary}` | **`#0075DE`**（Notion blue） | 唯一の構造アクセント。Primary CTA fill、インラインリンク、active-tab / focus ring。**装飾には使わない。** |
| `{colors.primary-active}` | **`#005BAB`** | Primary CTA の押下 / hover 状態。 |
| `{colors.secondary}` | `#213183` | 深 indigo hero-band（AE では限定的） |
| `{colors.on-primary}` | `#FFFFFF` | primary 上のテキスト |

> **旧製品上書きの撤回（FE10・2026-07-21）**: 旧 brand teal `#038B94` / pressed `#027078` は legacy 化し、audit C1 が再導入を機械的に禁止する。旧 legacy accent `#2383E2` 系トークン（`C.bgAccent` 等）の**値**は brand `#0075DE` に統合済み（トークン名は互換のため残存 — 新規実装は brand 系を使うこと）。

**装飾スティッカーパレット**（CTA・構造フィル禁止 — §2.5 参照）:
Sky `#62AEF0`、Purple `#D6B6F6` / Deep `#391C57`、Pink `#FF64C8`、Orange `#DD5B00` / Deep `#793400`、Teal `#2A9D99`、Green `#1AAE39`、Brown `#523410`

### 2.2 Surface

| トークン | 値 | 用途 |
|---|---|---|
| `{colors.canvas}` / `{colors.surface}` | `#FFFFFF` | カード・パネル・ナビ・フォームフィールド |
| `{colors.canvas-soft}` | `#F6F5F4` | ページ canvas・フッター帯 — document-like な暖色オフホワイト |
| `{colors.hairline}` | `#E6E6E6` | 1px カード境界・区切り線 |

### 2.3 Text

| トークン | 値 | 用途 |
|---|---|---|
| `{colors.ink}` | `#000000` | 見出し・本文（~95% alpha で soft true-black） |
| `{colors.ink-secondary}` | `#31302E` | 二次本文・フッターリンク |
| `{colors.ink-muted}` | `#615D59` | 補助・muted コピー |
| `{colors.ink-faint}` | `#A39E98` | キャプション・メタデータ・placeholder |

### 2.4 意味的カラー (Semantic Colors) — 唯一の DESIGN.md 逸脱（存続決定・FE10）

DESIGN.md の Semantic 節は「Notion の*マーケ表面*は専用 semantic ramp を持たない」という観察であり、削除の指令ではない。臨床安全（SPECIFICATION 2.1 — 全原則に優先）により、本システムは意味的カラーを **FE10 字義リブランド後も維持する**。`design-tokens.ts` で一元管理し、構造色用途（CTA・リンク・active/focus）には使わない。

- **危険 / 死亡 / 異常高**: `C.danger`（`#C0392B`、WCAG AA 7.1:1 準拠）
- **注意 / 期限間近**: `C.textWarning` / `C.bgWarning50`
- **正常 / 完了 / 生存**: `C.textStatusGreen` / `C.bgStatusGreen`
- **異常低 / 寒冷**: `C.textStatusBlue` 系

> 危険バッジ・死亡グレーアウト・RBAC 非活性表示など臨床安全 UI は、デザイン変更で退行させない。

### 2.5 スティッカーパレット（装飾専用）

DESIGN.md Do's/Don'ts に従い、以下は **CTA・構造フィルには使用しない**。バッジ・タグ・凡例ドットなど装飾用途に限定する。

- 実装: `C.dotBlue` / `C.dotPurple` / `C.dotPink` / `C.dotOrange` / `C.dotGreen` / `C.dotBrown` など
- **外部ブランド例外**: LINE 公式グリーン `#06C755`（`PALETTE.lineGreen`）は外部ブランド識別のため構造色ルールの対象外。ただし構造・CTA には使わない。

### 2.6 `design-tokens.ts` マッピング表

| 役割 | DESIGN.md トークン | 字義値 | design-tokens.ts（FE10 反転済み） |
|---|---|---|---|
| 構造色（Primary CTA・リンク・active/focus） | `{colors.primary}` | **`#0075DE`** | `PALETTE.brand` / `C.bgBrand` / `C.textBrand` / `C.borderBrand` = `#0075DE` ✅ |
| 構造色 Pressed | `{colors.primary-active}` | **`#005BAB`** | `PALETTE.brandHover` = `#005BAB` ✅ |
| ページ canvas | `{colors.canvas-soft}` | `#F6F5F4` | `PALETTE.bgMain`, `C.bgPage` ✅ |
| カード surface | `{colors.canvas}` | `#FFFFFF` | `C.bgWhite` / `C.bgSubtle` ✅ |
| Hairline | `{colors.hairline}` | **`#E6E6E6`** | `C.borderLight` / `globals.css` `--border` = `#E6E6E6` ✅（旧 `rgba(0,0,0,0.09)` を字義固体値化） |
| 入力 border | `text-input` 1px `rgb(221,221,221)` | `#DDDDDD` | `globals.css` `--input` ✅ |
| Ink 系 | `{colors.ink}` 〜 `{colors.ink-faint}` | DESIGN.md 準拠 | `C.text` / `C.text70` / `C.text60` / `C.text40` ✅ |
| checked / focus | `{colors.primary}` | **`#0075DE`** | `C.dataCheckedBgBrand` / `--ring` / `--shadow-focus-brand` ✅ |
| CSS 変数 | `{colors.primary}` | **`#0075DE`** | `globals.css` `--primary` / `--sidebar-primary` = `#0075de` ✅ |

> **legacy トークンの扱い（FE10）**:
> - 旧 teal `#038B94`/`#027078` は全反転済み・audit C1 で再導入禁止。
> - `C.accent`/`C.bgAccent` 系トークンは**値を brand `#0075DE` に統合**して残存（名前互換のみ）。route 表面からは従来どおり排除（C1/C5）。**新規実装は必ず `brand` 系を使うこと**。

---

## 3. Typography / タイポグラフィ

DESIGN.md `typography:` フロントマターに準拠。実装のフォントファミリーは **`'Inter', 'Noto Sans JP'`**（フォールバック: `-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`。`frontend/src/styles/globals.css:143`）。単一スタックで display から eyebrow まで担う。

### 3.1 Hierarchy

| トークン | Size | Weight | Line Height | Letter Spacing | 用途 |
|---|---|---|---|---|---|
| `{typography.display-1}` | 64px | 700 | 1.0 | −2.125px | Hero 見出し |
| `{typography.display-2}` | 54px | 700 | 1.04 | −1.875px | 大セクション見出し |
| `{typography.heading-1}` | 40px | 700 | 1.1 | −1px | ページ見出し |
| `{typography.heading-2}` | 26px | 700 | 1.23 | −0.625px | サブセクション |
| `{typography.heading-3}` | 22px | 700 | 1.27 | −0.25px | カードタイトル |
| `{typography.title}` | 20px | 600 | 1.4 | −0.125px | フィーチャータイトル |
| `{typography.body-md}` | 16px | 400 | 1.5 | 0 | 標準本文 |
| `{typography.body-sm}` | 15px | 400 | 1.33 | 0 | テーブル行・密な UI（`ex-data-table-cell` body） |
| `{typography.button}` | 16px | 500 | 1.5 | 0 | ボタンラベル |
| `{typography.caption}` | 14px | 400 | 1.43 | 0 | キャプション・注記 |
| `{typography.eyebrow}` | 12px | 600 | 1.33 | +0.125px | データテーブルヘッダー・小ラベル（`ex-data-table-cell` header） |

### 3.2 Principles

- 見出しは weight 700 + サイズに応じたネガティブ tracking — display は特に tight
- 本文は 1.5 line-height で document 可読性を確保
- 表現の主レバーは **700 見出し vs 400 本文** のコントラスト。装飾タイポグラフィは使わない
- Inter 非利用環境（フォールバック時）では上表の letter-spacing を明示適用（デフォルト tracking より緩く見える）

### 3.3 タブレットファースト拡大

診察室 iPad 利用を想定し、Tailwind v4 `@theme inline` でクリックターゲット・フォントサイズを標準より約 **10% 拡大**（§7 参照）。

### 3.4 採用範囲と実装マッピング（FE10 字義化・2026-07-21）

アプリ本体 UI が日常的に使うのは `{typography.title}` 以下（display/heading 系は hero/LP 等の大面積見出しでのみ使用機会がある — 用途列が示すとおり DESIGN.md 自身の使い分け）。サイズは **DESIGN.md 字義値**に整列済み。

| ロール | AE 実装（Tailwind） | 実サイズ | 用途 |
|---|---|---|---|
| title | `text-xl font-semibold` | 20px | ページ内最上位見出し（PageLayout タイトル等） |
| section | `text-lg font-semibold` | 18px | セクション見出し（title と body の中間段） |
| body-md | `text-base` | 16px | 標準本文・フォームラベル |
| body-sm | `text-sm` | 15px（`--text-sm`） | テーブル行・密な UI の既定 |
| caption | `text-xs` | **14px**（`--text-xs` — DESIGN.md caption 字義値。FE10 で 13px 上書きを撤回） | キャプション・メタ・placeholder |
| eyebrow | `text-2xs font-semibold tracking-wide` | **12px**（DESIGN.md eyebrow 字義値） | データテーブルヘッダー（`STYLE.tableHeaderCell`）・小ラベル |
| micro | `text-2xs` | **12px**（`--text-2xs` — FE10 で 11px 製品拡張を撤回し eyebrow 段に統合） | バッジ内文字・極小メタ表示。乱用禁止 — caption で足りるなら caption |

- **`text-[Npx]` 等の font-size 任意値は禁止**（audit C11 恒久ガード）。
- font-weight: 本文 400、強調 500（`font-medium`）、見出し 600（`font-semibold`）。**700（`font-bold`）は title/section 見出し専用** — 本文・数値セルに使わない（§3.2 の 700 vs 400 コントラスト原則）。

---

## 4. Layout / レイアウト

### 4.1 Spacing System

- **ベース単位**: 8px
- **トークン**: `{spacing.xxs}` 4px · `{spacing.xs}` 8px · `{spacing.sm}` 12px · `{spacing.md}` 16px · `{spacing.lg}` 24px · `{spacing.xl}` 28px · `{spacing.xxl}` 32px
- カード内 padding は `{spacing.lg}`（24px）前後。ユーティリティボタンは 4px/14px。フォームフィールドは 6px 程度。

### 4.2 Grid & Container

- コンテンツは max-width カラム（デスクトップ ~1080–1300px）中央配置、外側に十分な gutter
- フィーチャーセクションは full-width テキストと 2-up / 3-up カードグリッドを交互配置
- AE 臨床 UI: 一覧 + SidePeek（§6.2）の 2 ペイン構成を基本とする

### 4.3 Whitespace Philosophy

- セクション間は大きな垂直 gap でグルーピング（罫線より whitespace）
- カードは暖色 canvas 上に hairline で静かに配置 — document-like、スキャンしやすい

### 4.4 Responsive Strategy

| Name | Width | 主な変化 |
|---|---|---|
| Wide | 1440px+ | フルマルチカラム |
| Desktop | 1080–1300px | 標準 3-up グリッド |
| Tablet | 768–840px | 2-up 折りたたみ、ナビ condense |
| Mobile | ≤600px | 単一カラム、ハンバーガー、full-width CTA |

- **タッチターゲット**: 最小 44×44px（pill CTA / utility ボタンは vertical padding を維持）
- **折りたたみ**: タブレット以下でナビ condense、マルチカラム → スタック
- AE は **タブレットファースト**（§7）— iPad 横画面を primary breakpoint として設計

---

## 5. Elevation & Depth / エレベーション

DESIGN.md 準拠：**barely-there** — hairline + 複数レイヤーの極薄シャドウ。重い drop-shadow 禁止。

| Level | Treatment | 用途 |
|---|---|---|
| 0 — Flat | hairline `{colors.hairline}` のみ | 暖色 canvas 上の通常カード |
| 1 — Soft | 極薄 4 段: `rgba(0,0,0,0.01) 0 0.175px 1.041px`, `0.02 0 0.8px 2.925px`, `0.027 0 2.025px 7.847px`, `0.04 0 4px 18px` | 浮動カード・フォーカス中入力 |
| 2 — Elevated | 5 段 deep stack（末尾 `rgba(0,0,0,0.05) 0 23px 52px`） | モーダル・ポップオーバー |

shadcn `DialogContent`（`frontend/src/components/ui/dialog.tsx`）は `rounded-xl` + `p-6`（`{spacing.lg}`）+ `shadow-lg` を既定とし、`ex-modal-card`（`{rounded.xl}` / Level-2）に一致。

### 5.1 実装ルール（2026-07-21 追補 — FE9）

- 実装トークン（`globals.css` `@theme inline`・FE9-1 実装済み）:
  - `--shadow-level1`: `0 0.175px 1.041px rgba(0,0,0,0.01), 0 0.8px 2.925px rgba(0,0,0,0.02), 0 2.025px 7.847px rgba(0,0,0,0.027), 0 4px 18px rgba(0,0,0,0.04)`
  - `--shadow-level2`: `0 0.8px 2.9px rgba(0,0,0,0.02), 0 2px 7.8px rgba(0,0,0,0.027), 0 4px 18px rgba(0,0,0,0.04), 0 10px 32px rgba(0,0,0,0.045), 0 23px 52px rgba(0,0,0,0.05)`（中間段は Level-1 の等比を踏襲した5段・末尾は DESIGN.md 実測値）
  - 製品マイクロトークン（実態の公式化）: `--shadow-btn`（ボタン微細影 = 旧 `--notion-shadow-btn` 値）・`--shadow-panel`（SidePeek 等の左方向影）・`--shadow-focus-brand`（focus リング `0 0 0 1px rgba(0,117,222,.35)` — DESIGN.md「focus signal は primary」原則）・`--shadow-brand-glow`（ナビ進捗バーの brand グロー `0 0 8px rgba(0,117,222,.5)`）— FE10 リブランドで brand 系 rgba を `#0075DE` 基準へ反転済み
- 使い分け: 通常カード = **Level 0（hairline のみ・shadow なし）**／ドロップダウン・ポップオーバー・浮動パネル・フォーカス強調 = `shadow-level1`／モーダル・トースト = `shadow-level2`。
- **Tailwind 既定の `shadow-sm/md/lg/xl` と `shadow-[...]` 任意値は新規使用禁止**（実測: sm×55 / md×4 / lg×5 / 任意値×6・2026-07-21）。FE9 で level トークンへ移行し audit C10 で恒久ガード。`drop-shadow`・CSS 直書き `box-shadow` は 0 件を維持する。

### Decorative Depth

- AE では illustration より **意味的カラー**（§2.4）と hairline 階層で depth を表現
- スティッカーパレットは凡例・バッジなど小面積装飾に限定

---

## 6. Shapes / 形状・角丸

DESIGN.md `rounded:` フロントマターに準拠。**コンポーネント種別ごとに角丸値が意味を持つ**ため、単一値に統一しない。

| トークン | 値 | 用途 |
|---|---|---|
| `{rounded.xxs}` | **4px**（FE10 字義化 — 3px 製品拡張を撤回し DESIGN.md 最小段 xs に整列。クラス名は互換のため残存） | コンパクト入力・アイコンボタン。`--radius-xxs` in `globals.css` |
| `{rounded.xs}` | 4px | フォーム入力（`text-input`）・小タグ |
| `{rounded.sm}` | 5px | メニュー項目・リスト行・status pill |
| `{rounded.md}` | 8px | ユーティリティボタン（`button-utility`）・小カード |
| `{rounded.lg}` | 12px | フィーチャーカード・ illustration frame |
| `{rounded.xl}` | 16px | 大コンテナ・モーダル（`ex-modal-card`） |
| `{rounded.full}` | 9999px | Primary CTA pill・バッジ・円形アイコンボタン |

**Photography / メディア**: スクリーンショットは `{rounded.lg}` / `{rounded.xl}` well 内。アバターは `{rounded.full}`。

---

## 7. Components / UI コンポーネント

> **Hover 状態**: DESIGN.md と同様、Default / Active-Pressed のみ文書化。hover は実装詳細として各プリミティブに委譲。

### 7.1 Navigation

**`nav-bar`** — `{colors.canvas}` surface、`{colors.ink}` link、`{typography.body-sm}`、`{spacing.md}` padding。AE: アプリシェル上部 / サイドバー。

### 7.2 Buttons

| コンポーネント | 仕様 | AE 実装 |
|---|---|---|
| `button-primary` | bg `{colors.primary}`（`#0075DE`）、text `{colors.on-primary}`、`{typography.button}`、`{rounded.full}` pill | `SubmitButton` / `PrimaryButton` — `colorVariant="brand"`（既定・pill 実装済み） |
| `button-primary-pressed` | bg `{colors.primary-active}`（`#005BAB`） | `PALETTE.brandHover` = `#005BAB` ✅ |
| `button-secondary` | white surface、`{colors.ink}`、pill、Level-1 shadow | 二次 CTA |
| `button-utility` | white surface、`{rounded.md}`、4px 14px padding、hairline border | ナビ / ユーティリティ操作 |
| `button-icon-circular` | `rgba(0,0,0,0.05)` fill、`{rounded.full}` | カルーセル / メディア制御 |

> **FE10 字義化**: 旧「pill はマーケ専用」裁定（2026-07-21 旧版）は撤回。DESIGN.md 字義どおり **`button-primary` = pill**（実装は既に準拠）。フォーム入力への pill 適用は引き続き禁止（DESIGN.md Don't — 入力は `{rounded.xs}` 4px）。

### 7.3 Cards & Containers

| コンポーネント | 仕様 | AE 実装 |
|---|---|---|
| `feature-card` | white、`{rounded.lg}`、`{spacing.lg}` padding、Level-0 | フォームセクションカード |
| `feature-card-elevated` | 同上 + Level-1 shadow | 浮動パネル |
| `pricing-plan-card` | `{rounded.md}`、plan 列 | — |
| `pricing-plan-card-featured` | `{colors.canvas-soft}` tint | — |
| `hero-band` | `{colors.secondary}` full-bleed、`{typography.display-1}` | マーケ向け（AE 本体では限定的） |
| `badge-pill` | white surface、`{colors.primary}` text、eyebrow、pill | `BADGE.*` |
| `footer` | `{colors.canvas-soft}`、`{typography.caption}` | — |

### 7.4 Inputs & Forms

**`text-input`**: white surface、`{typography.body-sm}`、1px border、`{rounded.xs}`（4px）、padding 6px。Focus で Level-1 shadow。

- AE: `Input` / `Textarea` / `SelectTrigger` — `rounded-xs`（4px）準拠。`--radius-xs: 4px` in `globals.css`

### 7.5 Examples (`ex-*`) — キットミラー面

| トークン | 説明 | AE マッピング |
|---|---|---|
| `ex-pricing-tier` | canvas-soft surface + hairline border | — |
| `ex-pricing-tier-featured` | polarity-flipped featured tier | — |
| `ex-product-selector` | SaaS summary card | — |
| `ex-cart-drawer` | サブスクリプション summary | — |
| `ex-app-shell-row` | サイドバー行、active = `{colors.primary}` indicator | サイドバーナビ active 状態 |
| `ex-data-table-cell` | header: canvas-soft + eyebrow、body: body-sm、hairline 行区切り | `OwnersListTable` / `PetSelectionResultsTable` / `OwnerPetsSection` / `HistoryTable` / `AggregationOwnerTable` |
| `ex-auth-form-card` | feature-card + text-input | `OwnerForm` 等 |
| `ex-modal-card` | feature-card + Level-2 shadow | shadcn `Dialog` 全モーダル |
| `ex-empty-state-card` | canvas-soft + `{spacing.xxl}` padding | 空状態 |
| `ex-toast` | feature-card shape + medium shadow | トースト通知 |

> **テーブルヘッダ様式（FE10 字義化・2026-07-21）**: 旧 house 様式裁定（plain muted ヘッダ）は撤回。`STYLE.tableHeaderRow` / `STYLE.tableHeaderCell` を DESIGN.md `ex-data-table-cell` 字義（**canvas-soft ヘッダ帯 + eyebrow 型 12px/600/tracking + hairline 行区切り**）へ一括反転済み — STYLE トークン経由の全テーブルに一斉適用される。STYLE を経由しない手書きヘッダは FE10 R6 スイープで検出し同様式へ統一する（部分適用は不統一を生むため禁止）。

### 7.6 Animal Ekarte 固有パターン

#### 7.6.1 プロパティ編集 (`PropertyInput`)

通常はプレーンテキストのように見え、マウスホバーやフォーカスで初めて入力枠が現れる、ボーダーレスな入力体験。

#### 7.6.2 サイドピークパネル (`SidePeekPanel`)

一覧画面から詳細情報を「覗き見る」ためのサイドスライド形式。文脈（コンテキスト）を維持したまま編集が可能。

#### 7.6.3 非同期フィードバック (`FilteringIndicator`)

大規模データの検索中、UI をフリーズさせることなく「計算中」であることを透過アニメーションで表現する。

---

## 8. デバイス最適化 (タブレット・ファースト)

診察室での iPad 利用を想定し、Tailwind v4 `@theme inline` により全体的なクリックターゲットとフォントサイズを標準より **10% 拡大** している。タッチ操作の精度を高めつつ、高い可読性を確保する。

---

## 9. Do's and Don'ts

### Do

- `{colors.primary}`（**`#0075DE`**）は Primary CTA・インラインリンク・active/focus のみに使う。装飾には使わない。
- ページ canvas は暖色 `{colors.canvas-soft}`、カード・フィールドは白 `{colors.surface}` にする。
- スティッカーパレット（`{colors.accent-pink}`、`{colors.accent-teal}` 等）はバッジ・タグ・凡例ドットなど装飾用途にのみ使う。
- 見出し階層は §3.4 のロール表に従う。
- Primary CTA は pill `{rounded.full}`（`button-primary` 字義）、ユーティリティボタンは `{rounded.md}` — 対比は意図的。
- カード境界は hairline `#E6E6E6` + Level-1 の極薄シャドウで表現する。
- 深 indigo `{colors.secondary}` hero-band は単一の hero モーメントに限定する。

### Don't

- スティッカーパレットの色を CTA・構造フィルに使わない。
- `{colors.primary}` 以外の第二構造アクセントを新規導入しない。
- フォーム入力に pill（`{rounded.full}`）を使わない — 入力は `{rounded.xs}`（4px）。
- 重い drop-shadow を使わない。
- 本文を heavy weight にしない — 400 で可読性、700 は見出し専用。
- 全ページを純白 clinical white にしない — 暖色 `{colors.canvas-soft}` が brand calm の核心。
- 危険バッジ・死亡グレーアウト・RBAC 非活性表示など臨床安全 UI をデザイン変更で退行させない。
- コンポーネント内で hex 直書きしない — `design-tokens.ts` 経由で参照する。

---

## 10. 技術的 SSOT

デザインに関する全ての定数は、以下のファイルで一元管理されている。

- **`frontend/src/lib/design-tokens.ts`**: `PALETTE`（raw hex）、`C`（Tailwind クラス）、`STYLE`（複合クラスプリセット）、`LAYOUT`（寸法）、`BADGE`（バッジ配色コンボ）、`ICON`（アイコンサイズ）。
- 新しい色を追加する場合は、必ず `design-tokens.ts` に追加した上でコンポーネントから参照する。**コンポーネント内での hex 直書きは禁止。**
- **規約 vs 実装**: 本書 §2.6 のとおり、構造色 `#0075DE` は tokens・`globals.css` ともに同期済み（FE10）。route 表面の legacy 色（旧 teal `#038B94`/`#027078`・旧 accent `#2383E2`）は audit C1 が機械的に禁止する。

---
