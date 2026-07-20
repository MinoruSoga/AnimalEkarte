# Infra — Cloudflare (STG/PROD)

> AWS STG は 2026-07-20 に廃止済み(terraform destroy・課金停止)。経緯と実施記録は
> `../docs/ops/infra/_archive/migration-cloudflare.md`(Phase 8)と `docs/ops/infra/_archive/aws-legacy/` を参照。
> 再編計画は `infra-reorg-plan.md`(Phase B で envs/ + modules/ 構造へ移行予定)。

## 構成の2層境界(MANDATORY)

| 層 | 管理対象 | 場所 |
|---|---|---|
| Terraform | ゾーン・DNS・R2・Hyperdrive・通知 | `infra/cloudflare/`(現行=STG) / `infra/cloudflare/production/`(ドラフト) |
| Wrangler | Worker・Container・ルート・secrets | `backend/wrangler.jsonc`(STG) / `backend/wrangler.production.jsonc`(ドラフト) |

- Workers/Containers を Terraform で管理しない(デプロイパイプラインと競合する)
- secrets は `wrangler secret put` または GitHub Actions `worker-secret-sync.yml` 経由。
  `vars` は非機密のみ。Terraform リソース・state に secret を置かない
- 手動ダッシュボード操作は原則禁止。やむを得ない場合は実施記録を残し `cf-terraforming` で取り込む

## セキュリティ

- endpoint・role・bucket 名は operational-sensitive。公開ログ・Issue へ不用意に貼らない
- インフラ変更は plan → 差分確認 → 明示承認 → apply。destroy・credential 変更は必ず停止して承認を得る

## 運用スクリプト(`infra/scripts/`)

| script | 用途 |
|---|---|
| `cf-run-migrate.sh` | STG Worker 経由の DB migration 実行(CI からも使用) |
| `cf-crud-smoke.sh` | STG API の CRUD/混在会計スモーク |
| `pscale-create-stg.sh` | PlanetScale STG DB 作成手順 |
| `validate-schema.sql` | スキーマ互換性検証 |

## 参照ドキュメント

- デプロイハブ: `docs/ops/deploy/README.md`
- STG フルデモ投入/検証: `docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md`
- 本番構築手順: `../docs/ops/infra/production/setup.md`(#253)
