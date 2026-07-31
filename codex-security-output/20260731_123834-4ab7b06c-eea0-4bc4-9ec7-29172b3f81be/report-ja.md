# Codex Security レポート（日本語要約）
対象リポジトリ: AnimalEkarte
対象パス: /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
スキャンID: 4ab7b06c-eea0-4bc4-9ec7-29172b3f81be
実行日時: 2026/7/31 12:38:33
進捗ステータス: complete
スコープ: .
モード: standard

## 要約
- 検出件数: **6件**
- 深刻度別: 低: 1件 / 中: 5件
- コスト（推定）: $16.077245
- カバレッジ: 0/4067（partial）

## 検出内容

### [1] Deployment workflows use mutable third-party Action references
- 重大度: 中
- 信頼度: 高
- CWE: CWE-829
- 影響: N/A

#### 概要
Backend and frontend deployment jobs execute checkout, pnpm setup, and Node setup through version tags rather than immutable commit SHAs before credentialed deployment commands.

#### 根本原因
Privileged workflows trust mutable external action labels rather than reviewed, immutable action revisions.

#### 推奨対応
Pin every third-party Action in privileged workflows to a full commit SHA and use an automated reviewed updater for SHA changes.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/backend-deploy.yml` (33-41)
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/frontend-deploy.yml` (39-45)

### [2] Fractional treatment quantities bypass inventory decrement
- 重大度: 中
- 信頼度: 高
- CWE: CWE-682, CWE-840
- 影響: N/A

#### 概要
Inventory-backed treatments accept fractional quantities, but stock decrement truncates the value to an integer before enforcing availability and updating stock.

#### 根本原因
The treatment quantity type and inventory quantity semantics are inconsistent, with loss of precision at the enforcement point.

#### 推奨対応
Either reject non-integral quantities for integer-stock items before creating the treatment, or store and atomically decrement inventory using a precise decimal representation.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/inventory/repository.go` (152-157)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_request.go` (14-14)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/treatment_service.go` (337-340)

### [3] Tracked demo seed exposes privileged-account verifiers and staff identity data
- 重大度: 中
- 信頼度: 高
- CWE: CWE-359, CWE-522
- 影響: N/A

#### 概要
The local/test demo seed stores active system-admin account rows, staff email identities, and bcrypt password verifiers in the repository despite the repository policy prohibiting that material in seeds or Git history.

#### 根本原因
Privileged local/test identities and password verifiers are committed rather than provisioned through a non-versioned local or secret-managed path.

#### 推奨対応
Replace tracked identities and verifiers with synthetic disabled fixtures, and provision any necessary local accounts through a secret-managed one-time path.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/migrations/seeds/003_demo/accounts.csv` (2-2)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/migrate/main.go` (490-492)
- 証跡: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/migrations/CLAUDE.md` (69-69)

### [4] Performance workflow trusts an unpinned k6 installation path before credentialed execution
- 重大度: 中
- 信頼度: 高
- CWE: CWE-829, CWE-494
- 影響: N/A

#### 概要
The performance workflow imports a remote APT key through apt-key, installs an unversioned k6 package, and later runs it with staging demo credentials.

#### 根本原因
The workflow extends its package trust root dynamically without binding it to an expected key or artifact.

#### 推奨対応
Use a version-and-digest-pinned k6 artifact or image, or verify a known signing-key fingerprint and exact package version before installation.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/performance-tests.yml` (52-52)
- シンク: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/performance-tests.yml` (90-108)

### [5] Secret-sync workflow executes an unverified npm artifact
- 重大度: 中
- 信頼度: 高
- CWE: CWE-829, CWE-494
- 影響: N/A

#### 概要
The staging Worker secret-sync job downloads and executes Wrangler through npx in the same step that exposes Cloudflare and database secrets, without a lockfile or artifact-digest integrity binding.

#### 根本原因
A secret-bearing workflow resolves executable package code outside the repository's lockfile and without artifact verification.

#### 推奨対応
Install Wrangler through a frozen lockfile and invoke the resolved local binary, or verify a downloaded artifact against a maintained digest before exposing secrets.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.github/workflows/worker-secret-sync.yml` (25-38)

### [6] LSTEP migration report permits spreadsheet-formula injection
- 重大度: 低
- 信頼度: 高
- CWE: CWE-1236
- 影響: N/A

#### 概要
The operator CSV report writes stored owner names and error messages directly, allowing leading spreadsheet formula markers to be interpreted when the report is opened.

#### 根本原因
The report writer relies on CSV quoting but does not apply the leading-character neutralization required for spreadsheet consumers.

#### 推奨対応
Apply a shared CSV-cell sanitizer to every untrusted textual report field before calling csv.Writer.

#### 対象箇所
- 制御点: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/lstep-migrate/reporter.go` (21-27)
- ソース: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/cmd/lstep-migrate/migrator.go` (173-173)

## 補足（原文レポート）
- 英語版: `report.md`
- Scan show JSON: `scan-show.json`
