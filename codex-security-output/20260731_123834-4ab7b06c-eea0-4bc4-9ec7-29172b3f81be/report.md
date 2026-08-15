# Security Review: AnimalEkarte

## Scope

Repository-wide standard scan using a deterministic 4,064-file inventory.

- Scan mode: repository
- Target kind: git_worktree
- Target ID: target_sha256_cba2c45eb017a3da044d2efe21da80462474a971cc1df51888672c46e8ec8b5a
- Revision: e3f4f44dae0e586b36ae2ea21078bb5f93c50773
- Snapshot digest: codex-security-snapshot/v1:sha256:6489adf4633213698bbc929ea421747eab5152e827fc4b9af4d2c6bc3f347ca2
- Inventory strategy: repository
- Included paths: .
- Excluded paths: none
- Runtime or test status: Static review completed; Worker typecheck passed. Backend and Worker test execution was sandbox-blocked from temporary-file writes.
- Artifacts reviewed: artifacts/01_context/threat_model.md, artifacts/02_discovery/in_scope_files.txt, artifacts/02_discovery/candidate_ledger.jsonl, artifacts/03_coverage/reviewed_surfaces.md
- Scan context: Non-interactive repository-wide Codex Security scan.

Limitations and exclusions:
- No external deployment, CI-secret, or production-service access was used.
- The scan did not expose or attempt to crack sensitive seed-account material.
- Excluded .git/\*\*: Git metadata is excluded by the standard repository inventory procedure; the reviewed worktree content is bound by the supplied snapshot digest.

### Scan Summary

| Field | Value |
| --- | --- |
| Reportable findings | 6 |
| Severity mix | medium: 5, low: 1 |
| Confidence mix | high: 6 |
| Coverage | complete |
| Validation mode | Compact standard-scan validation and attack-path analysis over one combined candidate ledger. |

Canonical artifacts: `scan-manifest.json`, `findings.json`, and `coverage.json`. This report is a deterministic projection of those files.

## Threat Model

Multi-tenant veterinary-record application with browser, Go API, PostgreSQL, Cloudflare Worker, and CI/deployment trust boundaries.

### Assets

- clinic-scoped records
- billing and inventory integrity
- sessions and permissions
- integration and deployment credentials

### Trust Boundaries

- browser to API
- clinic selection to request-time authorization
- API to database and integrations
- repository to CI/deployment dependencies

### Attacker Capabilities

- untrusted request and stored-data input
- low-privilege authenticated actions
- supply-chain compromise of third-party CI dependencies

### Security Objectives

- tenant isolation
- record and inventory integrity
- credential confidentiality
- safe external integration

### Assumptions

- frontend controls are not authorization controls
- production and staging migration plans exclude demo seeds

## Findings

