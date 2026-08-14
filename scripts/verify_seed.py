#!/usr/bin/env python3
"""CSV seed-bundle verification for master/demo/staging seed data.

Prior to the CSV migration, this script statically parsed and replayed
INSERT/ON CONFLICT statements from backend/migrations/002-004_*.sql to
simulate the final row state. That replay engine is gone: the seed data now
lives in backend/migrations/seeds/<bundle>/*.csv + manifest.json, sourced by
dumping a disposable database after a real 001->004 apply (backend/cmd/
seed-export) — not by parsing SQL. Each CSV row already IS the final merged
state (COPY dumped the actual table after every INSERT/ON CONFLICT/DO block
had already executed), so there is nothing left to replay or merge here:
this script just reads the CSVs and checks business invariants against them.

Does not touch a database — reads only the committed CSV files.
"""

from __future__ import annotations

from collections import Counter, defaultdict
from datetime import datetime, timedelta, timezone
from pathlib import Path
import csv
import json
import re
import sys


ROOT = Path(__file__).resolve().parents[1]
SEEDS_ROOT = ROOT / "backend/migrations/seeds"
BUNDLE_DIRS = ("002_master", "003_demo", "004_staging")

# Remarks that denote a combination vaccine ("N種混合" / "混合ワクチン").
COMBO_VACCINE_REMARK_RE = re.compile(r"\d+種混合|混合ワクチン")

# exam_type_field semantic category -> substrings the owning exam_type name must
# contain. Blood markers are matched first because "BUN（尿素窒素）" contains "尿"
# yet is a blood biochemistry field, not a urine field.
EXAM_FIELD_BLOOD_TOKENS = (
    "WBC", "RBC", "HGB", "HCT", "PLT", "MCV", "MCH",
    "ALT", "AST", "GPT", "GOT", "ALP", "GGT", "BUN", "CRE", "GLU", "TBIL",
)
EXAM_FIELD_BLOOD_TOKENS_JP = (
    "白血球", "赤血球", "ヘマトクリット", "血小板", "クレアチニン", "尿素窒素", "血糖", "血球",
)
EXAM_FIELD_ECHO_TOKENS = ("エコー", "超音波")
EXAM_FIELD_XRAY_TOKENS = ("正面", "四肢", "レントゲン", "X線", "Xray")
EXAM_TYPE_EXPECTED_BY_CATEGORY = {
    "blood": ("血液",),
    "echo": ("エコー", "超音波", "Echo"),
    "xray": ("レントゲン", "Xray", "X線", "X-ray"),
    "urine": ("尿",),
}

JST = timezone(timedelta(hours=9))

SEVEN_MASTER_TABLES = (
    "exam_types",
    "vaccines",
    "medicines",
    "consultations",
    "procedures",
    "trimming_courses",
    "merchandise_items",
)

TRACKED_TABLES = {
    *SEVEN_MASTER_TABLES,
    "inventory_items",
    "medical_records",
    "treatments",
}

# Tables read directly by name-specific checks below, beyond TRACKED_TABLES.
EXTRA_TABLES = {
    "appointments",
    "audit_logs",
    "staffs",
    "vaccinations",
    "exam_type_fields",
}

EXPECTED_TREATMENTS = {
    5: {"procedure_id": 13019},
    101: {
        "item_type": "other",
        "consultation_id": None,
        "procedure_id": None,
        "medicine_id": None,
    },
    103: {"procedure_id": 13019},
    105: {"procedure_id": 13016},
    106: {"consultation_id": 4},
    107: {"procedure_id": 13029},
    301: {"consultation_id": 7},
    302: {"consultation_id": 6},
}

EXPECTED_MISSING_PROCEDURES = {13018, 13030}
EXPECTED_PRESENT_PROCEDURES = {13049, 13051}
IMPORTED_OWNER_ID_FLOOR = 300000
# Imported STG/demo dumps use high serial ids for appointments (see appointments.csv min id).
IMPORTED_APPOINTMENT_ID_FLOOR = 1_000_000
# When the clinical graph is dump-imported, treatments.csv is intentionally empty
# (legacy fixture rows removed) while medical_records is huge.
IMPORTED_MEDICAL_RECORD_MIN_COUNT = 1_000


