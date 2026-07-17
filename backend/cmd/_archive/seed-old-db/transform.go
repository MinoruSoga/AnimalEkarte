package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Reserved crosswalk column names emitted by the old_db generator. They are
// join-only metadata (old pet_number + old record_no/sno), never inserted into
// the target table. Keep these in sync with CHILD_CROSSWALK in
// old_db/scripts/generate-new-schema-import-artifacts.mjs.
const (
	crosswalkPetNoCol    = "__xw_pet_no"
	crosswalkRecordNoCol = "__xw_record_no"
	crosswalkSnoCol      = "__xw_sno"
)

// legacyProcedureItemCode is the legacy treatments item_type code (Ttri_Kbn) that
// maps to the 'procedure' enum label. It is the ONLY item_type for which
// chk_treatment_item_ref permits a non-NULL procedure_id, so the same constant
// gates both treatmentItemTypeExpr (the enum cast) and the procedure_id NULLing
// in normalizeArgs — keeping the two in lockstep if the mapping ever changes.
const legacyProcedureItemCode = "1"

// hachiojiClinicName is the canonical name of the single clinic every old_db row
// belongs to. The legacy export stores clinic_id as branch/section codes
// {01,05,06,07} that DO NOT exist as clinics(id) in the new schema; all old_db
// data is from the 八王子 hospital, so clinic_id is resolved to this base-seeded
// clinic BY NAME (see hachiojiClinicIDExpr), not by trusting the legacy code.
//
// Source of truth: backend/migrations/003_seed_demo.sql inserts
// `(1, 1, '八王子病院')`. TestHachiojiClinicNameMatchesBaseSeed pins this string
// to that seed so the two never drift.
const hachiojiClinicName = "八王子病院"

// crosswalkFKColumn returns the resolved FK column a recoverable child-detail
// table fills from its crosswalk columns, plus the cache that resolves it.
// Returns ("", nil) for tables that do not use the crosswalk. The cache is
// chosen by the parent the child links to:
//   - children of medical_records → cache.MedicalRecords (medical_record_id)
//   - billing_items (two-hop)      → cache.Billings       (billing_id)
//   - exam_results (two-hop)       → cache.Exams          (exam_id)
func crosswalkFKColumn(entry ManifestEntry) (string, func(*lookupCache) map[string]int64) {
	switch entry.TargetTable {
	case "vital_records", "clinical_plans", "inquiries", "treatments":
		return "medical_record_id", func(c *lookupCache) map[string]int64 { return c.MedicalRecords }
	case "billing_items":
		return "billing_id", func(c *lookupCache) map[string]int64 { return c.Billings }
	case "exam_results":
		return "exam_id", func(c *lookupCache) map[string]int64 { return c.Exams }
	}
	return "", nil
}

func isCrosswalkColumn(col string) bool {
	return col == crosswalkPetNoCol || col == crosswalkRecordNoCol || col == crosswalkSnoCol
}

// crosswalkNeedsPetID reports whether a crosswalk child also requires a resolved
// pet_id column (NOT NULL on the target) that the TSV does not carry directly.
// Only vital_records does — the other children reach pets transitively through
// their parent FK and have no pet_id column of their own.
func crosswalkNeedsPetID(entry ManifestEntry) bool {
	return entry.TargetTable == "vital_records"
}

// withCrosswalkFKColumn returns the read-column list extended with the resolved
// columns for crosswalk children: the parent FK (medical_record_id/billing_id/
// exam_id) and, for vital_records, pet_id. These appended columns have no TSV
// source position; normalizeArgs fills them from the caches. For non-crosswalk
// entries the input is returned unchanged.
//
// Read columns keep the crosswalk metadata (__xw_*) so normalizeArgs can see it;
// insertColumns (below) is what actually goes into the INSERT and excludes the
// crosswalk metadata while keeping the resolved columns.
func withCrosswalkFKColumn(entry ManifestEntry, cols []string) []string {
	fkCol, _ := crosswalkFKColumn(entry)
	if fkCol == "" {
		return cols
	}
	out := make([]string, len(cols), len(cols)+2)
	copy(out, cols)
	out = append(out, fkCol)
	if crosswalkNeedsPetID(entry) {
		out = append(out, "pet_id")
	}
	return out
}

// insertColumns returns, for the given read-column list, the columns that belong
// in the INSERT (crosswalk metadata removed) and the source index in the read
// args for each. The resolved FK column maps to its own slot. Synthetic columns
// are added separately by syntheticColumns.
func insertColumns(cols []string) (names []string, srcIdx []int) {
	for i, c := range cols {
		if isCrosswalkColumn(c) {
			continue // join-only metadata — never inserted
		}
		names = append(names, c)
		srcIdx = append(srcIdx, i)
	}
	return names, srcIdx
}

