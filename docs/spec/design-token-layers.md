# デザイントークン層構造 (Token Layer Architecture)

> **目的**: `design-tokens.ts` / `globals.css` の値を「Primitive → Semantic → Component」の3層モデルとして正式に位置づけ、どの層が何を参照し、新しい値をどこに足すかを一意に決める。Sparkle Design の4層モデル（primitive / semantic / component / applied）を本リポジトリの実装事情に写像したもの。
> **読者**: フロントエンド実装者・レビュアー・DS 保守者。
> **タイミング**: トークン追加・色変更・新規 UI プリミティブ作成時。
> **SSOT 関係**: 色の値は [design-system.md](design-system.md) が正本、タイポ/形状/寸法は [DESIGN.md](../../DESIGN.md) が正本。**本書は「配置と参照方向」の正本**であり、値そのものは規定しない。
> **最新更新**: 2026-09-30（初版 — FE-DS1）

---

## 1. 層モデル

```
L1 Primitive ──参照（値の唯一の出所）──→ L2 Semantic ──合成──→ L3 Component ──→ 画面 UI
   PALETTE / :root raw vars              C / :root semantic vars    STYLE / BADGE / ICON /
   （生の hex・rgba）                    （役割名つき単機能クラス）    LAYOUT / *-variants.ts
```

| 層 | 実体 | 例 | 変更理由 |
|---|------|-----|----------|
| **L1 Primitive** | `PALETTE`（`src/lib/design-tokens.ts` §1）+ `globals.css :root` の生値変数 | `#038B94`, `#EBECED`, `rgba(0,0,0,0.04)` | 製品パレットの変更のみ。呼び出し側の意味を持たない |
| **L2 Semantic** | `C`（`design-tokens.ts` §2）+ `globals.css` の意味名変数（`--primary`, `--border`, `--muted` 等）+ `@theme inline` 経由の Tailwind 色 | `C.bgActionPrimary`, `C.textInkMuted`, `C.borderLight` | 「汎用 CTA の hover 色」のような役割判断 |
| **L3 Component** | `STYLE` / `BADGE` / `ICON` / `LAYOUT` / `Z` / `TABLE_STYLES` + `components/ui/*-variants.ts`（cva） | `STYLE.tableHeaderCell`, `BADGE.green`, `buttonVariants` | コンポーネントの見た目仕様。design-system.md §7 / [design-states.md](design-states.md) が正本 |

## 2. 参照方向の規則（Dependency Rules）

1. **下向きのみ**: L3 は L2 の member を合成してよい。L3→L1・L2→画面・画面→L1 の参照は原則禁止。
   - 許容例外: `PALETTE` の生値は `style` prop / canvas 描画 / third-party への色渡しなど**文字列の生値が必要な箇所に限り**コンポーネントから参照可。className 合成用の値として新規参照する場合は L2 member を先に作る。
2. **L2 member は値をリテラルで持つ**（重要・後述 §3）。`C.text` が `"text-[#000000]"` と hex を内包するのは意図的。
3. **新しい値の流入経路**: 色 = design-system.md に決裁記録 → `PALETTE`/`:root` に L1 → `C`/semantic var に L2 → 必要なら `STYLE`/`*-variants` に L3。タイポ・寸法 = DESIGN.md → `globals.css @theme` / `LAYOUT`。
4. **削除も層の順で**: L3 消費 → L2 member → L1 値。L1 を消す前に L2 の全 member が退避済みであること。

## 3. 実装制約（なぜ L2 が生値を持つのか）

Tailwind v4 は任意値クラス（`text-[#31302E]` 等）を**ソースを走査した静的文字列**からしか生成しない。`C.text = \`text-[${PALETTE.primary}]\`` のような動的参照は CSS が出力されず壊れる。

したがって **L2 は L1 の値をテキストとして重複して保持する**のが構造上の正しい形であり、これは欠陥ではない。層の一致は以下で担保する:

- 色の値一致 = design-system.md §2.6 マッピング表（PALETTE ↔ C ↔ :root）
- 機械検証 = `pnpm design-audit`（C1–C19。新規 hex 直書き C3 / rgba 直値 C6b / 非仕様値 C11–C16 が L1 外への値混入を検出）

