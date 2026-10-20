---
version: 1.0
name: Animal Ekarte Design Language
description: 動物病院向け電子カルテ Animal Ekarte の意匠言語。タイポグラフィ・形状・余白・エレベーション・コンポーネント寸法の正本。色の値と採否は docs/spec/design-system.md が正本。

colors:
  # 値は docs/spec/design-system.md §2 が正本。ここでは本文のトークン参照のために同値を再掲する。
  brand: "#038b94"
  brand-active: "#027078"
  on-brand: "#ffffff"
  on-brand-active: "#ffffff"
  primary: "#038b94"
  primary-active: "#027078"
  on-primary: "#ffffff"
  on-primary-active: "#ffffff"
  secondary: "#213183"
  canvas: "#ffffff"
  canvas-soft: "#f6f5f4"
  surface: "#ffffff"
  ink: "#000000"
  ink-secondary: "#31302e"
  ink-muted: "#615d59"
  ink-faint: "#a39e98"
  hairline: "#e6e6e6"

typography:
  display-1:
    fontFamily: Inter
    fontSize: 64px
    fontWeight: 700
    lineHeight: 1.0
    letterSpacing: -2.125px
  display-2:
    fontFamily: Inter
    fontSize: 54px
    fontWeight: 700
    lineHeight: 1.04
    letterSpacing: -1.875px
  heading-1:
    fontFamily: Inter
    fontSize: 40px
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: -1px
  heading-2:
    fontFamily: Inter
    fontSize: 26px
    fontWeight: 700
    lineHeight: 1.23
    letterSpacing: -0.625px
  heading-3:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: 700
    lineHeight: 1.27
    letterSpacing: -0.25px
  title:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: 600
    lineHeight: 1.4
    letterSpacing: -0.125px
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: 0
  body-sm:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: 400
    lineHeight: 1.33
    letterSpacing: 0
  button:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 500
    lineHeight: 1.5
    letterSpacing: 0
  caption:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.43
    letterSpacing: 0
  eyebrow:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 600
    lineHeight: 1.33
    letterSpacing: 0.125px

rounded:
  xs: 4px
  sm: 5px
  md: 8px
  lg: 12px
  xl: 16px
  full: 9999px

spacing:
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 24px
  xl: 28px
  xxl: 32px

components:
  nav-bar:
    backgroundColor: "{colors.canvas-soft}"
    textColor: "{colors.ink}"
    typography: "{typography.body-sm}"
    padding: 16px
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button}"
    rounded: "{rounded.full}"
  button-primary-pressed:
    backgroundColor: "{colors.primary-active}"
    textColor: "{colors.on-primary-active}"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.full}"
  button-utility:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: 4px 14px
  badge-pill:
    # 色は固定しない: `BADGE.*`（design-states §2.1 の AA 実測表）から組み合わせを選ぶ。
    # 旧 surface + primary 文字（#038B94 on white = 4.10:1）は 12px で AA 未達のため撤回。
    colorPairs: "BADGE.* (design-states.md §2.1)"
    typography: "{typography.eyebrow}"
    rounded: "{rounded.full}"
    padding: 4px 8px
  feature-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 24px
  feature-card-elevated:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.lg}"
    padding: 24px
  text-input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.xs}"
    padding: 6px
  ex-app-shell-row:
    description: "サイドバーのナビ行。active 状態は semantic primary をインジケータに使う。"
    backgroundColor: "{colors.canvas-soft}"
    activeIndicator: "{colors.primary}"
    rounded: "{rounded.sm}"
    padding: "{spacing.sm} {spacing.md}"
  ex-data-table-cell:
    description: "業務 data-table の th / td 基本様式。ヘッダは eyebrow、本文は body-sm。"
    headerBackground: "{colors.canvas-soft}"
    headerTypography: "{typography.eyebrow}"
    bodyTypography: "{typography.body-sm}"
    cellPadding: "{spacing.sm} {spacing.md}"
    rowBorder: "{colors.hairline}"
  ex-auth-form-card:
    description: "ログイン・パスワード再設定などの認証フォームカード。feature-card の chrome に text-input を内包する。"
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.xl}"
    padding: "{spacing.lg}"
  ex-modal-card:
    description: "モーダルダイアログ面。feature-card の chrome + Level-2 shadow。"
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.xl}"
    padding: "{spacing.lg}"
  ex-empty-state-card:
    description: "空状態カード。"
    backgroundColor: "{colors.canvas-soft}"
    rounded: "{rounded.xl}"
    padding: "{spacing.xxl}"
    captionTypography: "{typography.body-md}"
  ex-toast:
    description: "トースト通知面。feature-card の形状 + Level-1 shadow。"
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.xl}"
    padding: "{spacing.sm} {spacing.md}"
    typography: "{typography.body-sm}"

