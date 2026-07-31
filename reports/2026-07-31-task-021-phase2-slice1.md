# TASK-021 Phase2 slice 1 — positive capability contract / known-client proof

| 項目 | 値 |
|---|---|
| Unit / packet | `TODO-MD-POST-FINAL-FOLLOWUP-ORCH-20260731` / `W-021-P2` |
| Branch / worktree | `agent/w-021-p2` / `AnimalEkarte-w021p2` |
| Baseline | `e1b51ceae0e8370146cdb13365a8324985728977` |
| Claim | `claim/TASK-021-P2`（本 packet では削除しない） |
| RED checkpoint | `5d0a53b46` |
| Disposition | **SLICE1 COMPLETE / CLEAN-GO・DROP・migrate HOLD** |
| External legacy endpoint use | **UNREPORTED** |

## 0. 結論

TASK-021 inventory §6.1–§6.3 のうち、破壊変更を伴わない最初の consumer/contract slice を実装した。

1. `ReservationStaff.capable_courses` と `CapableCourse` を OpenAPI の正式な positive contract にした。
2. `ExcludedCourse`、`excluded_courses`、Create/Update の `excluded_type_ids`、legacy exclusion endpoint は削除せず `deprecated: true` とした。
3. 既知の院内予約 FE consumer は wire object を spread せず positive field だけを投影する。legacy `excluded_courses` は query cache / downstream consumer へ伝播しない。
4. `capable_courses` 欠落・`null` はそれぞれ独立した fixture で `[]` への正規化を固定し、対応可能扱いへ fail-open しない。
5. BE response は capability 0 件でも `"capable_courses":[]` を返すことをテストで固定した。

本 slice は legacy route/field/model/table/seed を削除していない。外部 client の利用ゼロは access log・client registry・利用者確認を実施していないため **UNREPORTED** であり、CLEAN-GO ではない。

## 1. 変更した contract

### OpenAPI

- `CapableCourse` を追加し、`id` / `name` を required とした。
- `ReservationStaff.required` に `capable_courses` を追加した。
- `ReservationStaff.capable_courses` を `CapableCourse[]` として記載した。
- 次の互換面は残存させ、deprecated とした。
  - `StaffReservationExclusion`
  - `ExcludedCourse`
  - `ReservationStaff.excluded_courses`
  - `CreateReservationStaffRequest.excluded_type_ids`
  - `UpdateReservationStaffRequest.excluded_type_ids`
  - `GET|PUT /masters/staffs/{id}/excluded-reservation-types`
- positive `GET|PUT /masters/staffs/{id}/capable-reservation-types` は維持した。

### FE consumer boundary

`useGetReservationStaffs` は wire response から次だけを新しい object へ投影する。

- `id`
- `name`
- `is_active`
- `capable_courses`（欠落・`null` は `[]`）

これにより、backend が互換期間中に `excluded_courses` や他の wire field を返しても、既知の予約フォーム consumer へは伝播しない。

## 2. Known-client proof

### Production FE / LIFF / line-reserve の exclusion consumer

```text
$ rg -n 'excluded_type_ids|excluded_courses|excluded-reservation-types' frontend/src frontend/liff frontend/line-reserve --glob '!**/*.test.*' --glob '!**/types/generated/**'
exit 1
<no matches>
```

テスト内の `excluded_courses` fixture は「wire に legacy field があっても投影後に消える」ことを固定する negative fixture であり、production consumer ではない。

### Production positive capability consumer

```text
$ rg -n 'capable_courses|capable-reservation-types' frontend/src frontend/liff frontend/line-reserve --glob '!**/*.test.*' --glob '!**/types/generated/**'
exit 0
matches=10
frontend/src/lib/query-keys.ts:92:      resource: "capable-reservation-types" | "clinics" | "permission-groups",
frontend/src/hooks/use-reservation-types.ts:55:  capable_courses: ReservationStaffCourse[];
frontend/src/hooks/use-reservation-types.ts:62:  capable_courses?: ReservationStaffCourse[] | null;
frontend/src/hooks/use-reservation-types.ts:93:    capable_courses: staff.capable_courses ?? [],
frontend/src/hooks/use-reservation-types.ts:157: * capable_courses は院内予約フォームの担当者候補を肯定形で絞り込む（TASK-021 Stage B）。
frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts:14:  capable_courses?: ReadonlyArray<{ id: number; name?: string }>;
frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts:40:    const capable = reservationStaff.capable_courses ?? [];
frontend/src/features/master/api/staff-reservation-types.ts:12:  queryKeys.staffs.subResource(staffId, "capable-reservation-types");
frontend/src/features/master/api/staff-reservation-types.ts:19:        `/v1/masters/staffs/${staffId}/capable-reservation-types`,
frontend/src/features/master/api/staff-reservation-types.ts:39:      await axios.put(`/v1/masters/staffs/${staffId}/capable-reservation-types`, {
```

### External use boundary

- repository 内の known FE source は exclusion request/response を消費していない。
- repository 外の client、運用 script、API access log、第三者 integration は未調査。
- よって legacy endpoint の外部利用ゼロは **UNREPORTED**。deprecation 終了や route 削除を本報告から推論してはならない。

## 3. TDD evidence

