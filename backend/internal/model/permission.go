package model

// Resource はフロントエンドのページ識別子（権限管理用）
type Resource string

const (
	ResourceReception            Resource = "reception"
	ResourceOwners               Resource = "owners"
	ResourceReservations         Resource = "reservations"
	ResourceMedicalRecords       Resource = "medical-records"
	ResourceHospitalization      Resource = "hospitalization"
	ResourceTrimming             Resource = "trimming"
	ResourceExaminations         Resource = "examinations"
	ResourceExaminationUnconfirm Resource = "examination-unconfirm"
	ResourceAccounting           Resource = "accounting"
	ResourceVaccinations         Resource = "vaccinations"
	ResourceCheckups             Resource = "checkups"
	ResourceInventory            Resource = "inventory"
	ResourceEstimates            Resource = "estimates"
	ResourceShifts               Resource = "shifts"
	ResourceHospitalSettings     Resource = "hospital-settings"

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

	// FEAT-368: 集計・締め
	ResourceCashRegisterClose Resource = "cash-register-close" // レジ締め実行・履歴
	ResourceAccountingReports Resource = "accounting-reports"  // 月次売上集計（経理向け）
	ResourceClosingSettings   Resource = "closing-settings"    // 締め時間設定（管理者向け）

	// 支払方法マスタ
	ResourcePaymentMethod Resource = "master-payment-method"

	// FEAT-385: Lステップ CSV インポート・分析
	ResourceLstepCsvImport Resource = "lstep-csv-import"
	ResourceLstepAnalytics Resource = "lstep-analytics"

	// 取扱説明書（マニュアル）編集権限
	ResourceManualEdit Resource = "manual-edit"

	// #118: 会計キャンセル専用権限（ResourceAccounting "delete" から分離）
	ResourceAccountingCancel Resource = "accounting-cancel"

	// #115: 締め後編集専用権限（レジ締め済み期間の会計を特定権限で遡り編集）
	ResourceAccountingPostCloseEdit Resource = "accounting-post-close-edit"

	// lab import: 外部検査結果インポートジョブ管理
	ResourceLabImport Resource = "lab-import"

	// #239: 医院別 owner/pet を残したままの identity link（view / edit 分離・fail-closed default）
	ResourceIdentityLinks Resource = "identity-links"

	// TASK-374 / #211: 健診 package versioned import（default-deny）
	ResourceCheckupPackageImport Resource = "checkup-package-import"
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
	ResourceExaminationUnconfirm,
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
	ResourceCashRegisterClose,
	ResourceAccountingReports,
	ResourceClosingSettings,
	ResourcePaymentMethod,
	ResourceLstepCsvImport,
	ResourceLstepAnalytics,
	ResourceManualEdit,
	ResourceAccountingCancel,
	ResourceAccountingPostCloseEdit,
	ResourceLabImport,
	ResourceIdentityLinks,
	ResourceCheckupPackageImport,
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