---

# Animal Ekarte Design Language

> **位置づけ**: 本書は Animal Ekarte の**意匠言語**であり、**タイポグラフィ・形状（角丸）・余白・エレベーション・動き（Motion）・コンポーネント寸法の正本**である。実装はこれらの値に字義で従う。
> **色は本書の正本範囲外**: 色の値・採否・臨床/業務ステータス色は [docs/spec/design-system.md](docs/spec/design-system.md) §2 が正本（FE11 決裁・2026-07-21）。本書のフロントマター `colors:` は本文の `{colors.*}` 参照を解決するための同値再掲であり、ここを編集して色を変えてはならない。
> **関連正本**: 画面の組み立て（ナビ現在地・一覧の列設計・フォーム・固定要素・空状態）は [design-patterns.md](docs/spec/design-patterns.md)、状態（hover / focus-visible / disabled / loading / invalid）は [design-states.md](docs/spec/design-states.md)、トークンの配置と参照方向は [design-token-layers.md](docs/spec/design-token-layers.md)、部品契約は [components/](docs/spec/components/README.md)、準拠検査 C1〜C24 は [ui-design-compliance.md](docs/spec/ui-design-compliance.md)（C21〜C24 の定義は [design-token-layers.md §6](docs/spec/design-token-layers.md)）。全エージェント共通の入口は [ui-design-guide.md](docs/spec/ui-design-guide.md)。
> **履歴**: 初版（alpha）は Notion のマーケティングサイトを分析した意匠言語を転用したものだった。2026-10-20 に Animal Ekarte 自身の意匠言語として書き直し、本製品で使わないマーケ専用パターン（スティッカーパレット 8 色、hero-band、pricing-plan-card、footer、ex-pricing-tier、ex-product-selector、ex-cart-drawer、button-icon-circular）を削除した。**colors / typography / rounded / spacing の値はすべて初版と同一**であり、`design-tokens.ts` / `globals.css` / design-audit への影響はない。`nav-bar` と `ex-app-shell-row` の面だけは design-system.md §7.1 の既存決裁（canvas-soft）に合わせて参照先を揃えた。

## Overview

Animal Ekarte の画面は、**良い光の下で整理された診察机**のように見えることを目指す。診察室の iPad や受付の PC で一日中開かれる業務画面として、スタッフの集中を妨げず、臨床の安全に関わる情報（死亡・危険・期限超過）だけが確実に目に入る情報設計を最優先とする。

支配的な面は純白ではなく、暖色の紙のようなオフホワイト `{colors.canvas-soft}`（#f6f5f4）である。その上に白い `{colors.surface}` のカード・フォーム・テーブルが hairline `{colors.hairline}` で区切られて乗り、図と地の関係を作る。文字は Inter 系（実装は `'Inter', 'Noto Sans JP'`）を近黒 `{colors.ink}` で組み、見出しは weight 700 と負の letter-spacing で「置かれた」印象に、本文は 400 / line-height 1.5 で文書としての可読性を保つ。

