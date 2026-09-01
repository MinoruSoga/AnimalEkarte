package lintscan

// Write-side clinic-scope review-coverage gate — every service method that receives a
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
//	multiple package/function boundaries, which go/ast cannot do reliably. #124 (f4e7b7a7) is the
//	standing counterexample: the parent exam_type_id was validated while a NESTED
//	exam_type_field_id was not. A static rule that "looks 80% validated" is the exact
//	"felt-validated" failure mode that let #124 regress, so we deliberately do not build
//	one. Correctness is enforced at RUNTIME by cross_tenant_master_fk_write_test.go and
//	the *_clinic_isolation_test.go guards. THIS gate enforces only REVIEW COVERAGE:
//	it forces every master-FK write onto a curated allowlist with a status. It is the
//	write-side analogue of repository/preload_clinic_scope_lint_test.go (read side).
//
// Residual gaps this gate does NOT cover (documented in repository/CLAUDE.md's clinic-scope Preload rule):
//  1. Correctness of the ownership guard (see above) — runtime tests are the source of truth.
//  2. A master FK arriving as a BARE scalar parameter (e.g. `medicineID uint64`) rather
//     than via a DTO struct field. Detection keys on DTO struct fields by design (the
//     request-binding architecture binds master FKs into XxxInput DTOs). Known bare-param
//     sites (last verified by grep, not automated — re-check when new Set*IDs/Link* methods
//     are added): medicineDoseParamService.Upsert, staffService.Set{Excluded,Capable}ReservationTypeIDs,
//     staffService.SetPermissionGroupIDs (PermissionGroup; mitigated — repo UpdateStaffGroups
//     does a clinic-scoped IN-check), reservationTypeService.LinkOccupation (mitigated — FindByID).
//  3. A master FK propagated through a non-`model.` cross-package struct parameter. The sole
//     reviewed occurrence is owner.PetRegistrationIntent at the pet owner-registration adapter;
//     knownSafeParamQualifiers plus knownReviewedExternalParamOccurrences pin both its exact
//     type and call site so any new external command fails closed and is forced into review.
//  4. (field-name/shape matching limits, independent of role scope) A master-FK-bearing DTO reached only via an UNREGISTERED
//     field name (not a key of clinicScopedMasterFKField) is invisible by design — this gate
//     detects by field-name matching, not exhaustive type inspection. A struct type reachable
//     only through a shape masterFKsOf's recursion does not follow (an interface-typed field, a
//     map/func-typed field, a generic type parameter, or a type alias) is also invisible — the
//     recursion only walks named/embedded struct-typed fields reached by peeling pointers and
//     slices/arrays. This gate makes no claim of exhaustiveness against either gap, and scanning
//     more of internal/ does not close them — a single struct's fields can hide from field-
//     name/shape-based matching regardless of scan breadth. Cross-tenant ownership correctness
//     remains enforced at RUNTIME by cross_tenant_master_fk_write_test.go and the
//     *_clinic_isolation_test.go guards (see the SCOPE / ROLE BOUNDARY section above).
//
// Technique reused from repository/preload_clinic_scope_lint_test.go and
// model/audit_taxonomy_exhaustiveness_test.go, adapted for BE9-1: file discovery now goes
// through the shared internal/lintscan package (lintscan.WalkInternalTreeT), which walks the
// WHOLE module's internal/** tree via filepath.WalkDir — package-independent, not go:embed-
// rooted; see internal/lintscan/lintscan.go's package doc for the discovery-layer/content-
// classification-layer split this reuse relies on. lintscan already excludes *_test.go files,
// testdata/, and vendor/, so this gate never inspects its own fixtures or other test helpers.
// Parsing (go/parser+go/ast),
// the curated field set + allowlist, and pinning with good/bad fixtures, floor guards,
// bidirectional exhaustiveness, and a negative self-check so the gate can never pass vacuously
// are unchanged from the pre-BE9-1 go:embed-based version.
//
// SERVICE-WRITE ROLE SCOPE — analyzeRealServiceSource narrows lintscan's whole-tree walk down to
// the service-write role package(s): isServiceWriteRolePackage / serviceWriteRolePackagePrefixes,
// defined just above analyzeRealServiceSource below. This narrowing is the gate's permanent,
// by-design scope. This gate reviews SERVICE-WRITE persistence coverage — the layer where a
// request-derived DTO is validated and actually persisted. internal/model (GORM column/struct
// definitions) and internal/handler (HTTP request/response DTOs and transport binding) are
// DIFFERENT layers with DIFFERENT concerns — schema definition and transport binding, not
// persistence write logic — and are intentionally, permanently excluded from this gate's role
// scope; that is a deliberate layer boundary, not a coverage gap. A prior independent AST audit
// found 206 field-name occurrences of clinic-scoped master FK field names across internal/model
// and internal/handler combined; every one of them is correctly out of THIS gate's scope
// (schema/transport, not a service-write persistence decision), not silently allowlisted or
// missed. Widening the role scope to internal/model or internal/handler would not add
// write-review coverage — it would blur this gate's single responsibility with two layers it was
// never designed to audit. When BE9-2 migrates service-write persistence logic OUT of
// internal/service into a new domain package (e.g. internal/owner, internal/pet), that new
// package's lintscan prefix is added to serviceWriteRolePackagePrefixes — and ONLY there; no
// other code in this file changes to admit it. See
// TestMasterFKWriteInventory_RoleFilterPackageScope and
// TestMasterFKWriteInventory_RoleFilterExtensionPointGeneralizes for the extension-point proof,
// and TestMasterFKWriteInventory_RoleFilterIntegrationExcludesOtherLayers for the direct
// service-vs-model-vs-handler detection proof.
//
// CROSS-PACKAGE TYPE-NAME COLLISION SAFETY: the struct index is keyed by (containing-directory-
// of-the-file relative to internal/, type name) via mfkStructKey, not by bare type name alone —
// two unrelated internal/ packages may each declare an unrelated local struct with the same name
// (e.g. a handler.Input and a service.Input), and an unqualified param type reference is only
// resolved against entries whose directory matches that SAME file's own directory — mirroring
// real Go semantics, where an unqualified type reference can only resolve to a type declared in
// that file's own package. This makes analyzeServicePackage itself correct across the WHOLE
// internal/** tree, independent of whatever subset of files analyzeRealServiceSource happens to
// feed it — exercised directly against multi-package fixtures by
// TestMasterFKWriteInventory_AnalyzerDetectsViolationUnderNestedPathFilename and
// TestMasterFKWriteInventory_CrossPackageTypeNameCollisionIsolation. See the SERVICE-WRITE ROLE
// SCOPE section above for why the real feed itself stays permanently restricted to the
// service-write role scope regardless of the analyzer's own whole-tree correctness.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

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
	"Diagnosis2TypeID":      "DiagnosisType",
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

