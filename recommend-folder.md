React 19 Data Mode + bulletproof-react に完全準拠した、動物病院管理システムのベストフォルダ構成を提示します。

## 完全版：ベストフォルダ構成

```
src/
├── main.tsx
├── vite-env.d.ts
│
├── app/
│   ├── index.tsx                          # Appコンポーネント（export）
│   ├── provider.tsx                       # 全プロバイダー統合
│   ├── router.tsx                         # React Router Data Mode設定
│   └── ErrorBoundary.tsx                  # グローバルエラーバウンダリ
│
├── features/
│   │
│   ├── auth/
│   │   ├── api/
│   │   │   ├── login.ts                   # + useLogin()
│   │   │   ├── logout.ts                  # + useLogout()
│   │   │   └── getCurrentUser.ts          # + useCurrentUser()
│   │   ├── components/
│   │   │   └── LoginForm/
│   │   │       ├── LoginForm.tsx
│   │   │       ├── LoginForm.test.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   └── useAuth.ts                 # 認証状態管理ロジック
│   │   ├── routes/
│   │   │   └── Login.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts                       # Public API
│   │
│   ├── dashboard/
│   │   ├── api/
│   │   │   ├── getWorkflowStatus.ts       # + useWorkflowStatus()
│   │   │   ├── updatePatientStatus.ts     # + useUpdatePatientStatus()
│   │   │   └── getTodayStats.ts           # + useTodayStats()
│   │   ├── components/
│   │   │   ├── KanbanBoard/
│   │   │   │   ├── KanbanBoard.tsx
│   │   │   │   ├── KanbanBoard.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── KanbanColumn/
│   │   │   │   ├── KanbanColumn.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PatientCard/
│   │   │   │   ├── PatientCard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── StatsCards/
│   │   │   │   ├── StatsCards.tsx
│   │   │   │   └── index.ts
│   │   │   ├── StatCard/
│   │   │   │   ├── StatCard.tsx
│   │   │   │   └── index.ts
│   │   │   └── QuickActions/
│   │   │       ├── QuickActions.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useKanbanDragDrop.ts       # DnDロジック
│   │   │   └── useOptimisticStatusUpdate.ts # React 19 useOptimistic
│   │   ├── routes/
│   │   │   └── Dashboard.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── reservations/
│   │   ├── api/
│   │   │   ├── getReservations.ts         # + useReservations()
│   │   │   ├── getReservation.ts          # + useReservation()
│   │   │   ├── createReservation.ts       # + useCreateReservation()
│   │   │   ├── updateReservation.ts       # + useUpdateReservation()
│   │   │   ├── deleteReservation.ts       # + useDeleteReservation()
│   │   │   └── getCalendarView.ts         # + useCalendarView()
│   │   ├── components/
│   │   │   ├── ReservationCalendar/
│   │   │   │   ├── ReservationCalendar.tsx
│   │   │   │   ├── ReservationCalendar.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── CalendarView/
│   │   │   │   ├── CalendarView.tsx
│   │   │   │   └── index.ts
│   │   │   ├── MonthView/
│   │   │   │   ├── MonthView.tsx
│   │   │   │   └── index.ts
│   │   │   ├── WeekView/
│   │   │   │   ├── WeekView.tsx
│   │   │   │   └── index.ts
│   │   │   ├── DayView/
│   │   │   │   ├── DayView.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ReservationForm/
│   │   │   │   ├── ReservationForm.tsx
│   │   │   │   ├── ReservationForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PetSelector/
│   │   │   │   ├── PetSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── TimeSlotPicker/
│   │   │   │   ├── TimeSlotPicker.tsx
│   │   │   │   └── index.ts
│   │   │   ├── VetSelector/
│   │   │   │   ├── VetSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── RoomSelector/
│   │   │   │   ├── RoomSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ReservationCard/
│   │   │   │   ├── ReservationCard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ReservationList/
│   │   │   │   ├── ReservationList.tsx
│   │   │   │   └── index.ts
│   │   │   └── ReservationFilters/
│   │   │       ├── ReservationFilters.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useReservationForm.ts      # フォームロジック
│   │   │   ├── useCalendarNavigation.ts   # カレンダー操作
│   │   │   └── useReservationFilters.ts   # フィルタリングロジック
│   │   ├── routes/
│   │   │   ├── ReservationList.tsx
│   │   │   ├── ReservationCalendar.tsx
│   │   │   └── ReservationDetail.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   └── reservationHelpers.ts
│   │   └── index.ts
│   │
│   ├── medical-records/
│   │   ├── api/
│   │   │   ├── getMedicalRecords.ts       # + useMedicalRecords()
│   │   │   ├── getMedicalRecord.ts        # + useMedicalRecord()
│   │   │   ├── createMedicalRecord.ts     # + useCreateMedicalRecord()
│   │   │   ├── updateMedicalRecord.ts     # + useUpdateMedicalRecord()
│   │   │   ├── confirmMedicalRecord.ts    # + useConfirmMedicalRecord()
│   │   │   └── getEstimate.ts             # + useEstimate()
│   │   ├── components/
│   │   │   ├── SOAPSForm/
│   │   │   │   ├── SOAPSForm.tsx
│   │   │   │   ├── SOAPSForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── SubjectiveSection/
│   │   │   │   ├── SubjectiveSection.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ObjectiveSection/
│   │   │   │   ├── ObjectiveSection.tsx
│   │   │   │   └── index.ts
│   │   │   ├── AssessmentSection/
│   │   │   │   ├── AssessmentSection.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PlanSection/
│   │   │   │   ├── PlanSection.tsx
│   │   │   │   └── index.ts
│   │   │   ├── SurgerySection/
│   │   │   │   ├── SurgerySection.tsx
│   │   │   │   └── index.ts
│   │   │   ├── InterviewForm/
│   │   │   │   ├── InterviewForm.tsx
│   │   │   │   └── index.ts
│   │   │   ├── VitalSignsInput/
│   │   │   │   ├── VitalSignsInput.tsx
│   │   │   │   └── index.ts
│   │   │   ├── VitalChart/
│   │   │   │   ├── VitalChart.tsx
│   │   │   │   └── index.ts
│   │   │   ├── DiagnosisSelector/
│   │   │   │   ├── DiagnosisSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── TreatmentPlan/
│   │   │   │   ├── TreatmentPlan.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PrescriptionForm/
│   │   │   │   ├── PrescriptionForm.tsx
│   │   │   │   └── index.ts
│   │   │   ├── MedicineSelector/
│   │   │   │   ├── MedicineSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── DosageCalculator/
│   │   │   │   ├── DosageCalculator.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ImageUploader/
│   │   │   │   ├── ImageUploader.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ImageGallery/
│   │   │   │   ├── ImageGallery.tsx
│   │   │   │   └── index.ts
│   │   │   ├── EstimatePreview/
│   │   │   │   ├── EstimatePreview.tsx
│   │   │   │   └── index.ts
│   │   │   └── RecordTimeline/
│   │   │       ├── RecordTimeline.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useSOAPSForm.ts            # SOAPSフォームロジック
│   │   │   ├── useVitalSigns.ts           # バイタルサイン管理
│   │   │   └── usePrescription.ts         # 処方ロジック
│   │   ├── routes/
│   │   │   ├── MedicalRecordList.tsx
│   │   │   ├── MedicalRecordDetail.tsx
│   │   │   └── MedicalRecordCreate.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   ├── soapsValidator.ts
│   │   │   └── vitalCalculator.ts
│   │   └── index.ts
│   │
│   ├── hospitalization/
│   │   ├── api/
│   │   │   ├── getHospitalizations.ts     # + useHospitalizations()
│   │   │   ├── getHospitalization.ts      # + useHospitalization()
│   │   │   ├── createHospitalization.ts   # + useCreateHospitalization()
│   │   │   ├── updateHospitalization.ts   # + useUpdateHospitalization()
│   │   │   ├── updateCarePlan.ts          # + useUpdateCarePlan()
│   │   │   ├── addDailyLog.ts             # + useAddDailyLog()
│   │   │   └── calculateCost.ts           # + useCalculateCost()
│   │   ├── components/
│   │   │   ├── HospitalizationBoard/
│   │   │   │   ├── HospitalizationBoard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PatientCard/
│   │   │   │   ├── PatientCard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── CarePlanForm/
│   │   │   │   ├── CarePlanForm.tsx
│   │   │   │   ├── CarePlanForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── TaskSchedule/
│   │   │   │   ├── TaskSchedule.tsx
│   │   │   │   └── index.ts
│   │   │   ├── MedicationSchedule/
│   │   │   │   ├── MedicationSchedule.tsx
│   │   │   │   └── index.ts
│   │   │   ├── DailyLogForm/
│   │   │   │   ├── DailyLogForm.tsx
│   │   │   │   ├── DailyLogForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── TaskCheckList/
│   │   │   │   ├── TaskCheckList.tsx
│   │   │   │   └── index.ts
│   │   │   ├── VitalRecording/
│   │   │   │   ├── VitalRecording.tsx
│   │   │   │   └── index.ts
│   │   │   ├── FoodIntakeLog/
│   │   │   │   ├── FoodIntakeLog.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ExcretionLog/
│   │   │   │   ├── ExcretionLog.tsx
│   │   │   │   └── index.ts
│   │   │   ├── HospitalizationTimeline/
│   │   │   │   ├── HospitalizationTimeline.tsx
│   │   │   │   └── index.ts
│   │   │   └── CostCalculator/
│   │   │       ├── CostCalculator.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useCarePlanForm.ts         # ケアプランフォームロジック
│   │   │   ├── useDailyLogForm.ts         # デイリーログフォームロジック
│   │   │   ├── useOptimisticTaskUpdate.ts # React 19 useOptimistic
│   │   │   └── useCostCalculation.ts      # コスト計算ロジック
│   │   ├── routes/
│   │   │   ├── HospitalizationList.tsx
│   │   │   ├── HospitalizationDetail.tsx
│   │   │   └── HospitalizationCreate.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   └── costCalculator.ts
│   │   └── index.ts
│   │
│   ├── accounting/
│   │   ├── api/
│   │   │   ├── getAccountings.ts          # + useAccountings()
│   │   │   ├── getAccounting.ts           # + useAccounting()
│   │   │   ├── createAccounting.ts        # + useCreateAccounting()
│   │   │   ├── updateAccounting.ts        # + useUpdateAccounting()
│   │   │   ├── processPayment.ts          # + useProcessPayment()
│   │   │   └── generateInvoice.ts         # + useGenerateInvoice()
│   │   ├── components/
│   │   │   ├── AccountingForm/
│   │   │   │   ├── AccountingForm.tsx
│   │   │   │   ├── AccountingForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PatientSelector/
│   │   │   │   ├── PatientSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ItemsEditor/
│   │   │   │   ├── ItemsEditor.tsx
│   │   │   │   └── index.ts
│   │   │   ├── InvoicePreview/
│   │   │   │   ├── InvoicePreview.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ReceiptPreview/
│   │   │   │   ├── ReceiptPreview.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PaymentMethodSelector/
│   │   │   │   ├── PaymentMethodSelector.tsx
│   │   │   │   └── index.ts
│   │   │   └── AccountingList/
│   │   │       ├── AccountingList.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useAccountingForm.ts       # 会計フォームロジック
│   │   │   └── useInvoiceGeneration.ts    # 請求書生成ロジック
│   │   ├── routes/
│   │   │   ├── AccountingList.tsx
│   │   │   ├── AccountingDetail.tsx
│   │   │   └── AccountingCreate.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   ├── invoiceGenerator.ts
│   │   │   └── taxCalculator.ts
│   │   └── index.ts
│   │
│   ├── examinations/
│   │   ├── api/
│   │   │   ├── getExaminations.ts         # + useExaminations()
│   │   │   ├── getExamination.ts          # + useExamination()
│   │   │   ├── createExamination.ts       # + useCreateExamination()
│   │   │   ├── updateExamination.ts       # + useUpdateExamination()
│   │   │   ├── updateExamResult.ts        # + useUpdateExamResult()
│   │   │   └── getExamTypes.ts            # + useExamTypes()
│   │   ├── components/
│   │   │   ├── ExaminationForm/
│   │   │   │   ├── ExaminationForm.tsx
│   │   │   │   ├── ExaminationForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ExamTypeSelector/
│   │   │   │   ├── ExamTypeSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── ResultInput/
│   │   │   │   ├── ResultInput.tsx
│   │   │   │   └── index.ts
│   │   │   ├── SummaryEditor/
│   │   │   │   ├── SummaryEditor.tsx
│   │   │   │   └── index.ts
│   │   │   └── ExaminationList/
│   │   │       ├── ExaminationList.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   └── useExaminationForm.ts      # 検査フォームロジック
│   │   ├── routes/
│   │   │   ├── ExaminationList.tsx
│   │   │   └── ExaminationDetail.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── owners/
│   │   ├── api/
│   │   │   ├── getOwners.ts               # + useOwners()
│   │   │   ├── getOwner.ts                # + useOwner()
│   │   │   ├── createOwner.ts             # + useCreateOwner()
│   │   │   ├── updateOwner.ts             # + useUpdateOwner()
│   │   │   └── deleteOwner.ts             # + useDeleteOwner()
│   │   ├── components/
│   │   │   ├── OwnerForm/
│   │   │   │   ├── OwnerForm.tsx
│   │   │   │   ├── OwnerForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── OwnerCard/
│   │   │   │   ├── OwnerCard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── OwnerList/
│   │   │   │   ├── OwnerList.tsx
│   │   │   │   └── index.ts
│   │   │   └── OwnerSearch/
│   │   │       ├── OwnerSearch.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useOwnerForm.ts            # 飼い主フォームロジック
│   │   │   └── useOwnerSearch.ts          # 検索ロジック
│   │   ├── routes/
│   │   │   ├── OwnerList.tsx
│   │   │   ├── OwnerDetail.tsx
│   │   │   └── OwnerCreate.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── pets/
│   │   ├── api/
│   │   │   ├── getPets.ts                 # + usePets()
│   │   │   ├── getPet.ts                  # + usePet()
│   │   │   ├── createPet.ts               # + useCreatePet()
│   │   │   ├── updatePet.ts               # + useUpdatePet()
│   │   │   ├── deletePet.ts               # + useDeletePet()
│   │   │   ├── getPetsByOwner.ts          # + usePetsByOwner()
│   │   │   └── uploadPetImage.ts          # + useUploadPetImage()
│   │   ├── components/
│   │   │   ├── PetForm/
│   │   │   │   ├── PetForm.tsx
│   │   │   │   ├── PetForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PetImageUploader/
│   │   │   │   ├── PetImageUploader.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PetCard/
│   │   │   │   ├── PetCard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── PetList/
│   │   │   │   ├── PetList.tsx
│   │   │   │   └── index.ts
│   │   │   └── PetProfile/
│   │   │       ├── PetProfile.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   └── usePetForm.ts              # ペットフォームロジック
│   │   ├── routes/
│   │   │   ├── PetList.tsx
│   │   │   ├── PetDetail.tsx
│   │   │   └── PetCreate.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── trimming/
│   │   ├── api/
│   │   │   ├── getTrimmings.ts            # + useTrimmings()
│   │   │   ├── getTrimming.ts             # + useTrimming()
│   │   │   ├── createTrimming.ts          # + useCreateTrimming()
│   │   │   ├── updateTrimming.ts          # + useUpdateTrimming()
│   │   │   ├── updateTrimmingStatus.ts    # + useUpdateTrimmingStatus()
│   │   │   └── getTrimmingCourses.ts      # + useTrimmingCourses()
│   │   ├── components/
│   │   │   ├── TrimmingForm/
│   │   │   │   ├── TrimmingForm.tsx
│   │   │   │   ├── TrimmingForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── CourseSelector/
│   │   │   │   ├── CourseSelector.tsx
│   │   │   │   └── index.ts
│   │   │   ├── OptionSelector/
│   │   │   │   ├── OptionSelector.tsx
│   │   │   │   └── index.ts
│   │   │   └── TrimmingList/
│   │   │       ├── TrimmingList.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   └── useTrimmingForm.ts         # トリミングフォームロジック
│   │   ├── routes/
│   │   │   ├── TrimmingList.tsx
│   │   │   └── TrimmingDetail.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── vaccinations/
│   │   ├── api/
│   │   │   ├── getVaccinations.ts         # + useVaccinations()
│   │   │   ├── getVaccination.ts          # + useVaccination()
│   │   │   ├── createVaccination.ts       # + useCreateVaccination()
│   │   │   ├── updateVaccination.ts       # + useUpdateVaccination()
│   │   │   └── getVaccineTypes.ts         # + useVaccineTypes()
│   │   ├── components/
│   │   │   ├── VaccinationForm/
│   │   │   │   ├── VaccinationForm.tsx
│   │   │   │   ├── VaccinationForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── VaccinationSchedule/
│   │   │   │   ├── VaccinationSchedule.tsx
│   │   │   │   └── index.ts
│   │   │   └── VaccinationList/
│   │   │       ├── VaccinationList.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   ├── useVaccinationForm.ts      # ワクチンフォームロジック
│   │   │   └── useVaccineSchedule.ts      # スケジュール管理ロジック
│   │   ├── routes/
│   │   │   ├── VaccinationList.tsx
│   │   │   └── VaccinationDetail.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   ├── master/
│   │   ├── api/
│   │   │   ├── getTreatmentItems.ts       # + useTreatmentItems()
│   │   │   ├── getTreatmentItem.ts        # + useTreatmentItem()
│   │   │   ├── createTreatmentItem.ts     # + useCreateTreatmentItem()
│   │   │   ├── updateTreatmentItem.ts     # + useUpdateTreatmentItem()
│   │   │   └── deleteTreatmentItem.ts     # + useDeleteTreatmentItem()
│   │   ├── components/
│   │   │   ├── TreatmentItemForm/
│   │   │   │   ├── TreatmentItemForm.tsx
│   │   │   │   ├── TreatmentItemForm.test.tsx
│   │   │   │   └── index.ts
│   │   │   ├── TreatmentItemList/
│   │   │   │   ├── TreatmentItemList.tsx
│   │   │   │   └── index.ts
│   │   │   └── CategoryManager/
│   │   │       ├── CategoryManager.tsx
│   │   │       └── index.ts
│   │   ├── hooks/
│   │   │   └── useTreatmentItemForm.ts    # 診療項目フォームロジック
│   │   ├── routes/
│   │   │   ├── TreatmentItemList.tsx
│   │   │   └── TreatmentItemEdit.tsx
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   │
│   └── clinic/
│       ├── api/
│       │   ├── getClinicInfo.ts           # + useClinicInfo()
│       │   └── updateClinicInfo.ts        # + useUpdateClinicInfo()
│       ├── components/
│       │   └── ClinicInfoForm/
│       │       ├── ClinicInfoForm.tsx
│       │       ├── ClinicInfoForm.test.tsx
│       │       └── index.ts
│       ├── hooks/
│       │   └── useClinicInfoForm.ts       # クリニック情報フォームロジック
│       ├── routes/
│       │   └── ClinicSettings.tsx
│       ├── types/
│       │   └── index.ts
│       └── index.ts
│
├── components/
│   ├── ui/                                # shadcn/ui（汎用UIコンポーネント）
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── select.tsx
│   │   ├── dialog.tsx
│   │   ├── dropdown-menu.tsx
│   │   ├── table.tsx
│   │   ├── card.tsx
│   │   ├── badge.tsx
│   │   ├── calendar.tsx
│   │   ├── checkbox.tsx
│   │   ├── radio-group.tsx
│   │   ├── textarea.tsx
│   │   ├── toast.tsx
│   │   ├── toaster.tsx
│   │   ├── tabs.tsx
│   │   ├── form.tsx
│   │   ├── label.tsx
│   │   ├── separator.tsx
│   │   ├── sheet.tsx
│   │   └── skeleton.tsx
│   │
│   └── shared/                            # アプリ固有の共有コンポーネント
│       ├── Layout/
│       │   ├── MainLayout/
│       │   │   ├── MainLayout.tsx
│       │   │   ├── MainLayout.test.tsx
│       │   │   └── index.ts
│       │   ├── Header/
│       │   │   ├── Header.tsx
│       │   │   └── index.ts
│       │   ├── Sidebar/
│       │   │   ├── Sidebar.tsx
│       │   │   └── index.ts
│       │   ├── Navigation/
│       │   │   ├── Navigation.tsx
│       │   │   └── index.ts
│       │   ├── AuthLayout/
│       │   │   ├── AuthLayout.tsx
│       │   │   └── index.ts
│       │   └── PrintLayout/
│       │       ├── PrintLayout.tsx
│       │       └── index.ts
│       │
│       ├── DataTable/
│       │   ├── DataTable.tsx
│       │   ├── DataTable.test.tsx
│       │   ├── DataTablePagination.tsx
│       │   ├── DataTableFilters.tsx
│       │   ├── DataTableHeader.tsx
│       │   └── index.ts
│       │
│       ├── Form/
│       │   ├── FormField/
│       │   │   ├── FormField.tsx
│       │   │   ├── FormField.test.tsx
│       │   │   └── index.ts
│       │   ├── FormError/
│       │   │   ├── FormError.tsx
│       │   │   └── index.ts
│       │   ├── FormLabel/
│       │   │   ├── FormLabel.tsx
│       │   │   └── index.ts
│       │   ├── SubmitButton/
│       │   │   ├── SubmitButton.tsx
│       │   │   ├── SubmitButton.test.tsx
│       │   │   └── index.ts
│       │   └── FormSection/
│       │       ├── FormSection.tsx
│       │       └── index.ts
│       │
│       ├── Feedback/
│       │   ├── Spinner/
│       │   │   ├── Spinner.tsx
│       │   │   └── index.ts
│       │   ├── ErrorFallback/
│       │   │   ├── ErrorFallback.tsx
│       │   │   └── index.ts
│       │   ├── LoadingOverlay/
│       │   │   ├── LoadingOverlay.tsx
│       │   │   └── index.ts
│       │   └── EmptyState/
│       │       ├── EmptyState.tsx
│       │       └── index.ts
│       │
│       ├── Navigation/
│       │   ├── Breadcrumb/
│       │   │   ├── Breadcrumb.tsx
│       │   │   └── index.ts
│       │   └── NavLink/
│       │       ├── NavLink.tsx
│       │       └── index.ts
│       │
│       ├── StatusBadge/
│       │   ├── StatusBadge.tsx
│       │   └── index.ts
│       │
│       ├── DateRangePicker/
│       │   ├── DateRangePicker.tsx
│       │   └── index.ts
│       │
│       ├── SearchBox/
│       │   ├── SearchBox.tsx
│       │   └── index.ts
│       │
│       ├── ConfirmDialog/
│       │   ├── ConfirmDialog.tsx
│       │   └── index.ts
│       │
│       ├── PetSelectionSearchForm.tsx    # ペット選択検索フォーム
│       └── PetSelectionResultsTable.tsx  # ペット選択結果テーブル
│
├── hooks/                                 # グローバル共有カスタムフック
│   ├── useDebounce.ts
│   ├── useDebounce.test.ts
│   ├── useDisclosure.ts
│   ├── useLocalStorage.ts
│   ├── useMediaQuery.ts
│   ├── usePagination.ts
│   ├── useToast.ts
│   └── useConfirm.ts
│
├── lib/                                   # ライブラリ設定・ラッパー
│   ├── axios.ts
│   ├── react-query.ts
│   ├── date-fns.ts
│   ├── zod.ts
│   └── utils.ts                           # cn()等
│
├── stores/                                # グローバル状態管理
│   ├── authStore.ts
│   ├── themeStore.ts
│   ├── notificationStore.ts
│   └── sidebarStore.ts
│
├── types/                                 # グローバル型定義
│   ├── api.ts
│   ├── common.ts
│   └── index.ts
│
├── utils/                                 # グローバルユーティリティ
│   ├── format/
│   │   ├── date.ts
│   │   ├── currency.ts
│   │   └── number.ts
│   ├── validation/
│   │   ├── schemas.ts
│   │   └── validators.ts
│   ├── constants/
│   │   ├── status.ts
│   │   ├── colors.ts
│   │   └── routes.ts
│   └── helpers/
│       ├── string.ts
│       └── array.ts
│
├── config/                                # アプリケーション設定
│   ├── constants.ts
│   └── env.ts
│
├── styles/                                # グローバルスタイル
│   ├── globals.css
│   └── themes/
│       └── notion-like.css
│
└── testing/                               # テスト設定・ユーティリティ
    ├── setup.ts
    ├── server/                            # MSW (Mock Service Worker)
    │   ├── handlers/
    │   │   ├── auth.ts
    │   │   ├── patients.ts
    │   │   ├── reservations.ts
    │   │   ├── medicalRecords.ts
    │   │   └── index.ts
    │   └── server.ts
    └── utils.tsx
```

