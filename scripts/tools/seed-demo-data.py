#!/usr/bin/env python3
"""Seed synthetic demo data through the application API.

USER-RUN. This script never writes the database directly. It logs in as a
demo/operator account, sends clinic-scoped API requests, and records created
IDs in a gitignored state file so cleanup can delete only this run's rows.

Dry-run (default):
  DEMO_SEED_EMAIL=... DEMO_SEED_PASSWORD=... \\
    python3 scripts/tools/seed-demo-data.py --clinic-id 1

Apply:
  DEMO_SEED_EMAIL=... DEMO_SEED_PASSWORD=... \\
    python3 scripts/tools/seed-demo-data.py --clinic-id 1 --apply

Cleanup:
  ... --clinic-id 1 --cleanup --apply
"""
from __future__ import annotations

import argparse
import http.cookiejar
import json
import os
import sys
import urllib.error
import urllib.request
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timedelta, timezone
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_STATE_ROOT = ROOT / "sensitive-local/demo-seed"
JST = ZoneInfo("Asia/Tokyo")
RUN_TAG = "EMR443-DEMO"


def die(message: str) -> None:
    print(f"BLOCKED: {message}", file=sys.stderr)
    raise SystemExit(1)


@dataclass
class State:
    run_tag: str
    clinic_id: int
    created: list[dict] = field(default_factory=list)
    updated: list[dict] = field(default_factory=list)

    def record_created(self, kind: str, resource_id: int, name: str) -> None:
        self.created.append({"kind": kind, "id": resource_id, "name": name})

    def record_updated(self, kind: str, resource_id: int, name: str) -> None:
        self.updated.append({"kind": kind, "id": resource_id, "name": name})

    def save(self, path: Path) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(self.__dict__, ensure_ascii=False, indent=2) + "\n")
        path.chmod(0o600)


class ApiClient:
    def __init__(self, base_url: str, clinic_id: int) -> None:
        self.base_url = base_url.rstrip("/")
        self.clinic_id = clinic_id
        self.cookies = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.cookies)
        )

    def request(
        self,
        method: str,
        path: str,
        payload: dict | None = None,
        headers: dict[str, str] | None = None,
    ) -> tuple[int, dict | list | None]:
        data = None if payload is None else json.dumps(payload).encode("utf-8")
        request_headers = {
            "Accept": "application/json",
            "Content-Type": "application/json",
            "User-Agent": "animal-ekarte-demo-seed/1.0",
            "X-Requested-With": "XMLHttpRequest",
            "X-Clinic-ID": str(self.clinic_id),
        }
        if headers:
            request_headers.update(headers)
        req = urllib.request.Request(
            f"{self.base_url}{path}",
            data=data,
            method=method,
            headers=request_headers,
        )
        try:
            with self.opener.open(req, timeout=30) as res:
                body = res.read()
                return res.status, json.loads(body) if body else None
        except urllib.error.HTTPError as exc:
            body = exc.read().decode("utf-8", errors="replace")
            return exc.code, {"error": body[:500]}

    def login(self, email: str, password: str) -> None:
        code, body = self.request("POST", "/api/v1/login", {"email": email, "password": password})
        if code != 200:
            die(f"login failed: HTTP {code} {str(body)[:300]}")

    def rows(self, path: str) -> list[dict]:
        separator = "&" if "?" in path else "?"
        collected: list[dict] = []
        page = 1
        while True:
            code, body = self.request("GET", f"{path}{separator}limit=100&page={page}")
            if code != 200:
                die(f"GET {path}: HTTP {code} {str(body)[:300]}")
            if isinstance(body, list):
                return body
            if not isinstance(body, dict):
                return collected
            items = None
            for key in ("data", "items", "results"):
                value = body.get(key)
                if isinstance(value, list):
                    items = value
                    break
            if items is None:
                return collected
            collected.extend(items)
            total = body.get("total")
            if total is None or len(collected) >= int(total) or not items:
                return collected
            page += 1

    def find_by_name(self, path: str, name: str) -> dict | None:
        for row in self.rows(path):
            if row.get("name") == name or row.get("title") == name or row.get("owner_name") == name:
                return row
        return None

    def upsert(self, path: str, name: str, payload: dict, state: State, kind: str) -> int:
        existing = self.find_by_name(path, name)
        if existing:
            resource_id = int(existing["id"])
            update_payload = {key: value for key, value in payload.items() if key not in {"name"}}
            code, body = self.request("PATCH", f"{path}/{resource_id}", update_payload)
            if code not in (200, 204):
                die(f"PATCH {path}/{resource_id}: HTTP {code} {str(body)[:300]}")
            state.record_updated(kind, resource_id, name)
            print(f"updated {kind}: {name} (#{resource_id})")
            return resource_id
        code, body = self.request("POST", path, payload)
        if code not in (200, 201):
            die(f"POST {path}: HTTP {code} {str(body)[:300]}")
        data = body.get("data") if isinstance(body, dict) and isinstance(body.get("data"), dict) else body
        resource_id = int(data["id"])
        state.record_created(kind, resource_id, name)
        print(f"created {kind}: {name} (#{resource_id})")
        return resource_id

    def create(self, path: str, payload: dict) -> dict:
        code, body = self.request("POST", path, payload)
        if code not in (200, 201):
            die(f"POST {path}: HTTP {code} {str(body)[:300]}")
        data = body.get("data") if isinstance(body, dict) and isinstance(body.get("data"), dict) else body
        return data


