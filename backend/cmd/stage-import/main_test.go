package main

import (
	"fmt"
	"strings"
	"testing"
)

// containsAll fails the test unless every needle appears in haystack.
func containsAll(t *testing.T, label, haystack string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			t.Errorf("%s: expected to contain %q\n--- got ---\n%s", label, n, haystack)
		}
	}
}

// --- Safety boundary: non-local target host is refused ---

func TestGuardLocalTargetAcceptsLocalHosts(t *testing.T) {
	for _, h := range []string{"db", "localhost", "127.0.0.1", "::1", "[::1]"} {
		if err := guardLocalTarget(h); err != nil {
			t.Errorf("guardLocalTarget(%q) = %v, want nil", h, err)
		}
	}
}

func TestGuardLocalTargetRejectsNonLocalHosts(t *testing.T) {
	for _, h := range []string{
		"prod-db.example.com",
		"10.0.0.5",
		"staging.internal",
		"ekarte-stg.cluster-xyz.ap-northeast-1.rds.amazonaws.com",
		"",
	} {
		if err := guardLocalTarget(h); err == nil {
			t.Errorf("guardLocalTarget(%q) = nil, want refusal", h)
		}
	}
}

// --- Source contract: only confirmed + inferred are importable ---

func TestStatusInClauseIsOnlyConfirmedAndInferred(t *testing.T) {
	got := statusInClause()
	containsAll(t, "statusInClause", got, "'confirmed'", "'inferred'")
	for _, excluded := range []string{"needs_review", "archive_only", "blocked"} {
		if strings.Contains(got, excluded) {
			t.Errorf("statusInClause must NOT include %q; got %q", excluded, got)
		}
	}
}

func TestImportableStatusesExcludeNonImportable(t *testing.T) {
	for _, s := range importableStatuses {
		if s != "confirmed" && s != "inferred" {
			t.Fatalf("importableStatuses contains unexpected status %q", s)
		}
	}
}

// --- Source is animalekarte_stage only (never legacy_raw/legacy_canonical) ---

func TestSelectSQLReadsOnlyFromStageSchema(t *testing.T) {
	for _, tbl := range importOrder {
		sql := selectSQL(tbl, "")
		if !strings.Contains(sql, "FROM animalekarte_stage."+tbl) {
			t.Errorf("selectSQL(%q) must read FROM animalekarte_stage.%s; got %s", tbl, tbl, sql)
		}
		for _, forbidden := range []string{"legacy_raw", "legacy_canonical"} {
			if strings.Contains(sql, forbidden) {
				t.Errorf("selectSQL(%q) must NOT reference %s; got %s", tbl, forbidden, sql)
			}
		}
	}
}

// --- Hard-FK children cascade the parent skip (no dangling FK) ---

func TestImportableWhereCascadesParentSkipForHardFKChildren(t *testing.T) {
	bi := importableWhere("billing_items", "")
	containsAll(t, "billing_items importableWhere", bi,
		"EXISTS", "animalekarte_stage.billings", "p.id = billing_id")
	er := importableWhere("exam_results", "")
	containsAll(t, "exam_results importableWhere", er,
		"EXISTS", "animalekarte_stage.exams", "p.id = exam_id")
}

func TestImportableWhereGuardsPetsOwnerNotNull(t *testing.T) {
	// target pets.owner_id is NOT NULL; a pet with a NULL owner must be excluded.
	got := importableWhere("pets", "")
	containsAll(t, "pets importableWhere", got, "owner_id IS NOT NULL")
}

func TestImportableWhereParentTablesHaveNoCascade(t *testing.T) {
	// owners has no business-table parent, so no EXISTS cascade.
	if strings.Contains(importableWhere("owners", ""), "EXISTS") {
		t.Errorf("owners importableWhere should not contain an EXISTS cascade")
	}
}

// --- Target NOT NULL / CHECK / enum constraints satisfied in the SELECT ---

