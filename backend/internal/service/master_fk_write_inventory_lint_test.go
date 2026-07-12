package service

// Write-side P3.1 review-coverage gate — every service method that receives a
// REQUEST-DERIVED clinic-scoped master foreign key (FK) through one of its input
// parameters (a DTO struct field, possibly nested in a sub-struct / slice / embedded
// struct) must appear in the maintained masterFKWriteAllowlist with an explicit status.
// A NEW such write that is not on the allowlist FAILS this gate, so it cannot merge
// without a human reviewing it (and, where it persists, adding a runtime isolation test).
//
// ───────────────────────────────────────────────────────────────────────────────
// SCOPE / ROLE BOUNDARY — read this before extending the gate:
//
//	This gate does NOT verify that a FindByID(ctx, clinicID, …) ownership guard exists
//	or actually works. Proving "the request-supplied master FK was checked against the
//	caller's clinic before persistence" requires interprocedural taint analysis across
//	handler→service→repository, which go/ast cannot do reliably. #124 (f4e7b7a7) is the
//	standing counterexample: the parent exam_type_id was validated while a NESTED
//	exam_type_field_id was not. A static rule that "looks 80% validated" is the exact
//	"felt-validated" failure mode that let #124 regress, so we deliberately do not build
//	one. Correctness is enforced at RUNTIME by cross_tenant_master_fk_write_test.go and
//	the *_clinic_isolation_test.go guards. THIS gate enforces only REVIEW COVERAGE:
//	it forces every master-FK write onto a curated allowlist with a status. It is the
//	write-side analogue of repository/preload_clinic_scope_lint_test.go (read side).
//
// Residual gaps this gate does NOT cover (documented in repository/CLAUDE.md P3.1):
//  1. Correctness of the ownership guard (see above) — runtime tests are the source of truth.
//  2. A master FK arriving as a BARE scalar parameter (e.g. `medicineID uint64`) rather
//     than via a DTO struct field. Detection keys on DTO struct fields by design (the
//     request-binding architecture binds master FKs into XxxInput DTOs). Known bare-param
//     sites (last verified by grep, not automated — re-check when new Set*IDs/Link* methods
//     are added): medicineDoseParamService.Upsert, staffService.Set{Excluded,Capable}ReservationTypeIDs,
//     staffService.SetPermissionGroupIDs (PermissionGroup; mitigated — repo UpdateStaffGroups
//     does a clinic-scoped IN-check), reservationTypeService.LinkOccupation (mitigated — FindByID).
//  3. A master FK propagated through a non-`model.` cross-package struct parameter. Today
//     there are none; knownSafeParamQualifiers pins this so a new external param qualifier
//     fails closed and is forced into review.
//
// Technique reused verbatim from repository/preload_clinic_scope_lint_test.go and
// model/audit_taxonomy_exhaustiveness_test.go: go:embed the package source (correct under
// -trimpath), parse with go/parser+go/ast, curate the field set + allowlist, and pin
// everything with good/bad fixtures, floor guards, bidirectional exhaustiveness, and a
// negative self-check so the gate can never pass vacuously.

import (
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strings"
	"testing"
)

// serviceSourceFS embeds every .go file in this package directory at compile time.
// The glob also matches *_test.go (incl. this file); those are skipped at runtime in
// analyzeServicePackage so the gate never inspects its own fixtures or other test helpers.
//
//go:embed *.go
var serviceSourceFS embed.FS

// clinicScopedMasterFKField maps a Go struct field name to the clinic-scoped master model
// it references. When a service method's input DTO carries one of these fields (with an
// ID-shaped type), the method is a request-derived clinic-scoped master-FK write and must
// be on the allowlist. Symmetric with repository.clinicScopedMasterAssoc (read side).
//
// Every target master was confirmed to carry a clinic_id column in internal/model (see
// the model files named below). animal_species / manual_article are GLOBAL (no clinic_id)
// and are intentionally absent.
//
// GENERIC field names (ParentID, GroupID, CourseID, OptionIDs, ExcludedTypeIDs) are
// context-resolved: in the CURRENT service write DTOs they ONLY ever denote the listed
// clinic-scoped master (verified by grep). If one is later reused for a non-master FK,
// that new write fails closed (forced into review) — the same fail-closed tradeoff the
// read lint accepts for Parent/Group/Course/Options.
var clinicScopedMasterFKField = map[string]string{
	// ── specific, unambiguous names ──
	"MedicineID":            "Medicine",
	"VaccineID":             "Vaccine",
	"ProcedureID":           "Procedure",
	"ConsultationID":        "Consultation",
	"ExamTypeID":            "ExaminationType",
	"ExamTypeFieldID":       "ExamTypeField (sub-master of ExaminationType, #124)",
	"CheckupTypeFieldID":    "CheckupTypeField (sub-master of CheckupType, #211)",
	"ReservationTypeID":     "ReservationType",
	"TrimmingCourseID":      "TrimmingCourse",
	"TrimmingOptionID":      "TrimmingOption",
	"TrimmingOptionIDs":     "TrimmingOption",
	"CageID":                "Cage",
	"CheckupTypeID":         "CheckupType",
	"DiagnosisTypeID":       "DiagnosisType",
	"DiagnosisNameID":       "DiagnosisName",
	"Diagnosis1CategoryID":  "DiagnosisType",
	"Diagnosis1NameID":      "DiagnosisName",
	"Diagnosis2CategoryID":  "DiagnosisType",
	"Diagnosis2NameID":      "DiagnosisName",
	"InsuranceID":           "Insurance",
	"OccupationID":          "Occupation",
	"ChiefComplaintTypeID":  "ChiefComplaintType",
	"HospitalizationPlanID": "HospitalizationPlan",
	"MerchandiseItemID":     "MerchandiseItem",
	"CourseTypeID":          "TrimmingCourseType",
	"PaymentMethodID":       "PaymentMethodMaster",
	"InventoryID":           "InventoryItem (clinic-scoped stock catalog; treatment/medicine FK)",
	// ── generic, context-resolved (verified master-only in current write DTOs) ──
	"ParentID":        "self-ref master (checkup_type/consultation/exam_type/medicine/procedure/reservation_type/vaccine)",
	"GroupID":         "ReservationTypeGroup",
	"CourseID":        "TrimmingCourse (alias)",
	"OptionIDs":       "TrimmingOption (alias)",
	"ExcludedTypeIDs": "ReservationType (reservation_staff exclusions)",
	"TargetItemIDs":   "MerchandiseItem (campaign_target_items.merchandise_item_id)",
}

