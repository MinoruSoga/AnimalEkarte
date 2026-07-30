# Codex Security レポート（日本語要約）
対象リポジトリ: AnimalEkarte
対象パス: /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
スキャンID: 9e5ca7bb-54dc-4fbf-8105-620922e55229
実行日時: 2026/7/30 20:19:25
進捗ステータス: complete
スコープ: .
モード: standard

## 要約
- 検出件数: **15件**
- 深刻度別: 低: 10件 / 中: 5件
- コスト（推定）: $125.079194
- カバレッジ: 0/4022（partial）

## 検出内容

### [1] Migrations install an active system administrator with a repository-public password
- 重大度: 中
- 信頼度: 高
- CWE: CWE-798
- 影響: N/A

#### 概要
Every environment migration loads an active system-administrator account whose matching plaintext password is published in the repository, allowing unauthenticated takeover if a new shared or production deployment becomes reachable before manual cleanup.

#### 根本原因
The unconditional seed bundle order includes demo accounts in every environment, including an active system administrator with a fixed bcrypt verifier. Repository documentation publishes the matching password, and deployment relies on unenforced manual cleanup rather than environment gating, forced rotation, or pre-serve disablement.

#### 推奨対応
Exclude demo/staging account bundles from production and other shared environments through an explicit fail-closed environment allowlist. Generate non-public one-time credentials only when demo seeding is intentionally enabled, force rotation on first use, and keep seeded privileged accounts disabled until an operator activates them through a secure bootstrap flow.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/seedbundle/manifest.go` (18-23)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/migrations/seeds/003_demo/accounts.csv` (2-2)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/migrate/main.go` (475-510)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md` (73-78)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/e2e/README.md` (93-96)

### [2] Staged filenames can execute shell commands through the commit-quality hook
- 重大度: 中
- 信頼度: 高
- CWE: CWE-78
- 影響: N/A

#### 概要
A malicious repository pathname containing shell metacharacters reaches active shell-form `execSync` command templates, so a subsequent Claude-driven commit can execute arbitrary commands with developer or coding-agent privileges.

#### 根本原因
The active hook reads repository-controlled staged filenames and interpolates each filename into `git show` command strings passed to shell-form `execSync`. Quoting the Git revision expression does not escape embedded shell syntax, and newline-delimited parsing also fails to preserve arbitrary Git pathnames.

#### 推奨対応
Replace both shell-form calls with `execFileSync('git', ['show', ':' + file], options)` or `spawnSync` argv form, obtain filenames with `git diff --cached --name-only -z`, parse NUL-delimited bytes, and never concatenate repository pathnames into a shell command.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/settings.json` (173-178)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/hooks/pre-bash-commit-quality.js` (67-72)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/hooks/pre-bash-commit-quality.js` (79-86)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/hooks/pre-bash-commit-quality.js` (118-132)

### [3] Staging dump verification exposes restored clinical data on all host interfaces
- 重大度: 中
- 信頼度: 高
- CWE: CWE-668, CWE-798
- 影響: N/A

#### 概要
While the verifier runs, any network peer able to reach the published Docker port can use the fixed PostgreSQL superuser password to read or modify the restored staging database containing owner, pet, and clinical records.

#### 根本原因
The verification script combines predictable `postgres`/`verify` credentials with Docker's host-wide `${PORT}:5432` publication while restoring sensitive staging data. Random container naming and exit cleanup do not restrict access during the verification window.

#### 推奨対応
Avoid publishing the database port when possible and execute comparison commands inside the container network. If host access is required, bind explicitly to `127.0.0.1:${VERIFY_PORT}:5432`, generate a fresh high-entropy password per run, and fail closed if the requested bind address is not loopback.

#### 対象箇所
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/scripts/verify_seed_matches_stg_dump_full.sh` (39-48)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/scripts/verify_seed_matches_stg_dump_full.sh` (116-121)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/scripts/verify_seed_matches_stg_dump_full.sh` (181-185)

### [4] Unbounded reorder arrays amplify one request into hundreds of thousands of updates
- 重大度: 中
- 信頼度: 高
- CWE: CWE-400, CWE-770
- 影響: N/A

#### 概要
An authenticated master-data editor can send a near-body-limit array of repeated IDs and force the shared reorder helper to execute one database update per element inside a long-held transaction against the shared pool.

