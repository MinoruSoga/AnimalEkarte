#!/usr/bin/env python3
"""Static seed verification for master/demo migrations.

This script does not touch a database. It replays relevant INSERT statements from
the seed SQL files, taking ON CONFLICT behavior into account, and validates the
expected final state of the master/demo seed data.
"""

from __future__ import annotations

from collections import Counter, defaultdict
from dataclasses import dataclass
from decimal import Decimal
from pathlib import Path
import re
import sys


ROOT = Path(__file__).resolve().parents[1]
MIGRATION_FILES = [
    ROOT / "backend/migrations/002_seed_master.sql",
    ROOT / "backend/migrations/003_seed_demo.sql",
]

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


@dataclass(frozen=True)
class InsertStatement:
    table: str
    columns: tuple[str, ...]
    rows: tuple[tuple[object, ...], ...]
    conflict_action: str | None


class SeedState:
    def __init__(self) -> None:
        self.tables: dict[str, dict[int, dict[str, object]]] = defaultdict(dict)
        self.source_id_counts: dict[str, Counter[int]] = defaultdict(Counter)

    def apply(self, stmt: InsertStatement) -> None:
        if stmt.table not in TRACKED_TABLES:
            return

        for values in stmt.rows:
            row = dict(zip(stmt.columns, values, strict=True))
            row_id = row.get("id")
            if not isinstance(row_id, int):
                continue

            self.source_id_counts[stmt.table][row_id] += 1
            existing = self.tables[stmt.table].get(row_id)
            if existing is None:
                self.tables[stmt.table][row_id] = row
                continue

            if stmt.conflict_action == "update":
                merged = existing.copy()
                merged.update(row)
                self.tables[stmt.table][row_id] = merged


def split_sql_statements(text: str) -> list[str]:
    statements: list[str] = []
    buf: list[str] = []
    in_string = False
    i = 0
    while i < len(text):
        ch = text[i]
        buf.append(ch)
        if ch == "'":
            if in_string and i + 1 < len(text) and text[i + 1] == "'":
                buf.append(text[i + 1])
                i += 1
            else:
                in_string = not in_string
        elif ch == ";" and not in_string:
            statement = "".join(buf).strip()
            if statement:
                statements.append(statement)
            buf = []
        i += 1
    tail = "".join(buf).strip()
    if tail:
        statements.append(tail)
    return statements


def find_keyword_outside_quotes(text: str, keyword: str) -> int:
    upper = text.upper()
    keyword_upper = keyword.upper()
    in_string = False
    i = 0
    while i < len(text):
        ch = text[i]
        if ch == "'":
            if in_string and i + 1 < len(text) and text[i + 1] == "'":
                i += 2
                continue
            in_string = not in_string
        if not in_string and upper.startswith(keyword_upper, i):
            return i
        i += 1
    return -1


def parse_insert_statement(statement: str) -> InsertStatement | None:
    match = re.match(
        r"INSERT INTO\s+([a-z_]+)\s*\((.*?)\)\s*VALUES\s*(.*)",
        statement,
        flags=re.IGNORECASE | re.DOTALL,
    )
    if not match:
        return None

    table = match.group(1)
    columns = tuple(part.strip() for part in match.group(2).split(","))
    remainder = match.group(3).strip()

    conflict_index = find_keyword_outside_quotes(remainder, "ON CONFLICT")
    if conflict_index >= 0:
        values_text = remainder[:conflict_index].rstrip()
        conflict_text = remainder[conflict_index:].upper()
        if "DO UPDATE" in conflict_text:
            conflict_action = "update"
        elif "DO NOTHING" in conflict_text:
            conflict_action = "nothing"
        else:
            conflict_action = None
    else:
        values_text = remainder
        conflict_action = None

    rows = tuple(parse_values_block(strip_sql_line_comments(values_text)))
    return InsertStatement(table=table, columns=columns, rows=rows, conflict_action=conflict_action)


def parse_values_block(values_text: str) -> list[tuple[object, ...]]:
    chunks: list[str] = []
    start = -1
    depth = 0
    in_string = False
    i = 0
    while i < len(values_text):
        ch = values_text[i]
        if ch == "'":
            if in_string and i + 1 < len(values_text) and values_text[i + 1] == "'":
                i += 2
                continue
            in_string = not in_string
        elif not in_string:
            if ch == "(":
                if depth == 0:
                    start = i + 1
                depth += 1
            elif ch == ")":
                depth -= 1
                if depth == 0 and start >= 0:
                    chunks.append(values_text[start:i])
        i += 1

    return [tuple(parse_tuple(chunk)) for chunk in chunks]


