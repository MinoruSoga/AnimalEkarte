# TASK-021 Stage A Inventory — exclusion 面 consumer 残存（NO DELETE）

| 項目 | 値 |
|------|-----|
| Packet / Unit | `TODO-MD-OPEN-REMAINING-ORCH-WAVE-20260731-V2` / W-021A |
| Task | TASK-021 Stage A（exclusion 破壊的撤去の前段 inventory） |
| Worktree | `AnimalEkarte-w021a` / branch `agent/w-021a` |
| Claim | `claim/TASK-021`（coordinator 取得済・本 unit は削除しない） |
| Wave 0 RO-021 join | **Stage A = INVENTORY-ONLY** |
| 実施日 | 2026-07-31 |
| Stage B tip（todo.md） | `e9dddd921` |
| 決裁 | `reports/2026-07-31-todo-po-decisions-FINAL.md` § TASK-021 |

---

## 0. 判定（結論）

| 項目 | 判定 |
|------|------|
| **Stage A disposition** | **INVENTORY-ONLY（consumers remain）— NO CLEAN-GO** |
| 今すぐ code / route / model / table / seed / OpenAPI を削除してよいか | **NO** |
| drop migration を author してよいか | **NO**（本 packet 禁止） |
| Dual-write to physical `staff_reservation_exclusions`（production） | **NO**（facade only; write SoT = capabilities） |
| 今削除した場合の external contract 破壊 | **YES**（§4） |

**本 packet の唯一の成果物は本 inventory 報告である。product code は未変更。**

---

## 1. Stage B SoT 要約（capabilities only; facade still live）

Stage B（実装済・facade 存続）の契約:

1. **唯一の write / decision SoT** = `staff_reservation_capabilities`（肯定形）。
2. **exclusion 形状は期限付き互換 facade**:
   - **Read**: `excluded = clinic-scoped active universe \ capable`（`staff_affinity.go` の inverse mapping）。
   - **Write**: `excluded_type_ids` / exclusion PUT は **atomic capability replacement** に変換（`capable = universe \ excluded`）。
3. **物理テーブル `staff_reservation_exclusions` への production write はゼロ**。
   - 実装: `UpdateExcludedReservationTypes` → `replaceReservationCapabilitiesTx` のみ（`reservation_staff_repository.go:368-401`）。
   - 回帰: `staff_affinity_facade_test.go`, `reservation_staff_repository_tx_atomicity_test.go`, `reservation_staff_exclusion_clinic_isolation_test.go`, `reservation_staff_full_replacement_concurrency_test.go`。
4. **Legacy exclusion 行は SoT として無視**される（capabilities が勝つ; facade test で固定）。
5. **院内候補フィルタ**は肯定形 `capable_courses` + fail-closed（`filter-staff-candidates.ts`）。`excluded_courses` は型・レスポンスに残るが候補判定には使わない。
6. **`GET /v1/reservations/available-staffs` は作らない**（WONTFILE; `available_staffs_ban_test.go`）。
7. **維持する既存 surface**（Stage A でも壊さない）: `reservation-staffs`, `on-duty-staffs`, `available-times`。

Stage B は「二重 junction を捨てた」が「exclusion 契約を捨てた」わけではない。**facade consumer が残る限り Stage A 破壊削除は不可**。

---

## 2. Full consumer inventory（本 worktree で rg 再検証）

凡例: **Live** = production path が依存 / **Contract** = API・型・seed 契約 / **Test-only** = テスト・lint の参照 / **Legacy physical** = テーブル・seed が残存するが production write なし