#### 根本原因
The shared reorder contract enforces only a nonempty array. It neither caps nor deduplicates IDs, and `ReorderByClinicID` translates every supplied element—including repeats—into a separate update within one transaction.

#### 推奨対応
Reject reorder requests above a small domain-specific item limit, require unique positive IDs, verify the submitted set matches the clinic's reorderable objects, and perform ordering with one bounded bulk update rather than one statement per element.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/cage_handler.go` (103-118)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/httpapi/slice.go` (13-17)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/persistence/scope.go` (118-151)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/cage_repository.go` (90-92)

### [5] Invalid LINE webhooks trigger body-size-by-clinic-count cryptographic work
- 重大度: 中
- 信頼度: 高
- CWE: CWE-400
- 影響: N/A

#### 概要
Any remote unauthenticated sender can make the LINE webhook load and decrypt every clinic's channel secret and compute HMAC over an attacker-controlled body for each clinic before rejection, enabling scalable CPU, database, and decryption pressure.

#### 根本原因
Signature verification has no cheap, authenticated routing key that narrows an incoming webhook to one clinic/channel. The public handler accepts up to 2 MiB, and the verifier enumerates all settings, decrypts every secret, and HMACs the complete body for each candidate; per-IP limits do not bound distributed aggregate work.

#### 推奨対応
Parse only LINE's bounded destination/channel identifier needed for candidate selection, map it to one configured clinic secret through a cached index, then verify the exact raw body HMAC before any business parsing. Add a substantially smaller webhook body limit, global concurrency/work budget, and bounded secret-cache behavior.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/api/composition_runtime.go` (528-535)
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/api/main.go` (19-21)
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/line_link_handler.go` (17-17)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/routes.go` (224-230)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/line_link_handler.go` (51-72)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/line_link_service.go` (325-357)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/line_link_service.go` (230-238)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/line_link_service.go` (360-365)

### [6] Owner names can inject spreadsheet formulas into aggregation CSV exports
- 重大度: 低
- 信頼度: 高
- CWE: CWE-1236
- 影響: N/A

#### 概要
An attacker-influenced owner name beginning with a spreadsheet formula marker survives CSV quoting and is downloaded by clinic staff, so opening the aggregation export in a formula-capable spreadsheet can execute a cell formula in the staff workstation context.

#### 根本原因
The aggregation CSV encoder treats RFC-style quoting as the complete cell-safety control. It doubles embedded quotes but never neutralizes leading formula markers (`=`, `+`, `-`, or `@`) before owner-controlled values cross into a spreadsheet execution context.

#### 推奨対応
Centralize CSV cell encoding and, before quoting, prefix any text cell whose first non-whitespace character is `=`, `+`, `-`, or `@` with a literal apostrophe (or another consumer-approved neutralization). Apply the same encoder to every exported text column rather than only owner names.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx` (176-182)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/owner/http_request.go` (166-166)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/aggregation/api/get-aggregations.ts` (41-43)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/aggregation/components/aggregation-csv.ts` (11-16)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/aggregation_handler.go` (40-44)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/aggregation/components/aggregation-csv.ts` (73-84)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/aggregation/components/aggregation-csv.test.ts` (50-58)

### [7] Unbounded manual-article history accumulates and returns full-size snapshots
- 重大度: 低
- 信頼度: 高
- CWE: CWE-770
- 影響: N/A

#### 概要
A tenant-level manual editor can repeatedly add full 100 KB global article snapshots, after which an unpaginated history request materializes and returns the entire accumulated body set, consuming shared storage, memory, and response capacity.

#### 根本原因
Every article upsert appends a complete body snapshot with no retention, quota, rate, or compaction limit. The versions endpoint has no pagination or hard response cap and loads and serializes all historical bodies at once.