# ------------------------------------------------------------------
# CSV loading
# ------------------------------------------------------------------

def build_table_index() -> dict[str, Path]:
    """Map table name -> CSV path by reading every bundle's manifest.json."""
    index: dict[str, Path] = {}
    for bundle in BUNDLE_DIRS:
        manifest_path = SEEDS_ROOT / bundle / "manifest.json"
        if not manifest_path.exists():
            raise SystemExit(
                f"missing {manifest_path} — run backend/cmd/seed-export first "
                "(see backend/migrations/CLAUDE.md)"
            )
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        for entry in manifest["tables"]:
            index[entry["table"]] = SEEDS_ROOT / bundle / entry["csvFile"]
    return index


def to_int(value: str | None) -> int | None:
    if value is None or value == "":
        return None
    return int(value)


def to_bool(value: str) -> bool | None:
    if value == "":
        return None
    if value == "t":
        return True
    if value == "f":
        return False
    raise ValueError(f"unexpected boolean CSV value: {value!r}")


def read_rows(table_index: dict[str, Path], table: str) -> list[dict[str, str]]:
    path = table_index.get(table)
    if path is None:
        raise SystemExit(f"table {table!r} not found in any bundle manifest — was it renamed or removed?")
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f))


class SeedState:
    """Per-table id -> row dict, for tables whose business invariants are
    checked below. Unlike the pre-CSV version, there is no ON CONFLICT merge
    logic here: a CSV row already is the table's final state, and a
    table's primary key is unique by construction, so "loading" is just a
    direct read.
    """

    def __init__(self, table_index: dict[str, Path]) -> None:
        self.table_index = table_index
        self.tables: dict[str, dict[int, dict[str, object]]] = {}

    def load(self, table: str, int_columns: tuple[str, ...], bool_columns: tuple[str, ...] = ()) -> None:
        rows: dict[int, dict[str, object]] = {}
        for raw in read_rows(self.table_index, table):
            row: dict[str, object] = dict(raw)
            for col in int_columns:
                if col in row:
                    row[col] = to_int(row[col])
            for col in bool_columns:
                if col in row:
                    row[col] = to_bool(row[col])
            row_id = row.get("id")
            rows[row_id] = row
        self.tables[table] = rows


def load_seed_state() -> tuple[SeedState, dict[str, Path]]:
    table_index = build_table_index()
    state = SeedState(table_index)

    # Column sets are per-table because CSV headers differ; only columns the
    # checks below actually touch need int/bool coercion.
    state.load("exam_types", ("id", "clinic_id", "parent_id"), ("is_active",))
    state.load("vaccines", ("id", "clinic_id", "parent_id", "inventory_id"), ("is_active",))
    state.load("medicines", ("id", "clinic_id", "parent_id", "inventory_id"), ("is_active",))
    state.load("consultations", ("id", "clinic_id", "parent_id"), ("is_active",))
    state.load("procedures", ("id", "clinic_id", "parent_id"), ("is_active",))
    state.load("trimming_courses", ("id", "clinic_id", "course_type_id"), ())
    # Trigger-owned child of clinics; explicit CSV owns stable IDs for course_type_id FKs.
    state.load("trimming_course_types", ("id", "clinic_id"), ())
    # Trigger-owned child of clinics; explicit CSV owns stable IDs for payment_method_id FKs.
    state.load("payment_methods", ("id", "clinic_id"), ("is_active",))
    state.load("payments", ("id", "payment_method_id"), ())
    state.load("payment_splits", ("id", "clinic_id", "payment_method_id"), ())
    state.load("merchandise_items", ("id", "clinic_id"), ("is_active",))
    state.load("inventory_items", ("id", "clinic_id"), ())
    state.load("owners", ("id", "clinic_id"), ())
    # estimates.owner_id previously hard-coded 1–5 against imported owners ≥ 300000 (seed FK fail).
    state.load("estimates", ("id", "clinic_id", "owner_id"), ())
    state.load("medical_records", ("id", "clinic_id"), ())
    state.load(
        "treatments",
        ("id", "medical_record_id", "consultation_id", "procedure_id", "medicine_id", "inventory_id"),
        ("is_selected", "is_insurance"),
    )
    return state, table_index