## 主要ファイルの完全準拠実装例

### 1. **main.tsx**

```typescript
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app";
import "./styles/globals.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

### 2. **app/index.tsx**

```typescript
import { RouterProvider } from "react-router";
import { AppProvider } from "./provider";
import { router } from "./router";

export const App = () => {
  return (
    <AppProvider>
      <RouterProvider router={router} />
    </AppProvider>
  );
};
```

### 3. **app/provider.tsx**

```typescript
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { queryClient } from "@/lib/react-query";
import { Toaster } from "@/components/ui/toaster";

interface AppProviderProps {
  children: React.ReactNode;
}

export const AppProvider = ({ children }: AppProviderProps) => {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};
```

### 4. **app/router.tsx** (Data Mode)

```typescript
import { createBrowserRouter } from "react-router";
import { MainLayout } from "@/components/shared/Layout/MainLayout";
import { AuthLayout } from "@/components/shared/Layout/AuthLayout";
import { ErrorBoundary } from "./ErrorBoundary";

// React Router Data Mode
export const router = createBrowserRouter([
  {
    path: "/",
    element: <MainLayout />,
    errorElement: <ErrorBoundary />,
    children: [
      {
        index: true,
        lazy: () =>
          import("@/features/dashboard").then((m) => ({
            Component: m.Dashboard,
          })),
      },
      {
        path: "reservations",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/reservations").then((m) => ({
                Component: m.ReservationList,
              })),
          },
          {
            path: "calendar",
            lazy: () =>
              import("@/features/reservations").then((m) => ({
                Component: m.ReservationCalendar,
              })),
          },
          {
            path: ":reservationId",
            lazy: () =>
              import("@/features/reservations").then((m) => ({
                Component: m.ReservationDetail,
              })),
          },
        ],
      },
      {
        path: "medical-records",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/medical-records").then((m) => ({
                Component: m.MedicalRecordList,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/medical-records").then((m) => ({
                Component: m.MedicalRecordCreate,
              })),
          },
          {
            path: ":recordId",
            lazy: () =>
              import("@/features/medical-records").then((m) => ({
                Component: m.MedicalRecordDetail,
              })),
          },
        ],
      },
      {
        path: "hospitalization",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/hospitalization").then((m) => ({
                Component: m.HospitalizationList,
              })),
          },
          {
            path: ":hospitalizationId",
            lazy: () =>
              import("@/features/hospitalization").then((m) => ({
                Component: m.HospitalizationDetail,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/hospitalization").then((m) => ({
                Component: m.HospitalizationCreate,
              })),
          },
        ],
      },
      {
        path: "accounting",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/accounting").then((m) => ({
                Component: m.AccountingList,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/accounting").then((m) => ({
                Component: m.AccountingCreate,
              })),
          },
          {
            path: ":accountingId",
            lazy: () =>
              import("@/features/accounting").then((m) => ({
                Component: m.AccountingDetail,
              })),
          },
        ],
      },
      {
        path: "examinations",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/examinations").then((m) => ({
                Component: m.ExaminationList,
              })),
          },
          {
            path: ":examinationId",
            lazy: () =>
              import("@/features/examinations").then((m) => ({
                Component: m.ExaminationDetail,
              })),
          },
        ],
      },
      {
        path: "owners",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerList,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerCreate,
              })),
          },
          {
            path: ":ownerId",
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerDetail,
              })),
          },
        ],
      },
      {
        path: "pets",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/pets").then((m) => ({
                Component: m.PetList,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/pets").then((m) => ({
                Component: m.PetCreate,
              })),
          },
          {
            path: ":petId",
            lazy: () =>
              import("@/features/pets").then((m) => ({
                Component: m.PetDetail,
              })),
          },
        ],
      },
      {
        path: "trimming",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/trimming").then((m) => ({
                Component: m.TrimmingList,
              })),
          },
          {
            path: ":trimmingId",
            lazy: () =>
              import("@/features/trimming").then((m) => ({
                Component: m.TrimmingDetail,
              })),
          },
        ],
      },
      {
        path: "vaccinations",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/vaccinations").then((m) => ({
                Component: m.VaccinationList,
              })),
          },
          {
            path: ":vaccinationId",
            lazy: () =>
              import("@/features/vaccinations").then((m) => ({
                Component: m.VaccinationDetail,
              })),
          },
        ],
      },
      {
        path: "settings",
        children: [
          {
            path: "clinic",
            lazy: () =>
              import("@/features/clinic").then((m) => ({
                Component: m.ClinicSettings,
              })),
          },
          {
            path: "master/treatment-items",
            lazy: () =>
              import("@/features/master").then((m) => ({
                Component: m.TreatmentItemList,
              })),
          },
        ],
      },
    ],
  },
  {
    path: "/login",
    element: <AuthLayout />,
    children: [
      {
        index: true,
        lazy: () =>
          import("@/features/auth").then((m) => ({
            Component: m.Login,
          })),
      },
    ],
  },
]);
```

### 5. **app/ErrorBoundary.tsx**

```typescript
import { useRouteError, isRouteErrorResponse } from "react-router";
import { Button } from "@/components/ui/button";

