# UI Design Compliance — DESIGN.md / docs/DESIGN_SYSTEM.md 対応状況

> **目的**: メイン EMR フロントエンド（`frontend/src`）の全ルートページ UI を、[DESIGN.md](../DESIGN.md)（意匠 SSOT）および [docs/DESIGN_SYSTEM.md](DESIGN_SYSTEM.md)（実装規約）に対してドメイン単位・ページ単位・共通コンポーネント単位で監査・追跡する。
> **読者**: フロントエンド実装者・レビュアー。
> **タイミング**: UI デザイン準拠タスクの実施時・追跡時。
> **最新更新**: 2026-07-05

---

## 1. サマリー

| 項目 | 結果 |
|---|---|
| ドメイン数 | D01–D24（24） |
| status=完了 | 24 / 24（Known Exclusions を含むもの 2 件: D16, D17 は「完了（除外あり）」。D21 は 2026-07-05 に PATTERNS.md 対応完了し「完了」に更新） |
| Page Registry 行数（§3） | 98 行（データ行、ヘッダ除く）。lazy `Component:` ルート 82 + `/login`（`React.lazy`+`Suspense` 直書きで router `lazy` 未使用のため grep 対象外）1 + `Navigate` redirect 12 + 404 フォールバック 1 + 対象外（`frontend/liff`/`frontend/line-reserve`）2 |
| Shared Component Registry 行数（§4） | 35 行（§4.1〜§4.6、DESIGN.md 関連の主要 shared/ui コンポーネントのみ。97 shared ファイル全件は列挙しない）。34 行が `完了`（2026-07-05 追記2 で C011–C015 の残 5 行を解消）+ C025（`PrintPortal`）のみ印刷専用（Known Exclusion #1）で `完了（除外あり）`。`部分準拠` は 0 件（2026-07-05 再監査で C035 `SearchableSelect` を追加、後述） |
| 全ページ完了率 | 100%（84/84。分母 = 98 行 − redirect 12 − 対象外 2。うち 81 行が「完了」、3 行（D16×2 + D17×1）が「完了（除外あり）」） |
| `node scripts/check-design-primary-cta.mjs` | `PASS  no primary CTA accent reintroduction detected`（exit 0） |
| hex 直書き（`frontend/src/features/**` 実コード、コメント・print-only・design-tokens.ts 除く） | 0 件 |
| Tailwind 生色クラス（同上） | 0 件 |
| 臨床安全 UI（危険バッジ・死亡グレーアウト・RBAC 非活性表示） | 退行なし（本タスクでは D06/D07 のコードは変更していない。D15 は警告バナーの色トークン置換のみで表示条件・disabled ロジックは不変） |

**本タスクで実際にコードを変更したドメインは D15（accounting）と D18（lstep）の 2 ドメイン、計 4 ファイルのみ**。他の 22 ドメインは、本タスク開始前の先行コミット（`e93b53a2` "align owner and medical-record UI with DESIGN.md tokens" / `e7b8bc16` "DESIGN.md app-wide — brand CTA + table header unification" 等）で既に DESIGN.md 準拠が完了しており、監査の結果、追加の実コード修正は不要と判定した。

**2026-07-05 追記（Page Registry / Shared Component Registry 新設）**: 本追記では実コードの変更は行わず、`frontend/src/app/routes/*.tsx`（ルート SSOT）を全走査し §3 Page Registry を、`components/shared/*` / `components/ui/*` の DESIGN 関連コンポーネントを §4 Shared Component Registry として新規追加した。status は §2 の 2026-07-05 監査結果をページ単位・コンポーネント単位に展開したものであり、新規の visual QA は実施していない。

**2026-07-05 追記2（Shared Component Registry 100% 完了化）**: §4 で「部分準拠」「完了（除外あり）」だった C011–C015 を解消する実コード変更を実施した。
- **Phase A（Known Exclusion #3 解消）**: `SubmitButton` / `PrimaryButton` の既定 `colorVariant` を `"brand"`（DESIGN.md `button-primary` = brand blue `#0075DE` + pill）に変更。48 ファイルの call site から冗長な `colorVariant="brand"` を削除し、`PetEditModal.tsx` の `STYLE.confirmPrimary` + brand 上書きハックを `PrimaryButton` 使用に正規化。`scripts/check-design-primary-cta.mjs` に `colorVariant="default"`（旧 accent への明示的 opt-out）検出ルールを追加。
- **Phase B（Known Exclusion #2 解消）**: `globals.css` に `--radius-xs: 4px`（DESIGN.md `{rounded.xs}`。既存コメントで xs/sm/md/lg/xl=4/5/8/12/16px と明記されていたが xs のみ未定義だったため追加）を定義し、`Input` / `Textarea` / `SelectTrigger` の角丸を `rounded-md`（8px）→ `rounded-xs`（4px）に変更。`SelectContent` 等の非 text-input コンポーネントは対象外のまま維持。
- 結果、C011–C015 は全て `完了` に更新（§4.3）。§6 Known Exclusions #2/#3 は対応済みとして記録（取り消し線）。視覚QAは自動化不可のため、角丸変更は `rg` による構造確認（AC-4）を PASS 根拠とする。

**2026-07-05 追記3（検証ファースト再監査 — Phase 0/1）**: 追記2 までの内容が実際にコミットされているか、および全 84 ページで §5 Audit Procedure が再現するかを再検証した。
- **Phase 0（全24ドメイン再監査）**: D01–D24 全ドメインで `rg` 2種（hex 直書き・Tailwind 生色）を再実行。一致した全件を目視確認し、Issue 番号コメント（`#150`〜`#215` 等）・DESIGN.md 準拠を示すコメント・D15/D16/D17 の印刷専用コンポーネント（Known Exclusion #1）のいずれかであり、新規 violation は 0 件だった。`node scripts/check-design-primary-cta.mjs` も PASS（exit 0）。
- **新規 FAIL 1件（AC-2 候補ギャップの確定）**: `frontend/src/components/ui/searchable-select.tsx` の `PopoverTrigger`（Select 相当の text-input trigger）が `rounded-md`（8px）のままで、同一役割の `SelectTrigger`（`rounded-xs`、追記2 で対応済み）と角丸が不一致だった。DESIGN.md `text-input` = `{rounded.xs}` に統一するため `rounded-md` → `rounded-xs` に修正（1 ファイル・1 行）。Shared Component Registry に C035（`SearchableSelect`）として追加。18 ファイルから参照される shared/ui コンポーネントのため、`ReservationFormModal.test.tsx`（`SearchableSelect` を直接使用）+ 代表ドメイン `src/features/owners` の scoped vitest（10 test files / 75 tests）で PASS を確認。
- **ドキュメント矛盾 2件の解消（AC-6）**: (1) §1 のサマリーが「Shared Component Registry 全34行が完了、部分準拠/完了（除外あり）は0件」としていたが、C025（`PrintPortal`）は印刷専用のため `完了（除外あり）` が正しい状態であり矛盾していた → 「34行が完了 + C025のみ完了（除外あり）」に修正。(2) §7 完了条件チェック #12 が `rg 'STYLE\.confirmPrimary' frontend/src/features frontend/src/components/shared` の結果を「0件」としていたが、実際は `SubmitButton.tsx` 自身（`colorVariant="default"` の実装本体）に 2 件ヒットする（`check-design-primary-cta.mjs` の `EXCLUDE_FILE` が明示的に除外している定義ファイルであり violation ではない）→ 実測値を反映し除外理由を明記。
- 上記以外の 22 ドメイン・33 Shared Component は Phase 0 で PASS のため無変更（AC-4: 差分修正の最小性）。

