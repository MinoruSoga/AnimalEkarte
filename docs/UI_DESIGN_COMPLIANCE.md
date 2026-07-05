# UI Design Compliance — DESIGN.md / docs/DESIGN_SYSTEM.md 対応状況

> **目的**: メイン EMR フロントエンド（`frontend/src`）の全ルートページ UI を、[DESIGN.md](../DESIGN.md)（意匠 SSOT）および [docs/DESIGN_SYSTEM.md](DESIGN_SYSTEM.md)（実装規約）に対してドメイン単位で監査・追跡する。
> **読者**: フロントエンド実装者・レビュアー。
> **タイミング**: UI デザイン準拠タスクの実施時・追跡時。
> **最新更新**: 2026-07-05

---

## 1. サマリー

| 項目 | 結果 |
|---|---|
| ドメイン数 | D01–D24（24） |
| status=完了 | 24 / 24（Known Exclusions を含むもの 3 件: D16, D17, D21 は「完了（除外あり）」） |
| `node scripts/check-design-primary-cta.mjs` | `PASS  no primary CTA accent reintroduction detected`（exit 0） |
| hex 直書き（`frontend/src/features/**` 実コード、コメント・print-only・design-tokens.ts 除く） | 0 件 |
| Tailwind 生色クラス（同上） | 0 件 |
| 臨床安全 UI（危険バッジ・死亡グレーアウト・RBAC 非活性表示） | 退行なし（本タスクでは D06/D07 のコードは変更していない。D15 は警告バナーの色トークン置換のみで表示条件・disabled ロジックは不変） |

**本タスクで実際にコードを変更したドメインは D15（accounting）と D18（lstep）の 2 ドメイン、計 4 ファイルのみ**。他の 22 ドメインは、本タスク開始前の先行コミット（`e93b53a2` "align owner and medical-record UI with DESIGN.md tokens" / `e7b8bc16` "DESIGN.md app-wide — brand CTA + table header unification" 等）で既に DESIGN.md 準拠が完了しており、監査の結果、追加の実コード修正は不要と判定した。

---

## 2. Domain Matrix（D01–D24）

