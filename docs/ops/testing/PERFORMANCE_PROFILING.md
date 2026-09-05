# パフォーマンス測定・プロファイリングガイド

> **目的**: 現在利用できる測定手段、その安全境界、未実装部分を明示する。
> **最新更新**: 2026-09-05

本ファイルは測定手段の安全境界だけを持つ。STG 一覧の 2026-09-05 改善（auth / COUNT / preload）は `main` の実装に吸収済み。

## Local CI load auth boundary

The local API and spike k6 scripts require `LOAD_TEST_LOGIN_EMAIL` and
`LOAD_TEST_LOGIN_PASSWORD` with no fallback. The scheduled workflow sets `APP_ENV=test` and
uses the public synthetic non-production login only; it does not consume `STG_DEMO_*` values.
Both scripts retain fail-closed login-status, cookie, protected-endpoint, and zero-aggregate
checks. Actual Actions, fresh-DB, and k6 execution evidence is UNREPORTED/UNKNOWN. The STG
sustained script is a separate path and is not covered by this local-CI contract.

## 1. proposed targets と measurement state

以下は承認済み SLO/spec への link がなく、現時点では **proposed targets** である。

| target                   | value                 | automation state                                              |
| :----------------------- | :-------------------- | :------------------------------------------------------------ |
| initial display          | 1.5 s                 | not gated                                                     |
| incremental search       | 200 ms after debounce | not gated                                                     |
| save action              | 1.0 s                 | not gated                                                     |
| monthly report           | 3.0 s                 | not gated                                                     |
| process/container memory | 500 MB                | collector、duration、artifact、threshold enforcement が未実装 |

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

k6 の local API endpoint／spike scripts は `LOAD_TEST_LOGIN_EMAIL` / `LOAD_TEST_LOGIN_PASSWORD` のみを要求し、local CI の synthetic test context で fail-closed に login する。scheduled `load-test` job は `APP_ENV=test` を設定し、`STG_DEMO_*` を消費しない。

`k6-cf-stg-sustained.js` は別途承認済みの STG sustained path であり、その `STG_DEMO_*` 契約を local CI に再利用しない。ここで actual Actions、fresh DB、または k6 実行の結果を主張しない。production／共有 STG では実行せず、対象、rate window、stop condition、owner を個別承認で確定する。