---

## 2. Domain Matrix（D01–D24）

各ドメインのページ単位の内訳は §3 Page Registry を参照。

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
| D15 | accounting | /accounting/* | features/accounting | 完了 | **修正実施**: レジ締め後編集理由ラベル・クレジット確定後訂正の警告バナーが Tailwind 生色（`text-red-600` / `bg-red-50` / `border-red-200`）を使用していたため `C.danger` / `C.bgDanger8` / `C.borderDanger20` に置換。表示条件・disabled/権限ロジック・DOM構造は不変。`DailyAccountingTab.tsx` の印刷専用テーブル（`@media print` のみ表示・A4横帳簿レプリカ）は Known Exclusion（§6）とし対象外 | `routes/AccountingDetail.tsx`, `components/CreditCorrectionDialog.tsx` |
| D16 | cash-register | /accounting/close/* | features/cash-register | 完了（除外あり） | hex/tw 違反 0（画面 UI）。`ClosePrintArea.tsx` は印刷専用（`@media print` のみ表示）の A4 帳簿レプリカで Tailwind 生色グリッド線を使用 → Known Exclusion（§6） | なし |
| D17 | accounting-reports | /accounting/reports | features/accounting-reports | 完了（除外あり） | hex/tw 違反 0（画面 UI）。`MonthlyReportPrintArea.tsx` は印刷専用の月次帳票レプリカ → Known Exclusion（§6） | なし |
| D18 | lstep | /lstep/*, /settings/lstep/* | features/lstep, features/settings | 完了 | **修正実施**: 配信失敗警告バナー（`bg-red-50`/`border-red-200`/`text-red-600`/`text-red-700`）を `C.bgDanger8`/`C.borderDanger20`/`C.danger` に置換し `role="alert"` を付与。タグ管理テーブルのヘッダーを `STYLE.tableHeaderRow/Cell`（既定）から `DESIGN_TABLE_HEADER_ROW/CELL`（canvas-soft + eyebrow）に統一 | `pages/LstepDeliveryMonitorPageParts.tsx`, `components/TagSummaryTable.tsx` |
| D19 | line-reservation | /line-reservation/* | features/line-reservation | 完了 | hex/tw 違反 0 | なし |
| D20 | clinic-settings | /settings/clinic | features/clinic-settings | 完了 | hex/tw 違反 0（grep 一致は全て `#190` 等 Issue 番号コメント） | なし |
| D21 | master-settings | /settings/*（D14/D18 以外） | features/master, features/closing-settings | 完了 | hex/tw 違反 0（実行コード）。`features/master/PATTERNS.md` 内のドキュメント例示コードに残っていた旧 hex 直書き（`text-[#37352F]`/`bg-[#2383E2]` 等）は `C.*` トークン参照に置換済み（2026-07-05、Known Exclusion #7 対応済み） | `frontend/src/features/master/PATTERNS.md`（ドキュメントのみ、実行コード変更なし） |
| D22 | manual | /manual/* | features/manual | 完了 | hex/tw 違反 0 | なし |
| D23 | owner-report | /owners/:id/report | features/owner-report | 完了 | hex/tw 違反 0（先行コミットで HistoryTable 等 token 化済み。header は canvas-soft + eyebrow uppercase を実装済み） | なし |
| D24 | shared-shell | Layout 配下共通 | components/shared/*, components/ui/* | 完了 | hex/tw 違反 0。`DataTable.tsx`（`DESIGN_TABLE_HEADER_ROW/CELL` 提供元）、`SortableHeader.tsx`、`dialog.tsx`（`rounded-xl p-6 shadow-lg` = ex-modal-card 準拠）、`SubmitButton.tsx`/`PrimaryButton.tsx`（既定 `colorVariant="brand"`、2026-07-05 追記2 で変更）、`Input`/`Textarea`/`SelectTrigger`/`SearchableSelect`（`rounded-xs`、追記2 + 追記3）、`Sidebar.tsx` 等、いずれも準拠 | `SubmitButton.tsx`, `PrimaryButton.tsx`, `input.tsx`, `textarea.tsx`, `select.tsx`, `globals.css`, `PetEditModal.tsx`, `scripts/check-design-primary-cta.mjs` ほか call site 48件（2026-07-05 追記2）; `searchable-select.tsx`（2026-07-05 追記3、Phase 0 再監査で検出） |

---

## 3. Page Registry（全ルートページ表）

**SSOT**: `frontend/src/app/routes/*.tsx`（`appRoutes` ツリー全体）。静的解析のみで作成し、実コードの変更は行っていない。status・最終監査日は §2 Domain Matrix の 2026-07-05 監査結果をページ単位に展開したもので、新規の visual QA は行っていない。

### 3.0 列定義・凡例

| 列 | 内容 |
|---|---|
| P-ID | `P001` 連番（ソート用。ドメイン順 → §2 と同じ D01→D24 の順） |
| path | ルートパス |
| ページ名 | 日本語表示名（sidebar-menu.tsx / FormHeader タイトル / 機能名） |
| コンポーネント | React コンポーネント名（router `lazy` の `Component:` export、または直接 import） |
| ソース | 主要ファイルパス |
| ドメイン | D01–D24（§2 と一致） |
| Layout | `Layout内` / `スタンドアロン` |
| status | `完了` / `完了（除外あり）` / `リダイレクト` / `対象外` |
| 最終監査 | `2026-07-05`（§2 と同じ監査日。リダイレクト・対象外は `-`） |
| 備考 | Known Exclusion 参照・権限ガード・print-only 等 |

**件数根拠（AC-1）**:

```bash
rg -c "return \{ Component:" frontend/src/app/routes/   # 合計 82
rg -c "Navigate to" frontend/src/app/routes/             # 合計 12（settings-routes.tsx のみ）
```

Page Registry データ行数 98 = Component 数 82 + `/login`（1、router `lazy` を使わず `React.lazy`+`Suspense` を直書きしているため上記 grep には現れないが実在するページ） + Navigate 数 12 + 404 フォールバック（1、`path: "*"` の inline element） + 対象外（liff/line-reserve、2）。

> **既知の差分（正直な報告）**: タスク説明の見積りは「lazy Component 行 ≈72」「settings の旧 URL redirect 10 件」だったが、実測では Component 行 82・Navigate 行 12（`job-title`〜`inquiry-template` 11 件 + `shift-template` 1 件、BUG-383）だった。Truth Source Priority 1（ルート SSOT）に従い、本 Registry は実コードの件数を正としている。

### D01 auth（スタンドアロン）

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P001 | /login | ログイン | Login | `features/auth/routes/Login.tsx` | D01 | スタンドアロン | 完了 | 2026-07-05 | `React.lazy()`+`Suspense` 直書き（router `lazy` 未使用）。未認証ユーザー専用バンドル |
| P002 | /forgot-password | パスワードのリセット | ForgotPasswordPage | `features/auth/routes/ForgotPasswordPage.tsx` | D01 | スタンドアロン | 完了 | 2026-07-05 | - |
| P003 | /reset-password | 新しいパスワードの設定 | ResetPasswordPage | `features/auth/routes/ResetPasswordPage.tsx` | D01 | スタンドアロン | 完了 | 2026-07-05 | 無効トークン時は「無効なリンクです」表示 |

### D02 reception

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P004 | / | 当日の受付 | Reception | `features/reception` | D02 | Layout内 | 完了 | 2026-07-05 | `RequirePermission resource={ResourceReception}` |

### D03 owners

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P005 | /owners | 飼主・ペット一覧 | OwnersListPage | `app/pages/OwnersListPage.tsx`（→`features/owners/routes/OwnersList.tsx`） | D03 | Layout内 | 完了 | 2026-07-05 | ownersLoader 併用 |
| P006 | /owners/new | 飼主・ペット登録 | OwnerFormPage | `app/pages/OwnerFormPage.tsx`（→`features/owners/routes/OwnerForm.tsx`） | D03 | Layout内 | 完了 | 2026-07-05 | BUG-020: `RequirePermission action="create"` |
| P007 | /owners/:id | 飼主・ペット編集 | OwnerFormPage | `app/pages/OwnerFormPage.tsx`（→`features/owners/routes/OwnerForm.tsx`） | D03 | Layout内 | 完了 | 2026-07-05 | ownerLoader 併用 |

### D04 aggregation

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P008 | /aggregation | 顧客集計ダッシュボード | AggregationDashboardPage | `features/aggregation/AggregationDashboardPage.tsx` | D04 | Layout内 | 完了 | 2026-07-05 | AggregationOwnerTable は DESIGN_TABLE_HEADER_ROW/CELL 採用済み |

### D05 reservations

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P009 | /reservations | 予約管理 | ReservationManagement | `features/reservations/routes/ReservationManagement.tsx` | D05 | Layout内 | 完了 | 2026-07-05 | - |

### D06 medical-records

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P010 | /medical-records | カルテ | MedicalRecords | `features/medical-records` | D06 | Layout内 | 完了 | 2026-07-05 | index |
| P011 | /medical-records/select-pet | カルテ作成 - ペット選択 | MedicalRecordPetSelection | `features/medical-records/routes/MedicalRecordPetSelection.tsx` | D06 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P012 | /medical-records/new | カルテ入力 | MedicalRecordForm | `features/medical-records/routes/MedicalRecordForm.tsx` | D06 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P013 | /medical-records/:id | カルテ編集 | MedicalRecordForm | `features/medical-records/routes/MedicalRecordForm.tsx` | D06 | Layout内 | 完了 | 2026-07-05 | 臨床安全 UI（危険バッジ等）は本タスクで無変更 |

### D07 hospitalization

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P014 | /hospitalization | 入院・ホテル | HospitalizationList | `features/hospitalization` | D07 | Layout内 | 完了 | 2026-07-05 | HospitalizationListView は DESIGN_TABLE_HEADER_ROW/CELL 採用済み |
| P015 | /hospitalization/select-pet | 入院・ホテル登録 - ペット選択 | HospitalizationPetSelection | `features/hospitalization/routes/HospitalizationPetSelection.tsx` | D07 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P016 | /hospitalization/new | 入院登録 | HospitalizationForm | `features/hospitalization/routes/HospitalizationForm.tsx` | D07 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P017 | /hospitalization/:id | 入院詳細・カルテ | HospitalizationDetail | `features/hospitalization/routes/HospitalizationDetail.tsx` | D07 | Layout内 | 完了 | 2026-07-05 | 死亡グレーアウト等の臨床安全 UI は本タスクで無変更・退行なし |
| P018 | /hospitalization/:id/edit | 入院編集 | HospitalizationForm | `features/hospitalization/routes/HospitalizationForm.tsx` | D07 | Layout内 | 完了 | 2026-07-05 | BUG-020: `RequirePermission action="edit"` |

### D08 trimming

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P019 | /trimming | トリミング | TrimmingList | `features/trimming` | D08 | Layout内 | 完了 | 2026-07-05 | TrimmingListTable は DESIGN_TABLE_HEADER_ROW/CELL 採用済み |
| P020 | /trimming/select-pet | トリミング登録 - ペット選択 | TrimmingPetSelection | `features/trimming/routes/TrimmingPetSelection.tsx` | D08 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P021 | /trimming/new | トリミング登録 | TrimmingForm | `features/trimming/routes/TrimmingForm.tsx` | D08 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P022 | /trimming/:id | トリミング編集 | TrimmingForm | `features/trimming/routes/TrimmingForm.tsx` | D08 | Layout内 | 完了 | 2026-07-05 | - |

### D09 examinations

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P023 | /examinations | 検査管理 | ExaminationsList | `features/examinations` | D09 | Layout内 | 完了 | 2026-07-05 | - |
| P024 | /examinations/select-pet | 検査登録 - ペット選択 | ExaminationPetSelection | `features/examinations/routes/ExaminationPetSelection.tsx` | D09 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P025 | /examinations/new | 新規検査登録 | ExaminationForm | `features/examinations/routes/ExaminationForm.tsx` | D09 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P026 | /examinations/:id | 検査詳細・編集 | ExaminationForm | `features/examinations/routes/ExaminationForm.tsx` | D09 | Layout内 | 完了 | 2026-07-05 | - |

### D10 vaccinations

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P027 | /vaccinations | 予防接種 | VaccinationList | `features/vaccinations` | D10 | Layout内 | 完了 | 2026-07-05 | - |
| P028 | /vaccinations/select-pet | ワクチン接種 - ペット選択 | VaccinationPetSelection | `features/vaccinations/routes/VaccinationPetSelection.tsx` | D10 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P029 | /vaccinations/new | 新規予防接種登録 | VaccinationForm | `features/vaccinations/routes/VaccinationForm.tsx` | D10 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P030 | /vaccinations/:id | 予防接種詳細・編集 | VaccinationForm | `features/vaccinations/routes/VaccinationForm.tsx` | D10 | Layout内 | 完了 | 2026-07-05 | - |

### D11 checkups

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P031 | /checkups | 定期健診 | CheckupsList | `features/checkups` | D11 | Layout内 | 完了 | 2026-07-05 | - |
| P032 | /checkups/select-pet | 定期健診登録 - ペット選択 | CheckupPetSelection | `features/checkups/routes/CheckupPetSelection.tsx` | D11 | Layout内 | 完了 | 2026-07-05 | 他ドメインと異なり `RequirePermission` 未使用（`/checkups` の RequirePermission 内で保護） |
| P033 | /checkups/new | 定期健診登録 | CheckupForm | `features/checkups/routes/CheckupForm.tsx` | D11 | Layout内 | 完了 | 2026-07-05 | new/編集ともタイトル文言は同一 |

### D12 inventory

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P034 | /inventory | 在庫管理 | InventoryList | `features/inventory` | D12 | Layout内 | 完了 | 2026-07-05 | InventoryList は DESIGN_TABLE_HEADER_ROW/CELL 採用済み |
| P035 | /inventory/new | 在庫登録 | InventoryForm | `features/inventory/routes/InventoryForm.tsx` | D12 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P036 | /inventory/:id | 在庫編集 | InventoryForm | `features/inventory/routes/InventoryForm.tsx` | D12 | Layout内 | 完了 | 2026-07-05 | - |

### D13 estimates

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P037 | /estimates | 見積管理 | EstimateList | `features/estimates` | D13 | Layout内 | 完了 | 2026-07-05 | EstimateList は DESIGN_TABLE_HEADER_ROW/CELL 採用済み |
| P038 | /estimates/new | 新規見積書作成 | EstimateForm | `features/estimates/routes/EstimateForm.tsx` | D13 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード |
| P039 | /estimates/:id | 見積書 | EstimateDetail | `features/estimates/routes/EstimateDetail.tsx` | D13 | Layout内 | 完了 | 2026-07-05 | タイトルは見積番号（動的、例: 見積書 EST-001） |
| P040 | /estimates/:id/edit | 見積書編集 | EstimateForm | `features/estimates/routes/EstimateForm.tsx` | D13 | Layout内 | 完了 | 2026-07-05 | BUG-020: `RequirePermission action="edit"` |

### D14 shifts

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P041 | /shifts | シフト管理 | ShiftCalendarPage | `features/shifts/routes/ShiftCalendarPage.tsx` | D14 | Layout内 | 完了 | 2026-07-05 | - |
| P042 | /settings/shift-templates | シフトテンプレートマスタ | ShiftTemplateSettings | `features/shifts/routes/ShiftTemplateSettings.tsx` | D14 | Layout内 | 完了 | 2026-07-05 | DESIGN_TABLE_HEADER_ROW/CELL 採用済み |
| P043 | /settings/shift-template（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D14 | Layout内 | リダイレクト | - | BUG-383: 旧URL → `/settings/shift-templates`（P042） |

### D15 accounting

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P044 | /accounting | 会計管理 | AccountingList | `features/accounting` | D15 | Layout内 | 完了 | 2026-07-05 | - |
| P045 | /accounting/select-pet | 会計登録 - ペット選択 | AccountingPetSelection | `features/accounting/routes/AccountingPetSelection.tsx` | D15 | Layout内 | 完了 | 2026-07-05 | `RequirePermission action="create"` |
| P046 | /accounting/new | 会計精算 | AccountingDetailPage | `app/pages/AccountingDetailPage.tsx`（→`features/accounting/routes/AccountingDetail.tsx`） | D15 | Layout内 | 完了 | 2026-07-05 | BUG-020: create ガード。本タスクで修正実施（警告バナー色トークン置換） |
| P047 | /accounting/:id | 会計精算 | AccountingDetailPage | `app/pages/AccountingDetailPage.tsx`（→`features/accounting/routes/AccountingDetail.tsx`） | D15 | Layout内 | 完了 | 2026-07-05 | `DailyAccountingTab.tsx` の印刷専用テーブルは Known Exclusion #1（§6）で対象外 |

### D16 cash-register

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P048 | /accounting/close | レジ締め | CashRegisterClosePage | `features/cash-register/routes/CashRegisterClosePage.tsx` | D16 | Layout内 | 完了（除外あり） | 2026-07-05 | `ClosePrintArea.tsx`（印刷専用）は Known Exclusion #1（§6） |
| P049 | /accounting/close/history | 締め履歴 | CashRegisterHistoryPage | `features/cash-register/routes/CashRegisterHistoryPage.tsx` | D16 | Layout内 | 完了（除外あり） | 2026-07-05 | 同上 |

### D17 accounting-reports

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P050 | /accounting/reports | 月次集計レポート | AccountingReportsPage | `features/accounting-reports/routes/AccountingReportsPage.tsx` | D17 | Layout内 | 完了（除外あり） | 2026-07-05 | `MonthlyReportPrintArea.tsx`（印刷専用）は Known Exclusion #1（§6） |

### D18 lstep

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P051 | /lstep/checkup-sync | 健診リマインダー抽出 | CheckupSyncPage | `features/lstep/checkup-sync/CheckupSyncPage.tsx` | D18 | Layout内 | 完了 | 2026-07-05 | `RequirePermission resource={ResourceHospitalSettings}` |
| P052 | /lstep/delivery-monitor | 自動配信トリガー監視 | LstepDeliveryMonitorPage | `features/lstep/pages/LstepDeliveryMonitorPage.tsx` | D18 | Layout内 | 完了 | 2026-07-05 | 本タスクで修正実施（配信失敗警告バナー色トークン置換 + `role="alert"`） |
| P053 | /lstep/analytics | Lステップ分析レポート | LstepAnalyticsPage | `features/lstep/pages/LstepAnalyticsPage.tsx` | D18 | Layout内 | 完了 | 2026-07-05 | `RequirePermission resource={ResourceLstepAnalytics}` |
| P054 | /settings/integrations/lstep | Lステップ連携設定 | LstepSettingsPage | `features/settings/integrations/lstep/LstepSettingsPage.tsx` | D18 | Layout内 | 完了 | 2026-07-05 | feature ソースは `features/settings`（§2 D18 の feature 列に記載あり） |
| P055 | /settings/lstep/tags | Lステップタグ管理 | LstepTagManagementPage | `features/lstep/pages/LstepTagManagementPage.tsx` | D18 | Layout内 | 完了 | 2026-07-05 | 本タスクで修正実施（`TagSummaryTable` を DESIGN_TABLE_HEADER_ROW/CELL に統一） |

### D19 line-reservation

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P056 | /line-reservation | 基本設定 | LineReservationSettings | `features/line-reservation/routes/LineReservationSettings.tsx` | D19 | Layout内 | 完了 | 2026-07-05 | index（`/line-reservation/settings` と同一コンポーネント） |
| P057 | /line-reservation/settings | 基本設定 | LineReservationSettings | `features/line-reservation/routes/LineReservationSettings.tsx` | D19 | Layout内 | 完了 | 2026-07-05 | - |
| P058 | /line-reservation/page-editor | ページ編集 | LineReservationPageEditor | `features/line-reservation/routes/LineReservationPageEditor.tsx` | D19 | Layout内 | 完了 | 2026-07-05 | - |
| P059 | /line-reservation/slots | LINE予約枠 | LineReservationSlotsSettings | `features/master/routes/LineReservationSlotsSettings.tsx` | D19 | Layout内 | 完了 | 2026-07-05 | feature ソースは `features/master`（path は `/line-reservation/*` のため §2 の D19 route pattern に従う）。権限は `ResourceMasterReservationType` |

### D20 clinic-settings

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P060 | /settings/clinic | 医院マスタ | ClinicMasterSettings | `features/clinic-settings/routes/ClinicMasterSettings.tsx` | D20 | Layout内 | 完了 | 2026-07-05 | grep 一致は全て `#190` 等 Issue 番号コメント |

### D21 master-settings

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P061 | /settings | マスタ設定 | MasterSettingsIndex | `features/master/routes/MasterSettingsIndex.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | index。ガード不要（BUG-123 でカードフィルタリング対応） |
| P062 | /settings/staff | スタッフマスタ | StaffSettings | `features/master/routes/StaffSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P063 | /settings/treatment-items | 治療プランマスタ | TreatmentPlanMaster | `features/master/routes/TreatmentPlanMaster.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | `?tab=examination\|vaccine\|consultation\|procedure` を受け付ける（P088–P091 の遷移先） |
| P064 | /settings/diagnosis | 診断病名マスタ | DiagnosisSettings | `features/master/routes/DiagnosisSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | `?tab=diagnosis_type\|diagnosis_name` を受け付ける（P084–P085 の遷移先） |
| P065 | /settings/animal-species | 動物種類マスタ | AnimalSpeciesSettings | `features/master/routes/AnimalSpeciesSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P066 | /settings/trimming | トリミングマスタ | TrimmingSettings | `features/master/routes/TrimmingSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | `?tab=course\|option` を受け付ける（P086–P087 の遷移先） |
| P067 | /settings/trimming-course-type | トリミングコース種別マスタ | TrimmingCourseTypeSettings | `features/master/routes/TrimmingCourseTypeSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P068 | /settings/medicine | 薬剤マスタ | MedicineSettings | `features/master/routes/MedicineSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P069 | /settings/reservation-type | 予約区分マスタ | ReservationTypeSettings | `features/master/routes/ReservationTypeSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P070 | /settings/hospitalization | 入院マスタ | HospitalizationSettings | `features/master/routes/HospitalizationSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P071 | /settings/cage | ケージマスタ | CageSettings | `features/master/routes/CageSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P072 | /settings/merchandise-items | 物販・その他マスタ | MerchandiseItemSettings | `features/master/routes/MerchandiseItemSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P073 | /settings/insurance | 保険マスタ | InsuranceSettings | `features/master/routes/InsuranceSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P074 | /settings/occupations | 職種マスタ | OccupationSettings | `features/master/routes/OccupationSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P075 | /settings/permission-groups | 権限グループマスタ | PermissionGroupSettings | `features/master/routes/PermissionGroupSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P076 | /settings/inquiry-templates | 問診テンプレートマスタ | InterviewTemplateSettings | `features/master/routes/InterviewTemplateSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | P078 と同一コンポーネント（P092 の遷移先） |
| P077 | /settings/interview/chief-complaint | 主訴マスタ | ChiefComplaintSettings | `features/master/routes/ChiefComplaintSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | - |
| P078 | /settings/interview/templates | 問診テンプレートマスタ | InterviewTemplateSettings | `features/master/routes/InterviewTemplateSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | P076 と同一コンポーネント |
| P079 | /settings/closing-time | 締め時間設定 | ClosingSettingsPage | `features/closing-settings/routes/ClosingSettingsPage.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | FEAT-368 / STG-BLOCKER-002: `RequirePermission` 追加済み |
| P080 | /settings/payment-methods | 支払方法マスタ | PaymentMethodSettings | `features/master/routes/PaymentMethodSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | FEAT-368 / STG-BLOCKER-002: `RequirePermission` 追加済み |
| P081 | /settings/campaigns | 割引キャンペーンマスタ | CampaignSettings | `features/master/routes/CampaignSettings.tsx` | D21 | Layout内 | 完了 | 2026-07-05 | #81: `ResourceAccounting` 権限（会計割引マスタのため） |
| P082 | /settings/job-title（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | BUG-382/384: → `/settings/occupations`（P074） |
| P083 | /settings/service-type（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/reservation-type`（P069） |
| P084 | /settings/diagnosis-type（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/diagnosis?tab=diagnosis_type`（P064） |
| P085 | /settings/diagnosis-name（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/diagnosis?tab=diagnosis_name`（P064） |
| P086 | /settings/trimming-course（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/trimming?tab=course`（P066） |
| P087 | /settings/trimming-option（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/trimming?tab=option`（P066） |
| P088 | /settings/examination（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/treatment-items?tab=examination`（P063） |
| P089 | /settings/vaccine（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/treatment-items?tab=vaccine`（P063） |
| P090 | /settings/consultation（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/treatment-items?tab=consultation`（P063） |
| P091 | /settings/procedure（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/treatment-items?tab=procedure`（P063） |
| P092 | /settings/inquiry-template（redirect） | - | Navigate | `app/routes/settings-routes.tsx` | D21 | Layout内 | リダイレクト | - | → `/settings/inquiry-templates`（P076） |

### D22 manual

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P093 | /manual | 取扱説明書 | ManualPage | `features/manual/routes/ManualPage.tsx` | D22 | Layout内 | 完了 | 2026-07-05 | 記事一覧（index） |
| P094 | /manual/:category/:slug | 取扱説明書（記事） | ManualPage | `features/manual/routes/ManualPage.tsx` | D22 | Layout内 | 完了 | 2026-07-05 | 記事ごとの Markdown 見出しが動的タイトル |

### Standalone（D23 owner-report / 404）

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P095 | /owners/:id/report | 飼主レポート | OwnerReport | `features/owner-report/routes/OwnerReport.tsx` | D23 | スタンドアロン | 完了 | 2026-07-05 | #158: 別ウィンドウ用に Layout（サイドバー）外に登録。認証ガード・`medical-records:view` ゲートは OwnerReport 自身が保持 |
| P096 | * (404) | ページが見つかりません | （inline element） | `app/routes/app-routes.tsx`（`notFoundRoute`） | - | Layout内 | 完了 | 2026-07-05 | **実装上の注記**: `notFoundRoute` は `appRoutes` の `{ element: <Layout />, children: [...] }` の子として登録されており、実際は Layout 内（サイドバー表示）でレンダリングされる。router `lazy` は未使用（inline element） |

### 対象外（別アプリ）

| P-ID | path | ページ名 | コンポーネント | ソース | ドメイン | Layout | status | 最終監査 | 備考 |
|---|---|---|---|---|---|---|---|---|---|
| P097 | frontend/liff/* | - | - | `frontend/liff` | - | - | 対象外 | - | 別アプリのため対象外（Known Exclusion #4、§6） |
| P098 | frontend/line-reserve/* | - | - | `frontend/line-reserve` | - | - | 対象外 | - | 別アプリのため対象外（Known Exclusion #4、§6） |

---

## 4. Shared Component Registry（共通コンポーネント表）

**SSOT**: `frontend/src/components/shared/*`, `frontend/src/components/ui/*`。DESIGN.md/DESIGN_SYSTEM.md §6 に関連する横断 UI プリミティブのみを対象とし、全 97 shared ファイルは列挙しない。status は §2 D24（shared-shell）の 2026-07-05 監査結果（`components/shared/*`, `components/ui/*` 配下 hex/tw 違反 0）を各コンポーネントに展開したもの。

### 列定義

| 列 | 内容 |
|---|---|
| C-ID | `C001` 連番 |
| コンポーネント | 名前 |
| パス | `frontend/src/...` |
| DESIGN.md 対応 | `docs/DESIGN_SYSTEM.md` §6 の ID（`ex-data-table-cell` / `ex-modal-card` / `text-input` / `button-primary` / `badge-pill` 等）。§6 に明示 ID がない場合は最も近い分類を記載 |
| status | `完了` / `完了（除外あり）` / `部分準拠` |
| 備考 | Known Exclusion 参照・opt-in パターン等 |

### 4.1 App Shell（D24）

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C001 | Layout | `frontend/src/components/shared/Layout/Layout.tsx` | shared-shell（D24、個別ID対応なし） | 完了 | 全ページ共通の `<Outlet/>` ラッパー |
| C002 | Sidebar | `frontend/src/components/shared/Layout/Sidebar.tsx` | shared-shell（D24、個別ID対応なし） | 完了 | §2 D24 で明示的に準拠確認済み |
| C003 | SidebarItems | `frontend/src/components/shared/Layout/SidebarItems.tsx` | shared-shell（D24、個別ID対応なし） | 完了 | sidebar-menu.tsx のメニュー定義を描画 |
| C004 | FormHeader | `frontend/src/components/shared/Form/FormHeader.tsx` | feature-card 系（見出し） | 完了 | 各フォームページのタイトル表示に使用 |
| C005 | PageLayout | `frontend/src/components/shared/PageLayout/PageLayout.tsx` | shared-shell（D24、個別ID対応なし） | 完了 | - |

### 4.2 Data Display

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C006 | DataTable | `frontend/src/components/shared/DataTable/DataTable.tsx` | `ex-data-table-cell` | 完了 | `DESIGN_TABLE_HEADER_ROW`/`DESIGN_TABLE_HEADER_CELL` の提供元。§2 D24 で明示 |
| C007 | DataTableRow | `frontend/src/components/shared/DataTable/DataTableRow.tsx` | `ex-data-table-cell`（body） | 完了 | `STYLE.tableCell`（body-sm 相当）使用 |
| C008 | SortableDataTableRow | `frontend/src/components/shared/DataTable/SortableDataTableRow.tsx` | `ex-data-table-cell`（body） | 完了 | - |
| C009 | SortableHeader | `frontend/src/components/shared/SortableHeader/SortableHeader.tsx` | `ex-data-table-cell`（header） | 完了 | §2 D24 で明示的に準拠確認済み |
| C010 | DESIGN_TABLE_HEADER_ROW / DESIGN_TABLE_HEADER_CELL（定数） | `frontend/src/components/shared/DataTable/DataTable.tsx` | `ex-data-table-cell`（header: canvas-soft + eyebrow） | 完了 | 18 ファイルで採用（D04/D08/D12/D13/D14/D18/D20 等の一覧テーブル） |

### 4.3 Forms & Buttons

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C011 | SubmitButton | `frontend/src/components/shared/Form/SubmitButton.tsx` | `button-primary` | 完了 | 既定 `colorVariant` を `"brand"` に変更済み（2026-07-05 追記2、Known Exclusion #3 解消）。`colorVariant="default"`（旧 accent）は opt-out として残置し `check-design-primary-cta.mjs` が再混入を検出 |
| C012 | PrimaryButton | `frontend/src/components/shared/Form/PrimaryButton.tsx` | `button-primary` | 完了 | 同上 |
| C013 | Input | `frontend/src/components/ui/input.tsx` | `text-input` | 完了 | `rounded-xs`（4px）に変更済み（2026-07-05 追記2、Known Exclusion #2 解消）。`globals.css` に `--radius-xs: 4px` を追加 |
| C014 | Textarea | `frontend/src/components/ui/textarea.tsx` | `text-input` | 完了 | 同上 |
| C015 | SelectTrigger | `frontend/src/components/ui/select.tsx` | `text-input` | 完了 | 同上（`SelectContent` 等の非 text-input 部分は対象外のまま） |
| C016 | FormDialog | `frontend/src/components/shared/FormDialog/FormDialog.tsx` | `ex-modal-card` | 完了 | shadcn `dialog.tsx` ベース |
| C017 | PropertyInput（PropInput） | `frontend/src/components/shared/SidePeek/PropertyInput.tsx` | `text-input`（ボーダーレス変種、DESIGN_SYSTEM.md §6.1） | 完了 | ホバー/フォーカスで初めて枠が現れるプロパティ編集パターン |
| C035 | SearchableSelect | `frontend/src/components/ui/searchable-select.tsx` | `text-input` | 完了 | 2026-07-05 再監査で `rounded-md` → `rounded-xs` に修正（`SelectTrigger` との角丸不一致を解消）。D02/D03/D06/D07/D09/D11/D14/D21 等 18 ファイルから参照 |

### 4.4 Overlays

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C018 | dialog（Dialog） | `frontend/src/components/ui/dialog.tsx` | `ex-modal-card` | 完了 | `rounded-xl` + `p-6` + `shadow-lg`（DESIGN_SYSTEM.md §5 で明示的に準拠記載） |
| C019 | alert-dialog（AlertDialog） | `frontend/src/components/ui/alert-dialog.tsx` | `ex-modal-card` | 完了 | dialog.tsx と同系統のプリミティブ |
| C020 | sheet（Sheet） | `frontend/src/components/ui/sheet.tsx` | `ex-modal-card`（サイドスライド変種） | 完了 | §2 D24 の hex/tw 違反 0 判定に含まれる（`components/ui/*` 全体） |
| C021 | SidePeekPanel | `frontend/src/components/shared/SidePeek/SidePeekPanel.tsx` | サイドピークパネル（DESIGN_SYSTEM.md §6.2） | 完了 | 一覧画面から詳細を覗き見る形式 |
| C022 | SidePeekFooter | `frontend/src/components/shared/SidePeek/SidePeekFooter.tsx` | `button-primary` / `ex-modal-card` | 完了 | `c19b0dc7` で対応済み（master/PATTERNS.md 記載） |
| C023 | SidePeekBody | `frontend/src/components/shared/SidePeek/SidePeekBody.tsx` | `ex-modal-card` | 完了 | - |
| C024 | SidePeekToolbar | `frontend/src/components/shared/SidePeek/SidePeekToolbar.tsx` | `ex-modal-card` | 完了 | - |
| C025 | PrintPortal | `frontend/src/components/shared/PrintPortal.tsx` | 対象外（印刷専用、Known Exclusion #1） | 完了（除外あり） | `@media print` 専用スタイル注入。固定属性 `data-print-portal` で複数ポータル同居時の相殺を防止（#187） |

### 4.5 Filters & Navigation

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C026 | NotionFilter | `frontend/src/components/shared/NotionFilter/NotionFilter.tsx` | `ex-data-table-cell` 系フィルタUI | 完了 | §2 D24 の hex/tw 違反 0 判定に含まれる |
| C027 | CategoryChipsFilter | `frontend/src/components/shared/CategoryChipsFilter/CategoryChipsFilter.tsx` | `badge-pill` | 完了 | 同上 |
| C028 | ClinicScopeFilter | `frontend/src/components/shared/ClinicScopeFilter/ClinicScopeFilter.tsx` | `badge-pill` | 完了 | 同上 |
| C029 | UnifiedTabs | `frontend/src/components/shared/UnifiedTabs.tsx` | `badge-pill`（active タブ表現） | 完了 | 同上 |
| C030 | NotionDatePicker | `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` | `text-input` | 完了 | 同上 |

### 4.6 Feedback & Status

| C-ID | コンポーネント | パス | DESIGN.md 対応 | status | 備考 |
|---|---|---|---|---|---|
| C031 | StatusBadge | `frontend/src/components/shared/StatusBadge/StatusBadge.tsx` | `badge-pill` | 完了 | 配色は呼び出し側の `colorClass` に依存（コンポーネント自体は shadcn `Badge` ラッパー） |
| C032 | NotionStatusPill | `frontend/src/components/shared/StatusPill/NotionStatusPill.tsx` | `badge-pill` | 完了 | - |
| C033 | FilteringIndicator | `frontend/src/components/shared/FilteringIndicator/FilteringIndicator.tsx` | 非同期フィードバック（DESIGN_SYSTEM.md §6.3） | 完了 | 大規模データ検索中の透過アニメーション表現 |
| C034 | PermissionBadges | `frontend/src/components/shared/PermissionBadges/PermissionBadges.tsx` | `badge-pill` | 完了 | - |

---

## 5. Audit Procedure（監査手順）

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

## 6. Known Exclusions（既知の除外事項）

以下は本タスクのスコープ外、または DESIGN.md の適用対象外として明示的に除外する。

| # | 対象 | 除外理由 |
|---|---|---|
| 1 | 印刷専用コンポーネント（`DailyAccountingTab.tsx` の `DailyPrintArea`、`MonthlyReportPrintArea.tsx`、`ClosePrintArea.tsx`） | `hidden` 属性 + `@media print` でのみ表示される A4 帳票・帳簿のレプリカ。画面上の「ページ canvas」ではなく物理紙面の再現が目的のため、DESIGN.md のスクリーン UI トークン規約の対象外。Tailwind の `gray-300`/`gray-400` 等の生色グリッド線は帳簿の視認性を保つための意図的な選択であり変更しない |
| 2 | ~~`Input` / `SelectTrigger` の `rounded-md`（8px）~~ | **対応済み（2026-07-05 追記2）**。`globals.css` に `--radius-xs: 4px` を追加し、`Input` / `Textarea` / `SelectTrigger` を `rounded-md` → `rounded-xs`（DESIGN.md `{rounded.xs}`）に変更。`SelectContent`/`button-variants.ts`（`button-utility` = `{rounded.md}`）等、DESIGN 意味論上 `rounded-md` が正当な箇所は対象外のまま維持 |
| 3 | ~~`C.accent` → `C.bgBrand` のアプリ全体一括移行（`SubmitButton`/`PrimaryButton` 既定 `colorVariant` 変更を含む）~~ | **対応済み（2026-07-05 追記2）**。`SubmitButton`/`PrimaryButton` の既定 `colorVariant` を `"brand"` に変更し、48 ファイルの call site から冗長な `colorVariant="brand"` を削除。`PetEditModal.tsx` の `STYLE.confirmPrimary` 上書きハックも `PrimaryButton` 使用に正規化。`design-tokens.ts` の `STYLE.confirmPrimary` 定義本体・`colorVariant="default"` opt-out は残置（Known Exclusion #6 参照） |
| 4 | `frontend/liff`, `frontend/line-reserve` | 別アプリのため対象外 |
| 5 | `change-ui.md`（受付テレメトリー機能 / `checked_in_at` BE 連携） | 別タスクのため対象外 |
| 6 | `design-tokens.ts` の大規模リファクタ | 対象外。今回は既存トークン（`C.danger`/`C.bgDanger8`/`C.borderDanger20`）のみを使用し、新規トークン追加は行っていない |
| 7 | ~~`frontend/src/features/master/PATTERNS.md` 内のコード例示~~ | **対応済み（2026-07-05）**。内部向けパターン集ドキュメントに残っていた旧 hex 直書き例（`text-[#37352F]`、`bg-[#2383E2]` 等、計8箇所）を `C.text` / `C.text20` / `C.text65` / `C.danger` / `C.hoverBgDanger5` / `C.bgAccent` / `C.bgAccentLight` / `C.textAccentDark` / `C.bgPrimary10` / `C.bgInactive` / `C.text60` / `C.accent` / `C.hoverTextAccent` / `C.dataActiveBorderB` / `C.dataActiveText` に置換し、「サイドピークのトークン一覧」表も `design-tokens.ts` の合成元表記に更新。例の意味（SidePeek 保存ボタン・削除ボタン・バッジ色・タブ active 状態）は維持。ドキュメントのみの変更で実行コードは無変更 |

---

## 7. 完了条件チェック

| # | 条件 | 結果 |
|---|---|---|
| 1 | `docs/UI_DESIGN_COMPLIANCE.md` が存在し D01–D24 全行に status がある | PASS（本ファイル） |
| 2 | 全ドメイン status=完了（Known Exclusions のみ「完了（除外あり）」可） | PASS（D16/D17 が「完了（除外あり）」、他 22 ドメインが「完了」） |
| 3 | `node scripts/check-design-primary-cta.mjs` が exit 0 | PASS |
| 4 | 各ドメインで hex 直書き・Tailwind 生色・旧 accent Primary CTA が残っていない | PASS（§6 の Known Exclusions を除く） |
| 5 | 変更ドメインの scoped vitest が PASS | PASS（`docker compose exec frontend npx vitest run src/features/accounting src/features/lstep` → 19 test files / 210 passed, 3 skipped, 0 failed） |
| 6 | 臨床安全 UI の既存テストが退行していない | PASS（D06/D07 は本タスクで無変更。D15/D18 は警告バナーの色トークン置換のみで表示条件・disabled/権限ロジックは不変。code-reviewer subagent による重点レビュー実施: CRITICAL/HIGH 0件、APPROVE（条件付き）。指摘事項は D18 `TagSummaryTable.tsx` のヘッダーが `DESIGN_TABLE_HEADER_CELL`（eyebrow: uppercase + tracking-wide）へ変わる視覚差分のみで、これは DESIGN_SYSTEM.md §6 の `ex-data-table-cell` header 仕様への準拠そのものであり意図した変更） |
| 7 | Page Registry（§3）に appRoutes の lazy ルート + Navigate redirect が漏れなく載っている | PASS（§3 参照。82 Component + 12 Navigate + 1 login + 1 404 + 2 対象外 = 98 行。詳細は §3 の AC-1 件数根拠を参照） |
| 8 | Shared Component Registry（§4）に DESIGN_SYSTEM.md §6 相当の横断コンポーネントが表形式で載っている | PASS（35 行、§4.1–4.6。2026-07-05 追記3 で C035 `SearchableSelect` を追加） |
| 9 | C011/C012（SubmitButton/PrimaryButton）が `完了`（`部分準拠` 解消） | PASS（§4.3。既定 `colorVariant="brand"`） |
| 10 | C013–C015（Input/Textarea/SelectTrigger）が `完了`（`完了（除外あり）` 解消） | PASS（§4.3。`rounded-xs` 4px） |
| 11 | `node scripts/check-design-primary-cta.mjs` が既定 brand 化後も exit 0 | PASS（`colorVariant="default"` 検出ルール追加後も本番コード scan で違反 0。2026-07-05 追記3 でローカル node 実行により再確認） |
| 12 | feature 層に Primary CTA としての `STYLE.confirmPrimary` 直指定が残っていない | PASS（`rg 'STYLE\.confirmPrimary' frontend/src/features --glob '!*.test.*'` → 0 件。`components/shared/Form/SubmitButton.tsx` 自身（`colorVariant="default"` opt-out の実装本体）には 2 件ヒットするが、これは `check-design-primary-cta.mjs` の `EXCLUDE_FILE` が明示的に除外する定義ファイルであり violation ではない。2026-07-05 追記3 で対象範囲の記述を実測値に合わせて修正） |
| 13 | `input.tsx`/`textarea.tsx`/`select.tsx`（SelectTrigger）/`searchable-select.tsx`（trigger）に `rounded-md` が残っていない | PASS（4 ファイルとも `rounded-xs`。2026-07-05 追記3 で `searchable-select.tsx` を追加修正。`SelectContent` 等の非 text-input は対象外で `rounded-md` 維持） |
| 14 | §6 Known Exclusions #2/#3 が対応済みとして反映されている | PASS（取り消し線 + 「対応済み（2026-07-05 追記2）」） |
| 15 | 2026-07-05 追記3（検証ファースト再監査）: 全24ドメイン再監査 + Page Registry/Shared Component Registry の件数整合確認 | PASS（Phase 0 再監査表は本タスクの実行記録を参照。新規 violation は `searchable-select.tsx` 1件のみで修正済み。§1/C025 矛盾・§7 #12 の記述不正確を解消） |
