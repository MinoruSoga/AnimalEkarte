# 最終DDLとcanonical modelの対応

基準 `7f164f8a71baf5bfb23207ddce5b3049939e81a3`。ALTER/DROP後の241表・2,691列。全列は[対応表JSON](model-schema-map.json)と[列別TSV](columns.tsv)へ。DB未接続。新規モデルはFK列を持ち、ナビゲーション／自動書込みgraphを追加しない。

| テーブル | Model／ソース | 列数 | 主キー | FK数 | 扱い |
|:---|:---|---:|:---|---:|:---|
| `accounts` | [Account](../../../backend/internal/model/account.go) | 8 | id | 0 | 既存対応 |
| `actual_visit_facts` | [ActualVisitFact](../../../backend/internal/model/storage_clinical.go) | 14 | id | 8 | 追加 |
| `admission_records` | [AdmissionRecord](../../../backend/internal/model/storage_reservations.go) | 6 | id | 3 | 追加 |
| `animal_species` | [AnimalSpecies](../../../backend/internal/model/animal_species.go) | 6 | id | 0 | 既存対応 |
| `appointment_change_events` | [AppointmentChangeEvent](../../../backend/internal/model/storage_reservations.go) | 11 | id | 4 | 追加 |
| `appointment_trimming_details` | [AppointmentTrimmingDetail](../../../backend/internal/model/trimming.go) | 14 | id | 3 | 既存対応 |
| `appointments` | [Reservation](../../../backend/internal/model/reservation.go) | 31 | id | 11 | 既存対応 |
| `audit_logs` | [AuditLog](../../../backend/internal/model/audit_log.go) | 13 | id | 2 | 既存対応 |
| `billing_calculation_lines` | [BillingCalculationLine](../../../backend/internal/model/storage_finance.go) | 16 | clinic_id, calculation_id, billing_item_id | 3 | 追加 |
| `billing_calculation_snapshots` | [BillingCalculationSnapshot](../../../backend/internal/model/storage_finance.go) | 15 | id | 3 | 追加 |
| `billing_change_events` | [BillingChangeEvent](../../../backend/internal/model/storage_finance.go) | 12 | id | 5 | 追加 |
| `billing_fulfillment_events` | [BillingFulfillmentEvent](../../../backend/internal/model/storage_finance.go) | 11 | id | 5 | 追加 |
| `billing_items` | [BillingItem](../../../backend/internal/model/accounting.go) | 24 | id | 6 | 既存対応 |
| `billing_price_acknowledgements` | [BillingPriceAcknowledgement](../../../backend/internal/model/storage_finance.go) | 13 | id | 5 | 追加 |
| `billing_reissue_links` | [BillingReissueLink](../../../backend/internal/model/storage_finance.go) | 7 | clinic_id, new_billing_id | 5 | 追加 |
| `billing_source_bindings` | [BillingSourceBinding](../../../backend/internal/model/storage_finance.go) | 15 | clinic_id, billing_item_id | 7 | 追加 |
| `billing_source_difference_resolutions` | [BillingSourceDifferenceResolution](../../../backend/internal/model/storage_finance.go) | 6 | clinic_id, difference_id | 3 | 追加 |
| `billing_source_differences` | [BillingSourceDifference](../../../backend/internal/model/storage_finance.go) | 9 | id | 2 | 追加 |
| `billings` | [Billing](../../../backend/internal/model/accounting.go) | 17 | id | 9 | 既存対応 |
| `cages` | [Cage](../../../backend/internal/model/cage.go) | 13 | id | 2 | 既存対応 |
| `campaign_target_categories` | [CampaignTargetCategory](../../../backend/internal/model/campaign.go) | 3 | id | 1 | 既存対応 |
| `campaign_target_items` | [CampaignTargetItem](../../../backend/internal/model/campaign.go) | 3 | id | 2 | 既存対応 |
| `campaigns` | [Campaign](../../../backend/internal/model/campaign.go) | 12 | id | 1 | 既存対応 |
| `care_logs` | [CareLog](../../../backend/internal/model/hospitalization.go) | 11 | id | 3 | 既存対応 |
| `care_plan_items` | [CarePlanItem](../../../backend/internal/model/hospitalization.go) | 17 | id | 4 | 既存対応 |
| `cash_actual_corrections` | [CashActualCorrection](../../../backend/internal/model/storage_finance.go) | 14 | id | 5 | 追加 |
| `cash_adjustment_events` | [CashAdjustmentEvent](../../../backend/internal/model/storage_finance.go) | 9 | id | 3 | 追加 |
| `cash_close_contexts` | [CashCloseContext](../../../backend/internal/model/storage_finance.go) | 10 | clinic_id, close_id | 5 | 追加 |
| `cash_close_late_events` | [CashCloseLateEvent](../../../backend/internal/model/storage_finance.go) | 4 | clinic_id, source_event_id | 3 | 追加 |
| `cash_close_source_events` | [CashCloseSourceEvent](../../../backend/internal/model/storage_finance.go) | 4 | clinic_id, close_id, source_event_id | 3 | 追加 |
| `cash_opening_balances` | [CashOpeningBalance](../../../backend/internal/model/storage_finance.go) | 11 | id | 4 | 追加 |
| `cash_register_closes` | [CashRegisterClose](../../../backend/internal/model/cash_register_close.go) | 12 | id | 2 | 既存対応 |
| `checkup_field_results` | [CheckupFieldResult](../../../backend/internal/model/checkup_field.go) | 18 | id | 3 | 既存対応 |
| `checkup_package_import_receipts` | [CheckupPackageImportReceipt](../../../backend/internal/model/checkup_package_import.go) | 12 | id | 2 | 既存対応 |
| `checkup_result_revisions` | [CheckupResultRevision](../../../backend/internal/model/storage_clinical.go) | 12 | id | 5 | 追加 |
| `checkup_save_attempts` | [CheckupSaveAttempt](../../../backend/internal/model/storage_clinical.go) | 8 | id | 4 | 追加 |
| `checkup_save_stage_results` | [CheckupSaveStageResult](../../../backend/internal/model/storage_clinical.go) | 8 | clinic_id, attempt_id, stage | 5 | 追加 |
| `checkup_type_fields` | [CheckupTypeField](../../../backend/internal/model/checkup_field.go) | 17 | id | 2 | 既存対応 |
| `checkup_types` | [CheckupType](../../../backend/internal/model/checkup_type.go) | 17 | id | 2 | 既存対応 |
| `checkups` | [Checkup](../../../backend/internal/model/checkup_record.go) | 13 | id | 5 | 既存対応 |
| `chief_complaint_types` | [ChiefComplaintType](../../../backend/internal/model/chief_complaint_type.go) | 9 | id | 1 | 既存対応 |
| `clinic_holidays` | [ClinicHoliday](../../../backend/internal/model/clinic_holiday.go) | 6 | id | 1 | 既存対応 |
| `clinic_integrations` | [ClinicIntegration](../../../backend/internal/model/clinic_integration.go) | 7 | id | 1 | 既存対応 |
| `clinic_settings` | [ClinicSettings](../../../backend/internal/model/clinic_settings.go) | 32 | clinic_id | 1 | 既存対応 |
| `clinical_plan_diagnoses` | [ClinicalPlanDiagnosis](../../../backend/internal/model/storage_clinical.go) | 13 | id | 4 | 追加 |
| `clinical_plans` | [ClinicalPlan](../../../backend/internal/model/clinical_plan.go) | 10 | id | 3 | 既存対応 |
| `clinics` | [Clinic](../../../backend/internal/model/clinic.go) | 26 | id | 1 | 既存対応 |
| `closing_boundary_revisions` | [ClosingBoundaryRevision](../../../backend/internal/model/storage_finance.go) | 5 | clinic_id, revision | 2 | 追加 |
| `closing_special_periods` | [ClosingSpecialPeriod](../../../backend/internal/model/closing_special_period.go) | 9 | id | 1 | 既存対応 |
| `common_owner_basic_revisions` | [CommonOwnerBasicRevision](../../../backend/internal/model/storage_identity.go) | 24 | id | 5 | 追加 |
| `common_owner_clinic_bindings` | [CommonOwnerClinicBinding](../../../backend/internal/model/storage_identity.go) | 7 | id | 5 | 追加 |
| `common_owner_subjects` | [CommonOwnerSubject](../../../backend/internal/model/storage_identity.go) | 4 | id | 1 | 追加 |
| `common_pet_basic_revisions` | [CommonPetBasicRevision](../../../backend/internal/model/storage_identity.go) | 34 | id | 6 | 追加 |
| `common_pet_clinic_bindings` | [CommonPetClinicBinding](../../../backend/internal/model/storage_identity.go) | 7 | id | 5 | 追加 |
| `common_pet_subjects` | [CommonPetSubject](../../../backend/internal/model/storage_identity.go) | 4 | id | 1 | 追加 |
| `companies` | [Company](../../../backend/internal/model/company.go) | 14 | id | 0 | 既存対応 |
| `consultations` | [Consultation](../../../backend/internal/model/consultation.go) | 16 | id | 2 | 既存対応 |
| `daily_records` | [DailyRecord](../../../backend/internal/model/hospitalization.go) | 6 | id | 2 | 既存対応 |
| `diagnosis_names` | [DiagnosisName](../../../backend/internal/model/diagnosis.go) | 10 | id | 2 | 既存対応 |
| `diagnosis_types` | [DiagnosisType](../../../backend/internal/model/diagnosis.go) | 9 | id | 1 | 既存対応 |
| `doctor_review_records` | [DoctorReviewRecord](../../../backend/internal/model/storage_finance.go) | 7 | clinic_id, review_set_id, medical_record_id | 3 | 追加 |
| `doctor_review_sets` | [DoctorReviewSet](../../../backend/internal/model/storage_finance.go) | 9 | id | 3 | 追加 |
| `estimate_billing_handoffs` | [EstimateBillingHandoff](../../../backend/internal/model/storage_finance.go) | 7 | clinic_id, presentation_id, estimate_item_id | 5 | 追加 |
| `estimate_calculation_revisions` | [EstimateCalculationRevision](../../../backend/internal/model/storage_finance.go) | 15 | id | 4 | 追加 |
| `estimate_items` | [EstimateItem](../../../backend/internal/model/estimate.go) | 19 | id | 5 | 既存対応 |
| `estimate_presentation_lines` | [EstimatePresentationLine](../../../backend/internal/model/storage_finance.go) | 8 | clinic_id, presentation_id, estimate_item_id | 3 | 追加 |
| `estimate_presentations` | [EstimatePresentation](../../../backend/internal/model/storage_finance.go) | 13 | id | 5 | 追加 |
| `estimate_replies` | [EstimateReply](../../../backend/internal/model/storage_finance.go) | 9 | id | 4 | 追加 |
| `estimate_successor_links` | [EstimateSuccessorLink](../../../backend/internal/model/storage_finance.go) | 6 | clinic_id, successor_estimate_id | 4 | 追加 |
| `estimates` | [Estimate](../../../backend/internal/model/estimate.go) | 17 | id | 6 | 既存対応 |
| `exam_reference_ranges` | [ExamReferenceRange](../../../backend/internal/model/exam_reference_range.go) | 10 | id | 3 | 既存対応 |
| `exam_results` | [ExamResult](../../../backend/internal/model/examination_record.go) | 18 | id | 2 | 既存対応 |
| `exam_type_fields` | [ExamTypeField](../../../backend/internal/model/examination_type.go) | 10 | id | 2 | 既存対応 |
| `exam_types` | [ExaminationType](../../../backend/internal/model/examination_type.go) | 12 | id | 2 | 既存対応 |
| `examination_field_semantic_mappings` | [ExaminationFieldSemanticMapping](../../../backend/internal/model/storage_clinical.go) | 10 | id | 5 | 追加 |
| `examination_revision_items` | [ExaminationRevisionItem](../../../backend/internal/model/examination_revision.go) | 20 | id | 1 | 既存対応 |
| `examination_revisions` | [ExaminationRevision](../../../backend/internal/model/examination_revision.go) | 27 | id | 2 | 既存対応 |
| `exams` | [Examination](../../../backend/internal/model/examination_record.go) | 16 | id | 7 | 既存対応 |
| `finance_event_ingestions` | [FinanceEventIngestion](../../../backend/internal/model/storage_finance.go) | 5 | clinic_id, consumer_kind, source_event_id | 2 | 追加 |
| `finance_source_events` | [FinanceSourceEvent](../../../backend/internal/model/storage_finance.go) | 22 | id | 10 | 追加 |
| `financial_document_artifacts` | [FinancialDocumentArtifact](../../../backend/internal/model/storage_finance.go) | 6 | clinic_id, edition_id | 2 | 追加 |
| `financial_document_editions` | [FinancialDocumentEdition](../../../backend/internal/model/storage_finance.go) | 12 | id | 6 | 追加 |
| `financial_document_identities` | [FinancialDocumentIdentity](../../../backend/internal/model/storage_finance.go) | 9 | id | 5 | 追加 |
| `followup_confirmations` | [FollowupConfirmation](../../../backend/internal/model/storage_messaging.go) | 14 | id | 7 | 追加 |
| `fund_consumption_events` | [FundConsumptionEvent](../../../backend/internal/model/storage_finance.go) | 16 | id | 8 | 追加 |
| `fund_record_corrections` | [FundRecordCorrection](../../../backend/internal/model/storage_finance.go) | 15 | id | 5 | 追加 |
| `hospitalization_plans` | [HospitalizationPlan](../../../backend/internal/model/hospitalization_plan.go) | 14 | id | 1 | 既存対応 |
| `hospitalizations` | [Hospitalization](../../../backend/internal/model/hospitalization.go) | 17 | id | 8 | 既存対応 |
| `hotel_capacity_settings` | [HotelCapacitySetting](../../../backend/internal/model/storage_reservations.go) | 5 | clinic_id | 2 | 追加 |
| `initial_intake_entries` | [InitialIntakeEntry](../../../backend/internal/model/storage_clinical.go) | 16 | clinic_id, intake_revision_id, entry_key | 3 | 追加 |
| `initial_intake_revisions` | [InitialIntakeRevision](../../../backend/internal/model/storage_clinical.go) | 13 | id | 7 | 追加 |
| `inquiries` | [Inquiry](../../../backend/internal/model/inquiry.go) | 17 | id | 3 | 既存対応 |
| `inquiry_templates` | [InquiryTemplate](../../../backend/internal/model/inquiry_template.go) | 10 | id | 1 | 既存対応 |
| `insurances` | [Insurance](../../../backend/internal/model/insurance.go) | 11 | id | 1 | 既存対応 |
| `inventory_items` | [InventoryItem](../../../backend/internal/model/inventory.go) | 15 | id | 2 | 既存対応 |
| `inventory_movement_events` | [InventoryMovementEvent](../../../backend/internal/model/storage_clinical.go) | 20 | id | 8 | 追加 |
| `inventory_quantity_bases` | [InventoryQuantityBase](../../../backend/internal/model/storage_clinical.go) | 10 | id | 4 | 追加 |
| `lab_device_item_masters` | [LabDeviceItemMaster](../../../backend/internal/model/lab_device_item_master.go) | 11 | id | 2 | 既存対応 |
| `lab_device_mapping_draft_items` | [LabDeviceMappingDraftItem](../../../backend/internal/model/lab_device_mapping_draft.go) | 18 | id | 3 | 既存対応 |
| `lab_device_mapping_drafts` | [LabDeviceMappingDraft](../../../backend/internal/model/lab_device_mapping_draft.go) | 13 | id | 1 | 既存対応 |
| `lab_device_mapping_generation` | [LabDeviceMappingGeneration](../../../backend/internal/model/lab_device_mapping_draft.go) | 3 | clinic_id | 1 | 既存対応 |
| `lab_device_mapping_specs` | [LabDeviceMappingSpec](../../../backend/internal/model/lab_device_mapping_draft.go) | 12 | id | 3 | 既存対応 |
| `lab_device_station_settings` | [LabDeviceStationSettings](../../../backend/internal/model/lab_device_receive.go) | 5 | clinic_id | 1 | 既存対応 |
| `lab_device_waits` | [LabDeviceWait](../../../backend/internal/model/lab_device_receive.go) | 8 | id | 3 | 既存対応 |
| `lab_devices` | [LabDevice](../../../backend/internal/model/lab_device.go) | 9 | id | 2 | 既存対応 |
| `lab_import_events` | [LabImportEvent](../../../backend/internal/model/lab_import.go) | 12 | id | 2 | 既存対応 |
| `lab_import_exam_retraction_items` | [LabImportExamRetractionItem](../../../backend/internal/model/lab_import.go) | 8 | id | 1 | 既存対応 |
| `lab_import_exam_retractions` | [LabImportExamRetraction](../../../backend/internal/model/lab_import.go) | 8 | id | 2 | 既存対応 |
| `lab_import_job_items` | [LabImportJobItem](../../../backend/internal/model/lab_device_receive.go) | 11 | id | 3 | 既存対応 |
| `lab_import_jobs` | [LabImportJob](../../../backend/internal/model/lab_import.go) | 23 | id | 2 | 既存対応 |
| `lab_import_revert_receipts` | [LabImportRevertReceipt](../../../backend/internal/model/lab_import.go) | 10 | id | 1 | 既存対応 |
| `lab_import_usage_receipts` | [LabImportUsageReceipt](../../../backend/internal/model/lab_import.go) | 7 | id | 2 | 既存対応 |
| `line_customers` | [LineCustomer](../../../backend/internal/model/line_customer.go) | 10 | id | 1 | 既存対応 |
| `line_dispatch_attempts` | [LineDispatchAttempt](../../../backend/internal/model/storage_messaging.go) | 9 | id | 4 | 追加 |
| `line_dispatch_intents` | [LineDispatchIntent](../../../backend/internal/model/storage_messaging.go) | 14 | id | 5 | 追加 |
| `line_dispatch_results` | [LineDispatchResult](../../../backend/internal/model/storage_messaging.go) | 10 | id | 2 | 追加 |
| `line_link_tokens` | [LineLinkToken](../../../backend/internal/model/line_link_token.go) | 7 | id | 2 | 既存対応 |
| `line_reservation_settings` | [LineReservationSetting](../../../backend/internal/model/line_reservation_setting.go) | 28 | id | 1 | 既存対応 |
| `lstep_auto_managed_prefixes` | [LstepAutoManagedPrefix](../../../backend/internal/model/lstep_auto_managed_prefix.go) | 6 | id | 0 | 既存対応 |
| `lstep_condition_tag_mappings` | [LstepConditionTagMapping](../../../backend/internal/model/lstep_condition_tag_mapping.go) | 6 | id | 0 | 既存対応 |
| `lstep_csv_imports` | [LstepCsvImport](../../../backend/internal/model/lstep_csv_import.go) | 12 | id | 2 | 既存対応 |
| `lstep_delivery_trigger_log` | [LstepDeliveryTriggerLog](../../../backend/internal/model/lstep_delivery_trigger_log.go) | 12 | id | 2 | 既存対応 |
| `lstep_friend_attribute_snapshots` | [LstepFriendAttributeSnapshot](../../../backend/internal/model/lstep_friend_attribute_snapshot.go) | 14 | id | 2 | 既存対応 |
| `lstep_migration_progress` | [LstepMigrationProgress](../../../backend/internal/model/storage_messaging.go) | 9 | id | 2 | 追加 |
| `lstep_next_visit_sync_retries` | [LstepNextVisitSyncRetry](../../../backend/internal/model/storage_messaging.go) | 6 | clinic_id, owner_id | 2 | 追加 |
| `lstep_send_purpose_tag_prefixes` | [LstepSendPurposeTagPrefix](../../../backend/internal/model/lstep_send_purpose_tag_prefix.go) | 6 | id | 0 | 既存対応 |
| `lstep_settings` | [LstepSettings](../../../backend/internal/model/lstep_settings.go) | 6 | id | 1 | 既存対応 |
| `lstep_sync_error_counters` | [LstepSyncErrorCounter](../../../backend/internal/model/lstep_sync_error_counter.go) | 6 | id | 2 | 既存対応 |
| `lstep_tag_cache` | [LstepTagCache](../../../backend/internal/model/lstep_tag_cache.go) | 7 | id | 2 | 既存対応 |
| `lstep_tag_code_mappings` | [LstepTagCodeMapping](../../../backend/internal/model/lstep_tag_code_mapping.go) | 10 | id | 1 | 既存対応 |
| `lstep_trigger_priorities` | [LstepTriggerPriority](../../../backend/internal/model/lstep_trigger_priority.go) | 6 | id | 1 | 既存対応 |
| `manual_article_operations` | [ManualArticleOperation](../../../backend/internal/model/storage_messaging.go) | 14 | id | 7 | 追加 |
| `manual_article_versions` | [ManualArticleVersion](../../../backend/internal/model/manual_article.go) | 12 | id | 3 | 既存対応 |
| `manual_articles` | [ManualArticle](../../../backend/internal/model/manual_article.go) | 9 | id | 2 | 既存対応 |
| `medical_record_addenda` | [MedicalRecordAddendum](../../../backend/internal/model/medical_record_addendum.go) | 8 | id | 3 | 既存対応 |
| `medical_record_consultation_events` | [MedicalRecordConsultationEvent](../../../backend/internal/model/storage_clinical.go) | 12 | id | 5 | 追加 |
| `medical_record_correction_links` | [MedicalRecordCorrectionLink](../../../backend/internal/model/storage_clinical.go) | 13 | id | 6 | 追加 |
| `medical_record_dependency_versions` | [MedicalRecordDependencyVersion](../../../backend/internal/model/storage_clinical.go) | 4 | clinic_id, medical_record_id, owner_kind | 2 | 追加 |
| `medical_record_image_deletions` | [MedicalRecordImageDeletion](../../../backend/internal/model/storage_clinical.go) | 13 | id | 1 | 追加 |
| `medical_record_image_upload_quota` | [MedicalRecordImageUploadQuota](../../../backend/internal/model/storage_clinical.go) | 6 | id | 2 | 追加 |
| `medical_record_images` | [MedicalRecordImage](../../../backend/internal/model/medical_record_image.go) | 15 | id | 3 | 既存対応 |
| `medical_records` | [MedicalRecord](../../../backend/internal/model/medical_record.go) | 21 | id | 9 | 既存対応 |
| `medicine_dose_params` | [MedicineDoseParam](../../../backend/internal/model/medicine_dose_param.go) | 20 | id | 2 | 既存対応 |
| `medicine_dose_pattern_revisions` | [MedicineDosePatternRevision](../../../backend/internal/model/storage_clinical.go) | 15 | id | 4 | 追加 |
| `medicines` | [Medicine](../../../backend/internal/model/medicine.go) | 22 | id | 3 | 既存対応 |
| `merchandise_items` | [MerchandiseItem](../../../backend/internal/model/merchandise_item.go) | 12 | id | 1 | 既存対応 |
| `occupations` | [Occupation](../../../backend/internal/model/occupation.go) | 9 | id | 1 | 既存対応 |
| `operation_attempt_outcomes` | [OperationAttemptOutcome](../../../backend/internal/model/storage_clinical.go) | 7 | id | 2 | 追加 |
| `operation_bindings` | [OperationBinding](../../../backend/internal/model/storage_clinical.go) | 10 | id | 1 | 追加 |
| `operation_results` | [OperationResult](../../../backend/internal/model/storage_clinical.go) | 6 | clinic_id, binding_id | 2 | 追加 |
| `owner_deposit_accounts` | [OwnerDepositAccount](../../../backend/internal/model/storage_finance.go) | 4 | clinic_id, owner_id | 2 | 追加 |
| `owner_deposit_entries` | [OwnerDepositEntry](../../../backend/internal/model/storage_finance.go) | 15 | id | 6 | 追加 |
| `owner_deposit_operations` | [OwnerDepositOperation](../../../backend/internal/model/storage_finance.go) | 10 | clinic_id, owner_id, operation_key | 3 | 追加 |
| `owner_identity_confirmations` | [OwnerIdentityConfirmation](../../../backend/internal/model/storage_identity.go) | 9 | id | 5 | 追加 |
| `owner_identity_group_members` | [OwnerIdentityGroupMember](../../../backend/internal/model/identity_link.go) | 8 | id | 2 | 既存対応 |
| `owner_identity_groups` | [OwnerIdentityGroup](../../../backend/internal/model/identity_link.go) | 6 | id | 1 | 既存対応 |
| `owner_line_binding_events` | [OwnerLineBindingEvent](../../../backend/internal/model/storage_messaging.go) | 9 | id | 3 | 追加 |
| `owner_line_binding_states` | [OwnerLineBindingState](../../../backend/internal/model/storage_messaging.go) | 8 | clinic_id, owner_id | 6 | 追加 |
| `owners` | [Owner](../../../backend/internal/model/owner.go) | 19 | id | 1 | 既存対応 |
| `password_reset_tokens` | [PasswordResetToken](../../../backend/internal/model/password_reset_token.go) | 5 | id | 1 | 既存対応 |
| `patient_operation_gates` | [PatientOperationGate](../../../backend/internal/model/storage_reservations.go) | 3 | clinic_id, pet_id | 2 | 追加 |
| `payment_funds` | [PaymentFund](../../../backend/internal/model/storage_finance.go) | 12 | id | 5 | 追加 |
| `payment_methods` | [PaymentMethodMaster](../../../backend/internal/model/payment_method_master.go) | 9 | id | 1 | 既存対応 |
| `permission_group_rules` | [PermissionGroupRule](../../../backend/internal/model/permission_group.go) | 10 | id | 1 | 既存対応 |
| `permission_groups` | [PermissionGroup](../../../backend/internal/model/permission_group.go) | 10 | id | 1 | 既存対応 |
| `pet_chronic_conditions` | [PetChronicCondition](../../../backend/internal/model/pet_chronic_condition.go) | 11 | id | 2 | 既存対応 |
| `pet_identity_group_members` | [PetIdentityGroupMember](../../../backend/internal/model/identity_link.go) | 8 | id | 2 | 既存対応 |
| `pet_identity_groups` | [PetIdentityGroup](../../../backend/internal/model/identity_link.go) | 8 | id | 2 | 既存対応 |
| `pet_owners` | [PetOwner](../../../backend/internal/model/pet_owner.go) | 7 | id | 3 | 既存対応 |
| `pets` | [Pet](../../../backend/internal/model/pet.go) | 9 | id | 4 | 既存対応 |
| `preventive_medication_observations` | [PreventiveMedicationObservation](../../../backend/internal/model/storage_clinical.go) | 5 | id | 2 | 追加 |
| `preventive_medication_revisions` | [PreventiveMedicationRevision](../../../backend/internal/model/storage_clinical.go) | 20 | id | 6 | 追加 |
| `preventive_prescription_revisions` | [PreventivePrescriptionRevision](../../../backend/internal/model/storage_clinical.go) | 27 | id | 7 | 追加 |
| `preventive_prescriptions` | [PreventivePrescription](../../../backend/internal/model/storage_clinical.go) | 6 | id | 6 | 追加 |
| `procedures` | [Procedure](../../../backend/internal/model/procedure.go) | 17 | id | 2 | 既存対応 |
| `record_edit_right_events` | [RecordEditRightEvent](../../../backend/internal/model/storage_clinical.go) | 11 | id | 5 | 追加 |
| `record_edit_rights` | [RecordEditRight](../../../backend/internal/model/storage_clinical.go) | 9 | clinic_id, medical_record_id | 3 | 追加 |
| `refund_requests` | [RefundRequest](../../../backend/internal/model/storage_clinical.go) | 12 | id | 5 | 追加 |
| `reservation_capacity_gates` | [ReservationCapacityGate](../../../backend/internal/model/storage_reservations.go) | 4 | clinic_id, capacity_kind, resource_key | 1 | 追加 |
| `reservation_contract_states` | [ReservationContractState](../../../backend/internal/model/storage_reservations.go) | 6 | clinic_id, contract_kind | 1 | 追加 |
| `reservation_eligibility_evidence` | [ReservationEligibilityEvidence](../../../backend/internal/model/storage_reservations.go) | 7 | id | 2 | 追加 |
| `reservation_notification_eligibility` | [ReservationNotificationEligibility](../../../backend/internal/model/storage_reservations.go) | 9 | clinic_id, appointment_id | 5 | 追加 |
| `reservation_type_available_slots` | [ReservationTypeAvailableSlot](../../../backend/internal/model/reservation_type.go) | 10 | id | 2 | 既存対応 |
| `reservation_type_groups` | [ReservationTypeGroup](../../../backend/internal/model/reservation_type_group.go) | 10 | id | 1 | 既存対応 |
| `reservation_type_name_revisions` | [ReservationTypeNameRevision](../../../backend/internal/model/storage_reservations.go) | 8 | clinic_id, reservation_type_id, revision | 4 | 追加 |
| `reservation_type_unavailable_times` | [ReservationTypeUnavailableTime](../../../backend/internal/model/reservation_type.go) | 10 | id | 2 | 既存対応 |
| `reservation_types` | [ReservationType](../../../backend/internal/model/reservation_type.go) | 25 | id | 3 | 既存対応 |
| `revisit_intake_answers` | [RevisitIntakeAnswer](../../../backend/internal/model/storage_clinical.go) | 18 | id | 5 | 追加 |
| `revisit_intake_submit_receipts` | [RevisitIntakeSubmitReceipt](../../../backend/internal/model/storage_clinical.go) | 10 | clinic_id, author_customer_id, operation_key | 2 | 追加 |
| `revisit_intake_targets` | [RevisitIntakeTarget](../../../backend/internal/model/storage_clinical.go) | 6 | id | 3 | 追加 |
| `shared_files` | [SharedFile](../../../backend/internal/model/shared_file.go) | 13 | id | 3 | 既存対応 |
| `shift_bulk_operations` | [ShiftBulkOperation](../../../backend/internal/model/storage_reservations.go) | 6 | clinic_id, operation_id | 2 | 追加 |
| `shift_entries` | [ShiftEntry](../../../backend/internal/model/staff.go) | 11 | id | 2 | 既存対応 |
| `shift_entry_breaks` | [ShiftEntryBreak](../../../backend/internal/model/shift_entry_break.go) | 4 | id | 1 | 既存対応 |
| `shift_template_breaks` | [ShiftTemplateBreak](../../../backend/internal/model/staff.go) | 4 | id | 1 | 既存対応 |
| `shift_templates` | [ShiftTemplate](../../../backend/internal/model/staff.go) | 12 | id | 1 | 既存対応 |
| `staff_clinic_assignments` | [StaffClinicAssignment](../../../backend/internal/model/account.go) | 9 | id | 2 | 既存対応 |
| `staff_leave_changes` | [StaffLeaveChange](../../../backend/internal/model/storage_reservations.go) | 9 | id | 4 | 追加 |
| `staff_leave_periods` | [StaffLeavePeriod](../../../backend/internal/model/storage_reservations.go) | 10 | id | 2 | 追加 |
| `staff_notes` | [StaffNote](../../../backend/internal/model/hospitalization.go) | 7 | id | 2 | 既存対応 |
| `staff_permission_groups` | [StaffPermissionGroup](../../../backend/internal/model/permission_group.go) | 3 | staff_id, group_id | 2 | 既存対応 |
| `staff_record_draft_operations` | [StaffRecordDraftOperation](../../../backend/internal/model/storage_clinical.go) | 4 | clinic_id, draft_id, operation_id | 3 | 追加 |
| `staff_record_drafts` | [StaffRecordDraft](../../../backend/internal/model/storage_clinical.go) | 12 | id | 3 | 追加 |
| `staff_reservation_capabilities` | [StaffReservationCapability](../../../backend/internal/model/staff_reservation_capability.go) | 5 | id | 3 | 既存対応 |
| `staffs` | [Staff](../../../backend/internal/model/staff.go) | 18 | id | 3 | 既存対応 |
| `stay_cases` | [StayCase](../../../backend/internal/model/storage_reservations.go) | 10 | id | 3 | 追加 |
| `stay_extension_proposals` | [StayExtensionProposal](../../../backend/internal/model/storage_reservations.go) | 11 | id | 5 | 追加 |
| `stay_movement_events` | [StayMovementEvent](../../../backend/internal/model/storage_reservations.go) | 14 | id | 8 | 追加 |
| `stay_plan_change_events` | [StayPlanChangeEvent](../../../backend/internal/model/storage_reservations.go) | 15 | id | 8 | 追加 |
| `stay_price_agreements` | [StayPriceAgreement](../../../backend/internal/model/storage_reservations.go) | 11 | id | 4 | 追加 |
| `support_bug_reports` | [SupportBugReport](../../../backend/internal/model/support_bug_report.go) | 16 | id | 2 | 既存対応 |
| `support_chat_messages` | [SupportChatMessage](../../../backend/internal/model/support_chat_message.go) | 8 | id | 2 | 既存対応 |
| `support_delivery_attempts` | [SupportDeliveryAttempt](../../../backend/internal/model/storage_messaging.go) | 9 | id | 4 | 追加 |
| `support_delivery_destinations` | [SupportDeliveryDestination](../../../backend/internal/model/storage_messaging.go) | 5 | id | 2 | 追加 |
| `support_delivery_intents` | [SupportDeliveryIntent](../../../backend/internal/model/storage_messaging.go) | 11 | id | 4 | 追加 |
| `support_delivery_results` | [SupportDeliveryResult](../../../backend/internal/model/storage_messaging.go) | 9 | id | 2 | 追加 |
| `support_llm_budgets` | [SupportLLMBudget](../../../backend/internal/model/storage_messaging.go) | 3 | scope, model | 0 | 追加 |
| `support_manual_embeddings` | [SupportManualEmbedding](../../../backend/internal/model/storage_messaging.go) | 7 | cache_key | 0 | 追加 |
| `token_blacklist` | [TokenBlacklist](../../../backend/internal/model/token_blacklist.go) | 3 | jti | 0 | 既存対応 |
| `treatment_plans` | [TreatmentPlan](../../../backend/internal/model/hospitalization.go) | 15 | id | 3 | 既存対応 |
| `treatments` | [Treatment](../../../backend/internal/model/treatment.go) | 36 | id | 11 | 既存対応 |
| `trimming_booking_revision_options` | [TrimmingBookingRevisionOption](../../../backend/internal/model/storage_reservations.go) | 5 | clinic_id, booking_revision_id, option_id | 3 | 追加 |
| `trimming_booking_revisions` | [TrimmingBookingRevision](../../../backend/internal/model/storage_reservations.go) | 14 | id | 6 | 追加 |
| `trimming_course_included_services` | [TrimmingCourseIncludedService](../../../backend/internal/model/storage_reservations.go) | 5 | clinic_id, course_id, service_identity_id | 3 | 追加 |
| `trimming_course_types` | [TrimmingCourseType](../../../backend/internal/model/trimming_course_type.go) | 8 | id | 1 | 既存対応 |
| `trimming_courses` | [TrimmingCourse](../../../backend/internal/model/trimming_master.go) | 17 | id | 2 | 既存対応 |
| `trimming_detail_revision_lines` | [TrimmingDetailRevisionLine](../../../backend/internal/model/storage_reservations.go) | 11 | clinic_id, detail_revision_id, source_line_id | 3 | 追加 |
| `trimming_detail_revisions` | [TrimmingDetailRevision](../../../backend/internal/model/storage_reservations.go) | 8 | id | 3 | 追加 |
| `trimming_image_deletions` | [TrimmingImageDeletion](../../../backend/internal/model/storage_reservations.go) | 6 | clinic_id, object_key | 1 | 追加 |
| `trimming_option_rule_sets` | [TrimmingOptionRuleSet](../../../backend/internal/model/storage_reservations.go) | 7 | id | 2 | 追加 |
| `trimming_option_rules` | [TrimmingOptionRule](../../../backend/internal/model/storage_reservations.go) | 5 | clinic_id, rule_set_id, option_id | 3 | 追加 |
| `trimming_options` | [TrimmingOption](../../../backend/internal/model/trimming_master.go) | 14 | id | 2 | 既存対応 |
| `trimming_service_identities` | [TrimmingServiceIdentity](../../../backend/internal/model/storage_reservations.go) | 5 | id | 1 | 追加 |
| `trimming_service_lines` | [TrimmingServiceLine](../../../backend/internal/model/storage_reservations.go) | 6 | id | 4 | 追加 |
| `vaccination_certificate_artifacts` | [VaccinationCertificateArtifact](../../../backend/internal/model/storage_clinical.go) | 9 | id | 2 | 追加 |
| `vaccination_certificate_editions` | [VaccinationCertificateEdition](../../../backend/internal/model/storage_clinical.go) | 9 | id | 4 | 追加 |
| `vaccination_certificate_operations` | [VaccinationCertificateOperation](../../../backend/internal/model/storage_clinical.go) | 8 | clinic_id, operation_key | 3 | 追加 |
| `vaccination_certificate_series` | [VaccinationCertificateSeries](../../../backend/internal/model/storage_clinical.go) | 5 | id | 2 | 追加 |
| `vaccinations` | [Vaccination](../../../backend/internal/model/vaccination_record.go) | 19 | id | 7 | 既存対応 |
| `vaccines` | [Vaccine](../../../backend/internal/model/vaccine.go) | 15 | id | 3 | 既存対応 |
| `vital_records` | [VitalRecord](../../../backend/internal/model/vital.go) | 17 | id | 5 | 既存対応 |
