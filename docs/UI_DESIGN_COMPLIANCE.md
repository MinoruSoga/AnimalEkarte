# UI Design Compliance — 全ルートページ準拠監査

> 正本。[DESIGN.md](../DESIGN.md)（テンプレート）と [docs/DESIGN_SYSTEM.md](DESIGN_SYSTEM.md)（AE 製品上書き）への準拠を、`frontend/src/app/routes/` 配下の全リーフルートに対して機械的に判定した結果。監査日 **2026-07-06**。`frontend/docs/design-audit-pages.md` は本書に統合済みで更新終了。

## 1. 準拠チェック定義

| ID | 観点 | 参照 | 判定方法 |
|---|---|---|---|
| C1 | 構造色は brand teal `#038B94` 系のみ。legacy accent（`#0075DE`/`#2383E2`/`C.accent`）禁止 | DESIGN_SYSTEM §2.1, §9, §10 | `rg 'C\.accent\|#0075DE\|#2383E2'`（route/pages glob） |
| C2 | ページ canvas は暖色 canvas-soft（`PageLayout` / `STYLE.page*` / `C.bgPage`）。純白ページ禁止 | DESIGN.md canvas-soft／DESIGN_SYSTEM §2.2, §9 | 各リーフ Component の `PageLayout` 使用を追跡（直接 or 共有 shell 経由） |
| C3 | route 表面での hex 直書き禁止。`design-tokens.ts` 経由必須 | DESIGN_SYSTEM §9 Don't, §10 | `rg "['\"\`]#[0-9A-Fa-f]{3,8}['\"\`]"`（**issue番号コメント `#158` 等は文字列リテラルでないため除外**） |
| C4 | 一覧/詳細系ページは `PageLayout` または同等 shell を持つ | DESIGN_SYSTEM §4, §7.6 | C2 と同一判定（shell 欠落 = C2/C4 同時フラグ） |
| C5 | Primary CTA は `PrimaryButton`/`SubmitButton` の `colorVariant="brand"` | DESIGN_SYSTEM §7.2 | `rg 'colorVariant="[a-zA-Z]+"'` の値分布確認 |
| C6 | 臨床安全 UI（危険/死亡/RBAC 非活性表示）はデザイン変更で退行させない | DESIGN_SYSTEM §2.4, §9 | 静的 grep では網羅不可 — **コードレビュー要**。本監査では danger/warning の hex 直書き逸脱のみ C3 と合わせて確認（0 件） |

**AE 上書き注記**: DESIGN.md フロントマターの `{colors.primary}` は Notion Analysis テンプレート値 `#0075DE` のまま。製品正本は DESIGN_SYSTEM.md が定義する **`#038B94`**（SSOT 優先順位: 実装 `design-tokens.ts` > DESIGN_SYSTEM.md > DESIGN.md）。

**監査範囲の限界**: 本監査は「route が指す Component ファイル」と、そこから明確に import される共有 shell（`PageLayout` / `MasterPageShell` / `MasterListPage` / `MasterTabPage` 経由）までを追跡した。Component 内部が呼び出す個別 UI パーツ（カード・モーダル等）の hex は対象外。C6 は目視レビューが必要な領域が残る。

**判定方法（実行コマンド、2026-07-06 時点）**:
```bash
rg '#[0-9A-Fa-f]{3,8}' src/features --glob '**/routes/**' --glob '**/pages/**'   # 生hex regexは #158 等issue番号コメントを誤検知するため
rg "['\"\`]#[0-9A-Fa-f]{3,8}['\"\`]" src/features --glob '**/routes/**' --glob '**/pages/**' --glob '!**/*.test.*'  # 文字列リテラル限定＝実違反
rg 'C\.accent|#0075DE|#2383E2' src/features --glob '**/routes/**' --glob '**/pages/**'
rg --files-without-match 'PageLayout|STYLE\.page|C\.bgPage' <leaf-component-files>
rg -o 'colorVariant="[a-zA-Z]+"' src/features --glob '**/routes/**'
```

**C1/C3/C5 の機械化（2026-07-06 追加）**: 上記 grep のうち C1/C3/C5 は `frontend/scripts/design-system-audit.mjs` で自動判定できる（strict fail — 現状 0 件からの新規混入を検知）。C2/C4（`PageLayout` 追跡）は leaf component 解決が動的で自動化が複雑なため、引き続き §2 の手動更新で追跡する。
```bash
docker compose exec frontend pnpm design-audit
```

## 2. ページ別対応状況表

**サマリー: 準拠 83 / 一部 0 / 対象外 1（合計 84 リーフルート）**（監査日 2026-07-06、5件のPageLayout化により2026-07-06に解消）

| エリア | ページ名 | パス | コンポーネント | 準拠 | 備考 |
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
- DESIGN_SYSTEM §10 の既知baseline「route表面accent 0件」は再現。2026-07-06時点で `AccountingReportsPage` / `CashRegisterClosePage` / `CashRegisterHistoryPage` を `PageLayout` 化し、非PageLayout例は解消済み（DESIGN_SYSTEM.md §10 側の旧baseline記述は本書のみ更新、DESIGN_SYSTEM.md本文は対象外のため未修正）。`ManualPage` は二ペインdocビューア構造のため引き続き `PageLayout` 非採用・独自 canvas-soft shell（`C.bgPage`）で C2 準拠。
