package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- buildParamExpr: exhaustive column/table branch coverage ---
//
// buildParamExpr is the single function that decides how every legacy TSV value
// is cast into the target column's SQL type. A wrong branch here either aborts
// the whole seed run (bad cast) or silently corrupts data (wrong enum mapping),
// so every switch arm gets its own assertion on the emitted SQL fragment.
func TestBuildParamExprBranches(t *testing.T) {
	tests := []struct {
		name  string
		table string
		col   string
		want  string // substring that must appear in the emitted SQL
	}{
		{"doctor_id on medical_records uses staff lookup", "medical_records", "doctor_id", "public.staffs s"},
		{"entered_by on medical_records uses staff lookup", "medical_records", "entered_by", "public.staffs s"},
		{"doctor_id on other table falls through to bigint", "billings", "doctor_id", "~ '^-?[0-9]+$'"},
		{"insurance_id uses insurance lookup", "pets", "insurance_id", "public.insurances i"},
		{"birth_date uses plain date cast", "owners", "birth_date", "::date"},
		{"neutered_date uses plain date cast", "pets", "neutered_date", "::date"},
		{"date on medical_records coalesces to sentinel", "medical_records", "date", "COALESCE"},
		{"date on exams coalesces to sentinel", "exams", "date", "COALESCE"},
		{"date on other table uses plain date cast", "checkups", "date", "::date"},
		{"scheduled_date on billings coalesces to sentinel", "billings", "scheduled_date", "COALESCE"},
		{"scheduled_date on other table uses plain date cast", "reservations", "scheduled_date", "::date"},
		{"next_visit_recommended_date uses plain date cast", "medical_records", "next_visit_recommended_date", "::date"},
		{"created_at coalesces to now()", "owners", "created_at", "now()"},
		{"deleted_at uses plain timestamp cast", "owners", "deleted_at", "::timestamptz"},
		{"deceased_at uses plain timestamp cast", "pets", "deceased_at", "::timestamptz"},
		{"sort_order coalesces to zero", "exam_type_fields", "sort_order", "COALESCE"},
		{"heart_rate uses plain int cast", "vital_records", "heart_rate", "::integer"},
		{"respiration_rate uses plain int cast", "vital_records", "respiration_rate", "::integer"},
		{"unit_price coalesces to zero bigint", "treatments", "unit_price", "COALESCE"},
		{"discount_amount coalesces to zero bigint", "treatments", "discount_amount", "COALESCE"},
		{"total_amount coalesces to zero bigint", "billings", "total_amount", "COALESCE"},
		{"discount_rate uses plain numeric cast", "owners", "discount_rate", "::numeric"},
		{"temperature uses plain numeric cast", "vital_records", "temperature", "::numeric"},
		{"weight uses plain numeric cast", "pets", "weight", "::numeric"},
		{"status on medical_records uses medical_record_status", "medical_records", "status", "::medical_record_status"},
		{"status on billings uses billing_status", "billings", "status", "::billing_status"},
		{"status on pets uses pet_status", "pets", "status", "::pet_status"},
		{"status on exams uses exam_status", "exams", "status", "::exam_status"},
		{"status on unrecognized table falls back to text", "owners", "status", "COALESCE"},
		{"visit_type uses visit_type enum", "medical_records", "visit_type", "::visit_type"},
		{"gender uses pet_gender enum", "pets", "gender", "::pet_gender"},
		{"acquisition_type uses acquisition_type enum", "pets", "acquisition_type", "::acquisition_type"},
		{"danger_level uses danger_level enum", "pets", "danger_level", "::danger_level"},
		{"membership_type uses membership_type enum", "owners", "membership_type", "::membership_type"},
		{"item_type uses treatment_item_type enum", "treatments", "item_type", "::treatment_item_type"},
		{"tax_type uses tax_type enum", "billing_items", "tax_type", "::tax_type"},
		{"category uses item_category enum", "billing_items", "category", "::item_category"},
		{"source uses item_source enum", "billing_items", "source", "::item_source"},
		{"unrecognized column falls back to text", "owners", "remarks", "COALESCE(\\$1::text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := ManifestEntry{TargetTable: tt.table}
			expr := buildParamExpr(1, entry, tt.col)
			want := strings.ReplaceAll(tt.want, "\\$", "$")
			if !strings.Contains(expr, want) {
				t.Fatalf("buildParamExpr(table=%s, col=%s) = %q, want substring %q", tt.table, tt.col, expr, want)
			}
		})
	}
}