def is_unique_scope_row(table: str, row: dict[str, object]) -> bool:
    if row.get("deleted_at"):
        return False
    if table == "merchandise_items":
        return row.get("is_active") is True
    return True


def add_result(errors: list[str], condition: bool, message: str) -> None:
    if not condition:
        errors.append(message)


# ------------------------------------------------------------------
# Checks (business invariants) — logic unchanged from the SQL-replay
# version; only the row source changed (CSV dict instead of parsed INSERT).
# ------------------------------------------------------------------

def check_csv_row_integrity(state: SeedState, errors: list[str]) -> None:
    """Sanity-check the CSVs themselves: no duplicate/missing id per table.

    Replaces the old check_source_id_duplicates, which detected duplicate ids
    across multiple INSERT statements targeting the same PK during SQL
    replay. That scenario cannot occur from a CSV dump of a real table (a
    table's primary key is unique by construction) — this instead guards
    against a corrupted or hand-edited CSV file.
    """
    for table in SEVEN_MASTER_TABLES:
        rows = read_rows(state.table_index, table)
        ids = [to_int(r["id"]) for r in rows if r.get("id", "") != ""]
        duplicates = sorted(id_ for id_, count in Counter(ids).items() if count > 1)
        add_result(errors, not duplicates, f"{table}: duplicate id values in CSV: {duplicates}")


def check_unique_name_duplicates(state: SeedState, errors: list[str]) -> None:
    for table in SEVEN_MASTER_TABLES:
        seen: Counter[tuple[object, object]] = Counter()
        for row in state.tables[table].values():
            if is_unique_scope_row(table, row):
                seen[(row.get("clinic_id"), row.get("name"))] += 1
        duplicates = sorted(key for key, count in seen.items() if count > 1)
        add_result(errors, not duplicates, f"{table}: effective (clinic_id, name) duplicates found: {duplicates}")


def _int_keys(table: dict) -> list[int]:
    """table id keys that are int (CSV may inject None keys)."""
    return [k for k in table if isinstance(k, int)]


def is_imported_clinical_graph(state: SeedState) -> bool:
    """True when 003_demo is an imported clinical dump, not the tiny legacy demo fixtures.

    Signals (any sufficient):
    - all owner ids are above the import floor, or
    - treatments is empty while medical_records is large (legacy treatment fixtures removed).
    Mixed owner id ranges still count as imported when treatments are gone and the
    clinical graph is dump-scale — requiring EXPECTED_TREATMENTS then false-fails TASK-009.
    """
    owners = state.tables.get("owners", {})
    owner_ids = _int_keys(owners)
    if owner_ids and min(owner_ids) >= IMPORTED_OWNER_ID_FLOOR:
        return True
    treatments = state.tables.get("treatments", {})
    medical_records = state.tables.get("medical_records", {})
    if not treatments and medical_records and len(medical_records) >= IMPORTED_MEDICAL_RECORD_MIN_COUNT:
        return True
    return False


def check_expected_treatments(state: SeedState, errors: list[str]) -> None:
    # The legacy demo treatment fixtures are intentionally removed when the
    # sensitive-local owner/clinical graph replaces 003_demo. Keep these
    # historical assertions for the original demo bundle, but do not require
    # rows whose medical_record_ids no longer exist in the imported dataset.
    if is_imported_clinical_graph(state):
        return
    treatments = state.tables["treatments"]
    for treatment_id, expected in EXPECTED_TREATMENTS.items():
        row = treatments.get(treatment_id)
        add_result(errors, row is not None, f"treatments#{treatment_id}: row not found")
        if row is None:
            continue
        for column, expected_value in expected.items():
            add_result(
                errors,
                row.get(column) == expected_value,
                f"treatments#{treatment_id}: expected {column}={expected_value!r}, got {row.get(column)!r}",
            )