MASTER_SEEDS: list[tuple[str, str, list[dict]]] = [
    (
        "procedure",
        "/api/v1/masters/procedures",
        [
            {"name": "【デモ】一般診察", "price": 2000, "is_active": True, "description": "デモ用", "duration": 15, "anesthesia": "none", "is_surgery": False, "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 1},
            {"name": "【デモ】血液検査", "price": 4500, "is_active": True, "description": "デモ用", "duration": 30, "anesthesia": "none", "is_surgery": False, "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 2},
            {"name": "【デモ】レントゲン", "price": 6000, "is_active": True, "description": "デモ用", "duration": 30, "anesthesia": "none", "is_surgery": False, "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 3},
            {"name": "【デモ】避妊手術", "price": 25000, "is_active": True, "description": "デモ用", "duration": 90, "anesthesia": "general", "is_surgery": True, "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 4},
        ],
    ),
    (
        "medicine",
        "/api/v1/masters/medicines",
        [
            {"name": "【デモ】アモキシシリン", "price": 800, "is_active": True, "description": "デモ用", "dosage_form": "tablet", "medicine_unit": "per_tablet", "default_quantity": "1", "tax_type": "excluded", "tax_rate": "0.10", "sort_order": 1, "is_non_insurance": False, "calculation_type": "none"},
            {"name": "【デモ】メトクロプラミド", "price": 600, "is_active": True, "description": "デモ用", "dosage_form": "injection", "medicine_unit": "per_dose", "default_quantity": "1", "tax_type": "excluded", "tax_rate": "0.10", "sort_order": 2, "is_non_insurance": False, "calculation_type": "none"},
            {"name": "【デモ】プレドニゾロン", "price": 500, "is_active": True, "description": "デモ用", "dosage_form": "tablet", "medicine_unit": "per_tablet", "default_quantity": "1", "tax_type": "excluded", "tax_rate": "0.10", "sort_order": 3, "is_non_insurance": False, "calculation_type": "none"},
        ],
    ),
    (
        "vaccine",
        "/api/v1/masters/vaccines",
        [
            {"name": "【デモ】混合ワクチン（犬）", "price": 5000, "is_active": True, "description": "デモ用", "species": "dog", "interval": "1年", "sort_order": 1},
            {"name": "【デモ】狂犬病ワクチン", "price": 3500, "is_active": True, "description": "デモ用", "species": "dog", "interval": "1年", "sort_order": 2},
            {"name": "【デモ】混合ワクチン（猫）", "price": 4500, "is_active": True, "description": "デモ用", "species": "cat", "interval": "1年", "sort_order": 3},
        ],
    ),
    (
        "insurance",
        "/api/v1/masters/insurances",
        [
            {"name": "【デモ】ペット保険A", "is_active": True, "description": "デモ用", "coverage_rate": 70, "contact_phone": "", "sort_order": 1},
            {"name": "【デモ】ペット保険B", "is_active": True, "description": "デモ用", "coverage_rate": 50, "contact_phone": "", "sort_order": 2},
        ],
    ),
    (
        "diagnosis_type",
        "/api/v1/masters/diagnosis-types",
        [
            {"name": "【デモ】内科疾患", "is_active": True, "sort_order": 1},
            {"name": "【デモ】外科疾患", "is_active": True, "sort_order": 2},
            {"name": "【デモ】皮膚疾患", "is_active": True, "sort_order": 3},
        ],
    ),
    (
        "chief_complaint",
        "/api/v1/masters/chief-complaint-types",
        [
            {"name": "【デモ】食欲不振", "is_active": True, "sort_order": 1},
            {"name": "【デモ】呕吐", "is_active": True, "sort_order": 2},
            {"name": "【デモ】下痢", "is_active": True, "sort_order": 3},
            {"name": "【デモ】咳嗽", "is_active": True, "sort_order": 4},
            {"name": "【デモ】皮膚のかゆみ", "is_active": True, "sort_order": 5},
        ],
    ),
    (
        "cage",
        "/api/v1/masters/cages",
        [
            {"name": "【デモ】小型ケージ1", "price": 3000, "is_active": True, "description": "デモ用", "cage_type": "general", "cage_size": "small", "sort_order": 1},
            {"name": "【デモ】中型ケージ1", "price": 4000, "is_active": True, "description": "デモ用", "cage_type": "general", "cage_size": "medium", "sort_order": 2},
            {"name": "【デモ】ICUケージ1", "price": 8000, "is_active": True, "description": "デモ用", "cage_type": "icu", "cage_size": "medium", "sort_order": 3},
        ],
    ),
    (
        "hospitalization_plan",
        "/api/v1/masters/hospitalization-plans",
        [
            {"name": "【デモ】入院基本料（小型）", "price": 8000, "is_active": True, "description": "デモ用", "body_size": "small", "billing_unit": "per_day", "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 1},
            {"name": "【デモ】入院基本料（中型）", "price": 10000, "is_active": True, "description": "デモ用", "body_size": "medium", "billing_unit": "per_day", "tax_type": "excluded", "tax_rate": 0.1, "sort_order": 2},
        ],
    ),
    (
        "inventory",
        "/api/v1/inventory",
        [
            {"name": "【デモ】アモキシシリン錠", "category": "medicine", "quantity": 100, "unit": "錠", "min_stock_level": 20, "location": "薬品棚"},
            {"name": "【デモ】輸液セット", "category": "consumable", "quantity": 30, "unit": "セット", "min_stock_level": 10, "location": "処置室"},
            {"name": "【デモ】療法食（腎臓）", "category": "food", "quantity": 12, "unit": "袋", "min_stock_level": 3, "location": "倉庫"},
        ],
    ),
    (
        "inquiry_template",
        "/api/v1/masters/inquiry-templates",
        [
            {"category": "general", "title": "【デモ】初診問診", "content": "食欲・飲水量・排泄・既往歴を確認", "is_active": True, "sort_order": 1},
            {"category": "general", "title": "【デモ】再診問診", "content": "前回からの変化を確認", "is_active": True, "sort_order": 2},
        ],
    ),
    (
        "checkup_type",
        "/api/v1/masters/checkup-types",
        [
            {"name": "【デモ】犬の健康診断", "price": 8000, "is_active": True, "interval": "1年", "target_age": "全年齢", "description": "デモ用", "sort_order": 1},
            {"name": "【デモ】猫の健康診断", "price": 8000, "is_active": True, "interval": "1年", "target_age": "全年齢", "description": "デモ用", "sort_order": 2},
        ],
    ),
    (
        "trimming_course",
        "/api/v1/masters/trimming-courses",
        [
            {"name": "【デモ】シャンプーコース", "price": 5000, "is_active": True, "description": "デモ用", "target_size": "small", "duration": 60, "sort_order": 1},
            {"name": "【デモ】シャンプー＆カット", "price": 8000, "is_active": True, "description": "デモ用", "target_size": "medium", "duration": 90, "sort_order": 2},
        ],
    ),
    (
        "reservation_type_group",
        "/api/v1/masters/reservation-type-groups",
        [
            {"name": "【デモ】診療", "color": "#3B82F6", "is_active": True, "sort_order": 1},
            {"name": "【デモ】予防・健診", "color": "#22C55E", "is_active": True, "sort_order": 2},
        ],
    ),
]

DIAGNOSIS_NAMES = [
    ("【デモ】胃炎", "【デモ】内科疾患"),
    ("【デモ】膀胱炎", "【デモ】内科疾患"),
    ("【デモ】皮膚炎", "【デモ】皮膚疾患"),
]

OWNERS = [
    {
        "owner_name": "【デモ】山田 太郎",
        "owner_name_kana": "ヤマダ タロウ",
        "phone": "09000000001",
        "email": "demo-owner-1@example.test",
        "address1": "東京都デモ市1-1-1",
        "remarks": f"{RUN_TAG} synthetic owner",
        "pets": [
            {"name": "【デモ】ポチ", "species": "犬", "breed": "柴犬", "gender": "male", "birth_date": "2020-04-01", "weight": 10.5},
            {"name": "【デモ】ミケ", "species": "猫", "breed": "三毛猫", "gender": "female", "birth_date": "2021-06-15", "weight": 4.2},
        ],
    },
    {
        "owner_name": "【デモ】佐藤 花子",
        "owner_name_kana": "サトウ ハナコ",
        "phone": "09000000002",
        "email": "demo-owner-2@example.test",
        "address1": "東京都デモ市2-2-2",
        "remarks": f"{RUN_TAG} synthetic owner",
        "pets": [
            {"name": "【デモ】レオ", "species": "犬", "breed": "トイプードル", "gender": "male", "birth_date": "2019-11-20", "weight": 5.1},
        ],
    },
    {
        "owner_name": "【デモ】鈴木 一郎",
        "owner_name_kana": "スズキ イチロウ",
        "phone": "09000000003",
        "email": "demo-owner-3@example.test",
        "address1": "東京都デモ市3-3-3",
        "remarks": f"{RUN_TAG} synthetic owner",
        "pets": [
            {"name": "【デモ】モモ", "species": "猫", "breed": "茶トラ", "gender": "female", "birth_date": "2022-02-02", "weight": 3.8},
        ],
    },
]


def build_plan() -> list[str]:
    lines = ["masters:"]
    for kind, path, items in MASTER_SEEDS:
        lines.append(f"  {kind}: {len(items)} ({path})")
    lines.append("diagnosis_names: " + str(len(DIAGNOSIS_NAMES)))
    lines.append(f"owners: {len(OWNERS)}")
    lines.append(f"pets: {sum(len(owner['pets']) for owner in OWNERS)}")
    lines.append("reservations: 3")
    lines.append("scenario: medical records / billing / vaccines / checkups / exams / trimming / hospitalization (seed_scenario on --apply)")
    return lines


def seed_masters(client: ApiClient, state: State) -> dict[str, dict[str, int]]:
    ids: dict[str, dict[str, int]] = {}
    for kind, path, items in MASTER_SEEDS:
        ids[kind] = {}
        for item in items:
            name = item["name"] if "name" in item else item["title"]
            resource_id = client.upsert(path, name, item, state, kind)
            ids[kind][name] = resource_id
    for name, type_name in DIAGNOSIS_NAMES:
        type_id = ids["diagnosis_type"][type_name]
        payload = {"name": name, "diagnosis_type_id": type_id, "is_active": True, "sort_order": 1}
        ids.setdefault("diagnosis_name", {})[name] = client.upsert(
            "/api/v1/masters/diagnosis-names", name, payload, state, "diagnosis_name"
        )
    return ids


def find_species_id(client: ApiClient, name: str) -> int:
    for row in client.rows("/api/v1/masters/animal-species"):
        if row.get("name") == name:
            return int(row["id"])
    die(f"animal species not found: {name}")


def seed_owners_and_pets(client: ApiClient, state: State) -> list[dict]:
    result = []
    for owner in OWNERS:
        pets = owner["pets"]
        owner_payload = {key: value for key, value in owner.items() if key != "pets"}
        existing = client.find_by_name("/api/v1/owners", owner["owner_name"])
        if existing:
            owner_id = int(existing["id"])
            code, body = client.request("PATCH", f"/api/v1/owners/{owner_id}", owner_payload)
            if code not in (200, 204):
                die(f"PATCH owner {owner_id}: HTTP {code} {str(body)[:300]}")
            state.record_updated("owner", owner_id, owner["owner_name"])
            print(f"updated owner: {owner['owner_name']} (#{owner_id})")
        else:
            created = client.create("/api/v1/owners", owner_payload)
            owner_id = int(created["id"])
            state.record_created("owner", owner_id, owner["owner_name"])
            print(f"created owner: {owner['owner_name']} (#{owner_id})")
        pet_ids = []
        for pet in pets:
            pet_payload = {
                "owner_id": owner_id,
                "animal_species_id": find_species_id(client, pet["species"]),
                "name": pet["name"],
                "breed": pet["breed"],
                "gender": pet["gender"],
                "birth_date": pet["birth_date"],
                "weight": pet["weight"],
                "remarks": f"{RUN_TAG} synthetic pet",
            }
            existing_pet = client.find_by_name("/api/v1/pets", pet["name"])
            if existing_pet:
                pet_id = int(existing_pet["id"])
                pet_update = {key: value for key, value in pet_payload.items() if key not in {"owner_id", "animal_species_id"}}
                code, body = client.request("PATCH", f"/api/v1/pets/{pet_id}", pet_update)
                if code not in (200, 204):
                    die(f"PATCH pet {pet_id}: HTTP {code} {str(body)[:300]}")
                state.record_updated("pet", pet_id, pet["name"])
                print(f"updated pet: {pet['name']} (#{pet_id})")
            else:
                created_pet = client.create("/api/v1/pets", pet_payload)
                pet_id = int(created_pet["id"])
                state.record_created("pet", pet_id, pet["name"])
                print(f"created pet: {pet['name']} (#{pet_id})")
            pet_ids.append({"id": pet_id, "name": pet["name"]})
        result.append({"owner_id": owner_id, "owner_name": owner["owner_name"], "pets": pet_ids})
    return result


def pick_reservation_type_and_doctor(client: ApiClient) -> tuple[int, int]:
    reservation_type = next(
        (row for row in client.rows("/api/v1/masters/reservation-types") if row.get("is_active", True)),
        None,
    )
    doctor = next(
        (
            row for row in client.rows("/api/v1/masters/staffs")
            if row.get("staff_type") == "doctor" and row.get("is_active", True)
        ),
        None,
    )
    if reservation_type is None or doctor is None:
        die("no active reservation type or doctor found")
    return int(reservation_type["id"]), int(doctor["id"])


def ensure_doctor_capability(client: ApiClient, state: State, doctor_id: int, reservation_type_id: int) -> None:
    code, body = client.request("GET", f"/api/v1/masters/staffs/{doctor_id}/capable-reservation-types")
    if code != 200 or not isinstance(body, dict):
        die(f"GET capable reservation types for staff {doctor_id}: HTTP {code} {str(body)[:300]}")
    current = [int(value) for value in body.get("reservation_type_ids", [])]
    if reservation_type_id in current:
        return
    updated = sorted(set(current + [reservation_type_id]))
    code, body = client.request(
        "PUT",
        f"/api/v1/masters/staffs/{doctor_id}/capable-reservation-types",
        {"reservation_type_ids": updated},
    )
    if code not in (200, 204):
        die(f"PUT capable reservation types for staff {doctor_id}: HTTP {code} {str(body)[:300]}")
    state.created.append(
        {
            "kind": "staff_capability",
            "id": doctor_id,
            "name": f"{RUN_TAG} staff capability",
            "previous_ids": current,
        }
    )
    print(f"updated doctor capability: staff=#{doctor_id} + reservation_type=#{reservation_type_id}")


def ensure_shift(client: ApiClient, state: State, doctor_id: int, date: datetime) -> None:
    month = date.strftime("%Y-%m")
    code, body = client.request("GET", f"/api/v1/shifts?date={month}&staff_id={doctor_id}&limit=100&page=1")
    if code != 200:
        die(f"GET /api/v1/shifts: HTTP {code} {str(body)[:300]}")
    rows = body if isinstance(body, list) else body.get("data") or body.get("items") or []
    date_text = date.strftime("%Y-%m-%d")
    if any(row.get("date", "").startswith(date_text) for row in rows):
        return
    payload = {
        "staff_id": doctor_id,
        "date": date_text,
        "shift_type": "full",
        "start_time": "09:00",
        "end_time": "18:00",
        "note": f"{RUN_TAG} synthetic shift",
    }
    created = client.create("/api/v1/shifts", payload)
    state.record_created("shift", int(created["id"]), f"{RUN_TAG} {date_text}")
    print(f"created shift: {date_text} doctor=#{doctor_id}")


def find_reservation_type_id(client: ApiClient) -> int:
    for row in client.rows("/api/v1/masters/reservation-types"):
        if row.get("is_active", True):
            return int(row["id"])
    die("no active reservation type found")


def seed_reservations(client: ApiClient, state: State, owners: list[dict]) -> None:
    reservation_type_id, doctor_id = pick_reservation_type_and_doctor(client)
    ensure_doctor_capability(client, state, doctor_id, reservation_type_id)
    base = datetime.now(JST).replace(hour=10, minute=0, second=0, microsecond=0)
    for index, owner in enumerate(owners):
        pet = owner["pets"][0]
        start = base + timedelta(days=index)
        ensure_shift(client, state, doctor_id, start)
        payload = {
            "start_time": start.isoformat(),
            "end_time": (start + timedelta(minutes=30)).isoformat(),
            "owner_id": owner["owner_id"],
            "pet_id": pet["id"],
            "visit_type": "revisit",
            "reservation_type_id": reservation_type_id,
            "status": "confirmed",
            "source": "manual",
            "notes": f"{RUN_TAG} synthetic reservation",
        }
        payload["doctor_id"] = doctor_id
        label = f"{RUN_TAG} {index + 1}"
        if any(item["kind"] == "reservation" and item["name"] == label for item in state.created + state.updated):
            print(f"reservation exists in state: {label}")
            continue
        created = client.create("/api/v1/reservations", payload)
        reservation_id = int(created["id"])
        state.created.append(
            {
                "kind": "reservation",
                "id": reservation_id,
                "name": label,
                "pet_id": pet["id"],
                "owner_id": owner["owner_id"],
            }
        )
        print(f"created reservation: {label} (#{reservation_id})")


def find_record_for_reservation(client: ApiClient, pet_id: int, reservation_id: int) -> dict | None:
    for row in client.rows(f"/api/v1/medical-records?pet_id={pet_id}"):
        if row.get("appointment_id") == reservation_id:
            return row
    return None


def ensure_record_for_reservation(client: ApiClient, state: State, reservation: dict) -> dict | None:
    pet_id = int(reservation["pet_id"])
    reservation_id = int(reservation["id"])
    existing = find_record_for_reservation(client, pet_id, reservation_id)
    if existing is not None:
        return existing
    owner_id = reservation.get("owner_id")
    if owner_id is None:
        code, pet = client.request("GET", f"/api/v1/pets/{pet_id}")
        if code != 200 or not isinstance(pet, dict):
            print(f"warning: cannot load pet #{pet_id}: HTTP {code}")
            return None
        owner_id = pet.get("owner_id")
    _, doctor_id = pick_reservation_type_and_doctor(client)
    created = client.create(
        "/api/v1/medical-records",
        {
            "owner_id": int(owner_id),
            "pet_id": pet_id,
            "doctor_id": doctor_id,
            "appointment_id": reservation_id,
            "visit_date": datetime.now(JST).strftime("%Y-%m-%d"),
            "visit_type": "revisit",
            "status": "in_progress",
        },
    )
    print(f"created medical record: #{created['id']} reservation=#{reservation_id}")
    return created


def enrich_record(client: ApiClient, record: dict, ids: dict[str, dict[str, int]]) -> None:
    record_id = int(record["id"])
    code, plan = client.request("GET", f"/api/v1/medical-records/{record_id}/clinical-plan")
    version = plan.get("version") if code == 200 and isinstance(plan, dict) else None
    payload = {
        "physical_exam": "【デモ】一般状態良好。可視粘膜正常。",
        "diagnosis_type_id": ids["diagnosis_type"]["【デモ】内科疾患"],
        "diagnosis_name_id": ids["diagnosis_name"]["【デモ】胃炎"],
        "diagnosis_details": "【デモ】食欲不振のため来院。嘔吐なし。",
        "treatment_policy": "【デモ】内服加療と再診指示。",
    }
    if version is not None:
        payload["version"] = version
    code, body = client.request("PATCH", f"/api/v1/medical-records/{record_id}/clinical-plan", payload)
    if code not in (200, 204):
        die(f"PATCH clinical plan {record_id}: HTTP {code} {str(body)[:300]}")
    print(f"updated clinical plan: record #{record_id}")

    vital = {
        "recorded_at": record.get("date") or datetime.now(JST).isoformat(),
        "temperature": 38.2,
        "heart_rate": 110,
        "respiration_rate": 24,
        "weight": 10.4,
        "weight_unit": "Kg",
        "notes": f"{RUN_TAG} synthetic vital",
    }
    client.create(f"/api/v1/medical-records/{record_id}/vitals", vital)
    print(f"created vital: record #{record_id}")

    treatments = [
        {
            "item_type": "procedure",
            "procedure_id": ids["procedure"]["【デモ】一般診察"],
            "unit_price": 2000,
            "quantity": 1,
            "selected": True,
            "content": "【デモ】一般診察",
            "sort_order": 1,
        },
        {
            "item_type": "medicine",
            "medicine_id": ids["medicine"]["【デモ】アモキシシリン"],
            "unit_price": 800,
            "quantity": 1,
            "selected": True,
            "content": "【デモ】アモキシシリン",
            "sort_order": 2,
        },
    ]
    for treatment in treatments:
        client.create(f"/api/v1/medical-records/{record_id}/treatments", treatment)
    print(f"created treatments: record #{record_id}")


def create_estimate(client: ApiClient, state: State, record: dict) -> None:
    payload = {
        "medical_record_id": int(record["id"]),
        "owner_id": int(record["owner_id"]),
        "pet_id": int(record["pet_id"]),
        "title": "【デモ】見積書",
        "status": "draft",
        "comment": f"{RUN_TAG} synthetic estimate",
        "items": [
            {
                "name": "【デモ】一般診察",
                "category": "procedure",
                "unit_price": 2000,
                "quantity": 1,
                "tax_type": "excluded",
                "tax_rate": 0.1,
            }
        ],
    }
    created = client.create("/api/v1/estimates", payload)
    state.created.append({"kind": "estimate", "id": int(created["id"]), "name": "【デモ】見積書"})
    print(f"created estimate: #{created['id']}")


def billing_items() -> list[dict]:
    return [
        {
            "category": "procedure",
            "name": "【デモ】一般診察",
            "unit_price": 2000,
            "quantity": 1,
            "tax_type": "excluded",
            "tax_rate": 0.1,
            "is_insurance_applicable": False,
            "source": "manual",
        },
        {
            "category": "medicine",
            "name": "【デモ】アモキシシリン",
            "unit_price": 800,
            "quantity": 1,
            "tax_type": "excluded",
            "tax_rate": 0.1,
            "is_insurance_applicable": False,
            "source": "manual",
        },
    ]


def create_billing(client: ApiClient, state: State, record: dict, mode: str) -> None:
    scheduled_date = datetime.now(JST).replace(hour=0, minute=0, second=0, microsecond=0).isoformat()
    if mode == "paid":
        payload = {
            "medical_record_id": int(record["id"]),
            "scheduled_date": scheduled_date,
            "items": billing_items(),
            "payment_splits": [
                {"method": "cash", "amount": 3080, "received_amount": 3080, "change_amount": 0}
            ],
        }
        code, body = client.request(
            "POST",
            "/api/v1/accountings/complete",
            payload,
            headers={"Idempotency-Key": str(uuid.uuid4())},
        )
        if code not in (200, 201):
            die(f"POST /accountings/complete: HTTP {code} {str(body)[:300]}")
        data = body.get("data") if isinstance(body, dict) and isinstance(body.get("data"), dict) else body
        billing_id = int(data["id"])
        state.created.append({"kind": "billing", "id": billing_id, "name": f"{RUN_TAG} billing paid"})
        print(f"created billing: #{billing_id} mode=paid status={data.get('status')}")
        return

    header = {
        "medical_record_id": int(record["id"]),
        "subtotal": 2800,
        "tax_total": 280,
        "total_amount": 3080,
        "status": "waiting",
        "scheduled_date": scheduled_date,
        "memo": f"{RUN_TAG} synthetic {mode} billing",
    }
    created = client.create("/api/v1/accountings", header)
    billing_id = int(created["id"])
    for item in billing_items():
        item_payload = {key: value for key, value in item.items() if key not in {"source"}}
        item_payload["billing_id"] = billing_id
        client.create("/api/v1/billing-items", item_payload)
    state.created.append({"kind": "billing", "id": billing_id, "name": f"{RUN_TAG} billing {mode}"})
    print(f"created billing: #{billing_id} mode={mode} status={created.get('status')}")
    if mode == "partial":
        code, body = client.request(
            "PATCH",
            f"/api/v1/accountings/{billing_id}",
            {
                "payment_splits": [
                    {"method": "cash", "amount": 1000, "received_amount": 1000, "change_amount": 0}
                ]
            },
        )
        if code not in (200, 204):
            print(f"warning: partial payment not applied: HTTP {code} {str(body)[:200]}")
        else:
            print(f"applied partial payment: #{billing_id} amount=1000")


def create_clinical_extras(client: ApiClient, state: State, ids: dict[str, dict[str, int]], record: dict) -> None:
    today = datetime.now(JST).date()
    _, doctor_id = pick_reservation_type_and_doctor(client)
    pet_id = int(record["pet_id"])
    record_id = int(record["id"])

    vaccination = client.create(
        "/api/v1/vaccinations",
        {
            "medical_record_id": record_id,
            "pet_id": pet_id,
            "vaccine_id": ids["vaccine"]["【デモ】混合ワクチン（犬）"],
            "doctor_id": doctor_id,
            "date": today.isoformat(),
            "next_date": (today + timedelta(days=365)).isoformat(),
            "next_schedule_type": "1year",
            "supplemental": f"{RUN_TAG} synthetic vaccination",
        },
    )
    state.created.append({"kind": "vaccination", "id": int(vaccination["id"]), "name": "【デモ】ワクチン接種"})
    print(f"created vaccination: #{vaccination['id']}")

    checkup = client.create(
        f"/api/v1/medical-records/{record_id}/checkups",
        {
            "checkup_type_id": ids["checkup_type"]["【デモ】犬の健康診断"],
            "pet_id": pet_id,
            "doctor_id": doctor_id,
            "date": today.isoformat(),
            "result": "【デモ】異常なし",
        },
    )
    state.created.append(
        {
            "kind": "checkup",
            "id": int(checkup["id"]),
            "name": "【デモ】健康診断",
            "record_id": record_id,
        }
    )
    print(f"created checkup: #{checkup['id']}")

    exam_types = client.rows("/api/v1/masters/examination-types")
    if exam_types:
        exam = client.create(
            "/api/v1/examinations",
            {
                "medical_record_id": record_id,
                "pet_id": pet_id,
                "exam_type_id": int(exam_types[0]["id"]),
                "doctor_id": doctor_id,
                "date": datetime.now(JST).replace(hour=9, minute=0, second=0, microsecond=0).isoformat(),
                "status": "pending",
            },
        )
        state.created.append({"kind": "exam", "id": int(exam["id"]), "name": "【デモ】検査"})
        print(f"created exam: #{exam['id']}")
    else:
        print("warning: no examination type found; skipping exam")

    reservation_types = client.rows("/api/v1/masters/reservation-types")
    trimming_type = next(
        (row for row in reservation_types if row.get("category") == "trimming" and row.get("is_active", True)),
        next((row for row in reservation_types if row.get("is_active", True)), None),
    )
    if trimming_type is None:
        print("warning: no reservation type for trimming; skipping trimming")
    else:
        ensure_doctor_capability(client, state, doctor_id, int(trimming_type["id"]))
        trimming = client.create(
            "/api/v1/trimmings",
            {
                "date": today.isoformat(),
                "start_time": datetime.combine(today, datetime.min.time()).replace(hour=14, tzinfo=JST).isoformat(),
                "end_time": datetime.combine(today, datetime.min.time()).replace(hour=15, tzinfo=JST).isoformat(),
                "pet_id": pet_id,
                "staff_id": doctor_id,
                "course_id": ids["trimming_course"]["【デモ】シャンプーコース"],
                "reservation_type_id": int(trimming_type["id"]),
                "weight": "5.0",
                "status": "confirmed",
                "style_request": "【デモ】短め",
                "remarks": f"{RUN_TAG} synthetic trimming",
            },
        )
        state.created.append({"kind": "trimming", "id": int(trimming["id"]), "name": "【デモ】トリミング"})
        print(f"created trimming: #{trimming['id']}")

    hospitalization_payload = {
        "owner_id": int(record["owner_id"]),
        "pet_id": pet_id,
        "hospitalization_type": "hospitalization",
        "start_date": datetime.combine(today, datetime.min.time()).replace(hour=9, tzinfo=JST).isoformat(),
        "end_date": datetime.combine(today + timedelta(days=3), datetime.min.time()).replace(hour=9, tzinfo=JST).isoformat(),
        "status": "admitted",
        "doctor_id": doctor_id,
        "memo": f"{RUN_TAG} synthetic hospitalization",
        "owner_request": "【デモ】入院希望",
    }
    code, body = client.request("POST", "/api/v1/hospitalizations", hospitalization_payload)
    if code not in (200, 201):
        print(f"warning: hospitalization not created: HTTP {code} {str(body)[:200]}")
    else:
        data = body.get("data") if isinstance(body, dict) and isinstance(body.get("data"), dict) else body
        hospitalization_id = int(data["id"])
        state.created.append(
            {
                "kind": "hospitalization",
                "id": hospitalization_id,
                "name": "【デモ】入院",
                "retained": True,
            }
        )
        for item in state.created:
            if item["kind"] == "owner" and item["id"] == int(record["owner_id"]):
                item["retained"] = True
            if item["kind"] == "pet" and item["id"] == pet_id:
                item["retained"] = True
            if item["kind"] == "reservation" and item.get("pet_id") == pet_id:
                item["retained"] = True
        print(f"created hospitalization: #{hospitalization_id} (retained as history on cleanup)")


def seed_scenario(client: ApiClient, state: State, ids: dict[str, dict[str, int]]) -> None:
    reservations = [row for row in state.created if row["kind"] == "reservation"]
    if not reservations:
        print("no demo reservations; skipping scenario data")
        return
    records = []
    for reservation in reservations:
        record = ensure_record_for_reservation(client, state, reservation)
        if record is None:
            print(f"warning: medical record not found for reservation #{reservation['id']}")
            continue
        records.append(record)
    if not records:
        return
    enrich_record(client, records[0], ids)
    create_clinical_extras(client, state, ids, records[0])
    create_estimate(client, state, records[0])
    create_billing(client, state, records[0], mode="unpaid")
    if len(records) > 1:
        create_billing(client, state, records[1], mode="paid")
    if len(records) > 2:
        create_billing(client, state, records[2], mode="partial")


def cleanup(client: ApiClient, state: State) -> None:
    for checkup in [row for row in state.created if row["kind"] == "checkup"]:
        code, body = client.request(
            "DELETE",
            f"/api/v1/medical-records/{checkup['record_id']}/checkups/{checkup['id']}",
        )
        if code not in (200, 204, 404):
            print(f"FAIL cleanup: checkup #{checkup['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
        else:
            print(f"deleted checkup: #{checkup['id']}")
    for hospitalization in [row for row in state.created if row["kind"] == "hospitalization"]:
        code, body = client.request(
            "PATCH",
            f"/api/v1/hospitalizations/{hospitalization['id']}",
            {"status": "discharged"},
        )
        if code not in (200, 204):
            print(f"WARN cleanup: hospitalization #{hospitalization['id']} not discharged: HTTP {code} {str(body)[:120]}")
        else:
            print(f"discharged hospitalization: #{hospitalization['id']} (retained as history)")
    for billing in [row for row in state.created if row["kind"] == "billing"]:
        code, body = client.request("POST", f"/api/v1/accountings/{billing['id']}/cancel")
        if code not in (200, 204, 404, 409):
            print(f"FAIL cleanup: billing #{billing['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
        else:
            print(f"cancelled billing: #{billing['id']}")
    for estimate in [row for row in state.created if row["kind"] == "estimate"]:
        # The estimate delete endpoint refuses while items exist; clear them first.
        client.request("PATCH", f"/api/v1/estimates/{estimate['id']}", {"items": []})
        code, body = client.request("DELETE", f"/api/v1/estimates/{estimate['id']}")
        if code not in (200, 204, 404):
            print(f"FAIL cleanup: estimate #{estimate['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
        else:
            print(f"deleted estimate: #{estimate['id']}")
    # Creating a reservation also creates a linked medical record. Delete those
    # auto-created records before the reservation/pet, otherwise the API rejects
    # the reservation delete with 409.
    reservation_ids = {
        row["id"] for row in state.created if row["kind"] == "reservation"
    }
    pet_ids = {
        row["id"] for row in state.created
        if row["kind"] == "pet" and not row.get("retained")
    }
    for pet_id in sorted(pet_ids):
        code, body = client.request(
            "GET",
            f"/api/v1/medical-records?pet_id={pet_id}&limit=100&page=1",
        )
        if code != 200:
            print(f"FAIL cleanup: list records for pet #{pet_id}: HTTP {code}", file=sys.stderr)
            continue
        rows = body.get("data") if isinstance(body, dict) else body
        for record in rows or []:
            if record.get("appointment_id") not in reservation_ids:
                continue
            record_id = record["id"]
            client.request("DELETE", f"/api/v1/medical-records/{record_id}/clinical-plan")
            code, treatments = client.request("GET", f"/api/v1/medical-records/{record_id}/treatments")
            treatment_rows = treatments.get("data") if isinstance(treatments, dict) else treatments
            for treatment in treatment_rows or []:
                code, body = client.request("DELETE", f"/api/v1/medical-records/{record_id}/treatments/{treatment['id']}")
                if code not in (200, 204, 404):
                    print(f"FAIL cleanup: treatment #{treatment['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
            code, vitals = client.request("GET", f"/api/v1/medical-records/{record_id}/vitals")
            vital_rows = vitals.get("data") if isinstance(vitals, dict) else vitals
            for vital in vital_rows or []:
                code, body = client.request("DELETE", f"/api/v1/medical-records/{record_id}/vitals/{vital['id']}")
                if code not in (200, 204, 404):
                    print(f"FAIL cleanup: vital #{vital['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
            code, body = client.request("DELETE", f"/api/v1/medical-records/{record_id}")
            if code not in (200, 204, 404):
                print(f"FAIL cleanup: medical record #{record_id}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
            else:
                print(f"deleted medical record: #{record_id} (pet #{pet_id})")
    for item in [row for row in state.created if row["kind"] == "staff_capability"]:
        code, body = client.request(
            "PUT",
            f"/api/v1/masters/staffs/{item['id']}/capable-reservation-types",
            {"reservation_type_ids": item.get("previous_ids", [])},
        )
        if code not in (200, 204):
            print(f"FAIL cleanup: staff capability #{item['id']}: HTTP {code} {str(body)[:120]}", file=sys.stderr)
        else:
            print(f"restored staff capability: #{item['id']}")
    order = ["reservation", "shift", "vaccination", "exam", "trimming", "pet", "owner", "diagnosis_name", "chief_complaint", "diagnosis_type", "procedure", "medicine", "vaccine", "insurance", "cage", "hospitalization_plan", "inventory", "inquiry_template", "checkup_type", "trimming_course", "reservation_type_group"]
    resources = [
        row for row in state.created
        if row["kind"] not in {"staff_capability", "billing", "estimate", "checkup", "hospitalization"}
        and not row.get("retained")
    ]
    failures = []
    for kind in order:
        for item in [row for row in resources if row["kind"] == kind]:
            path = {
                "reservation": "/api/v1/reservations",
                "shift": "/api/v1/shifts",
                "vaccination": "/api/v1/vaccinations",
                "exam": "/api/v1/examinations",
                "trimming": "/api/v1/trimmings",
                "pet": "/api/v1/pets",
                "owner": "/api/v1/owners",
                "diagnosis_name": "/api/v1/masters/diagnosis-names",
                "chief_complaint": "/api/v1/masters/chief-complaint-types",
                "diagnosis_type": "/api/v1/masters/diagnosis-types",
                "procedure": "/api/v1/masters/procedures",
                "medicine": "/api/v1/masters/medicines",
                "vaccine": "/api/v1/masters/vaccines",
                "insurance": "/api/v1/masters/insurances",
                "cage": "/api/v1/masters/cages",
                "hospitalization_plan": "/api/v1/masters/hospitalization-plans",
                "inventory": "/api/v1/inventory",
                "inquiry_template": "/api/v1/masters/inquiry-templates",
                "checkup_type": "/api/v1/masters/checkup-types",
                "trimming_course": "/api/v1/masters/trimming-courses",
                "reservation_type_group": "/api/v1/masters/reservation-type-groups",
            }[kind]
            code, body = client.request("DELETE", f"{path}/{item['id']}")
            if code not in (200, 204, 404):
                if code == 409 and kind in {"procedure", "medicine", "vaccine", "insurance", "cage", "hospitalization_plan", "inventory", "inquiry_template", "checkup_type", "trimming_course", "reservation_type_group", "diagnosis_name", "diagnosis_type", "chief_complaint"}:
                    print(f"WARN cleanup: {kind} #{item['id']} still referenced; kept ({item['name']})")
                else:
                    failures.append(f"{kind} #{item['id']}: HTTP {code} {str(body)[:120]}")
            else:
                print(f"deleted {kind}: {item['name']} (#{item['id']})")
    if failures:
        for failure in failures:
            print(f"FAIL cleanup: {failure}", file=sys.stderr)
        raise SystemExit(1)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default=os.environ.get("DEMO_SEED_BASE_URL", "http://localhost:8080"))
    parser.add_argument("--clinic-id", type=int, default=1)
    parser.add_argument("--email", default=os.environ.get("DEMO_SEED_EMAIL", ""))
    parser.add_argument("--password-env", default="DEMO_SEED_PASSWORD")
    parser.add_argument("--apply", action="store_true", help="make API writes; default is dry-run")
    parser.add_argument("--cleanup", action="store_true", help="delete resources recorded in the state file")
    args = parser.parse_args()

    state_dir = DEFAULT_STATE_ROOT / RUN_TAG
    state_path = state_dir / "state.json"

    if not args.apply:
        print(f"DRY RUN: {RUN_TAG} clinic={args.clinic_id} base={args.base_url}")
        for line in build_plan():
            print(line)
        print("Re-run with --apply to write.")
        return

    password = os.environ.get(args.password_env, "")
    if not args.email or not password:
        die(f"set --email/--password-env ({args.password_env})")

    client = ApiClient(args.base_url, args.clinic_id)
    client.login(args.email, password)
    state = State(run_tag=RUN_TAG, clinic_id=args.clinic_id)
    if state_path.exists():
        loaded = json.loads(state_path.read_text())
        state.created = loaded.get("created", [])
        state.updated = loaded.get("updated", [])

    if args.cleanup:
        cleanup(client, state)
        state.created = []
        state.updated = []
        state.save(state_path)
        return

    try:
        ids = seed_masters(client, state)
        owners = seed_owners_and_pets(client, state)
        seed_reservations(client, state, owners)
        seed_scenario(client, state, ids)
    finally:
        state.save(state_path)
    print(f"state: {state_path}")
    print("W3b-3 applied; the hospitalization scenario is retained on cleanup. Inventory movements are not seeded (no movement API).")


if __name__ == "__main__":
    main()