#### 推奨対応
Enforce a bounded per-article version retention policy and per-principal/global write quota; paginate history with a strict page-size maximum and omit full bodies from list responses, fetching one version body by ID only when requested.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/manualarticle/handler.go` (195-216)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/manualarticle/request.go` (5-10)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/manualarticle/repository.go` (135-143)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/manualarticle/repository.go` (96-105)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/manualarticle/response.go` (34-43)

### [8] Unlimited multi-file selection creates an unbounded medical-image upload burst
- 重大度: 低
- 信頼度: 中
- CWE: CWE-770
- 影響: N/A

#### 概要
An ordinary medical-record creator can select arbitrarily many allowed files and make the shipped client queue all 10 MiB uploads concurrently, creating sustained shared storage, multipart, database, network, and browser pressure without an aggregate server budget.

#### 根本原因
Upload validation is per-file only. The UI accepts unrestricted multiple selection and uses `Promise.all` to start every upload, while the backend caps each request independently but provides no demonstrated per-user, per-clinic, or global concurrent-upload and aggregate-byte budget.

#### 推奨対応
Cap file count and aggregate selected bytes in the UI, replace `Promise.all` with a small bounded worker queue, and enforce authoritative server-side per-user/per-clinic concurrency, rate, and aggregate-byte/storage quotas so alternate clients cannot bypass the UI.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/medical_record_image_handler.go` (147-164)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/components/ImageGalleryFilter.tsx` (80-86)
- エントリーポイント/ラッパー: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/components/MedicalRecordImage.tsx` (52-57)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/components/ImageGalleryFilter.tsx` (60-70)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/api/medical-record-images.ts` (36-42)

### [9] Treatment discount updates can invalidate authorization before the write transaction
- 重大度: 低
- 信頼度: 高
- CWE: CWE-367, CWE-863
- 影響: N/A

#### 概要
A treatment editor without discount permission can race an authorized update and overwrite protected discount rate or amount fields because the permission comparison occurs before the transaction that commits the request.

#### 根本原因
The treatment handler compares requested discounts with a pre-transaction snapshot and lets equal values bypass the permission check. Although the service later serializes writes with the medical-record lock, it does not reauthorize discount fields against a current locked treatment row before persisting them.

#### 推奨対応
Within the same transaction and after acquiring the relevant medical-record/treatment lock, reload the treatment, compare both `discount_rate` and `discount_amount` with current values, require discount-edit permission for differences, and only then update. Add an expected-version guard to reject stale requests.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/httpapi/discount_permission.go` (25-50)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_handler.go` (180-215)
- エントリーポイント/ラッパー: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_service.go` (356-417)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_service.go` (417-465)

### [10] Treatment-plan discount updates can bypass permission through a stale authorization snapshot
- 重大度: 低
- 信頼度: 高
- CWE: CWE-367, CWE-863
- 影響: N/A

#### 概要
A treatment-plan editor without discount-edit permission can race an authorized discount change and overwrite the protected value because permission is checked against an earlier snapshot rather than the row state committed by the update transaction.

#### 根本原因
The handler treats equality with a pre-transaction treatment-plan snapshot as authorization to omit `discount:edit`. The service then starts a separate transaction, reloads the plan, and applies the caller-supplied discount without locking the compared version or re-evaluating discount authorization against current state.

#### 推奨対応
Move the discount-change decision into the update transaction: lock the treatment-plan row, compare requested discount fields with the locked current values, require `discount:edit` when they differ, and persist only after that decision. Alternatively require an expected version and reject stale updates atomically.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/httpapi/discount_permission.go` (25-50)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_plan_handler.go` (154-198)
- エントリーポイント/ラッパー: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_plan_handler.go` (232-261)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_plan_service.go` (202-267)

### [11] Unbounded tag-code mappings amplify one request into thousands of database inserts
- 重大度: 低
- 信頼度: 高
- CWE: CWE-400, CWE-770
- 影響: N/A

#### 概要
A hospital-settings editor can submit an oversized mapping collection that deletes existing mappings and performs one insert per attacker-supplied entry inside a long transaction, consuming shared database and storage capacity.

#### 根本原因
The mapping request limits individual code length but has no maximum for entries, codes per entry, or aggregate codes. `PutMappingsForTag` then expands that unbounded cardinality into individual transactional `Create` operations and retains every result for the response.

