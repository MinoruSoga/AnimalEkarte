package main

import (
	"strings"
	"testing"
)

func TestOrderByDependencyMovesParentTablesFirst(t *testing.T) {
	entries := []ManifestEntry{
		{PlanKey: "billings::legacy", TargetTable: "billings"},
		{PlanKey: "pets::legacy", TargetTable: "pets"},
		{PlanKey: "clinics::legacy", TargetTable: "clinics"},
		{PlanKey: "animal_species::legacy", TargetTable: "animal_species"},
	}

	got := orderByDependency(entries)
	gotKeys := []string{}
	for _, entry := range got {
		gotKeys = append(gotKeys, entry.PlanKey)
	}

	want := []string{
		"clinics::legacy",
		"animal_species::legacy",
		"pets::legacy",
		"billings::legacy",
	}
	if len(gotKeys) != len(want) {
		t.Fatalf("got %d entries, want %d", len(gotKeys), len(want))
	}
	for i := range want {
		if gotKeys[i] != want[i] {
			t.Fatalf("order[%d] = %q, want %q; all=%v", i, gotKeys[i], want[i], gotKeys)
		}
	}
}

// TestNormalizeArgsResolvesLegacyPetNumberAndRecordNo: billings resolves pet_id
// from the old pet_number and medical_record_id from the COMPOSITE
// (pet_no, record_no), not record_no alone. The billings parent TSV carries the
// __xw_pet_no crosswalk column that supplies the pet side of the composite.
func TestNormalizeArgsResolvesLegacyPetNumberAndRecordNo(t *testing.T) {
	entry := ManifestEntry{TargetTable: "billings"}
	cols := []string{"pet_id", "medical_record_id", "total_amount", crosswalkPetNoCol}
	args := []any{"300001-001", "REC-1", "1200", "300001-001"}
	cache := &lookupCache{
		Pets: map[string]petLookup{
			"300001-001": {ID: 42, ClinicID: 1, OwnerID: 300001},
		},
		MedicalRecords: map[string]int64{
			qualifyRecordNo("300001-001", "REC-1"): 99,
		},
	}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("normalizeArgs unexpectedly requested row skip")
	}
	if args[0] != "42" {
		t.Fatalf("pet_id = %v, want 42", args[0])
	}
	if args[1] != "99" {
		t.Fatalf("medical_record_id = %v, want 99 (composite-resolved)", args[1])
	}
	if args[2] != "1200" {
		t.Fatalf("total_amount = %v, want unchanged 1200", args[2])
	}
}

// TestMedicalRecordCompositeKeyPreventsMisLink is the core regression for the
// record_no-only mis-link bug. Two different pets share record_no "00000001"
// (the value that repeats 15,101× in the legacy data). A record_no-only cache
// would resolve both to the same medical_record; the composite cache must keep
// them distinct.
func TestMedicalRecordCompositeKeyPreventsMisLink(t *testing.T) {
	cache := &lookupCache{
		Pets: map[string]petLookup{
			"300001-001": {ID: 1},
			"300002-001": {ID: 2},
		},
		MedicalRecords: map[string]int64{
			qualifyRecordNo("300001-001", "00000001"): 1001,
			qualifyRecordNo("300002-001", "00000001"): 2002,
		},
	}

	// treatments child for pet 300002-001, record_no 00000001 must resolve to
	// 2002, NOT 1001 (which a record_no-only cache would have returned first).
	entry := ManifestEntry{TargetTable: "treatments"}
	cols := []string{"content", crosswalkPetNoCol, crosswalkRecordNoCol, "medical_record_id"}
	args := []any{"some treatment", "300002-001", "00000001", nil}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("normalizeArgs unexpectedly skipped a row with a resolvable composite key")
	}
	if args[3] != "2002" {
		t.Fatalf("medical_record_id = %v, want 2002 (composite for pet 300002-001); a value of 1001 means the record_no-only mis-link regressed", args[3])
	}
}