// knownSafeParamQualifiers are package qualifiers (the `pkg` in `pkg.Type`) that a service
// method parameter may use without the gate being able to introspect the type for master
// FKs. None of these resolve to a service-local DTO that could carry a clinic-scoped master
// FK field. A NEW qualifier appearing in a method parameter trips
// TestMasterFKWriteInventory_NoUnknownCrossPackageParam — forcing review of whether that
// external type carries a master FK (closing the cross-package false-negative gap).
var knownSafeParamQualifiers = map[string]struct{}{
	"context": {}, // context.Context
	"time":    {}, // time.Time
	"model":   {}, // model.* — domain models; service-local DTOs (which carry master FKs) are not here
	"uuid":    {}, // github.com/google/uuid
	"io":      {}, // io.Reader (file/stream import)
	// repository.* — read-filter/query structs (e.g. MedicalRecordListFilters), not persistence
	// write DTOs. Fields checked against clinicScopedMasterFKField: PetID/OwnerID/StartDate/
	// EndDate/Status/Search carry no master FK; DoctorID references Staff (explicitly exempt from
	// clinic-scope write checks — multi-clinic assignment, see repository/CLAUDE.md P3.1); and
	// AnimalSpeciesID is a global (non clinic-scoped) master. Re-review if repository.* gains a
	// write-path DTO carrying a clinic-scoped master FK field.
	"repository": {},
}

// masterFKWriteStatus records WHY a master-FK write is on the allowlist. The gate does not
// verify these; they are the human review record.
type masterFKWriteStatus string

const (
	// statusGuarded: a runtime isolation test (cross_tenant_master_fk_write_test.go or a
	// *_clinic_isolation_test.go) proves a cross-clinic master FK is rejected.
	statusGuarded masterFKWriteStatus = "guarded"
	// statusKnownUnguarded: reviewed; NO dedicated isolation test confirms ownership
	// rejection yet. Residual P1 risk tracked here (the candidate work list for adding tests).
	statusKnownUnguarded masterFKWriteStatus = "known-unguarded"
	// statusExempt: not a persistence write of the FK (preview/read/validate-only), or the
	// FK is validated by a different documented mechanism. Reason must justify it.
	// Currently NO allowlist entry uses this — every enumerated method persists its FK.
	// Reserved for a future preview/read-only endpoint that receives a master FK DTO.
	statusExempt masterFKWriteStatus = "exempt"
)

// masterFKWriteEntry is one allowlist row. key = "receiverType.Method" (unique per package).
// masterFKs is the SORTED set of clinic-scoped master FK field names the method's params
// transitively carry; the gate requires it to match the AST-computed set EXACTLY, so adding
// a new master FK to an already-listed DTO (the #124 nested-field shape) forces re-review.
type masterFKWriteEntry struct {
	key       string
	status    masterFKWriteStatus
	masterFKs []string
	reason    string
}