構造色は **1 色だけ**である。製品識別（`{colors.brand}`）と汎用操作・リンク・選択・focus（`{colors.primary}`）は同じ teal `#038b94` を共有し、意味役割だけをトークン名で分ける。それ以外の色は装飾ではなく**業務・臨床の意味論**（受付ステータス、予約区分、危険・死亡 sentinel、RBAC 非活性）を担い、その採否と値は design-system.md が決める。

**Key Characteristics**

- 暖色の paper-soft canvas `{colors.canvas-soft}`。全ページを純白にしない
- 近黒 `{colors.ink}` の Inter 系タイポ。見出しは 700 + 負の tracking、本文は 400 / 1.5
- 構造アクセントは `{colors.primary}` / `{colors.brand}`（`#038b94`）の 1 色のみ。第二の構造色を導入しない
- 状態・区分の色は業務識別子として design-system.md §2.4〜2.5 が所有し、構造や CTA には使わない。色だけに依存せず、文字・アイコンを必ず併置する
- Primary CTA は pill `{rounded.full}`、ユーティリティボタンは `{rounded.md}`（8px）、入力は `{rounded.xs}`（4px）— 対比は意図的
- エレベーションは hairline と極薄のレイヤードシャドウで表現し、重い drop-shadow を使わない
- タブレットファースト: タッチ面の基準は design-states.md §1.1（44×44px、dense 行アクションの例外は design-system.md §7.2）

## Colors（本書の参照範囲のみ）

色の正本は [design-system.md §2](docs/spec/design-system.md)。本節は本文で参照するトークン名の意味だけを定義し、値の決定・変更・追加はしない。

| トークン | 役割 | 値の所有 |
|---|---|---|
| `{colors.brand}` / `{colors.brand-active}` | 製品識別（認証画面・ロゴ・明示的 brand CTA） | design-system §2.1 |
| `{colors.primary}` / `{colors.primary-active}` | 汎用 Primary CTA・リンク・active タブ・選択・focus ring | design-system §2.1（塗りボタンの実装値は §7.2 の `action-primary-solid` 系） |
| `{colors.on-brand}` / `{colors.on-primary}`（active 含む） | brand / primary 面の上の文字 | design-system §2.1 |
| `{colors.secondary}` | 深 indigo。予約値。本体画面では使わない（採用する場合は design-system に決裁を記録してから） | design-system §2.1 |
| `{colors.canvas-soft}` | ページ canvas、サイドバー、テーブルヘッダ帯、空状態カード | design-system §2.2 |
| `{colors.canvas}` / `{colors.surface}` | カード・パネル・フォームフィールド・モーダル | design-system §2.2 |
| `{colors.hairline}` | 1px の境界線・行区切り | design-system §2.2 |
| `{colors.ink}` → `{colors.ink-secondary}` → `{colors.ink-muted}` → `{colors.ink-faint}` | 文字の 4 段階。faint は placeholder と非情報装飾のみ | design-system §2.3 |
| 臨床 semantic（危険・死亡・警告）、業務ステータス色、LINE グリーン | 状態の識別子 | design-system §2.4 / §2.5 |

## Typography

### Font Family

実装のフォントスタックは **`'Inter', 'Noto Sans JP'`**（フォールバック `-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`。`frontend/src/styles/globals.css`）。単一スタックで display から eyebrow まで担う。serif・monospace の表示用書体は使わない。日本語は Noto Sans JP が受け、下表の letter-spacing は欧文見出しを前提とした値であるため、和文見出しで詰まりすぎる場合も**ロールの値を上書きしない**（compliance C14）。必要ならロール自体の変更を本書の Hierarchy 表に決裁（日付・決定者）として記録する（タイポグラフィの決定は本書が所有し、design-system.md には記録しない）。

### Hierarchy