// TestNormalizeArgsResolvesChildCrosswalkToMedicalRecordID covers the
// medical_records children that need only medical_record_id (vital_records also
// needs pet_id and is covered by TestVitalRecordsResolvesPetID). A resolvable
// composite fills the FK; a missing parent skips the row so no orphan is inserted.
func TestNormalizeArgsResolvesChildCrosswalkToMedicalRecordID(t *testing.T) {
	cache := &lookupCache{
		MedicalRecords: map[string]int64{
			qualifyRecordNo("P-1", "R-1"): 555,
		},
	}
	for _, table := range []string{"clinical_plans", "inquiries", "treatments"} {
		entry := ManifestEntry{TargetTable: table}
		cols := []string{"notes", crosswalkPetNoCol, crosswalkRecordNoCol, "medical_record_id"}

		// Resolvable.
		args := []any{"x", "P-1", "R-1", nil}
		if skip := normalizeArgs(entry, cols, args, cache); skip {
			t.Fatalf("%s: unexpectedly skipped a resolvable row", table)
		}
		if args[3] != "555" {
			t.Fatalf("%s: medical_record_id = %v, want 555", table, args[3])
		}

		// Parent absent → skip, never orphan-insert.
		argsMissing := []any{"x", "P-1", "R-NOPE", nil}
		if skip := normalizeArgs(entry, cols, argsMissing, cache); !skip {
			t.Fatalf("%s: must skip a row whose parent medical_record is absent (no orphan insert)", table)
		}
	}
}

// TestNormalizeArgsResolvesTwoHopBillingAndExam covers billing_items → billings
// and exam_results → exams via their dedicated composite caches.
func TestNormalizeArgsResolvesTwoHopBillingAndExam(t *testing.T) {
	cache := &lookupCache{
		Billings: map[string]int64{qualifyRecordNo("P-9", "S-9"): 700},
		Exams:    map[string]int64{qualifyRecordNo("P-9", "E-9"): 800},
	}

	bi := ManifestEntry{TargetTable: "billing_items"}
	biCols := []string{"name", crosswalkPetNoCol, crosswalkRecordNoCol, "billing_id"}
	biArgs := []any{"item", "P-9", "S-9", nil}
	if skip := normalizeArgs(bi, biCols, biArgs, cache); skip {
		t.Fatal("billing_items: unexpectedly skipped a resolvable row")
	}
	if biArgs[3] != "700" {
		t.Fatalf("billing_items: billing_id = %v, want 700", biArgs[3])
	}

	er := ManifestEntry{TargetTable: "exam_results"}
	erCols := []string{"name", crosswalkPetNoCol, crosswalkSnoCol, "exam_id"}
	erArgs := []any{"res", "P-9", "E-9", nil}
	if skip := normalizeArgs(er, erCols, erArgs, cache); skip {
		t.Fatal("exam_results: unexpectedly skipped a resolvable row")
	}
	if erArgs[3] != "800" {
		t.Fatalf("exam_results: exam_id = %v, want 800", erArgs[3])
	}

	// Missing parent billing → skip.
	missing := []any{"item", "P-9", "S-NONE", nil}
	if skip := normalizeArgs(bi, biCols, missing, cache); !skip {
		t.Fatal("billing_items: must skip when the parent billing is absent")
	}
}

// TestInsertColumnsExcludeCrosswalkMetadata guards that __xw_* columns never
// reach the INSERT, while the resolved FK column does.
func TestInsertColumnsExcludeCrosswalkMetadata(t *testing.T) {
	entry := ManifestEntry{TargetTable: "treatments"}
	readCols := withCrosswalkFKColumn(entry, []string{"content", crosswalkPetNoCol, crosswalkRecordNoCol})
	names, srcIdx := insertColumns(readCols)

	got := toSet(names)
	for _, xw := range []string{crosswalkPetNoCol, crosswalkRecordNoCol, crosswalkSnoCol} {
		if got[xw] {
			t.Fatalf("crosswalk column %q must not appear in INSERT columns: %v", xw, names)
		}
	}
	if !got["content"] {
		t.Fatalf("real column 'content' missing from INSERT columns: %v", names)
	}
	if !got["medical_record_id"] {
		t.Fatalf("resolved FK 'medical_record_id' must be an INSERT column: %v", names)
	}
	if len(names) != len(srcIdx) {
		t.Fatalf("names/srcIdx length mismatch: %d vs %d", len(names), len(srcIdx))
	}
	// srcIdx must point into readCols correctly.
	for k, src := range srcIdx {
		if readCols[src] != names[k] {
			t.Fatalf("srcIdx[%d]=%d points to %q but INSERT col is %q", k, src, readCols[src], names[k])
		}
	}
}

