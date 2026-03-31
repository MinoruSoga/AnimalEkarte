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
	ResourceMaster           Resource = "master"
	ResourceHospitalSettings Resource = "hospital-settings"
)

// AllResources は全リソース一覧（system_admin/clinic_admin 全権限バイパス用）
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
	ResourceMaster,
	ResourceHospitalSettings,
}