// masterFKWriteAllowlist records every enumerated master-FK write with a reviewed status.
// Sorted by key for review diffs. masterFKs MUST equal the AST-computed sorted set exactly.
//
//	guarded         = ownership validation (FindByID(ctx, clinicID, …) or equivalent) covers
//	                  ALL master FKs the method carries. Verified in code (commits 03bf1cb5 /
//	                  f4e7b7a7 and direct inspection).
//	known-unguarded = at least one master FK is persisted WITHOUT an ownership check (incl.
//	                  partially-guarded methods). Residual P1 — the work list for runtime tests.
//	exempt          = does not persist the FK / validated by a documented alternate mechanism.
//
// REMINDER: status is a human review record. The gate does NOT verify it (see file header).
var masterFKWriteAllowlist = []masterFKWriteEntry{
	// ── guarded (FindByID ownership check covers every master FK; most have runtime isolation tests) ──
	{"accountingService.Update", statusGuarded, []string{"PaymentMethodID"}, "resolvePaymentMethodMasterID の mismatch 拒否ロジックで validated; test: TestAccountingService_Update_RejectsForeignPaymentMethodID"},
	{"campaignService.Create", statusGuarded, []string{"TargetItemIDs"}, "campaign_service.go: validateOwnedMerchandiseItemIDs loops merchandiseItemRepo.FindByID(ctx, clinicID, id) over every TargetItemIDs entry (X-5); test: TestCampaignService_Create_RejectsCrossClinicTargetItemFK"},
	{"campaignService.Update", statusGuarded, []string{"TargetItemIDs"}, "as Create — validateOwnedMerchandiseItemIDs guards *input.TargetItemIDs before ReplaceTargets (X-5); test: TestCampaignService_Update_RejectsCrossClinicTargetItemFK"},
	{"carePlanItemService.Create", statusGuarded, []string{"HospitalizationPlanID", "MedicineID", "ProcedureID"}, "care_plan_item_service.go: validateMasterFKs now covers all three — medicine/procedure (pre-existing) plus hospPlanRepo.FindByID(ctx, clinicID, HospitalizationPlanID) (X-14); test: TestCarePlanItemService_Create_RejectsCrossClinicHospitalizationPlanFK"},
	{"carePlanItemService.Update", statusGuarded, []string{"HospitalizationPlanID", "MedicineID", "ProcedureID"}, "as Create — validateMasterFKs guards *input.HospitalizationPlanID before persist (X-14); test: TestCarePlanItemService_Update_RejectsCrossClinicHospitalizationPlanFK"},
	{"checkupFieldResultService.ReplaceForCheckup", statusGuarded, []string{"CheckupTypeFieldID"}, "checkup_field_result_service.go: each CheckupTypeFieldID validated within the owned checkup_type's fields (#124 同型, #211); test in checkup_field_result_service_test.go"},
	{"checkupService.Create", statusGuarded, []string{"CheckupTypeID"}, "checkup_service.go: checkupTypeRepo.FindByID(ctx, clinicID, CheckupTypeID); test in cross_tenant_master_fk_write_test.go"},
	{"checkupService.Update", statusGuarded, []string{"CheckupTypeID"}, "checkup_service.go:223 checkupTypeRepo.FindByID(ctx, clinicID, *CheckupTypeID)"},
	{"clinicalPlanService.Update", statusGuarded, []string{"Diagnosis2CategoryID", "Diagnosis2NameID", "DiagnosisNameID", "DiagnosisTypeID"}, "validateDiagnosisFKs FindByID for all four slots (03bf1cb5); test present"},
	{"diagnosisNameService.Create", statusGuarded, []string{"DiagnosisTypeID"}, "diagnosis_service.go:295 typeRepo.FindByID(ctx, clinicID, DiagnosisTypeID) (#020)"},
	{"diagnosisNameService.Update", statusGuarded, []string{"DiagnosisTypeID"}, "diagnosis_service.go:329 typeRepo.FindByID(ctx, clinicID, *DiagnosisTypeID)"},
	{"examinationService.Create", statusGuarded, []string{"ExamTypeID"}, "examTypeRepo.FindByID(ctx, clinicID, ExamTypeID) (#124, 03bf1cb5); test present"},
	{"examinationService.ReplaceItems", statusGuarded, []string{"ExamTypeFieldID"}, "each ExamTypeFieldID validated within the owned exam_type's items (#124, f4e7b7a7); test present"},
	{"examinationService.Update", statusGuarded, []string{"ExamTypeID"}, "examTypeRepo.FindByID when non-nil (03bf1cb5); test present"},
	{"hospitalizationService.Create", statusGuarded, []string{"CageID"}, "hospitalization_service.go: repos.Cage.FindByID(ctx, clinicID, CageID) when non-nil (X-14); test: TestHospitalizationService_Create_RejectsCrossClinicCageFK"},
	{"hospitalizationService.Update", statusGuarded, []string{"CageID"}, "as Create — repos.Cage.FindByID(ctx, clinicID, *CageID) when non-nil (X-14); test: TestHospitalizationService_Update_RejectsCrossClinicCageFK"},
	{"ownerService.CreateWithPets", statusGuarded, []string{"InsuranceID"}, "owner_service_core.go: validateOwnerPetsInsuranceOwnership loops insuranceRepo.FindByID(ctx, clinicID, id) over every nested Pets[i].InsuranceID before repo.CreateWithPets (X-14 batch U5); test: TestOwnerService_CreateWithPets_RejectsCrossClinicInsuranceID"},
	{"petService.Create", statusGuarded, []string{"InsuranceID"}, "pet_service.go: insuranceRepo.FindByID(ctx, clinicID, InsuranceID) when non-nil — guard pre-existing, dedicated isolation test added (X-14 batch U5); test: TestPetService_Create_RejectsCrossClinicInsuranceID"},
	{"petService.Update", statusGuarded, []string{"InsuranceID"}, "pet_service.go: insuranceRepo.FindByID(ctx, clinicID, **InsuranceID) when non-nil and non-NULL — guard pre-existing, dedicated isolation test added (X-14 batch U5); test: TestPetService_Update_RejectsCrossClinicInsuranceID"},
	{"trimmingCourseService.Create", statusGuarded, []string{"CourseTypeID"}, "trimming_course_service.go:118 courseTypeRepo.FindByID when non-nil (#73)"},
	{"vaccinationService.Create", statusGuarded, []string{"VaccineID"}, "vaccineRepo.FindByID(ctx, clinicID, VaccineID) (#125, 03bf1cb5); test present"},
	{"vaccinationService.Update", statusGuarded, []string{"VaccineID"}, "vaccineRepo.FindByID when non-nil (03bf1cb5); test present"},

	// ── known-unguarded: at least one master FK persisted without an ownership check (residual P1) ──
	{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "PARTIAL (X-4): TrimmingCourseID/TrimmingOptionID now guarded via trimmingCourseRepo/trimmingOptionRepo.FindByID(ctx, clinicID, id) before persist; test: TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK. MerchandiseItemID remains unguarded but is a DEAD field for this write path — CreateItem never assigns input.MerchandiseItemID onto model.BillingItem (billing_item_service.go, item struct literal), so it carries no actual cross-tenant persistence risk today; out of scope for X-4."},
	{"checkupTypeService.Create", statusGuarded, []string{"ParentID"}, "checkup_type_service.go: validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestCheckupTypeService_Create_RejectsCrossClinicParentFK"},
	{"checkupTypeService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestCheckupTypeService_Update_RejectsCrossClinicParentFK"},
	{"consultationService.Create", statusGuarded, []string{"ParentID"}, "consultation_service.go: validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestConsultationService_Create_RejectsCrossClinicParentFK"},
	{"consultationService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestConsultationService_Update_RejectsCrossClinicParentFK"},
	{"examTypeService.Create", statusGuarded, []string{"ParentID"}, "exam_type_service.go: validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestExamTypeService_Create_RejectsCrossClinicParentFK"},
	{"examTypeService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestExamTypeService_Update_RejectsCrossClinicParentFK"},
	{"inquiryService.Save", statusGuarded, []string{"ChiefComplaintTypeID"}, "inquiry_service.go: chiefComplaintTypeRepo.FindByID(ctx, input.ClinicID, *ChiefComplaintTypeID) before persist (X-14 batch U4); test: TestInquiryService_Save_RejectsCrossClinicChiefComplaintType"},
	{"labImportExaminationService.PersistBatch", statusGuarded, []string{"ExamTypeID"}, "PersistBatch delegates each row to PersistExam, which now guards ExamTypeID (X-14 batch U3); test: TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamType"},
	{"labImportExaminationService.PersistExam", statusGuarded, []string{"ExamTypeID"}, "lab_import_examination_service.go: examTypeRepo.FindByID(ctx, clinicID, ExamTypeID) before dup-check/create (#124 同型, X-14 batch U3); test: TestLabImportExaminationService_PersistExam_RejectsCrossClinicExamType"},
	{"labResultImportService.Commit", statusGuarded, []string{"ExamTypeID"}, "Commit delegates to labImportExaminationService.PersistBatch/PersistExam, which now guards ExamTypeID (X-14 batch U3); test: TestLabResultImportService_Commit_RejectsCrossClinicExamType"},
	{"liffService.CreateReservation", statusGuarded, []string{"ReservationTypeID", "TrimmingCourseID", "TrimmingOptionIDs"}, "liffService.CreateReservation delegates fully to reservationValidators.ValidateAndCreate, which now guards all three FKs before tx (X-14 U6a); test: TestLiffService_CreateReservation_RejectsCrossClinicTrimmingFK"},
	{"medicalRecordService.CreateSubRecords", statusGuarded, []string{"ChiefComplaintTypeID", "Diagnosis1CategoryID", "Diagnosis1NameID", "Diagnosis2CategoryID", "Diagnosis2NameID"}, "medical_record_subrecords.go: chiefComplaintTypeRepo.FindByID before inquiry upsert; validateCreateSubRecordDiagnosisFKs (validateDiagnosisFKs-equivalent, unexported local helper) guards all four diagnosis FKs before clinicalPlanRepo.Update (X-14 batch U4); best-effort — failure skips the write (Warn), does not error. test: TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicChiefComplaintType, TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicDiagnosisFK"},
	{"medicineService.Create", statusGuarded, []string{"InventoryID", "ParentID"}, "medicine_service.go: validateParentOwnership (self-ref repo.FindByID) + validateInventoryOwnership (inventoryRepo.FindByID) before persist (X-14 batch U2); test: TestMedicineService_Create_RejectsCrossClinicParentFK, TestMedicineService_Create_RejectsCrossClinicInventoryFK"},
	{"medicineService.Update", statusGuarded, []string{"InventoryID", "ParentID"}, "as Create — validateParentOwnership/validateInventoryOwnership guard *input.ParentID/*input.InventoryID before repo.Update (X-14 batch U2); test: TestMedicineService_Update_RejectsCrossClinicParentFK, TestMedicineService_Update_RejectsCrossClinicInventoryFK"},
	{"procedureService.Create", statusGuarded, []string{"ParentID"}, "procedure_service.go: validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestProcedureService_Create_RejectsCrossClinicParentFK"},
	{"procedureService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestProcedureService_Update_RejectsCrossClinicParentFK"},
	{"reservationAdminService.Create", statusGuarded, []string{"ReservationTypeID"}, "appointment_admin_service.go:124 checkReservationTypeCapacity(ctx, s.resRepo, s.typeRepo, ...) unconditionally calls typeRepo.FindByID(ctx, clinicID, ReservationTypeID) before persist (X-14 U6b); test: TestReservationAdminService_Create_RejectsCrossClinicReservationType"},
	{"reservationService.Create", statusGuarded, []string{"ReservationTypeID"}, "reservation_service.go: added unconditional s.typeRepo.FindByID(ctx, input.ClinicID, ReservationTypeID) before persist — closes the shortcut-route hole where reception/exam_room/record_shortcut routes (or advanced statuses) set enforceBookingConstraints=false and previously skipped the FindByID embedded in checkReservationTypeCapacity (X-14 U6b); test: TestReservationService_Create_RejectsCrossClinicReservationType"},
	{"reservationService.Update", statusGuarded, []string{"ReservationTypeID"}, "reservation_service.go:389 updateWithConflictCheck unconditionally calls checkReservationTypeCapacity → typeRepo.FindByID(ctx, clinicID, resolvedReservationTypeID) whenever ReservationTypeID changes (needsConflictCheck gate); pre-existing guard, dedicated isolation test added (X-14 U6b); test: TestReservationService_Update_RejectsCrossClinicReservationType"},
	{"reservationStaffService.Create", statusGuarded, []string{"ExcludedTypeIDs"}, "reservation_staff_repository.go:227-238 UpdateExcludedReservationTypes counts clinic_id=? AND id IN ? AND deleted_at IS NULL before DELETE/INSERT and rejects on mismatch; pre-existing repo-level guard (X-14 U6b); test: TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation"},
	{"reservationStaffService.Update", statusGuarded, []string{"ExcludedTypeIDs"}, "as Create — same repo-level Count guard in UpdateExcludedReservationTypes (X-14 U6b); test: TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation"},
	{"reservationTypeService.Create", statusGuarded, []string{"GroupID", "ParentID"}, "ParentID validated via validateReservationTypeParent (pre-existing); GroupID now validated via new validateReservationTypeGroup → groupRepo.FindByID(ctx, clinicID, *GroupID) (X-14 U6b, groupRepo DI added to NewReservationTypeService); test: TestReservationTypeService_Create_RejectsCrossClinicGroupID"},
	{"reservationTypeService.Update", statusGuarded, []string{"GroupID", "ParentID"}, "as Create — validateReservationTypeGroup guards *input.GroupID before repo.Update; validateReservationTypeParent guards ParentID (X-14 U6b); test: TestReservationTypeService_Update_RejectsCrossClinicGroupID, TestReservationTypeService_Update_RejectsCrossClinicParentID"},
	{"reservationValidators.ValidateAndCreate", statusGuarded, []string{"ReservationTypeID", "TrimmingCourseID", "TrimmingOptionIDs"}, "reservation_validators.go: typeRepo/trimmingCourseRepo/trimmingOptionRepo.FindByID(ctx, clinicID, id) for all three FKs before tx (hard fail, no orphan appointment) (X-14 U6a); test: TestReservationValidators_ValidateAndCreate_RejectsCrossClinicReservationType, TestReservationValidators_ValidateAndCreate_RejectsCrossClinicTrimmingFK"},
	{"staffService.Create", statusGuarded, []string{"OccupationID"}, "staff_service_core.go: validateOccupationOwnership occupationRepo.FindByID(ctx, clinicID, *OccupationID) before persist (X-14 batch U7, occupationRepo now a mandatory NewStaffService dependency); test: TestStaffService_Create_RejectsCrossClinicOccupationID"},
	{"staffService.CreateWithAccount", statusGuarded, []string{"OccupationID"}, "as Create — validateOccupationOwnership guards *input.OccupationID before persist (X-14 batch U7); test: TestStaffService_CreateWithAccount_RejectsCrossClinicOccupationID"},
	{"staffService.Update", statusGuarded, []string{"OccupationID"}, "as Create — validateOccupationOwnership guards *input.OccupationID before repo.Update (X-14 batch U7); test: TestStaffService_Update_RejectsCrossClinicOccupationID"},
	{"treatmentService.Create", statusGuarded, []string{"ConsultationID", "InventoryID", "MedicineID", "ProcedureID"}, "treatment_service.go: validateTreatmentMasterFKs now covers all four — medicine/procedure/consultation (pre-existing, 03bf1cb5) plus Inventory.FindByID(ctx, clinicID, InventoryID) (X-14a, DecreaseStock itself still takes no clinicID); test: TestTreatmentService_Create_RejectsCrossClinicInventoryFK"},
	{"treatmentService.Update", statusGuarded, []string{"ConsultationID", "InventoryID", "MedicineID", "ProcedureID"}, "as Create — validateTreatmentMasterFKs guards InventoryID before persist (X-14a); test: TestTreatmentService_Update_RejectsCrossClinicInventoryFK"},
	{"trimmingCourseService.Update", statusGuarded, []string{"CourseTypeID"}, "trimming_course_service.go: Update now mirrors Create — courseTypeRepo.FindByID(ctx, clinicID, *CourseTypeID) before buildTrimmingCourseUpdate persists course_type_id (X-14b, symmetric with Create's guard); test: TestTrimmingCourseService_Update_RejectsCrossClinicCourseTypeFK"},
	{"trimmingService.Create", statusKnownUnguarded, []string{"CourseID", "OptionIDs", "ReservationTypeID"}, "PARTIAL: ReservationType validated; CourseID/OptionIDs persisted without FindByID."},
	{"trimmingService.Update", statusKnownUnguarded, []string{"CourseID", "OptionIDs"}, "CourseID/OptionIDs persisted without FindByID."},
	{"vaccineService.Create", statusGuarded, []string{"ParentID"}, "vaccine_service.go: validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestVaccineService_Create_RejectsCrossClinicParentFK"},
	{"vaccineService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestVaccineService_Update_RejectsCrossClinicParentFK"},
}

