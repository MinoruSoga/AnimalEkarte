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
| R6 | 全ページ視覚スイープ（下表・P1〜P7 + リブランド残渣 T 判定）| **完了: 84 ルート中 82 面 ✅**。残 = B2 入院詳細/編集 2 面（⏸ 入院データ 0 件・投入後に形式確認）+ owner-report の brand tick ⚠ 1 件（他セッション未コミット変更の commit 後に裁定） |

### 再開手順（次セッション向け）

1. backend が green（`curl http://localhost:8080/health` = 200）でログイン可能になったら、port 9222 の Chrome でデモアカウント（林 文明）ログイン → B7 残 7 面を probe（eyebrow th / brand pill / active tab / teal 残渣）。
2. **Vite 罠**: ホスト側でファイル編集後は `docker compose exec frontend touch <file>` しないと dev server が stale transform を配信し続ける（bind mount のイベント欠落・本スイープ中に実測）。
3. B2 入院詳細/編集はデータ投入後に実施。owner-report の brand tick ⚠ は当該 feature の他セッション未コミット変更が commit された後に裁定。
4. 全行 ✅ 後: USER フルゲート（lint/type-check/build/test 手動）+ 目視承認 → 本ファイル削除。

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
| B2 | /hospitalization | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | タブ active=brand 反映 ✓・空ケージ=dashed 空状態 idiom・リスト表示テーブル eyebrow 実測(12px/600) | ✅済 |
| B2 | /hospitalization/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 死亡行=グレーアウト+選択不可(臨床非活性維持 ✓)・検索=brand pill | ✅済 |
| B2 | /hospitalization/new | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 治療プラン表=eyebrow 様式 ✓・数値セル非 bold ✓ | ✅済 |
| B2 | /hospitalization/:id（詳細+デイリー記録） | | | | | | | | | 入院データ0件（八王子）のため未実施。デモデータ汚染回避で作成せず — データ投入後に実施 | ⏸データ待ち |
| B2 | /hospitalization/:id/edit | | | | | | — | | | 同上 | ⏸データ待ち |
| B3 | /medical-records（一覧） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ヘッダ eyebrow 実測(12px/600)・作成中/確定済/会計/担当医⚠は status・semantic（sanctioned） | ✅済 |
| B3 | /medical-records/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection（computed 全通過） | ✅済 |
| B3 | /medical-records/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | MedicalRecordForm 同一実体（:id で実測） | ✅済 |
| B3 | /medical-records/:id（問診/診察/治療プランタブ） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | タブ active=brand ✓・保存=brand pill・削除=danger ghost（semantic ✓）。治療タブは UI 切替不能（状態依存）だがヘッダは sectionLabel 経由で eyebrow 反映済み（静的確認） | ✅済 |
| B4 | /accounting（3タブ） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 保留=warning badge・カルテ=blue link（sanctioned） | ✅済 |
| B4 | /accounting/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection（computed） | ✅済 |
| B4 | /accounting/new | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | AccountingDetail 同一実体（:id で実測） | ✅済 |
| B4 | /accounting/:id（会計精算） | ✅ | ✅ | ✅ | ✅ | ✅ | 🔧 | ✅ | ✅ | 明細ヘッダ eyebrow 化（**shadcn TableHead 基底に eyebrow を設定=全 shadcn テーブル一括**）・物販ピッカー表も統一・sticky ヘッダ bg=canvas-soft 化。金額大字 bold=pricing idiom（許容） | ✅済 |
| B4 | /accounting/close | ✅ | ✅ | ✅ | ✅ | 🔧 | — | ✅ | ✅ | 区分選択 active=brand ✓。**brand フィル button=pill 字義を shadcn buttonVariants(default/primary)へ一括適用**・旧 medical accent blue(#2EAADC)を brand へ値統合（第二構造アクセント禁止） | ✅済 |
| B4 | /accounting/close/history | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | 空状態 ✓ | ✅済 |
| B4 | /accounting/reports | ✅ | ✅ | ✅ | ✅ | ✅ | 🔧 | ✅ | ✅ | 日次明細ヘッダを eyebrow 字義化（DailyBreakdownTable） | ✅済 |
| B4 | /estimates | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | eyebrow/brand pill computed 実測 | ✅済 |
| B4 | /estimates/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | 作成=brand pill・disabled 表現 ✓ | ✅済 |
| B4 | /estimates/:id | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 承認済み=green semantic badge ✓・明細ヘッダ eyebrow ✓ | ✅済 |
| B4 | /estimates/:id/edit | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | EstimateForm 同一実体 | ✅済 |
| B5 | /settings（トップ） | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | グループラベル=sectionLabel eyebrow 化反映（実測） | ✅済 |
| B5 | /settings/clinic | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過（eyebrow/brand pill/canvas） | ✅済 |
| B5 | /settings/staff | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **MasterCRUDPage シェル代表・スクショ精査**。権限バッジ=DB 色 dots・有効=status badge（sanctioned） | ✅済 |
| B5 | /settings/treatment-items | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **MasterTabPage シェル代表**（eyebrow/brand tab/brand pill probe 全通過） | ✅済 |
| B5 | /settings/diagnosis | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterTabPage シェル同一実体（代表実測） | ✅済 |
| B5 | /settings/animal-species | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/trimming | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterTabPage シェル同一実体 | ✅済 |
| B5 | /settings/trimming-course-type | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/medicine | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過。SidePanel ラベル=sectionLabel 統一済み（B4 静的修正） | ✅済 |
| B5 | /settings/reservation-type | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe: 実ヘッダ eyebrow ✓（空ハンドル列 th のブラウザ既定 16px/700 は空セルで視覚影響なし） | ✅済 |
| B5 | /settings/hospitalization | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/cage | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/merchandise-items | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/insurance | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/occupations | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/permission-groups | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | スクショ精査。RBAC 表示退行なし（C6a） | ✅済 |
| B5 | /settings/inquiry-templates | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/interview/chief-complaint | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/interview/templates | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | inquiry-templates と同一 Component | ✅済 |
| B5 | /settings/shift-templates | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe: eyebrow ✓・新規登録=brand テキストリンク（inline link=primary ✓） | ✅済 |
| B5 | /settings/closing-time | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | スクショ精査: 保存=brand pill（buttonVariants pill 化反映）・checked=brand | ✅済 |
| B5 | /settings/payment-methods | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/campaigns | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MasterCRUDPage シェル同一実体 | ✅済 |
| B5 | /settings/integrations/lstep | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | スクショ精査: form card・未設定 badge・4px 入力 ✓ | ✅済 |
| B5 | /settings/lstep/tags | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過 | ✅済 |
| B6 | /inventory | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過（eyebrow/brand pill） | ✅済 |
| B6 | /inventory/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | probe: 登録=brand pill・入力 4px・canvas ✓ | ✅済 |
| B6 | /inventory/:id | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | InventoryForm 同一実体 | ✅済 |
| B6 | /trimming | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過 | ✅済 |
| B6 | /trimming/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection（複数面で実測済み） | ✅済 |
| B6 | /trimming/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | フォーム共通 idiom（shell） | ✅済 |
| B6 | /trimming/:id | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | TrimmingForm 同一実体 | ✅済 |
| B6 | /vaccinations | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過 | ✅済 |
| B6 | /vaccinations/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection | ✅済 |
| B6 | /vaccinations/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | フォーム共通 idiom | ✅済 |
| B6 | /vaccinations/:id | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | VaccinationForm 同一実体 | ✅済 |
| B6 | /checkups | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過 | ✅済 |
| B6 | /checkups/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection | ✅済 |
| B6 | /checkups/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | フォーム共通 idiom | ✅済 |
| B6 | /examinations | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過 | ✅済 |
| B6 | /examinations/select-pet | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 共有 PetSelection | ✅済 |
| B6 | /examinations/new | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | フォーム共通 idiom | ✅済 |
| B6 | /examinations/:id | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | ExaminationForm 同一実体 | ✅済 |
| B7 | /shifts | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | スクショ精査: シフト種別チップ=装飾 status tint・週末=カレンダー semantic（sanctioned） | ✅済 |
| B7 | /lstep/checkup-sync | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | probe: 対象者を検索=brand。テーブルは STYLE.tableHeader* 経由（静的確認済み） | ✅済 |
| B7 | /lstep/delivery-monitor | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | probe 全通過（eyebrow 12px/600・teal 残渣なし） | ✅済 |
| B7 | /lstep/analytics | ✅ | ✅ | ✅ | ✅ | ✅ | 🔧 | ✅ | ✅ | 来院転換/CSV 取込/配信統計の手書き 15px ヘッダを eyebrow 字義化（reload 後 12px/600 実測） | ✅済 |
| B7 | /line-reservation/settings | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | probe: 設定を保存=brand pill・入力 4px・teal 残渣なし・canvas ✓ | ✅済 |
| B7 | /line-reservation/page-editor | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | probe: 変更を保存=brand pill・teal 残渣なし | ✅済 |
| B7 | /line-reservation/slots | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | probe 全通過（brand フィルなし=閲覧設定面・teal 残渣なし） | ✅済 |
| B7 | /manual | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | スクショ精査: 画面別トグル=brand・active 項目=brand tint・doc shell canvas ✓。markdown 文書表（bg-black/5 ヘッダ）は**文書レンダリング**であり app データテーブル規範の対象外と裁定（印刷帳票と同型） | ✅済 |
| B7 | /manual/:category/:slug | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | 同一 ManualPage（記事ビューをスクショ精査） | ✅済 |

（404 fallback は inline 簡易要素のため対象外 — ui-design-compliance §2 脚注どおり。リダイレクト専用 12 route も対象外）

## 検証規約

- R1〜R5 後: `node scripts/design-system-audit.mjs`（frontend/）緑 + 影響テストの `npx vitest run <path>`。
- R6 各 batch 修正後: 影響 feature の scoped vitest + design-audit 緑維持。1 batch = 1 コミット。Co-Authored-By 禁止。
- フル lint/type-check/build/test は USER 手動（Auto-Execution Prohibited 準拠）。リブランドは視覚全面変更のため、R6 完了後に USER のフルゲート + 目視承認を必須とする。