> Sparkle の CSS 変数 4 層は `--color-*` が変数参照で層を串刺しにする方式。本プロジェクトは Tailwind 静的走査制約により TS 側では「値の重複 + 監査による一致保証」を採る。両者は等価な層モデルの異なる実装である。

## 4. 層の編集者と消費者

| やりたいこと | 編集先 | 消費方法 |
|---|---|---|
| 新しい生色値 | `PALETTE` or `:root` raw var | 直接 style prop / canvas のみ |
| 汎用的な色の役割 | `C` member（`textInk` / `bgActionPrimary` 系の**意味名**で命名） | `className` に入れる |
| 繰り返す見た目パターン | `STYLE` / `BADGE` / `ICON` / `LAYOUT` | `className` に入れる |
| UI プリミティブの variant | `components/ui/*-variants.ts`（cva） | props で選ぶ |
| テーマ全体の Tailwind utility 化 | `globals.css @theme inline` | `bg-primary` 等の utility |

命名: 新規 L2 member は**役割名**を使う（`textInkMuted`, `bgActionPrimary`）。`text90` / `bgBrand` 系の値由来・旧名は互換用に残るが新規使用は非推奨（§6 の命名ラチェット項目を参照）。

## 5. 既知の層リーク（記録済み・別タスクで整理）

| 箇所 | 内容 | 扱い |
|---|---|---|
| ~~`PALETTE.hoverBgInput` / `focusBorderLegacyAccent` / `focusRingActionPrimary` / `focusRingBrand`~~ | 「raw values」と宣言された L1 に Tailwind クラス文字列が混入 | **解決済み — `STATE` バケット（L2.5）へ移設**。PALETTE は raw 値のみ、新規混入は C22 が hard fail |
| `C` の `text90`–`text15`（旧 ink 14段） | FE11-F2 で値は4段に字義化済みだが名前が残存 | 新規は `textInk*` 4エイリアスのみ使用。機械化は §6 の命名ラチェット項目（未実装） |
| `bgStatusBlueDot` 等の `bg-blue-500` 系 Tailwind 名前付き色 | L2 member が Tailwind パレットを直接指す | 業務ステータス色の製品決定として許容（design-system.md §2.5）。新規追加時は値を L1 に固定するか個別判断 |

## 6. 監査ルール（`scripts/design-system-audit.mjs`）

実装済み:

| ID | 内容 | 状態 |
|---|---|---|
| C21 | `src/components/ui/*.tsx` に colocated `*.stories.tsx` を要求（Storybook カタログの網羅担保） | ✅ 実装済み。全25 primitive が stories 保持、`C21_STORIES_ALLOWLIST` は空 |
| C22 | `PALETTE` に Tailwind クラス文字列メンバを追加することを禁止（生値のみ許可） | ✅ 実装済み。既存4メンバ（`dragOverlayShadow` / `brandGlow` / `primaryGlow` / `tableRowHover`）は `C22_PALETTE_CLASS_MEMBER_ALLOWLIST` で移行猶予 |
| C23 | `BADGE.*` コンボの text/bg を解決して WCAG コントラスト ≥4.5:1 を機械検証 | ✅ 実装済み（design-states.md §2 是正済み値で全パス） |
| C24 | `src/components/shared/` 層にも stories を要求（dir 単位 / トップレベルは sibling） | ✅ 実装済み。既存未カバー49エントリは `C24_SHARED_STORIES_ALLOWLIST` で棚卸し — **stories 追加時にエントリを削るラチェット運用（増やさない）** |

保留（未実装）:

| 検討 | 内容 | 保留理由 |
|---|---|---|
| 命名ラチェット | `C` の値由来名 member（`text[0-9]+` 等）の新規消費を検出 | 現状 1664 箇所の既存消費と新規消費を静的走査で区別できない。baseline ratchet（git diff 基盤）か allowlist 化が先に必要 |

> 注: 当初の提案番号 C20–C22 は、監査スクリプトで C20 が「runtime Tailwind synthesis」に既に使用されていたため、stories=C21 / PALETTE純粋性=C22 として実装した。