// TestSkipReasonUnblocksRecoverableChildrenWithCrosswalk verifies the gating:
// recoverable children load with crosswalk, are blocked without it, and the
// permanently-blocked tables stay blocked regardless.
func TestSkipReasonUnblocksRecoverableChildrenWithCrosswalk(t *testing.T) {
	recoverable := []struct {
		table string
		xw    []string
	}{
		{"vital_records", []string{crosswalkPetNoCol, crosswalkRecordNoCol}},
		{"clinical_plans", []string{crosswalkPetNoCol, crosswalkRecordNoCol}},
		{"inquiries", []string{crosswalkPetNoCol, crosswalkRecordNoCol}},
		{"treatments", []string{crosswalkPetNoCol, crosswalkRecordNoCol}},
		{"billing_items", []string{crosswalkPetNoCol, crosswalkRecordNoCol}},
		{"exam_results", []string{crosswalkPetNoCol, crosswalkSnoCol}},
	}
	for _, rc := range recoverable {
		withXW := ManifestEntry{TargetTable: rc.table, CrosswalkColumns: rc.xw}
		if r := skipReason(withXW); r != "" {
			t.Errorf("%s WITH crosswalk should load, got skip: %s", rc.table, r)
		}
		withoutXW := ManifestEntry{TargetTable: rc.table}
		if r := skipReason(withoutXW); r == "" {
			t.Errorf("%s WITHOUT crosswalk must be blocked", rc.table)
		}
	}

	// medical_record_addenda stays BLOCKED even with a crosswalk (author_user_id gap).
	addenda := ManifestEntry{TargetTable: "medical_record_addenda", CrosswalkColumns: []string{crosswalkPetNoCol, crosswalkRecordNoCol}}
	if r := skipReason(addenda); r == "" {
		t.Error("medical_record_addenda must stay BLOCKED (author_user_id has no source)")
	}

	// medical_records::TBL_TRI_DATA stays skipped by design.
	triMR := ManifestEntry{TargetTable: "medical_records", LegacyTable: "TBL_TRI_DATA", TargetColumns: []string{"record_no", "date"}}
	if r := skipReason(triMR); r == "" {
		t.Error("medical_records::TBL_TRI_DATA must stay skipped (TBL_KARTE_DATA is the source)")
	}
}

// TestChildDetailTablesLoadAfterParents asserts the line-item / child tables sort
// strictly after their parents so the composite caches are populated first.
func TestChildDetailTablesLoadAfterParents(t *testing.T) {
	entries := []ManifestEntry{
		{PlanKey: "exam_results::x", TargetTable: "exam_results"},
		{PlanKey: "billing_items::x", TargetTable: "billing_items"},
		{PlanKey: "treatments::x", TargetTable: "treatments"},
		{PlanKey: "exams::x", TargetTable: "exams"},
		{PlanKey: "billings::x", TargetTable: "billings"},
		{PlanKey: "medical_records::x", TargetTable: "medical_records"},
		{PlanKey: "pets::x", TargetTable: "pets"},
	}
	ordered := orderByDependency(entries)
	pos := map[string]int{}
	for i, e := range ordered {
		pos[e.TargetTable] = i
	}
	mustBefore := [][2]string{
		{"pets", "medical_records"},
		{"medical_records", "treatments"},
		{"billings", "billing_items"},
		{"exams", "exam_results"},
	}
	for _, mb := range mustBefore {
		if pos[mb[0]] >= pos[mb[1]] {
			t.Errorf("%s (pos %d) must load before %s (pos %d)", mb[0], pos[mb[0]], mb[1], pos[mb[1]])
		}
	}
}

// TestMedicalRecordsQualifiesRecordNo verifies medical_records stores record_no
// pet-qualified ("<pet_no>-<record_no>") so UNIQUE(clinic_id, record_no) does not
// collapse rows that merely share a legacy record_no across pets. Qualification
// must run on the old pet_no, before pet_id is converted to the new bigint.
func TestMedicalRecordsQualifiesRecordNo(t *testing.T) {
	entry := ManifestEntry{TargetTable: "medical_records", LegacyTable: "TBL_KARTE_DATA"}
	cols := []string{"pet_id", "record_no"}
	args := []any{"300002-001", "00000001"}
	cache := &lookupCache{Pets: map[string]petLookup{"300002-001": {ID: 7}}}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("medical_records row unexpectedly skipped")
	}
	if args[1] != "300002-001-00000001" {
		t.Fatalf("record_no = %v, want pet-qualified '300002-001-00000001'", args[1])
	}
	// pet_id must still resolve to the new bigint.
	if args[0] != "7" {
		t.Fatalf("pet_id = %v, want resolved bigint 7", args[0])
	}
	// The qualified record_no must equal the composite cache key children use.
	if qualifyRecordNo("300002-001", "00000001") != args[1] {
		t.Fatal("stored record_no must equal qualifyRecordNo() so children resolve via the same key")
	}
}

