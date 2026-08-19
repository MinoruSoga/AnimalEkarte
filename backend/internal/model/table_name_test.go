package model

import "testing"

// TestTableNames pins every GORM TableName() override to its expected snake_case
// table name. GORM relies on these being exact; a silent drift (typo, rename)
// would misroute queries to a non-existent or wrong table.
func TestTableNames(t *testing.T) {
	tests := []struct {
		name  string
		model interface{ TableName() string }
		want  string
	}{
		{"AuditLog", AuditLog{}, "audit_logs"},
		{"Campaign", Campaign{}, "campaigns"},
		{"CampaignTargetCategory", CampaignTargetCategory{}, "campaign_target_categories"},
		{"CampaignTargetItem", CampaignTargetItem{}, "campaign_target_items"},
		{"CheckupTypeField", CheckupTypeField{}, "checkup_type_fields"},
		{"CheckupFieldResult", CheckupFieldResult{}, "checkup_field_results"},
		{"ClinicIntegration", ClinicIntegration{}, "clinic_integrations"},
		{"LabImportJob", LabImportJob{}, "lab_import_jobs"},
		{"LabImportEvent", LabImportEvent{}, "lab_import_events"},
		{"LabDevice", LabDevice{}, "lab_devices"},
		{"LabDeviceItemMaster", LabDeviceItemMaster{}, "lab_device_item_masters"},
		{"LabImportJobItem", LabImportJobItem{}, "lab_import_job_items"},
		{"LabDeviceWait", LabDeviceWait{}, "lab_device_waits"},
		{"LabDeviceStationSettings", LabDeviceStationSettings{}, "lab_device_station_settings"},
		{"LineLinkToken", LineLinkToken{}, "line_link_tokens"},
		{"LineSendLog", LineSendLog{}, "line_send_logs"},
		{"LstepAutoManagedPrefix", LstepAutoManagedPrefix{}, "lstep_auto_managed_prefixes"},
		{"LstepConditionTagMapping", LstepConditionTagMapping{}, "lstep_condition_tag_mappings"},
		{"LstepCsvImport", LstepCsvImport{}, "lstep_csv_imports"},
		{"LstepDeliveryTriggerLog", LstepDeliveryTriggerLog{}, "lstep_delivery_trigger_log"},
		{"LstepFriendAttributeSnapshot", LstepFriendAttributeSnapshot{}, "lstep_friend_attribute_snapshots"},
		{"LstepSendPurposeTagPrefix", LstepSendPurposeTagPrefix{}, "lstep_send_purpose_tag_prefixes"},
		{"LstepSettings", LstepSettings{}, "lstep_settings"},
		{"LstepSyncErrorCounter", LstepSyncErrorCounter{}, "lstep_sync_error_counters"},
		{"LstepTagCache", LstepTagCache{}, "lstep_tag_cache"},
		{"LstepTagCodeMapping", LstepTagCodeMapping{}, "lstep_tag_code_mappings"},
		{"LstepTriggerPriority", LstepTriggerPriority{}, "lstep_trigger_priorities"},
		{"ManualArticle", ManualArticle{}, "manual_articles"},
		{"ManualArticleVersion", ManualArticleVersion{}, "manual_article_versions"},
		{"MedicalRecordAddendum", MedicalRecordAddendum{}, "medical_record_addenda"},
		{"PasswordResetToken", PasswordResetToken{}, "password_reset_tokens"},
		{"PetChronicCondition", PetChronicCondition{}, "pet_chronic_conditions"},
		{"PetOwner", PetOwner{}, "pet_owners"},
		{"OwnerIdentityGroup", OwnerIdentityGroup{}, "owner_identity_groups"},
		{"OwnerIdentityGroupMember", OwnerIdentityGroupMember{}, "owner_identity_group_members"},
		{"PetIdentityGroup", PetIdentityGroup{}, "pet_identity_groups"},
		{"PetIdentityGroupMember", PetIdentityGroupMember{}, "pet_identity_group_members"},
		{"SharedFile", SharedFile{}, "shared_files"},
		{"TokenBlacklist", TokenBlacklist{}, "token_blacklist"},
		{"TrimmingCourseType", TrimmingCourseType{}, "trimming_course_types"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.TableName(); got != tt.want {
				t.Errorf("%s.TableName() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