| トークン | Size | Weight | Line Height | Letter Spacing | 用途 |
|---|---|---|---|---|---|
| `{typography.display-1}` | 64px | 700 | 1.0 | −2.125px | 予約段。本体画面では使わない |
| `{typography.display-2}` | 54px | 700 | 1.04 | −1.875px | 予約段。本体画面では使わない |
| `{typography.heading-1}` | 40px | 700 | 1.1 | −1px | ページ級の大見出し。**例外的な大判画面専用**で、通常のページタイトル（`PageLayout` / `FormHeader`）は `{typography.title}` 20px（design-system §3.4、design-patterns §1.4）（`text-heading-1`） |
| `{typography.heading-2}` | 26px | 700 | 1.23 | −0.625px | 大型サイドパネルタイトル・文書型画面（`/manual/*`・単独ウィンドウ）の大見出し。本体画面のセクション見出しには使わない（2026-10-20 決裁、design-system §3.4）（`text-heading-2`） |
| `{typography.heading-3}` | 22px | 700 | 1.27 | −0.25px | ダイアログタイトル・文書型画面の見出し。本体画面のカード／セクション見出しには使わない（2026-10-20 決裁。それらは `title` 以下）（`text-heading-3`） |
| `{typography.title}` | 20px | 600 | 1.4 | −0.125px | フィーチャータイトル・callout（`text-xl`） |
| `{typography.body-md}` | 16px | 400 | 1.5 | 0 | 標準本文（`text-base`） |
| `{typography.body-sm}` | 15px | 400 | 1.33 | 0 | テーブル行・密な UI・ナビ（`text-sm`、`ex-data-table-cell` body） |
| `{typography.button}` | 16px | 500 | 1.5 | 0 | ボタンラベル |
| `{typography.caption}` | 14px | 400 | 1.43 | 0 | キャプション・注記（`text-xs`） |
| `{typography.eyebrow}` | 12px | 600 | 1.33 | +0.125px | テーブルヘッダ・小ラベル・バッジ（`text-2xs`、`ex-data-table-cell` header） |

実装クラスへの写像と採用範囲は design-system.md §3.4。上表にない段（Tailwind 既定の `text-lg` / `text-2xl` 以上、`text-[Npx]` 任意値）は compliance C11 / C12 が禁止する。

### Principles

- 表現の主レバーは **700 の見出し vs 400 の本文**のコントラストだけ。装飾タイポグラフィは使わない
- 本文を heavy weight にしない。強調は weight 600 の `title` か eyebrow で行う
- 数値は等幅化（`tabular-nums`）して桁を揃えてよいが、サイズ段は上表から選ぶ
- Inter が読み込めない環境でも上表の letter-spacing を明示適用する（既定 tracking では緩く見える）

## Layout

### Spacing System

- **基準単位**: 8px。半段として 4px を持つ
- **トークン**: `{spacing.xxs}` 4px · `{spacing.xs}` 8px · `{spacing.sm}` 12px · `{spacing.md}` 16px · `{spacing.lg}` 24px · `{spacing.xl}` 28px · `{spacing.xxl}` 32px
- カード内 padding は `{spacing.lg}`（24px）、テーブルセルは `{spacing.sm} {spacing.md}`（12px / 16px）、ユーティリティボタンは 4px / 14px、入力は 6px
- **20px はスケールにない**。`*-5`・負値・`[20px]` / `[1.25rem]` は compliance C16 が禁止する

### Grid & Container

ページ本文は `PageLayout` の中央寄せカラムに置く。最大幅は `LAYOUT.pageContentMaxWidth` のトークンから選び、`max-w-full` / `max-w-[Npx]` の生値は書かない（`PageLayout` の `maxWidth` prop は compliance C7 が機械検査する。本文・フォーム側の raw `max-w-[Npx]` は機械検査が無く、レビューで見る）。フォームの最大幅も同じトークンの既存段（`form` 1400px / `formMid` 1200px / `formNarrow` 900px）から選ぶ。より狭い段が必要になった場合は、本節に値と決裁（日付・決定者）を記録してからトークンへ追加する。他文書（design-patterns.md を含む）はフォーム幅の値を持たない。一覧系は横スクロールを許容し、識別情報・ステータス・操作列をスクロールしないと到達できない位置へ切り落とさない。

