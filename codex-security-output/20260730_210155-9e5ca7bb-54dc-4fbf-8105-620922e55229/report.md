# Security Review: AnimalEkarte

## Scope

Exhaustive security review of the complete deterministic repository inventory: 4,019 files, including source, tests, generated exports, documentation, deployment configuration, scripts, and 45 visually inspected PNG assets.

- Scan mode: repository
- Target kind: git_worktree
- Target ID: target_sha256_cba2c45eb017a3da044d2efe21da80462474a971cc1df51888672c46e8ec8b5a
- Revision: b11b772c34c5c171c66131a946a3579fee288dfc
- Snapshot digest: codex-security-snapshot/v1:sha256:0ac915b839d00b295806d3f3abd49a3d6f87de4b5d775bf0460014bf0a9fdbf2
- Inventory strategy: repository
- Included paths: .
- Excluded paths: none
- Runtime or test status: Static repository assessment completed with source, tests, configuration, and operational documentation used as evidence and counterevidence.
- Artifacts reviewed: artifacts/01_context/threat_model.md, artifacts/02_discovery/in_scope_files.txt, artifacts/02_discovery/candidate_ledger.jsonl, artifacts/03_coverage/repository_coverage_ledger.md
- Scan context: AnimalEkarte is a veterinary electronic medical-record platform with clinic-scoped healthcare, identity, billing, messaging, and operational data.

Limitations and exclusions:
- External deployment state, browser behavior, or data provenance is retained as deferred coverage where repository evidence cannot close the exact fact.
- Excluded .git/\*\*: Git internal object and index metadata are not repository source content and were excluded by the deterministic inventory command.

### Scan Summary

| Field | Value |
| --- | --- |
| Reportable findings | 15 |
| Severity mix | medium: 5, low: 10 |
| Confidence mix | high: 14, medium: 1 |
| Coverage | partial |
| Validation mode | Compact source-control-sink validation followed by compact attack-path analysis. |

Canonical artifacts: `scan-manifest.json`, `findings.json`, and `coverage.json`. This report is a deterministic projection of those files.

## Threat Model

The primary risks are cross-clinic access, stale or bypassed authorization, disclosure or corruption of veterinary medical and owner data, privileged credential exposure, unaudited sensitive changes, and externally triggerable resource exhaustion.

### Assets

- Owner contact data, pet identity, medical records, clinical images, hospitalization data, and billing records.
- Staff accounts, role and permission assignments, JWT sessions, password-reset material, and seeded administrative credentials.
- LINE, SMTP, database, object-storage, Cloudflare, Vercel, and migration credentials.
- Audit records, consent state, integration settings, deployment state, and availability of the clinical service.

### Trust Boundaries

- Public Internet and third-party webhooks into unauthenticated Gin routes.
- Browser and LIFF clients into authenticated API, CSRF, RBAC, and clinic-selection middleware.
- The clinic_id tenant boundary across handlers, services, repositories, foreign keys, and preload relationships.
- Application processes into PostgreSQL, R2/S3 object storage, LINE, SMTP, and other external services.
- Repository contributors and CI workflows into Cloudflare, Vercel, production databases, and deployment credentials.

### Attacker Capabilities

- An unauthenticated remote sender able to reach public API and webhook endpoints.
- A least-privileged authenticated clinic staff member able to supply normal application input and race concurrent requests.
- A malicious repository contributor, branch author, or local co-tenant able to influence files, developer workflows, or shared host resources.
- An actor who obtains a repository-published or locally exposed credential, session artifact, object URL, or operational log.

### Security Objectives

- Derive clinic scope from authenticated identity and reject every cross-tenant relationship at mutation time.
- Apply authentication, authorization, and current-account checks fail-closed except for explicitly accepted and bounded policy exceptions.
- Preserve confidentiality and integrity of owner, pet, clinical, billing, consent, and credential data.
- Keep privileged changes and sensitive external actions auditable even when dependent systems fail.
- Constrain untrusted work, uploaded content, concurrency, and external responses to protect service availability.
- Keep secrets out of repository content, process arguments, logs, public storage, and unprotected state.

### Assumptions

- Release configuration validation is active and production secrets are supplied by the deployment environment.
- The authoritative SECURITY.md policy applies throughout the repository, with the explicitly documented PO-005 authentication exception evaluated as an accepted limitation.
- Repository evidence is authoritative for code and declared deployment behavior; unavailable live environment values are not inferred.

## Findings