def check_treatment_constraints(state: SeedState, errors: list[str]) -> None:
    for treatment_id, row in sorted(state.tables["treatments"].items()):
        item_type = row.get("item_type")
        consultation_id = row.get("consultation_id")
        procedure_id = row.get("procedure_id")
        medicine_id = row.get("medicine_id")
        valid = (
            (item_type == "consultation" and procedure_id is None and medicine_id is None)
            or (item_type == "procedure" and consultation_id is None and medicine_id is None)
            or (item_type == "medicine" and consultation_id is None and procedure_id is None)
            or (
                item_type == "other"
                and consultation_id is None
                and procedure_id is None
                and medicine_id is None
            )
        )
        add_result(errors, valid, f"treatments#{treatment_id}: chk_treatment_item_ref violation")


def check_procedure_presence(state: SeedState, errors: list[str]) -> None:
    procedures = state.tables["procedures"]
    for procedure_id in EXPECTED_MISSING_PROCEDURES:
        add_result(errors, procedure_id not in procedures, f"procedures#{procedure_id}: expected absent")
    for procedure_id in EXPECTED_PRESENT_PROCEDURES:
        add_result(errors, procedure_id in procedures, f"procedures#{procedure_id}: expected present")


def check_fk_integrity(state: SeedState, errors: list[str]) -> None:
    fk_rules = (
        ("exam_types", "parent_id", "exam_types"),
        ("vaccines", "parent_id", "vaccines"),
        ("medicines", "parent_id", "medicines"),
        ("medicines", "inventory_id", "inventory_items"),
        ("consultations", "parent_id", "consultations"),
        ("procedures", "parent_id", "procedures"),
        ("treatments", "medical_record_id", "medical_records"),
        ("treatments", "consultation_id", "consultations"),
        ("treatments", "procedure_id", "procedures"),
        ("treatments", "medicine_id", "medicines"),
        ("treatments", "inventory_id", "inventory_items"),
        # Phase B harden: regression gates for recent seed ID failures (estimates owners,
        # trimming_courses → trimming_course_types after trigger/serial realign).
        ("estimates", "owner_id", "owners"),
        ("trimming_courses", "course_type_id", "trimming_course_types"),
        ("payments", "payment_method_id", "payment_methods"),
        ("payment_splits", "payment_method_id", "payment_methods"),
    )
    for table, column, ref_table in fk_rules:
        if table not in state.tables or ref_table not in state.tables:
            continue
        missing: list[tuple[int, object]] = []
        for row_id, row in sorted(state.tables[table].items()):
            ref_id = row.get(column)
            if ref_id is None:
                continue
            if ref_id not in state.tables[ref_table]:
                missing.append((row_id, ref_id))
        add_result(errors, not missing, f"{table}.{column}: missing FK targets {missing}")


