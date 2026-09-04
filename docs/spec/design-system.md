# UI デザインシステム 規約 (Design System)

> **目的**: [DESIGN.md](../../DESIGN.md)（ルート — Notion Analysis / 意匠言語）を Animal Ekarte の実装に落とし込むための規約を定義する。
> **読者**: フロントエンド実装者。
> **タイミング**: UI 実装・レビュー時。
> **最新更新**: 2026-07-27（brand と primary を Animal Ekarte teal に統一）

### SSOT 優先順位 — **軸ごとに正本が異なる**（FE11 決裁・2026-07-21 曽我）

| 軸 | 正本 | 理由 |
|---|---|---|
| **色（colors 全般）** | **本書 (`docs/spec/design-system.md`)** | 本システムの色は**装飾ではなく業務・臨床の意味論**を担う（ステータス識別・危険/死亡・RBAC 非活性）。DESIGN.md の色規定は「イラスト・アイコンタイル・カテゴリドット」という*装飾*用途を前提にしており、問題領域が異なる。よって色は製品判断（本書）を正本とする。 |
| タイポグラフィ / 形状 / 余白 / エレベーション / コンポーネント寸法 | [DESIGN.md](../../DESIGN.md) | 業務意味論を持たないため**字義で従う**。size / line-height / letter-spacing / radius / spacing / shadow は DESIGN.md の値をそのまま適用する（FE11-F1 で全ロール一致済み）。 |
| 実行時 | `frontend/src/lib/design-tokens.ts` | `C` / `STYLE` / `PALETTE` / `LAYOUT`。上記2軸の決定を実装する層であり、**独自判断で値を足さない**。 |

> **色が本書正本である帰結（FE11 で確定）**
> - **brand と primary はともに `#038B94` / active `#027078`** を採用する。意味役割をコード上で明示するためトークン名は分けるが、画面上の主要操作色はブランドカラーへ統一する。
> - **DESIGN.md のスティッカーパレット8色は採用しない**（§2.5）。
> - **ink ランプ（`#000000` / `#31302E` / `#615D59` / `#A39E98`）は DESIGN.md と同値を採用**する（§2.3）。装飾ではなく可読性の階層であり、DESIGN.md の4段が製品要件を満たすため一致させた（FE11-F2 実装済み）。
> - **nav-bar（サイドバー）は canvas-soft を維持**する。DESIGN.md は白 canvas を規定するが、白にすると本文カードとの図地関係（§1 核心原則）が消えるため、色軸の正本である本書の判断を優先する。

---

## 1. Overview / デザイン思想

本システムの UI は、**「臨床現場での集中力を妨げないクリーンな情報設計」**を最優先としている。DESIGN.md の言葉を借りれば、「良い光の下の整理されたデスク」— 暖色の紙のようなキャンバス（`{colors.canvas-soft}`）の上に、抑制されたモノクロと Animal Ekarte teal の UI が乗る。業務ステータス色・意味的カラーとは混同しない。

### 核心となる原則

- **ミニマリズム**: 装飾を削ぎ落とし、データそのものを主役にする。汎用操作と製品識別は `{colors.primary}` / `{colors.brand}`（`#038B94`）の同一色へ統一する。
- **直感的な階層**: Notion スタイルの「プロパティ」と「カード」による情報の整理。カードは白 surface + hairline、ページ canvas は暖色オフホワイト。
- **臨床安全の可視化**: 警告や異常値のみを、戦略的に配色された「意味的カラー」で強調する。危険バッジ・死亡グレーアウト・RBAC 非活性表示など、臨床安全 UI は本デザイン変更で退行させない。

### Key Characteristics（色=本書決定 / その他=DESIGN.md 字義）

