# CI/CD パイプライン構成書

> **目的**: checked-in workflow の実行契約と、実環境で必要な承認・検証を区別する。
> **照合**: 2026-09-06、`7c6592f9f`。GitHub [#253](https://github.com/MinoruSoga/AnimalEkarte/issues/253) は OPEN。Issue 本文・2026-08-20 コメントは当時の判断であり、現在の workflow 実装や billing/reviewers の実測を代替しない。

## 1. 現行のデプロイ経路

| 経路 | トリガー・設定 | 承認境界 |
|---|---|---|
| STG backend | `staging` push の `backend/**`、backend workflow、root package/lockfile 変更。または manual dispatch | `backend-deploy.yml` に GitHub Environment job binding はない。共有環境への dispatch / branch 更新は承認済み operator が行う |
| STG frontend | `staging` push の `frontend/**` または frontend workflow 変更。または preview dispatch | job は `Preview` Environment に bind。外部 protection の現在値は別途確認 |
| Production backend | **未実装**。backend workflow は STG Worker URL / `wrangler.jsonc` 固定 | production trigger・config 選択・protected Environment binding の実装と検証が済むまで実行不可 |
| Production frontend | `production` push の対象 path 変更。または `environment=production` dispatch | job は **`Production`** Environment に bind。production dispatch は `refs/heads/production` 以外を拒否。Required reviewers と branch protection の現在値は外部確認が必要 |
| main push | CI。STG deploy workflow の直接 trigger ではない | review 済み `main -> staging` PR で昇格する |

正本は `.github/workflows/backend-deploy.yml` と `frontend-deploy.yml`。
[#253](https://github.com/MinoruSoga/AnimalEkarte/issues/253) 本文の「main push → STG」は delivery 方針であり、現行の直接 trigger は `staging`。日常開発 `main` と昇格 `main -> staging` のプロジェクト規約を使う。

2026-08-20 の「frontend に Environment gate 無し」は後続実装で解消している。一方、binding の存在だけでは Required reviewers が有効とは証明できない。**大文字小文字を含む名前一致・reviewers・対象 ref・secret scope を実行時に再確認する。**

本番構築は [setup.md](../infra/production/setup.md)、稼働後の契約は [production runbook](../infra/production/runbook.md)。backend/frontend 両方の acceptance が満たされるまで本番リリース成功としない。

## 2. Backend pipeline

1. Checkout、pnpm/Node setup、frozen lockfile install。
2. `CLOUDFLARE_API_TOKEN` の存在確認と `wrangler whoami`。
3. `backend/` から `npx wrangler deploy`。
4. `infra/scripts/cf-run-migrate.sh` で `POST /_internal/migrate`（`MIGRATE_RUN_SECRET`）。
5. `/health` が HTTP 200 / `status: ok` になるまで最大12回、30秒間隔で確認。
6. `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` がある場合だけ `cf-crud-smoke.sh`。**continue-on-error の optional step** なので workflow green だけでは CRUD PASS を証明しない。

順序は **deploy → migrate → health → optional smoke**。新 binary が旧 schema に到達し得る deploy 完了〜migration 完了（`MIGRATE_TIMEOUT=150s`）の区間は workflow コメントに記録された既知の制約。schema compatibility を release 前に確認する。

CSV seed は全環境で `002_master` のみ。STG は `APP_ENV=staging` を Worker/Container/migrate に渡し、フェーズ3で合成ログインを upsert する。詳細は [seed operations](SEED_MIGRATION_OPERATIONS.md)。health は process liveness であり DB access の証明ではない。

### 手動 dispatch

named owner/approval、review 済み commit、target Worker/config、secret scope、共有環境の利用可否を先に記録する。

```bash
REVIEWED_SHA='<reviewed-commit>'
TARGET_REF='staging'
REMOTE_SHA="$(gh api "repos/{owner}/{repo}/git/ref/heads/${TARGET_REF}" --jq '.object.sha')"
test "$REMOTE_SHA" = "$(git rev-parse "${REVIEWED_SHA}^{commit}")" || exit 1
gh workflow run backend-deploy.yml --ref "$TARGET_REF"
```

dispatch 後も run の `headSha == REVIEWED_SHA` を確認する。不一致、migration/health failure、target 不一致は停止条件。production branch で現行 backend workflow を dispatch しても production config の選択にはならない。

## 3. Frontend pipeline

1. GitHub Environment `Preview` / `Production` を選び、production ref を検証。
2. Vercel CLI `pull` で対象 environment の project 設定を取得。
3. `VERCEL_ENV` / `VITE_VERCEL_ENV` を付けて `pnpm --dir frontend build`。
4. `.vercel/output` を生成し `vercel deploy --prebuilt`。preview は STG domain へ alias。

`frontend/vite.config.ts` が `VERCEL_ENV=preview` なら STG API、`production` なら production API の絶対 URL をビルド時に固定する。`frontend/.env.production` の STG 値はこの経路では override される。prebuilt config に `/api` rewrite はないため、same-origin `/api` を使う別の build path では API JSON/status を別途検証する。

実デプロイの SHA、API 接続先、cookie/CORS、assets の検証は [Vercel runbook](VERCEL-FRONTEND-STAGING-TEST.md)。設定を読んだだけで稼働済みにしない。

## 4. Rollback / monitoring / backup

- rollback は last-known-good Cloudflare artifact/ref と migration 互換性を確認して行う。[#99](https://github.com/MinoruSoga/AnimalEkarte/issues/99) は旧 ECS 経路不在の確認として CLOSED。AWS/ECS は切り戻し先ではない。
- DB 非互換の場合は forward fix または承認済み restore plan。復旧手順は [STG](../infra/staging/runbook.md) / [PROD](../infra/production/runbook.md)。
- health、5xx、Workers/Container logs、Actions failure、Vercel deployment を監視する。Cloudflare Notification Policy の存在、通知先、実配送は外部検証が必要。Terraform tombstone を有効な通知ポリシーとして数えない。
- backup は owner、target、取得方式、保護された保存先、取得時刻、サイズ、checksum、retention、receipt、隔離 restore rehearsal を記録する。[production runbook §4](../infra/production/runbook.md#4-backuprestore-rehearsal) が正本。RPO/RTO は承認済み目標と rehearsal 実測を用い、推測しない。
- secret/PHI を log、artifact、Issue に出さない。資格情報の投入・rotation は [外部資格情報 runbook](runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) の USER 作業。

## 5. #253 の未検証境界

#253 は latest main required CI、STG deploy/health/failure notification、production reviewers、rollback時間、隔離 restore の実証を要求する。2026-07 の billing failure や 2026-08-20 の reviewers 空という記録は履歴であり、現在の failure と断定しない。

実行時に Actions run URL/ID・headSha・required jobs、billing、Environment protection、secret scope、backup/restore を確認する。未確認は **UNKNOWN / HOLD**。本書の静的同期は課金復旧、release acceptance、Issue close を意味しない。
