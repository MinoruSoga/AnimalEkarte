#!/usr/bin/env python3
"""Activate migrated staffs using the authoritative old_db staff-activity map.

Policy (old_db artifact contract, schemaVersion staff-activity-map-manifest-v1):
set is_active=true for ACTIVE_CONFIRMED rows only. ACTIVE_LIKELY /
LIKELY_RETIRED / NO_EVIDENCE rows are never touched and this path never
deactivates. staffs.id / accounts / assignments / reservation_visible stay
untouched.

This file mirrors the selection policy of
scripts/sql/apply-old-db-staff-activity-map.sql for offline verification.
Never prints staff names.
"""
from __future__ import annotations

import csv
import sys

KNOWN_CLASSIFICATIONS = {
    "ACTIVE_CONFIRMED",
    "ACTIVE_LIKELY",
    "LIKELY_RETIRED",
    "NO_EVIDENCE",
}


def active_confirmed_ids(rows: list[dict], live_staff: dict[int, int]) -> list[int]:
    """Return sorted csv_staff_id list to activate.

    rows: map records with csv_staff_id/classification.
    live_staff: csv_staff_id -> clinic_id for non-deleted staffs rows.
    Raises ValueError on malformed input (fail-closed).
    """
    seen: set[int] = set()
    for r in rows:
        if r.get("classification") not in KNOWN_CLASSIFICATIONS:
            raise ValueError(f"unknown classification: {r.get('classification')}")
        staff_id = int(r["csv_staff_id"])
        if staff_id in seen:
            raise ValueError(f"duplicate csv_staff_id {staff_id}")
        seen.add(staff_id)

    out: list[int] = []
    for r in rows:
        if r["classification"] != "ACTIVE_CONFIRMED":
            continue
        staff_id = int(r["csv_staff_id"])
        if staff_id not in live_staff:
            raise ValueError(f"ACTIVE_CONFIRMED csv_staff_id {staff_id} not a live staff")
        out.append(staff_id)
    return sorted(out)


def load_map_csv(path: str) -> list[dict]:
    with open(path, newline="", encoding="utf-8") as fh:
        return list(csv.DictReader(fh))


def main(argv: list[str]) -> int:
    if len(argv) == 2 and argv[1] == "--self-test":
        return 0 if _self_test() else 1
    print("apply-old-db-staff-activity-map: use --self-test", file=sys.stderr)
    return 2


def _self_test() -> bool:
    live = {11: 1, 21: 2, 31: 3, 12: 1, 99: 2}
    rows = [
        {"csv_staff_id": "11", "classification": "ACTIVE_CONFIRMED"},
        {"csv_staff_id": "21", "classification": "ACTIVE_CONFIRMED"},
        {"csv_staff_id": "31", "classification": "ACTIVE_LIKELY"},
        {"csv_staff_id": "12", "classification": "LIKELY_RETIRED"},
        {"csv_staff_id": "99", "classification": "NO_EVIDENCE"},
    ]
    assert active_confirmed_ids(rows, live) == [11, 21]

    for bad in (
        rows + [{"csv_staff_id": "42", "classification": "ACTIVE_CONFIRMED"}],
        rows + [{"csv_staff_id": "11", "classification": "NO_EVIDENCE"}],
        [{"csv_staff_id": "11", "classification": "BOGUS"}],
    ):
        try:
            active_confirmed_ids(bad, live)
        except ValueError:
            pass
        else:
            return False
    return True


if __name__ == "__main__":
    sys.exit(main(sys.argv))
