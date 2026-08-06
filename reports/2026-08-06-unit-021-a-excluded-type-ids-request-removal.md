# UNIT-021-A — request `excluded_type_ids` 削除

| 項目 | 値 |
|------|-----|
| Date | 2026-08-06 |
| Authority | PO proxy Decision Pack D-021-A (`reports/2026-08-06-po-proxy-decision-pack.md`) |
| Scope | request DTO + OpenAPI property 削除のみ |

## Changes

- `CreateReservationStaffRequest` / `UpdateReservationStaffRequest`: `excluded_type_ids` 削除（`api.yaml`）
- request/service input structs: 同フィールド削除
- Create: 常に `UpdateExcludedReservationTypes(..., []uint64{})` で full universe seed
- Update: staff fields only（capability 変更は capable-reservation-types）
- OpenAPI contract test: request に property が **無い**ことを要求
- response `excluded_courses` · master exclusion route · migrate: **未変更（HOLD）**

## Tests (scoped, no integration DB)

```text
go test ./internal/reservation/ -run 'TestReservationStaffService_Create_Success|...AlwaysSeeds|...Update_DoesNotReplace|TestCreateReservationStaffRequest|TestUpdateReservationStaffRequest|TestBuildReservationStaffUpdate'
go test ./internal/apicontract/ -run 'ReservationStaff'
go test ./internal/lintscan/ -run 'MasterFKWriteInventory'
```

All green (host). Integration/repo tests require compose `db` hostname.

## Non-goals

- response / route DROP / DB DROP
