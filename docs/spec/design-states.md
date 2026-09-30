# インタラクション状態設計 (Interaction States)

> **目的**: 全インタラクティブ要素の状態（hover / focus / disabled / loading / invalid 等）の視覚仕様とトークンを一元化する正本。WCAG 2.2 AA を担保ラインとする。
> **読者**: フロントエンド実装者・レビュアー。
> **タイミング**: インタラクティブな UI を作る・触る全作業。
> **SSOT 関係**: 色値は [design-system.md](design-system.md)、寸法・形状は [DESIGN.md](../../DESIGN.md)、配置層は [design-token-layers.md](design-token-layers.md)。本書は**状態の見え方と必須実装**を規定する。
> **最新更新**: 2026-09-30（初版 — FE-DS1。実測コントラスト監査を含む）

---

## 1. 状態マトリクス（全インタラクティブ要素の共通仕様）

| 状態 | トリガー | 視覚仕様 | 実装トークン | 必須 |
|---|---|---|---|---|
| **default** | — | 各コンポーネント仕様（design-system.md §7） | — | ○ |
| **hover** | pointer hover | 塗り surface → 一段暗く / 薄い surface → `rgba(0,0,0,0.04)` 系 | `C.hoverBg*`, `C.hoverText*` | ○ |
| **active / pressed** | 押下中 | primary 系は `#027078`（primary-active） | `C.activeBgActionPrimary` 等 | ○ |
| **focus-visible** | キーボードフォーカス | **2px ring `#038B94` + offset-1**（塗り surface 上は offset で可視維持） | `focus-visible:ring-2` + `C.focusVisibleRingActionPrimary` + `ring-offset-1` | **○ 絶対** |
| **focus**（ポインタ起因） | クリック等 | インジケータを出さない（`outline-none` のみ） | — | ○ |
| **disabled** | `disabled` | `opacity-50` + `pointer-events-none`。入力系は `cursor-not-allowed` も | コンポーネント cva base | ○ |
| **readonly** | `readOnly` | disabled と区別: **ink を落とさない**、border を hairline 化。値は読める状態を維持 | `C.borderLight` | ○ |
| **loading** | 非同期処理中 | スピナー + `aria-busy` + 操作抑止（`disabled` 相当） | `Button loading` prop（§3） | ○ |
| **invalid / error** | バリデーション | border destructive + 2px ring `#C0392B`/30 + エラーテキスト `C.danger` | `aria-invalid:border-destructive`, `STYLE.formInputError` | 入力系で ○ |
| **selected / checked** | Radix `data-state` | primary bg `#038B94` + on-primary text | `C.dataCheckedBgActionPrimary` 等 | 該当部品で ○ |

### 1.1 ハードルール

- **focus-visible の除去禁止**: `outline-none` / `ring-0` を書いたら必ず代替インジケータを書く。過去の障害 EMR-207（Button 全 variant でフォーカス不可視）を根拠とする回帰防止。キーボード起因かどうかは `focus-visible:` のみで判定し、`focus:` にインジケータを載せない。
- **非色 cue**: 臨床 sentinel（死亡・危険・期限超過）とステータスは**色だけに依存しない** — バッジ文字・アイコン・テキストを必ず併置する（色覚多様性 + モノクロ印刷対策）。frontend/CLAUDE.md 臨床安全境界 1 と連動。
- **薄い surface 上の hover で ink を維持**: `hover:text-white` のような primary variant 由来の白文字化を薄背景に持ち込まない（design-system.md §7 冒頭の一般化済みルール）。
- **タッチターゲット**: tablet 以下の全タッチ面 44×44px（`min-h-11 min-w-11`）。dense 行アクション例外は design-system.md §7.2。

## 2. コントラスト実測監査（2026-09-30 計測）

WCAG 2.2 AA: 通常文字 4.5:1 / 大文字（18pt+ または 14pt bold+）3:1 / 非文字 UI 3:1。

### 2.1 バッジ色コンボ（`BADGE.*` — 実測値）

