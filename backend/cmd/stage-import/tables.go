package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// importOrder is the FK-safe parent-before-child insert order for the 7 business
// tables. clinics is intentionally absent: the single 八王子病院 clinic is
// base-seeded (id=1) and reused, never re-inserted (would dup the PK).
var importOrder = []string{
	"owners",
	"pets",
	"medical_records",
	"billings",
	"billing_items",
	"exams",
	"exam_results",
}

// deleteOrder is the FK-safe child-before-parent delete order for the destructive
// pre-import scrub of old_db-derived target rows. It is the reverse of the insert
// order, with one nuance: billings.pet_id is ON DELETE SET NULL (NOT cascade from
// pets), so billings/billing_items must be deleted explicitly before pets.
// exams/exam_results CASCADE from medical_records, but we delete them explicitly
// too so the scrub is exact and self-documenting (and counts are reported).
// All RESTRICT child tables (lstep_*, prescriptions, hospitalizations, estimates,
// payment_splits, billing_refunds, medical_record_addenda, pet_chronic_conditions)
// currently hold 0 old_db rows; if that ever changes the delete will fail loudly
// inside the transaction (RESTRICT) rather than leave a half-scrubbed DB.
var deleteOrder = []string{
	"exam_results",
	"exams",
	"billing_items",
	"billings",
	"medical_records",
	"pets",
	"owners",
}

// cascadeChildScrubOrder lists high-volume CASCADE children of medical_records
// that the direct seeder populated but this importer does NOT re-insert (they are
// out of scope). Deleting old_db medical_records would CASCADE into them in a
// single statement — millions of rows (treatments ~1.5M; inquiries/clinical_plans/
// vital_records ~0.4M each) — which dominates the delete time. Pre-deleting them
// (scoped to old_db MRs) makes the subsequent medical_records DELETE find empty
// children and finish fast. These run FIRST, before deleteOrder. Each is scoped by
// descent from old_db owners (owner_id >= oldDBOwnerIDFloor) via cascadeChildScope(),
// NOT by a record_no signal (which over-matches demo MRs), so demo rows are never
// touched. Tables with 0 old_db rows are harmless no-ops.
var cascadeChildScrubOrder = []string{
	"treatments",
	"inquiries",
	"clinical_plans",
	"vital_records",
}

// cascadeChildScope scopes a medical_records CASCADE child to old_db rows: its
// medical_record_id points at an old_db medical_record, identified by OWNER DESCENT
// (the MR's pet -> an old_db owner). All four scrub tables carry a medical_record_id
// column. Owner descent is the clean provenance axis; a record_no signal would
// over-match demo/staging MRs (which also use hyphenated record_no like 'R-2025-001').
func cascadeChildScope() string {
	return fmt.Sprintf(
		`medical_record_id IN (SELECT id FROM public.medical_records WHERE pet_id IN `+
			`(SELECT id FROM public.pets WHERE owner_id >= %d))`,
		oldDBOwnerIDFloor)
}

// deleteScope maps each table to the predicate identifying its old_db-derived
// rows. Demo (owner id <= 61) / master / config rows are never matched. Children
// are scoped by descent from old_db owners, never by a raw id-range guess.
func deleteScope(table string) string {
	switch table {
	case "owners":
		return fmt.Sprintf("id >= %d", oldDBOwnerIDFloor)
	case "pets":
		return fmt.Sprintf("owner_id >= %d", oldDBOwnerIDFloor)
	case "medical_records":
		return petScopedExists("medical_records")
	case "billings":
		// A prior seed run SET NULL on billings.pet_id/owner_id when their parent was
		// deleted, leaving old_db billings with pet_id IS NULL whose ids still fall in
		// the stage id range. pet_id IN (...) is NULL-blind and would skip them, so the
		// re-insert would PK-collide. Catch the NULL-pet path via medical_record_id too.
		return billingsOldDBScope("billings")
	case "exams":
		return petScopedExists("exams")
	case "billing_items":
		return fmt.Sprintf(`billing_id IN (SELECT id FROM public.billings b WHERE %s)`, billingsOldDBScope("b"))
	case "exam_results":
		return fmt.Sprintf(`exam_id IN (SELECT id FROM public.exams e WHERE %s)`, petScopedExists("e"))
	default:
		panic("deleteScope: unknown table " + table)
	}
}