- 暖色 paper-soft canvas `{colors.canvas-soft}` — 純白ページではなく document-like な落ち着き
- 近黒 `{colors.ink}` の Inter 系タイポ。display サイズではネガティブ tracking
- **semantic primary** — `{colors.primary}` = **`#038B94`** — 汎用 Primary CTA・リンク・active/focus
- **brand identity** — `{colors.brand}` = **`#038B94`** — ログインなど認証・製品識別・明示的 brand CTA
- 業務ステータス色（紫/橙/青/緑）は状態の識別子。構造には使わない（DESIGN.md sticker palette は不採用 — §2.5）
- Primary CTA は pill `{rounded.full}`、ユーティリティボタンは `{rounded.md}`（8px）— 意図的な対比
- hairline `#E6E6E6` + 極薄レイヤードシャドウによる elevation（重い drop-shadow 禁止）
- 深 indigo `{colors.secondary}` の hero-band はマーケ向けパターン。AE アプリ本体では臨床 UI に合わせて限定的に使用

---

## 2. Colors / カラーパレット

> **色は本書が正本**（FE11 決裁 — 冒頭 SSOT 表）。DESIGN.md の色値は参考として併記するが、採否は本書が決める。

### 2.1 Brand & Accent（構造色）

| トークン | 値（製品決定） | 用途 |
|---|---|---|
| `{colors.brand}` | **`#038B94`**（Animal Ekarte teal） | 製品識別。認証 CTA、ロゴ、明示的 brand surface。 |
| `{colors.brand-active}` | **`#027078`** | brand surface の hover / 押下状態。 |
| `{colors.on-brand}` | `#FFFFFF` | brand CTA 上の大きな太字テキスト、またはアイコン。通常サイズ本文には使わない。 |
| `{colors.on-brand-active}` | `#FFFFFF` | brand-active 背景上のテキスト。 |
| `{colors.primary}` | **`#038B94`** | semantic primary。汎用 Primary CTA、インラインリンク、active-tab / selection / focus ring。brand と同値。 |
| `{colors.primary-active}` | **`#027078`** | primary の hover / 押下状態。brand-active と同値。 |
| `{colors.secondary}` | `#213183` | 深 indigo hero-band（AE では限定的） |
| `{colors.on-primary}` | `#FFFFFF` | primary CTA 上のテキストとアイコン。 |
| `{colors.on-primary-active}` | `#FFFFFF` | primary-active 上の hover / 押下テキスト。`#027078` 上で WCAG AA 4.5:1 を満たす。 |

> **brand と primary は同じ色**: どちらも `#038B94`、active は `#027078`。製品識別と操作階層をコード上で読み分けるためトークン名は維持する。`PALETTE.accent` の旧値 `#2383E2` は compatibility 用に残る。一方 `C` に `accent` member は存在せず、`C.accent` の consumption は audit C1 で禁止する。

> **DESIGN.md のスティッカーパレット8色（Sky `#62AEF0` / Purple `#D6B6F6` / Pink `#FF64C8` / Orange `#DD5B00` / Teal `#2A9D99` / Green `#1AAE39` 他）は本システムでは採用しない** — 理由と代替は §2.5。

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

### 2.4 意味的カラー (Semantic Colors) — 臨床安全のための色（存続決定）

DESIGN.md の Semantic 節は「Notion の*マーケ表面*は専用 semantic ramp を持たない」という観察であり、削除の指令ではない。臨床安全（SPECIFICATION 2.1 — 全原則に優先）により、本システムは意味的カラーを **FE10 字義リブランド後も維持する**。`design-tokens.ts` で一元管理し、構造色用途（CTA・リンク・active/focus）には使わない。

- **危険 / 死亡 / 異常高**: `C.danger`（`#C0392B`。white 上の contrast は約 5.44:1 で、normal text の WCAG AA を満たす。`design-tokens.ts` の 7.1:1 comment は既知の source comment drift）
- **死亡の文脈修飾**: 一覧 surface ではグレーアウトし、単一患者画面では `C.danger` を使用する。
- **注意 / 期限間近**: `C.textWarning` / `C.bgWarning50`
- **正常 / 完了 / 生存**: `C.textStatusGreen` / `C.bgStatusGreen`
- **異常低 / 寒冷**: `C.textStatusBlue` 系

> 危険バッジ・死亡グレーアウト・RBAC 非活性表示など臨床安全 UI は、デザイン変更で退行させない。

本体 routes/pages で `bg-white` / `text-black` / `border-white` 等の white/black named color を直接指定せず、用途が追跡できる `C` / `STYLE` token を使う（audit C15）。

