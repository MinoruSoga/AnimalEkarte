# UI Design Compliance — 全ルートページ準拠監査

> 正本。[DESIGN.md](../../DESIGN.md)（タイポ・形状・余白・エレベーション・寸法）と [docs/spec/design-system.md](design-system.md)（AE 製品色）への準拠を、本体 86 製品ルート（+ wildcard 404）とその実装コンポーネントに対して判定した結果。最終監査日 **2026-07-24**（§2 在庫は 2026-07-31 に `/identity-links` を追加し 86 製品ルートへ更新）。

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
| C12 | DESIGN.md にない `text-lg/2xl/3xl/4xl+` 禁止。`text-heading-1/2/3`・`text-xl/base/sm/xs/2xs` へ写像 | design-system.md §3.4 | **機械化**（C12。本体 86 製品ルート対象） |
| C13 | ink 4段を迂回する黒アルファ text / placeholder 禁止 | design-system.md §2.1, §3.4 | **機械化**（C13） |
| C14 | ロール固有 letter-spacing を上書きする `tracking-*` built-in / 任意値 / shorthand 禁止 | design-system.md §3.1, §3.4 | **機械化**（C14。本体 86 製品ルート対象） |
| C15 | 本体 routes/pages の white/black named color 直接指定禁止。色token経由必須 | design-system.md §2, §9 | **機械化**（C15） |
| C16 | spacing scale にない20px utility（`*-5`、負値、`[20px]` / `[1.25rem]`）禁止 | design-system.md §4.1 | **機械化**（C16） |
| C17 | CSS の直接 `box-shadow:` / `filter: drop-shadow()` 禁止。elevation token経由必須 | design-system.md §5.1 | **機械化**（C17） |
| C18 | `TableHead` / `TableCell` とraw `th` / `td` で header eyebrow・body-sm・12px vertical / 16px horizontal paddingを非仕様値へ上書きしない | DESIGN.md `ex-data-table-cell` | **機械化**（C18。test除外、空stateと`data-c18-structural-cell`付きself-closing構造セルは許可。raw は `C18_RAW_CELL_ALLOWLIST` / `*PrintArea*` / FE11除外のみ。件数ratchetなし・違反はstrict fail） |
| C19 | `DataTableRow` / `SortableDataTableRow` / `TableRow` / raw `tr` に行全体の `onClick` を付けない。遷移はcell内のnative link、表示・編集はnative button、並べ替えは44px drag handleを使う | DESIGN.md §7.4, §8.2 | **機械化**（C19。production `.tsx` の4種のrow opening tagをmultiline対応で検査） |

**FE11 正本分割（2026-07-21、色役割更新 2026-07-27）**: 色は DESIGN_SYSTEM の製品判断、タイポ・形状・余白・エレベーション・寸法は DESIGN.md 字義を正本とする。brand と primary は同じ teal 値を使い、製品識別と汎用操作・選択・focus の意味役割だけをトークン名で分ける。臨床 semantic 色・業務 status 色・nav canvas-soft も DESIGN_SYSTEM の判断を維持する。

**監査範囲の限界**: `design-system-audit.mjs` は `src`・`liff/src`・`line-reserve/src` の非test TypeScriptを全数走査し、C17だけは CSS も走査する。FE11 の静的在庫対象は本体 86 製品ルートのため、C12/C14〜C17は LIFF・line-reserve・shared-liff を明示除外し、画面用規範と異なる `MedicalRecordPrintView` も除外する。C15 は本体 routes/pages、C18は `.tsx` の共通 Table primitive（`TableHead` / `TableCell`）とraw `th` / `td` opening tag、C19は4種のrow opening tagを対象にする。C18 は呼び出し側による非仕様 typography / vertical padding（raw は横padding・セル背景も含む）の再上書きを検出し、検出1件でも strict failする（件数ratchetなし。raw の正当な除外は `C18_RAW_CELL_ALLOWLIST`・`*PrintArea*` basename・FE11除外のみ）。ロール内での意味的誤選択、全viewport、C6a（臨床安全 UI）は静的判定できないためコードレビュー/ブラウザ確認を併用する。