// knownSafeParamQualifiers are package qualifiers (the `pkg` in `pkg.Type`) for which every
// parameter type is infrastructure/read-only and cannot carry a clinic-scoped master FK.
// A qualifier whose commands may carry a master FK must not be listed here; use the exact
// type-and-occurrence allowlist below instead.
var knownSafeParamQualifiers = map[string]struct{}{
	"context": {}, // context.Context
	"time":    {}, // time.Time
	"model":   {}, // model.* — domain models; service-local DTOs (which carry master FKs) are not here
	"uuid":    {}, // github.com/google/uuid
	"io":      {}, // io.Reader (file/stream import)
	// repository.* — read-filter/query structs (e.g. MedicalRecordListFilters), not persistence
	// write DTOs. Fields checked against clinicScopedMasterFKField: PetID/OwnerID/StartDate/
	// EndDate/Status/Search carry no master FK; DoctorID references Staff (explicitly exempt from
	// clinic-scope write checks — multi-clinic assignment, see repository/CLAUDE.md's clinic-scope Preload rule); and
	// AnimalSpeciesID is a global (non clinic-scoped) master. Re-review if repository.* gains a
	// write-path DTO carrying a clinic-scoped master FK field.
	"repository": {},
	// gin.* — added BE9-2C when serviceWriteRolePackagePrefixes grew "medicalrecord/": unlike
	// internal/service (pure application logic, no Gin dependency), a BE9-2 domain package
	// holds handler+service+repository in one directory (ADR-006), so this gate's scan now
	// also sees medicalrecord's *gin.Context-taking HTTP handler methods for the first time.
	// gin.Context is the HTTP request/response context — it is never a persisted DTO and
	// cannot carry a clinic-scoped master FK field; the actual master-FK-bearing DTOs those
	// handlers build (createExamTypeRequest.toServiceInput() → *CreateExamTypeInput, etc.)
	// are medicalrecord-local types with no package qualifier, so they're inspected by this
	// gate exactly as before — only the transport-layer gin.Context parameter is exempted.
	// Every future BE9-2C/2D/2E domain package will hit this same qualifier once it merges
	// handler code into role scope; no further per-domain qualifier addition should be needed.
	"gin": {},
	// gorm.DB is a transaction/query handle passed to repository-oriented helpers. It is not a
	// request DTO and cannot carry a clinic-scoped master foreign key.
	"gorm": {},
	// auth.Transactor / PasswordResetConfig / AuthAuditEntry are infrastructure and audit
	// boundary values. None contains a clinic-scoped master foreign key.
	"auth": {},
	// lstep.LifecycleAuditEntry is the transaction-local audit adapter input at the
	// composition boundary. It carries clinic/actor/resource IDs and audit labels only;
	// it is not a persistence write DTO and cannot carry a clinic-scoped master FK.
	"lstep": {},
	// NOTE (BE9-2D sub-batch③): a "medicalrecord" qualifier was temporarily exempted in the
	// Batch B middle state (labAuditAdapter.LogEntry(ctx, *medicalrecord.AuditEntry) in the now-
	// deleted lab_middle_state.go). Batch C relocated lab construction — and that adapter — to
	// cmd/api/main.go, so the service package no longer holds any *medicalrecord.* parameter and
	// the exemption is removed again (mirrors sub-batch②'s medicalrecord_middle_state.go lifecycle).
}

// knownReviewedExternalParamOccurrences is deliberately narrower than a qualifier-wide
// exemption. owner.PetRegistrationIntent carries nested InsuranceID values, so only the
// reviewed pet adapter occurrence is accepted. A new owner command, or reuse of the same
// command at another write entrypoint, must fail TestMasterFKWriteInventory_NoUnknownCrossPackageParam.
var knownReviewedExternalParamOccurrences = map[string]map[string]struct{}{
	"owner.PetRegistrationIntent": {
		"pet.OwnerRegistrationAdapter.CreateForOwnerRegistration": {},
	},
}

// masterFKWriteStatus records WHY a master-FK write is on the allowlist. The gate does not
// verify these; they are the human review record.
type masterFKWriteStatus string