// ───────────────────────────────────────────────────────────────────────────────
// Analyzer (pure over a set of (filename → source) files, so fixtures and the real
// embedded package exercise identical logic).
// ───────────────────────────────────────────────────────────────────────────────

type mfkFieldEntry struct {
	names    []string // empty => embedded field
	typeExpr ast.Expr
}

type mfkWriteFinding struct {
	receiver  string
	method    string
	file      string
	line      int
	masterFKs []string // sorted unique
}

type mfkStats struct {
	filesParsed     int
	structsIndexed  int
	methodsScanned  int
	externalParamQs map[string]struct{} // observed non-stdlib-builtin param qualifiers
}

func mfkKey(receiver, method string) string { return receiver + "." + method }

// isIDType reports whether expr is uint64 / *uint64 / []uint64 / *[]uint64 (the shape of a
// master FK column). The type guard prevents an accidental name collision (e.g. a future
// `GroupID string`) from being mistaken for a master FK.
func isIDType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return isIDType(t.X)
	case *ast.ArrayType:
		return isIDType(t.Elt)
	case *ast.Ident:
		return t.Name == "uint64"
	default:
		return false
	}
}

// localStructName returns the same-package struct type name reachable from expr by peeling
// pointers and slices/arrays (*T, []T, []*T, ...). For a cross-package type (pkg.Type) it
// returns ("", qualifier); the qualifier is reported so the param scan can pin known-safe
// external packages. For anything else it returns ("", "").
func localStructName(expr ast.Expr) (name, qualifier string) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return localStructName(t.X)
	case *ast.ArrayType:
		return localStructName(t.Elt)
	case *ast.Ident:
		return t.Name, ""
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return "", pkg.Name
		}
		return "", ""
	default:
		return "", ""
	}
}

