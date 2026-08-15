# リリース準備・検証ランブック (Release Readiness Check)

> **目的**: 本番反映前の最終検証ランブックを提供する。
> **読者**: DevOps/Team Lead。
> **タイミング**: 本番反映前の最終検証時。

> **Animal Ekarte**: 安全かつ確実なステージング反映と本番昇格のための手順
> **最新更新**: 2026-07-23 | **ステータス**: STG 推奨運用手順（Production 未構築）

---

## 1. 概要
本ランブックは、`main` ブランチの最新コードを `staging` 環境へ反映し、最終的なスモークテストを行うための、機械的かつ厳格なチェックリストです。本番環境は未構築のため、本チェックの PASS だけで本番昇格可能とは判定しません。

---

## 2. フェーズ 1: 事前監査 (Audit)

デプロイを開始する前に、環境設定と CI の状態を確定させます。

### 2.1 環境変数（Secret）の監査
- [ ] `.env.staging` 内に平文のパスワードや秘密鍵が混入していないか。
- [ ] `JWT_SECRET`, `INTEGRATION_ENCRYPTION_KEY` 等の必須変数が定義されているか。
- [ ] `CORS_ALLOWED_ORIGIN` に正しいフロントエンド URL が設定されているか。

### 2.2 CI パイプラインの検証
- [ ] `ci.yml` ワークフローが `success` であること。
- [ ] **注意**: `skipped` になったジョブ（バックエンド/フロントエンド等）がないか、`paths-filter` の挙動を再確認する。

---

## 3. フェーズ 2: デプロイ実行 (Deploy)

### 3.1 マイグレーションと、必要時のみ行う DB 再作成

現行の `backend-deploy.yml` は `workflow_dispatch` 可能ですが、`db_reset` 入力はありません。
通常は次の手動トリガー、または `staging` push に伴う自動実行で `wrangler deploy` → migrate → health を通します。

```bash
gh workflow run backend-deploy.yml --ref staging
```

共有 STG の DB 再作成は workflow のオプションではなく、データを破棄する別オペレーションです。
必要な場合だけ明示承認を得て [STG_PLANETSCALE_SEED_RUNBOOK.md](../STG_PLANETSCALE_SEED_RUNBOOK.md)
に従います。AWS ECS/RDS は廃止済みで、DB 再作成や切り戻しには使用できません。

- **監視項目**:
  - Worker の migrate レスポンス（`POST /_internal/migrate` の exit code）と `/health` ポーリング結果を確認。
  - fresh DB では直下 DDL 全数（`ls backend/migrations/*.sql` を正とする）+ seed バンドル（`002_master` → `003_demo` → `004_staging`）が完了し、`schema_migrations` の行数が「直下 DDL 本数 + seed バンドル数」に一致することを確認。
  - Checksum mismatch がないことを確認。旧 seed key の検出時は旧形式相当の 002〜004 が legacy translation されることを確認する。
- **シード突合検証**: `bash scripts/verify_seed_matches_stg_dump_full.sh` → exit 0 確認 (seed が STG dump と全テーブルで一致すること)。

---

## 4. フェーズ 3: 最終検証 (Post-Deploy Check)

### 4.1 ヘルスチェック
- [ ] `GET /health` -> `200 OK`（実 URL と workers.dev の両経路で確認し、DNS/Worker/Container のどこで失敗しているか切り分ける）。
- [ ] デプロイ直後は旧コンテナイメージが残りうるため、イメージ更新を伴う検証では 15 分静置後にも再確認する。

### 4.2 スモークテスト (CRUD Smoke Test)
[スモークテスト手順書](../CRUD-SMOKE-TEST.md) に従い、以下の基本導線をブラウザで確認します。
- **医院管理**: 既存医院の編集・保存。
- **権限管理**: 権限グループの変更と、ログアウト/再ログイン後の反映。
- **スタッフ管理**: ログイン可否と、所属医院の切り替え。

---

## 5. リリース失敗・復旧開始基準

以下のいずれか 1 つでも該当した場合、リリース成功とはせず、即座に復旧を開始してください。
1.  API エラーレートが **1%** を超過。
2.  画面の初期ロードに **5 秒** 以上を要する。
3.  スモークテスト中に 1 件でも致命的なバグ（500エラー等）を発見。

> AWS への切り戻しやホットスタンバイは存在しません。ここでいう復旧はリリース中止・
> Cloudflare 側の修正または再デプロイ・必要に応じたスナップショット + 現行 IaC からの再建を指します。
> 初動は [`../../infra/staging/runbook.md`](../../infra/staging/runbook.md) に従います。

---

## 6. 結果の記録

デプロイ完了後、以下のテンプレートを用いて `ops_deploy_YYYYMMDD.md` として記録を残します。

```markdown
### リリース記録 (YYYY-MM-DD)
- **対象 Commit**: [sha]
- **スモークテスト結果**: PASS / FAIL
- **エラーログ**: 0 件（Cloudflare Workers Logs / Containers）
- **特記事項**: [なし/あり]
```

---