def check_demo_id_harden_invariants(state: SeedState, errors: list[str]) -> None:
    """Static guards for known 003_demo ID/trigger failure modes.

    1) estimates.owner_id must resolve to an imported owner (≥ IMPORTED_OWNER_ID_FLOOR when
       the demo graph uses the stage-import owner range).
    2) trimming_courses.course_type_id must reference an explicit trimming_course_types row
       in the same clinic (trigger-only types are no longer the source of truth).
    3) payment_splits.payment_method_id must reference payment_methods in the same clinic
       when both rows are present (payments has no clinic_id column).
    """
    owners = state.tables.get("owners", {})
    estimates = state.tables.get("estimates", {})
    owner_ids = _int_keys(owners)
    if estimates and owner_ids and min(owner_ids) >= IMPORTED_OWNER_ID_FLOOR:
        low_owner_refs: list[tuple[int, object]] = []
        for est_id, row in sorted(estimates.items()):
            owner_id = row.get("owner_id")
            if owner_id is None:
                continue
            if not isinstance(owner_id, int) or owner_id < IMPORTED_OWNER_ID_FLOOR:
                low_owner_refs.append((est_id, owner_id))
        add_result(
            errors,
            not low_owner_refs,
            f"estimates.owner_id: below import floor {IMPORTED_OWNER_ID_FLOOR}: {low_owner_refs}",
        )

    courses = state.tables.get("trimming_courses", {})
    course_types = state.tables.get("trimming_course_types", {})
    if courses and course_types:
        clinic_mismatches: list[tuple[int, object, object, object]] = []
        for course_id, row in sorted(courses.items()):
            type_id = row.get("course_type_id")
            if type_id is None:
                continue
            target = course_types.get(type_id)
            if target is None:
                continue  # covered by check_fk_integrity
            if row.get("clinic_id") != target.get("clinic_id"):
                clinic_mismatches.append(
                    (course_id, row.get("clinic_id"), type_id, target.get("clinic_id"))
                )
        add_result(
            errors,
            not clinic_mismatches,
            f"trimming_courses.course_type_id: cross-clinic type refs {clinic_mismatches}",
        )

    splits = state.tables.get("payment_splits", {})
    pay_methods = state.tables.get("payment_methods", {})
    if splits and pay_methods:
        clinic_mismatches: list[tuple[int, object, object, object]] = []
        for split_id, row in sorted(splits.items()):
            method_id = row.get("payment_method_id")
            if method_id is None:
                continue
            target = pay_methods.get(method_id)
            if target is None:
                continue  # covered by check_fk_integrity
            if row.get("clinic_id") != target.get("clinic_id"):
                clinic_mismatches.append(
                    (split_id, row.get("clinic_id"), method_id, target.get("clinic_id"))
                )
        add_result(
            errors,
            not clinic_mismatches,
            f"payment_splits.payment_method_id: cross-clinic method refs {clinic_mismatches}",
        )


def check_cross_tenant(state: SeedState, errors: list[str]) -> None:
    self_ref_rules = (
        ("exam_types", "parent_id"),
        ("vaccines", "parent_id"),
        ("medicines", "parent_id"),
        ("consultations", "parent_id"),
        ("procedures", "parent_id"),
    )
    for table, column in self_ref_rules:
        mismatches: list[tuple[int, object, object]] = []
        for row_id, row in sorted(state.tables[table].items()):
            ref_id = row.get(column)
            if ref_id is None:
                continue
            target = state.tables[table].get(ref_id)
            if target is None:
                continue
            if row.get("clinic_id") != target.get("clinic_id"):
                mismatches.append((row_id, row.get("clinic_id"), target.get("clinic_id")))
        add_result(errors, not mismatches, f"{table}.{column}: cross-tenant references {mismatches}")

    medicine_inventory_mismatches: list[tuple[int, object, object]] = []
    for row_id, row in sorted(state.tables["medicines"].items()):
        ref_id = row.get("inventory_id")
        if ref_id is None:
            continue
        target = state.tables["inventory_items"].get(ref_id)
        if target is None:
            continue
        if row.get("clinic_id") != target.get("clinic_id"):
            medicine_inventory_mismatches.append((row_id, row.get("clinic_id"), target.get("clinic_id")))
    add_result(
        errors,
        not medicine_inventory_mismatches,
        f"medicines.inventory_id: cross-tenant references {medicine_inventory_mismatches}",
    )

    treatment_rules = (
        ("consultation_id", "consultations"),
        ("procedure_id", "procedures"),
        ("medicine_id", "medicines"),
        ("inventory_id", "inventory_items"),
    )
    for column, ref_table in treatment_rules:
        mismatches = []
        for row_id, row in sorted(state.tables["treatments"].items()):
            ref_id = row.get(column)
            if ref_id is None:
                continue
            medical_record = state.tables["medical_records"].get(row["medical_record_id"])
            target = state.tables[ref_table].get(ref_id)
            if medical_record is None or target is None:
                continue
            if medical_record.get("clinic_id") != target.get("clinic_id"):
                mismatches.append((row_id, medical_record.get("clinic_id"), target.get("clinic_id")))
        add_result(errors, not mismatches, f"treatments.{column}: cross-tenant references {mismatches}")


