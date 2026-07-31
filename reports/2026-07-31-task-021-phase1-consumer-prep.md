# TASK-021 Stage A Phase 1 — FE residual consumer prep（SAFE-CLEANUP）

| 項目 | 値 |
|------|-----|
| Packet / Unit | `TODO-MD-NEXT-ORCH-WAVE-20260731` / **W-021-P1** |
| Task | TASK-021 Stage A consumer prep **Phase 1 ONLY** |
| Worktree | `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p1` |
| Branch | `agent/w-021-p1` |
| Claim | `claim/W-021-P1`（取得済・本 unit は release しない） |
| 実施日 | 2026-07-31 |
| 親 inventory | `reports/2026-07-31-task-021-stage-a-inventory.md` |
| 決裁 | `reports/2026-07-31-todo-po-decisions-FINAL.md` § TASK-021 |
| **Disposition** | **SAFE-CLEANUP** |

---

## 0. 結論

| 項目 | 判定 |
|------|------|
| **Disposition** | **SAFE-CLEANUP** |
| production master staff UI が `excluded_*` API に bind しているか | **NO** — `capable-reservation-types` GET/PUT のみ |
| FE residual（query-key invalidate / deprecated field / fixture）を掃除したか | **YES** |
| backend route / model / table / drop migration を触ったか | **NO** |
| OpenAPI / generated models を手編集したか | **NO** |
| `StaffExcluded*` コンポーネント rename（cosmetic） | **NO**（禁止どおり未実施） |
| dual-write / available-staffs | **NO**（触らず） |
| CLEAN-GO（破壊削除）に進めるか | **NO** — §6 BE/OpenAPI/seed/DB 未達 |

本 Phase 1 は **FE のみ** の residual 掃除。BE exclusion facade / physical table / OpenAPI は inventory 通り残存。

---

## 1. 事前検証（rg — master UI binds capable_ only）

### 1.1 Master staff UI

| Surface | Evidence |
|---------|----------|
| API client | `frontend/src/features/master/api/staff-reservation-types.ts` — `useGetStaffCapableReservationTypes` / `useUpdateStaffCapableReservationTypes` のみ（`/capable-reservation-types`） |
| Re-export | `frontend/src/features/master/api/staffs.ts` — capable hooks のみ export |
| Settings | `StaffSettings.tsx` — `useUpdateStaffCapableReservationTypes` |
| Side panel | `StaffSidePanel.tsx` — `useGetStaffCapableReservationTypes` + `capableIds` / `capableIdSet` |
| Section UI | `StaffExcludedReservationTypesSection`（**名前は residual**）だが props は `capableIdSet` / checkbox は capable 肯定形 |

```text
rg capable-reservation-types|useGetStaffCapable|useUpdateStaffCapable frontend/src/features/master
# → staff-reservation-types.ts, staffs.ts, StaffSettings.tsx, StaffSidePanel.tsx のみ
```

### 1.2 excluded_* production FE consumer（Phase 1 前）

| Residual | Path | 用途 |
|----------|------|------|
| Query key + invalidate | `staff-reservation-types.ts` `STAFF_EXCLUDED_ST_KEY` | capable PUT 成功時に dead key を invalidate |
| Union member | `query-keys.ts` `"excluded-reservation-types"` | 上記専用 |
| Deprecated type field | `filter-staff-candidates.ts` `excluded_courses?` | 判定未使用 |
| Type + normalize | `use-reservation-types.ts` | レスポンス正規化のみ（読取なし） |
| Test fixtures | filter / ReservationFormModal tests | MSW / unit |

**excluded GET/PUT クライアント本体は存在しない**（Phase 1 前から）。

```text
rg "excluded-reservation-types|/excluded" frontend/src  # API path ヒットなし（query-key のみだった）
rg "\.excluded_courses" frontend/src                    # 読取プロパティアクセスなし
```

---