### Whitespace Philosophy

グルーピングは罫線ではなく余白で行う。セクションは大きめの縦余白で区切り、カードは暖色 canvas の上に hairline だけで置く。情報密度が高い臨床画面でも「詰める」のではなく「揃える」ことで読みやすさを保つ。

### Responsive Strategy

#### Breakpoints（タブレットファースト）

区分は連続（隙間なし）で定義する。サイドバーの折りたたみ境界は実装（`Sidebar.tsx` の `window.innerWidth < 1280` / `(max-width: 1279px)`）と同じ **1280px** である。

| 名称 | 幅 | 主な変化 |
|---|---|---|
| Wide | 1440px 以上 | 複数カラムの一覧・詳細を並置。サイドバー展開 |
| Desktop | 1280–1439px | 標準の中央寄せカラム。サイドバー展開 |
| Compact desktop | 1024–1279px | サイドバーは折りたたみ（アイコン幅）。一覧は補助列を `lg` で段階的に非表示にしてよい |
| Tablet | 768–1023px | 診察室 iPad の基準。サイドバー折りたたみ、補助列は非表示。識別・ステータス・操作列は常に到達可能 |
| Mobile | 767px 以下 | 本体画面の規定外（375px の規定は未整備。必要になった時点で本表へ決裁を追加する）。LIFF / line-reserve は別アプリとして design-system §2.6 の適用範囲外 |

#### Touch Targets

タッチ面の寸法・ヒット領域・dense 行アクションの例外は [design-states.md §1.1](docs/spec/design-states.md) と design-system.md §7.2 が正本。本書では再定義しない。ラベルが短くても縦 padding を削らない、という形状上の注意だけを置く。

#### Collapsing Strategy

サイドバーは 1280px 未満（Compact desktop 以下）で折りたたみ、一覧は補助列を `md` / `lg` で段階的に表示する。日付・識別情報・ステータス・操作の列は常に到達可能にする。

## Elevation & Depth

| Level | 扱い | 用途 | 実装トークン |
|---|---|---|---|
| 0 — Flat | hairline `{colors.hairline}` のみ、shadow なし | 既定のカード・テーブル・フォームセクション | border のみ |
| 1 — Soft | レイヤード微小シャドウ: `rgba(0,0,0,0.01) 0 0.175px 1.041px`, `rgba(0,0,0,0.02) 0 0.8px 2.925px`, `rgba(0,0,0,0.027) 0 2.025px 7.847px`, `rgba(0,0,0,0.04) 0 4px 18px` | 浮動パネル・ポップオーバー・ドロップダウン・Select・ツールチップ・secondary ボタン・入力 focus（design-system §5.1 と `popover.tsx` / `select.tsx` / `dropdown-menu.tsx` の実装に一致） | `shadow-level1` |
| 2 — Elevated | 5 段スタック。最終段 `rgba(0,0,0,0.05) 0 23px 52px` | モーダル（Dialog / AlertDialog）・Sheet / SidePeek | `shadow-level2` |

トースト（`sonner.tsx`）は現状ライブラリ既定の影で、elevation トークンを当てていない（既知のドリフト。当てるなら Level 1）。

エレベーションは **barely-there** を原則とする。多層の透明に近い影で面が紙から静かに浮く印象を作り、硬い影を落とさない。Tailwind 既定の `shadow-sm〜2xl`・`shadow-[...]` 任意値・CSS 直書きの `box-shadow` は compliance C10 / C17 が禁止し、elevation トークン（`shadow-level1` / `shadow-level2` / `shadow-btn` / `shadow-panel` / `shadow-focus-brand` 等）だけを使う。

## Motion

動きは**状態の変化を知らせるためだけ**に使い、装飾として使わない。値は実装（`tw-animate-css` の `animate-in` / `animate-out`、Tailwind の `transition-*`）に合わせて次の 4 段に固定する。