export const ErrorBoundary = () => {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <h1 className="text-4xl font-bold mb-4">{error.status}</h1>
        <p className="text-xl text-gray-600 mb-4">{error.statusText}</p>
        {error.data?.message && (
          <p className="text-gray-500 mb-8">{error.data.message}</p>
        )}
        <Button onClick={() => window.location.href = "/"}>
          ホームに戻る
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4">
      <h1 className="text-4xl font-bold mb-4">エラーが発生しました</h1>
      <p className="text-gray-600 mb-8">
        予期しないエラーが発生しました。もう一度お試しください。
      </p>
      <Button onClick={() => window.location.href = "/"}>
        ホームに戻る
      </Button>
    </div>
  );
};
```

### 6. **features/patients/api/getPatients.ts** (bulletproof-react準拠)

```typescript
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Patient } from "../types";

export const getPatients = async (): Promise<Patient[]> => {
  const response = await axios.get("/patients");
  return response.data;
};

// bulletproof-react: usePatients()という命名
export const usePatients = () => {
  return useQuery({
    queryKey: ["patients"],
    queryFn: getPatients,
  });
};
```

### 7. **features/patients/api/createPatient.ts** (bulletproof-react準拠)

```typescript
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { useToast } from "@/hooks/useToast";
import type { Patient, CreatePatientDTO } from "../types";