### RED

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/backend:/app backend test ./internal/apicontract ./internal/reservation -run 'Test(OpenAPIReservationStaff|ToReservationStaffResponse)' -count=1
exit 1
--- FAIL: TestOpenAPIReservationStaffUsesPositiveCapabilityContract
Messages: CapableCourse schema must be documented
FAIL github.com/animal-ekarte/backend/internal/apicontract
ok   github.com/animal-ekarte/backend/internal/reservation
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint npx -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend vitest run src/hooks/use-reservation-types.test.ts
exit 1
Test Files  1 failed (1)
Tests       2 failed (2)
Received included excluded_courses; the first case also propagated staff_type.
```

### GREEN / scoped verification

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/backend:/app backend test ./internal/apicontract ./internal/reservation -run 'Test(OpenAPIReservationStaff|ToReservationStaffResponse)' -count=1
exit 0
ok   github.com/animal-ekarte/backend/internal/apicontract 0.037s
ok   github.com/animal-ekarte/backend/internal/reservation 0.003s
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/backend:/app backend test -p 1 ./internal/apicontract ./internal/reservation -count=1
exit 0
ok   github.com/animal-ekarte/backend/internal/apicontract 0.751s
ok   github.com/animal-ekarte/backend/internal/reservation 20.900s
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint npx -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend vitest run src/hooks/use-reservation-types.test.ts
exit 0
Test Files  1 passed (1)
Tests       2 passed (2)
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint npx -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend eslint src/hooks/use-reservation-types.ts src/hooks/use-reservation-types.test.ts
exit 0
<no lint findings>
```

### Wave 2 evidence repair（RED failure ではない）

初回 GREEN は欠落 fixture だけで `null` も実装式から推論していた。report fidelity のため `capable_courses: null` と legacy `excluded_courses` を同時に返す独立 fixture を追加し、`[]` への正規化と legacy field 非伝播を直接固定した。既存実装の証拠補強であり、新しい RED failure ではない。

```text
$ shasum -a 256 frontend/src/hooks/use-reservation-types.test.ts
2a2df7097ae9864727e08baf4208852d842c8f840855f1084fafd810fe0a02ed  frontend/src/hooks/use-reservation-types.test.ts

$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint sh -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend -c 'sha256sum src/hooks/use-reservation-types.test.ts'
2a2df7097ae9864727e08baf4208852d842c8f840855f1084fafd810fe0a02ed  src/hooks/use-reservation-types.test.ts
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint npx -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend vitest run src/hooks/use-reservation-types.test.ts
exit 0
Test Files  1 passed (1)
Tests       3 passed (3)
```

```text
$ docker compose --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint npx -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p2/frontend:/app frontend eslint src/hooks/use-reservation-types.ts src/hooks/use-reservation-types.test.ts
exit 0
<no lint findings>
```

## 4. Failure Signature log

| Attempt | Expected | Actual / signature | New hypothesis / action | Result |
|---|---|---|---|---|
| pre-RED harness | Docker formatting only | backend default entrypoint ran startup migration checker; `applied=0 skipped=3`, then existing `003_demo` checksum mismatch aborted before gofmt | Default service entrypoint is unsafe for one-off worktree commands。以後すべて `--entrypoint gofmt|go|npx|sha256sum` を明示 | No migration/seed applied; unsafe command form not reused |
| RED parser repair 1 | Contract assertion failure | initial path decoder tried to decode path-level sequence keys into operation struct | Decode only `get` / `put` fields on a typed path item | Valid RED reached missing `CapableCourse` |
| GREEN repair 1 | Green contract test | one Docker bind-mount run reported YAML parse error in a later unchanged line; a subsequent diagnostic separately observed a transient truncated bind-mounted test file | Compare host/container SHA-256 before rerun; use targeted node decode rather than whole-spec typed projection | SHA matched; targeted and full package tests PASS |

## 5. Explicit non-actions and safety gates

```text
$ rg -n 'DROP TABLE( IF EXISTS)? staff_reservation_exclusions|DROP TABLE staff_reservation_exclusions' backend/migrations
exit 1
<no matches>

$ rg -n 'available-staffs|available_staffs' frontend/src backend/internal --glob '!**/*_test.go' --glob '!**/*.test.*'
exit 1
<no matches>
```

- migration author: none
- migration apply: none (`applied=0`; startup checker near miss is recorded above)
- seed / RLS / table / route deletion: none
- `make migrate`: not run
- generated type regeneration: not run; `frontend/src/types/generated/models.ts` の legacy `StaffReservationExclusion` は残存
- `available-staffs`: not added
- CLEAN-GO / DROP approval: not claimed
- external endpoint use zero: not claimed（**UNREPORTED**）

## 6. 次の Phase2 slice へ残すもの

1. legacy exclusion endpoint の外部利用を telemetry / client registry / named owner で実証する。
2. deprecation 期間と終了条件を明示した後、request `excluded_type_ids` と BE facade/ports の consumer を順序撤去する。
3. generated types、filter fixture、query-key residual、docs を capable-only へ揃える。
4. それらが完了しても §6.4–§6.6 と別の破壊変更承認までは CLEAN-GO / table DROP / seed delete / migrate を HOLD する。