## 2. 実施した変更（SAFE-CLEANUP）

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/features/master/api/staff-reservation-types.ts` | `STAFF_EXCLUDED_ST_KEY` 削除 + excluded invalidate 削除。capable key + reservation-staffs predicate は維持 |
| 2 | `frontend/src/lib/query-keys.ts` | `staffs.subResource` union から `"excluded-reservation-types"` 除去 |
| 3 | `frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts` | deprecated `excluded_courses?` フィールド削除 |
| 4 | `frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.test.ts` | exclusion fixture 削除; capable 肯定形 assertion 維持 |
| 5 | `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.test.tsx` | MSW `excluded_courses` 削除（capable のみ） |
| 6 | `frontend/src/hooks/use-reservation-types.ts` | `ReservationStaff.excluded_courses` 必須フィールドと normalize を除去（BE facade は comment で無視を明記） |
| 7 | `reports/2026-07-31-task-021-phase1-consumer-prep.md` | 本報告 |

**Files touched: 7（≤12）。FE-only + report。**

### 2.1 意図的に触っていないもの

| 対象 | 理由 |
|------|------|
| `StaffExcludedReservationTypesSection` 名称 | cosmetic rename 禁止 |
| `types/generated/models.ts` | 手編集禁止 |
| BE routes / handlers / facade / OpenAPI | Phase 1 外; CLEAN-GO 前段 |
| `backend/migrations/**` | drop 禁止; 本 unit ゼロ変更 |
| capable staff UI ロジック | 壊さない |

---

## 3. 事後 rg サンプル

```text
# production TS に excluded_courses / excluded-reservation-types の残存はコメント 1 行のみ
rg -n "excluded_courses|excluded-reservation-types|STAFF_EXCLUDED" frontend/src --glob '*.{ts,tsx}'
# → frontend/src/hooks/use-reservation-types.ts:81
#    // Backend may still emit excluded_courses (Stage B facade); FE ignores it.

# master UI は capable のみ
rg -n "capable-reservation-types|useGetStaffCapable|useUpdateStaffCapable" frontend/src/features/master

# filter は capable_courses のみ
rg -n "capable_courses|filterStaffCandidatesByCapability" frontend/src/components/shared/ReservationFormModal
```

---

## 4. テスト

### 4.1 Docker mount 注意

- 常駐 `frontend` コンテナは **main tree**（`AnimalEkarte`）を `./frontend:/app` で mount。
- 本 worktree の検証は main の compose から **one-shot run** で worktree を bind override:

```bash
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
docker compose run --rm --no-deps \
  -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w021p1/frontend:/app \
  -v ekarte-frontend-node-modules:/app/node_modules \
  frontend pnpm exec vitest run \
    src/components/shared/ReservationFormModal \
    src/features/master/api \
    src/hooks/use-reservation-types \
    src/lib/query-keys
```

### 4.2 結果（scoped）

| 結果 | 値 |
|------|-----|
| Test Files | **9 passed (9)** |
| Tests | **61 passed (61)** |
| Duration | ~6.5s |
| 対象 | ReservationFormModal/* + master/api + use-reservation-types + query-keys 配下 |

含まれる重要ケース:
- `filter-staff-candidates.test.ts`（5）— capable 肯定形 / fail-closed
- `ReservationFormModal.test.tsx` — 担当者候補が capable のみ残る
- `StaffSettings.test.tsx` 等 master api 関連

### 4.3 全件 run について

`pnpm test:run -- <path>` は本 repo では path 絞り込みにならず full suite が走る。full では `generated-model-response-boundary.test.ts` が **identity-links allowlist drift**（本 packet 無関係・main 既存）で 1 fail。Phase 1 変更とは無関係。

**V-SCOPE / land 後**: main 統合後に上記 scoped vitest を再実行すること（常駐 frontend mount は main）。

---

## 5. Zero drop migration 証明

```text
git status -- backend/migrations
# nothing to commit, working tree clean（migrations 差分なし）

git diff --name-only
# frontend/src/... 6 files only（commit 前）
# + reports/2026-07-31-task-021-phase1-consumer-prep.md
```

- drop migration **未作成**
- `make migrate` **未実行**
- `backend/migrations/**` **未変更**

---

## 6. 残 §6 checklist（CLEAN-GO 向け — 本 Phase 未完了）

親 inventory §6 のまま。Phase 1 で **FE の一部のみ** check 可能。

### 6.1 API / BE contract — **未着手**

- [ ] FE および既知クライアントが `excluded_type_ids` / `excluded_courses` を **送受信しない**ことを証明（rg + 契約テスト）  
  - Phase 1: FE 型からは除去済。**BE はまだ response に `excluded_courses` を emit**（facade）
- [ ] `GET|PUT .../excluded-reservation-types` の外部利用ゼロ（または deprecation 終了 + 承認）
- [ ] `reservation-staffs` Create/Update から `excluded_type_ids` 除去（capable 形へ）
- [ ] response から `excluded_courses` 除去（`capable_courses` を OpenAPI 正式記載後）
- [ ] staff service/handler/ports から Get/Set Excluded 除去
- [ ] reservation の `FindAllExcluded*` / `UpdateExcluded*` / inverse facade 除去
- [ ] `model.StaffReservationExclusion` と schema_drift / rls リストから除去

### 6.2 FE — Phase 1 進捗

- [x] `use-reservation-types.ts` から `excluded_courses` 型・normalize を除去
- [x] `filter-staff-candidates` の deprecated field とテストの exclusion fixture を掃除
- [x] `staff-reservation-types.ts` の excluded query-key invalidate と `query-keys` union を除去
- [ ] tygo 再生成後 `StaffReservationExclusion` が消えることを確認（**BE model 削除後**）
- [ ] （任意・cosmetic 禁止）`StaffExcluded*` コンポーネント rename は **CLEAN-GO でも必須ではない**

### 6.3 OpenAPI / docs — **未着手**

- [ ] `api.yaml` から exclusion schema / fields / paths 削除
- [ ] `capable_courses` / capable paths を契約として完全記載
- [ ] erd / reservation-to-record-flow 等 docs 更新

### 6.4 Seed / export / DB — **未着手**

- [ ] demo seed から `staff_reservation_exclusions.csv` + manifest 行削除
- [ ] `seed-export/tables.go` から table 名削除
- [ ] **新規 numbered migration** で DROP + RLS 除去（001_init 歴史改変禁止）
- [ ] ユーザー手動 `make migrate`（エージェント auto-apply 禁止）

### 6.5 Tests / gates — **未着手（BE）**

- [ ] exclusion facade 専用テストを capable-only に置換
- [ ] routes_test / OpenAPI snapshot / lintscan `ExcludedTypeIDs` 更新
- [ ] `available_staffs_ban_test` **維持**
- [ ] 破壊変更の PO / 運用明示承認

### 6.6 受け入れ

- [ ] exclusion production surface 削除
- [ ] drop migration あり
- [ ] Stage B 互換 consumer が無いことを inventory 再実行で証明

---

## 7. FREEZE 遵守

| 禁止 | 遵守 |
|------|------|
| drop migration | **遵守**（未作成） |
| delete exclusion routes | **遵守** |
| delete table/model | **遵守** |
| OpenAPI mass delete | **遵守** |
| dual-write reintro | **遵守** |
| available-staffs | **遵守** |
| make migrate | **遵守** |
| backend/migrations/** drop | **遵守** |

---

## 8. Integration note（merge agent / `agent/w-021-p1`）

1. **Merge 対象**: FE residual 6 files + 本 report。conflict しやすいのは `query-keys.ts` / `use-reservation-types.ts`。
2. **BE との共存**: backend は当面 `excluded_courses` を response に載せ続けてよい。FE は ignore。破壊削除は §6 完了後の別 packet。
3. **検証**: land 後に §4.1 の scoped vitest を main docker で再実行（常駐 frontend は main mount）。
4. **Claim**: `claim/W-021-P1` は USER が main 統合後に `git branch -D claim/W-021-P1` で release。エージェントは削除しない。
5. **Next packet**: CLEAN-GO ではない。残 §6.1/6.3/6.4/6.5（BE contract → OpenAPI → seed/DB → tests）を別 unit で。

---

## 9. Return summary

| 項目 | 値 |
|------|-----|
| Disposition | **SAFE-CLEANUP** |
| Files | 6 FE + 1 report |
| Commit | （本報告と同時 path-scoped commit）`refactor(fe): Stage A phase1 residual exclusion consumer prep` |
| Test evidence | scoped vitest **61/61 pass**（worktree bind override） |
| Drop migration | **ゼロ**（`backend/migrations` clean） |
| Claim release | USER only |
