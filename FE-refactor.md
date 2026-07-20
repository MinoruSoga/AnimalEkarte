# FE-refactor 第10期（FE10）— DESIGN.md 字義遵守リブランド＋全ページ視覚スイープ

> 起票: 2026-07-21（要件責任者: 曽我。決裁=「DESIGN.md 字義遵守（リブランド）」— AskUserQuestion で
> 「design-system.md 製品翻訳基準」との二択を提示し、字義遵守を明示選択）
> **本ファイルは対応後削除する使い捨てトラッカー**。恒久規約は `docs/spec/design-system.md`（本期 R5 で字義基準へ改訂）。
> 旧 FE9-3（製品翻訳基準・代表6画面クローズ）は本決裁により**撤回**。teal 期の ✅ は全て無効（P5 基準色が変わるため全画面再監査）。

## 決裁内容と唯一の例外

- **字義遵守 = DESIGN.md フロントマターのトークン値・コンポーネント定義をそのまま製品へ適用する**:
  - 構造色 primary: `#038B94`(teal) → **`#0075DE`**（primary-active `#027078` → **`#005BAB`**）
  - hairline: `rgba(0,0,0,0.09)` → **`#E6E6E6`** 固体値
  - text-input border: `rgba(0,0,0,0.16)` → **`#DDDDDD`**
  - caption 13px → **14px**、eyebrow/micro 11px → **12px**（DESIGN.md スケール最下段）
  - radius xxs 3px（製品拡張）→ **4px**（DESIGN.md 最小段 xs に整列）
  - `button-primary` = pill（実装は既に pill — §7.2 の「pill はマーケ専用」裁定を撤回）
  - データテーブルヘッダ = `ex-data-table-cell`（canvas-soft 帯 + eyebrow 型）— house 様式裁定を撤回
  - focus signal = primary（入力 focus の legacy 青 `#2383E2` 系も primary へ統一）
- **唯一の例外（削除しない）**: §2.4 臨床 semantic 色（danger `#C0392B`・warning・status green/blue）。
  DESIGN.md Semantic 節は「マーケ表面に semantic ランプが無い」という*観察*であり削除の指令ではない。
  臨床安全（SPECIFICATION 2.1・全原則に優先）により死亡表示・危険バッジ・RBAC 非活性の色は維持する。
  semantic 色を構造色として使わない従来ルールは継続。
- 対象 = 本体 84 ルート。liff/line-reserve は brand 色参照ゼロ（grep 実測）のためトークン反転の影響なし・ページ監査対象外（従来どおり）。
- チャートのデータ系列色（VitalsGraph 等）はデータ可視化パレットであり構造色規則の対象外（現状維持）。

## フェーズ

