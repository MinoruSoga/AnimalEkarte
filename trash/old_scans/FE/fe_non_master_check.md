# FE 業務系（非マスタ）コード規約違反 完全スキャン

## 目的
FEの業務系 feature（owners/pets/medical-records/hospitalization 等）が
「API層の命名規則・型安全性・エラーハンドリング・React 19 Action パターン」の規約に
準拠しているかを体系的に検査するためのチェックリストだ。

下記チェックリスト × 対象ファイルリストの**全組み合わせ**を検査し、
PASS/FAIL を表で出力せよ。

**新パターンの発見・起票は禁止。チェックリストに定義された10パターンのみを報告する。**

**対象外（専用ドキュメントを使用）:**
- `features/master/` → `tmp/FE/fe_master_check.md`
- `features/closing-settings/` → `tmp/FE/fe_closing_settings_check.md`

---

## チェックリスト（固定・10パターン）

### ■ API 層 (FA: Frontend API)

#### FA4: フック命名規則（api/）
Query/Mutation フックが以下の命名規則に従っているか？
- 一覧取得: `useGetXxx` または `useGetAllXxx`（`useXxx`・動詞省略は違反）
- 単件取得: `useGetXxx`
- 作成: `useCreateXxx`
- 更新: `useUpdateXxx`
- 削除: `useDeleteXxx`
- 違反例: `useOwners()`（動詞省略）/ `useFetchMedicalRecord()`（`Fetch` 動詞）
- 正しい例: `useGetOwners()`, `useCreateMedicalRecord()`, `useDeleteVaccination()`
- 対象: 上記「API」リストのファイル全件（export されるフック名）

#### FA5: onError での handleApiError（api/）
`useMutation` の `onError` コールバックで必ず `handleApiError(error, "コンテキスト")` が呼ばれているか？
- 違反例: `onError: (error) => console.error(error)` / `onError` なし
- 正しい例: `onError: (error) => handleApiError(error, "カルテの作成")`
- 対象: 上記「API」リストのファイル全件（全 `useMutation` の `onError`）

#### FA6: staleTime/gcTime 設定（api/）
一覧・単件取得フック（`useGetXxx`）に適切な `staleTime` と `gcTime` が設定されているか？
- マスタ参照系（変更頻度低）: `staleTime: QUERY_STALE_TIMES.STATIC, gcTime: QUERY_GC_TIMES.LONG`
- 業務データ系（変更頻度高）: `staleTime: QUERY_STALE_TIMES.SHORT` 等
- 違反例: `staleTime` 省略（デフォルト 0ms のまま）/ ハードコード数値
- 正しい例: `staleTime: QUERY_STALE_TIMES.SHORT, gcTime: QUERY_GC_TIMES.NORMAL`
- 対象: 上記「API」リストのファイル全件（`useQuery` を持つフックのみ）

#### FA8: transformXxx() / transforms.ts の存在（api/）
API ファイルまたは `transforms.ts` に `transformXxx()` 関数が存在し、BackendModel（snake_case）→ フロントエンドドメイン型（camelCase）に変換しているか？
- 違反例: transform なしで Backend 型をそのまま返す / `data.owner_id` を直接使う
- 正しい例: `export function transformOwner(data: ModelOwner) { return { id: String(data.id), name: data.name, ... }; }`
- **注意**: `transforms.ts` に集約して api ファイルからインポートする形式も OK。`types.ts` でドメイン型を手書きしている場合は違反。
- 対象: 上記「API」リストのファイル全件（`transforms.ts` が存在する feature はそちらも確認）

#### FA9: ドメイン型の導出方法（api/）
ドメイン型が `ReturnType<typeof transformXxx>` で型推論されているか？（手書き interface 禁止）
- 違反例: `export interface Owner { id: string; name: string; ... }`（手書き）
- 正しい例: `export type Owner = ReturnType<typeof transformOwner>;`
- **注意**: `types.ts` に手書きで interface を定義している場合は違反。`ReturnType` か zod schema からの型推論のみ許可。
- 対象: 上記「API」リストのファイル全件

---

### ■ 全体共通 (FG: Frontend General)

