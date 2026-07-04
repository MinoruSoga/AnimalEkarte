# リリース準備・検証ランブック (Release Readiness Check)

> **目的**: 本番反映前の最終検証ランブックを提供する。
> **読者**: DevOps/Team Lead。
> **タイミング**: 本番反映前の最終検証時。

> **Animal Ekarte**: 安全かつ確実なステージング反映と本番昇格のための手順
> **最新更新**: 2026-06-12 | **ステータス**: 推奨運用手順

---

## 1. 概要
本ランブックは、`main` ブランチの最新コードを `staging` 環境へ反映し、最終的なスモークテストを経て本番稼働へと昇格させるための、機械的かつ厳格なチェックリストです。

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

### 3.1 DB リセットとマイグレーション
大規模なスキーマ変更を伴う場合、初回デプロイのみ `DB_RESET=true` を指定して実行します。

- **実行コマンド (GitHub Actions)**:
  `gh workflow run backend-deploy.yml --ref staging -f db_reset=true`

- **監視項目**:
  - `aws logs tail` で `applying 001_init.sql` 〜 `004_seed_staging.sql` の成功を確認。
  - Checksum mismatch エラーが発生していないか。
- **シード突合検証**: `bash scripts/verify_seed_matches_stg_dump_full.sh` → exit 0 確認 (seed が STG dump と全テーブルで一致すること)。

---

## 4. フェーズ 3: 最終検証 (Post-Deploy Check)

### 4.1 ヘルスチェック
- [ ] `GET /health` -> `200 OK` (ALB 直接および CloudFront 経由)。
- [ ] ECS サービスのタスク実行数 (`runningCount`) が期待値と一致しているか。

### 4.2 スモークテスト (CRUD Smoke Test)
[スモークテスト手順書](../CRUD-SMOKE-TEST.md) に従い、以下の基本導線をブラウザで確認します。
- **医院管理**: 既存医院の編集・保存。
- **権限管理**: 権限グループの変更と、ログアウト/再ログイン後の反映。
- **スタッフ管理**: ログイン可否と、所属医院の切り替え。

---

## 5. ロールバック基準 (Rollback Criteria)

以下のいずれか 1 つでも該当した場合、**即座にロールバック**を実行してください。
1.  API エラーレートが **1%** を超過。
2.  画面の初期ロードに **5 秒** 以上を要する。
3.  スモークテスト中に 1 件でも致命的なバグ（500エラー等）を発見。

---

## 6. 結果の記録

デプロイ完了後、以下のテンプレートを用いて `ops_deploy_YYYYMMDD.md` として記録を残します。

```markdown
### リリース記録 (YYYY-MM-DD)
- **対象 Commit**: [sha]
- **スモークテスト結果**: PASS / FAIL
- **CloudWatch エラーログ**: 0 件
- **特記事項**: [なし/あり]
```

---
