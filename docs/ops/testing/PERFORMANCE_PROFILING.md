# パフォーマンス測定・プロファイリングガイド

> **目的**: 現在利用できる測定手段、その安全境界、未実装部分を明示する。
> **最新更新**: 2026-09-01

## 1. proposed targets と measurement state

以下は承認済み SLO/spec への link がなく、現時点では **proposed targets** である。

| target | value | automation state |
|:--|:--|:--|
| initial display | 1.5 s | not gated |
| incremental search | 200 ms after debounce | not gated |
| save action | 1.0 s | not gated |
| monthly report | 3.0 s | not gated |
| process/container memory | 500 MB | collector、duration、artifact、threshold enforcement が未実装 |

Lighthouse は metrics を記録し performance category score 75 を script 内で判定するが、workflow step は `continue-on-error: true`。上記 timing targets を gate しない。

## 2. frontend

React DevTools Profiler と Lighthouse artifact を使う。`medical-records`、`accounting`、`reception` の rerender を実測してから最適化する。`memo`、`useCallback`、`useMemo` を無条件に適用しない。

## 3. backend

`backend/scripts/profile.go` は稼働中 backend を測っていなかったため削除済み。`/debug/pprof` と `net/http/pprof` も公開されていない。現時点で live backend profiler は **未実装**。Lighthouse や k6 を pprof の代替と呼ばない。

N+1 regression tests:

- `backend/internal/lstep/perf_n1_regression_test.go`
- `backend/internal/lstep/lstep_tag_sync_perf_n1_regression_test.go`

```bash
docker compose exec backend go test -v ./internal/lstep/ -run 'TestPERF|TestH1' -count=1
```

## 4. database safety

`EXPLAIN ANALYZE` は query を実行する。approved safe dataset の read-only SELECT に限定する。write を調べる必要がある場合は、承認済み環境で明示的 transaction と rollback を用いる手順を DBA/owner が用意する。production で ad-hoc に実行しない。growth table の例は `billings` と `lstep_delivery_trigger_log`。

## 5. k6

- `load-tests/k6-api-endpoints.js`: 50 steady VUs、`p(95)<500`
- `load-tests/k6-spike-test.js`: 100 spike VUs、`p(95)<2000`
- `load-tests/k6-cf-stg-sustained.js`: approved STG sustained run only

k6 scripts は `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` と実在 account を要求し、login failure で fail-closed する。現行 scheduled workflow は fresh `002_master` DB に account を provision しないため **BLOCKED**。secret names が設定されていても DB account は作られない。

現状は未固定の local k6 installation を推奨しない。approved route は、ephemeral fixture/account provisioning と version-pinned k6 runtime を workflow または Docker runner に追加した後のその経路とする。修正前に load run を実行しない。特に production/共有 STG は禁止し、approved isolated UAT target、rate window、stop condition、owner を先に決める。