#### FG1: デザイントークン使用（routes/ + components/）
`C`, `STYLE`, `LAYOUT`, `ICON` 等のデザイントークン定数を使用し、Hex カラーや Tailwind のハードコードカラーを直接指定していないか？
- 違反例: `style={{ color: '#37352F' }}` / `className="text-gray-500 border-gray-200"`
- 正しい例: `style={{ color: C.TEXT_MAIN }}` / `className={cn(STYLE.FLEX_CENTER)}`
- **注意**: shadcn/ui コンポーネントの内部 className やサードパーティ由来の必須クラスは除外
- 対象: 上記「Routes」リストと「Components」リストの全件

#### FG2: 条件レンダーの三項演算子（routes/ + components/）
条件付きレンダリングが `condition ? (...) : null` を使っているか？（`&&` は禁止）
- 違反例: `{canEdit && <Button />}` / `{items.length && <List />}`
- 正しい例: `{canEdit ? <Button /> : null}`
- 対象: 上記「Routes」リストと「Components」リストの全件（全 JSX 中の条件レンダー）

#### FG3: any 型の不使用（api/ + routes/ + components/）
`any` 型を直接使用していないか？
- 違反例: `const data: any = response.data;` / `(e: any) => {}`
- 正しい例: `unknown` + 型ガード または 適切な型推論
- 対象: 上記「API」「Routes」「Components」リスト全件

#### FG4: useActionState + SubmitButton パターン（components/）
フォーム送信を伴うコンポーネントが `useActionState` + `<form action={...}>` + `SubmitButton` を使っているか？
（`useState` + `onSubmit` + 独自ローディング管理は禁止）
- 違反例: `const [isLoading, setIsLoading] = useState(false); const handleSubmit = async () => { setIsLoading(true); ... }`
- 正しい例: `const [, formAction, isPending] = useActionState(async (_prev, formData) => { ... }); return <form action={formAction}><SubmitButton>保存</SubmitButton></form>;`
- **注意**: 複雑なフォーム（multi-step、動的フィールド）で useActionState が困難な場合は `-`。判断基準: フォームの送信ボタンに isLoading/isPending を独自管理しているか否か。
- 対象: 上記「Components」リストのファイル全件（フォーム送信を行うコンポーネントのみ）

#### FG5: Feature Index 経由インポート（routes/ + components/）
`features/xxx` 内のコンポーネントを深いパスで直接インポートしていないか？ index.ts 経由のインポートを使っているか？
- 違反例: `import { OwnerForm } from '@/features/owners/components/OwnerForm'`（深いパス直接）
- 正しい例: `import { OwnerForm } from '@/features/owners'`（index.ts 経由）
- **注意**: 同一 feature 内の相対インポート（`./OwnerForm`）は問題なし。異なる feature や `@/` からの絶対インポートのみが対象。
- 対象: 上記「Routes」リストと「Components」リストの全件

---

## 対象ファイルリスト（全件）

### API（FA4, FA5, FA6, FA8, FA9, FG3 を検査）

#### owners
- frontend/src/features/owners/api/create-owner.ts
- frontend/src/features/owners/api/delete-owner.ts
- frontend/src/features/owners/api/get-owner.ts
- frontend/src/features/owners/api/get-insurances.ts
- frontend/src/features/owners/api/update-owner.ts
- frontend/src/features/owners/api/transforms.ts

#### pets
- frontend/src/features/pets/api/create-pet.ts
- frontend/src/features/pets/api/delete-pet.ts
- frontend/src/features/pets/api/get-pet.ts
- frontend/src/features/pets/api/get-pets.ts
- frontend/src/features/pets/api/update-pet.ts

#### medical-records
- frontend/src/features/medical-records/api/create-medical-record.ts
- frontend/src/features/medical-records/api/delete-medical-record.ts
- frontend/src/features/medical-records/api/get-chief-complaint-types.ts
- frontend/src/features/medical-records/api/get-diagnosis-options.ts
- frontend/src/features/medical-records/api/get-medical-record.ts
- frontend/src/features/medical-records/api/get-medical-record-images.ts
- frontend/src/features/medical-records/api/get-medical-records.ts
- frontend/src/features/medical-records/api/get-pet-vaccinations.ts
- frontend/src/features/medical-records/api/get-record-examinations.ts
- frontend/src/features/medical-records/api/update-medical-record.ts
- frontend/src/features/medical-records/api/billing-confirmation.ts
- frontend/src/features/medical-records/api/checkups.ts
- frontend/src/features/medical-records/api/clinical-plan.ts
- frontend/src/features/medical-records/api/inquiries.ts
- frontend/src/features/medical-records/api/medical-record-images.ts
- frontend/src/features/medical-records/api/save-estimate.ts
- frontend/src/features/medical-records/api/treatments.ts
- frontend/src/features/medical-records/api/vitals.ts
- frontend/src/features/medical-records/api/transforms.ts

