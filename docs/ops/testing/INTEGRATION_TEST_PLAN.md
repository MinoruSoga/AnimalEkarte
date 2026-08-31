# 統合テスト・品質保証計画書 (Integration Test Plan)

> **目的**: 自動テスト実装と性能試験の現状を定義する。層定義は [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md) を正本とする。
> **ステータス**: strategy document。execution state は CI、Linear、UAT report で追跡する。
> **最新更新**: 2026-09-01

## 1. Unit と API integration

- Backend: Go table-driven tests。必要に応じて testify を使い、`internal/<domain>` と `cmd/api` に colocate する。固定 Service/Repository layer を前提にしない。
- Frontend: Vitest + React Testing Library。
- Coverage: [coverage policy](../coverage-policy.md) の ratchet。
- CI: PostgreSQL 18 と全 migration を使う path-filtered `backend_build`、4 backend test shards、aggregate backend coverage ratchet。
- local `make ci`: inventory/guardrail family を含む。full project command の agent 自動実行は禁止されているため scoped check を選ぶ。

Scoped backend example:

```bash
docker compose exec backend go test ./internal/<domain>/... -run '<Name>' -count=1
```

## 2. E2E implementation state

実装済み specs は route/smoke/search と selected CRUD/flow regression が中心である。reservation-to-accounting や inpatient-to-billing の完全 journey を自動保証しない。詳細は [E2E guide](E2E_TESTING_GUIDE.md)。

Current supported route は local `make e2e` / `frontend/scripts/run-e2e.sh`。`.github/workflows/e2e.yml` は manual・non-gating であり、ephemeral account/fixture provisioning がないため authenticated suite は BLOCKED。

L4 の宣言 inventory は [scenarios/](scenarios/README.md)。wildcard/`要実測` gap が残るため exhaustive field coverage は未確立。

## 3. k6 / Lighthouse

| check | configured target | current automation state |
|:--|:--|:--|
| steady k6 | 50 VUs、`p(95)<500` | script に threshold あり。scheduled workflow の fresh master-only DB には login account provisioning がなく BLOCKED |
| spike k6 | 100 VUs、`p(95)<2000` | 同上 |
| sustained STG | `k6-cf-stg-sustained.js` | human-approved dedicated UAT target only |
| Lighthouse | performance score 75 | workflow は `continue-on-error`; §2 timing KPI の gate ではない |
| memory 500 MB | proposed/manual target | collector、duration、threshold enforcement、artifact が未実装 |

`.github/workflows/performance-tests.yml` は cron と `workflow_dispatch` が設定されているが、「configured」は「runnable/passing」を意味しない。k6 scripts は login failure で fail-closed し、workflow は account を作らない。修正までは BLOCKED と記録する。

## 4. 実行 boundary と report

1. `make up` は existing named Postgres volume を使うため「clean environment」ではない。
2. destructive clean DB contract (`make reset`) は user-only。agent は実行しない。
3. E2E/k6 は disposable local DB または approved isolated UAT tenant のみ。shared STG/production は禁止。
4. execution result は CI artifact、Linear、または `reports/uat-YYYY-MM-DD/` に記録する。strategy document を PASS evidence にしない。