#### 推奨対応
Define server-side maxima for entries per tag, codes per entry, and total encoded bytes/codes; reject duplicates and oversized requests before opening the transaction. Replace per-row inserts with a bounded bulk insert and cap response cardinality.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/lstep_tag_code_mapping_handler.go` (30-60)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/lstep_tag_code_mapping_request.go` (3-11)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/lstep/lstep_tag_code_mapping_service.go` (69-106)

### [12] Deployment workflow executes mutable `vercel@latest` with production credentials
- 重大度: 低
- 信頼度: 高
- CWE: CWE-829
- 影響: N/A

#### 概要
A compromised Vercel package publication or mutable `latest` tag can execute unreviewed code inside staging and production deployment jobs, exfiltrate Vercel credentials, or deploy attacker-controlled artifacts without any repository dependency change.

#### 根本原因
The credential-bearing deployment job globally installs the mutable `vercel@latest` package immediately before using it. No exact version, lockfile integrity, or provenance verification binds the executed CLI to a reviewed artifact.

#### 推奨対応
Pin the Vercel CLI to a reviewed exact version in the repository lockfile, install it with integrity-enforcing frozen dependency resolution, invoke the local binary, and update it only through reviewed dependency changes. Constrain the deployment token to the minimum project and environment scope.

#### 対象箇所
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/frontend-deploy.yml` (3-20)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/frontend-deploy.yml` (51-52)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/frontend-deploy.yml` (58-74)

### [13] Cage deletion can race a hospitalization assignment and hide an in-use cage
- 重大度: 低
- 信頼度: 高
- CWE: CWE-367
- 影響: N/A

#### 概要
A cage deleter can pass the not-in-use check immediately before another request assigns the cage, after which the soft delete succeeds and leaves a clinical hospitalization linked to a now-hidden cage.

#### 根本原因
The service performs `CountUsageByCageID` and soft deletion as separate, unserialized operations. Competing hospitalization assignment is not locked out and the delete statement has no atomic not-in-use predicate, while soft deletion does not trigger the database's physical-delete foreign-key behavior.

#### 推奨対応
Make assignment and deletion mutually serializable: lock the cage row in both paths, then perform an in-transaction usage check and soft delete, or issue an atomic conditional soft delete that succeeds only when no active hospitalization references the cage. Treat a zero-row update as an in-use conflict.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/cage_service.go` (168-187)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/cage_repository.go` (78-91)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/migrations/001_init.sql` (1341-1341)

### [14] Draft records allow new image uploads after the pet is deceased
- 重大度: 低
- 信頼度: 高
- CWE: CWE-841
- 影響: N/A

#### 概要
Authenticated staff can add a new clinical attachment to a retained draft record after the pet has died because neither the form-wide image gate nor the backend create path enforces the documented deceased-pet view-only state.

#### 根本原因
The product's deceased-pet sentinel removes authority for new operations, but image creation checks only record finalization, tenant ownership, record ID, and ordinary create permission. The liveness state is omitted from both the frontend mutation gate and the authoritative backend service.

#### 推奨対応
In the backend image-create transaction, load the medical record and pet under clinic scope and reject creation when the pet is deceased, while preserving only the explicitly allowed finalize/delete workflow for pre-existing drafts. Mirror the same `isDeceased` condition in the form-wide image controls for immediate feedback.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (149-149)
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (334-339)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/medical_record_image_handler.go` (134-145)
- エントリーポイント/ラッパー: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/components/MedicalRecordImage.tsx` (52-58)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/api/medical-record-images.ts` (21-32)
- 実装: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/medical_record_image_service.go` (125-150)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx` (211-217)

### [15] Owner discount updates can overwrite protected changes after a stale permission check
- 重大度: 低
- 信頼度: 高
- CWE: CWE-367, CWE-863
- 影響: N/A

#### 概要
An owner editor lacking discount-edit permission can race a legitimate discount update and restore a stale `discount_rate`, crossing the explicit financial-field permission boundary without holding the required privilege.

#### 根本原因
Owner discount authorization compares the request with an unlocked value read before the service transaction. The repository update has neither an expected-value/version predicate nor an in-transaction permission check, so the authorization decision is not bound to the row version that is written.

#### 推奨対応
Lock the owner row inside the update transaction, compare `discount_rate` against that locked value, and require `discount:edit` for any difference before writing. If row locking is unsuitable, use an optimistic version or expected-discount predicate and reject/retry stale requests.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/httpapi/discount_permission.go` (25-36)
- エントリーポイント: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/owner/http_owner.go` (96-129)
- エントリーポイント/ラッパー: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/owner/service_core.go` (102-145)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/owner/repository.go` (287-323)

## 補足（原文レポート）
- 英語版: `report.md`
- Scan show JSON: `scan-show.json`