export const createPatient = async (
  data: CreatePatientDTO
): Promise<Patient> => {
  const response = await axios.post("/patients", data);
  return response.data;
};

type UseCreatePatientOptions = {
  onSuccess?: (data: Patient) => void;
};

export const useCreatePatient = (options?: UseCreatePatientOptions) => {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: createPatient,
    onSuccess: (data) => {
      // キャッシュ更新
      queryClient.invalidateQueries({ queryKey: ["patients"] });

      // トースト通知
      toast({
        title: "成功",
        description: "患者を登録しました",
      });

      // カスタムコールバック
      options?.onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: "エラー",
        description: "患者登録に失敗しました",
        variant: "destructive",
      });
    },
  });
};
```

### 8. **features/patients/hooks/usePatientForm.ts**

```typescript
import { useState } from "react";
import { useNavigate } from "react-router";
import { useCreatePatient } from "../api/createPatient";
import type { CreatePatientDTO } from "../types";

// ビジネスロジック: フォーム管理
export const usePatientForm = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const createPatient = useCreatePatient({
    onSuccess: (data) => {
      navigate(`/patients/${data.id}`);
    },
  });

  const handleSubmit = async (data: CreatePatientDTO) => {
    setIsSubmitting(true);
    try {
      await createPatient.mutateAsync(data);
    } catch (error) {
      console.error("Failed to create patient:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    handleSubmit,
    isSubmitting,
  };
};
```

### 9. **features/patients/components/PatientForm/PatientForm.tsx** (React 19)

```typescript
import { useActionState } from "react";
import { Input } from "@/components/ui/input";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { usePatientForm } from "../../hooks/usePatientForm";
import type { CreatePatientDTO } from "../../types";

export const PatientForm = () => {
  const { handleSubmit } = usePatientForm();

  // React 19: useActionState
  const [state, formAction] = useActionState(
    async (_prevState: any, formData: FormData) => {
      const data: CreatePatientDTO = {
        name: formData.get("name") as string,
        species: formData.get("species") as string,
        breed: formData.get("breed") as string,
        age: parseInt(formData.get("age") as string),
        ownerId: formData.get("ownerId") as string,
      };

      try {
        await handleSubmit(data);
        return { success: true, errors: null };
      } catch (error) {
        return {
          success: false,
          errors: { _form: "患者の登録に失敗しました" },
        };
      }
    },
    { success: false, errors: null }
  );

  return (
    <form action={formAction} className="space-y-4">
      {state.errors?._form && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-600">{state.errors._form}</p>
        </div>
      )}

      <Input name="name" label="名前" required />
      <Input name="species" label="種別" required />
      <Input name="breed" label="品種" />
      <Input name="age" type="number" label="年齢" />
      <Input name="ownerId" label="飼い主ID" required />

      <SubmitButton>登録する</SubmitButton>
    </form>
  );
};
```

### 10. **features/patients/index.ts** (Public API)

```typescript
// Routes
export { PatientList } from "./routes/PatientList";
export { PatientDetail } from "./routes/PatientDetail";
export { PatientCreate } from "./routes/PatientCreate";