#### hospitalization
- frontend/src/features/hospitalization/api/create-hospitalization.ts
- frontend/src/features/hospitalization/api/delete-hospitalization.ts
- frontend/src/features/hospitalization/api/get-hospitalization.ts
- frontend/src/features/hospitalization/api/get-hospitalizations.ts
- frontend/src/features/hospitalization/api/update-hospitalization.ts
- frontend/src/features/hospitalization/api/care-plan-items.ts
- frontend/src/features/hospitalization/api/daily-records.ts
- frontend/src/features/hospitalization/api/discharge-with-billing.ts
- frontend/src/features/hospitalization/api/transforms.ts

#### checkups
- frontend/src/features/checkups/api/get-checkups.ts
- frontend/src/features/checkups/api/transforms.ts

#### examinations
- frontend/src/features/examinations/api/create-examination.ts
- frontend/src/features/examinations/api/delete-examination.ts
- frontend/src/features/examinations/api/get-examination.ts
- frontend/src/features/examinations/api/get-examinations.ts
- frontend/src/features/examinations/api/update-examination.ts
- frontend/src/features/examinations/api/transforms.ts

#### vaccinations
- frontend/src/features/vaccinations/api/create-vaccination.ts
- frontend/src/features/vaccinations/api/delete-vaccination.ts
- frontend/src/features/vaccinations/api/get-vaccination.ts
- frontend/src/features/vaccinations/api/get-vaccinations.ts
- frontend/src/features/vaccinations/api/update-vaccination.ts
- frontend/src/features/vaccinations/api/transforms.ts

#### estimates
- frontend/src/features/estimates/api/create-estimate.ts
- frontend/src/features/estimates/api/delete-estimate.ts
- frontend/src/features/estimates/api/get-estimate.ts
- frontend/src/features/estimates/api/get-estimates.ts
- frontend/src/features/estimates/api/update-estimate.ts
- frontend/src/features/estimates/api/transforms.ts

#### trimming
- frontend/src/features/trimming/api/create-trimming.ts
- frontend/src/features/trimming/api/delete-trimming.ts
- frontend/src/features/trimming/api/get-trimming.ts
- frontend/src/features/trimming/api/get-trimmings.ts
- frontend/src/features/trimming/api/update-trimming.ts
- frontend/src/features/trimming/api/transforms.ts

#### reservations
- frontend/src/features/reservations/api/create-reservation.ts
- frontend/src/features/reservations/api/delete-reservation.ts
- frontend/src/features/reservations/api/get-on-duty-staffs.ts
- frontend/src/features/reservations/api/get-reservation-types.ts
- frontend/src/features/reservations/api/get-reservations.ts
- frontend/src/features/reservations/api/update-reservation.ts
- frontend/src/features/reservations/api/transforms.ts

#### accounting
- frontend/src/features/accounting/api/cancel-accounting.ts
- frontend/src/features/accounting/api/create-accounting.ts
- frontend/src/features/accounting/api/create-billing-item.ts
- frontend/src/features/accounting/api/create-refund.ts
- frontend/src/features/accounting/api/get-accounting.ts
- frontend/src/features/accounting/api/get-accountings.ts
- frontend/src/features/accounting/api/get-merchandise-items.ts
- frontend/src/features/accounting/api/get-refunds.ts
- frontend/src/features/accounting/api/get-unpaid-billings.ts
- frontend/src/features/accounting/api/update-accounting.ts
- frontend/src/features/accounting/api/update-billing-item.ts
- frontend/src/features/accounting/api/transforms.ts

