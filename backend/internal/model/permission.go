package model

// Resource はフロントエンドのページ識別子（権限管理用）
type Resource string

const (
	ResourceDashboard        Resource = "dashboard"
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
	ResourceMasterServiceType     Resource = "master-service-type"
	ResourceMasterHospitalization Resource = "master-hospitalization"
	ResourceMasterTrimming        Resource = "master-trimming"
	ResourceMasterPermission      Resource = "master-permission"
	ResourceMasterStaff           Resource = "master-staff"
	ResourceMasterInsurance       Resource = "master-insurance"
	ResourceMasterMerchandise     Resource = "master-merchandise"
)

// AllResources は全リソース一覧（is_system_admin=true 全権限バイパス用）
var AllResources = []Resource{
	ResourceDashboard,
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
	ResourceMasterServiceType,
	ResourceMasterHospitalization,
	ResourceMasterTrimming,
	ResourceMasterPermission,
	ResourceMasterStaff,
	ResourceMasterInsurance,
	ResourceMasterMerchandise,
}
