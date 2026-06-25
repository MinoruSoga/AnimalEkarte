#!/usr/bin/env bash
# verify-old-db-seed.sh — row-count and relational checks after make seed-old-db
#
# Usage: bash scripts/verify-old-db-seed.sh
# Requires: Docker Compose stack is up (db service healthy)
#
# Exit codes:
#   0  all checks passed
#   1  one or more checks failed
set -euo pipefail

DC="docker compose --env-file .env.local"

pass=0
fail=0
warn=0

check() {
  local label="$1"
  local sql="$2"
  local expect_gt="${3:-0}"   # expected row count > this value
  local result
  result=$($DC exec -T db sh -c "psql -U \$POSTGRES_USER -d \$POSTGRES_DB -t -c \"$sql\"" 2>/dev/null | tr -d ' \n')
  if [[ -z "$result" ]]; then
    echo "WARN  $label — query returned no result"
    ((warn++)) || true
  elif [[ "$result" -gt "$expect_gt" ]]; then
    echo "PASS  $label = $result rows"
    ((pass++)) || true
  else
    echo "FAIL  $label = $result (expected > $expect_gt)"
    ((fail++)) || true
  fi
}

check_zero() {
  local label="$1"
  local sql="$2"
  local result
  result=$($DC exec -T db sh -c "psql -U \$POSTGRES_USER -d \$POSTGRES_DB -t -c \"$sql\"" 2>/dev/null | tr -d ' \n')
  if [[ -z "$result" ]]; then
    echo "WARN  $label — query returned no result"
    ((warn++)) || true
  elif [[ "$result" -eq 0 ]]; then
    echo "PASS  $label = 0 (no orphans)"
    ((pass++)) || true
  else
    echo "FAIL  $label = $result (expected 0)"
    ((fail++)) || true
  fi
}

echo "========================================="
echo "Old-DB Seed Verification"
echo "========================================="
echo ""
echo "--- Row count checks ---"

check "animal_species"     "SELECT COUNT(*) FROM animal_species"     0
check "clinics"            "SELECT COUNT(*) FROM clinics"            0
check "owners"             "SELECT COUNT(*) FROM owners"             0
check "pets"               "SELECT COUNT(*) FROM pets"               0
check "exam_types"         "SELECT COUNT(*) FROM exam_types"         0
check "exam_type_fields"   "SELECT COUNT(*) FROM exam_type_fields"   0
check "merchandise_items"  "SELECT COUNT(*) FROM merchandise_items"  0
check "procedures"         "SELECT COUNT(*) FROM procedures"         0
check "medical_records"    "SELECT COUNT(*) FROM medical_records"    0
check "billings"           "SELECT COUNT(*) FROM billings"           0
check "exams"              "SELECT COUNT(*) FROM exams"              0

echo ""
echo "--- Child-detail checks (composite crosswalk load) ---"

# Recoverable child-detail tables now load via the (pet_no, record_no/sno)
# composite crosswalk. A count of 0 here means the crosswalk regressed to BLOCKED.
check "vital_records"      "SELECT COUNT(*) FROM vital_records"      0
check "clinical_plans"     "SELECT COUNT(*) FROM clinical_plans"     0
check "inquiries"          "SELECT COUNT(*) FROM inquiries"          0
check "treatments"         "SELECT COUNT(*) FROM treatments"         0
check "billing_items"      "SELECT COUNT(*) FROM billing_items"      0
check "exam_results"       "SELECT COUNT(*) FROM exam_results"       0

echo "INFO  structurally skipped tables remain documented in the seed output"
echo "INFO  medical_record_addenda stays BLOCKED by design (author_user_id has no legacy source)"

echo ""
echo "--- Field preservation checks (silent-drop detection) ---"

# owners.dm_preference (migration 006, legacy MST_SIIK_INFO.SiiDM_Kbn).
# The source emits ~10,370 rows with DM codes 01/02, all of which map to a
# non-NULL boolean. If allowedColumns drops the column or the cast is wrong,
# every value lands as NULL and this check fails — catching the original bug
# that a pure row-count check would miss.
check "owners.dm_preference populated (not silently dropped)" \
  "SELECT COUNT(*) FROM owners WHERE dm_preference IS NOT NULL"

echo ""
echo "--- Clinic resolution checks (八王子病院 single-clinic binding) ---"

# clinics must NEVER contain an empty-name row. old_db has no real clinic master;
# the only legacy sources that ever targeted clinics were the pet/owner masters'
# hospital-CODE columns (QA-only crosswalk evidence) whose name is always NULL.
# Routing them into clinics inserted empty-name orphan clinic rows (ids 5/6/7,
# branch codes 05/06/07). clinics.name is NOT NULL with no default, so any blank
# name is invalid. A non-zero count means that defect regressed — the row-count
# check above (clinics >= 0) cannot catch it. Guarded at three layers: old_db
# generator (drops the entry), the seeder (shouldSkipRow), and here.
check_zero "clinics with empty/blank name" \
  "SELECT COUNT(*) FROM clinics WHERE COALESCE(TRIM(name), '') = ''"

# OLD_DB_OWNER_ID_FLOOR scopes the clinic-binding checks to OLD_DB-derived rows
# only, excluding the base-seed demo data (003_seed_demo.sql) that legitimately
# belongs to other clinics (城東センター病院 / 敷島医院). Legacy owner ids are the
# original MST_SIIK_INFO keys, preserved as the owners PK, and start at 300001;
# demo owner ids occupy 1–61. The wide gap makes this a safe, stable boundary.
# old_db pets carry no legacy numeric id, so they are scoped by their owner_id.
OLD_DB_OWNER_ID_FLOOR=300000