// --- Individual *Expr builders: SQL shape + enum mapping regression coverage ---

func TestIntExprCastsIntegerOrNull(t *testing.T) {
	expr := intExpr(paramRef(1))
	if !strings.Contains(expr, "::integer") || !strings.Contains(expr, "ELSE NULL") {
		t.Fatalf("intExpr must cast to integer with NULL fallback, got: %q", expr)
	}
}

func TestDateExprCastsDateOrNull(t *testing.T) {
	expr := dateExpr(paramRef(1))
	if !strings.Contains(expr, "::date") || !strings.Contains(expr, "ELSE NULL") {
		t.Fatalf("dateExpr must cast to date with NULL fallback, got: %q", expr)
	}
}

func TestTimestampExprCastsTimestamptzOrNull(t *testing.T) {
	expr := timestampExpr(paramRef(1))
	if !strings.Contains(expr, "::timestamptz") || !strings.Contains(expr, "ELSE NULL") {
		t.Fatalf("timestampExpr must cast to timestamptz with NULL fallback, got: %q", expr)
	}
}

func TestStaffIDLookupExprResolvesExistingStaffOnly(t *testing.T) {
	expr := staffIDLookupExpr(paramRef(1))
	if !strings.Contains(expr, "FROM public.staffs s") || !strings.Contains(expr, "ELSE NULL") {
		t.Fatalf("staffIDLookupExpr must resolve via staffs table with NULL fallback, got: %q", expr)
	}
}

func TestInsuranceIDLookupExprResolvesExistingInsuranceOnly(t *testing.T) {
	expr := insuranceIDLookupExpr(paramRef(1))
	if !strings.Contains(expr, "FROM public.insurances i") || !strings.Contains(expr, "ELSE NULL") {
		t.Fatalf("insuranceIDLookupExpr must resolve via insurances table with NULL fallback, got: %q", expr)
	}
}

func TestMedicalRecordStatusExprMapsFinalizedCodes(t *testing.T) {
	expr := medicalRecordStatusExpr(paramRef(1))
	if !strings.Contains(expr, "'finalized'") || !strings.Contains(expr, "'draft'") || !strings.Contains(expr, "::medical_record_status") {
		t.Fatalf("medicalRecordStatusExpr must map to finalized/draft, got: %q", expr)
	}
}

func TestBillingStatusExprMapsBlankOrLegacyCodeToCompleted(t *testing.T) {
	expr := billingStatusExpr(paramRef(1))
	if !strings.Contains(expr, "'completed'") || !strings.Contains(expr, "'waiting'") || !strings.Contains(expr, "::billing_status") {
		t.Fatalf("billingStatusExpr must map to completed/waiting, got: %q", expr)
	}
}

func TestPetStatusExprMapsDeceasedCode(t *testing.T) {
	expr := petStatusExpr(paramRef(1))
	if !strings.Contains(expr, "'deceased'") || !strings.Contains(expr, "'alive'") || !strings.Contains(expr, "::pet_status") {
		t.Fatalf("petStatusExpr must map to deceased/alive, got: %q", expr)
	}
}

func TestExamStatusExprDefaultsToCompleted(t *testing.T) {
	expr := examStatusExpr(paramRef(1))
	if !strings.Contains(expr, "'completed'") || !strings.Contains(expr, "::exam_status") {
		t.Fatalf("examStatusExpr must default to completed, got: %q", expr)
	}
}