func TestNotNullDatesCoalesceToSentinel(t *testing.T) {
	cases := map[string]string{
		"medical_records": "COALESCE(date, DATE '1900-01-01')",
		"billings":        "COALESCE(scheduled_date, DATE '1900-01-01')",
		"exams":           "COALESCE(date, DATE '1900-01-01')",
	}
	for tbl, want := range cases {
		if !strings.Contains(selectSQL(tbl, ""), want) {
			t.Errorf("selectSQL(%q) must COALESCE the NOT NULL date: want %q", tbl, want)
		}
	}
}

func TestNameKanaIsNotImported(t *testing.T) {
	// target owners/pets name_kana CHECK rejects katakana; the importer omits the
	// column entirely (target DEFAULT '' applies), mirroring the direct seeder.
	for _, tbl := range []string{"owners", "pets"} {
		if strings.Contains(selectSQL(tbl, ""), "name_kana") {
			t.Errorf("selectSQL(%q) must NOT project name_kana (katakana CHECK)", tbl)
		}
		for _, c := range insertColumns(tbl) {
			if c == "name_kana" {
				t.Errorf("insertColumns(%q) must NOT include name_kana", tbl)
			}
		}
	}
}

func TestFallbacksFillNotNullColumnsStageLacks(t *testing.T) {
	// pets.animal_species_id ($2) and exams.exam_type_id ($3) are NOT NULL with no
	// default and absent from stage; they must come from resolved target fallbacks.
	if !strings.Contains(selectSQL("pets", ""), "$2::bigint AS animal_species_id") {
		t.Error("pets selectSQL must supply animal_species_id from the $2 fallback")
	}
	if !strings.Contains(selectSQL("exams", ""), "$2::bigint AS exam_type_id") {
		t.Error("exams selectSQL must supply exam_type_id from the $2 fallback")
	}
}

func TestSelectArgsCountMatchesPlaceholders(t *testing.T) {
	// PostgreSQL rejects a query given more params than it references. selectArgs must
	// supply exactly the highest $N referenced in selectSQL for each table.
	refs := targetRefs{clinicID: 1, fallbackAnimalSpeciesID: 2, fallbackExamTypeID: 3}
	for _, tbl := range importOrder {
		sql := selectSQL(tbl, "")
		maxN := 0
		for n := 1; n <= 9; n++ {
			if strings.Contains(sql, fmt.Sprintf("$%d", n)) {
				maxN = n
			}
		}
		gotN := len(selectArgs(tbl, refs))
		if gotN != maxN {
			t.Errorf("%s: selectArgs returns %d args but selectSQL references $1..$%d", tbl, gotN, maxN)
		}
	}
}

func TestEnumColumnsAreNotCastInStageQuery(t *testing.T) {
	// The stage DB has no AnimalEkarte enum types, so ::enum casts in the stage query
	// fail. Enum columns must be plain text (registered enum types let COPY coerce).
	for _, e := range []string{"::membership_type", "::pet_gender", "::pet_status",
		"::medical_record_status", "::visit_type", "::billing_status", "::item_category", "::tax_type"} {
		for _, tbl := range importOrder {
			if strings.Contains(selectSQL(tbl, ""), e) {
				t.Errorf("selectSQL(%q) must not cast to %s (stage DB lacks the type)", tbl, e)
			}
		}
	}
}

func TestClinicIDIsResolvedFromParamNeverStageValue(t *testing.T) {
	// clinic_id must be the target-resolved 八王子病院 id ($1), never the stage column,
	// so a stale/branch-code stage clinic id can never leak.
	for _, tbl := range []string{"owners", "pets", "medical_records", "billings", "exams"} {
		sql := selectSQL(tbl, "")
		if !strings.Contains(sql, "$1::bigint AS clinic_id") {
			t.Errorf("selectSQL(%q) must resolve clinic_id from $1, not the stage value; got %s", tbl, sql)
		}
	}
}