| # | Surface | Path(s) | 状態 | 削除時影響 |
|---|---------|---------|------|------------|
| 1 | **GORM model** | `backend/internal/model/staff_reservation_exclusion.go` | Live model + TableName `staff_reservation_exclusions` | schema_drift / tygo / facade 戻り値型が壊れる |
| 2 | **schema_drift 登録** | `backend/internal/model/schema_drift_test.go` (`&model.StaffReservationExclusion{}`) | Live gate | drift test fail |
| 3 | **RLS inventory** | `backend/internal/model/rls_migration_test.go` (`"staff_reservation_exclusions"`) | Live gate | RLS migration test fail |
| 4 | **Migration 001 DDL + RLS** | `backend/migrations/001_init.sql` L2655–2663 (CREATE), L3150–3153 (RLS policy name `tenant_staff_reservation_exclusions_isolation`) | Legacy physical | drop なしに model/seed 削除不可; drop は **別 migration + 明示承認** が必須 |
| 5 | **Seed CSV** | `backend/migrations/seeds/003_demo/staff_reservation_exclusions.csv`（データ行あり） | Legacy physical / Contract | seed apply が table を期待 |
| 6 | **Seed manifest** | `backend/migrations/seeds/003_demo/manifest.json` L321–322 | Contract | manifest と CSV の対が残る |
| 7 | **seed-export allowlist** | `backend/cmd/seed-export/tables.go` L116 `"staff_reservation_exclusions"` | Contract | export 対象から外す前に table 生存確認が必要 |
| 8 | **Reservation facade — inverse helpers** | `backend/internal/reservation/staff_affinity.go` | Live | Stage B の中核; Stage A では capability-only API に置き換え後に削除可 |
| 9 | **Reservation facade — repository** | `backend/internal/reservation/reservation_staff_repository.go`: `FindAllExcluded*`, `UpdateExcludedReservationTypes`, `deriveExcludedFromUniverse` | Live facade | GET/PUT exclusion shape と response が依存 |
| 10 | **Reservation facade — service** | `backend/internal/reservation/reservation_staff_service.go`: `ExcludedTypeIDs` input, `ListExcludedByStaffIDs`, Create/Update/PatchStatus の exclusion readback | Live facade | reservation-staffs CRUD 契約 |
| 11 | **Reservation facade — request DTO** | `backend/internal/reservation/reservation_staff_request.go` `excluded_type_ids` | Live contract | Create/Update body 破壊 |
| 12 | **Reservation facade — response DTO** | `backend/internal/reservation/reservation_staff_response.go` `excluded_courses` (+ Stage B `capable_courses`) | Live contract | List/Create/Update response 破壊 |
| 13 | **Reservation routes** | `backend/internal/reservation/routes.go` `/clinics/:clinic_id/reservation-staffs`（CRUD 一式） | Live | ルート自体は維持対象だが payload の exclusion フィールドは facade |
| 14 | **Staff routes** | `backend/internal/staff/handler.go` L209–210: `GET|PUT /masters/staffs/:id/excluded-reservation-types` | Live contract | OpenAPI・外部クライアント破壊 |
| 15 | **Staff handlers** | `backend/internal/staff/staff_handler.go` `GetStaffExcludedReservationTypes` / `SetStaffExcludedReservationTypes` | Live | 同上 |
| 16 | **Staff service** | `backend/internal/staff/staff_service_permissions.go` `GetExcludedReservationTypeIDs` / `SetExcludedReservationTypeIDs` | Live facade → reservation repo | 同上 |
| 17 | **Staff ports** | `backend/internal/staff/ports.go` `FindAllExcluded*` / `UpdateExcluded*` | Live interface | staff ↔ reservation 境界 |
| 18 | **Staff service interface** | `backend/internal/staff/staff_service.go` Get/Set Excluded methods | Live | 同上 |
| 19 | **OpenAPI — schema** | `backend/docs/api.yaml`: `StaffReservationExclusion` (L6544+), `ReservationStaff.excluded_courses` (L8724), Create/Update `excluded_type_ids` (L8749/L8773) | Contract | codegen / 外部契約破壊。なお `capable_courses` は **api.yaml ReservationStaff に未記載**（実装 response には存在する Stage B 追加） |
| 20 | **OpenAPI — paths** | `backend/docs/api.yaml` `/masters/staffs/{id}/excluded-reservation-types` (L13872+) | Contract | path 削除は破壊 |
| 21 | **FE generated types** | `frontend/src/types/generated/models.ts` `StaffReservationExclusion` (L2975+) | Contract (tygo) | model 削除で再生成差分 |
| 22 | **FE reservation-staffs client** | `frontend/src/hooks/use-reservation-types.ts` — type `excluded_courses` + normalize map | Live type residual | フィールド削除で型・正規化が要更新（候補フィルタ自体は capable 依存） |
| 23 | **FE candidate filter** | `frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts` — `excluded_courses?` deprecated optional | Residual type only（判定は `capable_courses`） | 型掃除は Stage A と同時可だが **単独削除は不要・破壊でもない** |
| 24 | **FE master API** | `frontend/src/features/master/api/staff-reservation-types.ts` — capable GET/PUT が SoT; **excluded query key を invalidate する residual** (`STAFF_EXCLUDED_ST_KEY`) | Residual invalidate only（excluded GET/PUT クライアント本体は無し） | key 掃除は consumer 削除後 |
| 25 | **FE query-keys** | `frontend/src/lib/query-keys.ts` `"excluded-reservation-types"` union member | Residual | 同上 |
| 26 | **BE tests（exclusion facade / isolation / concurrency）** | 下記 §2.1 | Test-only だが Stage B 回帰の要 | 削除前にテスト書き換え必須 |
| 27 | **Staff association handler tests** | `backend/internal/staff/staff_association_handler_http_test.go` Get/Set Excluded characterization | Test-only | route 削除と同時 |
| 28 | **Staff routes snapshot** | `backend/internal/staff/routes_test.go` excluded path 2 本 | Test-only | 同上 |
| 29 | **lintscan FK inventory** | `backend/internal/lintscan/master_fk_write_inventory_lint_test.go` `ExcludedTypeIDs` | Test-only gate | facade フィールド削除時に更新 |
| 30 | **Docs / ERD / spec** | `docs/architecture/erd.md`, `docs/spec/reservation-to-record-flow.md`, `BE-pending.md` | Docs | drop 後に追随 |
| 31 | **Seed capabilities（SoT 側・削除対象外）** | `staff_reservation_capabilities.csv` + manifest | Live SoT | **維持** |