// TestVitalRecordsResolvesPetID guards the NOT NULL pet_id on vital_records: it is
// resolved from __xw_pet_no via the Pets cache, and a missing pet skips the row.
func TestVitalRecordsResolvesPetID(t *testing.T) {
	cache := &lookupCache{
		Pets:           map[string]petLookup{"P-1": {ID: 11}},
		MedicalRecords: map[string]int64{qualifyRecordNo("P-1", "R-1"): 555},
	}
	entry := ManifestEntry{TargetTable: "vital_records"}
	cols := withCrosswalkFKColumn(entry, []string{"temperature", crosswalkPetNoCol, crosswalkRecordNoCol})
	// cols now ends with medical_record_id, pet_id.
	args := make([]any, len(cols))
	args[0] = "38.5"
	args[1] = "P-1" // __xw_pet_no
	args[2] = "R-1" // __xw_record_no

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("vital_records row unexpectedly skipped")
	}
	mrIdx, petIdx := paramIndex(cols, "medical_record_id"), paramIndex(cols, "pet_id")
	if args[mrIdx] != "555" {
		t.Fatalf("medical_record_id = %v, want 555", args[mrIdx])
	}
	if args[petIdx] != "11" {
		t.Fatalf("pet_id = %v, want resolved 11", args[petIdx])
	}

	// Missing pet → skip (cannot insert NULL into NOT NULL pet_id).
	cache.Pets = map[string]petLookup{}
	args2 := make([]any, len(cols))
	args2[0] = "38.5"
	args2[1] = "P-1"
	args2[2] = "R-1"
	if skip := normalizeArgs(entry, cols, args2, cache); !skip {
		t.Fatal("vital_records must skip when pet_id cannot be resolved")
	}
}

// TestTreatmentsNullsProcedureIDWhenNotProcedure guards chk_treatment_item_ref:
// procedure_id may be non-NULL only when item_type is 'procedure' (legacy code
// '1'). A medicine/other row must have procedure_id NULLed.
func TestTreatmentsNullsProcedureIDWhenNotProcedure(t *testing.T) {
	cache := &lookupCache{MedicalRecords: map[string]int64{qualifyRecordNo("P", "R"): 1}}
	entry := ManifestEntry{TargetTable: "treatments"}
	cols := []string{"item_type", "procedure_id", crosswalkPetNoCol, crosswalkRecordNoCol, "medical_record_id"}

	// item_type '2' (medicine) → procedure_id must be NULLed.
	med := []any{"2", "990030", "P", "R", nil}
	if skip := normalizeArgs(entry, cols, med, cache); skip {
		t.Fatal("treatments medicine row unexpectedly skipped")
	}
	if med[1] != nil {
		t.Fatalf("procedure_id must be NULL for item_type=medicine, got %v", med[1])
	}

	// item_type '1' (procedure) → procedure_id retained.
	proc := []any{"1", "990030", "P", "R", nil}
	if skip := normalizeArgs(entry, cols, proc, cache); skip {
		t.Fatal("treatments procedure row unexpectedly skipped")
	}
	if proc[1] != "990030" {
		t.Fatalf("procedure_id must be kept for item_type=procedure, got %v", proc[1])
	}
}

// TestQuantityExprEnforcesPositive guards CHECK (quantity > 0): a legacy 0/NULL
// must coerce to a positive value, not 0.
func TestQuantityExprEnforcesPositive(t *testing.T) {
	expr := buildParamExpr(1, ManifestEntry{TargetTable: "billing_items"}, "quantity")
	if !strings.Contains(expr, "GREATEST") {
		t.Fatalf("quantity must be clamped positive (GREATEST), got: %q", expr)
	}
}

// TestProcedureAndMerchandiseFKUseExistsLookup guards the FK-existence cast so an
// unknown legacy item code becomes NULL instead of an FK-constraint error.
func TestProcedureAndMerchandiseFKUseExistsLookup(t *testing.T) {
	procExpr := buildParamExpr(1, ManifestEntry{TargetTable: "treatments"}, "procedure_id")
	if !strings.Contains(procExpr, "FROM public.procedures") {
		t.Fatalf("procedure_id must resolve via an existence lookup, got: %q", procExpr)
	}
	miExpr := buildParamExpr(1, ManifestEntry{TargetTable: "billing_items"}, "merchandise_item_id")
	if !strings.Contains(miExpr, "FROM public.merchandise_items") {
		t.Fatalf("merchandise_item_id must resolve via an existence lookup, got: %q", miExpr)
	}
	// Both must be NULL-safe (CASE … ELSE NULL) so a missing master row never errors.
	for _, e := range []string{procExpr, miExpr} {
		if !strings.Contains(e, "ELSE NULL") {
			t.Fatalf("FK existence lookup must fall back to NULL: %q", e)
		}
	}
}