def parse_timestamptz(value: str) -> datetime:
    """Parse a Postgres COPY CSV timestamptz value, e.g. '2026-05-22 09:00:00+09'."""
    if value.endswith(("+00", "+09")):
        base, offset = value[:-3], int(value[-3:])
        return datetime.fromisoformat(base).replace(tzinfo=timezone(timedelta(hours=offset)))
    return datetime.fromisoformat(value).replace(tzinfo=JST)


#  003_seed_demo.sql (pre-CSV) generated 21 synthetic dashboard-statistics
# appointments via `INSERT INTO appointments (...) SELECT ... FROM
# generate_series('2026-05-01', '2026-05-21', '1 day')`, deliberately placed
# at 06:00 with doctor_id=33 — the SQL's own comment reads "既存の枠と衝突
# しないよう 06:00 & doctor_id=33 に設定" (deliberately set to 06:00 &
# doctor_id=33 to avoid colliding with existing slots). The old SQL-replay
# script's regex-based scanner never saw these rows at all (a set-based
# `SELECT ... FROM generate_series` has no literal timestamp strings for its
# regex to match), so it silently never validated them either way. Reading
# the real CSV rows makes them visible, and doctor_id=33 is a safe, verified
# marker to exclude them by (checked: every row with doctor_id=33 starts at
# exactly 06:00, and no other row uses doctor_id=33) — this is intentional
# off-hours synthetic data, not a business-hours violation.
DASHBOARD_STATS_DOCTOR_ID = 33


def check_appointment_time_window(table_index: dict[str, Path], errors: list[str]) -> None:
    """Every appointment must fall within 09:00-19:00 JST on a single day, and
    each clinic-day's hourly distribution must not be too lopsided.

    Reads the single, fully-merged appointments.csv (appointments is 003-owned
    under earliest-file-wins bundle assignment — see cmd/seed-export/
    tables.go) rather than scanning 003 and 004's SQL text separately, since
    both files' contributions are already merged into one CSV. Excludes the
    synthetic dashboard-stats rows (see DASHBOARD_STATS_DOCTOR_ID above) —
    they are intentionally outside business hours by the original seed
    author's own design, not a data defect.
    """
    violations: list[str] = []
    day_hours: dict[str, Counter[int]] = defaultdict(Counter)

    for row in read_rows(table_index, "appointments"):
        if row.get("deleted_at"):
            continue
        if to_int(row.get("doctor_id")) == DASHBOARD_STATS_DOCTOR_ID:
            continue
        # Imported production-history appointments (id ≥ 1e6) are not synthetic
        # demo fixtures; business-hours / hourly-spread gates apply only to
        # hand-authored small-id demo rows.
        appt_id = to_int(row.get("id"))
        if appt_id is not None and appt_id >= IMPORTED_APPOINTMENT_ID_FLOOR:
            continue
        start_raw = row.get("start_time")
        end_raw = row.get("end_time")
        if not start_raw or not end_raw:
            continue
        start = parse_timestamptz(start_raw).astimezone(JST)
        end = parse_timestamptz(end_raw).astimezone(JST)
        start_minutes = start.hour * 60 + start.minute
        end_minutes = end.hour * 60 + end.minute
        day_key = f"appointments:{start.date().isoformat()}"
        day_hours[day_key][start.hour] += 1
        if start.date() != end.date() or start_minutes < 9 * 60 or end_minutes > 19 * 60:
            violations.append(
                f"appointments#{row['id']} {start.strftime('%Y-%m-%d %H:%M')} - {end.strftime('%Y-%m-%d %H:%M')}"
            )

    add_result(errors, not violations, f"appointments: time outside 09:00-19:00 JST {violations[:20]}")

    distribution_violations: list[str] = []
    for day_key, hours in sorted(day_hours.items()):
        hourly_counts = [hours.get(hour, 0) for hour in range(9, 19)]
        spread = max(hourly_counts) - min(hourly_counts)
        if spread > 3:
            distribution_violations.append(f"{day_key} hourly_counts={hourly_counts}")

    add_result(
        errors,
        not distribution_violations,
        f"appointments: uneven daily time distribution {distribution_violations[:20]}",
    )