**機械化（FE11 更新）**: C1/C3/C5/C6b/C7〜C19 は `frontend/scripts/design-system-audit.mjs` で strict fail（C18 も件数ratchetなし。違反0件が合格条件）。C2/C4 は C8 により `routes/*.tsx` を機械化済み。§2 表と C8 allowlist は新規リーフ追加時に同一コミットで更新する。C6a（臨床安全 UI）は引き続きコードレビュー要。
```bash
docker compose exec frontend pnpm design-audit
```

## 2. ページ別対応状況表

**静的在庫: 86 製品ルート + wildcard 404 = 87 行。** この docs refresh では source / route inventory を静的に照合したが、runtime E2E は再実行していない。表の静的列は現行 source との対応だけを示し、runtime 合格を意味しない。

**既知の E2E inventory gap**: `frontend/e2e/ui-design-compliance-readonly.spec.ts` は 85 製品 route しか持たず、実在する `/examinations/new` を欠く。`/identity-links` も追加後の route 単位 runtime 証跡が記録されていない。両 route の runtime は **pending**。E2E source と期待件数の修正・scoped rerun は別 task とし、本 docs-only 変更では行わない。2026-07-23 の 83 製品ページ aggregate run は履歴情報としてのみ残し、現在の 86 route へ拡張解釈しない。

