package medicalrecord

// BE9-2D ⑦注記: 旧 constructor の ownerRepo/petRepo は代入のみで一切未使用の dead 依存だったため
// 移動時に除去した（機能・挙動への影響なし。レビューで確認済みの上で簡素化）。
func NewMedicalRecordService(
	repo MedicalRecordRepository,
	inquiryRepo InquiryRepository,
	clinicalPlanRepo ClinicalPlanRepository,
	chiefComplaintTypeRepo ChiefComplaintTypeRepository,
	diagTypeRepo DiagnosisTypeRepository,
	diagNameRepo DiagnosisNameRepository,
	lineCustomerRepo mrLineCustomerRepo,
	reservationRepo mrReservationRepo,
	lstepDeliveryTrigger mrDeliveryTrigger,
	auditService mrAuditLogger,
	tx Transactor,
	tagSyncSvc ...mrTagSyncer,
) MedicalRecordService {
	return NewMedicalRecordServiceWithTxAudit(
		repo,
		inquiryRepo,
		clinicalPlanRepo,
		chiefComplaintTypeRepo,
		diagTypeRepo,
		diagNameRepo,
		lineCustomerRepo,
		reservationRepo,
		lstepDeliveryTrigger,
		auditService,
		nil,
		tx,
		tagSyncSvc...,
	)
}