| Finding | Severity | Confidence | Detailed write-up |
| --- | --- | --- | --- |
| [Deployment workflows use mutable third-party Action references](#finding-1) | medium | high | inline below |
| [Fractional treatment quantities bypass inventory decrement](#finding-2) | medium | high | inline below |
| [Tracked demo seed exposes privileged-account verifiers and staff identity data](#finding-3) | medium | high | inline below |
| [Performance workflow trusts an unpinned k6 installation path before credentialed execution](#finding-4) | medium | high | inline below |
| [Secret-sync workflow executes an unverified npm artifact](#finding-5) | medium | high | inline below |
| [LSTEP migration report permits spreadsheet-formula injection](#finding-6) | low | high | inline below |

### Confidence Scale

| Label | Meaning |
| --- | --- |
| high | Direct evidence supports the finding with no material unresolved blocker. |
| medium | Evidence supports a plausible issue, but material runtime or reachability proof remains. |
| low | Evidence is incomplete and the item is retained only for explicit follow-up. |

<a id="finding-1"></a>

### [1] Deployment workflows use mutable third-party Action references

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The action tags and later credentialed commands are directly visible in both workflows. |
| Category | CI supply chain |
| CWE | CWE-829 |
| Affected lines | .github/workflows/backend-deploy.yml:33-41, .github/workflows/frontend-deploy.yml:39-45 |

#### Summary

Backend and frontend deployment jobs execute checkout, pnpm setup, and Node setup through version tags rather than immutable commit SHAs before credentialed deployment commands.

#### Root Cause

Privileged workflows trust mutable external action labels rather than reviewed, immutable action revisions.

**Backend action tags** — `.github/workflows/backend-deploy.yml:33-41`

These semantic version tags are not immutable commit references.

```yaml
actions/checkout@v7.0.1, pnpm/action-setup@v6, and actions/setup-node@v7
```

#### Validation

The action tags and later credentialed commands are directly visible in both workflows. Validation details were not recorded separately.

Validation method: static CI supply-chain review

#### Dataflow

The canonical finding records the affected path at .github/workflows/backend-deploy.yml:33-41, .github/workflows/frontend-deploy.yml:39-45, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — A compromised or retargeted upstream Action tag can execute on a runner before Cloudflare or Vercel credentials are used; this requires a third-party publisher compromise.

Severity rises if the workflows become reachable from untrusted pull requests.

#### Remediation

Pin every third-party Action in privileged workflows to a full commit SHA and use an automated reviewed updater for SHA changes.

Tests:
- Add a workflow lint rule requiring full SHA pins for third-party actions in secret-bearing jobs.

Preventive controls:
- Scope tokens minimally and protect deployment environments.

<a id="finding-2"></a>

### [2] Fractional treatment quantities bypass inventory decrement

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The exact float-to-int conversion is visible on the path from HTTP request to atomic update. |
| Category | Business logic integrity |
| CWE | CWE-682, CWE-840 |
| Affected lines | backend/internal/inventory/repository.go:152-157, backend/internal/medicalrecord/treatment_request.go:14, backend/internal/medicalrecord/treatment_service.go:337-340 |

#### Summary

Inventory-backed treatments accept fractional quantities, but stock decrement truncates the value to an integer before enforcing availability and updating stock.

#### Root Cause

The treatment quantity type and inventory quantity semantics are inconsistent, with loss of precision at the enforcement point.

**Truncating stock decrement** — `backend/internal/inventory/repository.go:152-157`

Converting a positive fractional quantity to int yields zero before both the stock guard and decrement.

```go
func (r *repository) DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error { qty := int(quantity); ... UpdateColumn("quantity", gorm.Expr("quantity - ?", qty)) }
```

#### Validation

The exact float-to-int conversion is visible on the path from HTTP request to atomic update. Validation details were not recorded separately.

Validation method: static code understanding

#### Dataflow

The canonical finding records the affected path at backend/internal/inventory/repository.go:152-157, backend/internal/medicalrecord/treatment_request.go:14, backend/internal/medicalrecord/treatment_service.go:337-340, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — A clinic staff member authorized to create treatments can defeat stock accounting and availability checks for inventory-backed medical items.

Severity rises if fractional quantities are valid for regulated or high-value supplies.

#### Remediation

Either reject non-integral quantities for integer-stock items before creating the treatment, or store and atomically decrement inventory using a precise decimal representation.

Tests:
- Assert that quantity 0.5 is rejected or decrements precisely and that 1.5 cannot under-decrement stock.

Preventive controls:
- Use one domain quantity type for treatments and inventory.

<a id="finding-3"></a>

### [3] Tracked demo seed exposes privileged-account verifiers and staff identity data

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The tracked CSV, migration plan, and policy are explicit; no credential value was exposed in this report. |
| Category | Sensitive data exposure |
| CWE | CWE-359, CWE-522 |
| Affected lines | backend/migrations/seeds/003_demo/accounts.csv:2, backend/cmd/migrate/main.go:490-492, backend/migrations/CLAUDE.md:69 |

#### Summary

The local/test demo seed stores active system-admin account rows, staff email identities, and bcrypt password verifiers in the repository despite the repository policy prohibiting that material in seeds or Git history.

#### Root Cause

Privileged local/test identities and password verifiers are committed rather than provisioned through a non-versioned local or secret-managed path.

**Seed-data policy** — `backend/migrations/CLAUDE.md:69`

The repository explicitly defines the tracked data as prohibited.

```markdown
実スタッフの氏名・email・password hashなどのPII/credential verifierをseedやGit履歴へ追加しない。
```

#### Validation

The tracked CSV, migration plan, and policy are explicit; no credential value was exposed in this report. Validation details were not recorded separately.

Validation method: static repository and configuration review

#### Dataflow

The canonical finding records the affected path at backend/migrations/seeds/003_demo/accounts.csv:2, backend/cmd/migrate/main.go:490-492, backend/migrations/CLAUDE.md:69, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — Repository readers receive data usable for offline password attacks or test-environment access; production and staging have a seed-selection mitigation.

Severity rises if any demo credential, identity, or seed bundle is reused in a release environment.

#### Remediation

Replace tracked identities and verifiers with synthetic disabled fixtures, and provision any necessary local accounts through a secret-managed one-time path.

Tests:
- Add a seed-content test that rejects active privileged accounts, email addresses, and password-hash columns or values in tracked demo bundles.

Preventive controls:
- Run secret and PII scanning on seed exports before commit.

<a id="finding-4"></a>

### [4] Performance workflow trusts an unpinned k6 installation path before credentialed execution

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The package installation and later credentialed commands are explicit in one workflow. |
| Category | CI supply chain |
| CWE | CWE-829, CWE-494 |
| Affected lines | .github/workflows/performance-tests.yml:52, .github/workflows/performance-tests.yml:90-108 |

#### Summary

The performance workflow imports a remote APT key through apt-key, installs an unversioned k6 package, and later runs it with staging demo credentials.

#### Root Cause

The workflow extends its package trust root dynamically without binding it to an expected key or artifact.

**Unpinned k6 installation** — `.github/workflows/performance-tests.yml:52`

Neither an expected signing-key fingerprint nor a package version/digest is enforced.

```yaml
curl https://dl.k6.io/key.gpg | sudo apt-key add - && sudo add-apt-repository ... && sudo apt-get install k6
```

#### Validation

The package installation and later credentialed commands are explicit in one workflow. Validation details were not recorded separately.

Validation method: static CI supply-chain review

#### Dataflow

The canonical finding records the affected path at .github/workflows/performance-tests.yml:52, .github/workflows/performance-tests.yml:90-108, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — A compromised external key, repository, or package can run code on the hosted runner before credentialed k6 commands, but requires upstream distribution compromise.

Severity rises if the workflow handles production secrets.

#### Remediation

Use a version-and-digest-pinned k6 artifact or image, or verify a known signing-key fingerprint and exact package version before installation.

Tests:
- Add a CI policy test rejecting apt-key, unversioned privileged package installs, and unverified remote keys.

Preventive controls:
- Keep secrets unavailable until after verified tool setup where possible.

<a id="finding-5"></a>

### [5] Secret-sync workflow executes an unverified npm artifact

| Field | Value |
| --- | --- |
| Severity | medium |
| Confidence | high |
| Confidence rationale | The secret environment and npx execution are adjacent in the workflow. |
| Category | CI supply chain |
| CWE | CWE-829, CWE-494 |
| Affected lines | .github/workflows/worker-secret-sync.yml:25-38 |

#### Summary

The staging Worker secret-sync job downloads and executes Wrangler through npx in the same step that exposes Cloudflare and database secrets, without a lockfile or artifact-digest integrity binding.

#### Root Cause

A secret-bearing workflow resolves executable package code outside the repository's lockfile and without artifact verification.

**Credentialed npx execution** — `.github/workflows/worker-secret-sync.yml:25-38`

Registry-fetched code executes with the deployment token and database credentials available.

```yaml
CLOUDFLARE_API_TOKEN and staging DB secrets are exported before npx -y wrangler@4.107.0 secret put commands.
```

#### Validation

The secret environment and npx execution are adjacent in the workflow. Validation details were not recorded separately.

Validation method: static CI supply-chain review

#### Dataflow

The canonical finding records the affected path at .github/workflows/worker-secret-sync.yml:25-38, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Medium** — A substituted package can exfiltrate a Cloudflare deployment token and staging database credentials, though exploitation requires compromise of the external package-distribution path.

Severity rises if the token is production-wide or the workflow becomes reachable by untrusted code.

#### Remediation

Install Wrangler through a frozen lockfile and invoke the resolved local binary, or verify a downloaded artifact against a maintained digest before exposing secrets.

Tests:
- Add a workflow policy check rejecting npx or npm remote execution in secret-bearing steps.

Preventive controls:
- Use least-privilege, environment-scoped deployment tokens.

<a id="finding-6"></a>

### [6] LSTEP migration report permits spreadsheet-formula injection

| Field | Value |
| --- | --- |
| Severity | low |
| Confidence | high |
| Confidence rationale | Source-to-sink trace is direct and a neighboring CSV exporter implements the missing control. |
| Category | CSV formula injection |
| CWE | CWE-1236 |
| Affected lines | backend/cmd/lstep-migrate/reporter.go:21-27, backend/cmd/lstep-migrate/migrator.go:173 |

#### Summary

The operator CSV report writes stored owner names and error messages directly, allowing leading spreadsheet formula markers to be interpreted when the report is opened.

#### Root Cause

The report writer relies on CSV quoting but does not apply the leading-character neutralization required for spreadsheet consumers.

**Unsanitized report cells** — `backend/cmd/lstep-migrate/reporter.go:19-27`

OwnerName and ErrorMessage reach CSV cells without formula neutralization.

```go
row := []string{fmt.Sprintf("%d", r.OwnerID), r.OwnerName, r.Status, fmt.Sprintf("%d", r.TagsAdded), fmt.Sprintf("%d", r.TagsFailed), r.ErrorMessage}; cw.Write(row)
```

#### Validation

Source-to-sink trace is direct and a neighboring CSV exporter implements the missing control. Validation details were not recorded separately.

Validation method: static source-to-sink review

#### Dataflow

The canonical finding records the affected path at backend/cmd/lstep-migrate/reporter.go:21-27, backend/cmd/lstep-migrate/migrator.go:173, but no expanded source-to-sink narrative was recorded.

#### Reachability

Reachability was not recorded beyond the canonical finding summary and affected locations.

#### Severity

**Low** — Stored data can trigger spreadsheet execution only when an operator runs the migration and opens its report.

Severity rises if the report is automatically distributed to broader recipients.

#### Remediation

Apply a shared CSV-cell sanitizer to every untrusted textual report field before calling csv.Writer.

Tests:
- Add a table-driven reporter test for values beginning =, +, -, and @ in both owner and error fields.

Preventive controls:
- Centralize spreadsheet-safe CSV serialization.

## Reviewed Surfaces

| Surface | Risk Area | Outcome | Notes |
| --- | --- | --- | --- |
| Go API, domains, persistence, and migrations | tenant isolation, lifecycle integrity, record and inventory writes | Reported | 1,913 in-scope backend files were assigned across the backend review streams; reportable inventory and seed-data issues survived. Evidence: artifacts/02_discovery/in_scope_files.txt, artifacts/02_discovery/candidate_ledger.jsonl, artifacts/03_coverage/reviewed_surfaces.md |
| Cloudflare Worker and scheduler | JWT, scheduled operations, Worker secrets | No issue found | 24 Worker files reviewed. Typecheck passed; test runner temporary output was sandbox-blocked. Evidence: artifacts/02_discovery/in_scope_files.txt, artifacts/03_coverage/reviewed_surfaces.md |
| React, LIFF, E2E, and frontend configuration | XSS, session state, redirects, uploads | No issue found | 1,707 frontend files reviewed. The unrestricted E2E target candidate was not attacker reachable. Evidence: artifacts/02_discovery/in_scope_files.txt, artifacts/02_discovery/candidate_ledger.jsonl, artifacts/03_coverage/reviewed_surfaces.md |
| CI, deployment, infrastructure, scripts, and root configuration | secret-bearing supply chain and deployment integrity | Reported | 444 files reviewed; three CI dependency-trust issues survived. Evidence: artifacts/02_discovery/in_scope_files.txt, artifacts/02_discovery/candidate_ledger.jsonl, artifacts/03_coverage/reviewed_surfaces.md |
| Static binary assets | active binary content | Not applicable | 45 PNG assets were inventoried and have no executable review surface. Evidence: artifacts/02_discovery/in_scope_files.txt, artifacts/03_coverage/reviewed_surfaces.md |