// petScopedExists returns a predicate true when the given alias's pet_id belongs
// to an old_db owner. Used for tables that carry pet_id directly.
func petScopedExists(alias string) string {
	return fmt.Sprintf(
		`%s.pet_id IN (SELECT id FROM public.pets WHERE owner_id >= %d)`,
		alias, oldDBOwnerIDFloor)
}

// billingsOldDBScope identifies old_db billings for the given alias, NULL-safely,
// by owner descent: pet_id -> an old_db pet, OR owner_id -> an old_db owner. Stage
// billings always carry a resolved owner_id/pet_id, so on a re-import these links
// are present and the scope is exhaustive. It deliberately does NOT use a
// medical_record record_no signal: demo/staging medical_records ALSO use a
// hyphenated record_no (e.g. 'R-2025-001'), so that signal over-matches staging
// rows and would try to delete them — violating their RESTRICT children
// (billing_refunds) and the demo-preservation contract. Owner-id descent is the
// only provenance axis that cleanly separates old_db (>=300000) from demo (<=61).
func billingsOldDBScope(alias string) string {
	return fmt.Sprintf(
		`(%s OR %s.owner_id >= %d)`,
		petScopedExists(alias), alias, oldDBOwnerIDFloor)
}

// importableWhere returns the WHERE predicate (without the leading WHERE) that
// selects importable rows for a table: the status+run-id filter, plus a
// parent-importable cascade for hard-FK children and an owner_id-present guard for
// pets (target pets.owner_id is NOT NULL). selectSQL and buildPlan both use this
// so the dry-run count and the apply insert agree exactly.
func importableWhere(table, runID string) string {
	w := statusInClause() + runIDClause(runID)
	switch table {
	case "pets":
		// target pets.owner_id is NOT NULL; a stage pet with NULL owner is
		// needs_review already, but guard explicitly so the count matches.
		return w + " AND owner_id IS NOT NULL"
	case "billing_items":
		return w + " AND " + parentImportable("billing_id", "billings", runID)
	case "exam_results":
		return w + " AND " + parentImportable("exam_id", "exams", runID)
	default:
		return w
	}
}

