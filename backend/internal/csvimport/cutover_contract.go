package csvimport

import (
	"regexp"
)

const (
	cutoverManifestName        = "manifest.json"
	cutoverImportablePredicate = "mapping_status IN ('confirmed','inferred'), plus per-table CSV FK-parent eligibility guards (see scripts/lib/stage-csv-columns.mjs)"
	clinicBandSize             = int64(10_000_000)
	ownerBandOffset            = int64(300_000)
	nonOwnerBandOffset         = int64(1_000_000)
	applicationIDFloor         = int64(1_000_000_000)
	maxCutoverManifestBytes    = int64(4 << 20)
	maxCutoverCSVBytes         = int64(512 << 20)
	cutoverManifestSchema      = "animalekarte-cutover-v1"
	cutoverStageMappingSHA256  = "37df24e61efad4f55b546a376a97d7d5ab9eb39ddf12e3719198bcd001e94c77"
	cutoverCSVContractSHA256   = "4ab52e1421a6682870edf6f36504f046bd4dd3c2faf6e7a5d3634f7de0b3be0d"
)

var placeholderPattern = regexp.MustCompile(`\{\{[A-Z0-9_]+\}\}`)
var clinicCodePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)
var runIDPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]{0,62}[A-Za-z0-9])?$`)
var stageBuildIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ExpectedCutoverSource binds an operator-selected manifest digest to one
// clinic/run. The digest must come from the trusted producer run report rather
// than from the directory being imported.
type CutoverProvenanceMode string

const (
	CutoverProvenanceFormal           CutoverProvenanceMode = "formal"
	CutoverProvenanceLocalRehearsal   CutoverProvenanceMode = "local-rehearsal"
	CutoverProvenanceStagingRehearsal CutoverProvenanceMode = "staging-rehearsal"
)

type CutoverTargetBinding struct {
	Environment string
	Host        string
	Database    string
	ClinicID    int64
}

type CutoverProvenanceContract struct {
	Mode   CutoverProvenanceMode
	Target CutoverTargetBinding
}

type ExpectedCutoverSource struct {
	ManifestSHA256 string
	ClinicCode     string
	ClinicOrdinal  int64
	RunID          string
	// Provenance selects the exact accepted producer evidence. Staging
	// rehearsal is separate from local disposable rehearsal and must carry an
	// explicit target binding constructed only after operator confirmations.
	Provenance CutoverProvenanceContract
}

type CutoverIDBand struct {
	Base               int64 `json:"base"`
	EndExclusive       int64 `json:"endExclusive"`
	NonOwnerIDOffset   int64 `json:"nonOwnerIdOffset"`
	OwnerFloor         int64 `json:"ownerFloor"`
	ApplicationIDFloor int64 `json:"applicationIdFloor"`
}

type CutoverManifestTable struct {
	Table    string `json:"table"`
	File     string `json:"file"`
	RowCount int64  `json:"rowCount"`
	SHA256   string `json:"sha256"`
}

type CutoverSourceIdentity struct {
	SourceBackupSHA256    *string `json:"sourceBackupSha256"`
	SourceBackupSizeBytes *int64  `json:"sourceBackupSizeBytes"`
	BaseArchiveSHA256     *string `json:"baseArchiveSha256"`
	KNJOArchiveSHA256     *string `json:"knjoArchiveSha256"`
	Verified              bool    `json:"verified"`
}

type CutoverLayerDigests struct {
	Raw          string `json:"raw"`
	Intermediate string `json:"intermediate"`
	Stage        string `json:"stage"`
}

type CutoverLayerTimestamps struct {
	Raw          string `json:"raw"`
	Intermediate string `json:"intermediate"`
	Stage        string `json:"stage"`
}

type CutoverEvidenceDigests struct {
	BaseLoad     string `json:"baseLoad"`
	KNJORecovery string `json:"knjoRecovery"`
}

type CutoverManifest struct {
	GeneratedAt               string                 `json:"generatedAt"`
	Status                    string                 `json:"status"`
	SourceLayer               string                 `json:"sourceLayer"`
	SourceRunID               string                 `json:"sourceRunId"`
	ClinicCode                string                 `json:"clinicCode"`
	ClinicOrdinal             int64                  `json:"clinicOrdinal"`
	ClinicBandBase            int64                  `json:"clinicBandBase"`
	ClinicBandEndExclusive    int64                  `json:"clinicBandEndExclusive"`
	StageIDOffset             int64                  `json:"stageIdOffset"`
	IDBand                    CutoverIDBand          `json:"idBand"`
	OutputDir                 string                 `json:"outputDir"`
	Format                    string                 `json:"format"`
	ImportablePredicate       string                 `json:"importablePredicate"`
	PlaceholderColumns        map[string]string      `json:"placeholderColumns"`
	PlaceholderResolutionNote string                 `json:"placeholderResolutionNote"`
	ManifestSchemaVersion     string                 `json:"manifestSchemaVersion"`
	StageMappingSHA256        string                 `json:"stageMappingSha256"`
	CSVContractSHA256         string                 `json:"csvContractSha256"`
	SourceCompletenessStatus  string                 `json:"sourceCompletenessStatus"`
	SourceComplete            bool                   `json:"sourceComplete"`
	SourceProvenanceVerified  bool                   `json:"sourceProvenanceVerified"`
	SourceIdentity            CutoverSourceIdentity  `json:"sourceIdentity"`
	StageBuildID              string                 `json:"stageBuildId"`
	IncompleteSourceTables    *[]string              `json:"incompleteSourceTables"`
	HandoffEligibility        string                 `json:"handoffEligibility"`
	SourceSummarySHA256       CutoverLayerDigests    `json:"sourceSummarySha256"`
	SourceSummaryGeneratedAt  CutoverLayerTimestamps `json:"sourceSummaryGeneratedAt"`
	SourceEvidenceSHA256      CutoverEvidenceDigests `json:"sourceEvidenceSha256"`
	Tables                    []CutoverManifestTable `json:"tables"`
	// Local handoff package metadata. These fields are emitted by the old_db
	// handoff producer and remain subject to strict decoding so unknown fields fail closed.
	PackagingMode               string `json:"packagingMode"`
	SourcePackage               string `json:"sourcePackage"`
	Note                        string `json:"note"`
	SourcePackageManifestSHA256 string `json:"sourcePackageManifestSha256"`
}

type CutoverTableSpec struct {
	Name                string
	Columns             []string
	BandColumns         []string
	ForceNotNullColumns []string
}

type CutoverBundle struct {
	SourceDir  string
	Manifest   CutoverManifest
	Provenance CutoverProvenanceContract
}

// CutoverPlaceholderColumns is a copy of the exact producer-side placeholder
// inventory. Returning a fresh map prevents callers from changing the contract.
func CutoverPlaceholderColumns() map[string]string {
	return map[string]string{
		"staffs.clinic_id":            "{{CLINIC_ID}}",
		"procedures.clinic_id":        "{{CLINIC_ID}}",
		"merchandise_items.clinic_id": "{{CLINIC_ID}}",
		"owners.clinic_id":            "{{CLINIC_ID}}",
		"pets.clinic_id":              "{{CLINIC_ID}}",
		"pets.animal_species_id (fallback only, when unresolved)": "{{FALLBACK_ANIMAL_SPECIES_ID}}",
		"medical_records.clinic_id":                               "{{CLINIC_ID}}",
		"vital_records.clinic_id":                                 "{{CLINIC_ID}}",
		"appointments.clinic_id":                                  "{{CLINIC_ID}}",
		"appointments.reservation_type_id":                        "{{TRIMMING_RESERVATION_TYPE_ID}}",
		"appointment_trimming_details.clinic_id":                  "{{CLINIC_ID}}",
		"billings.clinic_id":                                      "{{CLINIC_ID}}",
		"payments.clinic_id":                                      "{{CLINIC_ID}}",
		"payments.payment_method_id (cash)":                       "{{PAYMENT_METHOD_CASH_ID}}",
		"payments.payment_method_id (credit_card)":                "{{PAYMENT_METHOD_CREDIT_CARD_ID}}",
		"payment_splits.clinic_id":                                "{{CLINIC_ID}}",
		"payment_splits.payment_method_id (cash)":                 "{{PAYMENT_METHOD_CASH_ID}}",
		"payment_splits.payment_method_id (credit_card)":          "{{PAYMENT_METHOD_CREDIT_CARD_ID}}",
		"estimates.clinic_id":                                     "{{CLINIC_ID}}",
		"exams.clinic_id":                                         "{{CLINIC_ID}}",
		"exams.exam_type_id":                                      "{{FALLBACK_EXAM_TYPE_ID}}",
		"vaccines.clinic_id":                                      "{{CLINIC_ID}}",
		"vaccinations.clinic_id":                                  "{{CLINIC_ID}}",
	}
}

// CutoverTableSpecs returns the immutable, parent-before-child F6 CSV contract.
func CutoverTableSpecs() []CutoverTableSpec {
	specs := []CutoverTableSpec{
		{"staffs", []string{"id", "clinic_id", "name", "license_number", "is_active", "reservation_visible"}, []string{"id"}, []string{"name", "license_number"}},
		{"procedures", []string{"id", "clinic_id", "name", "price", "is_active", "description", "duration", "anesthesia", "parent_id", "tax_type", "tax_rate", "sort_order", "is_surgery"}, []string{"id", "parent_id"}, []string{"name", "description"}},
		{"merchandise_items", []string{"id", "clinic_id", "name", "category", "unit_price", "tax_type", "tax_rate", "is_active", "sort_order"}, []string{"id"}, []string{"name"}},
		{"owners", []string{"id", "clinic_id", "name", "name_kana", "birth_date", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks", "is_dangerous", "discount_rate", "membership_type", "dm_preference"}, []string{"id"}, []string{"name", "name_kana", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks"}},
		{"pets", []string{"id", "clinic_id", "owner_id", "pet_number", "name", "name_kana", "animal_species_id", "gender", "status", "birth_date", "breed", "color", "weight", "neutered_date", "food", "remarks", "deceased_at"}, []string{"id", "owner_id"}, []string{"pet_number", "name", "name_kana", "breed", "color", "food", "remarks"}},
		{"medical_records", []string{"id", "clinic_id", "record_no", "date", "owner_id", "pet_id", "status", "visit_type", "doctor_id", "entered_by"}, []string{"id", "owner_id", "pet_id", "doctor_id", "entered_by"}, []string{"record_no"}},
		{"inquiries", []string{"id", "medical_record_id", "chief_complaint", "owner_observations", "history", "notes", "allergy_info", "current_medications", "staff_id"}, []string{"id", "medical_record_id", "staff_id"}, []string{"chief_complaint", "owner_observations", "history", "notes", "allergy_info", "current_medications"}},
		{"clinical_plans", []string{"id", "medical_record_id", "physical_exam", "diagnosis_details", "treatment_policy"}, []string{"id", "medical_record_id"}, []string{"physical_exam", "diagnosis_details", "treatment_policy"}},
		{"vital_records", []string{"id", "clinic_id", "medical_record_id", "pet_id", "recorded_at", "temperature", "weight", "weight_unit", "heart_rate", "respiration_rate", "staff_id", "notes"}, []string{"id", "medical_record_id", "pet_id", "staff_id"}, []string{"notes"}},
		{"appointments", []string{"id", "clinic_id", "start_time", "end_time", "owner_id", "pet_id", "visit_type", "reservation_type_id", "doctor_id", "status", "source"}, []string{"id", "owner_id", "pet_id", "doctor_id"}, nil},
		{"appointment_trimming_details", []string{"id", "clinic_id", "appointment_id", "remarks"}, []string{"id", "appointment_id"}, []string{"remarks"}},
		{"billings", []string{"id", "clinic_id", "medical_record_id", "owner_id", "pet_id", "total_amount", "status", "scheduled_date", "completed_at"}, []string{"id", "medical_record_id", "owner_id", "pet_id"}, nil},
		{"billing_items", []string{"id", "billing_id", "category", "name", "unit_price", "quantity", "tax_type", "is_insurance_applicable", "sort_order"}, []string{"id", "billing_id"}, []string{"name"}},
		{"payments", []string{"id", "clinic_id", "billing_id", "subtotal", "tax_total", "total_amount", "insurance_name", "insurance_ratio", "insurance_amount", "discount_amount", "billing_amount", "received_amount", "change_amount", "method", "payment_method_id", "paid_by", "created_at"}, []string{"id", "billing_id", "paid_by"}, []string{"insurance_name"}},
		{"payment_splits", []string{"id", "clinic_id", "billing_id", "method", "payment_method_id", "amount", "received_amount", "change_amount", "paid_by", "created_at"}, []string{"id", "billing_id", "paid_by"}, nil},
		{"estimates", []string{"id", "clinic_id", "estimate_no", "medical_record_id", "title", "owner_id", "status", "subtotal", "tax_total", "total_amount", "insurance_amount", "discount_amount", "valid_until", "comment", "notes", "created_by", "created_at"}, []string{"id", "medical_record_id", "owner_id", "created_by"}, []string{"estimate_no", "title", "comment", "notes"}},
		{"estimate_items", []string{"id", "estimate_id", "name", "category", "unit_price", "quantity", "tax_type", "tax_rate", "discount_rate", "discount_amount", "is_insurance_applicable", "consultation_id", "procedure_id", "medicine_id", "merchandise_item_id", "sort_order"}, []string{"id", "estimate_id", "consultation_id", "procedure_id", "medicine_id", "merchandise_item_id"}, []string{"name"}},
		{"exams", []string{"id", "clinic_id", "medical_record_id", "pet_id", "date", "exam_type_id", "result_summary"}, []string{"id", "medical_record_id", "pet_id"}, []string{"result_summary"}},
		{"exam_results", []string{"id", "exam_id", "name", "inspection_value", "normal_value", "sort_order"}, []string{"id", "exam_id"}, []string{"name", "inspection_value", "normal_value"}},
		{"vaccines", []string{"id", "clinic_id", "name", "price", "is_active", "description", "species", "interval", "inventory_id", "parent_id", "sort_order"}, []string{"id"}, []string{"name", "description", "interval"}},
		{"vaccinations", []string{"id", "clinic_id", "medical_record_id", "pet_id", "vaccine_id", "date", "next_date", "next_schedule_type", "doctor_id", "supplemental", "lot1", "lot2", "lot3", "lot4", "remarks"}, []string{"id", "medical_record_id", "pet_id", "vaccine_id", "doctor_id"}, []string{"supplemental", "lot1", "lot2", "lot3", "lot4", "remarks"}},
	}
	return cloneCutoverSpecs(specs)
}

func cloneCutoverSpecs(specs []CutoverTableSpec) []CutoverTableSpec {
	cloned := make([]CutoverTableSpec, len(specs))
	for i, spec := range specs {
		cloned[i] = CutoverTableSpec{
			Name:                spec.Name,
			Columns:             append([]string(nil), spec.Columns...),
			BandColumns:         append([]string(nil), spec.BandColumns...),
			ForceNotNullColumns: append([]string(nil), spec.ForceNotNullColumns...),
		}
	}
	return cloned
}

// Stable non-PHI operator error references for cutover audit trails (Issue #250).
// Errors may include table/column/CSV line coordinates and these refs only —
// never owner/pet names, addresses, phones, clinical text, or CSV cell values.
const (
	CutoverRefBandOccupied    = "CUTOVER_REF_BAND_OCCUPIED"
	CutoverRefClinicIsolation = "CUTOVER_REF_CLINIC_ISOLATION"
	CutoverRefRowCount        = "CUTOVER_REF_ROW_COUNT"
)

// CutoverIsolationClinicID is the verify path for tables that carry clinic_id.
const CutoverIsolationClinicID = "clinic_id_column"

// CutoverIsolationParentFK is the verify path for child tables without clinic_id:
// isolation is enforced by the clinic ID band plus validated parent FKs.
const CutoverIsolationParentFK = "id_band_and_parent_fk"

// CutoverMappingEntry is the non-PHI source→target inventory for one formal
// cutover table. It intentionally excludes source cell values and personal data.
type CutoverMappingEntry struct {
	TargetTable    string
	CSVFile        string
	ColumnCount    int
	BandColumns    []string
	HasClinicID    bool
	IsolationCheck string
	ConsumerStatus string
}

// CutoverMappingCoverage returns the immutable 21-table mapping inventory used
// by Issue #250 acceptance (source CSV → target table, isolation mode).
// Personal business owners remain a USER ops assignment outside this package.
func CutoverMappingCoverage() []CutoverMappingEntry {
	specs := CutoverTableSpecs()
	coverage := make([]CutoverMappingEntry, 0, len(specs))
	for _, spec := range specs {
		hasClinicID := hasColumn(spec.Columns, "clinic_id")
		isolation := CutoverIsolationParentFK
		if hasClinicID {
			isolation = CutoverIsolationClinicID
		}
		coverage = append(coverage, CutoverMappingEntry{
			TargetTable:    spec.Name,
			CSVFile:        spec.Name + ".csv",
			ColumnCount:    len(spec.Columns),
			BandColumns:    append([]string(nil), spec.BandColumns...),
			HasClinicID:    hasClinicID,
			IsolationCheck: isolation,
			ConsumerStatus: "formal_cutover_v1",
		})
	}
	return coverage
}
