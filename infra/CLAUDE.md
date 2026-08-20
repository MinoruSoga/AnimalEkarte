# Infra — Cloudflare (STG/PROD)

> AWS STG は 2026-07-20 に廃止済み（terraform destroy・課金停止）。現行構成は
> `../docs/ops/infra/architecture.md`。移行実施記録と旧 AWS 文書は git 履歴。
> 再編計画は `../docs/ops/infra/reorg-plan.md`（Phase Bでenvs/ + modules/構造へ移行予定）。

## 構成の2層境界（MANDATORY）

| 層 | 管理対象 | 場所 |
|---|---|---|
| Terraform | ゾーン・DNS・R2・Hyperdrive・通知 | `infra/cloudflare/`（現行=STG）/ `infra/cloudflare/production/`（ドラフト） |
| Wrangler | Worker・Container・ルート・secrets | `backend/wrangler.jsonc`（STG）/ `backend/wrangler.production.jsonc`（ドラフト） |

- Workers/ContainersをTerraformで管理しない（デプロイパイプラインと競合する）
- secretsは`wrangler secret put`またはGitHub Actions `worker-secret-sync.yml`経由で管理する
- `vars`は非機密のみ。Terraformリソース・stateにsecretを置かない
- 手動ダッシュボード操作は原則禁止。やむを得ない場合は実施記録を残し`cf-terraforming`で取り込む

## セキュリティ

- endpoint・role・bucket名はoperational-sensitive。公開ログ・Issueへ不用意に貼らない
- インフラ変更はplan → 差分確認 → 明示承認 → applyの順で行う
- destroy・credential変更は必ず停止して承認を得る

## 運用スクリプト（`infra/scripts/`）

| script | 用途 |
|---|---|
| `cf-run-migrate.sh` | STG Worker経由のDB migration実行（CIからも使用） |
| `cf-crud-smoke.sh` | STG APIのCRUD/混在会計スモーク |
| `pscale-create-stg.sh` | PlanetScale STG DB作成手順 |
| `validate-schema.sql` | スキーマ互換性検証 |

## 参照ドキュメント

- デプロイハブ: `../docs/ops/deploy/README.md`
- STGフルデモ投入/検証: `../docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md`
- 本番構築手順: `../docs/ops/infra/production/setup.md`（#253）