// selectSQL builds the typed SELECT against the STAGE schema for one table. The
// projected columns are exactly the target INSERT columns, in the same order as
// insertColumns(table). Stage-only columns (lineage) are excluded. Target NOT
// NULL / CHECK / enum constraints are satisfied here:
//   - clinic_id is the resolved TARGET 八王子病院 id (param), never the stage value;
//   - animal_species_id / exam_type_id are TARGET fallbacks (param) the stage lacks;
//   - NOT NULL dates (medical_records.date, billings.scheduled_date, exams.date)
//     COALESCE to 1900-01-01 when the legacy date was a sentinel/NULL;
//   - name_kana is NOT projected (target CHECK rejects katakana; matches seeder).
//
// Enum columns are selected as PLAIN TEXT (no ::enum cast): the AnimalEkarte enum
// types (membership_type, pet_gender, ...) exist only in the TARGET, not in the
// stage DB, so casting in the stage query fails ("type does not exist"). The stage
// values are already valid enum labels (guaranteed by migration-stage-check [10]),
// and COPY into the target enum column coerces the label (the target enum types are
// registered on the target pool so pgx can encode them).
//
// Each query references ONLY the params it needs (PostgreSQL rejects extra params):
// owners/medical_records/billings use $1=clinic_id; pets uses $1, $2=fallback
// animal_species_id; exams uses $1, $2=fallback exam_type_id; billing_items/
// exam_results take none. selectArgs() supplies exactly the matching args per table.
func selectSQL(table, runID string) string {
	where := importableWhere(table, runID)
	switch table {
	case "owners":
		// email dedup: target has a PARTIAL UNIQUE (clinic_id, email) WHERE email<>''.
		// All old_db owners land in one clinic, and legacy email is not unique (a few
		// share a placeholder address). Keep email only on the FIRST owner (lowest id)
		// per non-empty value; blank the rest so they fall outside the partial index.
		// Empty emails (the vast majority) are already excluded by the partial index.
		// No silent owner drop — only the redundant email string is cleared.
		return `SELECT id, $1::bigint AS clinic_id, COALESCE(name,'') AS name,
			birth_date, COALESCE(company,'') AS company, COALESCE(postal_code,'') AS postal_code,
			COALESCE(address1,'') AS address1, COALESCE(address2,'') AS address2,
			COALESCE(home_postal_code,'') AS home_postal_code, COALESCE(home_address1,'') AS home_address1,
			COALESCE(home_address2,'') AS home_address2, COALESCE(phone,'') AS phone,
			COALESCE(company_phone,'') AS company_phone,
			CASE WHEN COALESCE(email,'') <> '' AND row_number() OVER (
			       PARTITION BY email ORDER BY id) > 1
			     THEN '' ELSE COALESCE(email,'') END AS email,
			COALESCE(remarks,'') AS remarks, COALESCE(is_dangerous,false) AS is_dangerous,
			COALESCE(discount_rate,0) AS discount_rate,
			COALESCE(membership_type,'non_member') AS membership_type,
			dm_preference
			FROM animalekarte_stage.owners WHERE ` + where
	case "pets":
		return `SELECT id, $1::bigint AS clinic_id, owner_id, COALESCE(pet_number,'') AS pet_number,
			COALESCE(name,'') AS name, $2::bigint AS animal_species_id,
			COALESCE(gender,'unknown') AS gender,
			COALESCE(status,'alive') AS status,
			birth_date, COALESCE(breed,'') AS breed, COALESCE(color,'') AS color, weight,
			neutered_date, COALESCE(food,'') AS food, COALESCE(remarks,'') AS remarks, deceased_at
			FROM animalekarte_stage.pets WHERE ` + where
	case "medical_records":
		return `SELECT id, $1::bigint AS clinic_id, record_no,
			COALESCE(date, DATE '1900-01-01') AS date, owner_id, pet_id,
			COALESCE(status,'draft') AS status,
			visit_type
			FROM animalekarte_stage.medical_records WHERE ` + where
	case "billings":
		// total_amount: target CHECK requires >= 0. Some legacy rows carry negative
		// settlement amounts (refund/adjustment); clamp to 0 so the CHECK holds
		// (subtotal/tax_total are absent from stage and default to 0 in the target).
		return `SELECT id, $1::bigint AS clinic_id, medical_record_id, owner_id, pet_id,
			GREATEST(COALESCE(total_amount,0), 0) AS total_amount,
			COALESCE(status,'waiting') AS status,
			COALESCE(scheduled_date, DATE '1900-01-01') AS scheduled_date
			FROM animalekarte_stage.billings WHERE ` + where
	case "billing_items":
		// Hard FK billing_items.billing_id -> billings(id) NOT NULL. A line item
		// whose parent billing is NOT importable (needs_review/blocked/...) must be
		// skipped too, or the insert would dangle. parentImportable() enforces this.
		return `SELECT id, billing_id, COALESCE(category,'other') AS category,
			COALESCE(name,'') AS name, COALESCE(unit_price,0) AS unit_price,
			GREATEST(COALESCE(quantity,1),0.1) AS quantity,
			COALESCE(tax_type,'excluded') AS tax_type,
			COALESCE(is_insurance_applicable,false) AS is_insurance_applicable,
			COALESCE(sort_order,0) AS sort_order
			FROM animalekarte_stage.billing_items WHERE ` + where
	case "exams":
		// $1=clinic_id, $2=fallback_exam_type_id (exams takes two params; the second
		// is the exam_type fallback, NOT animal_species — see selectArgs).
		return `SELECT id, $1::bigint AS clinic_id, medical_record_id, pet_id,
			COALESCE(date, DATE '1900-01-01') AS date, $2::bigint AS exam_type_id,
			COALESCE(result_summary,'') AS result_summary
			FROM animalekarte_stage.exams WHERE ` + where
	case "exam_results":
		// Hard FK exam_results.exam_id -> exams(id) NOT NULL. Same parent-skip rule
		// as billing_items: skip a result whose exam is not importable.
		return `SELECT id, exam_id, COALESCE(name,'') AS name,
			COALESCE(inspection_value,'') AS inspection_value,
			COALESCE(normal_value,'') AS normal_value, COALESCE(sort_order,0) AS sort_order
			FROM animalekarte_stage.exam_results WHERE ` + where
	default:
		panic("selectSQL: unknown table " + table)
	}
}

