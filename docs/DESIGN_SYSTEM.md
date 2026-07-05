# UI デザインシステム 規約 (Design System)

> **目的**: DESIGN.md（ルート — Notion Analysis / 唯一の意匠仕様の源泉）を実装に落とし込むための規約を定義する。
> **読者**: フロントエンド実装者。
> **タイミング**: UI 実装・レビュー時。
> **真実の源泉**: 本書は [DESIGN.md](../DESIGN.md) の下位文書であり、値・原則が矛盾する場合は常に DESIGN.md が優先する。実装 SSOT は `frontend/src/lib/design-tokens.ts`（`C` / `STYLE` / `PALETTE` / `LAYOUT`）。

> **Animal Ekarte**: Notion ライクな体験を臨床現場へ
> **最新更新**: 2026-07-05

---

## 1. デザイン思想

本システムの UI は、**「臨床現場での集中力を妨げないクリーンな情報設計」**を最優先としている。DESIGN.md の言葉を借りれば、「良い光の下の整理されたデスク」— 暖色の紙のようなキャンバスの上に、抑制されたモノクロ＋単一ブルーのUIが乗り、装飾は必要な場所（スティッカーパレット・意味的カラー）にのみ許可される。

### 核心となる 3 つの原則
- **ミニマリズム**: 装飾を削ぎ落とし、データそのものを主役にする。構造色は `{colors.primary}` 一色のみとし、二番目の構造アクセントを導入しない。
- **直感的な階層**: Notion スタイルの「プロパティ」と「カード」による情報の整理。カードは白 surface + hairline、ページ canvas は暖色オフホワイト。
- **臨床安全の可視化**: 警告や異常値のみを、戦略的に配色された「意味的カラー」で強調する。危険バッジ・死亡グレーアウト・RBAC 非活性表示など、臨床安全 UI は本デザイン変更で退行させない。

---

## 2. カラーパレット

> 値は [DESIGN.md](../DESIGN.md) `colors:` フロントマターが正。以下は `design-tokens.ts` とのマッピング表。

### 2.1 構造色・基礎カラー

| 役割 | DESIGN.md トークン | 値 | design-tokens.ts |
|---|---|---|---|
| 構造ブルー（Primary CTA・リンク・active/focus のみ） | `{colors.primary}` | `#0075DE` | `PALETTE.brand`, `C.bgBrand` / `C.textBrand` / `C.borderBrand` |
| 構造ブルー Pressed | `{colors.primary-active}` | `#005BAB` | `PALETTE.brandHover`, `C.hoverBgBrand` |
| ページ canvas（暖色オフホワイト） | `{colors.canvas-soft}` | `#F6F5F4` | `PALETTE.bgMain`, `C.bgPage` |
| カード・フィールド surface | `{colors.canvas}` / `{colors.surface}` | `#FFFFFF` | `C.bgWhite` / `C.bgSubtle` |
| Hairline（区切り線・カード境界） | `{colors.hairline}` | `#E6E6E6` 相当 | `C.borderLight`（`rgba(0,0,0,0.09)` — hairline の実装上の等価トークン） |
| Ink（本文・見出し） | `{colors.ink}` | `#000000` | `C.text` |
| Ink secondary | `{colors.ink-secondary}` | `#31302E` 相当 | `C.text70` |
| Ink muted | `{colors.ink-muted}` | `#615D59` 相当 | `C.text60` |
| Ink faint（caption/placeholder） | `{colors.ink-faint}` | `#A39E98` 相当 | `C.text40` |

> **既知の技術的負債**: `C.accent` / `C.bgAccent` 系（`#2383E2`）は DESIGN.md 導入以前の旧 Notion Analysis ブルーであり、`C.bgBrand`（`#0075DE`）が DESIGN.md 準拠の現行構造色である。両者は当面共存するが、**新規実装は必ず `brand` 系を使うこと**。`accent` 系のアプリ全体一括移行は本書の範囲外（影響範囲が広いため別タスクで計画的に実施）。