// masterFKsOf returns the set of clinic-scoped master FK field names that structName
// transitively contains. `visiting` guards self/mutual recursion (struct graph cycles),
// which the audit-taxonomy precedent never needed but a struct graph requires.
//
// NOT memoized on purpose: a result computed while truncating a cycle is incomplete, and
// caching it would poison later lookups (a false-negative). DTO graphs are acyclic and tiny
// (whole gate runs in <0.2s), so recomputing per top-level param is both correct and cheap.
func masterFKsOf(structName string, index map[string][]mfkFieldEntry, visiting map[string]bool) map[string]struct{} {
	if visiting[structName] {
		return map[string]struct{}{} // cycle: contribution already accounted up the stack
	}
	visiting[structName] = true

	out := map[string]struct{}{}
	for _, fe := range index[structName] {
		// Direct scalar master FK field (name match + ID-shaped type).
		for _, n := range fe.names {
			if _, ok := clinicScopedMasterFKField[n]; ok && isIDType(fe.typeExpr) {
				out[n] = struct{}{}
			}
		}
		// Recurse into a same-package struct-typed field (named or embedded).
		if child, _ := localStructName(fe.typeExpr); child != "" {
			if _, ok := index[child]; ok {
				for k := range masterFKsOf(child, index, visiting) {
					out[k] = struct{}{}
				}
			}
		}
	}

	delete(visiting, structName)
	return out
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// analyzeServicePackage parses every supplied file, builds a unified struct index, then
// enumerates exported methods whose any parameter transitively carries a clinic-scoped
// master FK. Enumeration is triggered by master-FK CONTAINMENT (not the method verb), so a
// persistence path named Validate*/Confirm/Close/Sync* is captured the same as Create/Update.
func analyzeServicePackage(t *testing.T, files map[string]string) ([]mfkWriteFinding, mfkStats) {
	t.Helper()
	fset := token.NewFileSet()
	stats := mfkStats{externalParamQs: map[string]struct{}{}}

	index := map[string][]mfkFieldEntry{}
	parsed := make([]*ast.File, 0, len(files))

	// Deterministic file order for stable diagnostics.
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	// Pass 1: parse + build struct index across all files.
	for _, name := range names {
		f, err := parser.ParseFile(fset, name, files[name], 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		parsed = append(parsed, f)
		stats.filesParsed++
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			var fields []mfkFieldEntry
			for _, fld := range st.Fields.List {
				var fnames []string
				for _, id := range fld.Names {
					fnames = append(fnames, id.Name)
				}
				fields = append(fields, mfkFieldEntry{names: fnames, typeExpr: fld.Type})
			}
			index[ts.Name.Name] = fields
			stats.structsIndexed++
			return true
		})
	}

	// Pass 2: enumerate exported methods; classify by param containment.
	var findings []mfkWriteFinding
	for fi, f := range parsed {
		fname := names[fi]
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue // free functions are not service write entrypoints
			}
			if !ast.IsExported(fd.Name.Name) {
				continue // unexported helpers (buildXxxUpdateFields, validators) are excluded
			}
			recv, _ := localStructName(fd.Recv.List[0].Type)
			if recv == "" {
				continue
			}
			stats.methodsScanned++

			combined := map[string]struct{}{}
			if fd.Type.Params != nil {
				for _, p := range fd.Type.Params.List {
					child, qualifier := localStructName(p.Type)
					if qualifier != "" {
						stats.externalParamQs[qualifier] = struct{}{}
						continue
					}
					if child == "" {
						continue
					}
					if _, ok := index[child]; !ok {
						continue
					}
					for k := range masterFKsOf(child, index, map[string]bool{}) {
						combined[k] = struct{}{}
					}
				}
			}
			if len(combined) == 0 {
				continue
			}
			findings = append(findings, mfkWriteFinding{
				receiver:  recv,
				method:    fd.Name.Name,
				file:      baseName(fname),
				line:      fset.Position(fd.Pos()).Line,
				masterFKs: sortedKeys(combined),
			})
		}
	}

	return findings, stats
}