#### shifts
- frontend/src/features/shifts/api/clinic-holidays.ts
- frontend/src/features/shifts/api/create-shift.ts
- frontend/src/features/shifts/api/create-shift-template.ts
- frontend/src/features/shifts/api/delete-shift.ts
- frontend/src/features/shifts/api/delete-shift-template.ts
- frontend/src/features/shifts/api/get-shifts.ts
- frontend/src/features/shifts/api/get-shift-templates.ts
- frontend/src/features/shifts/api/get-staffs.ts
- frontend/src/features/shifts/api/reorder-shift-templates.ts
- frontend/src/features/shifts/api/update-shift.ts
- frontend/src/features/shifts/api/update-shift-template.ts
- frontend/src/features/shifts/api/transforms.ts

#### inventory
- frontend/src/features/inventory/api/inventory.ts

#### reception
- frontend/src/features/reception/api/get-reception.ts
- frontend/src/features/reception/api/get-staffs.ts
- frontend/src/features/reception/api/update-appointment-status.ts
- frontend/src/features/reception/api/transforms.ts

#### line-reservation
- frontend/src/features/line-reservation/api/get-line-customers.ts
- frontend/src/features/line-reservation/api/get-line-reservation-setting.ts
- frontend/src/features/line-reservation/api/update-line-reservation-setting.ts
- frontend/src/features/line-reservation/api/update-owner-link.ts

#### accounting-reports
- frontend/src/features/accounting-reports/api/export-monthly-csv.ts
- frontend/src/features/accounting-reports/api/get-monthly-report.ts

#### cash-register
- frontend/src/features/cash-register/api/create-cash-register-close.ts
- frontend/src/features/cash-register/api/get-cash-register-close.ts
- frontend/src/features/cash-register/api/get-cash-register-closes.ts
- frontend/src/features/cash-register/api/get-cash-register-preview.ts

#### clinic-settings
- frontend/src/features/clinic-settings/api/clinics.ts
- frontend/src/features/clinic-settings/api/transforms.ts

### Routes（FG1, FG2, FG3, FG5 を検査）

#### checkups
- frontend/src/features/checkups/routes/CheckupsList.tsx

#### owners
- frontend/src/features/owners/routes/OwnersList.tsx
- frontend/src/features/owners/routes/OwnerForm.tsx

#### medical-records
- frontend/src/features/medical-records/routes/MedicalRecords.tsx
- frontend/src/features/medical-records/routes/MedicalRecordForm.tsx
- frontend/src/features/medical-records/routes/MedicalRecordPetSelection.tsx

#### hospitalization
- frontend/src/features/hospitalization/routes/HospitalizationList.tsx
- frontend/src/features/hospitalization/routes/HospitalizationForm.tsx
- frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx
- frontend/src/features/hospitalization/routes/HospitalizationPetSelection.tsx

#### examinations
- frontend/src/features/examinations/routes/ExaminationsList.tsx
- frontend/src/features/examinations/routes/ExaminationForm.tsx
- frontend/src/features/examinations/routes/ExaminationPetSelection.tsx

#### vaccinations
- frontend/src/features/vaccinations/routes/VaccinationList.tsx
- frontend/src/features/vaccinations/routes/VaccinationForm.tsx
- frontend/src/features/vaccinations/routes/VaccinationPetSelection.tsx

#### trimming
- frontend/src/features/trimming/routes/TrimmingList.tsx
- frontend/src/features/trimming/routes/TrimmingForm.tsx
- frontend/src/features/trimming/routes/TrimmingPetSelection.tsx

#### estimates
- frontend/src/features/estimates/routes/EstimateList.tsx
- frontend/src/features/estimates/routes/EstimateForm.tsx
- frontend/src/features/estimates/routes/EstimateDetail.tsx
- frontend/src/features/estimates/components/EstimateLineItems/EstimateLineItems.tsx
- frontend/src/features/estimates/components/EstimateStatusBadge/EstimateStatusBadge.tsx

#### accounting
- frontend/src/features/accounting/routes/AccountingList.tsx
- frontend/src/features/accounting/routes/AccountingDetail.tsx
- frontend/src/features/accounting/routes/AccountingPetSelection.tsx

#### reservations
- frontend/src/features/reservations/routes/ReservationManagement.tsx

#### reception
- frontend/src/features/reception/routes/Reception.tsx

