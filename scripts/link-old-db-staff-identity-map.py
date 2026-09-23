#!/usr/bin/env python3
"""Link old_db imported staffs using the authoritative cross-clinic identity map.

Policy (old_db artifact contract, schemaVersion staff-identity-map-v1):
apply CONFIRMED groups only — NEEDS_REVIEW and UNRESOLVED rows keep their
own accounts. One accounts row per group shared by all member staffs rows
(staffs.id is not rewritten; clinical FKs stay clinic-banded), and every
member row gets staff_clinic_assignments to all member clinics.

This file mirrors the selection policy of
scripts/sql/link-old-db-staff-identity-map.sql for offline verification.
Never prints staff names.
"""
from __future__ import annotations

import csv
import hashlib
import sys
from collections import defaultdict


def confirmed_groups(rows: list[dict], live_staff: dict[int, int]) -> dict[str, list[int]]:
    """Return person_group_id -> [csv_staff_id] for linkable CONFIRMED groups.

    rows: map records with person_group_id/classification/clinic_code/csv_staff_id.
    live_staff: csv_staff_id -> clinic_id for non-deleted staffs rows.
    Raises ValueError on malformed CONFIRMED input (fail-closed).
    """
    confirmed = [r for r in rows if r.get("classification") == "CONFIRMED"]
    for r in confirmed:
        if r.get("clinic_code") == "hachioji":
            raise ValueError("hachioji rows must not be CONFIRMED")
        staff_id = int(r["csv_staff_id"])
        if staff_id not in live_staff:
            raise ValueError(f"CONFIRMED csv_staff_id {staff_id} not a live staff")

    by_group: dict[str, list[dict]] = defaultdict(list)
    for r in confirmed:
        by_group[r["person_group_id"]].append(r)

    for group_id, members in by_group.items():
        ids = [int(m["csv_staff_id"]) for m in members]
        if len(ids) != len(set(ids)):
            raise ValueError(f"{group_id}: duplicate csv_staff_id")
        clinics = {live_staff[i] for i in ids}
        if len(members) < 2 or len(clinics) < 2:
            raise ValueError(f"{group_id}: needs >=2 members at distinct clinics")

    seen: dict[int, str] = {}
    for group_id, members in by_group.items():
        for m in members:
            staff_id = int(m["csv_staff_id"])
            if staff_id in seen:
                raise ValueError(f"csv_staff_id {staff_id} in groups {seen[staff_id]} and {group_id}")
            seen[staff_id] = group_id

    return {g: sorted(int(m["csv_staff_id"]) for m in ms) for g, ms in by_group.items()}


def resolve_account(member_account_ids: list[int | None]) -> int | None:
    """Existing account to keep, or None when a placeholder must be created."""
    existing = [a for a in member_account_ids if a]
    return min(existing) if existing else None


def account_email_for_group(person_group_id: str) -> str:
    digest = hashlib.md5(person_group_id.encode("utf-8")).hexdigest()
    return f"olddb-map-{digest}@invalid.local"


def load_map_csv(path: str) -> list[dict]:
    with open(path, newline="", encoding="utf-8") as fh:
        return list(csv.DictReader(fh))


def main(argv: list[str]) -> int:
    if len(argv) == 2 and argv[1] == "--self-test":
        return 0 if _self_test() else 1
    print("link-old-db-staff-identity-map: use --self-test", file=sys.stderr)
    return 2


def _self_test() -> bool:
    live = {11: 1, 21: 2, 31: 3, 12: 1, 99: 2}
    rows = [
        {"person_group_id": "PG-S0001", "classification": "CONFIRMED", "clinic_code": "jouto", "csv_staff_id": "11"},
        {"person_group_id": "PG-S0001", "classification": "CONFIRMED", "clinic_code": "shikishima", "csv_staff_id": "21"},
        {"person_group_id": "PG-S0001", "classification": "CONFIRMED", "clinic_code": "hakobuneco", "csv_staff_id": "31"},
        {"person_group_id": "PG-H0001", "classification": "NEEDS_REVIEW", "clinic_code": "hachioji", "csv_staff_id": "12"},
        {"person_group_id": "PG-H0002", "classification": "UNRESOLVED", "clinic_code": "hachioji", "csv_staff_id": "99"},
    ]
    groups = confirmed_groups(rows, live)
    assert groups == {"PG-S0001": [11, 21, 31]}, groups

    # resolve_account keeps the smallest existing account id
    assert resolve_account([None, 50, 30]) == 30
    assert resolve_account([None, None]) is None
    assert account_email_for_group("PG-S0001").startswith("olddb-map-")

    # fail-closed guards
    for bad in (
        rows + [{"person_group_id": "PG-X", "classification": "CONFIRMED", "clinic_code": "hachioji", "csv_staff_id": "12"}],
        rows + [{"person_group_id": "PG-Y", "classification": "CONFIRMED", "clinic_code": "jouto", "csv_staff_id": "42"}],
        [
            {"person_group_id": "PG-Z", "classification": "CONFIRMED", "clinic_code": "jouto", "csv_staff_id": "11"},
            {"person_group_id": "PG-W", "classification": "CONFIRMED", "clinic_code": "shikishima", "csv_staff_id": "11"},
        ],
        [{"person_group_id": "PG-S", "classification": "CONFIRMED", "clinic_code": "jouto", "csv_staff_id": "11"}],
    ):
        try:
            confirmed_groups(bad, live)
        except ValueError:
            pass
        else:
            return False

    # duplicate id inside one group is rejected
    dup = rows[:2] + [
        {"person_group_id": "PG-S0001", "classification": "CONFIRMED", "clinic_code": "hakobuneco", "csv_staff_id": "11"},
    ]
    try:
        confirmed_groups(dup, live)
    except ValueError:
        pass
    else:
        return False
    return True


if __name__ == "__main__":
    sys.exit(main(sys.argv))