func baseName(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// analyzeRealServiceSource runs the analyzer over every embedded non-test .go file.
func analyzeRealServiceSource(t *testing.T) ([]mfkWriteFinding, mfkStats) {
	t.Helper()
	entries, err := fs.Glob(serviceSourceFS, "*.go")
	if err != nil {
		t.Fatalf("glob embedded service source: %v", err)
	}
	files := map[string]string{}
	for _, name := range entries {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := serviceSourceFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", name, err)
		}
		files[name] = string(src)
	}
	return analyzeServicePackage(t, files)
}

func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]int{}
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

// ───────────────────────────────────────────────────────────────────────────────
// The gate.
// ───────────────────────────────────────────────────────────────────────────────

// reconcileMasterFKWrites is the pure bidirectional check, separated so its three failure
// modes can be frozen by a fixture test (TestMasterFKWriteInventory_GateDetectsViolations)
// rather than only transiently during a RED run. It returns one human-readable violation per
// problem (empty => clean). found maps "receiver.Method" → sorted master FK field set.
func reconcileMasterFKWrites(found map[string][]string, allowlist []masterFKWriteEntry) []string {
	var violations []string

	allow := map[string]masterFKWriteEntry{}
	for _, e := range allowlist {
		if _, dup := allow[e.key]; dup {
			violations = append(violations, "duplicate allowlist entry "+e.key)
		}
		if e.status != statusGuarded && e.status != statusKnownUnguarded && e.status != statusExempt {
			violations = append(violations, "allowlist entry "+e.key+" has invalid status "+string(e.status))
		}
		allow[e.key] = e
	}

	// (i) every enumerated master-FK write must be allowlisted with a matching field set.
	for key, fks := range found {
		e, ok := allow[key]
		if !ok {
			violations = append(violations, "master-FK write "+key+" "+joinSet(fks)+
				" is NOT on the allowlist. Add an entry with a status (guarded/known-unguarded/exempt) "+
				"+ reason and pin its master FK fields. This gate enforces REVIEW COVERAGE, not "+
				"correctness — if the method persists the FK, also add a runtime isolation test.")
			continue
		}
		if !equalStringSets(e.masterFKs, fks) {
			violations = append(violations, "master-FK write "+key+" carries "+joinSet(fks)+
				" but allowlist pins "+joinSet(e.masterFKs)+". A master FK was added/removed from its DTO "+
				"— re-review (the #124 nested-field shape) and update the entry's masterFKs.")
		}
	}
	// (ii) no stale allowlist entry.
	for key := range allow {
		if _, ok := found[key]; !ok {
			violations = append(violations, "allowlist entry "+key+
				" no longer matches any master-FK write (renamed/removed/FK dropped). Delete the stale entry.")
		}
	}
	return violations
}

func joinSet(s []string) string { return "[" + strings.Join(s, " ") + "]" }