const (
	// statusGuarded: a runtime isolation test (cross_tenant_master_fk_write_test.go or a
	// *_clinic_isolation_test.go) proves a cross-clinic master FK is rejected.
	statusGuarded masterFKWriteStatus = "guarded"
	// statusKnownUnguarded: reviewed; NO dedicated isolation test confirms ownership
	// rejection yet. Residual high-priority risk tracked here (the candidate work list for adding tests).
	statusKnownUnguarded masterFKWriteStatus = "known-unguarded"
	// statusExempt: not a persistence write of the FK (preview/read/validate-only), or the
	// FK is validated by a different documented mechanism. Reason must justify it.
	// LabImportDuplicateCheckerDB.IsDuplicate uses this: ExamTypeID is a read filter only.
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
//	                  partially-guarded methods). Residual high-priority — the work list for runtime tests.
//	exempt          = does not persist the FK / validated by a documented alternate mechanism.
//
// REMINDER: status is a human review record. The gate does NOT verify it (see file header).
var masterFKWriteAllowlist = []masterFKWriteEntry{
	// ── guarded (FindByID ownership check covers every master FK; most have runtime isolation tests) ──
	{"accountingService.Complete", statusGuarded, []string{"MerchandiseItemID", "PaymentMethodID", "TrimmingCourseID", "TrimmingOptionID"}, "internal/billing/accounting_complete.go (BUG-018): every item FK reaches the DB only through s.itemWriter.CreateItemForComplete inside the completion transaction, which runs billingItemRepository.ValidateCreateReferences against input.ClinicID (see billingItemService.CreateItemForComplete); PaymentMethodID goes through resolvePaymentMethodMasterID against loadPaymentMethodSystemKeyToID(ctx, input.ClinicID) — the same mismatch-rejecting mechanism as accountingService.Update. COVERAGE CAVEAT: verified by code inspection plus the delegates' tests (TestBillingItemRepository_ValidateCreateReferences, TestBillingItemService_CreateItem_RuntimeReferenceIsolation, TestAccountingService_Update_RejectsForeignPaymentMethodID); there is NO cross-clinic rejection test driven through the Complete path itself — accounting_complete_test.go has no clinic-isolation case. Residual: add TestAccountingService_Complete_RejectsForeignMasterFK."},
	{"accountingService.Update", statusGuarded, []string{"PaymentMethodID"}, "(moved to internal/billing in B④): resolvePaymentMethodMasterID の mismatch 拒否ロジックで validated; test: TestAccountingService_Update_RejectsForeignPaymentMethodID"},
	{"billingItemService.CreateItem", statusGuarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "internal/billing/billing_item_service.go persists MerchandiseItemID and, inside the write transaction, billingItemRepository.ValidateCreateReferences locks and validates MerchandiseItemID plus attached TrimmingCourseID/TrimmingOptionID against the authenticated clinic before Create; tests: TestBillingItemRepository_ValidateCreateReferences, TestBillingItemService_CreateItem_RuntimeReferenceIsolation"},
	{"billingItemService.CreateItemForComplete", statusGuarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "internal/billing/billing_item_service.go (BUG-018): thin ambient-tx wrapper that delegates to the same createItemInAmbientTx as billingItemService.CreateItem — it locks the parent billing, rejects finalized status, then calls billingItemRepository.ValidateCreateReferences(ctx, input.ClinicID, …) for MerchandiseItemID/TrimmingCourseID/TrimmingOptionID before Create. Only totals recalculation and post-close recording are skipped (the Complete command performs those once); no FK validation is skipped. tests: TestBillingItemRepository_ValidateCreateReferences, TestBillingItemService_CreateItem_RuntimeReferenceIsolation (shared code path)."},
	{"campaignService.Create", statusGuarded, []string{"TargetItemIDs"}, "campaign_service.go: WithTx then uniqueSorted TargetItemIDs + validateOwnedMerchandiseItemIDs → merchandiseItemRepo.FindByID (ambient FOR SHARE) before Create (X-5 / BE-ACT-CAMPAIGN-TARGET-SERIALIZATION); tests: TestCampaignService_Create_RejectsCrossClinicTargetItemFK, TestCampaignService_Create_ValidatesMerchandiseTargetsInsideTransaction"},
	{"campaignService.Update", statusGuarded, []string{"TargetItemIDs"}, "as Create — target-changing Update validates uniqueSorted *input.TargetItemIDs inside WithTx before ReplaceTargets (X-5 / BE-ACT-CAMPAIGN-TARGET-SERIALIZATION); tests: TestCampaignService_Update_RejectsCrossClinicTargetItemFK, TestCampaignService_Update_ValidatesMerchandiseTargetsInsideTransaction"},
	{"carePlanItemService.Create", statusGuarded, []string{"HospitalizationPlanID", "MedicineID", "ProcedureID"}, "internal/medicalrecord/care_plan_item_service.go (BE9-2D ⑤, moved from internal/service): validateMasterFKs now covers all three — medicine/procedure (pre-existing) plus hospPlanRepo.FindByID(ctx, clinicID, HospitalizationPlanID) (X-14); test: TestCarePlanItemService_Create_RejectsCrossClinicHospitalizationPlanFK"},
	{"carePlanItemService.Update", statusGuarded, []string{"HospitalizationPlanID", "MedicineID", "ProcedureID"}, "as Create (internal/medicalrecord/care_plan_item_service.go) — validateMasterFKs guards *input.HospitalizationPlanID before persist (X-14); test: TestCarePlanItemService_Update_RejectsCrossClinicHospitalizationPlanFK"},
	{"checkupFieldResultService.ReplaceForCheckup", statusGuarded, []string{"CheckupTypeFieldID"}, "internal/medicalrecord/checkup_field_result_service.go (BE9-2D, moved from internal/service): each CheckupTypeFieldID validated within the owned checkup_type's fields (#124 同型, #211); test in internal/medicalrecord/checkup_field_result_service_test.go"},
	{"checkupService.Create", statusGuarded, []string{"CheckupTypeID"}, "internal/medicalrecord/checkup_service.go (BE9-2D, moved from internal/service): checkupTypeRepo.FindByID(ctx, clinicID, CheckupTypeID); test: TestCheckupService_Create_RejectsCrossClinicCheckupType (internal/medicalrecord/cross_tenant_master_fk_write_test.go)"},
	{"checkupService.Update", statusGuarded, []string{"CheckupTypeID"}, "internal/medicalrecord/checkup_service.go (BE9-2D): checkupTypeRepo.FindByID(ctx, clinicID, *CheckupTypeID)"},
	{"clinicalPlanService.Update", statusGuarded, []string{"Diagnosis2TypeID", "Diagnosis2NameID", "DiagnosisNameID", "DiagnosisTypeID"}, "internal/medicalrecord/clinical_plan_service.go (BE9-2D sub-batch④a, moved from internal/service): validateDiagnosisFKs FindByID for all four slots (03bf1cb5); tests: TestClinicalPlanService_Update_RejectsCrossClinicDiagnosisFK/Name (internal/medicalrecord/cross_tenant_master_fk_write_test.go)"},
	{"diagnosisNameService.Create", statusGuarded, []string{"DiagnosisTypeID"}, "internal/medicalrecord/diagnosis_service.go (BE9-2C, moved from internal/service/diagnosis_service.go): typeRepo.FindByID(ctx, clinicID, DiagnosisTypeID) (#020)"},
	{"diagnosisNameService.Update", statusGuarded, []string{"DiagnosisTypeID"}, "internal/medicalrecord/diagnosis_service.go (BE9-2C, moved from internal/service/diagnosis_service.go): typeRepo.FindByID(ctx, clinicID, *DiagnosisTypeID)"},
	{"examinationService.Create", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "ExamTypeID and nested Items[].ExamTypeFieldID are validated against the locked clinic-owned exam type; tests: TestExaminationService_Create_RejectsCrossClinicExamType and TestExaminationService_Create_RejectsCrossClinicExamTypeField"},
	{"examinationService.ReplaceItems", statusGuarded, []string{"ExamTypeFieldID"}, "each ExamTypeFieldID validated within the owned exam_type's items (#124, f4e7b7a7); test present"},
	{"examinationService.Update", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "ExamTypeID and nested Items[].ExamTypeFieldID are validated against the locked clinic-owned exam type when present; tests: TestExaminationService_Update_RejectsCrossClinicExamType and TestExaminationService_Update_RejectsCrossClinicExamTypeField"},
	{"hospitalizationService.Create", statusGuarded, []string{"CageID"}, "internal/medicalrecord/hospitalization_service.go (BE9-2D ⑤, moved from internal/service): s.cageRepo.FindByID(ctx, clinicID, CageID) when non-nil (X-14); test: TestHospitalizationService_Create_RejectsCrossClinicCageFK"},
	{"hospitalizationService.Update", statusGuarded, []string{"CageID"}, "as Create (internal/medicalrecord/hospitalization_service.go) — s.cageRepo.FindByID(ctx, clinicID, *CageID) when non-nil (X-14); test: TestHospitalizationService_Update_RejectsCrossClinicCageFK"},
	{"ownerService.CreateWithPets", statusGuarded, []string{"InsuranceID"}, "internal/owner/service_core.go: validateOwnerPetsInsuranceOwnership loops insuranceFinder.FindByID(ctx, clinicID, id) over every nested Pets[i].InsuranceID before repo.CreateWithPets; test: TestOwnerService_CreateWithPets_MapsMissingInsuranceToInvalidInput (internal/owner/service_error_contract_test.go)"},
	{"petService.Create", statusGuarded, []string{"InsuranceID"}, "internal/pet/service.go: insuranceRepo.FindByID(ctx, clinicID, InsuranceID) when non-nil; test: TestPetService_Create/rejects_insurance_not_in_clinic (internal/pet/service_test.go)"},
	{"petService.Update", statusGuarded, []string{"InsuranceID"}, "internal/pet/service.go: insuranceRepo.FindByID(ctx, clinicID, **InsuranceID) when non-nil and non-NULL; test: TestPetService_Update_InsuranceValidation (internal/pet/service_test.go)"},
	{"trimmingCourseService.Create", statusGuarded, []string{"CourseTypeID"}, "trimming_course_service.go:118 courseTypeRepo.FindByID when non-nil (#73)"},
	{"vaccinationService.Create", statusGuarded, []string{"VaccineID"}, "internal/medicalrecord/vaccination_service.go (BE9-2D, moved from internal/service): vaccineRepo.FindByID(ctx, clinicID, VaccineID) inside the relation-validation transaction (#125, BUG-420); test: TestVaccinationService_Create_RejectsCrossClinicVaccine"},
	{"vaccinationService.Update", statusGuarded, []string{"VaccineID"}, "internal/medicalrecord/vaccination_service.go (BE9-2D): validates the merged effective VaccineID on every PATCH inside the relation-validation transaction (BUG-420); test: TestVaccinationService_Update_RejectsCrossClinicVaccine"},

	{"checkupTypeService.Create", statusGuarded, []string{"ParentID"}, "internal/medicalrecord/checkup_type_service.go (BE9-2D, moved from internal/service): validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestCheckupTypeService_Create_RejectsCrossClinicParentFK (internal/medicalrecord/cross_tenant_master_fk_write_test.go)"},
	{"checkupTypeService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); internal/medicalrecord/checkup_type_service.go; test: TestCheckupTypeService_Update_RejectsCrossClinicParentFK"},
	{"consultationService.Create", statusGuarded, []string{"ParentID"}, "internal/medicalrecord/consultation_service.go (BE9-2D ⑥, moved): validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestConsultationService_Create_RejectsCrossClinicParentFK"},
	{"consultationService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestConsultationService_Update_RejectsCrossClinicParentFK"},
	{"examTypeService.Create", statusGuarded, []string{"ParentID"}, "internal/medicalrecord/exam_type_service.go (BE9-2C, moved from internal/service/exam_type_service.go): validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestExamTypeService_Create_RejectsCrossClinicParentFK (internal/medicalrecord/exam_type_cross_tenant_test.go)"},
	{"examTypeService.CreateField", statusGuarded, []string{"ExamTypeID"}, "CreateExamTypeFieldCommand.ExamTypeID is resolved through the clinic-scoped parent lookup inside the write transaction before CreateField; DB test: TestExamTypeService_FieldCRUDAndReorder_ClinicIsolation"},
	{"examTypeService.ReplaceReferenceRanges", statusGuarded, []string{"ExamTypeFieldID"}, "ReplaceReferenceRangesCommand.ExamTypeFieldID is locked with clinic_id and exam_type_id correlation before replacement in the same transaction; DB test: TestExamReferenceRangeService_ReplaceAtomicClinicSafeAndNoHistoryRewrite"},
	{"examTypeService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestExamTypeService_Update_RejectsCrossClinicParentFK (internal/medicalrecord/exam_type_cross_tenant_test.go)"},
	{"inquiryService.Save", statusGuarded, []string{"ChiefComplaintTypeID"}, "internal/medicalrecord/inquiry_service.go (BE9-2D, moved from internal/service): chiefComplaintTypeRepo.FindByID(ctx, input.ClinicID, *ChiefComplaintTypeID) before persist (X-14 batch U4); test: TestInquiryService_Save_RejectsCrossClinicChiefComplaintType (internal/medicalrecord/cross_tenant_master_fk_write_test.go)"},
	{"labDeviceItemMasterService.CreateDevice", statusGuarded, []string{"ExamTypeID"}, "internal/medicalrecord/lab_device_item_master_service.go: CreateDevice calls validateExamType → repo.FindExamType(ctx, clinicID, *examTypeID) before CreateDevice; test: TestLabDeviceService_CreateUpdateAndIsolation"},
	{"labDeviceItemMasterService.SaveConfiguration", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "internal/medicalrecord/lab_device_item_master_service.go: SaveConfiguration uses one transaction, validates the device ExamTypeID and every item ExamTypeFieldID with clinic-scoped lookups before persisting; test: TestLabDeviceItemMasterService_SaveConfigurationRollsBackAllChanges"},
	{"labDeviceItemMasterService.Update", statusGuarded, []string{"ExamTypeFieldID"}, "internal/medicalrecord/lab_device_item_master_service.go: Update calls repo.FindExamTypeField(ctx, clinicID, *ExamTypeFieldID) before persist; test: TestLabDeviceItemMasterService_UpdateAndResolve"},
	{"labDeviceItemMasterService.UpdateDevice", statusGuarded, []string{"ExamTypeID"}, "internal/medicalrecord/lab_device_item_master_service.go: UpdateDevice calls validateExamType → repo.FindExamType(ctx, clinicID, *examTypeID) before UpdateDevice; test: TestLabDeviceItemMasterService_UpdateDevice_RejectsCrossClinicExamType"},
	{"labImportExaminationService.PersistBatch", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "internal/medicalrecord/lab_import_examination_service.go: PersistBatch delegates each row to persistExam, which FindByID-guards ExamTypeID then requireOwnedExamTypeFields against examType.Items for nested ExamTypeFieldID; tests: TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamType, TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamTypeField"},
	{"labImportExaminationService.PersistExam", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "internal/medicalrecord/lab_import_examination_service.go: PersistExam is persistExam; ExamTypeID via examTypeRepo.FindByID(ctx, clinicID, ExamTypeID), nested Items[].ExamTypeFieldID via requireOwnedExamTypeFields (same membership check as examination replaceItemsTx); tests: TestLabImportExaminationService_PersistExam_RejectsCrossClinicExamType, TestLabImportExaminationService_PersistExam_RejectsCrossClinicExamTypeField"},
	// Issue #249 R-3: IsDuplicate takes LabExamPersistInput (now also carries nested ExamTypeFieldID)
	// but only filters existing exams by clinic_id+exam_type_id+date+payload — it does not INSERT/UPDATE
	// exam_type_id or exam_type_field_id. Ownership stays on persistExam.
	{"LabImportDuplicateCheckerDB.IsDuplicate", statusExempt, []string{"ExamTypeFieldID", "ExamTypeID"}, "internal/medicalrecord/lab_import_repository.go: read-only full-identical duplicate probe; ExamTypeID is a candidate filter only and ExamTypeFieldID is unused in the payload match (no persist of either FK). Write path: labImportExaminationService.persistExam."},
	{"labResultImportService.Commit", statusGuarded, []string{"ExamTypeFieldID", "ExamTypeID"}, "internal/medicalrecord/lab_result_import_service.go: Commit delegates to PersistBatch/persistExam, which guards ExamTypeID and nested ExamTypeFieldID; tests: TestLabResultImportService_Commit_RejectsCrossClinicExamType, TestLabResultImportService_Commit_RejectsCrossClinicExamTypeField"},
	{"medicalRecordService.CreateSubRecords", statusGuarded, []string{"ChiefComplaintTypeID", "Diagnosis1CategoryID", "Diagnosis1NameID", "Diagnosis2TypeID", "Diagnosis2NameID"}, "medical_record_subrecords.go: chiefComplaintTypeRepo.FindByID before inquiry upsert; validateCreateSubRecordDiagnosisFKs (validateDiagnosisFKs-equivalent, unexported local helper) guards all four diagnosis FKs before clinicalPlanRepo.Update (X-14 batch U4); best-effort — failure skips the write (Warn), does not error. test: TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicChiefComplaintType, TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicDiagnosisFK"},
	{"medicineService.Create", statusGuarded, []string{"InventoryID", "ParentID"}, "internal/medicalrecord/medicine_service.go (BE9-2D ⑥, moved): validateParentOwnership (self-ref repo.FindByID) + validateInventoryOwnership (inventoryRepo.FindByID) before persist (X-14 batch U2); test: TestMedicineService_Create_RejectsCrossClinicParentFK, TestMedicineService_Create_RejectsCrossClinicInventoryFK"},
	{"medicineService.Update", statusGuarded, []string{"InventoryID", "ParentID"}, "as Create — validateParentOwnership/validateInventoryOwnership guard *input.ParentID/*input.InventoryID before repo.Update (X-14 batch U2); test: TestMedicineService_Update_RejectsCrossClinicParentFK, TestMedicineService_Update_RejectsCrossClinicInventoryFK"},
	{"procedureService.Create", statusGuarded, []string{"ParentID"}, "internal/medicalrecord/procedure_service.go (BE9-2D ⑥, moved): validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestProcedureService_Create_RejectsCrossClinicParentFK"},
	{"procedureService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); test: TestProcedureService_Update_RejectsCrossClinicParentFK"},
	{"reservationRepository.CreateForTrimming", statusGuarded, []string{"ReservationTypeID"}, "BE9-2E-0 appointment write owner: assertTrimmingReservationType scopes id+category to clinic before insert; test: TestReservationRepository_CreateForTrimmingRejectsForeignReservationType"},
	{"reservationAdminService.Create", statusGuarded, []string{"ReservationTypeID"}, "appointment_admin_service.go:124 reservation.CheckReservationTypeCapacity(ctx, s.resRepo, s.typeRepo, ...) unconditionally calls typeRepo.FindByID(ctx, clinicID, ReservationTypeID) before persist (X-14 U6b); test: TestReservationAdminService_Create_RejectsCrossClinicReservationType"},
	{"reservationService.Create", statusGuarded, []string{"ReservationTypeID"}, "reservation_service.go: added unconditional s.typeRepo.FindByID(ctx, input.ClinicID, ReservationTypeID) before persist — closes the shortcut-route hole where reception/exam_room/record_shortcut routes (or advanced statuses) set enforceBookingConstraints=false and previously skipped the FindByID embedded in reservation.CheckReservationTypeCapacity (X-14 U6b); test: TestReservationService_Create_RejectsCrossClinicReservationType"},
	{"reservationService.CreateBatch", statusGuarded, []string{"ReservationTypeID"}, "reservation_service.go: CreateBatch validates ReservationTypeID via typeRepo.FindByID(ctx, input.ClinicID, ...) before its atomic transaction, then validates staff capability, owner/pet links, and each persisted reservation; tests: TestReservationService_CreateBatch_RejectsCrossClinicReservationType and batch atomicity coverage."},
	{"reservationService.Update", statusGuarded, []string{"ReservationTypeID"}, "reservation_service.go:389 updateWithConflictCheck unconditionally calls reservation.CheckReservationTypeCapacity → typeRepo.FindByID(ctx, clinicID, resolvedReservationTypeID) whenever ReservationTypeID changes (needsConflictCheck gate); pre-existing guard, dedicated isolation test added (X-14 U6b); test: TestReservationService_Update_RejectsCrossClinicReservationType"},
	{"reservationTypeService.Create", statusGuarded, []string{"GroupID", "ParentID"}, "ParentID validated via validateReservationTypeParent (pre-existing); GroupID now validated via new validateReservationTypeGroup → groupRepo.FindByID(ctx, clinicID, *GroupID) (X-14 U6b, groupRepo DI added to NewReservationTypeService); test: TestReservationTypeService_Create_RejectsCrossClinicGroupID"},
	{"reservationTypeService.Update", statusGuarded, []string{"GroupID", "ParentID"}, "as Create — validateReservationTypeGroup guards *input.GroupID before repo.Update; validateReservationTypeParent guards ParentID (X-14 U6b); test: TestReservationTypeService_Update_RejectsCrossClinicGroupID, TestReservationTypeService_Update_RejectsCrossClinicParentID"},
	{"liffService.CreateReservation", statusGuarded, []string{"ReservationTypeID", "TrimmingCourseID", "TrimmingOptionIDs"}, "liffService.CreateReservation delegates the whole write graph to reservationValidators.ValidateAndCreate; the validator checks active/public clinic-owned masters plus explicit staff capability and reservation visibility inside the write tx, then atomically writes appointment/detail/options; tests: TestLiffService_CreateReservation_RejectsCrossClinicTrimmingFK, TestReservationValidators_ValidateAndCreate_TrimmingWritesAreAtomicAndActive"},
	{"reservationValidators.ValidateAndCreate", statusGuarded, []string{"ReservationTypeID", "TrimmingCourseID", "TrimmingOptionIDs"}, "reservation_validators.go: type/course/option FindByID calls run in the ambient write tx, LIFF explicit staff must be active and reservation-visible, and concrete repositories hold SHARE locks through appointment/detail/options persistence (hard fail, no orphan appointment); tests: TestReservationValidators_ValidateAndCreate_RejectsCrossClinicReservationType, TestReservationValidators_ValidateAndCreate_RejectsCrossClinicTrimmingFK, TestReservationValidators_ValidateAndCreate_RejectsHiddenStaffDirectPost, TestReservationValidators_ValidateAndCreate_RejectsInactiveStaffDirectPost, TestReservationValidators_RealDBRejectsForeignStaffAndRollsBackTrimmingGraph"},
	{"staffService.Create", statusGuarded, []string{"OccupationID"}, "internal/staff/staff_service_core.go: lockOccupationOwnership calls occupationRepo.LockActiveByIDForShare(ctx, clinicID, *OccupationID) inside the write transaction before persist; tests: TestStaffService_Create_RejectsCrossClinicOccupationID (internal/staff/staff_cross_tenant_test.go), TestStaffServiceCore_Create_LocksOccupationInsideWriteTransaction (internal/staff/staff_service_core_test.go)"},
	{"staffService.CreateWithAccount", statusGuarded, []string{"OccupationID"}, "internal/staff/staff_service_account.go: the write transaction calls lockOccupationOwnership before account/staff/assignment writes; test: TestStaffService_CreateWithAccount_RejectsCrossClinicOccupationID (internal/staff/staff_cross_tenant_test.go)"},
	{"staffService.Update", statusGuarded, []string{"OccupationID"}, "internal/staff/staff_service_core.go: the write transaction locks staff, assignments, then the clinic-scoped occupation before repo.Update; tests: TestStaffService_Update_RejectsCrossClinicOccupationID (internal/staff/staff_cross_tenant_test.go), TestStaffServiceCore_Update_UsesStaffAssignmentOccupationLockOrder (internal/staff/staff_service_core_update_delete_test.go)"},
	{"treatmentService.Create", statusGuarded, []string{"ConsultationID", "InventoryID", "MedicineID", "ProcedureID"}, "internal/medicalrecord/treatment_service.go (BE9-2D ④b, moved from internal/service): validateTreatmentMasterFKs covers all four — medicine/procedure/consultation (pre-existing, 03bf1cb5) plus Inventory.FindByID(ctx, clinicID, InventoryID); DecreaseStock(ctx, clinicID, InventoryID, qty) independently scopes the atomic stock update (X-14a / INV-SEC P1); test: TestTreatmentService_Create_RejectsCrossClinicInventoryFK"},
	{"treatmentService.Update", statusGuarded, []string{"ConsultationID", "InventoryID", "MedicineID", "ProcedureID"}, "as Create — validateTreatmentMasterFKs guards InventoryID before persist (X-14a); test: TestTreatmentService_Update_RejectsCrossClinicInventoryFK"},
	{"trimmingCourseService.Update", statusGuarded, []string{"CourseTypeID"}, "internal/trimming/trimming_course_service.go (BE9-2E): Update now mirrors Create — courseTypeRepo.FindByID(ctx, clinicID, *CourseTypeID) before buildTrimmingCourseUpdate persists course_type_id (X-14b, symmetric with Create's guard); test: TestTrimmingCourseService_Update_RejectsCrossClinicCourseTypeFK"},
	{"trimmingService.Create", statusGuarded, []string{"CourseID", "OptionIDs", "ReservationTypeID"}, "internal/trimming/trimming_service.go (BE9-2E): fail-fast validation is repeated in the write tx; concrete type/course/option/staff repositories hold SHARE locks through appointment/detail/options persistence, and an explicitly referenced master/staff with a missing repository fails closed before writes; tests: TestTrimmingService_Create_RejectsCrossClinicCourseFK, TestTrimmingService_Create_RejectsCrossClinicOptionFK, TestTrimmingService_Create_RevalidatesStaffCapabilityInsideTransaction, TestTrimmingService_Create_MissingCourseRepositoryFailsClosed"},
	{"trimmingService.Update", statusGuarded, []string{"CourseID", "OptionIDs"}, "internal/trimming/trimming_service.go (BE9-2E): validateTrimmingCourseAndOptions guards *input.CourseID/*input.OptionIDs before persist and missing required repositories fail closed; tests: TestTrimmingService_Update_RejectsCrossClinicCourseFK, TestTrimmingService_Update_RejectsCrossClinicOptionFK, TestTrimmingService_Update_MissingOptionRepositoryFailsClosed"},
	{"vaccineService.Create", statusGuarded, []string{"ParentID"}, "internal/medicalrecord/vaccine_service.go (BE9-2D, moved from internal/service): validateParentOwnership FindByID(ctx, clinicID, *ParentID) before persist (X-14 batch3); test: TestVaccineService_Create_RejectsCrossClinicParentFK (internal/medicalrecord/cross_tenant_master_fk_write_test.go)"},
	{"vaccineService.Update", statusGuarded, []string{"ParentID"}, "as Create — validateParentOwnership guards *input.ParentID before repo.Update (X-14 batch3); internal/medicalrecord/vaccine_service.go; test: TestVaccineService_Update_RejectsCrossClinicParentFK"},
	{"writer.Create", statusGuarded, []string{"InsuranceID"}, "internal/pet/owner_registration.go: lockOwnerRegistrationMasters checks every non-nil InsuranceID by clinic under a SHARE lock before insert; test: TestPetRepository_Create_RejectsCrossClinicInsuranceAndRollsBackPet (internal/repository/owner_pet_create_write_owner_test.go)"},
	{"writer.CreateForOwnerRegistration", statusGuarded, []string{"InsuranceID"}, "internal/pet/owner_registration.go: the ambient transaction path delegates to lockOwnerRegistrationMasters before any nested pet insert; test: TestOwnerRepository_CreateWithPets_RejectsCrossClinicInsuranceAndRollsBackGraph (internal/repository/owner_pet_create_write_owner_test.go)"},
}

// ───────────────────────────────────────────────────────────────────────────────
// Analyzer (pure over a set of (filename → source) files, so fixtures and the real
// embedded package exercise identical logic).
// ───────────────────────────────────────────────────────────────────────────────

type mfkFieldEntry struct {
	names    []string // empty => embedded field
	typeExpr ast.Expr
}

// mfkStructKey identifies a locally-declared struct type by BOTH its containing directory
// (relative to internal/, e.g. "service", "repository/paymentmethod", "" for a bare fixture
// filename with no "/") and its bare type name. Keying by name alone would let two unrelated
// internal/ packages that happen to declare a same-named local struct (e.g. both defining
// "Input") contaminate each other's field sets — see the CROSS-PACKAGE TYPE-NAME COLLISION
// SAFETY note at the top of this file.
type mfkStructKey struct {
	dir  string
	name string
}

// mfkDirOf returns the directory portion of a lintscan-style forward-slash path (e.g.
// "repository/paymentmethod/repository.go" -> "repository/paymentmethod"), or "" for a
// bare filename with no directory component (the shape inline test fixtures use, all
// implicitly sharing the single "" package).
func mfkDirOf(filePath string) string {
	if i := strings.LastIndex(filePath, "/"); i >= 0 {
		return filePath[:i]
	}
	return ""
}

type mfkWriteFinding struct {
	receiver  string
	method    string
	file      string
	line      int
	masterFKs []string // sorted unique
}

type mfkExternalParam struct {
	dir       string
	qualifier string
	typeName  string
	receiver  string
	method    string
	file      string
	line      int
}

func (p mfkExternalParam) qualifiedType() string {
	return p.qualifier + "." + p.typeName
}

func (p mfkExternalParam) occurrence() string {
	parts := make([]string, 0, 3)
	if p.dir != "" {
		parts = append(parts, p.dir)
	}
	parts = append(parts, p.receiver, p.method)
	return strings.Join(parts, ".")
}

type mfkStats struct {
	filesParsed    int
	structsIndexed int
	methodsScanned int
	externalParams []mfkExternalParam
	// wholeTreeFilesDiscovered is the file count lintscan.WalkInternalTreeT(t) returned BEFORE
	// the service-write role filter (isServiceWriteRolePackage) narrowed it down. Populated only
	// by analyzeRealServiceSource (zero when analyzeServicePackage is called directly on inline
	// fixtures). Distinguishes "lintscan discovery broke" from "the role filter legitimately
	// narrowed a healthy whole-tree result" as two different failure modes — see the floor guard
	// in TestMasterFKWriteInventory_AllowlistMatchesRealSource.
	wholeTreeFilesDiscovered int
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

// qualifiedStructName returns the package qualifier and selected type name reachable from
// expr by peeling pointers and slices/arrays. Keeping the selected name (rather than recording
// only the qualifier) lets the external-param gate make exact type-and-occurrence exceptions.
func qualifiedStructName(expr ast.Expr) (qualifier, typeName string) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return qualifiedStructName(t.X)
	case *ast.ArrayType:
		return qualifiedStructName(t.Elt)
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name, t.Sel.Name
		}
	}
	return "", ""
}