### 2.1 主要テスト消費者（抜粋）

| Test file | 役割 |
|-----------|------|
| `reservation/staff_affinity_facade_test.go` | inverse mapping + **zero dual-write** |
| `reservation/staff_affinity_test.go` | pure inverse helpers |
| `reservation/reservation_staff_exclusion_clinic_isolation_test.go` | exclusion write facade → capabilities only |
| `reservation/reservation_staff_exclusion_response_clinic_isolation_test.go` | derived excluded response の clinic scope |
| `reservation/reservation_staff_repository_test.go` | FindAllExcluded* derived behavior |
| `reservation/reservation_staff_repository_tx_atomicity_test.go` | tx atomicity + zero exclusion rows |
| `reservation/reservation_staff_full_replacement_concurrency_test.go` | concurrent replace + zero exclusion rows |
| `reservation/reservation_staff_junction_lock_race_test.go` | Stage B: both kinds count capabilities |
| `reservation/reservation_staff_service_test.go` | ExcludedTypeIDs create/update/list |
| `reservation/reservation_staff_request_test.go` | request mapping |
| `reservation/reservation_staff_handler_test.go` | ListExcluded mocks |
| `reservation/reservation_staff_service_readback_atomicity_test.go` | readback failures |
| `staff/staff_service_permissions_test.go` | Get/Set Excluded |
| `staff/staff_association_handler_http_test.go` | HTTP characterization |
| `frontend/.../filter-staff-candidates.test.ts` | excluded は ignore、capable が判定 |
| `frontend/.../ReservationFormModal.test.tsx` | MSW payload に excluded_courses 残存 |

---

## 3. Dual-write to physical table

| 観点 | 結果 |
|------|------|
| Production write path が `staff_reservation_exclusions` に INSERT/DELETE するか | **NO** |
| 根拠（実装） | `UpdateExcludedReservationTypes`: 「Replace capabilities only — zero production writes to exclusions table」（`reservation_staff_repository.go:391-393`） |
| 根拠（回帰） | `assert.Zero(..., "production write must not insert staff_reservation_exclusions")` 他 |
| 物理テーブルの残存理由 | seed / schema / RLS / model / 歴史的 dual-table; **read SoT ではない** |
| Seed が exclusion CSV をまだ入れるか | **YES**（`003_demo` CSV + manifest）— production runtime write とは別経路 |

**結論**: Stage B 要件「production dual-write ゼロ」は満たしている。Stage A が消すべきは **契約・型・route・table・seed の exclusion 面**であり、今は consumer が残る。

---

## 4. External contract breakage if deleted now: **YES**

今すぐ exclusion 面を削除すると壊れる surface:

1. **HTTP**
   - `GET|PUT /api/v1/masters/staffs/:id/excluded-reservation-types`
   - `POST|PUT /api/v1/clinics/:clinic_id/reservation-staffs` の body `excluded_type_ids`
   - 同 resource の response field `excluded_courses`