| コンボ | text / bg | 比率 | AA 通常文字 | 主な用途 |
|---|---|---:|---|---|
| `BADGE.blue` | `#183B56` / `#D3E5EF` | **9.02:1** | ✅ | 受付済・入院中 |
| `BADGE.purple` | `#6940A5` / `#EEE0F7` | **5.78:1** | ✅ | 診療中・予防接種 |
| `BADGE.green` | `#0F7B6C` / `#DDEDEA` | 4.27:1 | ⚠️ あと0.23不足 | 完了・健康 |
| `BADGE.red` | `#E03E3E` / `#FFE2DD` | 3.49:1 | ❌ | 会計待ち・在庫切れ |
| `BADGE.muted` | `#787774` / `#F1F1EF` | 3.96:1 | ❌ | デフォルトフォールバック |
| `BADGE.orange` | `#D9730D` / `#FAEBDD` | 2.81:1 | ❌ | トリミング進行中 |
| `BADGE.yellow` | `#C29243` / `#FDECC8` | 2.41:1 | ❌ | 依頼中・在庫僅少 |
| `BADGE.gray` | `#9B9A97` / `#EBECED` | **2.38:1** | ❌ | **死亡・確定済・無効 — 臨床 sentinel を含む最重要箇所** |

> §2.1 は是正前（2026-09-30 実測）のベースライン。text 値は §2.3 で `textBadge*` へ差替え済み — 背景色は据え置き。primary CTA の white-on-teal のみ未適用（製品判断事項）。

### 2.2 その他の実測

| 箇所 | 比率 | 判定 |
|---|---:|---|
| `C.danger` `#C0392B` on white | 5.44:1 | ✅（design-system.md の 7.1:1 記述は既知のドリフト — 実測 5.44） |
| ink-muted `#615D59` on canvas `#F6F5F4` | 5.99:1 | ✅ |
| ink-faint `#A39E98` on canvas | 2.44:1 | ❌ placeholder 以外（キャプション・メタ実情報）への使用は非適合 |
| **white on primary `#038B94`**（主要 CTA） | **4.10:1** | ❌ 通常文字未達（4.5 未満）。white on `#027078` は 5.85:1 で適合 |
| brandDark `#025F66` on brandLight `#E1F3F4` | 6.47:1 | ✅ |

### 2.3 是正案 → 実装済み（2026-09-30）

背景パステルは維持し、文字側を暗くする方針（業務ステータスの色相は変更しない = 学習済みの色意味を保全）。`C.textBadge*` メンバを新設し `BADGE.*` の text 参照のみ差し替え — 共有 `C.text*` メンバの値は変更していないため波及面はバッジに限定される。

| コンボ | 採用 text 値 | 実測比率 | 備考 |
|---|---|---:|---|
| gray | `#615D59`（`textBadgeGray`） | 5.52:1 | 死亡 sentinel の可読性を実質改善 |
| yellow | `#7A5C00`（`textBadgeYellow`、textCheckupDueSoon 同値） | 5.36:1 | |
| orange | `#793400`（`textBadgeOrange`、DESIGN.md accent-orange-deep） | 7.81:1 | |
| red | `#B03A2E`（`textBadgeRed`） | ~5.0:1 | 4.45 は境界値のため #B03A2E を採用 |
| green | `#0C6E5F`（`textBadgeGreen`） | 5.09:1 | |
| muted | `#615D59`（`textBadgeMuted`） | 5.77:1 | |
| primary CTA | 塗りを `#027078` に変更、hover を一層暗く | 5.85:1 | **依然として製品判断事項 — 未適用**。ブランド主要色の変更となるため決裁要 |

## 3. Loading 状態（新規標準 — FE-DS1）

Sparkle `Button` の `loading` prop に対応する本システムの標準:

- `<Button loading>` は `disabled` + `aria-busy="true"` + ラベル先頭に `Loader2 animate-spin`（`ICON.sm`）を表示。テキストは維持（スピナーで置き換えない = 幅確保と内容把握のため）。
- `loading` 中は `onClick` を発火させない（`disabled` 経由）。
- フォーム送信は React 19 `useActionState` の `isPending` を `loading` に接続するのが正パターン（`SubmitButton` 参照）。

## 4. 状態トークンの配置（将来タスク）

`PALETTE` 末尾に混入している状態クラス 4 件（`hoverBgInput` / `focusBorderLegacyAccent` / `focusRingActionPrimary` / `focusRingBrand`）は、将来 `export const STATE`（L2.5: interaction-state バケット）に移す正本化を予定する。当面は design-token-layers.md §5 のリーク記録どおり現状維持。

## 5. 検証手段

- `pnpm design-audit` — 値・構造ルール（C1–C19）
- `button.test.tsx` / `dialog.test.tsx` 等 — focus-visible ring の回帰テスト（EMR-207 由来）
- **Storybook + `@storybook/addon-a11y`** — 全 variant × 全状態の目視/axe 検査（導入後はカタログが状態仕様のリビングドキュメントになる）
- コントラスト再計測 — 本書 §2 の表を再計算して更新する（値変更時に同コミットで更新すること）