func normalizeArgs(entry ManifestEntry, cols []string, args []any, cache *lookupCache) bool {
	if cache == nil {
		return false
	}

	// medical_records stores record_no PET-QUALIFIED ("<pet_no>-<record_no>") so it
	// is unique within the single synthetic clinic all old_db rows resolve to (legacy
	// record_no is unique only per pet). Do this BEFORE pet_id is converted to the new bigint below,
	// reading the old pet_no from the pet_id slot. The qualified value is also the
	// composite cache key every child resolves against.
	if entry.TargetTable == "medical_records" {
		petNo := stringArg(cols, args, "pet_id")
		for i, col := range cols {
			if col == "record_no" {
				if rec, _ := args[i].(string); rec != "" && petNo != "" {
					args[i] = qualifyRecordNo(petNo, rec)
				}
			}
		}
	}

	// Child-detail crosswalk resolution: read (__xw_pet_no, __xw_record_no/sno),
	// resolve the parent's NEW id from the composite cache via the qualified
	// record_no, and write it into the resolved FK slot. Missing parent → skip the
	// row (never insert an orphan).
	if fkCol, pickCache := crosswalkFKColumn(entry); fkCol != "" {
		var petNo, recordNo string
		for i, col := range cols {
			switch col {
			case crosswalkPetNoCol:
				petNo, _ = args[i].(string)
			case crosswalkRecordNoCol, crosswalkSnoCol:
				// A child TSV carries exactly one of __xw_record_no / __xw_sno
				// (KARTE/TRI children use record_no, KNS children use sno), so these
				// share a branch; both feed the same composite-key position.
				recordNo, _ = args[i].(string)
			}
		}
		if petNo == "" || recordNo == "" {
			return true // no crosswalk key → cannot link → skip
		}
		id, ok := pickCache(cache)[qualifyRecordNo(petNo, recordNo)]
		if !ok {
			return true // parent not loaded → skip rather than mis-link/orphan
		}
		resolved := strconv.FormatInt(id, 10)

		// vital_records also needs a NOT NULL pet_id; resolve the new bigint from
		// the old __xw_pet_no via the Pets cache. A missing pet → skip (the FK is
		// NOT NULL, so we must not insert a NULL).
		var resolvedPetID string
		if crosswalkNeedsPetID(entry) {
			pet, ok := cache.Pets[petNo]
			if !ok {
				return true
			}
			resolvedPetID = strconv.FormatInt(pet.ID, 10)
		}

		for i, col := range cols {
			switch col {
			case fkCol:
				args[i] = resolved
			case "pet_id":
				if resolvedPetID != "" {
					args[i] = resolvedPetID
				}
			}
		}
	}

	for i, col := range cols {
		raw, _ := args[i].(string)
		if raw == "" {
			continue
		}
		switch {
		case col == "pet_id" && usesOldPetNumber(entry):
			if pet, ok := cache.Pets[raw]; ok {
				args[i] = strconv.FormatInt(pet.ID, 10)
			} else {
				args[i] = nil
			}
		case col == "medical_record_id" && usesOldRecordNo(entry):
			// Parent billings/exams resolve their own medical_record_id from the
			// old per-pet record number via the composite cache. The cache key is
			// the qualified record_no; the pet side comes from the __xw_pet_no the
			// parent TSV carries.
			petNo := crosswalkPetNoFromArgs(cols, args)
			if id, ok := cache.MedicalRecords[qualifyRecordNo(petNo, raw)]; ok {
				args[i] = strconv.FormatInt(id, 10)
			} else {
				args[i] = nil
			}
		case entry.TargetTable == "pets" && col == "owner_id":
			id, ok := parseInt64(raw)
			if !ok || !cache.OwnerIDs[id] {
				return true
			}
		case entry.TargetTable == "pets" && col == "animal_species_id":
			id, ok := parseInt64(raw)
			if ok && cache.AnimalSpeciesIDs[id] {
				continue
			}
			if cache.FallbackAnimalSpeciesID == 0 {
				return true
			}
			args[i] = strconv.FormatInt(cache.FallbackAnimalSpeciesID, 10)
		}
	}

	// treatments.chk_treatment_item_ref CHECK: only item_type='procedure' may have
	// a non-NULL procedure_id (consultation/medicine/other require it NULL). The
	// legacy Ttri_Kbn code maps '1'→procedure, '2'→medicine, else→other (see
	// treatmentItemTypeExpr). NULL out procedure_id whenever item_type ≠ procedure
	// so the resolved legacy item code does not violate the CHECK.
	if entry.TargetTable == "treatments" {
		itemType := stringArg(cols, args, "item_type")
		if itemType != legacyProcedureItemCode { // non-procedure → procedure_id must be NULL
			for i, col := range cols {
				if col == "procedure_id" {
					args[i] = nil
				}
			}
		}
	}

	return false
}

// crosswalkPetNoFromArgs returns the raw __xw_pet_no value from the row args, or
// "" if the column is absent.
func crosswalkPetNoFromArgs(cols []string, args []any) string {
	return stringArg(cols, args, crosswalkPetNoCol)
}