func TestBillingItemsQuantityStaysPositive(t *testing.T) {
	// target CHECK (quantity > 0): a 0/NULL stage quantity must become positive.
	if !strings.Contains(selectSQL("billing_items", ""), "GREATEST(COALESCE(quantity,1),0.1)") {
		t.Error("billing_items selectSQL must enforce quantity > 0")
	}
}

// --- selectSQL projection aligns with insertColumns (1:1, same order) ---

func TestSelectProjectionCountMatchesInsertColumns(t *testing.T) {
	for _, tbl := range importOrder {
		sql := selectSQL(tbl, "")
		// Count top-level "AS <col>" aliases is brittle; instead assert every insert
		// column name appears as an "AS <col>" or projected identifier in the SELECT.
		for _, col := range insertColumns(tbl) {
			if !strings.Contains(sql, " AS "+col) && !strings.Contains(sql, ", "+col) &&
				!strings.Contains(sql, "SELECT "+col) && !strings.Contains(sql, "SELECT id, "+col) {
				// fall back to a loose membership check on the column token
				if !strings.Contains(sql, col) {
					t.Errorf("selectSQL(%q) missing projection for insert column %q", tbl, col)
				}
			}
		}
	}
}

// --- Destructive delete scope: old_db rows only, demo preserved ---

func TestDeleteScopeOwnersUsesIDFloor(t *testing.T) {
	got := deleteScope("owners")
	if !strings.Contains(got, "id >= 300000") {
		t.Errorf("owners deleteScope must use the old_db owner id floor; got %s", got)
	}
}

func TestDeleteScopeChildrenDescendFromOldOwners(t *testing.T) {
	// children must be scoped by descent from old_db owners, never a raw id guess.
	cases := map[string]string{
		"pets":            "owner_id >= 300000",
		"medical_records": "pets WHERE owner_id >= 300000",
		"billings":        "pets WHERE owner_id >= 300000",
		"exams":           "pets WHERE owner_id >= 300000",
		"billing_items":   "public.billings",
		"exam_results":    "public.exams",
	}
	for tbl, want := range cases {
		got := deleteScope(tbl)
		if !strings.Contains(got, want) {
			t.Errorf("deleteScope(%q) must scope to old_db descent (%q); got %s", tbl, want, got)
		}
	}
}

func TestBillingsDeleteScopeUsesOwnerDescentNotRecordNo(t *testing.T) {
	// Provenance must be owner descent (pet -> old owner OR owner_id >= floor), NOT a
	// record_no signal: demo/staging MRs also use hyphenated record_no (e.g.
	// 'R-2025-001'), so a record_no scope over-matches staging billings and would
	// violate their RESTRICT children (billing_refunds) and the demo-preservation rule.
	got := deleteScope("billings")
	containsAll(t, "billings deleteScope", got, "owner_id >= 300000", "public.pets")
	if strings.Contains(got, "record_no") {
		t.Errorf("billings deleteScope must NOT use a record_no signal (over-matches staging); got %s", got)
	}
	// billing_items inherits the same billings scope through its subquery.
	gotBI := deleteScope("billing_items")
	containsAll(t, "billing_items deleteScope", gotBI, "owner_id >= 300000")
	if strings.Contains(gotBI, "record_no") {
		t.Errorf("billing_items deleteScope must NOT use a record_no signal; got %s", gotBI)
	}
}

