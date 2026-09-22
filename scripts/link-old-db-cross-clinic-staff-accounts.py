#!/usr/bin/env python3
"""Link old_db imported staffs that share one unambiguous name across clinics.

Policy (operator 2026-09-22): one account, memberships in each clinic.
Does not rewrite staffs.id (clinical FKs stay clinic-banded).
Skips names that appear twice in the same clinic, or that have two different
non-empty license numbers.
Never prints staff names.
"""
from __future__ import annotations

import hashlib
import re
import sys
from collections import defaultdict


def normalize_staff_name(name: str) -> str:
    collapsed = (name or "").replace("\u3000", " ").strip()
    return re.sub(r"\s+", " ", collapsed)


def eligible_cross_clinic_groups(rows: list[dict]) -> list[list[dict]]:
    """rows: id, clinic_id, name, license_number, account_id (optional)."""
    by_name: dict[str, list[dict]] = defaultdict(list)
    for row in rows:
        key = normalize_staff_name(str(row.get("name") or ""))
        if not key:
            continue
        by_name[key].append(row)

    groups: list[list[dict]] = []
    for members in by_name.values():
        per_clinic: dict[object, list[dict]] = defaultdict(list)
        for member in members:
            per_clinic[member["clinic_id"]].append(member)
        if len(per_clinic) < 2:
            continue
        if any(len(v) != 1 for v in per_clinic.values()):
            continue
        licenses = {
            str(m.get("license_number") or "").strip()
            for m in members
            if str(m.get("license_number") or "").strip()
        }
        if len(licenses) > 1:
            continue
        account_ids = {m.get("account_id") for m in members if m.get("account_id") not in (None, "")}
        if len(account_ids) > 1:
            continue
        groups.append(members)
    return groups


def account_email_for_name_key(name_key: str) -> str:
    digest = hashlib.md5(name_key.encode("utf-8")).hexdigest()
    return f"olddb-link-{digest}@invalid.local"


def main(argv: list[str]) -> int:
    if len(argv) == 2 and argv[1] == "--self-test":
        return 0 if _self_test() else 1
    print("link-old-db-cross-clinic-staff-accounts: use --self-test or the reset hook", file=sys.stderr)
    return 2


def _self_test() -> bool:
    rows = [
        {"id": 1, "clinic_id": 1, "name": "林　文明", "license_number": "", "account_id": None},
        {"id": 2, "clinic_id": 2, "name": "林 文明", "license_number": "", "account_id": None},
        {"id": 3, "clinic_id": 1, "name": "山田", "license_number": "", "account_id": None},
        {"id": 4, "clinic_id": 2, "name": "山田", "license_number": "A", "account_id": None},
        {"id": 5, "clinic_id": 2, "name": "山田", "license_number": "B", "account_id": None},
        {"id": 6, "clinic_id": 1, "name": "佐藤", "license_number": "X", "account_id": None},
        {"id": 7, "clinic_id": 2, "name": "佐藤", "license_number": "Y", "account_id": None},
        {"id": 8, "clinic_id": 3, "name": "鈴木", "license_number": "", "account_id": None},
    ]
    groups = eligible_cross_clinic_groups(rows)
    ids = {tuple(sorted(m["id"] for m in g)) for g in groups}
    assert ids == {(1, 2)}, ids
    assert account_email_for_name_key("林 文明").startswith("olddb-link-")
    assert "@invalid.local" in account_email_for_name_key("林 文明")
    return True


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