// TestMasterFKWriteInventory_AllowlistMatchesRealSource is the gate: every service method
// whose input DTO transitively carries a clinic-scoped master FK must be on the allowlist
// with a matching pinned field set, and every allowlist entry must still match a real
// method. Floors prevent a vacuous green if the embed glob or AST matching silently breaks.
func TestMasterFKWriteInventory_AllowlistMatchesRealSource(t *testing.T) {
	findings, stats := analyzeRealServiceSource(t)

	if stats.filesParsed < 100 {
		t.Fatalf("only %d non-test service files parsed; embed glob likely broken (would vacuously pass)", stats.filesParsed)
	}
	if stats.structsIndexed < 150 {
		t.Fatalf("only %d structs indexed; struct-index pass likely broken", stats.structsIndexed)
	}
	if stats.methodsScanned < 200 {
		t.Fatalf("only %d exported methods scanned; method walk likely broken", stats.methodsScanned)
	}
	// Floor, not an exact pin: master-FK writes grow with the schema. A broken matcher would
	// drop this near 0. Raise only if it ever exceeds the real count.
	if len(findings) < 20 {
		t.Fatalf("only %d master-FK writes enumerated; param-containment matching likely broken", len(findings))
	}

	found := map[string][]string{}
	for _, f := range findings {
		found[mfkKey(f.receiver, f.method)] = f.masterFKs
	}

	for _, v := range reconcileMasterFKWrites(found, masterFKWriteAllowlist) {
		t.Error(v)
	}
}

// TestMasterFKWriteInventory_GateDetectsViolations freezes the gate's three failure modes so
// the old #124-style slip-through cannot regress: (a) a new master-FK write missing from the
// allowlist FAILS, (b) a stale allowlist entry FAILS, (c) a new master FK added to an already
// listed DTO (field-set drift) FAILS. The clean baseline must report zero.
func TestMasterFKWriteInventory_GateDetectsViolations(t *testing.T) {
	base := []masterFKWriteEntry{
		{"svc.Create", statusGuarded, []string{"MedicineID"}, "fixture"},
	}

	t.Run("clean baseline reports no violation", func(t *testing.T) {
		got := reconcileMasterFKWrites(map[string][]string{"svc.Create": {"MedicineID"}}, base)
		if len(got) != 0 {
			t.Fatalf("expected 0 violations, got %v", got)
		}
	})
	t.Run("unlisted master-FK write fails (the core regression guard)", func(t *testing.T) {
		got := reconcileMasterFKWrites(map[string][]string{
			"svc.Create":     {"MedicineID"},
			"svc.NewPersist": {"VaccineID"}, // not on the allowlist
		}, base)
		if len(got) != 1 || !strings.Contains(got[0], "svc.NewPersist") || !strings.Contains(got[0], "NOT on the allowlist") {
			t.Fatalf("expected the unlisted write to be flagged, got %v", got)
		}
	})
	t.Run("stale allowlist entry fails", func(t *testing.T) {
		got := reconcileMasterFKWrites(map[string][]string{}, base)
		if len(got) != 1 || !strings.Contains(got[0], "stale") {
			t.Fatalf("expected the stale entry to be flagged, got %v", got)
		}
	})
	t.Run("new master FK added to listed DTO fails (#124 nested-field shape)", func(t *testing.T) {
		got := reconcileMasterFKWrites(map[string][]string{
			"svc.Create": {"MedicineID", "ProcedureID"}, // a master FK was added since review
		}, base)
		if len(got) != 1 || !strings.Contains(got[0], "added/removed") {
			t.Fatalf("expected the field-set drift to be flagged, got %v", got)
		}
	})
	t.Run("duplicate allowlist key fails", func(t *testing.T) {
		dup := []masterFKWriteEntry{
			{"svc.Create", statusGuarded, []string{"MedicineID"}, "a"},
			{"svc.Create", statusGuarded, []string{"MedicineID"}, "b"},
		}
		got := reconcileMasterFKWrites(map[string][]string{"svc.Create": {"MedicineID"}}, dup)
		if len(got) != 1 || !strings.Contains(got[0], "duplicate") {
			t.Fatalf("expected duplicate key to be flagged, got %v", got)
		}
	})
}

// TestMasterFKWriteInventory_StatusesAreLive proves the status taxonomy is exercised (not a
// dead enum) and that the guarded/known-unguarded split actually reflects the codebase.
func TestMasterFKWriteInventory_StatusesAreLive(t *testing.T) {
	counts := map[masterFKWriteStatus]int{}
	for _, e := range masterFKWriteAllowlist {
		counts[e.status]++
	}
	if counts[statusGuarded] == 0 {
		t.Error("no 'guarded' allowlist entries; the guarded write sites (treatment/vaccination/…) drifted")
	}
	if counts[statusKnownUnguarded] == 0 {
		t.Error("no 'known-unguarded' allowlist entries; the residual P1 list is empty — verify, don't assume")
	}
}

// TestMasterFKWriteInventory_NoUnknownCrossPackageParam closes the cross-package false-negative
// gap: a service method receiving a struct from a package the gate cannot introspect could hide
// a master FK. All such qualifiers must be in knownSafeParamQualifiers (none of which is a
// service-local DTO). A NEW qualifier trips this and forces review.
func TestMasterFKWriteInventory_NoUnknownCrossPackageParam(t *testing.T) {
	_, stats := analyzeRealServiceSource(t)
	for q := range stats.externalParamQs {
		if _, ok := knownSafeParamQualifiers[q]; !ok {
			t.Errorf("service method parameter uses unknown package qualifier %q. If %q.<Type> cannot "+
				"carry a clinic-scoped master FK, add it to knownSafeParamQualifiers; otherwise the gate "+
				"must be extended to introspect it (cross-package master-FK propagation).", q, q)
		}
	}
}