### 2.5 業務ステータス色 — DESIGN.md スティッカーパレットは不採用（FE11 決裁）

**DESIGN.md の sticker palette 8色は本システムでは採用しない。** DESIGN.md におけるあの8色は「イラスト・アプリアイコンタイル・カテゴリドット」という*装飾*の規定であり、本システムの色は**スタッフが状態を読むための業務識別子**として機能している（受付ステータス・予約区分・在庫状態など）。用途が異なるものに同じ規定を当てても遵守にならないため、色の正本である本書の判断として不採用とする。

**実装（`design-tokens.ts` が正本）**:

| 用途 | トークン | 値 |
|---|---|---|
| ステータス: 灰（未着手/無効） | `C.bgStatusGrayMedium` | `#9B9A97` |
| ステータス: 紫（診療中） | `C.bgStatusPurpleDot` | `#6940A5` |
| ステータス: 橙（会計待ち/割引） | `C.bgDiscount` | `#D9730D` |
| ステータス: 青（受付済/checked_in） | `C.bgStatusBlueDot` | `bg-blue-500` |
| ステータス: 空/緑/emerald 系 | `C.bgStatusSkyDot` / `C.bgStatusEmeraldDot` 他 | Tailwind パレット |

**規律（DESIGN.md と共通で維持する原則）**:

- ステータス色を **CTA・構造フィルに使わない**（汎用構造色は primary）。
- **primary（構造色）と brand（製品識別色）をステータスや装飾に使わない**。値変更時は消費者を操作・製品識別・ステータス・装飾で必ず仕分ける。
- **外部ブランド例外**: LINE 公式グリーン `#06C755`（`PALETTE.lineGreen`）は外部ブランド識別のため構造色ルールの対象外。構造・CTA には使わない。

### 2.7 適用範囲外（媒体が異なる面・FE11 明文化）

本規約は**画面 UI（screen chrome）**を対象とする。以下は媒体が異なり要件も異なるため対象外とし、それぞれの内部で一貫していればよい:

| 対象外の面 | 実体 | 理由 |
|---|---|---|
| **印刷帳票** | `PrintPortal` 配下（`MonthlyReportPrintArea` / `ClosePrintArea`）。raw Tailwind グレー 92 箇所 | 紙・モノクロプリンタ出力が要件。canvas-soft／hairline／ink ランプは画面の図地設計であり紙面には適用しない |
| **マニュアル本文** | `ManualContent` の markdown レンダリング（`bg-black/5` ヘッダ等） | 文書レンダリングであり、アプリのデータテーブル規範（`ex-data-table-cell`）の対象外 |
| **LINE ミニアプリ** | `liff/` `line-reserve/` `src/shared-liff/` | 別アプリ。FE10 でスイープ対象外を宣言済み。本体 route 数は [ui-design-compliance.md](ui-design-compliance.md) の静的在庫を正とする。audit C12/C14 も明示 allowlist で除外 |

> 対象外にする場合は**必ず本節に列挙する**。列挙のない暗黙の除外は「対象ゼロ」を「全件合格」に見せかけるため禁止。

### 2.6 `design-tokens.ts` マッピング表

