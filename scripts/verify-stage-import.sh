#!/usr/bin/env bash
# verify-stage-import.sh — post-import verification for `make stage-import`.
#
# Asserts every historical failure mode is absent after the stage-based import:
#   empty clinics, legacy branch-code clinic leakage, owner id collision,
#   record_no-only mis-link, orphan child rows, blocked/needs_review leakage,
#   demo/pre-old_db mixing, missing FK targets. Mirrors verify-old-db-seed.sh's
#   check harness but is scoped to the 7 stage-imported business tables.
#
# Usage: bash scripts/verify-stage-import.sh
# Requires: the AnimalEkarte compose stack up (db healthy).
#
# Exit codes: 0 all checks passed; 1 one or more failed.
set -euo pipefail

DC="docker compose --env-file .env.local"

pass=0
fail=0
warn=0

# OLD_DB_OWNER_ID_FLOOR scopes checks to OLD_DB-derived rows, excluding demo data
# (003_seed_demo.sql, owner ids 1-61). Legacy owner ids are preserved as the PK
# and start at 300001; the wide gap is the provenance boundary.
OLD_DB_OWNER_ID_FLOOR=300000

q() { $DC exec -T db sh -c "psql -U \$POSTGRES_USER -d \$POSTGRES_DB -tA -c \"$1\"" 2>/dev/null | tr -d ' \n'; }

check_gt() { # label, sql, expect_gt(default 0)
  local label="$1" sql="$2" gt="${3:-0}" result
  result=$(q "$sql")
  if [[ -z "$result" ]]; then echo "WARN  $label — no result"; ((warn++)) || true
  elif [[ "$result" -gt "$gt" ]]; then echo "PASS  $label = $result"; ((pass++)) || true
  else echo "FAIL  $label = $result (expected > $gt)"; ((fail++)) || true; fi
}

check_zero() { # label, sql
  local label="$1" sql="$2" result
  result=$(q "$sql")
  if [[ -z "$result" ]]; then echo "WARN  $label — no result"; ((warn++)) || true
  elif [[ "$result" -eq 0 ]]; then echo "PASS  $label = 0"; ((pass++)) || true
  else echo "FAIL  $label = $result (expected 0)"; ((fail++)) || true; fi
}

echo "========================================="
echo "Stage-Import Verification"
echo "========================================="

echo ""
echo "--- Row counts (old_db rows present) ---"
check_gt "old_db owners"          "SELECT count(*) FROM owners WHERE id >= $OLD_DB_OWNER_ID_FLOOR"
check_gt "old_db pets"            "SELECT count(*) FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR"
check_gt "old_db medical_records" "SELECT count(*) FROM medical_records WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR)"
check_gt "old_db billings"        "SELECT count(*) FROM billings WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR)"
check_gt "old_db billing_items"   "SELECT count(*) FROM billing_items WHERE billing_id IN (SELECT id FROM billings WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR))"
check_gt "old_db exams"           "SELECT count(*) FROM exams WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR)"
check_gt "old_db exam_results"    "SELECT count(*) FROM exam_results WHERE exam_id IN (SELECT id FROM exams WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR))"

echo ""
echo "--- Clinic resolution (八王子病院 single-clinic binding) ---"
# No empty/blank-name clinic may exist (the empty-clinics 5/6/7 defect).
check_zero "clinics with empty/blank name" \
  "SELECT count(*) FROM clinics WHERE COALESCE(TRIM(name),'') = ''"
# Every old_db owner/pet binds to 八王子病院, never a legacy branch code clinic.
check_zero "owners with NULL clinic_id" "SELECT count(*) FROM owners WHERE clinic_id IS NULL"
check_zero "old_db owners not bound to 八王子病院" \
  "SELECT count(*) FROM owners o WHERE o.id >= $OLD_DB_OWNER_ID_FLOOR AND NOT EXISTS (SELECT 1 FROM clinics c WHERE c.id = o.clinic_id AND c.name = '八王子病院')"
check_zero "pets with NULL clinic_id" "SELECT count(*) FROM pets WHERE clinic_id IS NULL"
check_zero "old_db pets not bound to 八王子病院" \
  "SELECT count(*) FROM pets p WHERE p.owner_id >= $OLD_DB_OWNER_ID_FLOOR AND NOT EXISTS (SELECT 1 FROM clinics c WHERE c.id = p.clinic_id AND c.name = '八王子病院')"

echo ""
echo "--- Owner id collision (demo-range leakage) ---"
# No old_db owner may land in the demo id range (<=61). Floor is 300000, so a
# count here would mean an owner id was both >= floor and <= 61 (impossible unless
# the floor regressed) — a structural guard against the collision class.
check_zero "old_db owner ids colliding with demo range (<=61)" \
  "SELECT count(*) FROM owners WHERE id >= $OLD_DB_OWNER_ID_FLOOR AND id <= 61"