| # | 内容 | 状態 |
|---|------|------|
| R1 | 構造色反転: `design-tokens.ts`（brand/brandHover/brandLight/brandDark/focus/checked/glow）+ `globals.css`（--primary/--ring/--shadow-focus-brand/--shadow-brand-glow/--sidebar-*）+ STYLE.confirmPrimary brand 化 + コメント/テスト追随 | ✅済 |
| R2 | 字義トークン化: hairline `#E6E6E6`・input border `#DDDDDD`・--text-xs 14px・--text-2xs 12px・--radius-xxs 4px | ✅済 |
| R3 | テーブルヘッダ `ex-data-table-cell` 化: `STYLE.tableHeaderRow`（canvas-soft 帯）+ `tableHeaderCell`（eyebrow 型）一括反転（部分適用禁止 — 旧 §7.5） | ✅済 |
| R4 | 機械ガード反転: audit C1 = `C.accent`/`#2383E2`/**`#038B94`/`#027078`**（teal を legacy 化・`#0075DE` を解禁）+ audit テスト追随 | ✅済 |
| R5 | 文書同期: `docs/spec/design-system.md`（字義基準へ改訂・製品上書き撤回・臨床例外のみ残す）+ `ui-design-compliance.md` C1 行/注記 + `frontend/CLAUDE.md` | ✅済 |
| R6 | 全ページ視覚スイープ（下表・P1〜P7 + リブランド残渣 T 判定）| 進行中 |

## R6 方法（1画面ずつ）

ログイン済みブラウザ（port 9222）で screenshot → 判定 → 逸脱を最小差分修正 → batch 単位コミット → 本表更新。

```
P1 図地   : ページ canvas=#F6F5F4・カード/フィールド=白 surface。全面純白でない
P2 境界   : カード境界=hairline #E6E6E6。通常カードに重い影/二重枠がない
P3 階層   : 見出しロール適合。font-bold が本文/数値セルに漏れてない
P4 余白   : spacing スケール(4/8/12/16/24/28/32)で階段状。罫線でなく whitespace でグルーピング
P5 アクセント: primary #0075DE が CTA/リンク/active/focus のみ。teal 残渣ゼロ。sticker/semantic が構造に漏れてない
P6 テーブル: ex-data-table-cell 様式（canvas-soft ヘッダ帯 + eyebrow 型 + hairline 行区切り）
P7 状態   : hover/focus 知覚可能。disabled が RBAC 非活性表現を退行させてない
T  残渣   : teal/旧 accent 青の直値・画像・ハードコードが視覚に残っていない
```

判定: ✅ 準拠 / 🔧 修正した / ⚠️ 裁定要 / — 該当なし

## R6 進捗表（ルート正本 = docs/spec/ui-design-compliance.md §2・84 ルート）

| Batch | ルート | P1 | P2 | P3 | P4 | P5 | P6 | P7 | T | 所見 | 状態 |
|-------|--------|----|----|----|----|----|----|----|---|------|------|
| A | /login | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | forgot リンクを muted ink→brand 化（DESIGN.md「inline link=primary」） | ✅済 |
| A | /forgot-password | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | 戻るリンク×2 brand 化。**副次発見: 未認証で到達不能（BUG-419 起票・機能バグ）** | ✅済 |
| A | /reset-password | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | リンク×2 brand 化（無効リンク状態で実測） | ✅済 |
| A | /（受付カンバン） | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | CTA/カルテリンク blue pill 反映確認。カンバン列パステル=カテゴリ装飾（sanctioned） | ✅済 |
| A | /owners（一覧） | ✅ | ✅ | ✅ | ✅ | ✅ | 🔧 | ✅ | ✅ | ヘッダを eyebrow 字義化（sectionLabel 16px→12px/600 一括・computed 実測 12px/600・band #F6F5F4・hairline #E6E6E6） | ✅済 |
| A | /owners/new | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 会員区分 active chip=brand ✓・required\*=semantic ✓ | ✅済 |
| A | /owners/:id | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 未払い warning=semantic 帯（§2.4 維持対象） | ✅済 |
| A | /owners/:id/report | ✅ | ✅ | ✅ | ✅ | ⚠ | — | ✅ | ✅ | セクション見出しの brand 青 tick=decorative-primary 疑い。**owner-report feature に他セッション未コミット変更があり衝突回避で保留**（当該 workstream commit 後に裁定） | ⚠保留 |
| A | /aggregation | ✅ | ✅ | 🔧 | ✅ | 🔧 | ✅ | ✅ | ✅ | active タブ ink→brand 化（UnifiedTabs/dataActive\* トークン一括=全タブ画面に波及・computed 実測 #0075DE） | ✅済 |
| B1 | /reservations（週カレンダー） | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | CTA/表示トグル/今日ハイライト=brand ✓。カテゴリ凡例=DB 設定色 dots（sanctioned） | ✅済 |
| B1 | 予約詳細モーダル | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | LINE 未連携警告を danger 赤→warning 帯へ（§2.4 階調適正化・danger 希釈防止）。カテゴリ帯/状態紫=装飾 status（sanctioned） | ✅済 |
| B1 | 予約登録モーダル | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | 検索=ink 黒フィル→brand・初診選択状態=danger→brand（computed 実測）。**同型 ink フィル action を全域一括反転: paginationBtnActive・DatePicker 選択日・ConfirmDialog 非 danger 主ボタン・MasterSelectModal チェック円** | ✅済 |
| B2 | /hospitalization | | | | | | | | | | 未 |
| B2 | /hospitalization/select-pet | | | | | | | | | | 未 |
| B2 | /hospitalization/new | | | | | | — | | | | 未 |
| B2 | /hospitalization/:id（詳細+デイリー記録） | | | | | | | | | | 未 |
| B2 | /hospitalization/:id/edit | | | | | | — | | | | 未 |
| B3 | /medical-records（一覧） | | | | | | | | | | 未 |
| B3 | /medical-records/select-pet | | | | | | | | | | 未 |
| B3 | /medical-records/new | | | | | | — | | | | 未 |
| B3 | /medical-records/:id（問診/診察/治療プランタブ） | | | | | | | | | | 未 |
| B4 | /accounting | | | | | | | | | | 未 |
| B4 | /accounting/select-pet | | | | | | | | | | 未 |
| B4 | /accounting/new | | | | | | | | | | 未 |
| B4 | /accounting/:id | | | | | | | | | | 未 |
| B4 | /accounting/close | | | | | | | | | | 未 |
| B4 | /accounting/close/history | | | | | | | | | | 未 |
| B4 | /accounting/reports | | | | | | | | | | 未 |
| B4 | /estimates | | | | | | | | | | 未 |
| B4 | /estimates/new | | | | | | — | | | | 未 |
| B4 | /estimates/:id | | | | | | | | | | 未 |
| B4 | /estimates/:id/edit | | | | | | — | | | | 未 |
| B5 | /settings（トップ） | | | | | | — | | | | 未 |
| B5 | /settings/clinic | | | | | | | | | | 未 |
| B5 | /settings/staff | | | | | | | | | | 未 |
| B5 | /settings/treatment-items | | | | | | | | | | 未 |
| B5 | /settings/diagnosis | | | | | | | | | | 未 |
| B5 | /settings/animal-species | | | | | | | | | | 未 |
| B5 | /settings/trimming | | | | | | | | | | 未 |
| B5 | /settings/trimming-course-type | | | | | | | | | | 未 |
| B5 | /settings/medicine | | | | | | | | | | 未 |
| B5 | /settings/reservation-type | | | | | | | | | | 未 |
| B5 | /settings/hospitalization | | | | | | | | | | 未 |
| B5 | /settings/cage | | | | | | | | | | 未 |
| B5 | /settings/merchandise-items | | | | | | | | | | 未 |
| B5 | /settings/insurance | | | | | | | | | | 未 |
| B5 | /settings/occupations | | | | | | | | | | 未 |
| B5 | /settings/permission-groups | | | | | | | | | | 未 |
| B5 | /settings/inquiry-templates | | | | | | | | | | 未 |
| B5 | /settings/interview/chief-complaint | | | | | | | | | | 未 |
| B5 | /settings/interview/templates | | | | | | | | | | 未 |
| B5 | /settings/shift-templates | | | | | | | | | | 未 |
| B5 | /settings/closing-time | | | | | | | | | | 未 |
| B5 | /settings/payment-methods | | | | | | | | | | 未 |
| B5 | /settings/campaigns | | | | | | | | | | 未 |
| B5 | /settings/integrations/lstep | | | | | | | | | | 未 |
| B5 | /settings/lstep/tags | | | | | | | | | | 未 |
| B6 | /inventory | | | | | | | | | | 未 |
| B6 | /inventory/new | | | | | | — | | | | 未 |
| B6 | /inventory/:id | | | | | | — | | | | 未 |
| B6 | /trimming | | | | | | | | | | 未 |
| B6 | /trimming/select-pet | | | | | | | | | | 未 |
| B6 | /trimming/new | | | | | | — | | | | 未 |
| B6 | /trimming/:id | | | | | | — | | | | 未 |
| B6 | /vaccinations | | | | | | | | | | 未 |
| B6 | /vaccinations/select-pet | | | | | | | | | | 未 |
| B6 | /vaccinations/new | | | | | | — | | | | 未 |
| B6 | /vaccinations/:id | | | | | | — | | | | 未 |
| B6 | /checkups | | | | | | | | | | 未 |
| B6 | /checkups/select-pet | | | | | | | | | | 未 |
| B6 | /checkups/new | | | | | | — | | | | 未 |
| B6 | /examinations | | | | | | | | | | 未 |
| B6 | /examinations/select-pet | | | | | | | | | | 未 |
| B6 | /examinations/new | | | | | | — | | | | 未 |
| B6 | /examinations/:id | | | | | | — | | | | 未 |
| B7 | /shifts | | | | | | — | | | | 未 |
| B7 | /lstep/checkup-sync | | | | | | | | | | 未 |
| B7 | /lstep/delivery-monitor | | | | | | | | | | 未 |
| B7 | /lstep/analytics | | | | | | | | | | 未 |
| B7 | /line-reservation/settings | | | | | | | | | | 未 |
| B7 | /line-reservation/page-editor | | | | | | | | | | 未 |
| B7 | /line-reservation/slots | | | | | | | | | | 未 |
| B7 | /manual | | | | | | — | | | | 未 |
| B7 | /manual/:category/:slug | | | | | | — | | | | 未 |

（404 fallback は inline 簡易要素のため対象外 — ui-design-compliance §2 脚注どおり。リダイレクト専用 12 route も対象外）

## 検証規約

- R1〜R5 後: `node scripts/design-system-audit.mjs`（frontend/）緑 + 影響テストの `npx vitest run <path>`。
- R6 各 batch 修正後: 影響 feature の scoped vitest + design-audit 緑維持。1 batch = 1 コミット。Co-Authored-By 禁止。
- フル lint/type-check/build/test は USER 手動（Auto-Execution Prohibited 準拠）。リブランドは視覚全面変更のため、R6 完了後に USER のフルゲート + 目視承認を必須とする。