| 段 | 時間 | 用途 | 実装 |
|---|---|---|---|
| micro | 150ms | hover / focus の色・不透明度、ツールチップの出現 | `transition-colors`（既定 150ms）、`tooltip.tsx` の `duration-150` |
| overlay | 200ms（**未実装**） | ダイアログ・ポップオーバー・ドロップダウン・Select の出現（fade + zoom 95%）と消失。実装には `animate-in` / `fade-in-0` / `zoom-in-95` が書かれているが、`tw-animate-css` が未導入のためこれらのクラスは効かず、現状は即時表示。導入して有効化するか、クラスを除去して「overlay は動かない」と確定するかは決裁待ち | `data-[state=open]:animate-in fade-in-0 zoom-in-95` + `duration-200` |
| panel | 300ms（閉）/ 500ms（開） | Sheet / SidePeek のスライド、サイドバーの折りたたみ（300ms） | `sheet.tsx`、`STYLE.sidebarContainer` の `duration-300` |
| indeterminate | 1.5s ループ | 読込中のプログレス帯、スピナーは連続回転 | `.animate-progress-indeterminate`、`Loader2 animate-spin` |

- トーストの自動消去は **4,000ms**（`sonner.tsx`）。エラートーストも同じ時間で消えるため、操作が必要なエラーはトーストではなくインラインのエラー文で出す（design-states §1 invalid）。
- 位置が動くアニメーション（スライド・ズーム・カードの移動）は `prefers-reduced-motion: reduce` で止める。判定は共有フック `useReducedMotion`（`src/hooks/use-reduced-motion.ts`。参照実装 `WeekViewAppointmentCard`）か `motion-reduce:` を使う。色・不透明度の遷移は止めなくてよい。
- 入力中・保存中にレイアウトが動くアニメーション（高さが変わる展開、行の挿入でのスライド）を本文に置かない。読込はスケルトンで高さを先に確保する（[design-patterns.md §6.2](docs/spec/design-patterns.md)）。
- 上表にない時間（`duration-700`、任意値 `duration-[Nms]`）と、無限ループの装飾アニメーションは使わない。

## Shapes

### Border Radius Scale

| トークン | 値 | 用途 |
|---|---|---|
| `{rounded.xs}` | 4px | フォーム入力（`text-input`）・小タグ・インラインチップ。`--radius-xs`（互換名 `--radius-xxs` も同値） |
| `{rounded.sm}` | 5px | メニュー項目・リスト行・サイドバー行・status pill |
| `{rounded.md}` | 8px | ユーティリティボタン・小カード |
| `{rounded.lg}` | 12px | フィーチャーカード・コンテンツタイル・画像枠 |
| `{rounded.xl}` | 16px | 大コンテナ・モーダル・空状態カード・トースト |
| `{rounded.full}` | 9999px | Primary / Secondary CTA の pill・バッジ・円形アイコンボタン・アバター |

`rounded-[Npx]` の任意値は compliance C9 が禁止する。入力に pill を使わない（入力は常に `{rounded.xs}`）。

## Components

> 本節は**形状・寸法・タイポの基本様式**を定義する。色の実装値は design-system.md §7、状態（hover / focus-visible / disabled / loading / invalid）は design-states.md §1、部品ごとの契約は components/ が正本。

### Navigation

**`nav-bar`** — サイドバー / アプリシェルのナビ
- 面は `{colors.canvas-soft}`（本文カードとの図地関係を保つため白にしない — design-system §7.1 FE11 決裁）。リンクは `{colors.ink}` の `{typography.body-sm}`、padding `{spacing.md}`。1280px 未満で折りたたむ（Breakpoints 表）。
- 項目名は省略しない。現在地（`aria-current="page"`）は初期表示で見える位置へスクロールする（[design-patterns.md §1](docs/spec/design-patterns.md)）。

**`ex-app-shell-row`** — サイドバーのナビ行
- `{rounded.sm}`、padding `{spacing.sm} {spacing.md}`。active は `{colors.primary}` のインジケータと文字色で示し、色以外に太字または左端バーを併置する。