| Finding | Severity | Confidence | Detailed write-up |
| --- | --- | --- | --- |
| [Migrations install an active system administrator with a repository-public password](#finding-1) | medium | high | inline below |
| [Staged filenames can execute shell commands through the commit-quality hook](#finding-2) | medium | high | inline below |
| [Staging dump verification exposes restored clinical data on all host interfaces](#finding-3) | medium | high | inline below |
| [Unbounded reorder arrays amplify one request into hundreds of thousands of updates](#finding-4) | medium | high | inline below |
| [Invalid LINE webhooks trigger body-size-by-clinic-count cryptographic work](#finding-5) | medium | high | inline below |
| [Owner names can inject spreadsheet formulas into aggregation CSV exports](#finding-6) | low | high | inline below |
| [Unbounded manual-article history accumulates and returns full-size snapshots](#finding-7) | low | high | inline below |
| [Unlimited multi-file selection creates an unbounded medical-image upload burst](#finding-8) | low | medium | inline below |
| [Treatment discount updates can invalidate authorization before the write transaction](#finding-9) | low | high | inline below |
| [Treatment-plan discount updates can bypass permission through a stale authorization snapshot](#finding-10) | low | high | inline below |
| [Unbounded tag-code mappings amplify one request into thousands of database inserts](#finding-11) | low | high | inline below |
| [Deployment workflow executes mutable `vercel@latest` with production credentials](#finding-12) | low | high | inline below |
| [Cage deletion can race a hospitalization assignment and hide an in-use cage](#finding-13) | low | high | inline below |
| [Draft records allow new image uploads after the pet is deceased](#finding-14) | low | high | inline below |
| [Owner discount updates can overwrite protected changes after a stale permission check](#finding-15) | low | high | inline below |

### Confidence Scale

| Label | Meaning |
| --- | --- |
| high | Direct evidence supports the finding with no material unresolved blocker. |
| medium | Evidence supports a plausible issue, but material runtime or reachability proof remains. |
| low | Evidence is incomplete and the item is retained only for explicit follow-up. |

<a id="finding-1"></a>

### [1] Migrations install an active system administrator with a repository-public password

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | Every migration bundle includes an active system administrator whose password is repository-public, and the production bootstrap explicitly runs that migration with only manual post-migrate cleanup. |
| Category | Hardcoded credentials / insecure default account |
| CWE | CWE-798 |
| Affected lines | backend/migrations/seeds/003_demo/accounts.csv:2, backend/internal/seedbundle/manifest.go:18-23, backend/cmd/migrate/main.go:475-510, docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md:73-78, frontend/e2e/README.md:93-96 |

#### Summary

Every environment migration loads an active system-administrator account whose matching plaintext password is published in the repository, allowing unauthenticated takeover if a new shared or production deployment becomes reachable before manual cleanup.

#### Root Cause

The unconditional seed bundle order includes demo accounts in every environment, including an active system administrator with a fixed bcrypt verifier. Repository documentation publishes the matching password, and deployment relies on unenforced manual cleanup rather than environment gating, forced rotation, or pre-serve disablement.

#### Validation

Every migration bundle includes an active system administrator whose password is repository-public, and the production bootstrap explicitly runs that migration with only manual post-migrate cleanup. Validation details were not recorded separately.

Validation method: static_seed_to_production_trace

#### Dataflow

The canonical finding records the affected path at backend/migrations/seeds/003_demo/accounts.csv:2, backend/internal/seedbundle/manifest.go:18-23, backend/cmd/migrate/main.go:475-510, docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md:73-78, frontend/e2e/README.md:93-96, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — The path can provide unauthenticated system-administrator takeover of a shared/production environment, but exploitation depends on a fresh deployment becoming reachable before cleanup. High impact with medium likelihood yields medium.

Raise if migration and exposure are automatic with a demonstrated public window; lower or ignore if production migration excludes demo bundles or forces credential rotation/disablement before serving traffic.

#### Remediation

Exclude demo/staging account bundles from production and other shared environments through an explicit fail-closed environment allowlist. Generate non-public one-time credentials only when demo seeding is intentionally enabled, force rotation on first use, and keep seeded privileged accounts disabled until an operator activates them through a secure bootstrap flow.

Tests:
- Run the production migration configuration and assert no demo/staging accounts are inserted and no active account accepts any repository-documented password.
- Run an explicitly enabled disposable demo migration and assert generated privileged credentials are unique per deployment, initially disabled or forced to rotate, and cannot be reused after bootstrap.

Preventive controls:
- Add CI policy that fails when production bundle order contains demo seeds or active privileged seed accounts with fixed credential material.

<a id="finding-2"></a>

### [2] Staged filenames can execute shell commands through the commit-quality hook

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | An active Bash-tool hook interpolates attacker-influenced staged filenames into shell-form execSync templates; Git permits metacharacters that become shell syntax. |
| Category | OS command injection |
| CWE | CWE-78 |
| Affected lines | .claude/settings.json:173-178, .claude/hooks/pre-bash-commit-quality.js:67-72, .claude/hooks/pre-bash-commit-quality.js:79-86, .claude/hooks/pre-bash-commit-quality.js:118-132 |

#### Summary

A malicious repository pathname containing shell metacharacters reaches active shell-form `execSync` command templates, so a subsequent Claude-driven commit can execute arbitrary commands with developer or coding-agent privileges.

#### Root Cause

The active hook reads repository-controlled staged filenames and interpolates each filename into `git show` command strings passed to shell-form `execSync`. Quoting the Git revision expression does not escape embedded shell syntax, and newline-delimited parsing also fails to preserve arbitrary Git pathnames.

#### Validation

An active Bash-tool hook interpolates attacker-influenced staged filenames into shell-form execSync templates; Git permits metacharacters that become shell syntax. Validation details were not recorded separately.

Validation method: static_shell_injection_trace

#### Dataflow

The canonical finding records the affected path at .claude/settings.json:173-178, .claude/hooks/pre-bash-commit-quality.js:67-72, .claude/hooks/pre-bash-commit-quality.js:79-86, .claude/hooks/pre-bash-commit-quality.js:118-132, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — Successful exploitation yields arbitrary local code execution and access to developer/agent credentials, but requires a malicious staged filename and a specific commit workflow. High impact with medium likelihood yields medium.

Raise if untrusted branches are routinely auto-staged/committed with sensitive credentials present; ignore once all pathname handling uses NUL-delimited input and argv-form child processes.

#### Remediation

Replace both shell-form calls with `execFileSync('git', ['show', ':' + file], options)` or `spawnSync` argv form, obtain filenames with `git diff --cached --name-only -z`, parse NUL-delimited bytes, and never concatenate repository pathnames into a shell command.

Tests:
- Stage files whose names contain command substitution, quotes, spaces, newlines, semicolons, and leading dashes; invoke the hook and assert no marker command executes and each exact pathname is inspected.
- Stub child-process calls and assert every `git show` invocation uses argv form with the entire `:<path>` value as one argument and no shell enabled.

Preventive controls:
- Ban shell-form child-process APIs for commands containing repository-derived values and enforce the rule with linting or code review checks in hook scripts.

<a id="finding-3"></a>

### [3] Staging dump verification exposes restored clinical data on all host interfaces

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The verifier publishes a full restored staging database with fixed credentials using Docker's all-interface port mapping. |
| Category | Sensitive data exposure / insecure network binding |
| CWE | CWE-668, CWE-798 |
| Affected lines | scripts/verify_seed_matches_stg_dump_full.sh:39-48, scripts/verify_seed_matches_stg_dump_full.sh:116-121, scripts/verify_seed_matches_stg_dump_full.sh:181-185 |

#### Summary

While the verifier runs, any network peer able to reach the published Docker port can use the fixed PostgreSQL superuser password to read or modify the restored staging database containing owner, pet, and clinical records.

#### Root Cause

The verification script combines predictable `postgres`/`verify` credentials with Docker's host-wide `${PORT}:5432` publication while restoring sensitive staging data. Random container naming and exit cleanup do not restrict access during the verification window.

#### Validation

The verifier publishes a full restored staging database with fixed credentials using Docker's all-interface port mapping. Validation details were not recorded separately.

Validation method: static_network_exposure_trace

#### Dataflow

The canonical finding records the affected path at scripts/verify_seed_matches_stg_dump_full.sh:39-48, scripts/verify_seed_matches_stg_dump_full.sh:116-121, scripts/verify_seed_matches_stg_dump_full.sh:181-185, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — The path can expose owner, pet, and clinical rows to an unauthenticated local-network actor, but only during a manual run on a reachable host. High impact with medium likelihood yields medium.

Raise if the script is routinely run on internet-reachable hosts; lower or ignore if publication is loopback-only or an enforced firewall makes untrusted peers unreachable.

#### Remediation

Avoid publishing the database port when possible and execute comparison commands inside the container network. If host access is required, bind explicitly to `127.0.0.1:${VERIFY_PORT}:5432`, generate a fresh high-entropy password per run, and fail closed if the requested bind address is not loopback.

Tests:
- Run the verifier and assert the PostgreSQL listener is reachable from loopback only and is not present on any external host interface.
- Assert two consecutive runs use different generated credentials and that cleanup removes the container and credential material on success, failure, and interruption.

Preventive controls:
- Add a shell-policy check that rejects Docker database port publication lacking an explicit loopback address and forbids fixed credentials in data-restoration utilities.

<a id="finding-4"></a>

### [4] Unbounded reorder arrays amplify one request into hundreds of thousands of updates

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The shared reorder request has no maximum item count and the generic persistence helper executes one update per supplied ID inside a transaction. |
| Category | Resource exhaustion / unbounded request cardinality |
| CWE | CWE-400, CWE-770 |
| Affected lines | backend/internal/medicalrecord/cage_handler.go:103-118, backend/internal/httpapi/slice.go:13-17, backend/internal/persistence/scope.go:118-151, backend/internal/medicalrecord/cage_repository.go:90-92 |

#### Summary

An authenticated master-data editor can send a near-body-limit array of repeated IDs and force the shared reorder helper to execute one database update per element inside a long-held transaction against the shared pool.

#### Root Cause

The shared reorder contract enforces only a nonempty array. It neither caps nor deduplicates IDs, and `ReorderByClinicID` translates every supplied element—including repeats—into a separate update within one transaction.

#### Validation

The shared reorder request has no maximum item count and the generic persistence helper executes one update per supplied ID inside a transaction. Validation details were not recorded separately.

Validation method: static_resource_amplification_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/medicalrecord/cage_handler.go:103-118, backend/internal/httpapi/slice.go:13-17, backend/internal/persistence/scope.go:118-151, backend/internal/medicalrecord/cage_repository.go:90-92, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — The amplification is directly attacker-controlled, single-request, and large enough to threaten a shared 50-connection database pool. Medium impact with high likelihood yields medium.

Lower or ignore if a strict ID-count/uniqueness cap or measured database budget bounds the path; raise only if reliable prolonged cross-tenant outage is demonstrated.

#### Remediation

Reject reorder requests above a small domain-specific item limit, require unique positive IDs, verify the submitted set matches the clinic's reorderable objects, and perform ordering with one bounded bulk update rather than one statement per element.

Tests:
- Assert duplicate IDs and arrays above the configured maximum are rejected before a transaction starts on cage and each sibling master-data reorder route.
- For the maximum valid unique list, assert the repository uses a bounded bulk statement count and produces the exact requested order.

Preventive controls:
- Adopt shared slice validators for maximum cardinality and uniqueness, plus database-statement budget assertions for bulk endpoints.

<a id="finding-5"></a>

### [5] Invalid LINE webhooks trigger body-size-by-clinic-count cryptographic work

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | Every invalid unauthenticated webhook request can enumerate all clinic settings, decrypt each secret, and HMAC a body up to 2 MiB; rate limiting is per-IP and still permits a high burst. |
| Category | Resource exhaustion / unauthenticated computational amplification |
| CWE | CWE-400 |
| Affected lines | backend/internal/lstep/routes.go:224-230, backend/internal/lstep/line_link_handler.go:51-72, backend/cmd/api/composition_runtime.go:528-535, backend/cmd/api/main.go:19-21, backend/internal/lstep/line_link_handler.go:17, backend/internal/lstep/line_link_service.go:325-357, backend/internal/lstep/line_link_service.go:230-238, backend/internal/lstep/line_link_service.go:360-365 |

#### Summary

Any remote unauthenticated sender can make the LINE webhook load and decrypt every clinic's channel secret and compute HMAC over an attacker-controlled body for each clinic before rejection, enabling scalable CPU, database, and decryption pressure.

#### Root Cause

Signature verification has no cheap, authenticated routing key that narrows an incoming webhook to one clinic/channel. The public handler accepts up to 2 MiB, and the verifier enumerates all settings, decrypts every secret, and HMACs the complete body for each candidate; per-IP limits do not bound distributed aggregate work.

#### Validation

Every invalid unauthenticated webhook request can enumerate all clinic settings, decrypt each secret, and HMAC a body up to 2 MiB; rate limiting is per-IP and still permits a high burst. Validation details were not recorded separately.

Validation method: static_public_resource_amplification_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/lstep/routes.go:224-230, backend/internal/lstep/line_link_handler.go:51-72, backend/cmd/api/composition_runtime.go:528-535, backend/cmd/api/main.go:19-21, backend/internal/lstep/line_link_handler.go:17, backend/internal/lstep/line_link_service.go:325-357, backend/internal/lstep/line_link_service.go:230-238, backend/internal/lstep/line_link_service.go:360-365, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — A remote unauthenticated actor can cheaply trigger database, decryption, and O(body×clinic-count) CPU work; controls reduce but do not close the path. Medium impact with high likelihood yields medium.

Raise only with evidence of prolonged clinical-service outage at realistic traffic; lower if verification selects one clinic/channel before expensive work or a global work budget caps aggregate cost.

#### Remediation

Parse only LINE's bounded destination/channel identifier needed for candidate selection, map it to one configured clinic secret through a cached index, then verify the exact raw body HMAC before any business parsing. Add a substantially smaller webhook body limit, global concurrency/work budget, and bounded secret-cache behavior.

Tests:
- With many clinic channel settings, send an invalid webhook for one destination and assert only that destination's secret is loaded/decrypted and only one HMAC is computed.
- Exercise oversized, unknown-destination, burst, and distributed-source invalid requests; assert early rejection and a fixed global verification/concurrency budget independent of clinic count.

Preventive controls:
- Monitor webhook verification work, invalid-signature rate, secret decrypts, and global concurrency, with alerts keyed to amplification rather than only request count.

<a id="finding-6"></a>

### [6] Owner names can inject spreadsheet formulas into aggregation CSV exports

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | The owner-name source, CSV encoder, browser download sink, and missing formula-prefix neutralization are directly visible; the regression test confirms only quoting is covered. |
| Category | CSV injection / spreadsheet formula injection |
| CWE | CWE-1236 |
| Affected lines | frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx:176-182, backend/internal/owner/http_request.go:166, frontend/src/features/aggregation/api/get-aggregations.ts:41-43, frontend/src/features/aggregation/components/aggregation-csv.ts:11-16, backend/internal/lstep/aggregation_handler.go:40-44, frontend/src/features/aggregation/components/aggregation-csv.ts:73-84, frontend/src/features/aggregation/components/aggregation-csv.test.ts:50-58 |

#### Summary

An attacker-influenced owner name beginning with a spreadsheet formula marker survives CSV quoting and is downloaded by clinic staff, so opening the aggregation export in a formula-capable spreadsheet can execute a cell formula in the staff workstation context.

#### Root Cause

The aggregation CSV encoder treats RFC-style quoting as the complete cell-safety control. It doubles embedded quotes but never neutralizes leading formula markers (`=`, `+`, `-`, or `@`) before owner-controlled values cross into a spreadsheet execution context.

#### Validation

The owner-name source, CSV encoder, browser download sink, and missing formula-prefix neutralization are directly visible; the regression test confirms only quoting is covered. Validation details were not recorded separately.

Validation method: static_source_control_sink_trace

#### Dataflow

The canonical finding records the affected path at frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx:176-182, backend/internal/owner/http_request.go:166, frontend/src/features/aggregation/api/get-aggregations.ts:41-43, frontend/src/features/aggregation/components/aggregation-csv.ts:11-16, backend/internal/lstep/aggregation_handler.go:40-44, frontend/src/features/aggregation/components/aggregation-csv.ts:73-84, frontend/src/features/aggregation/components/aggregation-csv.test.ts:50-58, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — The path can cross users and expose workstation/network actions, but requires authenticated data entry, a victim export/open action, and spreadsheet-dependent behavior. The impact-by-likelihood matrix yields low.

Raise only if a supported spreadsheet is shown to execute a concrete high-impact payload without meaningful warning; ignore if all supported consumers provably neutralize formula-prefixed quoted cells.

#### Remediation

Centralize CSV cell encoding and, before quoting, prefix any text cell whose first non-whitespace character is `=`, `+`, `-`, or `@` with a literal apostrophe (or another consumer-approved neutralization). Apply the same encoder to every exported text column rather than only owner names.

Tests:
- Export rows whose owner names begin with each of `=`, `+`, `-`, and `@`; assert the decoded CSV cells contain a neutralizing prefix and cannot begin with a formula marker.
- Open a generated regression fixture in every supported spreadsheet consumer and verify formula-like owner names render as inert text while commas, quotes, Unicode, and newlines still round-trip.

Preventive controls:
- Require all spreadsheet exports to use one tested formula-neutralizing CSV encoder and add a static review checklist item for formula-capable output formats.

<a id="finding-7"></a>

### [7] Unbounded manual-article history accumulates and returns full-size snapshots

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Every authorized manual edit can store another full body up to 100 KB, and history reads materialize and return all versions without retention or pagination. |
| Category | Resource exhaustion / unbounded persistent history |
| CWE | CWE-770 |
| Affected lines | backend/internal/manualarticle/handler.go:195-216, backend/internal/manualarticle/request.go:5-10, backend/internal/manualarticle/repository.go:135-143, backend/internal/manualarticle/repository.go:96-105, backend/internal/manualarticle/response.go:34-43 |

#### Summary

A tenant-level manual editor can repeatedly add full 100 KB global article snapshots, after which an unpaginated history request materializes and returns the entire accumulated body set, consuming shared storage, memory, and response capacity.

#### Root Cause

Every article upsert appends a complete body snapshot with no retention, quota, rate, or compaction limit. The versions endpoint has no pagination or hard response cap and loads and serializes all historical bodies at once.

#### Validation

Every authorized manual edit can store another full body up to 100 KB, and history reads materialize and return all versions without retention or pagination. Validation details were not recorded separately.

Validation method: static_persistent_resource_amplification_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/manualarticle/handler.go:195-216, backend/internal/manualarticle/request.go:5-10, backend/internal/manualarticle/repository.go:135-143, backend/internal/manualarticle/repository.go:96-105, backend/internal/manualarticle/response.go:34-43, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — The path crosses from one clinic executive into global persistent and read amplification, but requires accumulation and capacity impact is unquantified. The matrix yields low.

Raise if testing shows reliable cross-tenant outage at modest request counts; ignore if version retention, quota, pagination, and response limits bound worst-case cost.

#### Remediation

Enforce a bounded per-article version retention policy and per-principal/global write quota; paginate history with a strict page-size maximum and omit full bodies from list responses, fetching one version body by ID only when requested.

Tests:
- Create more versions than the retention limit and assert only the allowed number remains while the current article and audit metadata are preserved.
- Request version history for an article with many 100 KB snapshots and assert page-size limits, stable pagination, bounded response bytes, and absence of full bodies from the list payload.

Preventive controls:
- Set database/storage budgets and telemetry for append-only version tables, and require pagination plus payload-size limits on all history endpoints.

<a id="finding-8"></a>

### [8] Unlimited multi-file selection creates an unbounded medical-image upload burst

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | medium |
| Confidence rationale | The shipped UI permits unlimited file selection and immediately fans every file into parallel upload requests; each can carry 10 MiB and the server provides only per-request limits. |
| Category | Resource exhaustion / unbounded upload concurrency |
| CWE | CWE-770 |
| Affected lines | frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:80-86, frontend/src/features/medical-records/components/MedicalRecordImage.tsx:52-57, frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:60-70, backend/internal/medicalrecord/medical_record_image_handler.go:147-164, frontend/src/features/medical-records/api/medical-record-images.ts:36-42 |

#### Summary

An ordinary medical-record creator can select arbitrarily many allowed files and make the shipped client queue all 10 MiB uploads concurrently, creating sustained shared storage, multipart, database, network, and browser pressure without an aggregate server budget.

#### Root Cause

Upload validation is per-file only. The UI accepts unrestricted multiple selection and uses `Promise.all` to start every upload, while the backend caps each request independently but provides no demonstrated per-user, per-clinic, or global concurrent-upload and aggregate-byte budget.

#### Validation

The shipped UI permits unlimited file selection and immediately fans every file into parallel upload requests; each can carry 10 MiB and the server provides only per-request limits. Validation details were not recorded separately.

Validation method: static_concurrency_amplification_trace

#### Dataflow

The canonical finding records the affected path at frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:80-86, frontend/src/features/medical-records/components/MedicalRecordImage.tsx:52-57, frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:60-70, backend/internal/medicalrecord/medical_record_image_handler.go:147-164, frontend/src/features/medical-records/api/medical-record-images.ts:36-42, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — A tenant actor can generate sustained shared upload pressure through shipped functionality, but reliable cross-tenant outage is not proven. Medium impact and medium likelihood yield low.

Raise if bounded load testing demonstrates repeatable shared outage or cost abuse; ignore if server-side per-user/clinic concurrency and aggregate-byte quotas bound the worst case.

#### Remediation

Cap file count and aggregate selected bytes in the UI, replace `Promise.all` with a small bounded worker queue, and enforce authoritative server-side per-user/per-clinic concurrency, rate, and aggregate-byte/storage quotas so alternate clients cannot bypass the UI.

Tests:
- Select more than the allowed file count or aggregate bytes and assert the UI rejects the batch before starting any request; verify a maximum valid batch never exceeds the configured client concurrency.
- Bypass the UI with parallel multipart requests and assert server-side per-user/clinic concurrency and byte quotas throttle or reject excess uploads while normal bounded batches complete.

Preventive controls:
- Instrument concurrent uploads, aggregate bytes, storage writes, and quota rejections by user and clinic, and alert on sustained saturation.

<a id="finding-9"></a>

### [9] Treatment discount updates can invalidate authorization before the write transaction

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Treatment discount authorization occurs before the service transaction and the final write has no version or permission guard for the current discount. |
| Category | Authorization bypass / TOCTOU race |
| CWE | CWE-367, CWE-863 |
| Affected lines | backend/internal/medicalrecord/treatment_handler.go:180-215, backend/internal/medicalrecord/treatment_service.go:356-417, backend/internal/httpapi/discount_permission.go:25-50, backend/internal/medicalrecord/treatment_service.go:417-465 |

#### Summary

A treatment editor without discount permission can race an authorized update and overwrite protected discount rate or amount fields because the permission comparison occurs before the transaction that commits the request.

#### Root Cause

The treatment handler compares requested discounts with a pre-transaction snapshot and lets equal values bypass the permission check. Although the service later serializes writes with the medical-record lock, it does not reauthorize discount fields against a current locked treatment row before persisting them.

#### Validation

Treatment discount authorization occurs before the service transaction and the final write has no version or permission guard for the current discount. Validation details were not recorded separately.

Validation method: static_authorization_toctou_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/medicalrecord/treatment_handler.go:180-215, backend/internal/medicalrecord/treatment_service.go:356-417, backend/internal/httpapi/discount_permission.go:25-50, backend/internal/medicalrecord/treatment_service.go:417-465, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — This is a real but narrow protected-field authorization bypass with a race precondition. Medium impact and low likelihood yield low.

Raise if a deterministic or broader accounting impact is shown; ignore if discount authorization is repeated against the locked current treatment or guarded by version comparison.

#### Remediation

Within the same transaction and after acquiring the relevant medical-record/treatment lock, reload the treatment, compare both `discount_rate` and `discount_amount` with current values, require discount-edit permission for differences, and only then update. Add an expected-version guard to reject stale requests.

Tests:
- Race an ordinary treatment edit with an authorized `discount_rate` change and assert the stale ordinary request cannot restore the old rate.
- Repeat the interleaving for `discount_amount`, and verify non-discount edits remain available to users who lack discount-edit permission.

Preventive controls:
- Use a shared transactional helper for protected financial-field authorization so handler snapshots cannot authorize later writes.

<a id="finding-10"></a>

### [10] Treatment-plan discount updates can bypass permission through a stale authorization snapshot

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Discount permission is decided from a pre-transaction snapshot while the service later writes under a separate transaction without binding authorization to the current protected value. |
| Category | Authorization bypass / TOCTOU race |
| CWE | CWE-367, CWE-863 |
| Affected lines | backend/internal/medicalrecord/treatment_plan_handler.go:154-198, backend/internal/medicalrecord/treatment_plan_handler.go:232-261, backend/internal/httpapi/discount_permission.go:25-50, backend/internal/medicalrecord/treatment_plan_service.go:202-267 |

#### Summary

A treatment-plan editor without discount-edit permission can race an authorized discount change and overwrite the protected value because permission is checked against an earlier snapshot rather than the row state committed by the update transaction.

#### Root Cause

The handler treats equality with a pre-transaction treatment-plan snapshot as authorization to omit `discount:edit`. The service then starts a separate transaction, reloads the plan, and applies the caller-supplied discount without locking the compared version or re-evaluating discount authorization against current state.

#### Validation

Discount permission is decided from a pre-transaction snapshot while the service later writes under a separate transaction without binding authorization to the current protected value. Validation details were not recorded separately.

Validation method: static_authorization_toctou_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/medicalrecord/treatment_plan_handler.go:154-198, backend/internal/medicalrecord/treatment_plan_handler.go:232-261, backend/internal/httpapi/discount_permission.go:25-50, backend/internal/medicalrecord/treatment_plan_service.go:202-267, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — The path crosses an explicit financial permission boundary but is narrowly scoped and timing-dependent. Medium impact with low likelihood yields low.

Raise if a deterministic interleaving or broader financial overwrite is demonstrated; ignore if authorization and the compared value are bound inside the final transaction.

#### Remediation

Move the discount-change decision into the update transaction: lock the treatment-plan row, compare requested discount fields with the locked current values, require `discount:edit` when they differ, and persist only after that decision. Alternatively require an expected version and reject stale updates atomically.

Tests:
- Use two concurrent transactions so an authorized actor changes the discount after the ordinary editor's read; assert the ordinary editor's stale update is rejected and the authorized value remains.
- Verify an ordinary editor may update non-discount fields when discount values match the locked current row, while any actual discount change still requires `discount:edit` on both medical-record and hospitalization endpoints.

Preventive controls:
- Encapsulate protected-field authorization and mutation in one repository transaction with row locking or optimistic version predicates.

<a id="finding-11"></a>

### [11] Unbounded tag-code mappings amplify one request into thousands of database inserts

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Mapping replacement accepts unbounded entries/codes and performs individual inserts in a transaction after deleting existing mappings. |
| Category | Resource exhaustion / unbounded request cardinality |
| CWE | CWE-400, CWE-770 |
| Affected lines | backend/internal/lstep/lstep_tag_code_mapping_handler.go:30-60, backend/internal/lstep/lstep_tag_code_mapping_request.go:3-11, backend/internal/lstep/lstep_tag_code_mapping_service.go:69-106 |

#### Summary

A hospital-settings editor can submit an oversized mapping collection that deletes existing mappings and performs one insert per attacker-supplied entry inside a long transaction, consuming shared database and storage capacity.

#### Root Cause

The mapping request limits individual code length but has no maximum for entries, codes per entry, or aggregate codes. `PutMappingsForTag` then expands that unbounded cardinality into individual transactional `Create` operations and retains every result for the response.

#### Validation

Mapping replacement accepts unbounded entries/codes and performs individual inserts in a transaction after deleting existing mappings. Validation details were not recorded separately.

Validation method: static_resource_amplification_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/lstep/lstep_tag_code_mapping_handler.go:30-60, backend/internal/lstep/lstep_tag_code_mapping_request.go:3-11, backend/internal/lstep/lstep_tag_code_mapping_service.go:69-106, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — A malicious or compromised clinic administrator has a realistic shared-service resource-amplification path, but capacity impact is not quantified. Medium impact and medium likelihood yield low.

Raise if bounded load testing demonstrates reliable cross-tenant outage; ignore if an aggregate cardinality/query budget makes one tenant's worst-case request harmless.

#### Remediation

Define server-side maxima for entries per tag, codes per entry, and total encoded bytes/codes; reject duplicates and oversized requests before opening the transaction. Replace per-row inserts with a bounded bulk insert and cap response cardinality.

Tests:
- Submit exactly the documented maximum mapping cardinality and assert success, then submit one additional entry or code and assert a validation error occurs before any delete or insert.
- Instrument the repository for a maximum-size valid replacement and assert the operation uses a bounded number of database statements and leaves the previous mapping set intact on rejection.

Preventive controls:
- Require explicit cardinality, aggregate-size, and query-budget limits for every array-valued API request that expands into database work.

<a id="finding-12"></a>

### [12] Deployment workflow executes mutable `vercel@latest` with production credentials

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Staging and production jobs install a mutable latest Vercel CLI package immediately before executing it with deployment credentials, so upstream package substitution executes inside a privileged CI context. |
| Category | Software supply-chain compromise / unpinned executable dependency |
| CWE | CWE-829 |
| Affected lines | .github/workflows/frontend-deploy.yml:3-20, .github/workflows/frontend-deploy.yml:51-52, .github/workflows/frontend-deploy.yml:58-74 |

#### Summary

A compromised Vercel package publication or mutable `latest` tag can execute unreviewed code inside staging and production deployment jobs, exfiltrate Vercel credentials, or deploy attacker-controlled artifacts without any repository dependency change.

#### Root Cause

The credential-bearing deployment job globally installs the mutable `vercel@latest` package immediately before using it. No exact version, lockfile integrity, or provenance verification binds the executed CLI to a reviewed artifact.

#### Validation

Staging and production jobs install a mutable latest Vercel CLI package immediately before executing it with deployment credentials, so upstream package substitution executes inside a privileged CI context. Validation details were not recorded separately.

Validation method: static_ci_supply_chain_trace

#### Dataflow

The canonical finding records the affected path at .github/workflows/frontend-deploy.yml:3-20, .github/workflows/frontend-deploy.yml:51-52, .github/workflows/frontend-deploy.yml:58-74, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — A malicious resolved CLI could steal production-capable credentials or deploy code, but compromise of the trusted upstream/latest channel is relatively unlikely. High impact with low likelihood yields low.

Raise if package provenance or an actual malicious release is observed; ignore the mutable-resolution path once an exact reviewed version and integrity-controlled installation are enforced.

#### Remediation

Pin the Vercel CLI to a reviewed exact version in the repository lockfile, install it with integrity-enforcing frozen dependency resolution, invoke the local binary, and update it only through reviewed dependency changes. Constrain the deployment token to the minimum project and environment scope.

Tests:
- Assert the deployment workflow contains no mutable package tags and a clean frozen install resolves the exact reviewed Vercel CLI version and integrity from the lockfile.
- Simulate a newer `latest` publication and verify the workflow still executes the pinned local version; verify deployment credentials are scoped and unavailable to unrelated steps.

Preventive controls:
- Enforce CI policy that blocks mutable tags or global package installs in secret-bearing jobs and require provenance/integrity review for deployment tooling updates.

<a id="finding-13"></a>

### [13] Cage deletion can race a hospitalization assignment and hide an in-use cage

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Cage deletion performs a separate use-count check and soft delete, allowing a hospitalization assignment to commit between them; the database foreign-key delete behavior does not fire for soft deletion. |
| Category | Race condition / clinical state integrity |
| CWE | CWE-367 |
| Affected lines | backend/internal/medicalrecord/cage_service.go:168-187, backend/internal/medicalrecord/cage_repository.go:78-91, backend/migrations/001_init.sql:1341 |

#### Summary

A cage deleter can pass the not-in-use check immediately before another request assigns the cage, after which the soft delete succeeds and leaves a clinical hospitalization linked to a now-hidden cage.

#### Root Cause

The service performs `CountUsageByCageID` and soft deletion as separate, unserialized operations. Competing hospitalization assignment is not locked out and the delete statement has no atomic not-in-use predicate, while soft deletion does not trigger the database's physical-delete foreign-key behavior.

#### Validation

Cage deletion performs a separate use-count check and soft delete, allowing a hospitalization assignment to commit between them; the database foreign-key delete behavior does not fire for soft deletion. Validation details were not recorded separately.

Validation method: static_state_transition_race_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/medicalrecord/cage_service.go:168-187, backend/internal/medicalrecord/cage_repository.go:78-91, backend/migrations/001_init.sql:1341, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — The race bypasses a business-state integrity guard on a clinical association, but it is same-tenant and concurrency-dependent. The matrix yields low.

Raise if the detached association creates direct patient-safety impact at scale; ignore if deletion is serialized with assignments or uses an atomic not-in-use condition.

#### Remediation

Make assignment and deletion mutually serializable: lock the cage row in both paths, then perform an in-transaction usage check and soft delete, or issue an atomic conditional soft delete that succeeds only when no active hospitalization references the cage. Treat a zero-row update as an in-use conflict.

Tests:
- Run a deterministic concurrency test that pauses deletion after its usage check, commits a hospitalization assignment, then resumes deletion; assert deletion is rejected and the cage remains active.
- Verify deletion succeeds for a truly unused cage and that concurrent assignment either waits and succeeds before a rejected delete or fails after a completed delete without creating a hidden reference.

Preventive controls:
- Model check-then-delete business guards as transactional invariants and require concurrency regression tests for every `CountUsage`-to-delete path.

<a id="finding-14"></a>

### [14] Draft records allow new image uploads after the pet is deceased

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | The documented deceased-pet state is view-only, yet the image child mutation gate checks only finalization/permission and the backend image create path does not enforce pet liveness. |
| Category | Business-logic authorization bypass / clinical state integrity |
| CWE | CWE-841 |
| Affected lines | backend/internal/medicalrecord/medical_record_image_handler.go:134-145, frontend/src/features/medical-records/components/MedicalRecordImage.tsx:52-58, frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:149, frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:334-339, frontend/src/features/medical-records/api/medical-record-images.ts:21-32, backend/internal/medicalrecord/medical_record_image_service.go:125-150, frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx:211-217 |

#### Summary

Authenticated staff can add a new clinical attachment to a retained draft record after the pet has died because neither the form-wide image gate nor the backend create path enforces the documented deceased-pet view-only state.

#### Root Cause

The product's deceased-pet sentinel removes authority for new operations, but image creation checks only record finalization, tenant ownership, record ID, and ordinary create permission. The liveness state is omitted from both the frontend mutation gate and the authoritative backend service.

#### Validation

The documented deceased-pet state is view-only, yet the image child mutation gate checks only finalization/permission and the backend image create path does not enforce pet liveness. Validation details were not recorded separately.

Validation method: static_business_state_authorization_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/medicalrecord/medical_record_image_handler.go:134-145, frontend/src/features/medical-records/components/MedicalRecordImage.tsx:52-58, frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:149, frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:334-339, frontend/src/features/medical-records/api/medical-record-images.ts:21-32, backend/internal/medicalrecord/medical_record_image_service.go:125-150, frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx:211-217, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — The path bypasses an explicit clinical-state authorization rule and changes protected record integrity, but it is same-tenant and narrow. Medium impact and medium likelihood yield low.

Raise if this enables consequential clinical fraud or broad post-mortem modification; ignore if product policy explicitly permits retrospective images and the documentation/authorization model is updated consistently.

#### Remediation

In the backend image-create transaction, load the medical record and pet under clinic scope and reject creation when the pet is deceased, while preserving only the explicitly allowed finalize/delete workflow for pre-existing drafts. Mirror the same `isDeceased` condition in the form-wide image controls for immediate feedback.

Tests:
- Create a draft medical record, mark its pet deceased, then call the image-create API with valid permission; assert rejection and no object-storage or database artifact is created.
- Verify historical images remain viewable and the documented finalize/delete actions for pre-existing drafts still work, while image-create controls remain disabled for deceased pets in the UI.

Preventive controls:
- Centralize pet lifecycle mutation authorization in a backend policy helper and require every child-create endpoint to invoke it, with a route-level policy matrix regression suite.

<a id="finding-15"></a>

### [15] Owner discount updates can overwrite protected changes after a stale permission check

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Owner discount authorization uses an unlocked pre-service value and the repository later updates the owner without an expected-value or permission predicate. |
| Category | Authorization bypass / TOCTOU race |
| CWE | CWE-367, CWE-863 |
| Affected lines | backend/internal/owner/http_owner.go:96-129, backend/internal/owner/service_core.go:102-145, backend/internal/httpapi/discount_permission.go:25-36, backend/internal/owner/repository.go:287-323 |

#### Summary

An owner editor lacking discount-edit permission can race a legitimate discount update and restore a stale `discount_rate`, crossing the explicit financial-field permission boundary without holding the required privilege.

#### Root Cause

Owner discount authorization compares the request with an unlocked value read before the service transaction. The repository update has neither an expected-value/version predicate nor an in-transaction permission check, so the authorization decision is not bound to the row version that is written.

#### Validation

Owner discount authorization uses an unlocked pre-service value and the repository later updates the owner without an expected-value or permission predicate. Validation details were not recorded separately.

Validation method: static_authorization_toctou_trace

#### Dataflow

The canonical finding records the affected path at backend/internal/owner/http_owner.go:96-129, backend/internal/owner/service_core.go:102-145, backend/internal/httpapi/discount_permission.go:25-36, backend/internal/owner/repository.go:287-323, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — Unauthorized financial-field integrity is meaningful but narrow and race-dependent. The matrix yields low.

Raise if the race can be made deterministic or affects settled billing; ignore if the final write uses a locked value/expected version and in-transaction authorization.

#### Remediation

Lock the owner row inside the update transaction, compare `discount_rate` against that locked value, and require `discount:edit` for any difference before writing. If row locking is unsuitable, use an optimistic version or expected-discount predicate and reject/retry stale requests.

Tests:
- Interleave an authorized discount change between an ordinary editor's initial read and final update; assert the stale request fails without reverting the protected rate.
- Assert an ordinary owner editor can still change permitted fields when the expected version and discount match, and cannot change `discount_rate` without `discount:edit`.

Preventive controls:
- Prohibit authorization decisions based on mutable database snapshots unless the decision and write share a lock or compare-and-swap predicate.

## Reviewed Surfaces

| Surface | Risk Area | Outcome | Notes |
| --- | --- | --- | --- |
| Authentication, authorization, tenant isolation, and identity | Authentication and authorization | Reported | Clinic scope, role/permission propagation, stale claims, discount authorization, staff relationships, and administrative bootstrap credentials were traced through handlers, services, repositories, tests, and deployment guidance. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_3.md, artifacts/03_coverage/shard_4.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Medical records, hospitalization, cages, pets, and clinical child resources | Clinical data integrity | Reported | Tenant relations, mutation-time state, transactional revalidation, concurrency, deletion, and partial-write paths were assessed for attacker-reachable integrity impact. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_2.md, artifacts/03_coverage/shard_3.md, artifacts/03_coverage/shard_7.md, artifacts/02_discovery/candidate_ledger.jsonl |
| LINE/LSTEP, integration settings, request amplification, and service availability | External integrations and resource controls | Reported | Public webhook verification, authenticated bulk operations, mapping replacement, reorder behavior, external responses, and messaging controls were reviewed. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_2.md, artifacts/03_coverage/shard_6.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Frontend state, CSV export, file upload, rendering, and browser navigation | Browser and content handling | Reported | Spreadsheet formula handling and aggregate upload fan-out produced findings; browser URL-scheme behavior remains explicitly deferred. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_6.md, artifacts/03_coverage/shard_7.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Clinical image URLs, object storage access, deletion, and lifecycle | Sensitive object storage | Needs follow-up | Repository code proves persistence and deletion behavior, but authoritative deployed bucket access and lifecycle policy are external. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_3.md, artifacts/03_coverage/shard_7.md, artifacts/02_discovery/candidate_ledger.jsonl |
| CI/CD, dependency execution, repository hooks, and developer workflows | Software supply chain | Reported | Credential-bearing deployment jobs, mutable tools, action pinning, staged-path handling, and local session artifacts were assessed. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_1.md, artifacts/03_coverage/shard_8.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Docker, Cloudflare, Vercel, Terraform, operational scripts, and runbooks | Deployment and operational security | Reported | Network publication, dump handling, production readiness, process-argument exposure, local state, and environment separation were reviewed. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_5.md, artifacts/03_coverage/shard_8.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Audit durability, sensitive logging, and secret transport | Audit and secret exposure | Rejected | Validated defects were retained in the candidate ledger but attack-path policy rejected operator-only or no-privilege-delta instances from canonical findings. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_2.md, artifacts/03_coverage/shard_8.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Migrations, seed bundles, bootstrap accounts, and production initialization | Default credentials and environment separation | Reported | The fixed seed order and environment-independent production bootstrap path were traced end-to-end. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_4.md, artifacts/02_discovery/candidate_ledger.jsonl |
| Seed-data identity and credential-verifier provenance | Personal and credential data provenance | Needs follow-up | Repository evidence identifies a personal-looking record that violates seed guidance, but only the data owner can attest whether it is genuine. Evidence: artifacts/03_coverage/repository_coverage_ledger.md, artifacts/03_coverage/shard_4.md, artifacts/02_discovery/candidate_ledger.jsonl |

## Open Questions And Follow Up

- What is the authoritative production Vercel VITE_API_URL, and does a deployed production request reach only api.noah-karte.com?
  - Follow-up prompt: Capture the production Vercel environment value and one deployed browser or edge request trace without exposing credentials.
- Which persisted image URL schemes execute or navigate under each supported deployed browser and CSP?
  - Follow-up prompt: Exercise controlled non-HTTP(S) values in a staging clinic and record navigation, origin, opener, and CSP outcomes.
- Is the personal-looking staff seed identity synthetic, and has its bcrypt verifier ever been used by a live account?
  - Follow-up prompt: Obtain a data-owner attestation and credential-reuse check without disclosing the underlying password.
- Are deleted clinical-image objects publicly readable, and what lifecycle or revocation removes them?
  - Follow-up prompt: Inspect authoritative R2 bucket/domain policy and verify access before and after a record-image deletion.
- Authoritative Vercel production environment values and a deployed production request trace are unavailable.
  - Follow-up prompt: Review deferred unit candidate-c63e110aee32f1b1 and close its stated proof gap. Paths: package.json, frontend/src/lib/axios.ts, frontend/vercel.json, infra/scripts/cf-crud-smoke.sh. Surfaces: surface_infrastructure_operations.
- Supported-browser testing under the deployed CSP is required to determine accepted navigation schemes and meaningful victim-origin impact.
  - Follow-up prompt: Review deferred unit candidate-95f8b72f314e1c43 and close its stated proof gap. Paths: backend/internal/medicalrecord/routes.go, backend/internal/medicalrecord/medical_record_image_request.go, backend/internal/medicalrecord/medical_record_image_handler.go, frontend/src/features/medical-records/api/get-medical-record-images.ts, frontend/src/features/medical-records/components/ImageGalleryGroup.tsx. Surfaces: surface_frontend_exports_uploads.
- A data-owner provenance attestation and credential-reuse check are required to determine whether the seed record is genuine.
  - Follow-up prompt: Review deferred unit candidate-d05e6116c44de734 and close its stated proof gap. Paths: backend/migrations/seeds/003_demo/accounts.csv, backend/migrations/seeds/003_demo/staffs.csv, backend/internal/seedbundle/manifest.go, backend/cmd/migrate/main.go, backend/migrations/CLAUDE.md. Surfaces: surface_seed_provenance.
- Authoritative R2 bucket/public-domain access policy, object lifecycle rules, and the intended retention contract for deleted clinical images are external.
  - Follow-up prompt: Review deferred unit candidate-4c994efcd1016c7a and close its stated proof gap. Paths: backend/internal/medicalrecord/routes.go, backend/internal/medicalrecord/medical_record_image_handler.go, backend/internal/medicalrecord/medical_record_image_service.go, backend/internal/infra/s3_uploader.go, backend/internal/medicalrecord/handler_deps.go. Surfaces: surface_storage_url_lifecycle.