| 役割 | DESIGN.md トークン | 製品採用値 | design-tokens.ts（製品決定同期） |
|---|---|---|---|
| brand identity | `{colors.brand}` | **`#038B94`** | `PALETTE.brand` / `C.bgBrandIdentity` / `C.borderBrandIdentity` ✅ |
| brand Hover / Pressed | `{colors.brand-active}` | **`#027078`** | `PALETTE.brandHover` / `C.activeBgBrandIdentity` ✅ |
| brand text | アクセシビリティ派生色 | **`#025F66`**（light）/ **`#079BA5`**（dark） | `C.textBrandIdentity`。各 surface で通常文字 4.5:1 を満たす ✅ |
| semantic primary | `{colors.primary}` | **`#038B94`** | `PALETTE.actionPrimary` / `C.bgActionPrimary` / `C.borderActionPrimary` ✅ |
| primary Hover / Pressed | `{colors.primary-active}` | **`#027078`** | `PALETTE.actionPrimaryActive` / `C.activeBgActionPrimary` ✅ |
| ページ canvas | `{colors.canvas-soft}` | `#F6F5F4` | `PALETTE.bgMain`, `C.bgPage` ✅ |
| カード surface | `{colors.canvas}` | `#FFFFFF` | `C.bgWhite` / `C.bgSubtle` ✅ |
| Hairline | `{colors.hairline}` | **`#E6E6E6`** | `C.borderLight` / `globals.css` `--border` = `#E6E6E6` ✅（旧 `rgba(0,0,0,0.09)` を字義固体値化） |
| 入力 border | `text-input` 1px `rgb(221,221,221)` | `#DDDDDD` | `globals.css` `--input` ✅ |
| Ink 系（4段） | `{colors.ink}` 〜 `{colors.ink-faint}` | `#000000` / `#31302E` / `#615D59` / `#A39E98` | **FE11-F2 で実値化**。`C.text`=ink／`C.text90-70`=ink-secondary／`C.text65-50`=ink-muted／`C.text45` 以下と placeholder=ink-faint。新規実装は `C.textInk` / `textSecondary` / `textMuted` / `textFaint` を使う ✅ |
| checked / focus | `{colors.primary}` | **`#038B94`** | `C.dataCheckedBgActionPrimary` / `--ring`。汎用 selection / focus は semantic primary に統一 ✅ |
| CSS 変数 | brand / primary | **`#038B94` / `#038B94`** | `globals.css` `--brand` と `--primary` は意味名を維持しつつ同値。`--sidebar-primary` も同じ teal ✅ |

> **legacy トークンの扱い（FE10）**:
> - compatibility 用の `PALETTE.accent` は旧値 `#2383E2` を保持する。`C.accent` member は存在せず、その consumption は audit C1 で禁止する。
> - 新規の汎用操作は action-primary 系、製品識別が明示された表面だけ brand 系を使う。

---

## 3. Typography / タイポグラフィ

DESIGN.md `typography:` フロントマターに準拠。実装のフォントファミリーは **`'Inter', 'Noto Sans JP'`**（フォールバック: `-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`。`frontend/src/styles/globals.css` の `body` / typography selector）。単一スタックで display から eyebrow まで担う。

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

アプリ本体 UI は `{typography.title}` 以下を中心に使い、大きな見出しが必要な箇所だけ heading ロールを使う。display 2段は描画箇所がないため未実装である。実装済みサイズは **DESIGN.md 字義値**に整列済み。

| ロール | AE 実装（Tailwind） | 実サイズ | 用途 |
|---|---|---|---|
| heading-1 | `text-heading-1 font-bold` | 40px | ページ級の大見出し |
| heading-2 | `text-heading-2 font-bold` | 26px | 大セクション見出し・大型サイドパネルタイトル |
| heading-3 | `text-heading-3 font-bold` | 22px | カードタイトル |
| title | `text-xl font-semibold` | 20px | ページ内最上位見出し（PageLayout タイトル等） |
| body-md | `text-base` | 16px | 標準本文・フォームラベル |
| body-sm | `text-sm` | 15px（`--text-sm`） | テーブル行・密な UI の既定 |
| caption | `text-xs` | **14px**（`--text-xs` — DESIGN.md caption 字義値。FE10 で 13px 上書きを撤回） | キャプション・メタ・placeholder |
| eyebrow | `text-2xs font-semibold`（字送りは `--text-2xs--letter-spacing: 0.125px` が伴走） | **12px**（DESIGN.md eyebrow 字義値） | データテーブルヘッダー（`STYLE.tableHeaderCell`）・小ラベル |
| micro | `text-2xs` | **12px**（`--text-2xs` — FE10 で 11px 製品拡張を撤回し eyebrow 段に統合） | バッジ内文字・極小メタ表示。乱用禁止 — caption で足りるなら caption |