### Buttons

**`button-primary`** — 汎用 Primary CTA（保存・登録・確定）
- `{typography.button}`、pill `{rounded.full}`、文字 `{colors.on-primary}`。塗りの実装値は design-system §7.2（`action-primary-solid` 系。`#038b94` 地に白文字は AA 非達のため塗りには使わない）。
- 押下は `button-primary-pressed`（`{colors.primary-active}` 系）。focus-visible は 2px ring + offset を必ず残す（design-states §1.1）。
- 1 画面（1 ダイアログ）に Primary は原則 1 つ。

**`button-brand`** — 認証・製品識別 CTA（ログイン等）
- 形状・タイポは `button-primary` と同じ。塗りの実装値は design-system §7.2 の `brand-identity-solid` 系（`C.bgBrandIdentitySolid` = `#027078`、白文字 5.85:1。hover / active は `*BrandIdentitySolid` 系 `#025F66`）。`{colors.brand}`（`#038b94`）は白文字と 4.10:1 で AA 非達のため**塗りには使わず**、ロゴ・装飾・focus ring など非テキスト用途に限る（design-states §2.3 決裁。`SubmitButton` の brand 分岐が参照実装）。`colorVariant="brand"` を明示する（compliance C5）。

**`button-secondary`** — 二次 CTA（キャンセル・戻る・別アクション）
- `{colors.surface}` 面、`{colors.ink}` 文字、`{typography.button}`、pill `{rounded.full}`、Level-1 shadow。Primary の隣に置く。薄い面上の hover で文字を白くしない。

**`button-utility`** — ナビ / ツールバー / 行アクションのユーティリティ操作
- `{colors.surface}` 面、`{colors.ink}` 文字、`{typography.button}`（行内では `{typography.body-sm}`）、`{rounded.md}`（8px）、padding `4px 14px`、1px `{colors.hairline}` border。pill が大きすぎる場所に使う。アイコン単独の場合は必ずアクセシブルな名前を付ける。

### Cards & Containers

**`feature-card`** — フォームセクション・情報カード
- `{colors.surface}`、`{colors.ink}`、`{typography.body-md}`、`{rounded.lg}`（12px）、padding `{spacing.lg}`（24px）、Level-0（hairline のみ）。業務画面の基本単位。

**`feature-card-elevated`** — 浮動パネル
- `feature-card` と同じ chrome に Level-1 shadow。サイドピークパネルや canvas から浮かせたい面に使う。

**`badge-pill`** — ステータス・区分バッジ
- `{typography.eyebrow}`（12px / 600）、pill `{rounded.full}`、padding `4px 8px`。色の組み合わせは `BADGE.*`（design-states §2.1 の実測コントラスト表）から選ぶ。バッジは識別用でありクリック操作を持たせない。色だけで意味を伝えず、文字を必ず入れる。

### Inputs & Forms

**`text-input`** — テキスト / 数値 / 選択トリガー
- `{colors.surface}` 面、`{colors.ink}` 文字、`{typography.body-sm}`、1px border、`{rounded.xs}`（4px）、padding `6px`。角は CTA より意図的に角ばらせる。focus-visible で Level-1 shadow + ring。invalid は destructive border + エラーテキスト（design-states §1）。readonly は disabled と区別し、文字色を落とさず border を hairline 化する。必須項目は必須マークを示す。

### Data Table