// stringArg returns the string value at the position of col in cols, or "".
func stringArg(cols []string, args []any, col string) string {
	for i, c := range cols {
		if c == col && i < len(args) {
			s, _ := args[i].(string)
			return s
		}
	}
	return ""
}

// buildParamExpr returns the full SQL expression for a parameter.
//
// The old DB export stores many domain values as legacy numeric codes, not
// PostgreSQL enum labels. It also stores FK-ish values such as pet_id as old
// pet_number strings. Keep the coercion here explicit so bad legacy values do
// not abort the whole seed run.
func buildParamExpr(paramN int, entry ManifestEntry, col string) string {
	ref := paramRef(paramN)
	table := entry.TargetTable

	if (col == "doctor_id" || col == "entered_by") && table == "medical_records" {
		return staffIDLookupExpr(ref)
	}
	if col == "insurance_id" {
		return insuranceIDLookupExpr(ref)
	}
	// Legacy item codes (TChri_No) map to procedures/merchandise_items, but not
	// every code exists in the loaded master tables. Both FK columns are nullable
	// (ON DELETE SET NULL); resolve to the id only if it exists, else NULL, so an
	// unknown legacy code does not abort the whole child load with an FK error.
	if col == "procedure_id" {
		return existsIDLookupExpr(ref, "procedures")
	}
	if col == "merchandise_item_id" {
		return existsIDLookupExpr(ref, "merchandise_items")
	}

	switch col {
	case "dm_preference":
		return dmPreferenceExpr(ref)
	case "is_dangerous", "is_insurance", "is_insurance_applicable":
		return boolExpr(ref)
	case "birth_date", "neutered_date", "date", "scheduled_date",
		"next_visit_recommended_date":
		if (col == "date" && (table == "medical_records" || table == "exams")) || (col == "scheduled_date" && table == "billings") {
			return fmt.Sprintf("COALESCE(%s, DATE '1900-01-01')", dateExpr(ref))
		}
		return dateExpr(ref)
	case "created_at", "deleted_at", "deceased_at":
		if col == "created_at" {
			return fmt.Sprintf("COALESCE(%s, now())", timestampExpr(ref))
		}
		return timestampExpr(ref)
	case "id", "owner_id", "pet_id", "medical_record_id",
		"billing_id", "exam_id",
		"doctor_id", "entered_by", "exam_type_id", "parent_id", "animal_species_id",
		"procedure_id", "merchandise_item_id", "insurance_id", "exam_type_field_id",
		"chief_complaint_type_id":
		// clinic_id is intentionally NOT handled here: it is dropped from the TSV
		// columns (allowedColumns) and supplied as a synthetic column resolved by
		// hachiojiClinicIDExpr, so the legacy branch code is never inserted.
		return bigintExpr(ref)
	case "sort_order", "heart_rate", "respiration_rate":
		if col == "sort_order" {
			return fmt.Sprintf("COALESCE(%s, 0)", intExpr(ref))
		}
		return intExpr(ref)
	case "unit_price", "discount_amount", "total_amount":
		return fmt.Sprintf("COALESCE(%s, 0)", bigintExpr(ref))
	case "discount_rate", "temperature", "weight":
		return numericExpr(ref)
	case "quantity":
		// treatments/billing_items enforce CHECK (quantity > 0); a legacy 0/NULL/
		// invalid value must become a positive default, not 0.
		return fmt.Sprintf("GREATEST(COALESCE(%s, 1), 1)", numericExpr(ref))
	// ENUM columns — cast to the correct type based on table+column
	case "status":
		switch table {
		case "medical_records":
			return medicalRecordStatusExpr(ref)
		case "billings":
			return billingStatusExpr(ref)
		case "pets":
			return petStatusExpr(ref)
		case "exams":
			return examStatusExpr(ref)
		}
		return textExpr(ref)
	case "visit_type":
		return visitTypeExpr(ref)
	case "gender":
		return petGenderExpr(ref)
	case "acquisition_type":
		return acquisitionTypeExpr(ref)
	case "danger_level":
		return dangerLevelExpr(ref)
	case "membership_type":
		return membershipTypeExpr(ref)
	case "item_type":
		return treatmentItemTypeExpr(ref)
	case "tax_type":
		return taxTypeExpr(ref)
	case "category":
		return itemCategoryExpr(ref)
	case "source":
		return itemSourceExpr(ref)
	default:
		return textExpr(ref)
	}
}

func paramRef(paramN int) string {
	return fmt.Sprintf("$%d", paramN)
}

func textExpr(ref string) string {
	return fmt.Sprintf("COALESCE(%s::text, '')", ref)
}

func boolExpr(ref string) string {
	return fmt.Sprintf("(COALESCE(NULLIF(%s::text, ''), '0') IN ('1', 'true', 't', 'yes', 'y'))", ref)
}

