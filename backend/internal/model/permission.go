package model

// Resource はフロントエンドのページ識別子（権限管理用）
type Resource string

const (
	ResourceReception        Resource = "reception"
	ResourceOwners           Resource = "owners"
	ResourceReservations     Resource = "reservations"
	ResourceMedicalRecords   Resource = "medical-records"
	ResourceHospitalization  Resource = "hospitalization"
	ResourceTrimming         Resource = "trimming"
	ResourceExaminations     Resource = "examinations"
	ResourceAccounting       Resource = "accounting"
	ResourceVaccinations     Resource = "vaccinations"
	ResourceCheckups         Resource = "checkups"
	ResourceInventory        Resource = "inventory"
	ResourceEstimates        Resource = "estimates"
	ResourceShifts           Resource = "shifts"
	ResourceHospitalSettings Resource = "hospital-settings"

	// マスタ設定: 個別リソース
	ResourceMasterAnimalSpecies   Resource = "master-animal-species"
	ResourceMasterMedical         Resource = "master-medical"
	ResourceMasterReservationType Resource = "master-reservation-type"
	ResourceMasterHospitalization Resource = "master-hospitalization"
	ResourceMasterTrimming        Resource = "master-trimming"
	ResourceMasterPermission      Resource = "master-permission"
	ResourceMasterStaff           Resource = "master-staff"
	ResourceMasterInsurance       Resource = "master-insurance"
	ResourceMasterMerchandise     Resource = "master-merchandise"

	// BUG-372: 割引フィールド専用権限（飼主/治療/入院/見積/会計の全割引フィールドを保護）
	ResourceDiscount Resource = "discount"
)

// AllResources は全リソース一覧（is_system_admin=true 全権限バイパス用）
var AllResources = []Resource{
	ResourceReception,
	ResourceOwners,
	ResourceReservations,
	ResourceMedicalRecords,
	ResourceHospitalization,
	ResourceTrimming,
	ResourceExaminations,
	ResourceAccounting,
	ResourceVaccinations,
	ResourceCheckups,
	ResourceInventory,
	ResourceEstimates,
	ResourceShifts,
	ResourceHospitalSettings,
	ResourceMasterAnimalSpecies,
	ResourceMasterMedical,
	ResourceMasterReservationType,
	ResourceMasterHospitalization,
	ResourceMasterTrimming,
	ResourceMasterPermission,
	ResourceMasterStaff,
	ResourceMasterInsurance,
	ResourceMasterMerchandise,
	ResourceDiscount,
}

// IsValidResource は指定されたリソース名が有効かどうかを判定する（BUG-146）
func IsValidResource(r string) bool {
	for _, res := range AllResources {
		if string(res) == r {
			return true
		}
	}
	return false
}