#### shifts
- frontend/src/features/shifts/routes/ShiftCalendarPage.tsx
- frontend/src/features/shifts/routes/ShiftTemplateSettings.tsx

#### inventory
- frontend/src/features/inventory/routes/InventoryList.tsx
- frontend/src/features/inventory/routes/InventoryForm.tsx

#### clinic-settings
- frontend/src/features/clinic-settings/routes/ClinicMasterSettings.tsx

#### accounting-reports
- frontend/src/features/accounting-reports/routes/AccountingReportsPage.tsx

#### cash-register
- frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx
- frontend/src/features/cash-register/routes/CashRegisterHistoryPage.tsx

#### line-reservation
- frontend/src/features/line-reservation/routes/LineReservationSettings.tsx
- frontend/src/features/line-reservation/routes/LineReservationPageEditor.tsx

### Components（FG1, FG2, FG3, FG4, FG5 を検査）

#### owners
- frontend/src/features/owners/components/PetEditModal.tsx

#### medical-records
- frontend/src/features/medical-records/components/DiagnosisHeader.tsx
- frontend/src/features/medical-records/components/DiagnosisHeaderChiefComplaint.tsx
- frontend/src/features/medical-records/components/DiagnosisHeaderDiagnosis.tsx
- frontend/src/features/medical-records/components/DiagnosisHeaderPhysicalExam.tsx
- frontend/src/features/medical-records/components/EstimateForm.tsx
- frontend/src/features/medical-records/components/ExaminationFilter.tsx
- frontend/src/features/medical-records/components/ExaminationGroup.tsx
- frontend/src/features/medical-records/components/ExaminationImportDialog.tsx
- frontend/src/features/medical-records/components/ImageGalleryFilter.tsx
- frontend/src/features/medical-records/components/ImageGalleryGroup.tsx
- frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx
- frontend/src/features/medical-records/components/InterviewHistory.tsx
- frontend/src/features/medical-records/components/InterviewTreatmentPolicy.tsx
- frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx
- frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx
- frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx
- frontend/src/features/medical-records/components/MedicalRecordExamination.tsx
- frontend/src/features/medical-records/components/MedicalRecordImage.tsx
- frontend/src/features/medical-records/components/MedicalRecordInterview.tsx
- frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx
- frontend/src/features/medical-records/components/MedicalRecordTreatment.tsx
- frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx
- frontend/src/features/medical-records/components/StaffSelectionModal.tsx
- frontend/src/features/medical-records/components/TreatmentDetailedSummary.tsx
- frontend/src/features/medical-records/components/TreatmentTable.tsx
- frontend/src/features/medical-records/components/VaccinationForm.tsx
- frontend/src/features/medical-records/components/VaccinationHistory.tsx
- frontend/src/features/medical-records/components/VitalsModal.tsx
- frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx
- frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx
- frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx
- frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx
- frontend/src/features/medical-records/components/VitalsTab/VitalsGraph.tsx
- frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx

#### hospitalization
- frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx
- frontend/src/features/hospitalization/components/HospitalizationBasicInfo.tsx
- frontend/src/features/hospitalization/components/HospitalizationBoard.tsx
- frontend/src/features/hospitalization/components/HospitalizationCostSummary.tsx
- frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx
- frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx
- frontend/src/features/hospitalization/components/HospitalizationListView.tsx
- frontend/src/features/hospitalization/components/HospitalizationNoteCard.tsx
- frontend/src/features/hospitalization/components/HospitalizationPatientHeader.tsx
- frontend/src/features/hospitalization/components/HospitalizationTabbedView.tsx
- frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx
- frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx
- frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx
- frontend/src/features/hospitalization/components/DailyRecordsTab/DailyDateNav.tsx
- frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx
- frontend/src/features/hospitalization/components/DailyRecordsTab/DailyStaffNotesSection.tsx
- frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx

#### examinations
- frontend/src/features/examinations/components/ExaminationCard.tsx

#### vaccinations
- frontend/src/features/vaccinations/components/VaccinationCard.tsx

#### accounting
- frontend/src/features/accounting/components/AccountingDocument.tsx
- frontend/src/features/accounting/components/UnpaidTab.tsx