// dmPreferenceExpr maps the legacy MST_SIIK_INFO.SiiDM_Kbn ("DM発送区分") code to
// owners.dm_preference (boolean NULL). Per table_definition_v4 codebook:
//
//	'01' = DM発送可   → true
//	'02' = DM発送不可 → false
//	anything else / empty → NULL (未設定, unknown-safe; matches migration 006 intent —
//	                              never fabricate a preference that was not recorded)
//
// Leading-zero codes are matched directly; the trimmed numeric forms ('1'/'2')
// are accepted defensively in case an upstream export strips the zero padding.
// The explicit ::boolean keeps the cast-suffix convention shared by every other
// typed *Expr helper, so the target DDL type is visible at the call site.
func dmPreferenceExpr(ref string) string {
	return fmt.Sprintf(
		"(CASE WHEN %s::text IN ('01', '1') THEN true "+
			"WHEN %s::text IN ('02', '2') THEN false ELSE NULL END)::boolean",
		ref, ref,
	)
}

func bigintExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^-?[0-9]+$' THEN (%s::text)::bigint ELSE NULL END)", ref, ref)
}

func intExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^-?[0-9]+$' THEN (%s::text)::integer ELSE NULL END)", ref, ref)
}

func numericExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^-?[0-9]+(\\.[0-9]+)?$' THEN (%s::text)::numeric ELSE NULL END)", ref, ref)
}

func dateExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}$' THEN (%s::text)::date ELSE NULL END)", ref, ref)
}

func timestampExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}' THEN (%s::text)::timestamptz ELSE NULL END)", ref, ref)
}

func staffIDLookupExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^[0-9]+$' THEN (SELECT s.id FROM public.staffs s WHERE s.id = (%s::text)::bigint LIMIT 1) ELSE NULL END)", ref, ref)
}

func insuranceIDLookupExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text ~ '^[0-9]+$' THEN (SELECT i.id FROM public.insurances i WHERE i.id = (%s::text)::bigint LIMIT 1) ELSE NULL END)", ref, ref)
}

// hachiojiClinicIDExpr resolves clinic_id to the single base-seeded 八王子病院
// clinic, ignoring the legacy branch code in the TSV entirely.
//
// The legacy clinic_id values {01,05,06,07} are section codes that do not exist
// as clinics(id) in the new schema (the base seed creates only 1=八王子病院,
// 2, 3); casting them straight to bigint would write 5/6/7 and violate the
// NOT NULL FK owners.clinic_id → clinics(id) (the empty-owners.clinic_id bug).
// Because every old_db row belongs to the 八王子 hospital, resolution is by NAME
// against the already-seeded clinics table — deterministic, tied to the existing
// seed path, and free of any hardcoded numeric id.
//
// The COALESCE fallback to MIN(id) handles the case where the named clinic is
// absent (e.g. a renamed base seed): it lands on the lowest-id clinic instead.
// This is a belt-and-suspenders default, not a guarantee — it returns NULL if the
// clinics table is entirely empty (migrations/base seed not applied), and binds to
// whatever clinic happens to have the smallest id if 八王子病院 is missing. The
// verify script asserts the name binding post-seed, so a wrong/NULL fallback is
// caught there rather than silently shipped.
//
// It takes no parameter: the legacy clinic_id column is dropped from the INSERT
// (removed from allowedColumns for owners/pets) and clinic_id is supplied entirely
// as a synthetic column whose value is this subquery. So every table resolves the
// same single clinic the same way, and the unreliable legacy code is never read.
func hachiojiClinicIDExpr() string {
	return fmt.Sprintf(
		"COALESCE((SELECT c.id FROM public.clinics c WHERE c.name = '%s' ORDER BY c.id LIMIT 1), (SELECT MIN(id) FROM public.clinics))",
		hachiojiClinicName,
	)
}

// existsIDLookupExpr resolves a legacy numeric id to itself only if a row with
// that id exists in the given master table, else NULL. Used for nullable FK
// columns whose legacy codes are not guaranteed to exist in the loaded masters
// (procedures, merchandise_items), so an unknown code becomes NULL instead of an
// FK-constraint error. table is a fixed internal literal, never user input.
func existsIDLookupExpr(ref, table string) string {
	return fmt.Sprintf(
		"(CASE WHEN %s::text ~ '^[0-9]+$' THEN (SELECT m.id FROM public.%s m WHERE m.id = (%s::text)::bigint LIMIT 1) ELSE NULL END)",
		ref, table, ref,
	)
}

func medicalRecordStatusExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('4', '5', 'finalized') THEN 'finalized' ELSE 'draft' END)::medical_record_status", ref)
}

func billingStatusExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN COALESCE(NULLIF(%s::text, ''), '') IN ('', '\\N', '.', '1', 'completed') THEN 'completed' ELSE 'waiting' END)::billing_status", ref)
}

func petStatusExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('1', 'deceased') THEN 'deceased' ELSE 'alive' END)::pet_status", ref)
}

func examStatusExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('completed', 'confirmed') THEN %s::text ELSE 'completed' END)::exam_status", ref, ref)
}

func visitTypeExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '1' THEN 'first' ELSE 'revisit' END)::visit_type", ref)
}

func petGenderExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('1', 'male') THEN 'male' WHEN %s::text IN ('2', 'female') THEN 'female' ELSE 'unknown' END)::pet_gender", ref, ref)
}

func acquisitionTypeExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '1' THEN 'purchased' WHEN %s::text = '2' THEN 'transferred' WHEN %s::text = '3' THEN 'rescued' ELSE 'other' END)::acquisition_type", ref, ref, ref)
}

func dangerLevelExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('1', 'high') THEN 'high' WHEN %s::text IN ('2', 'medium') THEN 'medium' ELSE 'low' END)::danger_level", ref, ref)
}

func membershipTypeExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '1' THEN 'member' WHEN %s::text = '2' THEN 'deceased' WHEN %s::text = '3' THEN 'transferred' ELSE 'non_member' END)::membership_type", ref, ref, ref)
}

func treatmentItemTypeExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '%s' THEN 'procedure' WHEN %s::text = '2' THEN 'medicine' ELSE 'other' END)::treatment_item_type", ref, legacyProcedureItemCode, ref)
}

func taxTypeExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '1' THEN 'included' WHEN %s::text = '3' THEN 'exempt' ELSE 'excluded' END)::tax_type", ref, ref)
}

func itemCategoryExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text = '1' THEN 'procedure' WHEN %s::text = '2' THEN 'medicine' WHEN %s::text = '3' THEN 'goods' ELSE 'other' END)::item_category", ref, ref, ref)
}

func itemSourceExpr(ref string) string {
	return fmt.Sprintf("(CASE WHEN %s::text IN ('medical_record', 'manual', 'hospitalization', 'trimming') THEN %s::text ELSE 'manual' END)::item_source", ref, ref)
}

func usesOldPetNumber(entry ManifestEntry) bool {
	return entry.TargetTable == "medical_records" ||
		entry.TargetTable == "billings" ||
		entry.TargetTable == "exams"
}

func usesOldRecordNo(entry ManifestEntry) bool {
	return entry.TargetTable == "billings" || entry.TargetTable == "exams"
}

// clinicScopedTables lists every target table whose clinic_id is supplied as a
// SYNTHETIC column resolved to the single 八王子病院 clinic (hachiojiClinicIDExpr),
// rather than read from the TSV. owners and pets DO carry a legacy clinic_id
// column, but its values are unreliable branch codes {01,05,06,07} that are not
// valid clinics(id); allowedColumns drops that column for them so this synthetic
// resolver is the sole source of clinic_id. The remaining tables carry no clinic
// key at all. Centralising the list keeps every table on one deterministic,
// no-magic-number resolution path.
var clinicScopedTables = map[string]bool{
	"owners":            true,
	"pets":              true,
	"exam_types":        true,
	"procedures":        true,
	"merchandise_items": true,
	"staffs":            true,
	"medical_records":   true,
	"billings":          true,
	"exams":             true,
	"vital_records":     true,
}

// syntheticColumns returns (cols, exprs) for NOT NULL columns absent from the TSV.
// filteredCols is the actual ordered column list from the TSV header (after filtering).
// exprs are SQL literals or subqueries embedded in the INSERT VALUES clause.
func syntheticColumns(entry ManifestEntry, filteredCols []string) ([]string, []string) {
	table := entry.TargetTable
	has := toSet(filteredCols) // use actual loaded cols, not manifest TargetColumns
	cols := []string{}
	exprs := []string{}

	// clinic_id: every clinic-scoped table resolves the single 八王子病院 clinic by
	// name. allowedColumns drops any legacy clinic_id from the TSV, so it is always
	// absent here and always supplied synthetically — uniform across tables.
	if clinicScopedTables[table] && !has["clinic_id"] {
		cols = append(cols, "clinic_id")
		exprs = append(exprs, hachiojiClinicIDExpr())
	}

	switch table {
	case "animal_species":
		if !has["name"] {
			cols = append(cols, "name")
			if has["id"] {
				idIdx := paramIndex(filteredCols, "id") + 1
				exprs = append(exprs, fmt.Sprintf("'legacy_species_' || %s::text", paramRef(idIdx)))
			} else {
				exprs = append(exprs, "'legacy_species'")
			}
		}
	case "clinics":
		// company_id → the base-seeded 本部 company (companies id=1, 002_seed_master).
		// This is the company, not the clinic, and has its own single-row default.
		if !has["company_id"] {
			cols = append(cols, "company_id")
			exprs = append(exprs, "1")
		}
	}
	return cols, exprs
}

// paramIndex returns the 0-based index of col in cols slice; -1 if not found.
func paramIndex(cols []string, col string) int {
	for i, c := range cols {
		if c == col {
			return i
		}
	}
	return -1
}