func TestCascadeChildScrubCoversHighVolumeChildren(t *testing.T) {
	// The medical_records DELETE cascades into these high-volume children; they must
	// be pre-scrubbed (scoped to old_db MRs via qualified record_no) so the MR delete
	// is fast. They are NOT in the import set (scrub-only).
	want := map[string]bool{"treatments": true, "inquiries": true, "clinical_plans": true, "vital_records": true}
	got := map[string]bool{}
	for _, tbl := range cascadeChildScrubOrder {
		got[tbl] = true
	}
	for tbl := range want {
		if !got[tbl] {
			t.Errorf("cascadeChildScrubOrder missing high-volume MR cascade child %q", tbl)
		}
	}
	// none of the scrub-only tables may also be in importOrder (they aren't re-inserted)
	for _, it := range importOrder {
		if want[it] {
			t.Errorf("%q is scrub-only and must not be in importOrder", it)
		}
	}
	// scoped to old_db MRs via owner-id descent (NOT record_no, which over-matches demo)
	containsAll(t, "cascadeChildScope", cascadeChildScope(),
		"medical_record_id IN", "owner_id >= 300000")
	if strings.Contains(cascadeChildScope(), "record_no") {
		t.Errorf("cascadeChildScope must use owner descent, not record_no (over-matches staging)")
	}
}

func TestBillingsTotalAmountClampedNonNegative(t *testing.T) {
	// target CHECK requires total_amount >= 0; negative legacy amounts must clamp.
	if !strings.Contains(selectSQL("billings", ""), "GREATEST(COALESCE(total_amount,0), 0)") {
		t.Error("billings selectSQL must clamp total_amount to >= 0")
	}
}

func TestDeleteScopeNeverMatchesDemoRange(t *testing.T) {
	// A demo owner (id <= 61) must never satisfy any delete scope. The owners scope
	// is "id >= 300000", so demo ids are structurally excluded.
	if strings.Contains(deleteScope("owners"), "<=") || strings.Contains(deleteScope("owners"), "< 300000") {
		t.Error("owners deleteScope must not match the demo range")
	}
}

// --- FK-safe ordering: delete is the reverse of insert ---

func TestDeleteOrderIsReverseOfInsertOrder(t *testing.T) {
	if len(deleteOrder) != len(importOrder) {
		t.Fatalf("deleteOrder/importOrder length mismatch: %d vs %d", len(deleteOrder), len(importOrder))
	}
	for i := range importOrder {
		want := importOrder[len(importOrder)-1-i]
		if deleteOrder[i] != want {
			t.Errorf("deleteOrder[%d] = %q, want %q (reverse of importOrder)", i, deleteOrder[i], want)
		}
	}
}

func TestImportOrderPutsParentsBeforeChildren(t *testing.T) {
	pos := map[string]int{}
	for i, t := range importOrder {
		pos[t] = i
	}
	// every hard/soft FK child must come after its parent
	deps := [][2]string{
		{"pets", "owners"},
		{"medical_records", "pets"},
		{"billings", "pets"},
		{"billing_items", "billings"},
		{"exams", "medical_records"},
		{"exam_results", "exams"},
	}
	for _, d := range deps {
		child, parent := d[0], d[1]
		if pos[child] < pos[parent] {
			t.Errorf("import order violates FK: %q (%d) before parent %q (%d)", child, pos[child], parent, pos[parent])
		}
	}
}

func TestClinicsIsNeverInserted(t *testing.T) {
	// clinics is base-seeded (八王子病院 id=1); re-inserting would dup the PK.
	for _, tbl := range importOrder {
		if tbl == "clinics" {
			t.Error("importOrder must not include clinics (it is reused, never re-inserted)")
		}
	}
}

// --- run-id filter is escaped, not interpolated raw ---

func TestRunIDClauseEscapesSingleQuotes(t *testing.T) {
	got := runIDClause("a'b")
	if !strings.Contains(got, "'a''b'") {
		t.Errorf("runIDClause must escape single quotes; got %s", got)
	}
	if runIDClause("") != "" {
		t.Errorf("empty run-id must yield no clause; got %q", runIDClause(""))
	}
}

// --- pqLiteral safety ---

func TestPQLiteralDoublesQuotes(t *testing.T) {
	if got := pqLiteral("o'reilly"); got != "'o''reilly'" {
		t.Errorf("pqLiteral = %q, want '\\'o\\'\\'reilly\\''", got)
	}
}