# check_no_active_trimming_appointments (SQL-replay version) is intentionally
# NOT reimplemented. Its own docstring said it checked only 004's "STG diff
# patch" text section and explicitly stated "Pre-snapshot legacy trimming
# appointments in the base dump are not checked here" — i.e. clinic_id=1
# already has known, accepted legacy trimming appointments (verified: ids
# 101-108, 227-229, 246-248 in the real seed data), and the check's real
# purpose was narrower than "no trimming for clinic 1 ever": it guarded one
# specific regression path (a new trimming appointment being reintroduced via
# a hand-edited SQL patch section, incident cd6c55f1). That path is
# structurally gone post-CSV-migration — 004 no longer has any appointments
# DML at all, so there is no more "patch section" a future edit could slip a
# new row into without a full, reviewed seed-export re-run. Re-implementing
# this against the whole merged CSV would flag the long-accepted legacy rows
# as false positives (confirmed by running it: it fired on exactly the ids
# above, none of which are new). Removed rather than ported as a broader — and
# wrong — check.


def check_audit_log_actor_tenant(table_index: dict[str, Path], errors: list[str]) -> None:
    """Every staff-actor audit_logs row must be owned by the actor's clinic.

    Non-staff actors (actor_type != 'staff') and system rows (actor_id empty)
    are exempt — they are not bound to a single clinic's staff roster.
    """
    staff_clinic = {
        to_int(row["id"]): to_int(row["clinic_id"])
        for row in read_rows(table_index, "staffs")
        if row.get("id", "") != ""
    }
    mismatches: list[str] = []
    for row in read_rows(table_index, "audit_logs"):
        if row.get("actor_type") != "staff":
            continue
        actor_id_raw = row.get("actor_id", "")
        if actor_id_raw == "":
            continue
        actor_id = to_int(actor_id_raw)
        clinic_id = to_int(row["clinic_id"])
        actor_clinic = staff_clinic.get(actor_id)
        if actor_clinic is None:
            mismatches.append(f"actor_id={actor_id} not found (row clinic_id={clinic_id})")
        elif actor_clinic != clinic_id:
            mismatches.append(f"actor_id={actor_id} is clinic_id={actor_clinic} but row clinic_id={clinic_id}")
    add_result(errors, not mismatches, f"audit_logs: cross-tenant actor references {mismatches[:20]}")


def _is_acceptable_combo_vaccine_master_name(name: str) -> bool:
    """True if master name is a real vaccine or an imported-dump placeholder.

    Placeholders: external-clinic markers / RV product codes that do not
    contain the literal substring "ワクチン" but are not filaria-style
    non-vaccine prophylactics (the original #125 failure class).
    """
    if "ワクチン" in name:
        return True
    if "他院" in name:
        return True
    name_norm = name.replace("Ｖ", "V").replace("Ｒ", "R").replace("ｖ", "v").replace("ｒ", "r")
    compact = "".join(name_norm.split()).upper()
    return compact in {"RV", "R.V.", "R.V"}


def check_vaccination_vaccine_category(table_index: dict[str, Path], errors: list[str]) -> None:
    """Combination-vaccine records ("N種混合"/"混合ワクチン" in remarks) must
    reference an actual vaccines row whose name contains "ワクチン" — never a
    prophylactic drug/injection such as "フィラリア予防注射" (#125).

    Imported dump placeholders (他院で接種 / RV product codes) are allowed:
    rewriting historical vaccination FKs by hand is out of TASK-009 static gate.
    """
    vaccines = {
        to_int(row["id"]): row["name"]
        for row in read_rows(table_index, "vaccines")
        if row.get("id", "") != ""
    }
    mismatches: list[str] = []
    for row in read_rows(table_index, "vaccinations"):
        remarks = row.get("remarks") or ""
        if not COMBO_VACCINE_REMARK_RE.search(remarks):
            continue
        vaccine_id = to_int(row["vaccine_id"])
        name = vaccines.get(vaccine_id)
        if name is None:
            mismatches.append(f"vaccination id={row['id']} vaccine_id={vaccine_id} not found in vaccines master")
        elif not _is_acceptable_combo_vaccine_master_name(name):
            mismatches.append(
                f"vaccination id={row['id']} remarks={remarks!r} references vaccine_id={vaccine_id} ({name!r}) which is not a vaccine"
            )
    add_result(
        errors,
        not mismatches,
        f"vaccinations: combination-vaccine records point to non-vaccine entries {mismatches[:20]}",
    )


