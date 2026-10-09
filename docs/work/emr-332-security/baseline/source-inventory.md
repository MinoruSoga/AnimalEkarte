# U01 aux — Source Inventory (機械的導出の証跡)

Attempt att-u01-1 が baseline/route-middleware-map.md / baseline/test-classification.md を書くために読んだ登録箇所の一覧。行番号は base `de28e9765a29091339a872a85f1875a947e07e5b` 時点のもの。

## Staging origins

| ファイル | 行 | 内容 |
|---------|-----|------|
| frontend/wrangler.jsonc | 12, 16, 20 | name `animalekarte-stg-frontend`, workers_dev=true, route `stg.noah-karte.com/*` |
| backend/wrangler.jsonc | 24, 32, 37 | name `animalekarte-stg-api`, workers_dev=true, route `api.stg.noah-karte.com/*` |
| SECURITY.md | 91 | 2026-10-03 live 確認: 両 workers.dev ホスト名 (`*.baritech-soga.workers.dev`) |

## Route registration chain

| ファイル | 行 | 内容 |
|---------|-----|------|
| backend/cmd/api/main.go | 186-209 | `newRouter`: global Use チェーン (Recovery, SecurityHeaders, RequestID, CORS, RequestLoggingMiddleware, SanitizeNullBytes) |
| backend/cmd/api/main.go | 19-27 | webhook 50rps/200burst・support chat 20/min/burst10 のレート定数 |
| backend/cmd/api/base_routes.go | 77-94 | `/health`, `/api/v1/health`, `/health/db`, `/uploads` (STORAGE_TYPE 条件), scheduled jobs 呼出 |
| backend/cmd/api/batch_scheduler.go | 68-99 | `registerScheduledJobRoutes` + `requireSchedulerInternalToken` (X-Scheduler-Token, fail-closed) |
| backend/internal/scheduler/handler.go | 123-125 | `POST /_internal/scheduled-jobs/:jobAction` |
| backend/cmd/api/composition_auth_routes.go | 21-42 | authComposition.registerRoutes — CSRF=RequireXRequestedWith, Authenticate=AuthWithStaffValidationFailureNotifier, RateLimitFactory |
| backend/internal/auth/http_routes.go | 53-93 | 公開 auth 経路 6 本 (login/logout/refresh-logout/refresh は CSRF+RL、forgot/reset-password は RL のみ) + protected group (Authenticate+CSRF) + `/me`・`/users/me/password` |
| backend/cmd/api/composition_runtime.go | 508-538 | `registerRoutes`: base routes → auth registerRoutes → `protected.Use(LimitRequestBody)` → masters permission groups |
| backend/cmd/api/composition_runtime.go | 540-578 | `registerDomainRoutes`: owner(:551) pet(:552) staff(:553-556) clinic(:557) manualarticle(:600) support(:611) identitylink(:620) inventory(:625) trimming(:636) medicalrecord(:644) reservation(:562) billing(:567) lstep(:569) + LIFF(:563) + UAT(:568) + webhook(:570-577) |
| backend/cmd/api/composition_runtime.go | 581-589 | registerClinicRoutes (Clinic/ClinicHoliday/Company/ClosingSettings) |
| backend/cmd/api/composition_runtime.go | 591-645 | registerExistingDomainRoutes (manualarticle/support/identitylink/inventory/trimming/medicalrecord) |
| backend/cmd/api/composition_runtime.go | 662-692 | newLstepHandler / newReservationHandler (LiffAuth・LiffRateLimit・LiffBodyLimit 注入) |
| backend/internal/reservation/routes.go | 76-163 | 保護 reservation 経路 (masters/clinics/reservations) |
| backend/internal/reservation/routes.go | 167-199 | RegisterLiffRoutes — `/api/liff/:clinicId` 公開面 + LiffAuth authed グループ |
| backend/internal/lstep/routes.go | 118-224 | 保護 lstep 経路群 (settings/owners line/lstep*/tag-config/checkup-sync/lifecycle/line-customers/shared-files) |
| backend/internal/lstep/routes.go | 226-233 | RegisterWebhookRoutes — `POST /api/line/webhook` (JWT なし・HMAC+rate limit) |
| backend/internal/billing/synthetic_closing_http.go | 44-52, 137-155 | RegisterUATRoutes `/api/v1/uat/synthetic-closings` + allowHTTP ゲート (env+Host+Bearer secret, 404 fail-closed) |
| backend/internal/support/handler.go | 48-74 | `/support` chat (+POST chat rate limit) / bug-reports 共有ボード |
| backend/internal/medicalrecord/routes.go | 130-147 | `/masters`, `/medical-records`, `/hospitalizations` + registerLabRoutes 他 |
| backend/worker/index.ts | 503-561 | Worker fetch 分岐: `/_internal/migrate`, `/_internal/scheduler*`, その他 `/_internal/*` → 404, OPTIONS → edge preflight, else → Container (XFF 転記) |
| backend/worker/scheduler-ops.ts | 35, 83-107 | SCHEDULER_OPS_PREFIX=`/_internal/scheduler`, isInternalProxyPath (decode+normalize で `/_internal` 検出) |

## Test classification evidence

| 対象 | 結果 | 証拠 |
|------|------|------|
| internal/middleware/*_test.go (9 ファイル) | memory-only | 全ファイル `net/http/httptest` のみの外部依存; `driver.ErrBadConn` は fake エラー値; testdb import なし |
| internal/authjwt/ | テスト 0 件 | `claims.go` のみ (N/A) |
| internal/imageupload/ | テスト 0 件 | `imageupload.go` のみ (N/A) |
| internal/audit | service_test/service_operations_test/service_doubles_test/repository_test = memory-only; repository_db_test/repository_real_ddl_test(+helpers) = DB-dependent | `repository_db_test.go:33-36` setupAuditTestDB→setupAuditRealDDLTestDB; `repository_real_ddl_test.go:49` testdb.SetupIsolatedTestDB; `repository_test_helpers_test.go` TestMain→CloseSharedTestDB (nil-safe, testdb.go:127-133) |
| internal/testdb | DB 接続既定 `db:5432` | testdb.go:564-615 (`DB_HOST` 既定 `db` = Compose サービス名) |
| internal/auth | 55 テストファイル、うち 11 が testdb import | grep -l 'internal/testdb' による列挙 (baseline/test-classification.md §3) |
| cmd/api | 23 テストファイル、testdb import 0 | route_composition_smoke / staging_access_cache_config / get_head_permission_* を含む全件静的・httptest 系 |
| cmd/api/testdata/get_head_permissions.json | 210 GET/HEAD 経路 | class 内訳: clinic-fixed 159 / cross-clinic 21 / liff 9 / shared-master 9 / public 6 / self 3 / global 3 |

## その他確認済み事実

- go.mod `go 1.26.0`; docker-compose.yml:131 / Dockerfile.production:4 / Dockerfile.dev:1 とも `golang:1.26-alpine` — isolated Docker コマンドのイメージと一致。
- `git status` (w01 worktree、書込み前): clean。ブランチ `feat/emr-332-security-defense-w01`。
- `docs/work/emr-332-security/` は本ユニット着手前に存在しなかった (canonical・w01 両方で確認)。