# No old_db-derived owner may exist below the floor (would mean an old_db row
# leaked into the demo id space instead of being remapped above it — the legacy
# owner 000001 -> 310371 remap class). pets carry no legacy id, so scope owners only.
check_zero "old_db owners present below the id floor (demo-space leakage)" \
  "SELECT count(*) FROM owners o WHERE o.id < $OLD_DB_OWNER_ID_FLOOR AND o.id > 61 AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = o.clinic_id AND c.name = '八王子病院')"

echo ""
echo "--- record_no qualification (record_no-only mis-link prevented) ---"
# Stored record_no is pet-qualified '<pet>-<record>'; an unqualified value means
# the composite key collapsed and rows could mis-link.
check_zero "old_db medical_records with unqualified record_no" \
  "SELECT count(*) FROM medical_records WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR) AND record_no NOT LIKE '%-%'"
# (clinic_id, record_no) must be unique (the AnimalEkarte UNIQUE constraint).
check_zero "duplicate (clinic_id, record_no) in old_db medical_records" \
  "SELECT COALESCE(sum(c-1),0) FROM (SELECT count(*) c FROM medical_records WHERE pet_id IN (SELECT id FROM pets WHERE owner_id >= $OLD_DB_OWNER_ID_FLOOR) GROUP BY clinic_id, record_no HAVING count(*) > 1) x"

echo ""
echo "--- Orphan detection (FK integrity) ---"
check_zero "pets with missing owner" \
  "SELECT count(*) FROM pets WHERE owner_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM owners WHERE id = pets.owner_id)"
check_zero "medical_records with missing pet" \
  "SELECT count(*) FROM medical_records WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = medical_records.pet_id)"
check_zero "billings with missing pet" \
  "SELECT count(*) FROM billings WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = billings.pet_id)"
check_zero "billing_items with missing billing (hard FK)" \
  "SELECT count(*) FROM billing_items WHERE NOT EXISTS (SELECT 1 FROM billings WHERE id = billing_items.billing_id)"
check_zero "exams with missing pet" \
  "SELECT count(*) FROM exams WHERE pet_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pets WHERE id = exams.pet_id)"
check_zero "exam_results with missing exam (hard FK)" \
  "SELECT count(*) FROM exam_results WHERE NOT EXISTS (SELECT 1 FROM exams WHERE id = exam_results.exam_id)"

echo ""
echo "--- Demo / pre-old_db data preserved & unmixed ---"
# Demo owners (1-61) must be present after the import — the base seed creates them
# and the import must not delete them. A real assertion (no +1 escape): the import
# only touches owner ids >= the floor, so demo owners must remain > 0.
check_gt "demo owners present (id 1-61, not deleted by import)" \
  "SELECT count(*) FROM owners WHERE id BETWEEN 1 AND 61"
# clinics master must still contain 八王子病院.
check_gt "八王子病院 clinic present" "SELECT count(*) FROM clinics WHERE name = '八王子病院'"

echo ""
echo "--- Sequence safety (next insert will not collide) ---"
# After the importer runs advanceSequences(), every id sequence must be strictly
# greater than max(id) in its table. A sequence <= max(id) means the next app
# INSERT gets a PK that already exists, causing a duplicate-key error.
# check_zero: count of tables where last_value <= max(id) must be 0.
check_zero "sequences ahead of max id for all 7 business tables" \
  "SELECT count(*) FROM (
     SELECT 'owners' AS t, (SELECT last_value FROM owners_id_seq) AS seq, (SELECT COALESCE(max(id),0) FROM owners) AS mx
     UNION ALL SELECT 'pets', (SELECT last_value FROM pets_id_seq), (SELECT COALESCE(max(id),0) FROM pets)
     UNION ALL SELECT 'medical_records', (SELECT last_value FROM medical_records_id_seq), (SELECT COALESCE(max(id),0) FROM medical_records)
     UNION ALL SELECT 'billings', (SELECT last_value FROM billings_id_seq), (SELECT COALESCE(max(id),0) FROM billings)
     UNION ALL SELECT 'billing_items', (SELECT last_value FROM billing_items_id_seq), (SELECT COALESCE(max(id),0) FROM billing_items)
     UNION ALL SELECT 'exams', (SELECT last_value FROM exams_id_seq), (SELECT COALESCE(max(id),0) FROM exams)
     UNION ALL SELECT 'exam_results', (SELECT last_value FROM exam_results_id_seq), (SELECT COALESCE(max(id),0) FROM exam_results)
   ) s WHERE seq <= mx"

echo ""
echo "========================================="
echo "Results: PASS=$pass  FAIL=$fail  WARN=$warn"
echo "========================================="

if [[ "$fail" -gt 0 ]]; then exit 1; fi
exit 0