// masterFKsOf returns the set of clinic-scoped master FK field names that (dir, structName)
// transitively contains. dir stays fixed across the whole recursion: an unqualified struct
// type name reached from a field of a struct declared in directory dir can, per real Go
// semantics, only resolve to another type declared in that SAME directory/package — never to a
// same-named type in a different internal/ package (see the CROSS-PACKAGE TYPE-NAME COLLISION
// SAFETY note at the top of this file). `visiting` guards self/mutual recursion (struct graph
// cycles), which the audit-taxonomy precedent never needed but a struct graph requires; it is
// keyed by mfkStructKey (not bare name) for the same collision-safety reason.
//
// NOT memoized on purpose: a result computed while truncating a cycle is incomplete, and
// caching it would poison later lookups (a false-negative). DTO graphs are acyclic and tiny
// (whole gate runs in <0.2s), so recomputing per top-level param is both correct and cheap.
func masterFKsOf(dir, structName string, index map[mfkStructKey][]mfkFieldEntry, visiting map[mfkStructKey]bool) map[string]struct{} {
	key := mfkStructKey{dir: dir, name: structName}
	if visiting[key] {
		return map[string]struct{}{} // cycle: contribution already accounted up the stack
	}
	visiting[key] = true

	out := map[string]struct{}{}
	for _, fe := range index[key] {
		// Direct scalar master FK field (name match + ID-shaped type).
		for _, n := range fe.names {
			if _, ok := clinicScopedMasterFKField[n]; ok && isIDType(fe.typeExpr) {
				out[n] = struct{}{}
			}
		}
		// Recurse into a same-package (same dir) struct-typed field (named or embedded). A
		// same-directory lookup is the only Go-semantics-correct resolution for an unqualified
		// type name — never search a different directory for it.
		if child, _ := localStructName(fe.typeExpr); child != "" {
			childKey := mfkStructKey{dir: dir, name: child}
			if _, ok := index[childKey]; ok {
				for k := range masterFKsOf(dir, child, index, visiting) {
					out[k] = struct{}{}
				}
			}
		}
	}

	delete(visiting, key)
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
	stats := mfkStats{}

	index := map[mfkStructKey][]mfkFieldEntry{}
	parsed := make([]*ast.File, 0, len(files))

	// Deterministic file order for stable diagnostics.
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	// Pass 1: parse + build struct index across all files, keyed by (dir, type name) so
	// same-named local structs in different internal/ packages cannot collide (see the
	// CROSS-PACKAGE TYPE-NAME COLLISION SAFETY note at the top of this file).
	for _, name := range names {
		f, err := parser.ParseFile(fset, name, files[name], 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		parsed = append(parsed, f)
		stats.filesParsed++
		dir := mfkDirOf(name)
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
			index[mfkStructKey{dir: dir, name: ts.Name.Name}] = fields
			stats.structsIndexed++
			return true
		})
	}

	// Pass 2: enumerate exported methods; classify by param containment.
	var findings []mfkWriteFinding
	for fi, f := range parsed {
		fname := names[fi]
		dir := mfkDirOf(fname)
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
						qualifiedParamQualifier, qualifiedParamType := qualifiedStructName(p.Type)
						stats.externalParams = append(stats.externalParams, mfkExternalParam{
							dir:       dir,
							qualifier: qualifiedParamQualifier,
							typeName:  qualifiedParamType,
							receiver:  recv,
							method:    fd.Name.Name,
							file:      baseName(fname),
							line:      fset.Position(fd.Pos()).Line,
						})
						continue
					}
					if child == "" {
						continue
					}
					// Unqualified type reference: resolve ONLY within this file's own
					// directory/package (mirrors real Go semantics) — never search a
					// different internal/ package's same-named local struct.
					childKey := mfkStructKey{dir: dir, name: child}
					if _, ok := index[childKey]; !ok {
						continue
					}
					for k := range masterFKsOf(dir, child, index, map[mfkStructKey]bool{}) {
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

// serviceWriteRolePackagePrefixes is the SINGLE extension point for widening which
// lintscan-discovered files this gate treats as "service-write role" source, i.e. eligible for
// the master-FK write review-coverage check. Matching is prefix-based against the
// lintscan-relative path key (e.g. "service/foo.go", "service/sub/deep/foo.go"), not
// depth-limited, so a subpackage nested under a listed prefix is admitted without any logic
// change. See the "SERVICE-WRITE ROLE SCOPE" note at the top of this file for why this scope is
// permanent rather than a placeholder.
//
// BE9-2 EXTENSION POINT: when service-write persistence logic migrates out of internal/service
// into a new domain package (e.g. internal/owner, internal/pet), add that package's lintscan
// prefix here — and ONLY here. Do not touch analyzeRealServiceSource's body, the analyzer
// (analyzeServicePackage / masterFKsOf), or knownSafeParamQualifiers to admit a new package; this
// var (and isServiceWriteRolePackage below, which consults it) is the one and only place that
// decides what counts as "in role scope" for this gate.
//
// "medicalrecord/" added BE9-2C: diagnosisNameService.Create/Update and
// examTypeService.Create/Update moved from internal/service to internal/medicalrecord (master-
// CRUD slice, boundary map §3.7 sub-batch ①). Their allowlist keys are unchanged (receiver
// type names carried over verbatim) — only the evidence comments below now point at the new
// file locations. "lstep/" was added when tag/checkup-sync write roles moved in L③.
var serviceWriteRolePackagePrefixes = []string{
	"service/",
	"medicalrecord/",
	"reservation/",
	"billing/",
	"lstep/",
	"trimming/",
	"inventory/",
	"pet/",
	"owner/",
	"staff/",
	"clinic/",
	"auth/",
}

// isServiceWriteRolePackage reports whether key — a lintscan.WalkInternalTreeT path key such as
// "service/foo.go" or "service/sub/deep/foo.go" — belongs to the service-write role scope this
// gate audits, per serviceWriteRolePackagePrefixes.
func isServiceWriteRolePackage(key string) bool {
	return matchesRolePackagePrefixes(key, serviceWriteRolePackagePrefixes)
}

// matchesRolePackagePrefixes is the prefix-matching primitive isServiceWriteRolePackage applies
// to the production serviceWriteRolePackagePrefixes list. It is factored out separately (rather
// than inlined into isServiceWriteRolePackage) so a test can prove the extension-point property —
// admitting a future BE9-2 domain package requires only a DATA addition to a prefix list, not a
// logic change — by exercising this SAME matching primitive against a locally constructed prefix
// list, without ever mutating the production serviceWriteRolePackagePrefixes var. See
// TestMasterFKWriteInventory_RoleFilterExtensionPointGeneralizes.
func matchesRolePackagePrefixes(key string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// analyzeRealServiceSource runs the analyzer over every production (non-test) .go file in the
// service-write role scope (isServiceWriteRolePackage), sourced via the shared lintscan package
// rather than the old package-local go:embed. Discovery walks the WHOLE module's internal/** tree
// (lintscan.WalkInternalTreeT) — a pure, package-structure-independent discovery step — and this
// function then narrows that whole-tree result down to the service-write role scope. That
// narrowing is the gate's permanent, by-design scope: see the "SERVICE-WRITE ROLE SCOPE" section
// at the top of this file for why internal/model and internal/handler are deliberately,
// permanently outside it. lintscan already excludes *_test.go files, testdata/, and vendor/, so
// no additional filtering for those is needed here.
func analyzeRealServiceSource(t *testing.T) ([]mfkWriteFinding, mfkStats) {
	t.Helper()
	byteFiles := WalkInternalTreeT(t)
	wholeTreeFiles := len(byteFiles)
	files := make(map[string]string, len(byteFiles))
	for name, src := range byteFiles {
		if !isServiceWriteRolePackage(name) {
			continue
		}
		files[name] = string(src)
	}
	findings, stats := analyzeServicePackage(t, files)
	stats.wholeTreeFilesDiscovered = wholeTreeFiles
	return findings, stats
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
// method. Floors prevent a vacuous green if the lintscan walk or AST matching silently breaks.
func TestMasterFKWriteInventory_AllowlistMatchesRealSource(t *testing.T) {
	findings, stats := analyzeRealServiceSource(t)

	// Pre-filter floor: proves lintscan.WalkInternalTreeT itself returned a healthy whole-tree
	// result BEFORE isServiceWriteRolePackage narrowed it. Distinguishes "discovery broke" from
	// "the role filter legitimately narrowed a healthy result" — a low pre-filter count means
	// discovery broke; a low post-filter count (below) with a healthy pre-filter count means the
	// role filter itself is misconfigured. 300 is comfortably below the real whole-tree count
	// (~748 non-test files under internal/ at last measurement) and comfortably above what the
	// service-only filter alone would ever produce (~202), so it cannot be satisfied by
	// accidentally discovering only the already-narrowed subset.
	if stats.wholeTreeFilesDiscovered < 300 {
		t.Fatalf("only %d whole-tree internal/** files discovered by lintscan.WalkInternalTreeT "+
			"before the service-write role filter narrowed them; lintscan discovery itself likely "+
			"broke (would vacuously pass)", stats.wholeTreeFilesDiscovered)
	}
	if stats.filesParsed < 100 {
		t.Fatalf("only %d non-test internal/service files parsed; lintscan walk or the "+
			"isServiceWriteRolePackage service-write role filter likely broken (would vacuously pass)", stats.filesParsed)
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

// TestMasterFKWriteInventory_StatusesAreLive pins the current audited state:
// every enumerated write is guarded and no known-unguarded residual remains.
func TestMasterFKWriteInventory_StatusesAreLive(t *testing.T) {
	counts := map[masterFKWriteStatus]int{}
	for _, e := range masterFKWriteAllowlist {
		counts[e.status]++
	}
	if counts[statusGuarded] == 0 {
		t.Error("no 'guarded' allowlist entries; the guarded write sites (treatment/vaccination/…) drifted")
	}
	if counts[statusKnownUnguarded] != 0 {
		t.Errorf(
			"known-unguarded master-FK writes were reintroduced: %d",
			counts[statusKnownUnguarded],
		)
	}
}

func isReviewedExternalParam(param mfkExternalParam) bool {
	if _, ok := knownSafeParamQualifiers[param.qualifier]; ok {
		return true
	}
	occurrences, ok := knownReviewedExternalParamOccurrences[param.qualifiedType()]
	if !ok {
		return false
	}
	_, ok = occurrences[param.occurrence()]
	return ok
}

// TestMasterFKWriteInventory_NoUnknownCrossPackageParam closes the cross-package false-negative
// gap: a service method receiving a struct from a package the gate cannot introspect could hide
// a master FK. Infrastructure-only qualifiers may be accepted wholesale; master-bearing command
// packages require an exact qualified type + write-entrypoint occurrence. Any other external
// parameter trips this gate and forces review.
func TestMasterFKWriteInventory_NoUnknownCrossPackageParam(t *testing.T) {
	_, stats := analyzeRealServiceSource(t)
	for _, param := range stats.externalParams {
		if !isReviewedExternalParam(param) {
			t.Errorf("%s:%d: %s receives unreviewed external parameter %s. If every %s type is "+
				"infrastructure-only, add the qualifier to knownSafeParamQualifiers; if the type can "+
				"carry a clinic-scoped master FK, review and pin only this exact type + occurrence.",
				param.file, param.line, param.occurrence(), param.qualifiedType(), param.qualifier)
		}
	}
}

func TestMasterFKWriteInventory_ExternalParamExceptionIsExactAndLive(t *testing.T) {
	reviewed := mfkExternalParam{
		dir:       "pet",
		qualifier: "owner",
		typeName:  "PetRegistrationIntent",
		receiver:  "OwnerRegistrationAdapter",
		method:    "CreateForOwnerRegistration",
	}
	if !isReviewedExternalParam(reviewed) {
		t.Fatal("the exact reviewed owner.PetRegistrationIntent adapter occurrence must be accepted")
	}

	for _, param := range []mfkExternalParam{
		{
			dir:       "pet",
			qualifier: "owner",
			typeName:  "FutureWriteCommand",
			receiver:  "OwnerRegistrationAdapter",
			method:    "CreateForOwnerRegistration",
		},
		{
			dir:       "pet",
			qualifier: "owner",
			typeName:  "PetRegistrationIntent",
			receiver:  "AnotherAdapter",
			method:    "CreateForOwnerRegistration",
		},
		{
			dir:       "pet",
			qualifier: "owner",
			typeName:  "PetRegistrationIntent",
			receiver:  "OwnerRegistrationAdapter",
			method:    "AnotherWrite",
		},
	} {
		if isReviewedExternalParam(param) {
			t.Errorf("owner exception is too broad: unexpectedly accepted %s at %s",
				param.qualifiedType(), param.occurrence())
		}
	}

	_, stats := analyzeRealServiceSource(t)
	observed := false
	for _, param := range stats.externalParams {
		if param.qualifiedType() == reviewed.qualifiedType() &&
			param.occurrence() == reviewed.occurrence() {
			observed = true
			break
		}
	}
	if !observed {
		t.Fatalf("reviewed external-param exception is stale: %s at %s was not observed",
			reviewed.qualifiedType(), reviewed.occurrence())
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

// TestMasterFKWriteInventory_AnalyzerDetectsViolationUnderNestedPathFilename proves the
// analyzer itself (not just the lintscan discovery mechanism — see
// TestMasterFKWriteInventory_AllowlistMatchesRealSource's floor guards for that) flags a
// known-bad master-FK write when the source is presented under filenames simulating different
// top-level internal/ packages and nesting depths — the shape lintscan.WalkInternalTreeT now
// produces for the WHOLE internal/** tree (not just internal/service). Mirrors
// repository/preload_clinic_scope_lint_test.go's
// TestPreloadClinicScope_AnalyzerDetectsViolationUnderNestedPathFilename, and doubles as proof
// that analyzeServicePackage's directory-scoped struct-key handles arbitrary top-level internal/
// packages correctly — useful both for internal/service's own possible future subdirectories and
// for whatever package a later serviceWriteRolePackagePrefixes addition brings into scope.
func TestMasterFKWriteInventory_AnalyzerDetectsViolationUnderNestedPathFilename(t *testing.T) {
	cases := []struct {
		name     string
		filename string
	}{
		{"top-level file directly under a different internal/ package", "handler/foo.go"},
		{"one level of nesting under a different internal/ package", "repository/paymentmethod/repository.go"},
		{"two levels of nesting under yet another internal/ package", "model/sub/deeper/thing.go"},
	}

	src := "package p\n" +
		`type T struct { MedicineID *uint64 }` + "\n" +
		`func (s *svc) Create(in *T) error { return nil }` + "\n"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, _ := analyzeServicePackage(t, map[string]string{tc.filename: src})
			if len(findings) != 1 {
				t.Fatalf("got %d findings for a nested-path-filename violation at %q, want 1: %+v", len(findings), tc.filename, findings)
			}
			if got := mfkKey(findings[0].receiver, findings[0].method); got != "svc.Create" {
				t.Fatalf("got key %s, want svc.Create", got)
			}
			if !equalStringSets(findings[0].masterFKs, []string{"MedicineID"}) {
				t.Fatalf("got fields %v, want [MedicineID]", findings[0].masterFKs)
			}
		})
	}
}

// TestMasterFKWriteInventory_CrossPackageTypeNameCollisionIsolation is the regression fixture
// for the directory-scoped mfkStructKey fix (see the CROSS-PACKAGE TYPE-NAME COLLISION SAFETY
// note at the top of this file). Two different simulated internal/ packages each declare a
// local struct named "Input": handler's has no master FK field, service's does. A flat
// (bare-name-only) index would let service.Input's MedicineID field leak into handler's index
// entry (false positive) — or, depending on map iteration/overwrite order, silently drop
// service.Input's real finding (false negative). The directory-scoped index must keep both
// packages' same-named "Input" struct completely isolated.
func TestMasterFKWriteInventory_CrossPackageTypeNameCollisionIsolation(t *testing.T) {
	files := map[string]string{
		"handler/foo.go": "package handler\n" +
			`type Input struct { Name string }` + "\n" +
			`func (h *fooHandler) Bind(in *Input) error { return nil }` + "\n",
		"service/bar.go": "package service\n" +
			`type Input struct { MedicineID *uint64 }` + "\n" +
			`func (s *barService) Create(in *Input) error { return nil }` + "\n",
	}

	findings, _ := analyzeServicePackage(t, files)
	got := map[string][]string{}
	for _, f := range findings {
		got[mfkKey(f.receiver, f.method)] = f.masterFKs
	}

	if fks, ok := got["fooHandler.Bind"]; ok {
		t.Errorf("fooHandler.Bind must NOT be flagged: handler.Input carries no master FK field; "+
			"a same-named service.Input struct in a DIFFERENT package incorrectly contaminated it "+
			"via a non-directory-scoped type-name index. got masterFKs=%v", fks)
	}
	if !equalStringSets(got["barService.Create"], []string{"MedicineID"}) {
		t.Errorf("barService.Create: got %v, want [MedicineID] (its own package's Input.MedicineID "+
			"field must still be detected despite the same-named handler.Input in a different package)",
			got["barService.Create"])
	}
}

// TestMasterFKWriteInventory_RoleFilterPackageScope pins isServiceWriteRolePackage directly:
// service/ paths (at any nesting depth — the filter is prefix-based, not depth-limited) are in
// role scope; model/, handler/, repository/, and lintscan/ paths are not. This is the Solution A
// permanent role filter's own unit-level proof, independent of the whole-tree integration fixture
// below.
func TestMasterFKWriteInventory_RoleFilterPackageScope(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{"service/foo.go", true},
		{"service/sub/deep/foo.go", true}, // prefix-based, not depth-limited
		{"owner/foo.go", true},
		{"model/foo.go", false},
		{"handler/foo.go", false},
		{"repository/foo.go", false},
		{"lintscan/foo.go", false},
	}
	for _, tc := range cases {
		if got := isServiceWriteRolePackage(tc.key); got != tc.want {
			t.Errorf("isServiceWriteRolePackage(%q) = %v, want %v", tc.key, got, tc.want)
		}
	}
}

// TestMasterFKWriteInventory_RoleFilterExtensionPointGeneralizes proves the BE9-2 extension-point
// claim in serviceWriteRolePackagePrefixes's doc comment: admitting a future domain package
// requires only a DATA addition to a prefix list, not a logic change. It builds a LOCAL copy of
// the prefix list (simulating a future BE9-2 change that appends "futureowner/"), feeds it through the
// SAME matchesRolePackagePrefixes primitive isServiceWriteRolePackage itself uses, and asserts
// the simulated-future path is now admitted while model/handler paths remain excluded — all
// without ever mutating the production serviceWriteRolePackagePrefixes var.
func TestMasterFKWriteInventory_RoleFilterExtensionPointGeneralizes(t *testing.T) {
	future := append(append([]string{}, serviceWriteRolePackagePrefixes...), "futureowner/")

	cases := []struct {
		key  string
		want bool
	}{
		{"futureowner/foo.go", true}, // admitted only via the simulated future prefix
		{"owner/foo.go", true},       // admitted via the production owner prefix
		{"service/foo.go", true},     // still admitted via the existing production prefix
		{"model/foo.go", false},
		{"handler/foo.go", false},
	}
	for _, tc := range cases {
		if got := matchesRolePackagePrefixes(tc.key, future); got != tc.want {
			t.Errorf("matchesRolePackagePrefixes(%q, future-prefixes) = %v, want %v", tc.key, got, tc.want)
		}
	}

	// The production var itself must be untouched by this simulation.
	for _, p := range serviceWriteRolePackagePrefixes {
		if p == "futureowner/" {
			t.Fatal("serviceWriteRolePackagePrefixes must not be mutated by this test")
		}
	}
	if isServiceWriteRolePackage("futureowner/foo.go") {
		t.Fatal(`isServiceWriteRolePackage("futureowner/foo.go") must still be false: the production ` +
			"var was not changed, only a local copy was")
	}
}

// TestMasterFKWriteInventory_RoleFilterIntegrationExcludesOtherLayers is the direct integration
// proof for the SERVICE-WRITE ROLE SCOPE decision: an identical violating shape (a struct field
// carrying a registered master FK, exposed via an exported method) is placed under three
// simulated top-level internal/ packages — service/, model/, and handler/ — in a fabricated file
// map shaped exactly like what lintscan.WalkInternalTreeT(t) would hand analyzeRealServiceSource.
// The SAME two-step pipeline analyzeRealServiceSource itself runs (filter via
// isServiceWriteRolePackage, then analyzeServicePackage) is applied here, and only the service/
// placement must be detected — model/ and handler/ must NOT false-positive despite carrying the
// byte-identical violating shape, proving the exclusion is a deliberate role-scope decision, not
// an accidental gap.
func TestMasterFKWriteInventory_RoleFilterIntegrationExcludesOtherLayers(t *testing.T) {
	violatingSrc := "package p\n" +
		`type T struct { MedicineID *uint64 }` + "\n" +
		`func (s *svc) Create(in *T) error { return nil }` + "\n"

	// Shape mirrors lintscan.WalkInternalTreeT(t)'s output: a flat map of lintscan-relative path
	// key -> file source, spanning multiple top-level internal/ packages.
	wholeTree := map[string]string{
		"service/violating.go": violatingSrc,
		"model/violating.go":   violatingSrc,
		"handler/violating.go": violatingSrc,
	}

	// Mirrors analyzeRealServiceSource's own two-step pipeline exactly, without touching the real
	// repository tree via lintscan.WalkInternalTreeT.
	filtered := make(map[string]string, len(wholeTree))
	for name, src := range wholeTree {
		if !isServiceWriteRolePackage(name) {
			continue
		}
		filtered[name] = src
	}

	findings, _ := analyzeServicePackage(t, filtered)
	if len(findings) != 1 {
		t.Fatalf("expected exactly 1 finding (service/ only; model/ and handler/ must be excluded "+
			"by role scope despite the identical violating shape), got %d: %+v", len(findings), findings)
	}
	if got := mfkKey(findings[0].receiver, findings[0].method); got != "svc.Create" {
		t.Fatalf("got key %s, want svc.Create", got)
	}
	if !equalStringSets(findings[0].masterFKs, []string{"MedicineID"}) {
		t.Fatalf("got fields %v, want [MedicineID]", findings[0].masterFKs)
	}
}