### 2.2 意味的カラー (Semantic Colors)
臨床上の判断を助けるための特別な配色ルール。DESIGN.md の Notion Analysis は装飾用スティッカーパレット（sky/purple/pink/orange/teal/green/brown）を定義するのみで専用の semantic ramp を持たないため、本システムでは独自に意味的カラーを定義し、`design-tokens.ts` で一元管理する。
- **危険 / 死亡 / 異常高**: `C.danger`（`#C0392B`、WCAG AA 7.1:1 準拠）
- **注意 / 期限間近**: `C.textWarning` / `C.bgWarning50`
- **正常 / 完了 / 生存**: `C.textStatusGreen` / `C.bgStatusGreen`
- **異常低 / 寒冷**: `C.textStatusBlue` 系

### 2.3 スティッカーパレット（装飾専用）
DESIGN.md の Do's/Don'ts に従い、以下は **CTA・構造フィルには使用しない**。バッジ・タグ・凡例ドットなど装飾用途に限定する。
- Sky `#62AEF0` 相当、Purple `#D6B6F6`、Pink `#FF64C8`、Orange `#DD5B00`、Teal `#2A9D99`、Green `#1AAE39`、Brown `#523410`
- 実装: `C.dotBlue` / `C.dotPurple` / `C.dotPink` / `C.dotOrange` / `C.dotGreen` / `C.dotBrown` などの dot / badge 系トークン。
- **外部ブランド例外**: LINE 公式グリーン `#06C755`（`PALETTE.lineGreen`）は外部ブランド識別のため構造色ルールの対象外。ただし構造・CTA には使わない。

---

## 3. タイポグラフィ

DESIGN.md の `typography:` フロントマターに準拠。フォントファミリーは `NotionInter`（フォールバック: Inter）。

| DESIGN.md トークン | サイズ / Weight | 用途 |
|---|---|---|
| `{typography.heading-1}` 〜 `{typography.heading-3}` | 40px/700 〜 22px/700 | ページ見出し・セクション見出し |
| `{typography.title}` | 20px/600 | カードタイトル |
| `{typography.body-md}` | 16px/400 | 標準本文 |
| `{typography.body-sm}` | 15px/400 | テーブル行・密なUI（`ex-data-table-cell` body） |
| `{typography.button}` | 16px/500 | ボタンラベル |
| `{typography.caption}` | 14px/400 | キャプション・注記 |
| `{typography.eyebrow}` | 12px/600, +0.125px | データテーブルヘッダー・小ラベル（`ex-data-table-cell` header） |

タブレットファースト運用のため、Tailwind v4 `@theme inline` で全体のクリックターゲット・フォントサイズを標準より約 10% 拡大している（診察室での iPad 利用を想定）。

---

## 4. 形状・角丸スケール

DESIGN.md `rounded:` フロントマターに準拠。**コンポーネントの種別ごとに角丸値が意味を持つ**ため、統一した「1つの角丸」に寄せない。

| トークン | 値 | 用途 |
|---|---|---|
| `{rounded.xs}` | 4px | フォーム入力（`text-input`）・小タグ |
| `{rounded.sm}` | 5px | メニュー項目・リスト行 |
| `{rounded.md}` | 8px | ユーティリティボタン（`button-utility`）・小カード |
| `{rounded.lg}` | 12px | フィーチャーカード |
| `{rounded.xl}` | 16px | 大コンテナ・モーダル（`ex-modal-card`） |
| `{rounded.full}` | pill | Primary CTA（`button-primary`）・バッジ |

**Don't**: フォーム入力に `{rounded.full}`（pill）を使わない。入力は `{rounded.xs}` に留める。

---

## 5. エレベーション

DESIGN.md 準拠：ハードシャドウは使わず、hairline + 複数レイヤーの極薄シャドウで「紙から少し浮いている」質感を出す。

| Level | 用途 |
|---|---|
| 0 — Flat（hairline のみ） | 通常カード |
| 1 — Soft（極薄レイヤードシャドウ） | 浮動ボタン・フォーカス中の入力 |
| 2 — Elevated（深めの5段シャドウ） | モーダル・ポップオーバー |

shadcn `DialogContent`（`frontend/src/components/ui/dialog.tsx`）は `rounded-xl` + `p-6`（`{spacing.lg}`）+ `shadow-lg` を既定とし、これは DESIGN.md `ex-modal-card` の仕様（`{rounded.xl}` / `{spacing.lg}` padding / Level-2 shadow）に一致する。

---

## 6. UI コンポーネント・パターン（DESIGN.md マッピング）

