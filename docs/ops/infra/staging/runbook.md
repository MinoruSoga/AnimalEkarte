# STG 運用 Runbook（Cloudflare）

> External state is verification-required. この文書は HEAD の workflow/config contract のみを説明する。

## Deploy

- 自動: `staging` pushで`backend/**`、`.github/workflows/backend-deploy.yml`、root `package.json`、root `pnpm-lock.yaml`のいずれかが変わった場合。`infra/cloudflare/**`単独ではtriggerされず、Terraformは別の承認済みplan/apply手順で扱う。
- 順序: **deploy → migrate → `/health` → optional smoke**。
- 手動 dispatch も存在する。実行は人が対象 ref、approval、secret scope を確認して行う。

## DB and seed

- migrate は `POST /_internal/migrate` と `MIGRATE_RUN_SECRET` を使う workflow contract。
- 現行 `BundleOrderForEnv` は全環境で `002_master` のみ。`003_demo` / `004_staging` や full-demo 再投入を前提にしない。UAT/synthetic data は承認済みの明示 import と lifecycle owner を別途定義する。
- 過去の「public schema 109 objects」「REASSIGN が唯一解」は dated incident observation であり current fact ではない。schema owner と provider-supported remediation を PlanetScale/runtime で再検証してから ALTER migration を行う。
- credential rotation、DB access、shared STG operation は人による明示承認が必要。

## Incident triage

1. workers.dev path と configured domain の `/health` を比較する。
2. deploy直後の rolling update を考慮する。
3. DB connection error は pool/slot metrics を確認する。過去事例を current diagnosis とみなさない。
4. provider status を確認する。
5. AWS rollback target は repository decision 上存在しない。Cloudflare 側の修正または検証済み backup + IaC restore contract を使う。

runtime、DB 内容、credential、provider status は本更新では確認していない。