def classify_exam_field(name: str) -> str | None:
    """Infer an exam_type_field's semantic category from its name, or None if
    unknown. Blood markers are tested first so "BUN（尿素窒素）" classifies as
    blood, not urine, despite containing "尿".
    """
    upper = name.upper()
    if any(tok in upper for tok in EXAM_FIELD_BLOOD_TOKENS) or any(
        tok in name for tok in EXAM_FIELD_BLOOD_TOKENS_JP
    ):
        return "blood"
    if any(tok in name for tok in EXAM_FIELD_ECHO_TOKENS):
        return "echo"
    if any(tok in name for tok in EXAM_FIELD_XRAY_TOKENS):
        return "xray"
    if "尿" in name:
        return "urine"
    return None


def check_exam_type_field_category(table_index: dict[str, Path], errors: list[str]) -> None:
    """Each exam_type_field must bind to a semantically matching exam_type
    (blood/echo/xray/urine — #124).
    """
    exam_types = {
        to_int(row["id"]): row["name"]
        for row in read_rows(table_index, "exam_types")
        if row.get("id", "") != ""
    }
    mismatches: list[str] = []
    for row in read_rows(table_index, "exam_type_fields"):
        name = row.get("name") or ""
        category = classify_exam_field(name)
        if category is None:
            continue
        exam_type_id = to_int(row["exam_type_id"])
        type_name = exam_types.get(exam_type_id)
        if type_name is None:
            mismatches.append(
                f"exam_type_field id={row['id']} exam_type_id={exam_type_id} not found in exam_types master"
            )
            continue
        expected = EXAM_TYPE_EXPECTED_BY_CATEGORY[category]
        if not any(sub in type_name for sub in expected):
            mismatches.append(
                f"exam_type_field id={row['id']} {name!r} ({category}) -> exam_type_id={exam_type_id} ({type_name!r}) is not a {category} exam_type"
            )
    add_result(
        errors,
        not mismatches,
        f"exam_type_fields: fields bound to semantically wrong exam_type {mismatches[:20]}",
    )


def print_summary(state: SeedState, errors: list[str]) -> None:
    tracked_counts = ", ".join(
        f"{table}={len(state.tables[table])}" for table in sorted(TRACKED_TABLES) if state.tables[table]
    )
    if errors:
        print("FAIL")
        print(tracked_counts)
        for message in errors:
            print(f"- {message}")
    else:
        print("OK")
        print(tracked_counts)
        treatment_note = (
            "imported clinical graph (legacy treatment fixtures skipped)"
            if is_imported_clinical_graph(state)
            else "treatments drift fixes"
        )
        print(
            f"verified: 7 masters, {treatment_note}, CHECK equivalent, "
            "procedure presence, FK (incl. estimates.owner_id / trimming_courses.course_type_id / payments.payment_method_id), "
            "demo ID harden, cross-tenant, appointment time window, daily distribution, "
            "audit log actor tenant, vaccination vaccine category, exam_type_field category"
        )


def main() -> int:
    state, table_index = load_seed_state()
    errors: list[str] = []
    check_csv_row_integrity(state, errors)
    check_unique_name_duplicates(state, errors)
    check_expected_treatments(state, errors)
    check_treatment_constraints(state, errors)
    check_procedure_presence(state, errors)
    check_fk_integrity(state, errors)
    check_demo_id_harden_invariants(state, errors)
    check_cross_tenant(state, errors)
    check_appointment_time_window(table_index, errors)
    check_audit_log_actor_tenant(table_index, errors)
    check_vaccination_vaccine_category(table_index, errors)
    check_exam_type_field_category(table_index, errors)
    print_summary(state, errors)
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
