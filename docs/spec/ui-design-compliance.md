# UI Design Compliance — 全ルートページ準拠監査

> 正本。[DESIGN.md](../../DESIGN.md)（タイポ・形状・余白・エレベーション・寸法）と [docs/spec/design-system.md](design-system.md)（AE 製品色）への準拠を、本体85リーフルートとその実装コンポーネントに対して判定した結果。最終監査日 **2026-07-24**（§2 在庫は 2026-07-31 に `/identity-links` を追加し 85 リーフへ更新）。`frontend/docs/design-audit-pages.md` は本書に統合済みで更新終了。

## 1. 準拠チェック定義

| ID | 観点 | 参照 | 判定方法 |
|---|---|---|---|
| C1 | brand と semantic primary は同じ teal `#038B94` / active `#027078` を使う。意味役割はトークン名で分け、legacy accent `#2383E2`・`C.accent` は禁止 | DESIGN_SYSTEM §2.1, §9, §10 | **機械化**（C1）＋意味役割はレビュー |
| C2 | ページ canvas は暖色 canvas-soft（`PageLayout` / `STYLE.page*` / `C.bgPage`）。純白ページ禁止 | DESIGN_SYSTEM §2.2, §9 | **C8 機械化**（`src/features/*/routes/*.tsx` の PageLayout / Master*Page / allowlist）。§2 表は新規リーフ追加時に更新 |
| C3 | コンポーネントでの hex 直書き禁止。`design-tokens.ts` 経由必須 | DESIGN_SYSTEM §9 Don't, §10 | **機械化**（C3。issue番号コメントは対象外） |
| C4 | 一覧/詳細系ページは `PageLayout` または同等 shell を持つ | DESIGN_SYSTEM §4, §7.6 | **C8 機械化**（C2 と同一。shell 欠落 = C2/C4 同時フラグ） |
| C5 | 汎用 Primary CTA は `colorVariant="primary"`（既定）、認証・製品識別 CTA だけ `colorVariant="brand"` を明示する | DESIGN_SYSTEM §7.2 | **機械化**（C5 の許可値）＋意味役割はレビュー |
| C6a | 臨床安全 UI（危険/死亡/RBAC 非活性表示）はデザイン変更で退行させない | DESIGN_SYSTEM §2.4, §9 | 静的 grep では網羅不可 — **コードレビュー要**。danger/warning の hex 直書き逸脱は C3 と合わせて確認 |
| C6b | rgba/rgb/hsla/hsl 直値禁止 | design-system-audit.mjs C6 | **機械化**（`pnpm design-audit` / make ci） |
| C7 | PageLayout `maxWidth` 生値禁止（`max-w-full` / `max-w-[Npx]`） | LAYOUT.pageContentMaxWidth | **機械化**（C7） |
| C8 | routes/*.tsx は PageLayout / Master*Page / allowlist | PageLayout | **機械化**（C8・相対パス完全一致 allowlist 14件 = 独自page 9 + helper 5） |
| C9 | `rounded-[Npx]` 任意値禁止 | `--radius-xxs/xs/...` | **機械化**（C9） |
| C10 | Tailwind 既定影（`shadow-2xs〜2xl`）と `shadow-[...]` 任意値禁止。`shadow-level1/level2/btn/panel/focus-brand` 等の elevation トークンのみ使用可 | design-system.md §5.1 | **機械化**（`pnpm design-audit` / make ci） |
| C11 | `text-[Npx\|Nrem]` font-size 任意値禁止 | design-system.md §3.4 | **機械化**（C11） |
| C12 | DESIGN.md にない `text-lg/2xl/3xl/4xl+` 禁止。`text-heading-1/2/3`・`text-xl/base/sm/xs/2xs` へ写像 | design-system.md §3.4 | **機械化**（C12。本体85ルート対象） |
| C13 | ink 4段を迂回する黒アルファ text / placeholder 禁止 | design-system.md §2.1, §3.4 | **機械化**（C13） |
| C14 | ロール固有 letter-spacing を上書きする `tracking-*` built-in / 任意値 / shorthand 禁止 | design-system.md §3.1, §3.4 | **機械化**（C14。本体85ルート対象） |
| C15 | 本体 routes/pages の white/black named color 直接指定禁止。色token経由必須 | design-system.md §2, §9 | **機械化**（C15） |
| C16 | spacing scale にない20px utility（`*-5`、負値、`[20px]` / `[1.25rem]`）禁止 | design-system.md §4.1 | **機械化**（C16） |
| C17 | CSS の直接 `box-shadow:` / `filter: drop-shadow()` 禁止。elevation token経由必須 | design-system.md §5.1 | **機械化**（C17） |
| C18 | `TableHead` / `TableCell` とraw `th` / `td` で header eyebrow・body-sm・12px vertical / 16px horizontal paddingを非仕様値へ上書きしない | DESIGN.md `ex-data-table-cell` | **機械化**（C18。test除外、空stateと`data-c18-structural-cell`付きself-closing構造セルは許可。raw は `C18_RAW_CELL_ALLOWLIST` / `*PrintArea*` / FE11除外のみ。件数ratchetなし・違反はstrict fail） |
| C19 | `DataTableRow` / `SortableDataTableRow` / `TableRow` / raw `tr` に行全体の `onClick` を付けない。遷移はcell内のnative link、表示・編集はnative button、並べ替えは44px drag handleを使う | DESIGN.md §7.4, §8.2 | **機械化**（C19。production `.tsx` の4種のrow opening tagをmultiline対応で検査） |

**FE11 正本分割（2026-07-21、色役割更新 2026-07-27）**: 色は DESIGN_SYSTEM の製品判断、タイポ・形状・余白・エレベーション・寸法は DESIGN.md 字義を正本とする。brand と primary は同じ teal 値を使い、製品識別と汎用操作・選択・focus の意味役割だけをトークン名で分ける。臨床 semantic 色・業務 status 色・nav canvas-soft も DESIGN_SYSTEM の判断を維持する。

**監査範囲の限界**: `design-system-audit.mjs` は `src`・`liff/src`・`line-reserve/src` の非test TypeScriptを全数走査し、C17だけは CSS も走査する。FE11の対象は本体85ルートのため、C12/C14〜C17は LIFF・line-reserve・shared-liff を明示除外し、画面用規範と異なる `MedicalRecordPrintView` も除外する。C15 は本体 routes/pages、C18は `.tsx` の共通 Table primitive（`TableHead` / `TableCell`）とraw `th` / `td` opening tag、C19は4種のrow opening tagを対象にする。C18 は呼び出し側による非仕様 typography / vertical padding（raw は横padding・セル背景も含む）の再上書きを検出し、検出1件でも strict failする（件数ratchetなし。raw の正当な除外は `C18_RAW_CELL_ALLOWLIST`・`*PrintArea*` basename・FE11除外のみ）。ロール内での意味的誤選択、全viewport、C6a（臨床安全 UI）は静的判定できないためコードレビュー/ブラウザ確認を併用する。

**機械化（FE11 更新）**: C1/C3/C5/C6b/C7〜C19 は `frontend/scripts/design-system-audit.mjs` で strict fail（C18 も件数ratchetなし。違反0件が合格条件）。C2/C4 は C8 により `routes/*.tsx` を機械化済み。§2 表と C8 allowlist は新規リーフ追加時に同一コミットで更新する。C6a（臨床安全 UI）は引き続きコードレビュー要。
```bash
docker compose exec frontend pnpm design-audit
```

## 2. ページ別対応状況表

**サマリー: 機械適合 84 / P1〜P10・該当V1〜V8 完全確認 84 / BLOCKED 0 / 対象外 1（合計 85 リーフルート）**（在庫更新 2026-07-31: `/identity-links` 追加。runtime 全数再確認の基準日は 2026-07-23 の 83 製品ページ監査）。製品ページ 84（+ 404 対象外 1 = 本体 85）を対象とする。2026-07-23 時点では 83 製品ページを 1440 / 1200 / 800 / 500px で production component 経由描画し、console error、HTTP 4xx/5xx、document overflow、無名 control、44px 未満 control、continued business non-GET がすべて 0 であることを確認した。`/identity-links` を含む 84 製品の再 runtime は e2e inventory 更新後に再確認する。臨床 5 route は generated API type・feature type・transform test に適合する明白な sentinel fixture だけを Playwright 内で fulfill し、共有 DB への business write は 0 件とした。最終 runtime 正本は `frontend/e2e/ui-design-compliance-readonly.spec.ts` と `frontend/e2e/fixtures/ui-design-clinical.ts`。

| エリア | ページ名 | パス | コンポーネント | 機械適合 | 備考 |
|---|---|---|---|---|---|
| 認証/共通 | ログイン | /login | Login | ✅ | |
| 認証/共通 | パスワードを忘れた方 | /forgot-password | ForgotPasswordPage | ✅ | |
| 認証/共通 | パスワード再設定 | /reset-password | ResetPasswordPage | ✅ | |
| 認証/共通 | 飼主カルテレポート(単独ウィンドウ) | /owners/:id/report | OwnerReport | ✅ | |
| 認証/共通 | 404 Not Found | * | (inline element) | — | lazy Component非使用・簡易fallback |
| 会計 | 会計一覧 | /accounting | AccountingList | ✅ | |
| 会計 | 会計 - ペット選択 | /accounting/select-pet | AccountingPetSelection | ✅ | |
| 会計 | 会計登録(新規) | /accounting/new | AccountingDetailPage→AccountingDetail | ✅ | 薄いwrapper、shell実体はAccountingDetail |
| 会計 | 会計詳細 | /accounting/:id | AccountingDetailPage→AccountingDetail | ✅ | 同上 |
| 会計 | レジ締め | /accounting/close | CashRegisterClosePage | ✅ | PageLayout化済（2026-07-06） |
| 会計 | レジ締め履歴 | /accounting/close/history | CashRegisterHistoryPage | ✅ | PageLayout化済（2026-07-06） |
| 会計 | 月次集計レポート | /accounting/reports | AccountingReportsPage | ✅ | PageLayout化済（2026-07-06、既知baseline解消） |
| 在庫/見積/シフト | 在庫一覧 | /inventory | InventoryList | ✅ | |
| 在庫/見積/シフト | 在庫登録(新規) | /inventory/new | InventoryForm | ✅ | |
| 在庫/見積/シフト | 在庫編集 | /inventory/:id | InventoryForm | ✅ | |
| 在庫/見積/シフト | 見積一覧 | /estimates | EstimateList | ✅ | |
| 在庫/見積/シフト | 見積作成(新規) | /estimates/new | EstimateForm | ✅ | |
| 在庫/見積/シフト | 見積詳細 | /estimates/:id | EstimateDetail | ✅ | |
| 在庫/見積/シフト | 見積編集 | /estimates/:id/edit | EstimateForm | ✅ | |
| 在庫/見積/シフト | シフトカレンダー | /shifts | ShiftCalendarPage | ✅ | |
| カルテ | カルテ一覧 | /medical-records | MedicalRecords | ✅ | |
| カルテ | カルテ作成 - ペット選択 | /medical-records/select-pet | MedicalRecordPetSelection | ✅ | |
| カルテ | カルテ作成(新規) | /medical-records/new | MedicalRecordForm | ✅ | synthetic pet GET、予約/カルテ作成POSTをlocal fulfillし、synthetic detailへproduction遷移。backend write 0 |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | ✅ | |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | ✅ | |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | ✅ | |
| 入院/ホテル | 入院・ホテル登録(新規) | /hospitalization/new | HospitalizationForm | ✅ | |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | ✅ | typed `BackendHospitalization` / care-plan fixtureで臨床状態・担当医・単価を4 viewport確認 |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | ✅ | typed recordのform/treatment-plan hydration、治療・割引の参照専用状態、費用、mount時business non-GET 0を確認 |
| トリミング | トリミング一覧 | /trimming | TrimmingList | ✅ | |
| トリミング | トリミング登録 - ペット選択 | /trimming/select-pet | TrimmingPetSelection | ✅ | |
| トリミング | トリミング登録(新規) | /trimming/new | TrimmingForm | ✅ | |
| トリミング | トリミング編集 | /trimming/:id | TrimmingForm | ✅ | typed `BackendTrimming` とcourse/options/staff master fixtureでedit stateを4 viewport確認 |
| 検査 | 検査一覧 | /examinations | ExaminationsList | ✅ | |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | ✅ | |
| 検査 | 検査登録(新規) | /examinations/new | ExaminationForm | ✅ | |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | ✅ | |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | ✅ | |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | ✅ | |
| ワクチン | ワクチン登録(新規) | /vaccinations/new | Navigate → /vaccinations/select-pet | ✅ | /vaccinations/new は select-pet へリダイレクト。新規作成はカルテ予防接種タブから |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | ✅ | typed `BackendVaccination` とpet/vaccine/doctor fixtureでedit state・担当医帰属を4 viewport確認 |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | ✅ | |
| 定期健診 | 定期健診登録 - ペット選択 | /checkups/select-pet | CheckupPetSelection | ✅ | |
| 定期健診 | 定期健診登録(新規) | /checkups/new | CheckupForm | ✅ | |
| 受付/飼主/予約 | 受付(当日) | / | Reception | ✅ | |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersListPage→OwnersList | ✅ | 薄いwrapper、shell実体はOwnersList |
| 受付/飼主/予約 | 飼主登録(新規) | /owners/new | OwnerFormPage→OwnerForm | ✅ | 薄いwrapper、shell実体はOwnerForm |
| 受付/飼主/予約 | 飼主編集 | /owners/:id | OwnerFormPage→OwnerForm | ✅ | 同上 |
| 受付/飼主/予約 | 集計ダッシュボード | /aggregation | AggregationDashboardPage | ✅ | |
| 受付/飼主/予約 | 予約管理 | /reservations | ReservationManagement | ✅ | |
| 運用/Lステップ | Lステップ健診連携 | /lstep/checkup-sync | CheckupSyncPage | ✅ | |
| 運用/Lステップ | Lステップ配信モニター | /lstep/delivery-monitor | LstepDeliveryMonitorPage | ✅ | |
| 運用/Lステップ | Lステップ分析 | /lstep/analytics | LstepAnalyticsPage | ✅ | |
| 運用/LINE予約 | LINE予約設定(index) | /line-reservation | LineReservationSettings | ✅ | |
| 運用/LINE予約 | LINE予約設定 | /line-reservation/settings | LineReservationSettings | ✅ | |
| 運用/LINE予約 | LINE予約ページエディタ | /line-reservation/page-editor | LineReservationPageEditor | ✅ | |
| 運用/LINE予約 | LINE予約枠設定 | /line-reservation/slots | LineReservationSlotsSettings | ✅ | |
| 運用/医院 | 医院マスタ設定 | /settings/clinic | ClinicMasterSettings→ClinicMasterList | ✅ | 薄いwrapper、shell実体はClinicMasterList |
| 運用/マニュアル | マニュアル(トップ) | /manual | ManualPage | ✅ | custom doc shell（canvas-soft、2026-07-06 C.bgPage化）。二ペインdocビューア+印刷+固定FABのためPageLayout非採用 |
| 運用/マニュアル | マニュアル記事 | /manual/:category/:slug | ManualPage | ✅ | 同上 |
| 運用/横断 | 同一飼主・ペット連携 | /identity-links | IdentityLinksPage | ✅ | PageLayout化（canView で Navigate ゲート。所属医院内の手動リンクのみ） |
| 設定/マスタ | 設定トップ | /settings | MasterSettingsIndex | ✅ | |
| 設定/マスタ | 職員マスタ | /settings/staff | StaffSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 診療項目マスタ | /settings/treatment-items | TreatmentPlanMaster | ✅ | MasterTabPage経由 |
| 設定/マスタ | 診断名マスタ | /settings/diagnosis | DiagnosisSettings | ✅ | MasterTabPage経由 |
| 設定/マスタ | 動物種マスタ | /settings/animal-species | AnimalSpeciesSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | トリミングマスタ | /settings/trimming | TrimmingSettings | ✅ | MasterTabPage経由 |
| 設定/マスタ | トリミングコース種別マスタ | /settings/trimming-course-type | TrimmingCourseTypeSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 薬剤マスタ | /settings/medicine | MedicineSettings | ✅ | |
| 設定/マスタ | 予約種別マスタ | /settings/reservation-type | ReservationTypeSettings | ✅ | |
| 設定/マスタ | 入院・ホテルマスタ | /settings/hospitalization | HospitalizationSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | ケージマスタ | /settings/cage | CageSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 物販品マスタ | /settings/merchandise-items | MerchandiseItemSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 保険マスタ | /settings/insurance | InsuranceSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 職種マスタ | /settings/occupations | OccupationSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 権限グループマスタ | /settings/permission-groups | PermissionGroupSettings | ✅ | C6(RBAC表示)は目視要 |
| 設定/マスタ | 問診テンプレートマスタ | /settings/inquiry-templates | InterviewTemplateSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 主訴マスタ | /settings/interview/chief-complaint | ChiefComplaintSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 問診テンプレート(interview) | /settings/interview/templates | InterviewTemplateSettings | ✅ | 同一Component別path |
| 設定/マスタ | シフトテンプレートマスタ | /settings/shift-templates | ShiftTemplateSettings | ✅ | |
| 設定/マスタ | 締め時間設定 | /settings/closing-time | ClosingSettingsPage | ✅ | |
| 設定/マスタ | 支払方法マスタ | /settings/payment-methods | PaymentMethodSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 割引キャンペーンマスタ | /settings/campaigns | CampaignSettings | ✅ | MasterCRUDPage経由 |
| 設定/マスタ | 検査機器マスタ | /settings/lab-device-item-masters | LabDeviceItemMasterSettings | ✅ | 日常送信経路に出さない |
| 臨床 | 検査受信 | /lab-device | LabDeviceBoard | ✅ | 1画面掲示板。本日診療中カルテカード＋日別受信一覧。確認ダイアログなし |
| 設定/マスタ | Lステップ連携設定 | /settings/integrations/lstep | LstepSettingsPage | ✅ | |
| 設定/マスタ | Lステップタグ管理 | /settings/lstep/tags | LstepTagManagementPage | ✅ | |

**脚注**:
- 最終runtime: `cd frontend && ./scripts/run-e2e.sh e2e/ui-design-compliance-readonly.spec.ts --workers=1` は 2026-07-23 時点で **92/92 PASS**（route inventory 1 + 製品ページ83 + known RBAC 3 + clinical P10 5）。§2 在庫更新後の期待内訳は route inventory 1 + 製品ページ84 + known RBAC 3 + clinical P10 5 = **93**（`/identity-links` 追加分）。再実行結果は e2e inventory 同期後に更新する。全ページで4 viewport、console/network/overflow/accessible name/44pxを検査し、synthetic interceptorは attempted / locally fulfilled / blocked / continued-to-backend を排他的に集計する。
- write safety: synthetic non-GETは完全一致allowlistの `POST /api/v1/reservations` と `POST /api/v1/medical-records` だけをlocal fulfillし、それ以外のbusiness non-GETはabortする。same-origin、sentinel `X-Clinic-ID`、non-GETの`X-Requested-With`も検証し、`frontend/e2e/helpers/synthetic-api.spec.ts` 2件と `frontend/e2e/medical-records-create.spec.ts` 1件は **3/3 PASS**、continued-to-backend 0。
- targeted coverage: 明示したproduct logic 18 source / 17 test file / 138 testで statements **84.58%**、lines **87.38%**、branches **75.89%**、functions **83.10%**。`.coverage-baseline` は変更していない。
- リダイレクト専用route（`<Navigate replace />`）は表から除外: `/settings/job-title`, `/service-type`, `/diagnosis-type`, `/diagnosis-name`, `/trimming-course`, `/trimming-option`, `/examination`, `/vaccine`, `/consultation`, `/procedure`, `/inquiry-template`, `/shift-template` の **12件**。
- C1（legacy accent）・C3（route表面hex直書き）は全85ルート対象の機械検査で **0件**（`/identity-links` 追加後も C1/C3 逸脱なしを前提。再確認は `pnpm design-audit`）。生regex `#[0-9A-Fa-f]{3,8}` は issue番号コメント（`#158`等）を誤検知するため、文字列リテラル限定パターンで再検証済み。
- C5（Primary CTA colorVariant）の現行分布は `primary` 16件、認証の明示的 `brand` 3件。`default` の明示使用と未定義 variant は0件。
- C15（route named white/black）・C16（非仕様20px spacing）・C17（CSS shadow直書き）・C18（Table primitive / raw cell override）・C19（4種のrow全体click）は 2026-07-24 監査時点で **0件**（`docker compose exec frontend pnpm design-audit`）。`/identity-links` 追加と並行 shell 修正後も同条件を維持する前提で、再確認は design-audit に委ねる。C18 は件数ratchetを持たず、呼び出し側 override を strict failする。raw の除外は allowlist / print 帳票 / FE11 対象外に限定する。C16は画面用26件を仕様内scaleへ移行し、印刷ビュー5件は対象外を明示した。C19では共有row 29箇所に加えraw `<TableRow onClick>` 7箇所・raw `<tr onClick>` 3箇所をcell内native controlへ移行し、同じ失敗modeの再導入をunit testで禁止した。
- DESIGN_SYSTEM §10 の既知baseline「route表面accent 0件」は再現。2026-07-06時点で `AccountingReportsPage` / `CashRegisterClosePage` / `CashRegisterHistoryPage` を `PageLayout` 化し、非PageLayout例は解消済み（DESIGN_SYSTEM.md §10 側の旧baseline記述は本書のみ更新、DESIGN_SYSTEM.md本文は対象外のため未修正）。`ManualPage` は二ペインdocビューア構造のため引き続き `PageLayout` 非採用・独自 canvas-soft shell（`C.bgPage`）で C2 準拠。`IdentityLinksPage`（`/identity-links`）は PageLayout shell を採用する（C2/C4/C8）。