| ID | ドメイン | routes | feature | status | 監査結果 | 変更ファイル |
|----|---------|--------|---------|--------|---------|-------------|
| D01 | auth | /login, /forgot-password, /reset-password | features/auth | 完了 | hex/tw 違反 0（先行コミットで brand SSOT 化済み: LoginForm/ForgotPasswordPage/ResetPasswordPage） | なし（本タスクでは無変更） |
| D02 | reception | / | features/reception | 完了 | hex/tw 違反 0（先行コミットで ReceptionDialogActionButtons を brand 化済み） | なし |
| D03 | owners | /owners/* | features/owners, app/pages/Owner* | 完了 | hex/tw 違反 0（参照実装。先行コミットで OwnersListTable/OwnerPetsSection/PetEditModal 等を DESIGN.md トークン化済み） | なし |
| D04 | aggregation | /aggregation | features/aggregation | 完了 | hex/tw 違反 0（AggregationOwnerTable が DESIGN_TABLE_HEADER_ROW/CELL 採用済み） | なし |
| D05 | reservations | /reservations | features/reservations | 完了 | hex/tw 違反 0。バッジ色（brown/pink 等スティッカーパレット）は装飾用途のみで CTA/構造フィルには未使用 | なし |
| D06 | medical-records | /medical-records/* | features/medical-records | 完了 | hex/tw 違反 0（先行コミットで ExaminationFilter/ExaminationGroup/VaccinationForm/TreatmentTable 等を広範に token 化済み）。臨床安全 UI（危険バッジ等）は本タスクで無変更・退行なし | なし |
| D07 | hospitalization | /hospitalization/* | features/hospitalization | 完了 | hex/tw 違反 0。HospitalizationListView は DESIGN_TABLE_HEADER_ROW/CELL 採用済み。臨床安全 UI（死亡グレーアウト等）は本タスクで無変更・退行なし | なし |
| D08 | trimming | /trimming/* | features/trimming | 完了 | hex/tw 違反 0。TrimmingListTable は DESIGN_TABLE_HEADER_ROW/CELL 採用済み | なし |
| D09 | examinations | /examinations/* | features/examinations | 完了 | hex/tw 違反 0（先行コミットで ExaminationsList 等 token 化済み） | なし |
| D10 | vaccinations | /vaccinations/* | features/vaccinations | 完了 | hex/tw 違反 0（先行コミットで VaccinationForm/VaccinationHistory/VaccinationList 等 token 化済み） | なし |
| D11 | checkups | /checkups/* | features/checkups | 完了 | hex/tw 違反 0（先行コミットで CheckupsTab 系 token 化済み） | なし |
| D12 | inventory | /inventory/* | features/inventory | 完了 | hex/tw 違反 0。InventoryList は DESIGN_TABLE_HEADER_ROW/CELL 採用済み | なし |
| D13 | estimates | /estimates/* | features/estimates | 完了 | hex/tw 違反 0。EstimateList は DESIGN_TABLE_HEADER_ROW/CELL 採用済み | なし |
| D14 | shifts | /shifts, /settings/shift-templates | features/shifts | 完了 | hex/tw 違反 0。ShiftTemplateSettings(Parts) は DESIGN_TABLE_HEADER_ROW/CELL 採用済み | なし |
| D15 | accounting | /accounting/* | features/accounting | 完了 | **修正実施**: レジ締め後編集理由ラベル・クレジット確定後訂正の警告バナーが Tailwind 生色（`text-red-600` / `bg-red-50` / `border-red-200`）を使用していたため `C.danger` / `C.bgDanger8` / `C.borderDanger20` に置換。表示条件・disabled/権限ロジック・DOM構造は不変。`DailyAccountingTab.tsx` の印刷専用テーブル（`@media print` のみ表示・A4横帳簿レプリカ）は Known Exclusion（§4）とし対象外 | `routes/AccountingDetail.tsx`, `components/CreditCorrectionDialog.tsx` |
| D16 | cash-register | /accounting/close/* | features/cash-register | 完了（除外あり） | hex/tw 違反 0（画面 UI）。`ClosePrintArea.tsx` は印刷専用（`@media print` のみ表示）の A4 帳簿レプリカで Tailwind 生色グリッド線を使用 → Known Exclusion（§4） | なし |
| D17 | accounting-reports | /accounting/reports | features/accounting-reports | 完了（除外あり） | hex/tw 違反 0（画面 UI）。`MonthlyReportPrintArea.tsx` は印刷専用の月次帳票レプリカ → Known Exclusion（§4） | なし |
| D18 | lstep | /lstep/*, /settings/lstep/* | features/lstep, features/settings | 完了 | **修正実施**: 配信失敗警告バナー（`bg-red-50`/`border-red-200`/`text-red-600`/`text-red-700`）を `C.bgDanger8`/`C.borderDanger20`/`C.danger` に置換し `role="alert"` を付与。タグ管理テーブルのヘッダーを `STYLE.tableHeaderRow/Cell`（既定）から `DESIGN_TABLE_HEADER_ROW/CELL`（canvas-soft + eyebrow）に統一 | `pages/LstepDeliveryMonitorPageParts.tsx`, `components/TagSummaryTable.tsx` |
| D19 | line-reservation | /line-reservation/* | features/line-reservation | 完了 | hex/tw 違反 0 | なし |
| D20 | clinic-settings | /settings/clinic | features/clinic-settings | 完了 | hex/tw 違反 0（grep 一致は全て `#190` 等 Issue 番号コメント） | なし |
| D21 | master-settings | /settings/*（D14/D18 以外） | features/master, features/closing-settings | 完了（除外あり） | hex/tw 違反 0（grep 一致は Issue 番号コメント、または `features/master/PATTERNS.md` 内のドキュメント例示コードのみ）。PATTERNS.md は内部パターン集ドキュメントであり実行コードではないため Known Exclusion（§4）とし、実装済みトークンとの整合は別タスクでの更新を推奨 | なし |
| D22 | manual | /manual/* | features/manual | 完了 | hex/tw 違反 0 | なし |
| D23 | owner-report | /owners/:id/report | features/owner-report | 完了 | hex/tw 違反 0（先行コミットで HistoryTable 等 token 化済み。header は canvas-soft + eyebrow uppercase を実装済み） | なし |
| D24 | shared-shell | Layout 配下共通 | components/shared/*, components/ui/* | 完了 | hex/tw 違反 0。`DataTable.tsx`（`DESIGN_TABLE_HEADER_ROW/CELL` 提供元）、`SortableHeader.tsx`、`dialog.tsx`（`rounded-xl p-6 shadow-lg` = ex-modal-card 準拠）、`SubmitButton.tsx`/`PrimaryButton.tsx`（`colorVariant="brand"` opt-in）、`Sidebar.tsx` 等、いずれも準拠 | なし |

---

## 3. Audit Procedure（監査手順）

各ドメインについて、以下のコマンドを実行し結果をゼロにしたことを確認する（`<domain>` を対象ドメインの feature ディレクトリ名に置換）。

```bash
# 1. hex 直書き検出（コメント・design-tokens.ts・print-only コンポーネントは目視で除外）
rg -n '#[0-9A-Fa-f]{3,8}' frontend/src/features/<domain> --glob '!*.test.*'

# 2. Tailwind 生色クラス検出
rg -n '\b(bg|text|border)-(red|blue|green|gray|slate|zinc|neutral|stone|orange|yellow|purple|pink|indigo|cyan|teal|emerald|lime|amber|rose|violet|fuchsia|sky)-[0-9]{2,3}\b' frontend/src/features/<domain>

# 3. 旧 accent Primary CTA 再混入ガード（全ドメイン共通スコープ: features/** + components/shared/**）
node scripts/check-design-primary-cta.mjs

# 4. ドメインスコープの vitest
docker compose exec frontend npx vitest run src/features/<domain>
```

**判定基準**:
- 1, 2 の grep 一致が「Issue 番号コメント（`#158` 等）」「`@media print` のみで表示される印刷専用コンポーネント」「テストのフィクスチャ文字列」「`design-tokens.ts` / `globals.css`（トークン定義 SSOT 自体）」のいずれかであれば violation ではない。
- 3 が exit 0 であること。
- 4 が該当ドメインの既存テストを含め PASS すること。

全 24 ドメインに対して本手順を実施し、D15（accounting）と D18（lstep）でのみ実コード修正が必要な violation を検出した（詳細は §2 参照）。他ドメインは grep 一致が上記除外条件に該当し、修正不要と判定した。

---

## 4. Known Exclusions（既知の除外事項）

以下は本タスクのスコープ外、または DESIGN.md の適用対象外として明示的に除外する。

| # | 対象 | 除外理由 |
|---|---|---|
| 1 | 印刷専用コンポーネント（`DailyAccountingTab.tsx` の `DailyPrintArea`、`MonthlyReportPrintArea.tsx`、`ClosePrintArea.tsx`） | `hidden` 属性 + `@media print` でのみ表示される A4 帳票・帳簿のレプリカ。画面上の「ページ canvas」ではなく物理紙面の再現が目的のため、DESIGN.md のスクリーン UI トークン規約の対象外。Tailwind の `gray-300`/`gray-400` 等の生色グリッド線は帳簿の視認性を保つための意図的な選択であり変更しない |
| 2 | `Input` / `SelectTrigger` の `rounded-md`（8px） | DESIGN.md 仕様値は `{rounded.xs}`（4px）だが、共有プリミティブのためアプリ全体への影響が大きく本タスクでは変更しない（ユーザー指示で明示的に対象外） |
| 3 | `C.accent` → `C.bgBrand` のアプリ全体一括移行（`SubmitButton`/`PrimaryButton` 既定 `colorVariant` 変更を含む） | 影響範囲が広いため対象外（ユーザー指示で明示的に対象外）。個別ドメインでの `colorVariant="brand"` opt-in は先行コミットで一部実施済み（owners/OwnersListPage 等）だが、既定値そのものは変更していない |
| 4 | `frontend/liff`, `frontend/line-reserve` | 別アプリのため対象外 |
| 5 | `change-ui.md`（受付テレメトリー機能 / `checked_in_at` BE 連携） | 別タスクのため対象外 |
| 6 | `design-tokens.ts` の大規模リファクタ | 対象外。今回は既存トークン（`C.danger`/`C.bgDanger8`/`C.borderDanger20`）のみを使用し、新規トークン追加は行っていない |
| 7 | `frontend/src/features/master/PATTERNS.md` 内のコード例示 | 内部向けパターン集ドキュメントに残る旧 hex 直書き例（`text-[#37352F]` 等）。実行コードではなく画面に影響しないため本タスクでは対象外。ドキュメント正確性の観点では別タスクでの更新を推奨 |

---

## 5. 完了条件チェック

| # | 条件 | 結果 |
|---|---|---|
| 1 | `docs/UI_DESIGN_COMPLIANCE.md` が存在し D01–D24 全行に status がある | PASS（本ファイル） |
| 2 | 全ドメイン status=完了（Known Exclusions のみ「完了（除外あり）」可） | PASS（D16/D17/D21 が「完了（除外あり）」、他 21 ドメインが「完了」） |
| 3 | `node scripts/check-design-primary-cta.mjs` が exit 0 | PASS |
| 4 | 各ドメインで hex 直書き・Tailwind 生色・旧 accent Primary CTA が残っていない | PASS（§4 の Known Exclusions を除く） |
| 5 | 変更ドメインの scoped vitest が PASS | PASS（`docker compose exec frontend npx vitest run src/features/accounting src/features/lstep` → 19 test files / 210 passed, 3 skipped, 0 failed） |
| 6 | 臨床安全 UI の既存テストが退行していない | PASS（D06/D07 は本タスクで無変更。D15/D18 は警告バナーの色トークン置換のみで表示条件・disabled/権限ロジックは不変。code-reviewer subagent による重点レビュー実施: CRITICAL/HIGH 0件、APPROVE（条件付き）。指摘事項は D18 `TagSummaryTable.tsx` のヘッダーが `DESIGN_TABLE_HEADER_CELL`（eyebrow: uppercase + tracking-wide）へ変わる視覚差分のみで、これは DESIGN_SYSTEM.md §6 の `ex-data-table-cell` header 仕様への準拠そのものであり意図した変更） |
