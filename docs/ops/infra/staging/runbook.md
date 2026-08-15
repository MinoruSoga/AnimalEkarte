# STG 運用 Runbook（Cloudflare）

> **目的**: STG の日常運用と障害初動。**読者**: 運用担当・開発者。手順の詳細は各正本へのポインタで示す（二重管理しない）。

## デプロイ

- **自動**: `staging` ブランチへ push（`backend/**` 変更時）→ `backend-deploy.yml`（deploy → /health → migrate → smoke）
- **手動**: `gh workflow run backend-deploy.yml --ref staging`
- main→staging は **merge commit** 方式の PR（squash はコミット履歴の祖先関係を切る — 運用注意）

## DB（PlanetScale）

- 接続調査は **TTL 付き診断ロール**で: `pscale role create animalekarte-stg main <name> --org noah-animalekarte --inherited-roles postgres --ttl 30m`（使い捨て・値は保存しない）
- migrate は CI が実行（`POST /_internal/migrate`・`MIGRATE_RUN_SECRET` 認証）。seed/checksum の扱い・フルデモ再投入は [deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](../../deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) が正本
- **資格情報ローテーション**: `pscale role reset-default` → GH Secrets（STG_DB_USER/PASSWORD）更新 → `worker-secret-sync.yml` 実行 → 再デプロイ
- ⚠️ 未解決: public スキーマ 109 オブジェクトの所有者が失効ロールのまま（PlanetScale サポートへ REASSIGN 依頼が唯一の解 — ALTER 系 migration の前に必須。詳細 = [_archive/migration-cloudflare.md](../_archive/migration-cloudflare.md) P7-2 観測 #2）

## 障害初動

1. `/health` 確認（workers.dev 直行と実 URL の両方 — 切り分けになる）
2. デプロイ直後なら**ローリング更新の旧イメージ残留**を疑う（15 分静置 → 再確認）
3. 全断+DB 接続エラーなら**接続スロット枯渇**を疑う（プール設定 `DB_MAX_OPEN_CONNS` と滞留接続 — 過去事例 = 観測 #2）
4. Cloudflare 側の障害確認: https://www.cloudflarestatus.com/
5. 切り戻し先は無い（AWS 廃止済み）。復旧は CF 側修正 or スナップショット+IaC 再建

## 検証コマンド

- スモーク: `STG_DEMO_EMAIL=... STG_DEMO_PASSWORD=... ./infra/scripts/cf-crud-smoke.sh`
- migrate 単発: `./infra/scripts/cf-run-migrate.sh`（`MIGRATE_RUN_SECRET` 必要）