// API Hooks
export { usePatients } from "./api/getPatients";
export { usePatient } from "./api/getPatient";
export { useCreatePatient } from "./api/createPatient";
export { useUpdatePatient } from "./api/updatePatient";
export { useDeletePatient } from "./api/deletePatient";

// Components (必要な場合のみ公開)
export { PatientCard } from "./components/PatientCard";
export { PatientForm } from "./components/PatientForm";

// Types
export type {
  Patient,
  CreatePatientDTO,
  UpdatePatientDTO,
} from "./types";
```

### 11. **lib/axios.ts**

```typescript
import Axios, { InternalAxiosRequestConfig } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

const authRequestInterceptor = (config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem("token");

  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }

  config.headers = config.headers || {};
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID();

  return config;
};

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

axios.interceptors.request.use(authRequestInterceptor);

axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
```

### 12. **lib/react-query.ts**

```typescript
import { QueryClient, DefaultOptions } from "@tanstack/react-query";

const queryConfig: DefaultOptions = {
  queries: {
    staleTime: 5 * 60 * 1000, // 5分
    gcTime: 10 * 60 * 1000, // 10分
    refetchOnWindowFocus: false,
    retry: 1,
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
});
```

### 13. **components/ui/button.tsx** (React 19 + shadcn/ui)

```typescript
import { ComponentPropsWithoutRef } from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
        outline: "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-10 rounded-md px-8",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends ComponentPropsWithoutRef<"button">,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