// TestMasterFKWriteInventory_Analyzer pins the analyzer behaviour on inline fixtures: the
// detection-worthy shapes (direct/slice/nested/embedded/alias/non-verb name) must be detected,
// and the must-not shapes (no master FK, unexported, wrong type, bare param) must not. This
// both drives the analyzer (RED→GREEN) and freezes the failure modes (esp. the #124 nested
// shape and the verb-prefix false-negative) against reintroduction.
func TestMasterFKWriteInventory_Analyzer(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantKey string   // "" => expect zero findings
		wantFKs []string // expected pinned master FK set
	}{
		{
			name:    "direct master FK field",
			src:     `type T struct { MedicineID *uint64 }` + "\n" + `func (s *svc) Create(in *T) error { return nil }`,
			wantKey: "svc.Create", wantFKs: []string{"MedicineID"},
		},
		{
			// The real method shape: ctx (SelectorExpr qualifier "context" → skipped),
			// clinicID (bare uint64 not in index → skipped), then the *Input struct param.
			name:    "real signature (ctx, clinicID, *Input)",
			src:     `type T struct { MedicineID *uint64 }` + "\n" + `func (s *svc) Create(ctx context.Context, clinicID uint64, in *T) error { return nil }`,
			wantKey: "svc.Create", wantFKs: []string{"MedicineID"},
		},
		{
			name:    "slice param (#124 ReplaceItems shape)",
			src:     `type Item struct { ExamTypeFieldID *uint64 }` + "\n" + `func (s *svc) ReplaceItems(in []Item) error { return nil }`,
			wantKey: "svc.ReplaceItems", wantFKs: []string{"ExamTypeFieldID"},
		},
		{
			name:    "two-level nested sub-struct slice",
			src:     `type Sub struct { VaccineID uint64 }` + "\n" + `type Outer struct { Items []Sub }` + "\n" + `func (s *svc) Create(in *Outer) error { return nil }`,
			wantKey: "svc.Create", wantFKs: []string{"VaccineID"},
		},
		{
			name:    "embedded (anonymous) struct field",
			src:     `type Base struct { CageID *uint64 }` + "\n" + `type Outer struct { Base }` + "\n" + `func (s *svc) Update(in Outer) error { return nil }`,
			wantKey: "svc.Update", wantFKs: []string{"CageID"},
		},
		{
			name:    "alias field names (CourseID/OptionIDs)",
			src:     `type T struct { CourseID *uint64; OptionIDs []uint64 }` + "\n" + `func (s *svc) Save(in *T) error { return nil }`,
			wantKey: "svc.Save", wantFKs: []string{"CourseID", "OptionIDs"},
		},
		{
			name:    "non-verb name captured by containment (verb-independence)",
			src:     `type T struct { TrimmingCourseID *uint64 }` + "\n" + `func (v *validators) ValidateAndCreate(in *T) error { return nil }`,
			wantKey: "validators.ValidateAndCreate", wantFKs: []string{"TrimmingCourseID"},
		},
		{
			name: "no master FK (BulkUpdateSortOrder false-positive regression)",
			src:  `type T struct { ID uint64; SortOrder int }` + "\n" + `func (s *svc) BulkUpdateSortOrder(in *T) error { return nil }`,
		},
		{
			name: "unexported method excluded",
			src:  `type T struct { MedicineID *uint64 }` + "\n" + `func (s *svc) create(in *T) error { return nil }`,
		},
		{
			name: "type guard: master name but non-ID type ignored",
			src:  `type T struct { GroupID string }` + "\n" + `func (s *svc) Create(in *T) error { return nil }`,
		},
		{
			name: "bare scalar param not detected (documented limitation)",
			src:  `func (s *svc) Upsert(medicineID uint64) error { return nil }`,
		},
		{
			name:    "cycle in struct graph terminates",
			src:     `type A struct { B *B; MedicineID uint64 }` + "\n" + `type B struct { A *A }` + "\n" + `func (s *svc) Create(in *A) error { return nil }`,
			wantKey: "svc.Create", wantFKs: []string{"MedicineID"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package p\n" + tc.src + "\n"
			findings, _ := analyzeServicePackage(t, map[string]string{"fixture.go": src})
			if tc.wantKey == "" {
				if len(findings) != 0 {
					t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
				}
				return
			}
			if len(findings) != 1 {
				t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
			}
			got := findings[0]
			if mfkKey(got.receiver, got.method) != tc.wantKey {
				t.Fatalf("got key %s, want %s", mfkKey(got.receiver, got.method), tc.wantKey)
			}
			if !equalStringSets(got.masterFKs, tc.wantFKs) {
				t.Fatalf("got fields %v, want %v", got.masterFKs, tc.wantFKs)
			}
		})
	}
}

// TestMasterFKWriteInventory_MatcherIsLiveOnRealFieldSet is a positive liveness probe: a
// minimal MedicineID DTO IS detected by the REAL clinicScopedMasterFKField set, proving the
// matcher is wired to the live field set. (The negative path — "a broken matcher fails rather
// than vacuously passes" — is enforced by the floor guards in _AllowlistMatchesRealSource.)
func TestMasterFKWriteInventory_MatcherIsLiveOnRealFieldSet(t *testing.T) {
	src := "package p\n" +
		`type T struct { MedicineID *uint64 }` + "\n" +
		`func (s *svc) Create(in *T) error { return nil }` + "\n"

	findings, _ := analyzeServicePackage(t, map[string]string{"fixture.go": src})
	if len(findings) != 1 {
		t.Fatalf("matcher should detect MedicineID fixture; got %d findings", len(findings))
	}
}

// TestMasterFKWriteInventory_NoCyclePoisoning pins the correctness of the deliberately
// un-memoized masterFKsOf: with a mutual cycle A→B→A where A carries the master FK, BOTH a
// method taking *A and a method taking *B must report MedicineID. A memoized implementation
// would cache an incomplete set for B (visited mid-cycle) and silently drop B's finding — the
// false-negative the memo removal prevents. Acyclic in practice; this freezes the property.
func TestMasterFKWriteInventory_NoCyclePoisoning(t *testing.T) {
	src := "package p\n" +
		`type A struct { B *B; MedicineID uint64 }` + "\n" +
		`type B struct { A *A }` + "\n" +
		`func (s *svc) Create(in *A) error { return nil }` + "\n" +
		`func (s *svc2) List(in *B) error { return nil }` + "\n"

	findings, _ := analyzeServicePackage(t, map[string]string{"fixture.go": src})
	got := map[string][]string{}
	for _, f := range findings {
		got[mfkKey(f.receiver, f.method)] = f.masterFKs
	}
	for _, key := range []string{"svc.Create", "svc2.List"} {
		if !equalStringSets(got[key], []string{"MedicineID"}) {
			t.Errorf("%s: got %v, want [MedicineID] (cycle poisoning would drop B's finding)", key, got[key])
		}
	}
}