func TestVisitTypeExprMapsFirstVisitCode(t *testing.T) {
	expr := visitTypeExpr(paramRef(1))
	if !strings.Contains(expr, "'first'") || !strings.Contains(expr, "'revisit'") || !strings.Contains(expr, "::visit_type") {
		t.Fatalf("visitTypeExpr must map to first/revisit, got: %q", expr)
	}
}

func TestPetGenderExprMapsMaleFemaleCodes(t *testing.T) {
	expr := petGenderExpr(paramRef(1))
	if !strings.Contains(expr, "'male'") || !strings.Contains(expr, "'female'") || !strings.Contains(expr, "'unknown'") || !strings.Contains(expr, "::pet_gender") {
		t.Fatalf("petGenderExpr must map male/female/unknown, got: %q", expr)
	}
}

func TestAcquisitionTypeExprMapsAllCodes(t *testing.T) {
	expr := acquisitionTypeExpr(paramRef(1))
	for _, want := range []string{"'purchased'", "'transferred'", "'rescued'", "'other'", "::acquisition_type"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("acquisitionTypeExpr missing %q, got: %q", want, expr)
		}
	}
}

func TestDangerLevelExprMapsAllCodes(t *testing.T) {
	expr := dangerLevelExpr(paramRef(1))
	for _, want := range []string{"'high'", "'medium'", "'low'", "::danger_level"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("dangerLevelExpr missing %q, got: %q", want, expr)
		}
	}
}

func TestMembershipTypeExprMapsAllCodes(t *testing.T) {
	expr := membershipTypeExpr(paramRef(1))
	for _, want := range []string{"'member'", "'deceased'", "'transferred'", "'non_member'", "::membership_type"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("membershipTypeExpr missing %q, got: %q", want, expr)
		}
	}
}

func TestTreatmentItemTypeExprUsesLegacyProcedureCode(t *testing.T) {
	expr := treatmentItemTypeExpr(paramRef(1))
	if !strings.Contains(expr, "'"+legacyProcedureItemCode+"'") || !strings.Contains(expr, "'procedure'") ||
		!strings.Contains(expr, "'medicine'") || !strings.Contains(expr, "'other'") {
		t.Fatalf("treatmentItemTypeExpr must map legacy code %q to procedure, got: %q", legacyProcedureItemCode, expr)
	}
}

func TestTaxTypeExprMapsAllCodes(t *testing.T) {
	expr := taxTypeExpr(paramRef(1))
	for _, want := range []string{"'included'", "'exempt'", "'excluded'", "::tax_type"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("taxTypeExpr missing %q, got: %q", want, expr)
		}
	}
}

func TestItemCategoryExprMapsAllCodes(t *testing.T) {
	expr := itemCategoryExpr(paramRef(1))
	for _, want := range []string{"'procedure'", "'medicine'", "'goods'", "'other'", "::item_category"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("itemCategoryExpr missing %q, got: %q", want, expr)
		}
	}
}

func TestItemSourceExprAllowsKnownSourcesOnly(t *testing.T) {
	expr := itemSourceExpr(paramRef(1))
	for _, want := range []string{"'medical_record'", "'manual'", "'hospitalization'", "'trimming'", "::item_source"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("itemSourceExpr missing %q, got: %q", want, expr)
		}
	}
}

// --- conflictClause / quoteIdent ---

func TestConflictClauseIsDoNothing(t *testing.T) {
	if got := conflictClause(); got != " ON CONFLICT DO NOTHING" {
		t.Fatalf("conflictClause() = %q, want %q", got, " ON CONFLICT DO NOTHING")
	}
}