| エリア | ページ名 | パス | コンポーネント | 静的監査 (2026-08-31) | runtime 監査 | 備考 |
|---|---|---|---|---|---|---|
| 認証/共通 | ログイン | /login | Login | ✅ | 未記録（現行在庫） |  |
| 認証/共通 | パスワードを忘れた方 | /forgot-password | ForgotPasswordPage | ✅ | 未記録（現行在庫） |  |
| 認証/共通 | パスワード再設定 | /reset-password | ResetPasswordPage | ✅ | 未記録（現行在庫） |  |
| 認証/共通 | 飼主カルテレポート(単独ウィンドウ) | /owners/:id/report | OwnerReport | ✅ | 未記録（現行在庫） |  |
| 認証/共通 | 404 Not Found | * | (inline element) | — | 対象外 | lazy Component非使用・簡易fallback |
| 会計 | 会計一覧 | /accounting | AccountingList | ✅ | 未記録（現行在庫） |  |
| 会計 | 会計 - ペット選択 | /accounting/select-pet | AccountingPetSelection | ✅ | 未記録（現行在庫） |  |
| 会計 | 会計登録(新規) | /accounting/new | AccountingDetailPage→AccountingDetail | ✅ | 未記録（現行在庫） | 薄いwrapper、shell実体はAccountingDetail |
| 会計 | 会計詳細 | /accounting/:id | AccountingDetailPage→AccountingDetail | ✅ | 未記録（現行在庫） | 同上 |
| 会計 | レジ締め | /accounting/close | CashRegisterClosePage | ✅ | 未記録（現行在庫） | PageLayout化済（2026-07-06） |
| 会計 | レジ締め履歴 | /accounting/close/history | CashRegisterHistoryPage | ✅ | 未記録（現行在庫） | PageLayout化済（2026-07-06） |
| 会計 | 月次集計レポート | /accounting/reports | AccountingReportsPage | ✅ | 未記録（現行在庫） | PageLayout化済（2026-07-06、既知baseline解消） |
| 在庫/見積/シフト | 在庫一覧 | /inventory | InventoryList | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 在庫登録(新規) | /inventory/new | InventoryForm | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 在庫編集 | /inventory/:id | InventoryForm | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 見積一覧 | /estimates | EstimateList | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 見積作成(新規) | /estimates/new | EstimateForm | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 見積詳細 | /estimates/:id | EstimateDetail | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | 見積編集 | /estimates/:id/edit | EstimateForm | ✅ | 未記録（現行在庫） |  |
| 在庫/見積/シフト | シフトカレンダー | /shifts | ShiftCalendarPage | ✅ | 未記録（現行在庫） |  |
| カルテ | カルテ一覧 | /medical-records | MedicalRecords | ✅ | 未記録（現行在庫） |  |
| カルテ | カルテ作成 - ペット選択 | /medical-records/select-pet | MedicalRecordPetSelection | ✅ | 未記録（現行在庫） |  |
| カルテ | カルテ作成(新規) | /medical-records/new | MedicalRecordForm | ✅ | 未記録（現行在庫） | synthetic pet GET、予約/カルテ作成POSTをlocal fulfillし、synthetic detailへproduction遷移。backend write 0 |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | ✅ | 未記録（現行在庫） |  |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | ✅ | 未記録（現行在庫） |  |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | ✅ | 未記録（現行在庫） |  |
| 入院/ホテル | 入院・ホテル登録(新規) | /hospitalization/new | HospitalizationForm | ✅ | 未記録（現行在庫） |  |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | ✅ | 未記録（現行在庫） | typed `BackendHospitalization` / care-plan fixtureで臨床状態・担当医・単価を4 viewport確認 |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | ✅ | 未記録（現行在庫） | typed recordのform/treatment-plan hydration、治療・割引の参照専用状態、費用、mount時business non-GET 0を確認 |
| トリミング | トリミング一覧 | /trimming | TrimmingList | ✅ | 未記録（現行在庫） |  |
| トリミング | トリミング登録 - ペット選択 | /trimming/select-pet | TrimmingPetSelection | ✅ | 未記録（現行在庫） |  |
| トリミング | トリミング登録(新規) | /trimming/new | TrimmingForm | ✅ | 未記録（現行在庫） |  |
| トリミング | トリミング編集 | /trimming/:id | TrimmingForm | ✅ | 未記録（現行在庫） | typed `BackendTrimming` とcourse/options/staff master fixtureでedit stateを4 viewport確認 |
| 検査 | 検査一覧 | /examinations | ExaminationsList | ✅ | 未記録（現行在庫） |  |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | ✅ | 未記録（現行在庫） |  |
| 検査 | 検査登録(新規) | /examinations/new | ExaminationForm | ✅ | pending（未実行） |  |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | ✅ | 未記録（現行在庫） |  |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | ✅ | 未記録（現行在庫） |  |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | ✅ | 未記録（現行在庫） |  |
| ワクチン | ワクチン登録(新規) | /vaccinations/new | Navigate → /vaccinations/select-pet | ✅ | 未記録（現行在庫） | /vaccinations/new は select-pet へリダイレクト。新規作成はカルテ予防接種タブから |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | ✅ | 未記録（現行在庫） | typed `BackendVaccination` とpet/vaccine/doctor fixtureでedit state・担当医帰属を4 viewport確認 |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | ✅ | 未記録（現行在庫） |  |
| 定期健診 | 定期健診登録 - ペット選択 | /checkups/select-pet | CheckupPetSelection | ✅ | 未記録（現行在庫） |  |
| 定期健診 | 定期健診登録(新規) | /checkups/new | CheckupForm | ✅ | 未記録（現行在庫） |  |
| 受付/飼主/予約 | 受付(当日) | / | Reception | ✅ | 未記録（現行在庫） |  |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersListPage→OwnersList | ✅ | 未記録（現行在庫） | 薄いwrapper、shell実体はOwnersList |
| 受付/飼主/予約 | 飼主登録(新規) | /owners/new | OwnerFormPage→OwnerForm | ✅ | 未記録（現行在庫） | 薄いwrapper、shell実体はOwnerForm |
| 受付/飼主/予約 | 飼主編集 | /owners/:id | OwnerFormPage→OwnerForm | ✅ | 未記録（現行在庫） | 同上 |
| 受付/飼主/予約 | 集計ダッシュボード | /aggregation | AggregationDashboardPage | ✅ | 未記録（現行在庫） |  |
| 受付/飼主/予約 | 予約管理 | /reservations | ReservationManagement | ✅ | 未記録（現行在庫） |  |
| 運用/Lステップ | Lステップ健診連携 | /lstep/checkup-sync | CheckupSyncPage | ✅ | 未記録（現行在庫） |  |
| 運用/Lステップ | Lステップ配信モニター | /lstep/delivery-monitor | LstepDeliveryMonitorPage | ✅ | 未記録（現行在庫） |  |
| 運用/Lステップ | Lステップ分析 | /lstep/analytics | LstepAnalyticsPage | ✅ | 未記録（現行在庫） |  |
| 運用/LINE予約 | LINE予約設定(index) | /line-reservation | LineReservationSettings | ✅ | 未記録（現行在庫） |  |
| 運用/LINE予約 | LINE予約設定 | /line-reservation/settings | LineReservationSettings | ✅ | 未記録（現行在庫） |  |
| 運用/LINE予約 | LINE予約ページエディタ | /line-reservation/page-editor | LineReservationPageEditor | ✅ | 未記録（現行在庫） |  |
| 運用/LINE予約 | LINE予約枠設定 | /line-reservation/slots | LineReservationSlotsSettings | ✅ | 未記録（現行在庫） |  |
| 運用/医院 | 医院マスタ設定 | /settings/clinic | ClinicMasterSettings→ClinicMasterList | ✅ | 未記録（現行在庫） | 薄いwrapper、shell実体はClinicMasterList |
| 運用/マニュアル | マニュアル(トップ) | /manual | ManualPage | ✅ | 未記録（現行在庫） | custom doc shell（canvas-soft、2026-07-06 C.bgPage化）。二ペインdocビューア+印刷+固定FABのためPageLayout非採用 |
| 運用/マニュアル | マニュアル記事 | /manual/:category/:slug | ManualPage | ✅ | 未記録（現行在庫） | 同上 |
| 運用/横断 | 同一飼主・ペット連携 | /identity-links | IdentityLinksPage | ✅ | pending（未実行） | PageLayout化（canView で Navigate ゲート。所属医院内の手動リンクのみ） |
| 設定/マスタ | 設定トップ | /settings | MasterSettingsIndex | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | 職員マスタ | /settings/staff | StaffSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 診療項目マスタ | /settings/treatment-items | TreatmentPlanMaster | ✅ | 未記録（現行在庫） | MasterTabPage経由 |
| 設定/マスタ | 診断名マスタ | /settings/diagnosis | DiagnosisSettings | ✅ | 未記録（現行在庫） | MasterTabPage経由 |
| 設定/マスタ | 動物種マスタ | /settings/animal-species | AnimalSpeciesSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | トリミングマスタ | /settings/trimming | TrimmingSettings | ✅ | 未記録（現行在庫） | MasterTabPage経由 |
| 設定/マスタ | トリミングコース種別マスタ | /settings/trimming-course-type | TrimmingCourseTypeSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 薬剤マスタ | /settings/medicine | MedicineSettings | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | 予約種別マスタ | /settings/reservation-type | ReservationTypeSettings | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | 入院・ホテルマスタ | /settings/hospitalization | HospitalizationSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | ケージマスタ | /settings/cage | CageSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 物販品マスタ | /settings/merchandise-items | MerchandiseItemSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 保険マスタ | /settings/insurance | InsuranceSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 職種マスタ | /settings/occupations | OccupationSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 権限グループマスタ | /settings/permission-groups | PermissionGroupSettings | ✅ | 未記録（現行在庫） | C6(RBAC表示)は目視要 |
| 設定/マスタ | 問診テンプレートマスタ | /settings/inquiry-templates | InterviewTemplateSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 主訴マスタ | /settings/interview/chief-complaint | ChiefComplaintSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 問診テンプレート(interview) | /settings/interview/templates | InterviewTemplateSettings | ✅ | 未記録（現行在庫） | 同一Component別path |
| 設定/マスタ | シフトテンプレートマスタ | /settings/shift-templates | ShiftTemplateSettings | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | 締め時間設定 | /settings/closing-time | ClosingSettingsPage | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | 支払方法マスタ | /settings/payment-methods | PaymentMethodSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 割引キャンペーンマスタ | /settings/campaigns | CampaignSettings | ✅ | 未記録（現行在庫） | MasterCRUDPage経由 |
| 設定/マスタ | 検査機器マスタ | /settings/lab-device-item-masters | LabDeviceItemMasterSettings | ✅ | 未記録（現行在庫） | 日常送信経路に出さない |
| 臨床 | 検査受信 | /lab-device | LabDeviceBoard | ✅ | 未記録（現行在庫） | 1画面掲示板。本日診療中カルテカード＋日別受信一覧。確認ダイアログなし |
| 設定/マスタ | Lステップ連携設定 | /settings/integrations/lstep | LstepSettingsPage | ✅ | 未記録（現行在庫） |  |
| 設定/マスタ | Lステップタグ管理 | /settings/lstep/tags | LstepTagManagementPage | ✅ | 未記録（現行在庫） |  |