// TestCrosswalkColumnContractMatchesGenerator pins the synthetic column names to
// the contract shared with old_db's generator. If either side renames a column,
// this fails before a silent crosswalk break ships.
func TestCrosswalkColumnContractMatchesGenerator(t *testing.T) {
	if crosswalkPetNoCol != "__xw_pet_no" {
		t.Errorf("crosswalkPetNoCol drifted from generator contract: %q", crosswalkPetNoCol)
	}
	if crosswalkRecordNoCol != "__xw_record_no" {
		t.Errorf("crosswalkRecordNoCol drifted: %q", crosswalkRecordNoCol)
	}
	if crosswalkSnoCol != "__xw_sno" {
		t.Errorf("crosswalkSnoCol drifted: %q", crosswalkSnoCol)
	}
}

func TestNormalizeArgsSkipsPetsWithMissingOwner(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number", "owner_id", "animal_species_id"}
	args := []any{"P-1", "300001", "1"}
	cache := &lookupCache{
		OwnerIDs:                map[int64]bool{},
		AnimalSpeciesIDs:        map[int64]bool{1: true},
		FallbackAnimalSpeciesID: 1,
	}

	if skip := normalizeArgs(entry, cols, args, cache); !skip {
		t.Fatal("normalizeArgs should skip pets whose owner_id is absent")
	}
}

func TestNormalizeArgsUsesFallbackAnimalSpecies(t *testing.T) {
	entry := ManifestEntry{TargetTable: "pets"}
	cols := []string{"pet_number", "owner_id", "animal_species_id"}
	args := []any{"P-1", "300001", "999"}
	cache := &lookupCache{
		OwnerIDs:                map[int64]bool{300001: true},
		AnimalSpeciesIDs:        map[int64]bool{1: true},
		FallbackAnimalSpeciesID: 1,
	}

	if skip := normalizeArgs(entry, cols, args, cache); skip {
		t.Fatal("normalizeArgs unexpectedly requested row skip")
	}
	if args[2] != "1" {
		t.Fatalf("animal_species_id = %v, want fallback 1", args[2])
	}
}

// TestOwnersAllowedColumnsIncludesDMPreference guards the original silent-drop
// bug: owners.dm_preference (migration 006) must survive the allowedColumns
// whitelist, otherwise filterColumnsWithIndex drops the TSV column silently.
func TestOwnersAllowedColumnsIncludesDMPreference(t *testing.T) {
	allowed := allowedColumns("owners")
	if allowed == nil {
		t.Fatal("allowedColumns(\"owners\") must not be nil (would accept-all and mask drift)")
	}
	if !allowed["dm_preference"] {
		t.Fatal("owners whitelist must include dm_preference; otherwise it is silently dropped")
	}
}

// TestOwnersDMPreferenceColumnSurvivesFilter exercises the real filter path with
// the production owners TSV header order to prove dm_preference is kept, not dropped.
func TestOwnersDMPreferenceColumnSurvivesFilter(t *testing.T) {
	header := []string{
		"id", "deleted_at", "name", "name_kana", "birth_date", "postal_code",
		"address1", "address2", "phone", "company_phone", "email",
		"home_postal_code", "home_address1", "home_address2", "dm_preference",
		"discount_rate", "membership_type", "created_at", "remarks", "clinic_id",
		"is_dangerous", "company",
	}
	kept, _, dropped := filterColumnsWithIndex(header, allowedColumns("owners"))

	keptSet := toSet(kept)
	if !keptSet["dm_preference"] {
		t.Fatalf("dm_preference was filtered out; kept=%v dropped=%v", kept, dropped)
	}
	for _, d := range dropped {
		if d == "dm_preference" {
			t.Fatalf("dm_preference appeared in dropped columns: %v", dropped)
		}
	}
	// name_kana is intentionally excluded (CHECK constraint) — confirm the filter
	// still drops it, so the test is actually filtering and not a no-op.
	if keptSet["name_kana"] {
		t.Fatal("name_kana should remain excluded by the owners whitelist")
	}
}