func shouldSkipRow(entry ManifestEntry, cols []string, args []any) bool {
	values := map[string]any{}
	for i, col := range cols {
		if i < len(args) {
			values[col] = args[i]
		}
	}

	switch entry.TargetTable {
	case "animal_species", "exam_types", "owners", "procedures", "merchandise_items":
		if v, ok := values["id"]; ok && !isIntegerValue(v) {
			return true
		}
	case "clinics":
		if v, ok := values["id"]; ok && !isIntegerValue(v) {
			return true
		}
		// Defense-in-depth against empty-name clinic rows. old_db has no real clinic
		// master; the only legacy sources that ever targeted clinics were the pet/owner
		// masters' hospital-code columns (QA-only crosswalk evidence), whose name is
		// always NULL. Loading them inserted empty-name orphan clinic rows (id = branch
		// code 05/06/07). The generator now drops those entries (isQaOnlyCrosswalkCandidate
		// in old_db), so no clinics TSV is produced — but this guard ensures the seeder
		// never inserts a nameless clinic even if a clinics entry resurfaces. clinics.name
		// is NOT NULL with no default, so a missing/blank name is never valid here.
		if !isTextValue(values["name"]) {
			return true
		}
	case "pets":
		if !isTextValue(values["pet_number"]) || !isIntegerValue(values["owner_id"]) {
			return true
		}
		if v, ok := values["animal_species_id"]; ok && !isIntegerValue(v) {
			return true
		}
	case "exam_type_fields":
		if !isIntegerValue(values["exam_type_id"]) {
			return true
		}
	}
	return false
}

func isTextValue(v any) bool {
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) != ""
}

