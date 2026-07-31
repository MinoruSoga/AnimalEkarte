# TASK-021 Phase2 slice 2 — reject deprecated `excluded_type_ids` on staff write

| 項目 | 値 |
|---|---|
| Unit / packet | `TODO-MD-R05-PHASEB-021-SLICE2-ORCH-20260731` / `W-021-S2` |
| Branch / worktree | `agent/w-021-s2` / `AnimalEkarte-w021s2` |
| Baseline | `e17e749ecf83b7292da9a6cd3a783ab633fc0083` |
| Claim | `claim/TODO-MD-R05-PHASEB-021-SLICE2-ORCH-20260731` / `claim/TASK-021-S2`（本 packet では削除しない） |
| Disposition | **SLICE2 COMPLETE / CLEAN-GO・DROP・migrate HOLD** |
| External legacy endpoint use (access logs) | **UNREPORTED (USER)** |

## 0. 結論

TASK-021 Phase2 slice2 として、reservation-staffs **Create/Update write path** から inverse 置換を受け付ける経路を閉じた。

1. **Create**: `len(excluded_type_ids) > 0` → `400 INVALID_INPUT`。空配列・省略は従来どおり inverse facade で full active universe を seed する。
2. **Update**: `excluded_type_ids` が **存在する**（空配列ポインタ含む）→ `400 INVALID_INPUT`。省略時は staff 本体フィールドのみ更新し、capability は触らない。
3. **response** `excluded_courses` は deprecation 期間中維持（capabilities からの派生 facade）。
4. **master routes** `GET|PUT /masters/staffs/:id/excluded-reservation-types` は **削除しない**（本 slice の対象外・live 維持）。
5. **OpenAPI**: property は削除せず、reject 条件を description に明記。
6. 院内 FE/LIFF/line-reserve の production consumer は **ZERO_IN_REPO**（slice1 投影 + 本 inventory 再確認）。
7. 外部 access log / client registry による実利用ゼロ証明は未実施のため **UNREPORTED (USER)**。CLEAN-GO ではない。

本 slice は table/route の mass delete・migrate・available-staffs 変更を行っていない。

## 1. External / legacy use inventory（in-repo）

### 1.1 Production FE / LIFF / line-reserve

```text
$ rg -n 'excluded_type_ids|excluded_courses|excluded-reservation-types' \
    frontend/src frontend/liff frontend/line-reserve \
    --glob '!**/*.test.*' --glob '!**/types/generated/**'
exit 1
<no matches>
```

**Disposition: ZERO_IN_REPO** for production FE/LIFF/line-reserve consumers.

- 院内候補フィルタは `capable_courses` のみ（Phase1 / slice1 済み）。
- master staff 画面も `capable-reservation-types` を使用（`frontend/src/features/master/api/staff-reservation-types.ts`）。
- FE tests 内の `excluded_courses` は「wire に legacy field があっても投影後に消える」negative fixture のみ。

### 1.2 OpenAPI client / generated types

```text
$ rg -n 'excluded_type_ids|excluded_courses|excluded-reservation-types' frontend/src/types
exit 1
<no matches>
```

generated OpenAPI client への in-repo 参照は検出されず。

### 1.3 Remaining BE / OpenAPI surfaces（live / deprecation）

| Surface | Path / symbol | Disposition |
|---|---|---|
| OpenAPI response field | `ReservationStaff.excluded_courses` | **KEEP** deprecation window（派生 facade） |
| OpenAPI request field | `CreateReservationStaffRequest.excluded_type_ids` | **KEEP property** / **reject non-empty**（本 slice） |
| OpenAPI request field | `UpdateReservationStaffRequest.excluded_type_ids` | **KEEP property** / **reject if present**（本 slice） |
| OpenAPI path | `GET\|PUT /masters/staffs/{id}/excluded-reservation-types` | **KEEP live**（削除禁止） |
| Route | `backend/internal/staff/handler.go` L209–210 | **KEEP live** |
| Handler | `staff_handler.go` Get/SetStaffExcludedReservationTypes | **KEEP live** |
| Request DTO | `reservation_staff_request.go` `excluded_type_ids` | **KEEP bind** → service reject |
| Response DTO | `reservation_staff_response.go` `excluded_courses` | **KEEP emit** |
| Service write | `reservationStaffService.Create/Update` | **reject**（本 slice） |
| Repo inverse facade | `UpdateExcludedReservationTypes` | **KEEP**（Create empty seed + master exclusion PUT 用） |