- **`text-[Npx]` 等の font-size 任意値は禁止**（audit C11）。DESIGN.md にない `text-lg/2xl/3xl/4xl+` も禁止し、heading/title ロールへ写像する（audit C12）。
- 黒アルファによる ink 段の迂回は禁止する（audit C13）。letter-spacing はロールに伴走するため、`tracking-wide` 等や任意値で上書きしない（audit C14）。
- font-weight: 本文 400、強調・ボタン 500（`font-medium`）、title/eyebrow 600（`font-semibold`）、heading 700（`font-bold`）。**700 は heading 専用**で、本文・数値セルに使わない（§3.2 の 700 vs 400 コントラスト原則）。

---

## 4. Layout / レイアウト

### 4.1 Spacing System

- **ベース単位**: 8px
- **トークン**: `{spacing.xxs}` 4px · `{spacing.xs}` 8px · `{spacing.sm}` 12px · `{spacing.md}` 16px · `{spacing.lg}` 24px · `{spacing.xl}` 28px · `{spacing.xxl}` 32px
- カード内 padding は `{spacing.lg}`（24px）前後。ユーティリティボタンは 4px/14px。フォームフィールドは 6px 程度。
- `p-5` / `m-5` / `gap-5`、負値、`[20px]` / `[1.25rem]` 等の20px spacing utilityは禁止（audit C16）。

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

Tailwind の既定/任意 shadow と drop-shadow は audit C10、CSS の直接 `box-shadow:` / `filter: drop-shadow()` は audit C17 が禁止する。

| Level | Treatment | 用途 |
|---|---|---|
| 0 — Flat | hairline `{colors.hairline}` のみ | 暖色 canvas 上の通常カード |
| 1 — Soft | 極薄 4 段: `rgba(0,0,0,0.01) 0 0.175px 1.041px`, `0.02 0 0.8px 2.925px`, `0.027 0 2.025px 7.847px`, `0.04 0 4px 18px` | 浮動カード・フォーカス中入力 |
| 2 — Elevated | 5 段 deep stack（末尾 `rgba(0,0,0,0.05) 0 23px 52px`） | モーダル・ポップオーバー |

shadcn `DialogContent`（`frontend/src/components/ui/dialog.tsx`）は `rounded-xl` + `p-6`（`{spacing.lg}`）+ `shadow-level2` を使い、`ex-modal-card`（`{rounded.xl}` / Level-2）に一致。Tailwind 既定 shadow を許可する例ではない。

### 5.1 実装ルール（2026-07-21 追補 — FE9）

- 実装トークン（`globals.css` `@theme inline`・FE9-1 実装済み）:
  - `--shadow-level1`: `0 0.175px 1.041px rgba(0,0,0,0.01), 0 0.8px 2.925px rgba(0,0,0,0.02), 0 2.025px 7.847px rgba(0,0,0,0.027), 0 4px 18px rgba(0,0,0,0.04)`
  - `--shadow-level2`: `0 0.8px 2.9px rgba(0,0,0,0.02), 0 2px 7.8px rgba(0,0,0,0.027), 0 4px 18px rgba(0,0,0,0.04), 0 10px 32px rgba(0,0,0,0.045), 0 23px 52px rgba(0,0,0,0.05)`（中間段は Level-1 の等比を踏襲した5段・末尾は DESIGN.md 実測値）
  - 製品マイクロトークン（実態の公式化）: `--shadow-btn`（ボタン微細影）・`--shadow-panel`（SidePeek 等の左方向影）・`--shadow-focus-primary`（汎用 focus）・`--shadow-focus-brand`（brand surface 専用 focus）・`--shadow-brand-glow`（brand 表現専用）。汎用 focus と `--ring` は `#038B94` を使う
- 使い分け: 通常カード = **Level 0（hairline のみ・shadow なし）**／ドロップダウン・ポップオーバー・浮動パネル・フォーカス強調 = `shadow-level1`／モーダル・トースト = `shadow-level2`。
- **Tailwind 既定の `shadow-sm/md/lg/xl` と `shadow-[...]` 任意値は新規使用禁止**（実測: sm×55 / md×4 / lg×5 / 任意値×6・2026-07-21）。FE9 で level トークンへ移行し audit C10 で恒久ガード。`drop-shadow`・CSS 直書き `box-shadow` は 0 件を維持する。