#### reservations
- frontend/src/features/reservations/components/MonthView.tsx
- frontend/src/features/reservations/components/ReservationDetailModal.tsx
- frontend/src/features/reservations/components/WeekView.tsx

#### shifts
- frontend/src/features/shifts/components/ClinicHolidayModal/ClinicHolidayModal.tsx
- frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx
- frontend/src/features/shifts/components/ShiftCell/ShiftCell.tsx
- frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx

#### reception
- frontend/src/features/reception/components/AppointmentCard.tsx
- frontend/src/features/reception/components/KanbanColumn.tsx
- frontend/src/features/reception/components/ReceptionDetailModal.tsx

#### line-reservation
- frontend/src/features/line-reservation/components/LinkedLineCustomers.tsx

#### accounting-reports
- frontend/src/features/accounting-reports/components/DailyBreakdownTable.tsx
- frontend/src/features/accounting-reports/components/MonthlySummaryCards.tsx

#### cash-register
- frontend/src/features/cash-register/components/BillingDetailTable.tsx
- frontend/src/features/cash-register/components/CashReconciliationCard.tsx
- frontend/src/features/cash-register/components/CategoryPaymentMatrix.tsx

---

## 実行方法（AgentTeam 推奨）

以下の5チームで並列実行せよ。各チームは担当ファイルのみを読む。

| チーム | 担当パターン | 担当ファイル |
|--------|------------|------------|
| Team-API-Core | FA4, FA5, FA6, FA8, FA9, FG3 | owners/pets/medical-records/hospitalization の「API」リスト |
| Team-API-Clinical | FA4, FA5, FA6, FA8, FA9, FG3 | checkups/examinations/vaccinations/trimming/estimates の「API」リスト |
| Team-API-Business | FA4, FA5, FA6, FA8, FA9, FG3 | accounting/reservations/shifts/inventory/reception/line-reservation/accounting-reports/cash-register/clinic-settings の「API」リスト |
| Team-Routes | FG1, FG2, FG3, FG5 | 上記「Routes」リスト全件 |
| Team-Components | FG1, FG2, FG3, FG4, FG5 | 上記「Components」リスト全件 |

---

## 出力フォーマット（必須）

| ファイル | FA4 | FA5 | FA6 | FA8 | FA9 | FG1 | FG2 | FG3 | FG4 | FG5 | 違反詳細 |
|---------|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|---------|
| create-owner.ts | OK | FAIL | - | - | - | - | - | OK | - | - | FA5:onError なし |
| OwnerForm.tsx | - | - | - | - | - | OK | FAIL | OK | OK | OK | FG2:行88 `{isEdit &&` で && 使用 |

凡例:
- `OK` = 問題なし
- `FAIL` = 違反あり（違反詳細列にファイル名:行番号と内容を必ず記載）
- `-` = 該当パターンなし（このファイルに対象メソッド/構造が存在しない）
- FG4 は Components のみ対象。Routes ファイルは `-` とする。
- FA8/FA9 は `transforms.ts` を持たない feature の個別 api ファイルでは `-` とし、`transforms.ts` でまとめて判定する。

---

## 禁止事項（遵守必須）

1. **新パターンの発見・起票禁止** — FA4/FA5/FA6/FA8/FA9, FG1〜FG5 以外の問題を見つけても記録しない
2. **推測判定禁止** — 必ずファイルを Read してから判定する。コードを読まずに OK/FAIL を出力しない
3. **曖昧出力禁止** — 「〜かもしれない」「要確認」は使わない。`OK` か `FAIL` かのみ
4. **ファイル追加禁止** — 上記リスト外のファイルをスキャンしない
5. **スキャン中の即時起票禁止** — 全ファイルスキャン完了後に PASS/FAIL 表と違反サマリを出力してから起票する
6. **スキップ禁止** — ファイルリストの全件を読むこと

---

## 完了条件

1. 上記全ファイル × 全パターンの PASS/FAIL 表が出力される
2. FAIL セルの一覧をまとめた「違反サマリ」を出力する
3. `docs/tasks/open/code-quality/` と `docs/tasks/closed/code-quality/` の既存タスクタイトルと照合し、**未起票の違反のみ**を新規タスクとして `docs/tasks/open/code-quality/` に起票する（タスク番号は既存の最大番号+1から採番）