| DESIGN.md コンポーネント例 | 本システムでの実装 | 備考 |
|---|---|---|
| `ex-data-table-cell`（header: canvas-soft + eyebrow、body: body-sm、hairline 区切り） | `OwnersListTable` / `PetSelectionResultsTable` / `OwnerPetsSection` / `HistoryTable`（owner-report） / `AggregationOwnerTable` | header は `C.bgPage`（canvas-soft）+ eyebrow 相当タイポグラフィ、body 行は `STYLE.tableCell`（body-sm 相当）、行区切りは `C.borderLight` / `C.divideDivider`（hairline の実装上の等価トークン） |
| `feature-card` / `ex-auth-form-card` | フォームセクションカード（`OwnerForm` の飼主情報カード、`PetSelectionSearchForm` など） | white surface + `rounded-lg` + `p-4`〜`p-6` |
| `ex-modal-card` | `PetEditModal` / `OwnerSearchModal` / `LstepTagAddDialog` 等、shadcn `Dialog` ベースの全モーダル | `rounded-xl` + `p-6` + `shadow-lg`（既定で準拠） |
| `text-input` | `Input` / `Textarea` / `SelectTrigger`（`components/ui/`） | `{rounded.xs}`（4px）に準拠済み（2026-07-05）。`--radius-xs: 4px` を `globals.css` に追加し、3 プリミティブの角丸を `rounded-md`（8px）から `rounded-xs` へ変更 |
| `button-primary` | Primary CTA（保存・登録ボタン） | brand ブルー `#0075DE` + pill（`{rounded.full}`）に準拠済み（2026-07-05）。共有 `SubmitButton` / `PrimaryButton` の既定 `colorVariant` を `"brand"` に変更し、feature 層からの旧 accent 上書きを撤去 |
| `button-utility` | ユーティリティボタン（8px 角丸） | `{rounded.md}` に既定で一致 |
| `badge-pill` | ステータスバッジ | `BADGE.*`（pill 形状・eyebrow 相当タイポグラフィ） |

### 6.1 プロパティ編集 (`PropInput`)
通常はプレーンテキストのように見え、マウスホバーやフォーカスで初めて入力枠が現れる、ボーダーレスな入力体験。

### 6.2 サイドピークパネル (`SidePeekPanel`)
一覧画面から詳細情報を「覗き見る」ためのサイドスライド形式。文脈（コンテキスト）を維持したまま編集が可能。

### 6.3 非同期フィードバック (`FilteringIndicator`)
大規模データの検索中、UI をフリーズさせることなく「計算中」であることを透過アニメーションで表現する。

---

## 7. デバイス最適化 (タブレット・ファースト)

診察室での iPad 利用を想定し、Tailwind v4 の `@theme inline` により全体的なクリックターゲットとフォントサイズを標準より **10% 拡大** している。タッチ操作の精度を高めつつ、高い可読性を確保する。

---

## 8. Do's and Don'ts（DESIGN.md 準拠）

### Do
- `{colors.primary}`（`#0075DE`）は Primary CTA・インラインリンク・active/focus のみに使う。
- ページ canvas は暖色 `{colors.canvas-soft}`、カード・フィールドは白 `{colors.surface}` にする。
- スティッカーパレットはバッジ・タグ・凡例ドットなど装飾用途にのみ使う。
- フォーム入力は `{rounded.xs}`、Primary CTA は `{rounded.full}`、ユーティリティボタンは `{rounded.md}`。
- カード境界は hairline + Level-1 の極薄シャドウで表現する。

### Don't
- スティッカーパレットの色を CTA・構造フィルに使わない。
- `{colors.primary}` 以外の第二構造アクセントを新規導入しない。
- フォーム入力に pill（`{rounded.full}`）を使わない。
- 重いドロップシャドウを使わない。
- 危険バッジ・死亡グレーアウト・RBAC 非活性表示など臨床安全 UI をデザイン変更で退行させない。

---

## 9. 技術的 SSOT

デザインに関する全ての定数は、以下のファイルで一元管理されている。
- **`frontend/src/lib/design-tokens.ts`**: `PALETTE`（raw hex）、`C`（Tailwind クラス）、`STYLE`（複合クラスプリセット）、`LAYOUT`（寸法）、`BADGE`（バッジ配色コンボ）、`ICON`（アイコンサイズ）。
- 新しい色を追加する場合は、必ず `design-tokens.ts` に追加した上でコンポーネントから参照する。コンポーネント内での hex 直書きは禁止。

---
