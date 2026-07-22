# UI Design Compliance — 全ルートページ準拠監査

> 正本。[DESIGN.md](../../DESIGN.md)（タイポ・形状・余白・エレベーション・寸法）と [docs/spec/design-system.md](design-system.md)（AE 製品色）への準拠を、本体84リーフルートとその実装コンポーネントに対して判定した結果。最終監査日 **2026-07-22**。`frontend/docs/design-audit-pages.md` は本書に統合済みで更新終了。

## 1. 準拠チェック定義

| ID | 観点 | 参照 | 判定方法 |
|---|---|---|---|
| C1 | 構造色は製品採用の brand `#0075DE` 系のみ。legacy（旧 teal `#038B94`/`#027078`・旧 accent `#2383E2`・`C.accent`）禁止 | DESIGN_SYSTEM §2.1, §9, §10 | **機械化**（C1） |
| C2 | ページ canvas は暖色 canvas-soft（`PageLayout` / `STYLE.page*` / `C.bgPage`）。純白ページ禁止 | DESIGN_SYSTEM §2.2, §9 | **C8 機械化**（`src/features/*/routes/*.tsx` の PageLayout / Master*Page / allowlist）。§2 表は新規リーフ追加時に更新 |
| C3 | コンポーネントでの hex 直書き禁止。`design-tokens.ts` 経由必須 | DESIGN_SYSTEM §9 Don't, §10 | **機械化**（C3。issue番号コメントは対象外） |
| C4 | 一覧/詳細系ページは `PageLayout` または同等 shell を持つ | DESIGN_SYSTEM §4, §7.6 | **C8 機械化**（C2 と同一。shell 欠落 = C2/C4 同時フラグ） |
| C5 | Primary CTA は `PrimaryButton`/`SubmitButton` の `colorVariant="brand"` | DESIGN_SYSTEM §7.2 | `rg 'colorVariant="[a-zA-Z]+"'` の値分布確認 |
| C6a | 臨床安全 UI（危険/死亡/RBAC 非活性表示）はデザイン変更で退行させない | DESIGN_SYSTEM §2.4, §9 | 静的 grep では網羅不可 — **コードレビュー要**。danger/warning の hex 直書き逸脱は C3 と合わせて確認 |
| C6b | rgba/rgb/hsla/hsl 直値禁止 | design-system-audit.mjs C6 | **機械化**（`pnpm design-audit` / make ci） |
| C7 | PageLayout `maxWidth` 生値禁止（`max-w-full` / `max-w-[Npx]`） | LAYOUT.pageContentMaxWidth | **機械化**（C7） |
| C8 | routes/*.tsx は PageLayout / Master*Page / allowlist | PageLayout | **機械化**（C8・相対パス完全一致 allowlist 14件 = 独自page 9 + helper 5） |
| C9 | `rounded-[Npx]` 任意値禁止 | `--radius-xxs/xs/...` | **機械化**（C9） |
| C10 | Tailwind 既定影（`shadow-2xs〜2xl`）と `shadow-[...]` 任意値禁止。`shadow-level1/level2/btn/panel/focus-brand` 等の elevation トークンのみ使用可 | design-system.md §5.1 | **機械化**（`pnpm design-audit` / make ci） |
| C11 | `text-[Npx\|Nrem]` font-size 任意値禁止 | design-system.md §3.4 | **機械化**（C11） |
| C12 | DESIGN.md にない `text-lg/2xl/3xl/4xl+` 禁止。`text-heading-1/2/3`・`text-xl/base/sm/xs/2xs` へ写像 | design-system.md §3.4 | **機械化**（C12。本体84ルート対象） |
| C13 | ink 4段を迂回する黒アルファ text / placeholder 禁止 | design-system.md §2.1, §3.4 | **機械化**（C13） |
| C14 | ロール固有 letter-spacing を上書きする `tracking-*` built-in / 任意値 / shorthand 禁止 | design-system.md §3.1, §3.4 | **機械化**（C14。本体84ルート対象） |
| C15 | 本体 routes/pages の white/black named color 直接指定禁止。色token経由必須 | design-system.md §2, §9 | **機械化**（C15） |
| C16 | spacing scale にない20px utility（`*-5`、負値、`[20px]` / `[1.25rem]`）禁止 | design-system.md §4.1 | **機械化**（C16） |
| C17 | CSS の直接 `box-shadow:` / `filter: drop-shadow()` 禁止。elevation token経由必須 | design-system.md §5.1 | **機械化**（C17） |
| C18 | `TableHead` / `TableCell` 呼び出し側で header eyebrow・body-sm・12px vertical paddingを非仕様値へ上書きしない | DESIGN.md `ex-data-table-cell` | **機械化**（C18。test除外、`data-empty-state` + `colSpan` + `text-center` の空state追加余白と子要素は許可） |
| C19 | `DataTableRow` / `SortableDataTableRow` に行全体の `onClick` を付けない。遷移はcell内のnative link、表示・編集はnative button、並べ替えは44px drag handleを使う | DESIGN.md §7.4, §8.2 | **機械化**（C19。production `.tsx` の共有row opening tagをmultiline対応で検査） |

**FE11 正本分割（2026-07-21）**: 色は DESIGN_SYSTEM の製品判断、タイポ・形状・余白・エレベーション・寸法は DESIGN.md 字義を正本とする。brand `#0075DE` は製品採用値であり、臨床 semantic 色・業務 status 色・nav canvas-soft も DESIGN_SYSTEM の判断を維持する。

**監査範囲の限界**: `design-system-audit.mjs` は `src`・`liff/src`・`line-reserve/src` の非test TypeScriptを全数走査し、C17だけは CSS も走査する。FE11の対象は本体84ルートのため、C12/C14〜C17は LIFF・line-reserve・shared-liff を明示除外し、画面用規範と異なる `MedicalRecordPrintView` も除外する。C15 は本体 routes/pages に限定し、C18は `.tsx` の共通 Table primitive opening tag、C19は共有 DataTable row opening tagを対象にする。raw `<TableRow>` / `<tr>` の行クリック、ロール内での誤選択、手書き `<th>/<td>`、全viewport、C6a（臨床安全 UI）は静的判定できないためコードレビュー/ブラウザ確認を併用する。

**機械化（FE11 更新）**: C1/C3/C5/C6b/C7〜C19 は `frontend/scripts/design-system-audit.mjs` で strict fail。C2/C4 は C8 により `routes/*.tsx` を機械化済み。§2 表と C8 allowlist は新規リーフ追加時に同一コミットで更新する。C6a（臨床安全 UI）は引き続きコードレビュー要。
```bash
docker compose exec frontend pnpm design-audit
```

## 2. ページ別対応状況表

**サマリー: 機械適合 83 / P1〜P10 完全確認 0 / 対象外 1（合計 84 リーフルート）**（監査日 2026-07-22）。下表の `✅` は C1〜C19 と shell baseline の機械適合を示し、4 viewport・意味レビュー・臨床状態まで含む完全準拠ではない。完全確認は [`FE-refactor.md`](../../FE-refactor.md) で追跡する。

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
| カルテ | カルテ作成(新規) | /medical-records/new | MedicalRecordForm | ✅ | |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | ✅ | |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | ✅ | |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | ✅ | |
| 入院/ホテル | 入院・ホテル登録(新規) | /hospitalization/new | HospitalizationForm | ✅ | |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | ✅ | C6:死亡表示等は目視要 |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | ✅ | |
| トリミング | トリミング一覧 | /trimming | TrimmingList | ✅ | |
| トリミング | トリミング登録 - ペット選択 | /trimming/select-pet | TrimmingPetSelection | ✅ | |
| トリミング | トリミング登録(新規) | /trimming/new | TrimmingForm | ✅ | |
| トリミング | トリミング編集 | /trimming/:id | TrimmingForm | ✅ | |
| 検査 | 検査一覧 | /examinations | ExaminationsList | ✅ | |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | ✅ | |
| 検査 | 検査登録(新規) | /examinations/new | ExaminationForm | ✅ | |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | ✅ | |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | ✅ | |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | ✅ | |
| ワクチン | ワクチン登録(新規) | /vaccinations/new | VaccinationForm | ✅ | |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | ✅ | |
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
| 設定/マスタ | Lステップ連携設定 | /settings/integrations/lstep | LstepSettingsPage | ✅ | |
| 設定/マスタ | Lステップタグ管理 | /settings/lstep/tags | LstepTagManagementPage | ✅ | |

**脚注**:
- リダイレクト専用route（`<Navigate replace />`）は表から除外: `/settings/job-title`, `/service-type`, `/diagnosis-type`, `/diagnosis-name`, `/trimming-course`, `/trimming-option`, `/examination`, `/vaccine`, `/consultation`, `/procedure`, `/inquiry-template`, `/shift-template` の **12件**。
- C1（legacy accent）・C3（route表面hex直書き）は全84ルートで **0件**。生regex `#[0-9A-Fa-f]{3,8}` は issue番号コメント（`#158`等）を誤検知するため、文字列リテラル限定パターンで再検証済み。
- C5（Primary CTA colorVariant）は route表面で確認できた9件全てが `brand`。legacy variant使用は0件。
- C15（route named white/black）・C16（非仕様20px spacing）・C17（CSS shadow直書き）・C18（Table primitive非仕様override）・C19（共有DataTable rowの行全体click）は現行treeで **0件**。C16は画面用26件を仕様内scaleへ移行し、印刷ビュー5件は対象外を明示した。C18導入時に36ファイル195行、C19導入時に26ファイル29箇所を検出し、primitive標準とcell内native controlへ移行した。raw `<TableRow onClick>` 7箇所とraw `<tr onClick>` 3箇所はC19対象外で、次バッチとして追跡する。
- DESIGN_SYSTEM §10 の既知baseline「route表面accent 0件」は再現。2026-07-06時点で `AccountingReportsPage` / `CashRegisterClosePage` / `CashRegisterHistoryPage` を `PageLayout` 化し、非PageLayout例は解消済み（DESIGN_SYSTEM.md §10 側の旧baseline記述は本書のみ更新、DESIGN_SYSTEM.md本文は対象外のため未修正）。`ManualPage` は二ペインdocビューア構造のため引き続き `PageLayout` 非採用・独自 canvas-soft shell（`C.bgPage`）で C2 準拠。