def strip_sql_line_comments(text: str) -> str:
    out: list[str] = []
    in_string = False
    i = 0
    while i < len(text):
        ch = text[i]
        if ch == "'":
            out.append(ch)
            if in_string and i + 1 < len(text) and text[i + 1] == "'":
                out.append(text[i + 1])
                i += 2
                continue
            in_string = not in_string
            i += 1
            continue
        if not in_string and ch == "-" and i + 1 < len(text) and text[i + 1] == "-":
            while i < len(text) and text[i] != "\n":
                i += 1
            continue
        out.append(ch)
        i += 1
    return "".join(out)


def parse_tuple(chunk: str) -> list[object]:
    tokens: list[str] = []
    buf: list[str] = []
    in_string = False
    depth = 0
    i = 0
    while i < len(chunk):
        ch = chunk[i]
        if ch == "'":
            buf.append(ch)
            if in_string and i + 1 < len(chunk) and chunk[i + 1] == "'":
                buf.append(chunk[i + 1])
                i += 2
                continue
            in_string = not in_string
        elif not in_string and ch == "(":
            depth += 1
            buf.append(ch)
        elif not in_string and ch == ")":
            depth -= 1
            buf.append(ch)
        elif not in_string and depth == 0 and ch == ",":
            tokens.append("".join(buf).strip())
            buf = []
        else:
            buf.append(ch)
        i += 1
    if buf:
        tokens.append("".join(buf).strip())
    return [parse_literal(token) for token in tokens]


def parse_literal(token: str) -> object:
    upper = token.upper()
    lower = token.lower()
    if upper == "NULL":
        return None
    if lower == "true":
        return True
    if lower == "false":
        return False
    if token.startswith("'") and token.endswith("'"):
        return token[1:-1].replace("''", "'")
    if re.fullmatch(r"-?\d+", token):
        return int(token)
    if re.fullmatch(r"-?\d+\.\d+", token):
        return Decimal(token)
    return token


def load_seed_state() -> SeedState:
    state = SeedState()
    for path in MIGRATION_FILES:
        text = path.read_text(encoding="utf-8")
        for statement in split_sql_statements(text):
            insert_index = statement.upper().find("INSERT INTO")
            if insert_index < 0:
                continue
            trimmed = statement[insert_index:].strip()
            table_match = re.match(r"INSERT INTO\s+([a-z_]+)", trimmed, flags=re.IGNORECASE)
            if table_match is None or table_match.group(1) not in TRACKED_TABLES:
                continue
            parsed = parse_insert_statement(trimmed)
            if parsed is not None:
                state.apply(parsed)
    return state


def is_unique_scope_row(table: str, row: dict[str, object]) -> bool:
    if row.get("deleted_at") is not None:
        return False
    if table == "merchandise_items":
        return row.get("is_active") is True
    return True


def add_result(errors: list[str], condition: bool, message: str) -> None:
    if not condition:
        errors.append(message)


def check_source_id_duplicates(state: SeedState, errors: list[str]) -> None:
    for table in SEVEN_MASTER_TABLES:
        duplicates = sorted(row_id for row_id, count in state.source_id_counts[table].items() if count > 1)
        add_result(errors, not duplicates, f"{table}: source id duplicates found: {duplicates}")


def check_unique_name_duplicates(state: SeedState, errors: list[str]) -> None:
    for table in SEVEN_MASTER_TABLES:
        seen: Counter[tuple[object, object]] = Counter()
        for row in state.tables[table].values():
            if is_unique_scope_row(table, row):
                seen[(row.get("clinic_id"), row.get("name"))] += 1
        duplicates = sorted(key for key, count in seen.items() if count > 1)
        add_result(errors, not duplicates, f"{table}: effective (clinic_id, name) duplicates found: {duplicates}")


def check_expected_treatments(state: SeedState, errors: list[str]) -> None:
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
    )
    for table, column, ref_table in fk_rules:
        missing: list[tuple[int, object]] = []
        for row_id, row in sorted(state.tables[table].items()):
            ref_id = row.get(column)
            if ref_id is None:
                continue
            if ref_id not in state.tables[ref_table]:
                missing.append((row_id, ref_id))
        add_result(errors, not missing, f"{table}.{column}: missing FK targets {missing}")


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
        mismatches: list[tuple[int, object, object]] = []
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
        print("verified: 7 masters, treatments drift fixes, CHECK equivalent, procedure presence, FK, cross-tenant")


def main() -> int:
    state = load_seed_state()
    errors: list[str] = []
    check_source_id_duplicates(state, errors)
    check_unique_name_duplicates(state, errors)
    check_expected_treatments(state, errors)
    check_treatment_constraints(state, errors)
    check_procedure_presence(state, errors)
    check_fk_integrity(state, errors)
    check_cross_tenant(state, errors)
    print_summary(state, errors)
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