2. **OpenAPI** `StaffReservationExclusion`, `excluded_courses`, `excluded_type_ids`, excluded path
3. **FE**
   - `ReservationStaff.excluded_courses` 正規化
   - query-keys / invalidate residual
   - generated `StaffReservationExclusion`
4. **DB / seed / export**
   - table DDL + RLS
   - demo seed CSV/manifest
   - seed-export table list
5. **Go API surface**
   - `model.StaffReservationExclusion` と多数の service/repo シグネチャ
6. **Test / lint gates** 上記一式

**維持必須（削除対象にしない）**: `reservation-staffs` ルート自体、`on-duty-staffs`、`available-times`、`capable-reservation-types`、`staff_reservation_capabilities`。

---

## 5. Stage A recommendation

### **INVENTORY-ONLY（consumers remain）— NO CLEAN-GO**

| 項目 | 内容 |
|------|------|
| 本 unit で実施したこと | consumer inventory の再検証 + 本報告のみ |
| 本 unit で **実施しなかった**こと | code 削除、route 削除、model 削除、table drop migration 作成、seed 削除、OpenAPI 変更、dual-write 再導入、`available-staffs` 触手、`make migrate` |
| CLEAN-GO 条件 | §6 の preconditions が **全て** 完了し、破壊変更の **明示承認** があること |
| 現状 | preconditions **未達** → Stage A 破壊削除は **BLOCKED by consumers** |

---

## 6. Preconditions for future Stage A delete（consumer migration checklist）

破壊削除（CLEAN-GO）に進む前に、以下を **順に** 完了させる。いずれも本 packet の範囲外。

### 6.1 API / BE contract

- [ ] FE および既知クライアントが `excluded_type_ids` / `excluded_courses` を **送受信しない**ことを証明（rg + 契約テスト）
- [ ] `GET|PUT .../excluded-reservation-types` の **外部利用ゼロ**（または deprecation 期間終了 + 承認）
- [ ] `reservation-staffs` Create/Update を **capable 形**（または capability 専用 master API のみ）に寄せ、request から `excluded_type_ids` を除去
- [ ] response から `excluded_courses` を除去（`capable_courses` を OpenAPI に正式記載してから）
- [ ] staff service/handler/ports から Get/Set Excluded を除去
- [ ] reservation service/repo から `FindAllExcluded*` / `UpdateExcluded*` / inverse facade を除去（capable API のみ残す）
- [ ] `model.StaffReservationExclusion` と schema_drift / rls リストから除去

### 6.2 FE

- [ ] `use-reservation-types.ts` から `excluded_courses` 型・normalize を除去
- [ ] `filter-staff-candidates` の deprecated field とテストの exclusion fixture を掃除
- [ ] `staff-reservation-types.ts` の excluded query-key invalidate と `query-keys` union を除去
- [ ] tygo 再生成後 `StaffReservationExclusion` が消えることを確認

### 6.3 OpenAPI / docs

- [ ] `api.yaml` から exclusion schema / fields / paths を削除
- [ ] `capable_courses` / capable paths を契約として完全記載
- [ ] `erd.md` / `reservation-to-record-flow.md` / 関連 docs を capabilities-only に更新

### 6.4 Seed / export / DB

- [ ] demo seed から `staff_reservation_exclusions.csv` と manifest 行を削除（capabilities seed は維持）
- [ ] `seed-export/tables.go` から table 名を削除
- [ ] **新規 numbered migration** で `staff_reservation_exclusions` DROP + RLS policy 除去（**001_init の歴史的書き換えはしない**方針に従うこと）
- [ ] エージェントは migrate を auto-apply しない; ユーザーが `make migrate` を手動実行

### 6.5 Tests / gates

- [ ] exclusion facade 専用テストを capable-only 回帰に置換または削除
- [ ] routes_test / OpenAPI snapshot / lintscan `ExcludedTypeIDs` を更新
- [ ] `available_staffs_ban_test` は **維持**（WONTFILE のまま）
- [ ] 破壊変更の **PO / 運用明示承認** を記録

### 6.6 受け入れ（todo.md 再掲）

- [ ] exclusion production surface 削除
- [ ] drop migration あり
- [ ] Stage B 互換 consumer が無いことを inventory 再実行で証明

---

## 7. Explicit non-actions（本 packet）