**`ex-data-table-cell`** — 業務 data-table の th / td 基本様式
- ヘッダ: `{colors.canvas-soft}` 帯 + `{typography.eyebrow}`（12px / 600 / +0.125px）。本文: `{typography.body-sm}`。セル padding `{spacing.sm} {spacing.md}`（縦 12px / 横 16px）。行区切り `{colors.hairline}`。
- 実装は `STYLE.tableHeaderRow` / `STYLE.tableHeaderCell` と `TableHead` / `TableCell` が保証する。呼び出し側で typography や padding を上書きしない（compliance C18）。行全体に `onClick` を付けず、遷移は cell 内の native link、操作は native button、並べ替えは 44px のドラッグハンドルで行う（compliance C19）。
- 空状態の行だけ `data-empty-state` + `colSpan` + `text-center` で縦 padding を追加できる。
- 列の役割（識別列は折り返さない・長文列は `max-width` + 全文到達手段・操作列は右端 sticky）、横スクロール時に残す列、空値 `—`・件数の表記は [design-patterns.md §2](docs/spec/design-patterns.md) が正本。

### Dialogs, Toasts, Empty States

**`ex-modal-card`** — モーダルダイアログ
- `{colors.surface}`、`{rounded.xl}`、padding `{spacing.lg}`、Level-2 shadow。タイトル・説明・focus 復元・ESC を持ち、閉じる操作そのものに業務上の取消・保存・削除の意味を与えない（components/dialog.md）。

**`ex-auth-form-card`** — 認証フォームカード
- `feature-card` の chrome（`{rounded.xl}`、padding `{spacing.lg}`）に `text-input` と `button-brand` を内包する。

**`ex-empty-state-card`** — 空状態
- `{colors.canvas-soft}`、`{rounded.xl}`、padding `{spacing.xxl}`、キャプション `{typography.body-md}`。中に書く内容（何が無いか／条件／次の操作）は [design-patterns.md §6.1](docs/spec/design-patterns.md#empty-state) が正本。

**`ex-toast`** — トースト通知
- `{colors.surface}`、`{rounded.xl}`、padding `{spacing.sm} {spacing.md}`、`{typography.body-sm}`、Level-1 shadow。成功・失敗はアイコンと文言で区別し、色だけに頼らない。

### Animal Ekarte 固有パターン

形状・寸法は上記の基本様式に従い、振る舞いは design-system.md §7.6 と各部品仕様が定義する: `PropertyInput`（ボーダーレスのプロパティ編集）、`SidePeekPanel`（文脈を保つサイドピーク）、`FilteringIndicator`（非同期フィードバック）。

## Do's and Don'ts

### Do

- ページ canvas は暖色 `{colors.canvas-soft}`、カード・フィールドは白 `{colors.surface}` にして図地関係を作る
- 構造色は `{colors.primary}` / `{colors.brand}` の 1 色に限定し、汎用操作は primary、製品識別は brand のトークン名で意味を分ける
- 見出し・本文・ラベルは Hierarchy 表のロールから選び、実装クラスは design-system §3.4 の写像を使う
- Primary CTA は pill `{rounded.full}`、ユーティリティボタンは `{rounded.md}`、入力は `{rounded.xs}` — 対比を保つ
- 既定のカードは hairline の Level-0、浮動面だけを Level-1、モーダルだけを Level-2 にする
- 臨床 sentinel・ステータスは色 + 文字（+ アイコン）で示す
- タッチ面は design-states.md §1.1 の基準に従う

### Don't

- 第二の構造アクセントを導入しない。ステータス色・LINE グリーンを CTA や構造フィルに使わない
- 全ページを純白にしない
- フォーム入力に pill を使わない
- Tailwind 既定の shadow・任意値 shadow・CSS 直書き box-shadow を使わない
- 本文を heavy weight にしない。ロールの letter-spacing を `tracking-*` で上書きしない
- スケールにないサイズ（`text-lg` / `text-2xl` 以上・`text-[Npx]`）、角丸（`rounded-[Npx]`）、余白（20px 系 `*-5`）を使わない
- コンポーネントで hex / rgba / 名前付き色を直書きしない。`hover:${C.x}` のような実行時クラス合成をしない（`design-tokens.ts` と静的クラス文字列だけを使う）
- 危険・死亡・RBAC 非活性などの臨床安全 UI を、見た目の変更で退行させない
- 本書に色の決定を書かない。色は design-system.md へ決裁（日付・決定者）を記録してからトークンへ流す