func isIntegerValue(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == '-' && i == 0 {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseInt64(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// skipReason returns a non-empty string for entries that must be skipped.
func skipReason(entry ManifestEntry) string {
	table := entry.TargetTable
	has := toSet(entry.TargetColumns)

	// medical_records requires record_no (NOT NULL) and date (NOT NULL).
	// The MST_PET_INFO source only maps next_visit_recommended_date — not insertable standalone.
	if table == "medical_records" && !has["record_no"] && !has["date"] {
		return "medical_records: requires record_no and date (NOT NULL) — MST_PET_INFO source is update-only; skipped"
	}

	// exam_types::TBL_KNS_HIST has only 'name' (1.3M rows) — inserting without id
	// would create duplicates with no FK value (exam_type_fields reference by id).
	if table == "exam_types" && !has["id"] && !has["parent_id"] {
		return "exam_types: TBL_KNS_HIST source only has name column — would insert 1.3M duplicate rows with no FK; skipped"
	}

	// staffs::TBL_KARTE_DATA has only 'name' (425K rows) — inserting 425K staff records
	// with just names into clinic_id=1 is not useful for dev.
	if table == "staffs" && entry.LegacyTable == "TBL_KARTE_DATA" {
		return "staffs: TBL_KARTE_DATA source only has name (425K rows, no FK value) — would insert non-useful duplicates; skipped"
	}

	// TBL_TRI_DATA is line-item/detail-like data. Its medical_records projection
	// has only record_no/date and no pet or owner key. TBL_KARTE_DATA is the real
	// medical_records source (1:1, unique composite); loading TRI as a second
	// medical_records parent would create duplicate orphan records and corrupt the
	// composite cache the line-item children depend on. Stays skipped by design —
	// billings/treatments/billing_items already carry the TRI data via their own
	// plans and resolve the medical_record link through the KARTE-built cache.
	if table == "medical_records" && entry.LegacyTable == "TBL_TRI_DATA" {
		return "medical_records: SKIPPED by design — TBL_KARTE_DATA is the medical_records source; TBL_TRI_DATA rows reach the DB as billings/treatments/billing_items and link back via the composite crosswalk cache"
	}

	switch table {
	case "shared_files":
		return "shared_files: requires uploaded_by/file_type/file_key/file_size (NOT NULL) — not in TSV"
	case "medical_record_addenda":
		// author_user_id is NOT NULL REFERENCES staffs with NO legacy source — the
		// legacy chart carries no per-addendum staff key. The composite crosswalk
		// now resolves medical_record_id, but author_user_id cannot be filled
		// without fabricating an audit attribution that never existed. Permanently
		// BLOCKED by design until the legacy export carries a staff key.
		return "medical_record_addenda: BLOCKED — medical_record_id is now resolvable via the composite crosswalk, but author_user_id (NOT NULL → staffs) has no legacy source. Required input: a per-addendum legacy staff key; fabricating one is prohibited"
	case "vital_records", "clinical_plans", "inquiries", "treatments", "billing_items", "exam_results":
		// Recoverable: unblocked once the TSV carries the composite crosswalk key
		// (__xw_pet_no + __xw_record_no/__xw_sno). normalizeArgs resolves the
		// parent FK from the composite cache and skips any row whose parent is
		// absent, so no orphan is ever inserted.
		if !hasCrosswalkKey(entry) {
			return fmt.Sprintf(
				"%s: BLOCKED — missing composite crosswalk columns (%s + %s/%s). Required input: old_db generator emits the crosswalk key on this child TSV",
				table, crosswalkPetNoCol, crosswalkRecordNoCol, crosswalkSnoCol,
			)
		}
	}
	return ""
}

// hasCrosswalkKey reports whether the entry's TSV carries the composite crosswalk
// key: __xw_pet_no plus one of __xw_record_no / __xw_sno.
func hasCrosswalkKey(entry ManifestEntry) bool {
	xw := toSet(entry.CrosswalkColumns)
	return xw[crosswalkPetNoCol] && (xw[crosswalkRecordNoCol] || xw[crosswalkSnoCol])
}

// conflictClause returns ON CONFLICT handling.
//
// Every target uses DO NOTHING. For tables with a natural unique key (owners.id,
// pets.pet_number, medical_records, billings.medical_record_id, …) this makes
// re-runs idempotent and absorbs duplicate legacy rows. The child-detail /
// line-item tables (treatments, billing_items, exam_results, vital_records,
// clinical_plans, inquiries) have only a BIGSERIAL PK, so DO NOTHING is a no-op
// for them — re-running would duplicate their rows. Those tables are therefore
// TRUNCATEd up front (see truncateChildDetailTables) to keep the seed
// re-runnable. The signature is argument-free because conflict targets are
// uniform today; reintroduce parameters if a table needs ON CONFLICT (col) DO UPDATE.
func conflictClause() string {
	return " ON CONFLICT DO NOTHING"
}

// allowedColumns returns the set of target-table columns we accept from TSV.
// Returns nil for tables not in the map (accept all columns).
func allowedColumns(table string) map[string]bool {
	known := map[string]map[string]bool{
		"animal_species": {
			"id": true, "name": true,
		},
		"clinics": {
			"id": true, "name": true,
		},
		"owners": {
			// name_kana excluded: CHECK (name_kana !~ '[ァ-ヶ]') — old data may contain katakana; use DEFAULT ''
			// clinic_id excluded: the legacy value is an unreliable branch code
			// {01,05,06,07} that is not a valid clinics(id). Dropping it here lets
			// syntheticColumns supply the single 八王子病院 clinic via
			// hachiojiClinicIDExpr (see clinicScopedTables) — fixes empty clinic_id.
			"id": true, "name": true, "birth_date": true,
			"company": true, "postal_code": true, "address1": true, "address2": true,
			"home_postal_code": true, "home_address1": true, "home_address2": true,
			"phone": true, "company_phone": true, "email": true, "remarks": true,
			"is_dangerous": true, "discount_rate": true, "membership_type": true,
			// dm_preference: legacy MST_SIIK_INFO.SiiDM_Kbn ("DM発送区分"). Must be
			// whitelisted AND given a boolean cast in buildParamExpr (codes 01/02),
			// or it is silently dropped (migration 006 added it as boolean NULL).
			"dm_preference": true,
			"created_at":    true, "deleted_at": true,
		},
		"pets": {
			// name_kana excluded: same CHECK constraint as owners
			"pet_number": true, "deleted_at": true, "status": true, "name": true,
			"owner_id": true, "animal_species_id": true, "breed": true,
			"gender": true, "color": true, "birth_date": true, "deceased_at": true,
			"acquisition_type": true, "remarks": true, "neutered_date": true, "created_at": true,
			// clinic_id excluded for the same reason as owners: the legacy pets
			// clinic_id is a branch code {05,07}, not a valid clinics(id). Supplied
			// synthetically via hachiojiClinicIDExpr (clinicScopedTables) instead.
			"danger_level": true, "insurance_id": true, "weight": true,
			"environment": true, "food": true,
		},
		"staffs": {
			"name": true,
		},
		"exam_types": {
			"id": true, "parent_id": true, "description": true, "name": true,
		},
		"exam_type_fields": {
			"exam_type_id": true, "sort_order": true, "name": true, "normal_value": true,
		},
		"medical_records": {
			"pet_id": true, "record_no": true, "visit_type": true, "status": true,
			"doctor_id": true, "entered_by": true, "date": true,
			"next_visit_recommended_date": true, "recommendation_reason": true,
		},
		"billings": {
			"pet_id": true, "medical_record_id": true, "scheduled_date": true,
			"total_amount": true, "status": true,
			// Crosswalk columns must survive filtering so normalizeArgs can resolve
			// medical_record_id from (pet_no, record_no); stripped before INSERT.
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"merchandise_items": {
			"id": true, "name": true, "category": true, "unit_price": true,
			"sort_order": true, "tax_type": true,
		},
		"procedures": {
			"id": true, "name": true, "sort_order": true, "description": true,
		},
		"exams": {
			"pet_id": true, "medical_record_id": true, "date": true,
			"exam_type_id": true, "result_summary": true,
			// Crosswalk columns survive filtering for medical_record_id resolution;
			// stripped before INSERT.
			crosswalkPetNoCol: true, crosswalkSnoCol: true,
		},
		// ── Recoverable child-detail tables ──────────────────────────────────
		// Real attribute columns (from manifest targetColumns) plus the reserved
		// crosswalk columns, which must survive filtering so normalizeArgs can
		// read them; they are stripped from the INSERT by insertColumns(). The
		// resolved FK (medical_record_id/billing_id/exam_id) is appended by
		// withCrosswalkFKColumn, not sourced from the TSV, so it is not listed.
		"vital_records": {
			"temperature": true, "weight": true, "heart_rate": true, "respiration_rate": true,
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"clinical_plans": {
			"physical_exam": true, "diagnosis_details": true, "treatment_policy": true,
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"inquiries": {
			"chief_complaint_type_id": true, "chief_complaint": true,
			"owner_observations": true, "history": true, "notes": true,
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"treatments": {
			"sort_order": true, "procedure_id": true, "content": true, "unit_price": true,
			"discount_amount": true, "quantity": true, "item_type": true,
			"admin_route": true, "memo": true, "is_insurance": true,
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"billing_items": {
			"sort_order": true, "merchandise_item_id": true, "name": true, "unit_price": true,
			"discount_amount": true, "quantity": true, "tax_type": true, "category": true,
			"source": true, "is_insurance_applicable": true,
			crosswalkPetNoCol: true, crosswalkRecordNoCol: true,
		},
		"exam_results": {
			"sort_order": true, "name": true, "inspection_value": true, "result": true,
			"normal_value": true, "reference_value": true,
			crosswalkPetNoCol: true, crosswalkSnoCol: true,
		},
	}
	return known[table] // nil for unknown tables → accept all
}

// filterColumnsWithIndex returns (keptCols, keptIndexes, droppedCols).
func filterColumnsWithIndex(cols []string, allowed map[string]bool) ([]string, []int, []string) {
	if allowed == nil {
		indexes := make([]int, len(cols))
		for i := range cols {
			indexes[i] = i
		}
		return cols, indexes, nil
	}
	kept, indexes, dropped := []string{}, []int{}, []string{}
	for i, c := range cols {
		if allowed[c] {
			kept = append(kept, c)
			indexes = append(indexes, i)
		} else {
			dropped = append(dropped, c)
		}
	}
	return kept, indexes, dropped
}

// quoteIdent safely double-quotes a PostgreSQL identifier.
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func toSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}

// tableTier maps target table names to their load order tier.
// Lower tier numbers are loaded first to satisfy FK dependencies.
// Tables at the same tier are loaded in their original manifest order.
var tableTier = map[string]int{
	"animal_species":    1, // no FKs
	"clinics":           1, // company_id is synthetic literal
	"exam_types":        2, // clinic_id → clinics; parent_id self-ref (OK with ON CONFLICT)
	"owners":            2, // clinic_id → clinics
	"merchandise_items": 2, // clinic_id → clinics
	"procedures":        2, // clinic_id → clinics
	"staffs":            2, // clinic_id → clinics
	"exam_type_fields":  3, // exam_type_id → exam_types
	"pets":              3, // owner_id → owners, clinic_id → clinics
	"medical_records":   4, // pet_id → pets, clinic_id derived from pets
	"billings":          5, // medical_record_id → medical_records, pet_id → pets
	"exams":             5, // exam_type_id → exam_types, pet_id → pets
	// Child-detail tables resolve their parent FK from composite caches that are
	// refreshed only after the parent loads. medical_records children need the
	// medical_records cache (tier 4); billing_items/exam_results are two-hop and
	// need the billings/exams caches (tier 5), so all children load at tier 6.
	"vital_records":  6, // medical_record_id via medical_records cache
	"clinical_plans": 6, // medical_record_id via medical_records cache
	"inquiries":      6, // medical_record_id via medical_records cache
	"treatments":     6, // medical_record_id via medical_records cache
	"billing_items":  6, // billing_id via billings cache (two-hop)
	"exam_results":   6, // exam_id via exams cache (two-hop)
}

// orderByDependency sorts entries by FK tier so parents are always loaded before children.
func orderByDependency(entries []ManifestEntry) []ManifestEntry {
	out := make([]ManifestEntry, len(entries))
	copy(out, entries)
	// Stable sort preserves original manifest ordering within the same tier.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			tierA := tableTier[out[j-1].TargetTable]
			tierB := tableTier[out[j].TargetTable]
			if tierA == 0 {
				tierA = 99
			}
			if tierB == 0 {
				tierB = 99
			}
			if tierB < tierA {
				out[j-1], out[j] = out[j], out[j-1]
			} else {
				break
			}
		}
	}
	return out
}

func loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is operator config
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