| 禁止事項 | 遵守 |
|----------|------|
| `available-staffs` route/client の追加・再オープン | **遵守**（WONTFILE; 触らない） |
| drop migration の author / apply | **遵守**（author せず、migrate も実行せず） |
| dual-write の再導入（exclusions 表への production write） | **遵守**（変更なし; 現状 zero write を維持） |
| backend product code の変更 | **遵守** |
| 他 worktree / main の変更の revert | **遵守** |
| claim branch の削除 | **遵守**（USER のみ release） |

---

## 8. Evidence commands / counts（本 worktree 再検証）

> 実行手段: workspace `rg` 相当の content search（パス `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021a`）。

| Pattern | Scope | Approx hits / note |
|---------|-------|--------------------|
| `staff_reservation_exclusions` | `*.{go,sql,csv,json,yml,yaml,ts,tsx}` | **25** matching lines（DDL/seed/export/model/tests/comments） |
| `StaffReservationExclusion` | `*.{go,ts,tsx}` | **105** matching lines（model + facade signatures + tests + generated TS） |
| `excluded_courses` \| `excluded_type_ids` \| `ExcludedTypeIDs` \| `excluded-reservation-types` | `*.{go,ts,tsx,yml,yaml}` | **48** matching lines（API/FE/BE request-response） |
| dual-write 否定コメント / assert | `*.go` | production write ゼロを固定する assert/comment **≥5** |
| `capable-reservation-types` staff routes | `staff/handler.go` | GET+PUT live（SoT 側・削除対象外） |
| `available-staffs` ban | `staff/available_staffs_ban_test.go` | WONTFILE gate 存在 |

### 8.1 Key file paths（再確認済み）

```
backend/internal/model/staff_reservation_exclusion.go
backend/migrations/001_init.sql                          # CREATE + RLS
backend/migrations/seeds/003_demo/staff_reservation_exclusions.csv
backend/migrations/seeds/003_demo/manifest.json
backend/cmd/seed-export/tables.go
backend/internal/reservation/staff_affinity.go
backend/internal/reservation/reservation_staff_repository.go
backend/internal/reservation/reservation_staff_service.go
backend/internal/reservation/reservation_staff_request.go
backend/internal/reservation/reservation_staff_response.go
backend/internal/reservation/routes.go
backend/internal/staff/handler.go                        # excluded + capable routes
backend/internal/staff/staff_handler.go
backend/internal/staff/staff_service_permissions.go
backend/internal/staff/ports.go
backend/docs/api.yaml
frontend/src/types/generated/models.ts
frontend/src/hooks/use-reservation-types.ts
frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts
frontend/src/features/master/api/staff-reservation-types.ts
frontend/src/lib/query-keys.ts
reports/2026-07-31-todo-po-decisions-FINAL.md            # PO decision B→A
todo.md                                                  # Stage A remaining
```

### 8.2 Dual-write evidence quotes

- `reservation_staff_repository.go:368-370`:  
  `UpdateExcludedReservationTypes is a Stage B compatibility write facade.`  
  `... does not write staff_reservation_exclusions.`
- `staff_affinity_facade_test.go:35`:  
  `production write must not insert staff_reservation_exclusions`
- `reservation_staff_repository_tx_atomicity_test.go:127`:  
  `Stage B: exclusions table must not receive production writes`

---

## 9. Packet deliverable checklist

| 要求 | 結果 |
|------|------|
| Stage B SoT summary | §1 |
| Full consumer inventory table | §2 |
| Dual-write = NO in production | §3 |
| External breakage if deleted now = YES | §4 |
| Stage A = INVENTORY-ONLY / NO CLEAN-GO | §5 |
| Preconditions checklist | §6 |
| Explicit non-actions | §7 |
| Evidence commands / paths | §8 |
| product code 変更 | **なし** |
| drop migration | **なし** |
| **NO DELETE** 明示 | **本報告全体 / §0 / §5** |

---

## 10. Coordinator handoff

| Item | Value |
|------|-------|
| Report path | `reports/2026-07-31-task-021-stage-a-inventory.md` |
| Stage A disposition | **INVENTORY-ONLY — NO CLEAN-GO（consumers remain）** |
| Files changed (expected) | **this report only** under `reports/` |
| Next step (human / later packet) | §6 consumer migrations → 破壊変更承認 → Stage A delete + drop migration（別 packet） |
| Claim release | USER only: `git branch -D claim/TASK-021` after integrate/abandon |

**END OF INVENTORY — NO CODE DELETED.**
