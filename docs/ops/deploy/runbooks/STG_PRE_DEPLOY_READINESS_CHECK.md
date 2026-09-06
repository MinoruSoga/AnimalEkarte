# STG リリース準備チェック

> **目的**: reviewed `main -> staging` delivery後のrepository-derived gatesを定義する。
> **境界**: Production approval gatesは別途実装・検証が必要。このcheckのPASSだけでproductionへmerge/pushしない。

## 1. Configuration audit

- [ ] target `backend/wrangler.jsonc`（またはproduction target file）の`secrets.required`をnames-only SSOTとして全nameを確認した。値を表示・記録していない。
- [ ] Wrangler `vars`を別に確認した。secretとvarを混同していない。
- [ ] workflowが要求するGitHub secret **names** をcurrent workflowから導出し、names-onlyで確認した。
- [ ] `.env.staging`をcurrent/tracked config surfaceとして使っていない。
- [ ] CIのrequired jobsがsuccessで、path-filterによる`skipped`をcomponentのgreen証跡にしていない。

## 2. Migration / seed gate

Normal backend deliveryは`wrangler deploy -> POST /_internal/migrate -> /health`の順。

- [ ] current `backend/migrations/*.sql`を全てplanへ含めた。
- [ ] exactly `BundleOrderForEnv(APP_ENV)`を適用した。現在 CSV は`002_master`のみ。
- [ ] staging では migrate log に `Login seed applied` と `seeds/003_login` の coverage を確認した。
- [ ] `002_master/manifest.json`から12-table inventory/load orderを導出し、listed filesが揃う。
- [ ] migrate logの`Migration key coverage missing=0`を確認した。固定row/key/table countを使っていない。
- [ ] checksum mismatchがない。
- [ ] legacy seed key がある場合は [現行 translation 契約](../SEED_MIGRATION_OPERATIONS.md#2-legacy-seed-keys) と対象 DB の master 完全性を確認した。不整合時は release stop とし、reviewed recovery plan を待つ。

`verify_seed_matches_stg_dump_full.sh`はmandatory gateではない。approved non-repository `STG_DUMP=/absolute/path`、data handling approval、master-only contractが揃う別rehearsalでのみ使える。通常checkoutにdumpが無いことをfailureにしない。

Shared DB rebuildはworkflow optionではない。[STG_PLANETSCALE_SEED_RUNBOOK.md](../STG_PLANETSCALE_SEED_RUNBOOK.md)のtarget/data-owner/backup/rollback/approval gateを満たす場合だけapproved operatorが行う。

## 3. Post-deploy checks

### 3.1 Infrastructure

- [ ] custom-domain `/health` が`200` / `{"status":"ok"}`。
- [ ] failure時はapproved workers.dev endpointと比較し、DNS/routeとWorker/Container/DBを分離した。
- [ ] backend workflowのdeploy/migrate/post-migrate healthが全てsuccess。
- [ ] image更新を伴う場合はdocumented rolling window後にも確認した。
- [ ] frontend の build-time API target、cookie/CORS、API JSON/status を [Vercel runbook](../VERCEL-FRONTEND-STAGING-TEST.md) で確認した。same-origin `/api` を使う build は rewrite も検証した。

### 3.2 Corrected CRUD cases

[CRUD-SMOKE-TEST.md](../CRUD-SMOKE-TEST.md)の次を実行する。

- A-1 `GET /clinics?scope=all` with `hospital-settings:view` -> 200
- A-2 same route without permission -> 403
- A-3 approved existing clinic PATCH、readback、元値restore
- B-1 valid permission-group payload -> 201
- B-2 active staff assignmentのあるclinic-scoped group DELETE -> 409
- B-3 smoke-created unused group cleanup -> 204
- C-1 CRUD-only staff create -> 201
- C-2 loginが必要ならapproved provisioning + `POST /api/v1/login`。remote mechanism未承認ならBLOCKED
- C-3 dependency protectionとsmoke-created staff cleanup

Clinic/staffはHTTP/resource state、permission-group成功mutationだけexplicit auditを確認する。run-created IDsだけをcleanupし、cookie/password/token/PHIを記録しない。

## 4. Stop criteria

1つでも該当すればrelease successにしない。

- migration coverage missing、checksum mismatch、legacy translation後のmaster不整合
- health failure、unexpected 4xx/5xx、tenant isolation failure
- frontend が実際に使う API 接続先が誤っている、または API request が SPA `index.html` へ fallback
- restore/cleanup failure
- required approval gate、backup/rollback、account/provisioningが未確認

AWSはretiredでrollback先ではない。Cloudflare/Vercel側の修正・last-known-good rebuild/redeployとcurrent infra runbookで復旧する。

## 5. Record

commit、workflow/run ID、target、各gateのPASS/FAIL/BLOCKED、nonsecret evidence、cleanup IDs/count、deferred blockerを記録する。external deployed stateをrepository configだけで「確認済み」にしない。