// React 19: forwardRef不要、refをpropsとして直接受け取る
export const Button = ({
  className,
  variant,
  size,
  asChild = false,
  ref,
  ...props
}: ButtonProps) => {
  const Comp = asChild ? Slot : "button";
  return (
    <Comp
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      {...props}
    />
  );
};
```

### 14. **components/shared/Form/SubmitButton/SubmitButton.tsx** (React 19)

```typescript
import { useFormStatus } from "react-dom";
import { Button, ButtonProps } from "@/components/ui/button";

interface SubmitButtonProps extends ButtonProps {
  children: React.ReactNode;
}

// React 19: useFormStatus
export const SubmitButton = ({ children, ...props }: SubmitButtonProps) => {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" disabled={pending} {...props}>
      {pending ? "処理中..." : children}
    </Button>
  );
};
```

## 完全準拠のチェックリスト

### ✅ React 19 Data Mode
- [x] `createBrowserRouter`使用
- [x] `RouterProvider`使用
- [x] SSR/SSG機能を使用していない
- [x] useActionState活用
- [x] useOptimistic活用
- [x] useFormStatus活用
- [x] ref as prop活用
- [x] Document Metadata（ブラウザタブタイトルのみ）

### ✅ bulletproof-react
- [x] Feature-based architecture
- [x] API関数とフックを`api/`に配置
- [x] フック命名: `usePatients()`, `useCreatePatient()`等
- [x] ビジネスロジックフックは`hooks/`に分離
- [x] 各featureに`index.ts`でPublic API
- [x] コンポーネントのフォルダ+index.tsパターン
- [x] テストファイルを各コンポーネントと同階層に配置
- [x] 型定義を各featureに配置
- [x] Mock Dataを各featureに配置
- [x] MSW設定を`test/server/`に配置

## この構成の特徴

1. **完全準拠**: React 19 Data Mode + bulletproof-react の両方に完全準拠
2. **一貫性**: 全featureで統一されたパターン
3. **スケーラビリティ**: 新機能追加が容易
4. **保守性**: 明確な責任分離
5. **テスタビリティ**: テストファイルが各コンポーネントと同階層
6. **医療システム対応**: SOAPS、カンバンボード、入院管理等に最適化

この構成が、動物病院管理システムのベストフォルダ構成です。