**脚注**:
- runtime 履歴: 2026-07-23 の aggregate は 83 製品ページを対象とした。現行 86 route の route-level runtime 証跡ではない。現行 E2E inventory は 85 製品 route で `/examinations/new` が欠落し、scoped rerun も未実施。旧 92/92 や「expected pass」件数は現在の完了結果として使わない。
- write safety: synthetic non-GETは完全一致allowlistの `POST /api/v1/reservations` と `POST /api/v1/medical-records` だけをlocal fulfillし、それ以外のbusiness non-GETはabortする。same-origin、sentinel `X-Clinic-ID`、non-GETの`X-Requested-With`も検証し、`frontend/e2e/helpers/synthetic-api.spec.ts` 2件と `frontend/e2e/medical-records-create.spec.ts` 1件は **3/3 PASS**、continued-to-backend 0。
- targeted coverage: 明示したproduct logic 18 source / 17 test file / 138 testで statements **84.58%**、lines **87.38%**、branches **75.89%**、functions **83.10%**。`.coverage-baseline` は変更していない。
- リダイレクト専用route（`<Navigate replace />`）は表から除外: `/settings/job-title`, `/settings/service-type`, `/settings/diagnosis-type`, `/settings/diagnosis-name`, `/settings/trimming-course`, `/settings/trimming-option`, `/settings/examination`, `/settings/vaccine`, `/settings/consultation`, `/settings/procedure`, `/settings/inquiry-template`, `/settings/shift-template` の **12件**。
- C1（legacy accent）・C3（route表面hex直書き）の現行静的結果は design-audit で再確認する。本 docs refresh は runtime / design-audit を再実行しておらず、86 route 全体の新しい合格値を記録しない。
- C5（Primary CTA colorVariant）の現行分布は `primary` 16件、認証の明示的 `brand` 3件。`default` の明示使用と未定義 variant は0件。
- C15〜C19 の 2026-07-24 の監査値は履歴であり、現行 86 route の再実行結果ではない。新規 route を含む静的 gate は `docker compose exec frontend pnpm design-audit` の scoped operational execution で別途確認する。
- DESIGN_SYSTEM §10 の既知baseline「route表面accent 0件」は再現。2026-07-06時点で `AccountingReportsPage` / `CashRegisterClosePage` / `CashRegisterHistoryPage` を `PageLayout` 化し、非PageLayout例は解消済み（DESIGN_SYSTEM.md §10 側の旧baseline記述は本書のみ更新、DESIGN_SYSTEM.md本文は対象外のため未修正）。`ManualPage` は二ペインdocビューア構造のため引き続き `PageLayout` 非採用・独自 canvas-soft shell（`C.bgPage`）で C2 準拠。`IdentityLinksPage`（`/identity-links`）は PageLayout shell を採用する（C2/C4/C8）。