// parentImportable returns an EXISTS clause asserting that the row's hard-FK
// parent (in parentTable) is itself importable (confirmed/inferred, same run-id).
// This makes the status skip CASCADE: a child of a skipped parent is skipped too,
// so a hard NOT NULL FK can never dangle. fkCol is the child's FK column name.
func parentImportable(fkCol, parentTable, runID string) string {
	return fmt.Sprintf(
		`EXISTS (SELECT 1 FROM animalekarte_stage.%s p WHERE p.id = %s AND p.%s%s)`,
		parentTable, fkCol, statusInClause(), prefixRunID("p", runID))
}

// prefixRunID is runIDClause with a table alias on migration_run_id, for use in a
// correlated subquery where the bare column would be ambiguous.
func prefixRunID(alias, runID string) string {
	if runID == "" {
		return ""
	}
	return fmt.Sprintf(" AND %s.migration_run_id = %s", alias, pqLiteral(runID))
}

// insertColumns returns the target column list (order matches selectSQL projection).
func insertColumns(table string) []string {
	switch table {
	case "owners":
		return []string{"id", "clinic_id", "name", "birth_date", "company", "postal_code",
			"address1", "address2", "home_postal_code", "home_address1", "home_address2",
			"phone", "company_phone", "email", "remarks", "is_dangerous", "discount_rate",
			"membership_type", "dm_preference"}
	case "pets":
		return []string{"id", "clinic_id", "owner_id", "pet_number", "name", "animal_species_id",
			"gender", "status", "birth_date", "breed", "color", "weight", "neutered_date",
			"food", "remarks", "deceased_at"}
	case "medical_records":
		return []string{"id", "clinic_id", "record_no", "date", "owner_id", "pet_id",
			"status", "visit_type"}
	case "billings":
		return []string{"id", "clinic_id", "medical_record_id", "owner_id", "pet_id",
			"total_amount", "status", "scheduled_date"}
	case "billing_items":
		return []string{"id", "billing_id", "category", "name", "unit_price", "quantity",
			"tax_type", "is_insurance_applicable", "sort_order"}
	case "exams":
		return []string{"id", "clinic_id", "medical_record_id", "pet_id", "date",
			"exam_type_id", "result_summary"}
	case "exam_results":
		return []string{"id", "exam_id", "name", "inspection_value", "normal_value", "sort_order"}
	default:
		panic("insertColumns: unknown table " + table)
	}
}

// selectArgs returns exactly the bind args the table's selectSQL references, in
// order. PostgreSQL rejects a query supplied with more params than it references,
// so this must stay in lockstep with the $N placeholders in selectSQL.
func selectArgs(table string, refs targetRefs) []any {
	switch table {
	case "owners", "medical_records", "billings":
		return []any{refs.clinicID} // $1
	case "pets":
		return []any{refs.clinicID, refs.fallbackAnimalSpeciesID} // $1,$2
	case "exams":
		return []any{refs.clinicID, refs.fallbackExamTypeID} // $1,$2
	case "billing_items", "exam_results":
		return nil // no params
	default:
		panic("selectArgs: unknown table " + table)
	}
}

// copyRowsFromStage streams a stage SELECT into the target table via COPY. Because
// stage and target are SEPARATE Postgres servers, the rows flow stage -> Go ->
// target; pgx CopyFrom on the target side is the fast bulk path. Runs inside the
// caller's target transaction (tx).
func copyRowsFromStage(
	ctx context.Context,
	tx pgx.Tx,
	stagePool *pgxpool.Pool,
	table, runID string,
	refs targetRefs,
) (int64, error) {
	rows, err := stagePool.Query(ctx, selectSQL(table, runID), selectArgs(table, refs)...)
	if err != nil {
		return 0, fmt.Errorf("stage query %s: %w", table, err)
	}
	defer rows.Close()

	cols := insertColumns(table)
	// pgx.CopyFromFunc pulls each row from the stage cursor and hands its values to
	// the target COPY stream. rows.Values() returns []any already typed by pgx, so
	// no string round-trip and no manual casting.
	var scanErr error
	src := pgx.CopyFromFunc(func() ([]any, error) {
		if !rows.Next() {
			return nil, rows.Err()
		}
		vals, err := rows.Values()
		if err != nil {
			scanErr = err
			return nil, err
		}
		return vals, nil
	})

	n, err := tx.CopyFrom(ctx, pgx.Identifier{"public", table}, cols, src)
	if err != nil {
		return 0, fmt.Errorf("copy into %s: %w", table, err)
	}
	if scanErr != nil {
		return 0, fmt.Errorf("scan stage %s: %w", table, scanErr)
	}
	return n, nil
}