# owners.clinic_id must NEVER be NULL for ANY row (demo or old_db). The legacy
# export carries clinic_id as branch codes {01,05,06,07} that are not valid
# clinics(id); the seeder drops that column and supplies the single 八王子病院
# clinic synthetically. A non-zero count means resolution regressed and owners
# landed with an empty clinic_id — the exact bug this check guards.
check_zero "owners with NULL clinic_id" \
  "SELECT COUNT(*) FROM owners WHERE clinic_id IS NULL"

# Every OLD_DB owner must resolve to the canonical 八王子病院 clinic — not to a
# legacy branch code (5/6/7) or any other clinic. A non-zero count means an old_db
# owner is bound to a clinic other than 八王子病院, i.e. the legacy code leaked
# through. Demo owners (id < floor) are excluded by design.
check_zero "old_db owners not bound to the 八王子病院 clinic" \
  "SELECT COUNT(*) FROM owners o WHERE o.id >= $OLD_DB_OWNER_ID_FLOOR AND NOT EXISTS (SELECT 1 FROM clinics c WHERE c.id = o.clinic_id AND c.name = '八王子病院')"

# pets carry the same legacy branch code and must also bind to 八王子病院.
check_zero "pets with NULL clinic_id" \
  "SELECT COUNT(*) FROM pets WHERE clinic_id IS NULL"

# Scope to old_db pets via their owner: a pet whose owner is an old_db owner
# (owner_id >= floor) must bind to 八王子病院. Demo pets are excluded.
check_zero "old_db pets not bound to the 八王子病院 clinic" \
  "SELECT COUNT(*) FROM pets p WHERE p.owner_id >= $OLD_DB_OWNER_ID_FLOOR AND NOT EXISTS (SELECT 1 FROM clinics c WHERE c.id = p.clinic_id AND c.name = '八王子病院')"

echo ""
echo "--- Relational checks (orphan detection) ---"

# pets with valid owner reference
check_zero "pets with missing owner_id in owners" \
  "SELECT COUNT(*) FROM pets WHERE owner_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM owners WHERE id = pets.owner_id)"

# pets with valid clinic reference
check_zero "pets with missing clinic_id in clinics" \
  "SELECT COUNT(*) FROM pets WHERE clinic_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM clinics WHERE id = pets.clinic_id)"

# exam_type_fields FK to exam_types
check_zero "exam_type_fields with missing exam_type_id" \
  "SELECT COUNT(*) FROM exam_type_fields WHERE exam_type_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM exam_types WHERE id = exam_type_fields.exam_type_id)"

# medical_records, billings, and exams resolve old pet_number / record_no values through seed-time caches.
check_zero "medical_records with missing pet_id" \
  "SELECT COUNT(*) FROM medical_records WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = medical_records.pet_id)"

check_zero "medical_records with missing clinic_id" \
  "SELECT COUNT(*) FROM medical_records WHERE clinic_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM clinics WHERE id = medical_records.clinic_id)"

check_zero "billings with missing pet_id" \
  "SELECT COUNT(*) FROM billings WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = billings.pet_id)"

check_zero "billings with missing medical_record_id" \
  "SELECT COUNT(*) FROM billings WHERE medical_record_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM medical_records WHERE id = billings.medical_record_id)"

check_zero "exams with missing pet_id" \
  "SELECT COUNT(*) FROM exams WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = exams.pet_id)"

check_zero "exams with missing medical_record_id" \
  "SELECT COUNT(*) FROM exams WHERE medical_record_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM medical_records WHERE id = exams.medical_record_id)"

check_zero "exams with missing exam_type_id" \
  "SELECT COUNT(*) FROM exams WHERE exam_type_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM exam_types WHERE id = exams.exam_type_id)"

echo ""
echo "--- Child-detail orphan detection (composite crosswalk must never mis-link) ---"

# Every recoverable child row must resolve to an existing parent via the
# composite crosswalk; unresolved rows are skipped at seed time, never inserted
# as orphans. A non-zero count means a FK was written that does not resolve.
check_zero "vital_records with missing medical_record_id" \
  "SELECT COUNT(*) FROM vital_records WHERE medical_record_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM medical_records WHERE id = vital_records.medical_record_id)"

check_zero "clinical_plans with missing medical_record_id" \
  "SELECT COUNT(*) FROM clinical_plans WHERE NOT EXISTS (SELECT 1 FROM medical_records WHERE id = clinical_plans.medical_record_id)"

check_zero "inquiries with missing medical_record_id" \
  "SELECT COUNT(*) FROM inquiries WHERE NOT EXISTS (SELECT 1 FROM medical_records WHERE id = inquiries.medical_record_id)"

check_zero "treatments with missing medical_record_id" \
  "SELECT COUNT(*) FROM treatments WHERE NOT EXISTS (SELECT 1 FROM medical_records WHERE id = treatments.medical_record_id)"

check_zero "billing_items with missing billing_id" \
  "SELECT COUNT(*) FROM billing_items WHERE NOT EXISTS (SELECT 1 FROM billings WHERE id = billing_items.billing_id)"

check_zero "exam_results with missing exam_id" \
  "SELECT COUNT(*) FROM exam_results WHERE NOT EXISTS (SELECT 1 FROM exams WHERE id = exam_results.exam_id)"

echo ""
echo "========================================="
echo "Results: PASS=$pass  FAIL=$fail  WARN=$warn"
echo "========================================="

if [[ "$fail" -gt 0 ]]; then
  exit 1
fi
exit 0