### Decorative Depth

- AE では illustration より **意味的カラー**（§2.4）と hairline 階層で depth を表現
- 業務ステータス色は凡例・バッジなど小面積の状態表示に限定

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

**`nav-bar`** — AE では **`{colors.canvas-soft}` surface**（`C.bgPage`）、`{colors.ink}` link、`{typography.body-sm}`、`{spacing.md}` padding。DESIGN.md は白 canvas を規定するが、白にすると本文カードとの図地関係が消えるため色軸の正本である本書の判断で canvas-soft を採用する（FE11 決裁）。

### 7.2 Buttons

| コンポーネント | 仕様 | AE 実装 |
|---|---|---|
| `button-primary` | bg `{colors.primary}`（`#038B94`）、text `{colors.on-primary}`、`{typography.button}`、`{rounded.full}` pill | `SubmitButton` / `PrimaryButton` — `colorVariant="primary"`（既定） |
| `button-primary-pressed` | bg `{colors.primary-active}`（`#027078`）、text `{colors.on-primary-active}`（`#FFFFFF`） | `PALETTE.actionPrimaryActive` / `C.activeTextOnActionPrimary` ✅ |
| `button-brand` | bg `{colors.brand}`（`#038B94`）、hover/pressed `{colors.brand-active}`（`#027078`） | 認証などで `colorVariant="brand"` を明示 |
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

- `{colors.primary}` と `{colors.brand}` は同じ **`#038B94`** を使う。active はともに **`#027078`**。
- 汎用操作では primary トークン、認証・ロゴなど製品識別では brand トークンを使い、意味をコード上で明示する。
- ページ canvas は暖色 `{colors.canvas-soft}`、カード・フィールドは白 `{colors.surface}` にする。
- 業務ステータス色（§2.5）はバッジ・タグ・凡例ドット等の状態表示にのみ使う。primary と brand をステータスや装飾に流用しない。
- 見出し階層は §3.4 のロール表に従う。
- Primary CTA は pill `{rounded.full}`（`button-primary` 字義）、ユーティリティボタンは `{rounded.md}` — 対比は意図的。
- 通常カードは hairline `#E6E6E6` の Level 0、浮動面だけを Level-1 の極薄シャドウで表現する。
- 深 indigo `{colors.secondary}` hero-band は単一の hero モーメントに限定する。

### Don't

- 業務ステータス色を CTA・構造フィルに使わない（汎用構造色は primary）。
- `{colors.primary}` 以外の第二構造アクセントを新規導入しない。
- フォーム入力に pill（`{rounded.full}`）を使わない — 入力は `{rounded.xs}`（4px）。
- 重い drop-shadow を使わない。
- 本文を heavy weight にしない — 400 で可読性、700 は見出し専用。
- 全ページを純白 clinical white にしない — 暖色 `{colors.canvas-soft}` が brand calm の核心。
- 危険バッジ・死亡グレーアウト・RBAC 非活性表示など臨床安全 UI をデザイン変更で退行させない。
- コンポーネント内で hex 直書きしない — `design-tokens.ts` 経由で参照する。

---

## 10. 技術的 SSOT

TypeScriptから参照する共有デザイントークンは`design-tokens.ts`、CSS/theme変数は`globals.css`を正本とする。component-localなraw値は追加しない。

- **`frontend/src/lib/design-tokens.ts`**: `PALETTE`（raw hex）、`C`（Tailwind クラス）、`STYLE`（複合クラスプリセット）、`LAYOUT`（寸法）、`BADGE`（バッジ配色コンボ）、`ICON`（アイコンサイズ）。
- 新しい色を追加する場合は、必ず `design-tokens.ts` に追加した上でコンポーネントから参照する。**コンポーネント内での hex 直書きは禁止。**
- **規約 vs 実装**: 本書 §2.6 のとおり、brand と primary は tokens・`globals.css` で意味名を分けつつ、値は `#038B94`（active `#027078`）へ統一する。compatibility 用 `PALETTE.accent` は旧値 `#2383E2` を保持するが、存在しない `C.accent` の consumption は audit C1 が機械的に禁止する。

---