func TestQuoteIdentEscapesDoubleQuotes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain identifier", "owners", `"owners"`},
		{"identifier with embedded quote", `weird"col`, `"weird""col"`},
		{"empty identifier", "", `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quoteIdent(tt.in); got != tt.want {
				t.Fatalf("quoteIdent(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// --- withCrosswalkFKColumn: non-crosswalk pass-through ---

func TestWithCrosswalkFKColumnPassesThroughNonCrosswalkTable(t *testing.T) {
	entry := ManifestEntry{TargetTable: "owners"}
	cols := []string{"id", "name"}
	got := withCrosswalkFKColumn(entry, cols)
	if len(got) != len(cols) {
		t.Fatalf("non-crosswalk table must return cols unchanged, got %v want %v", got, cols)
	}
	for i := range cols {
		if got[i] != cols[i] {
			t.Fatalf("non-crosswalk table must return cols unchanged, got %v want %v", got, cols)
		}
	}
}

// --- paramIndex ---

func TestParamIndexFindsColumnPosition(t *testing.T) {
	cols := []string{"a", "b", "c"}
	if idx := paramIndex(cols, "b"); idx != 1 {
		t.Fatalf("paramIndex(cols, \"b\") = %d, want 1", idx)
	}
}

func TestParamIndexReturnsNegativeOneWhenAbsent(t *testing.T) {
	cols := []string{"a", "b", "c"}
	if idx := paramIndex(cols, "z"); idx != -1 {
		t.Fatalf("paramIndex(cols, \"z\") = %d, want -1", idx)
	}
}

// --- isIntegerValue / parseInt64 ---

func TestIsIntegerValue(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"positive integer string", "123", true},
		{"negative integer string", "-123", true},
		{"zero", "0", true},
		{"empty string", "", false},
		{"non-numeric string", "abc", false},
		{"float string is not integer", "1.5", false},
		{"non-string type", 123, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIntegerValue(tt.in); got != tt.want {
				t.Fatalf("isIntegerValue(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantVal int64
		wantOk  bool
	}{
		{"valid positive", "42", 42, true},
		{"valid negative", "-42", -42, true},
		{"whitespace padded", "  42  ", 42, true},
		{"empty string", "", 0, false},
		{"non-numeric", "abc", 0, false},
		{"float string is invalid", "1.5", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := parseInt64(tt.in)
			if gotVal != tt.wantVal || gotOk != tt.wantOk {
				t.Fatalf("parseInt64(%q) = (%d, %v), want (%d, %v)", tt.in, gotVal, gotOk, tt.wantVal, tt.wantOk)
			}
		})
	}
}

// --- isTextValue ---

func TestIsTextValue(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"non-empty string", "hello", true},
		{"whitespace-only string is blank", "   ", false},
		{"empty string", "", false},
		{"non-string type", 42, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTextValue(tt.in); got != tt.want {
				t.Fatalf("isTextValue(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// --- shouldSkipRow: per-table id/name guards beyond the clinics case already covered ---

func TestShouldSkipRowValidatesIDForMasterTables(t *testing.T) {
	for _, table := range []string{"animal_species", "exam_types", "owners", "procedures", "merchandise_items"} {
		t.Run(table+" non-integer id is skipped", func(t *testing.T) {
			entry := ManifestEntry{TargetTable: table}
			cols := []string{"id", "name"}
			args := []any{"not-a-number", "x"}
			if !shouldSkipRow(entry, cols, args) {
				t.Fatalf("%s: expected skip for non-integer id", table)
			}
		})
		t.Run(table+" integer id is kept", func(t *testing.T) {
			entry := ManifestEntry{TargetTable: table}
			cols := []string{"id", "name"}
			args := []any{"42", "x"}
			if shouldSkipRow(entry, cols, args) {
				t.Fatalf("%s: expected no skip for valid integer id", table)
			}
		})
	}
}

func TestShouldSkipRowValidatesPets(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number", "owner_id", "animal_species_id"}

	cases := []struct {
		name string
		args []any
		want bool
	}{
		{"blank pet_number skipped", []any{"", "300001", "1"}, true},
		{"non-integer owner_id skipped", []any{"P-1", "abc", "1"}, true},
		{"non-integer animal_species_id skipped", []any{"P-1", "300001", "abc"}, true},
		{"all valid kept", []any{"P-1", "300001", "1"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldSkipRow(entry, cols, tc.args); got != tc.want {
				t.Fatalf("%s: shouldSkipRow = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestShouldSkipRowValidatesExamTypeFields(t *testing.T) {
	entry := ManifestEntry{TargetTable: "exam_type_fields"}
	cols := []string{"exam_type_id", "name"}

	if !shouldSkipRow(entry, cols, []any{"abc", "x"}) {
		t.Fatal("exam_type_fields: non-integer exam_type_id must be skipped")
	}
	if shouldSkipRow(entry, cols, []any{"1", "x"}) {
		t.Fatal("exam_type_fields: valid exam_type_id must not be skipped")
	}
}

func TestShouldSkipRowUnlistedTableNeverSkips(t *testing.T) {
	entry := ManifestEntry{TargetTable: "vital_records"}
	cols := []string{"temperature"}
	if shouldSkipRow(entry, cols, []any{"38.5"}) {
		t.Fatal("unlisted table must never be skipped by shouldSkipRow (no per-table guard defined)")
	}
}

// --- syntheticColumns: clinic-scoped resolution + per-table special cases ---

func TestSyntheticColumnsSkipsClinicIDWhenAlreadyPresent(t *testing.T) {
	cols, exprs := syntheticColumns(ManifestEntry{TargetTable: "owners"}, []string{"id", "name", "clinic_id"})
	for _, c := range cols {
		if c == "clinic_id" {
			t.Fatalf("clinic_id already present in filteredCols; syntheticColumns must not add it again, got cols=%v exprs=%v", cols, exprs)
		}
	}
}

func TestSyntheticColumnsAnimalSpeciesNameFallbackWithID(t *testing.T) {
	cols, exprs := syntheticColumns(ManifestEntry{TargetTable: "animal_species"}, []string{"id"})
	idx := paramIndex(cols, "name")
	if idx < 0 {
		t.Fatalf("animal_species without name must synthesize a name column, got cols=%v", cols)
	}
	if !strings.Contains(exprs[idx], "legacy_species_") {
		t.Fatalf("animal_species synthetic name must reference legacy_species_, got %q", exprs[idx])
	}
}

func TestSyntheticColumnsAnimalSpeciesNameFallbackWithoutID(t *testing.T) {
	cols, exprs := syntheticColumns(ManifestEntry{TargetTable: "animal_species"}, []string{})
	idx := paramIndex(cols, "name")
	if idx < 0 {
		t.Fatalf("animal_species without id or name must still synthesize a name column, got cols=%v", cols)
	}
	if exprs[idx] != "'legacy_species'" {
		t.Fatalf("animal_species synthetic name without id must be the static literal, got %q", exprs[idx])
	}
}

func TestSyntheticColumnsClinicsCompanyIDDefault(t *testing.T) {
	cols, exprs := syntheticColumns(ManifestEntry{TargetTable: "clinics"}, []string{"id", "name"})
	idx := paramIndex(cols, "company_id")
	if idx < 0 {
		t.Fatalf("clinics without company_id must synthesize it, got cols=%v", cols)
	}
	if exprs[idx] != "1" {
		t.Fatalf("clinics synthetic company_id must default to base-seeded company 1, got %q", exprs[idx])
	}
}

func TestSyntheticColumnsUnlistedTableReturnsEmpty(t *testing.T) {
	// treatments is neither clinic-scoped (clinicScopedTables) nor special-cased
	// (animal_species/clinics), so syntheticColumns must add nothing.
	cols, exprs := syntheticColumns(ManifestEntry{TargetTable: "treatments"}, []string{"content"})
	if len(cols) != 0 || len(exprs) != 0 {
		t.Fatalf("non clinic-scoped, non special-cased table must synthesize nothing, got cols=%v exprs=%v", cols, exprs)
	}
}

// --- filterColumnsWithIndex ---

func TestFilterColumnsWithIndexNilAllowedKeepsEverything(t *testing.T) {
	cols := []string{"a", "b", "c"}
	kept, idx, dropped := filterColumnsWithIndex(cols, nil)
	if len(kept) != 3 || len(dropped) != 0 {
		t.Fatalf("nil allowed must keep all columns, got kept=%v dropped=%v", kept, dropped)
	}
	for i, want := range []int{0, 1, 2} {
		if idx[i] != want {
			t.Fatalf("index[%d] = %d, want %d", i, idx[i], want)
		}
	}
}

func TestFilterColumnsWithIndexDropsDisallowedAndTracksIndex(t *testing.T) {
	cols := []string{"id", "secret", "name"}
	allowed := map[string]bool{"id": true, "name": true}
	kept, idx, dropped := filterColumnsWithIndex(cols, allowed)

	if len(kept) != 2 || kept[0] != "id" || kept[1] != "name" {
		t.Fatalf("kept = %v, want [id name]", kept)
	}
	if len(dropped) != 1 || dropped[0] != "secret" {
		t.Fatalf("dropped = %v, want [secret]", dropped)
	}
	if len(idx) != 2 || idx[0] != 0 || idx[1] != 2 {
		t.Fatalf("idx = %v, want [0 2] (original positions of kept columns)", idx)
	}
}

// --- orderByDependency: same-tier stability + unknown-table fallback tier ---

func TestOrderByDependencyPreservesOrderWithinSameTier(t *testing.T) {
	entries := []ManifestEntry{
		{PlanKey: "exam_types::a", TargetTable: "exam_types"},
		{PlanKey: "owners::b", TargetTable: "owners"},
	}
	got := orderByDependency(entries)
	if got[0].PlanKey != "exam_types::a" || got[1].PlanKey != "owners::b" {
		t.Fatalf("entries in the same tier must preserve original order, got %v", got)
	}
}

func TestOrderByDependencyUnknownTableSortsLast(t *testing.T) {
	entries := []ManifestEntry{
		{PlanKey: "mystery::x", TargetTable: "mystery_table"},
		{PlanKey: "animal_species::y", TargetTable: "animal_species"},
	}
	got := orderByDependency(entries)
	if got[len(got)-1].PlanKey != "mystery::x" {
		t.Fatalf("unknown table (no tier entry) must sort last, got order %v", got)
	}
}

// --- loadManifest ---

func TestLoadManifestReadsValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	want := Manifest{
		FileCount: 2,
		TotalRows: 100,
		Entries: []ManifestEntry{
			{PlanKey: "owners::legacy", TargetTable: "owners", RowCount: 50},
		},
		SafetyPolicy: "local-only",
	}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("failed to marshal fixture manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write fixture manifest: %v", err)
	}

	got, err := loadManifest(path)
	if err != nil {
		t.Fatalf("loadManifest(%q) unexpected error: %v", path, err)
	}
	if got.FileCount != want.FileCount || got.TotalRows != want.TotalRows || got.SafetyPolicy != want.SafetyPolicy {
		t.Fatalf("loadManifest(%q) = %+v, want %+v", path, got, want)
	}
	if len(got.Entries) != 1 || got.Entries[0].TargetTable != "owners" {
		t.Fatalf("loadManifest(%q) entries = %+v, want 1 entry for owners", path, got.Entries)
	}
}

func TestLoadManifestMissingFileReturnsError(t *testing.T) {
	_, err := loadManifest(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("loadManifest for a missing file must return an error")
	}
}

func TestLoadManifestMalformedJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	_, err := loadManifest(path)
	if err == nil {
		t.Fatal("loadManifest for malformed JSON must return an error")
	}
}

// --- normalizeArgs: remaining branch coverage beyond main_test.go ---

func TestNormalizeArgsSkipsPetsOwnerIDNonInteger(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number", "owner_id", "animal_species_id"}
	args := []any{"P-1", "not-a-number", "1"}
	cache := &lookupCache{
		OwnerIDs:                map[int64]bool{},
		AnimalSpeciesIDs:        map[int64]bool{1: true},
		FallbackAnimalSpeciesID: 1,
	}
	if skip := normalizeArgs(entry, cols, args, cache); !skip {
		t.Fatal("normalizeArgs must skip pets whose owner_id does not parse as an integer")
	}
}

func TestNormalizeArgsKeepsAnimalSpeciesIDWhenPresentInCache(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number", "owner_id", "animal_species_id"}
	args := []any{"P-1", "300001", "7"}
	cache := &lookupCache{
		OwnerIDs:                map[int64]bool{300001: true},
		AnimalSpeciesIDs:        map[int64]bool{7: true},
		FallbackAnimalSpeciesID: 1,
	}
	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("normalizeArgs unexpectedly requested row skip")
	}
	if args[2] != "7" {
		t.Fatalf("animal_species_id already valid must be kept unchanged, got %v", args[2])
	}
}

func TestNormalizeArgsBillingsMedicalRecordIDUnresolvedBecomesNil(t *testing.T) {
	entry := ManifestEntry{TargetTable: "billings"}
	cols := []string{"medical_record_id", crosswalkPetNoCol}
	args := []any{"REC-MISSING", "P-1"}
	cache := &lookupCache{MedicalRecords: map[string]int64{}}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("billings: an unresolved medical_record_id must null the FK, not skip the row")
	}
	if args[0] != nil {
		t.Fatalf("medical_record_id = %v, want nil when the composite key is not in cache", args[0])
	}
}

func TestNormalizeArgsPetIDUnresolvedBecomesNil(t *testing.T) {
	entry := ManifestEntry{TargetTable: "billings"}
	cols := []string{"pet_id"}
	args := []any{"UNKNOWN-PET"}
	cache := &lookupCache{Pets: map[string]petLookup{}}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("billings: an unresolved pet_id must null the FK, not skip the row")
	}
	if args[0] != nil {
		t.Fatalf("pet_id = %v, want nil when the pet is not in cache", args[0])
	}
}

// --- skipReason: remaining table-specific gates ---

func TestSkipReasonMedicalRecordsWithoutRecordNoOrDate(t *testing.T) {
	entry := ManifestEntry{TargetTable: "medical_records", TargetColumns: []string{"pet_id", "next_visit_recommended_date"}}
	if got := skipReason(entry); got == "" {
		t.Fatal("medical_records without record_no and date must be skipped (MST_PET_INFO update-only source)")
	}
}

func TestSkipReasonExamTypesWithoutIDOrParentID(t *testing.T) {
	entry := ManifestEntry{TargetTable: "exam_types", TargetColumns: []string{"name"}}
	if got := skipReason(entry); got == "" {
		t.Fatal("exam_types without id and parent_id must be skipped (TBL_KNS_HIST name-only source)")
	}
}

func TestSkipReasonStaffsFromKarteDataIsSkipped(t *testing.T) {
	entry := ManifestEntry{TargetTable: "staffs", LegacyTable: "TBL_KARTE_DATA", TargetColumns: []string{"name"}}
	if got := skipReason(entry); got == "" {
		t.Fatal("staffs sourced from TBL_KARTE_DATA must be skipped (name-only, no FK value)")
	}
}

func TestSkipReasonSharedFilesAlwaysSkipped(t *testing.T) {
	entry := ManifestEntry{TargetTable: "shared_files"}
	if got := skipReason(entry); got == "" {
		t.Fatal("shared_files must always be skipped (required NOT NULL columns absent from TSV)")
	}
}

func TestSkipReasonValidEntryIsNotSkipped(t *testing.T) {
	entry := ManifestEntry{TargetTable: "owners", TargetColumns: []string{"id", "name"}}
	if got := skipReason(entry); got != "" {
		t.Fatalf("a fully valid owners entry must not be skipped, got reason: %q", got)
	}
}

func TestNormalizeArgsNilCacheNeverSkips(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number"}
	args := []any{"P-1"}
	if skip := normalizeArgs(entry, cols, args, nil); skip {
		t.Fatal("normalizeArgs with a nil cache must return false (no-op), not skip")
	}
}