### 1.4 External access logs

本番/ステージングの access log・client registry・利用者確認は本 agent では実施していない。

**Disposition: UNREPORTED (USER)** — CLEAN-GO / route DROP の前提に使わない。

## 2. 実装内容

### 2.1 Service reject

`backend/internal/reservation/reservation_staff_service.go`

- `errDeprecatedExcludedTypeIDs` 定数を追加。
- **Create**: `len(input.ExcludedTypeIDs) > 0` なら `apperrors.WrapInvalidInput` を返し、tx / create / inverse write に入らない。
- **Create empty**: 従来どおり `UpdateExcludedReservationTypes(..., [])` で full universe seed。
- **Update**: `input.ExcludedTypeIDs != nil` なら同様に reject。nil 省略時のみ staff fields + exclusion readback。

### 2.2 OpenAPI honesty（非破壊）

`backend/docs/api.yaml`

- Create `excluded_type_ids`: 非空送信 → 400。空/省略は seed 互換と明記。
- Update `excluded_type_ids`: フィールド存在（空含む）→ 400。property は削除しないと明記。
- property / path 自体は削除していない。

### 2.3 Tests

| Test | 期待 |
|---|---|
| `Create_RejectsNonEmptyExcludedTypeIDs` | InvalidInput / create 未実行 |
| `Create_EmptyExcludedTypeIDsStillSeedsUniverse` | success / replace `[]` |
| `Create_TxUpdateExcludedError` | empty seed path の repo error を維持 |
| `Update_RejectsExcludedTypeIDsPresent` | non-empty / empty pointer とも InvalidInput |
| `Update_OmitsExcludedTypeIDsDoesNotReplace` | nil 時は replace 非実行 |
| apicontract deprecation | property 残存 + description に `400` と `capable-reservation-types` |

## 3. 検証証拠

worktree の backend を mount し、main compose の `ekarte-network` と DB env を用いて実行（本 worktree に `.env.local` / 稼働 compose は無し）。

```text
$ docker run --rm --entrypoint go --network ekarte-network \
    -v "$PWD/backend:/app" -w /app \
    -e DB_HOST=db -e DB_PORT=5432 -e DB_USER -e DB_PASSWORD -e DB_NAME \
    animalekarte-backend \
    test ./internal/reservation/ -count=1 -run 'ReservationStaff|Excluded'
ok  github.com/animal-ekarte/backend/internal/reservation  5.374s

$ docker run --rm --entrypoint go \
    -v "$PWD/backend:/app" -w /app animalekarte-backend \
    test ./internal/apicontract/ -count=1 -run 'ReservationStaff|Capability|excluded'
ok  github.com/animal-ekarte/backend/internal/apicontract  0.064s
```

## 4. 変更ファイル（allowlist ⊆）

1. `backend/internal/reservation/reservation_staff_service.go`
2. `backend/internal/reservation/reservation_staff_service_test.go`
3. `backend/docs/api.yaml`
4. `backend/internal/apicontract/openapi_reservation_staff_capability_contract_test.go`
5. `reports/2026-07-31-task-021-phase2-slice2.md`

## 5. 意図的に触っていないもの

- `GET|PUT .../excluded-reservation-types` route / handler / table
- available-staffs
- migrate / DROP
- response `excluded_courses` 除去
- FE（projection-only 済みのため residual なし）
- R-05 / line_reservation_setting / lstep / todo.md

## 6. 次 slice への引き継ぎ

1. USER が access log / 外部 client 利用を確認（UNREPORTED 解消）するまで CLEAN-GO 不可。
2. 解消後: request DTO / OpenAPI property 削除、response `excluded_courses` 削除、exclusion master route の deprecation 終了判断。
3. master exclusion PUT がまだ capability inverse write の live 入口である点は、route DROP 時に capable-only へ一本化する必要がある。

## 7. Claim

- 保持: `claim/TODO-MD-R05-PHASEB-021-SLICE2-ORCH-20260731`, `claim/TASK-021-S2`
- エージェントは claim を削除しない。統合後に USER が `git branch -D` で解放する。