// TestDMPreferenceExprMapsLegacyCodes verifies the legacy DM code → boolean cast.
// table_definition_v4 codebook: 01=DM発送可→true, 02=DM発送不可→false, else NULL.
func TestDMPreferenceExprMapsLegacyCodes(t *testing.T) {
	entry := ManifestEntry{TargetTable: "owners"}
	expr := buildParamExpr(15, entry, "dm_preference")

	// Must reference the right parameter and be a boolean CASE — not the
	// default textExpr (which would fail to insert into a boolean column).
	if !strings.Contains(expr, "$15") {
		t.Fatalf("expr does not reference $15: %q", expr)
	}
	if !strings.Contains(expr, "CASE WHEN") {
		t.Fatalf("dm_preference must use a CASE expression, got: %q", expr)
	}
	if strings.Contains(expr, "::text, '')") {
		t.Fatalf("dm_preference must not fall through to textExpr (boolean column): %q", expr)
	}
	if !strings.Contains(expr, "::boolean") {
		t.Fatalf("dm_preference expr must cast to ::boolean: %q", expr)
	}
	// Directionality matters: a swapped true/false arm is the worst silent bug
	// (every owner's DM preference inverted). Assert the structural order
	// '01' … true … '02' … false so an arm swap fails the test.
	o1 := strings.Index(expr, "'01'")
	trueIdx := strings.Index(expr, "true")
	o2 := strings.Index(expr, "'02'")
	falseIdx := strings.Index(expr, "false")
	if o1 < 0 || trueIdx < 0 || o2 < 0 || falseIdx < 0 {
		t.Fatalf("expr must contain '01', true, '02', false: %q", expr)
	}
	if !(o1 < trueIdx && trueIdx < o2 && o2 < falseIdx) {
		t.Fatalf("expr must map '01'→true then '02'→false (got order o1=%d true=%d o2=%d false=%d): %q",
			o1, trueIdx, o2, falseIdx, expr)
	}
}

// TestBooleanColumnsHaveNonTextCast is a contract/regression test (harness P1).
//
// A column that maps to a non-text DDL type (e.g. boolean) MUST be (a) present in
// its table's allowedColumns whitelist so it is not silently dropped, and (b)
// produce a cast expression other than the default textExpr, so the generated
// INSERT does not try to write ” into a typed column. This is exactly the
// failure mode that hid owners.dm_preference: whitelisted-but-missing AND
// text-cast-instead-of-boolean. Adding a new typed legacy column without wiring
// both halves now fails here instead of silently dropping data.
func TestBooleanColumnsHaveNonTextCast(t *testing.T) {
	// (table, column) pairs that map to a boolean DDL column AND are actually
	// emitted by the old_db export (manifest targetColumns), so they flow through
	// the whitelist + cast path. Source of truth: new-schema-import-manifest.json
	// targetColumns ∩ boolean columns in migrations 001/006.
	// NOTE: pets exports danger_level (not is_dangerous), so pets.is_dangerous is
	// intentionally absent here — it is never present in the pets TSV.
	booleanColumns := []struct{ table, col string }{
		{"owners", "dm_preference"},                  // migration 006, MST_SIIK_INFO.SiiDM_Kbn
		{"owners", "is_dangerous"},                   // MST_SIIK_INFO
		{"billing_items", "is_insurance_applicable"}, // TBL_TRI_DATA (table currently skipped)
		{"treatments", "is_insurance"},               // TBL_TRI_DATA (table currently skipped)
	}

	for _, bc := range booleanColumns {
		allowed := allowedColumns(bc.table)
		if allowed != nil && !allowed[bc.col] {
			t.Errorf("%s.%s maps to a boolean column but is missing from allowedColumns(%q) — it would be silently dropped",
				bc.table, bc.col, bc.table)
		}

		expr := buildParamExpr(1, ManifestEntry{TargetTable: bc.table}, bc.col)
		if expr == textExpr(paramRef(1)) {
			t.Errorf("%s.%s falls through to textExpr but the DDL type is boolean — INSERT would write '' into a boolean column; add a cast in buildParamExpr",
				bc.table, bc.col)
		}
	}
}

func TestSkipReasonRequiresStructuralParentKeys(t *testing.T) {
	entry := ManifestEntry{
		PlanKey:       "billing_items::legacy",
		TargetTable:   "billing_items",
		LegacyTable:   "legacy",
		TargetColumns: []string{"name", "unit_price"},
	}

	if got := skipReason(entry); got == "" {
		t.Fatal("skipReason should reject billing_items without billing_id")
	}
}
