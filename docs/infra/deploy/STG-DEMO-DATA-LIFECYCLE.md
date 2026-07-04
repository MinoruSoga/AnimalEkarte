# STG デモデータライフサイクル (Demo Data Lifecycle)

> **Animal Ekarte**: ステージング環境におけるデモ・テストデータの作成から廃棄までの完全ガイド
> **最新更新**: 2026-05-26 | **ステータス**: STG-001 インシデント対応済

---

## 1. STG データの分類

ステージング環境に存在するデータは、以下の 4 つのカテゴリに分類されます。それぞれ異なるライフサイクルと管理ポリシーを持ちます。

| カテゴリ | 説明 | 作成時期 | 保有者 | 削除義務 |
|---------|------|--------|-------|--------|
| **Seed Data** | migration で自動生成。system_admin、base clinics、デフォルト permission_groups | DB 初期化時 | インフラ/DevOps | なし（永続） |
| **Demo Accounts** | 営業デモ、顧客向けプレゼンテーション用。固定アカウント数、予め作成済 | 環境構築時 | PO/営業チーム | なし（永続） |
| **Smoke Test Data** | デプロイ直後の機能検証用。CRUD-SMOKE-TEST.md に従い作成・削除 | デプロイ直後 | DevOps/エンジニア | **あり（デプロイ直後に全削除）** |
| **Investigation Data** | バグ調査、仕様検証用の一時データ。調査完了後は廃棄対象 | 随時（必要時） | エンジニア | あり（調査終了時） |

---

## 2. それぞれの目的

### 2.1 Seed Data
- **目的**: DB 初期化時の必須マスタデータ
- **具体例**: 
  - システム管理者グループ（permission_group id=1）
  - 初期医院レコード（clinic id=1）
  - デフォルト権限マスタ（resource + action 組み合わせ）
- **生存期間**: 環境存在期間中は永続
- **管理**: migrations で定義、DB_RESET で復元

### 2.2 Demo Accounts
- **目的**: 営業・顧客向けデモ環境でのロールプレイ
- **具体例**: 
  - demo-admin@example.com（システム管理者権限）
  - demo-staff@example.com（医院スタッフ権限、clinic_id 限定）
  - demo-owner@example.com（飼い主アカウント）
- **生存期間**: 環境存在期間中は永続
- **管理**: seed data として migration 定義、または デプロイ初期化スクリプトで作成
- **重要**: 本番移行時に全削除。ステージング環境でのみ有効

### 2.3 Smoke Test Data
- **目的**: デプロイ直後の機能確認（CRUD 操作、外部キー保護、権限チェック）
- **具体例**: 
  - テスト医院（TEST-CLINIC-xxxxxx）
  - テストスタッフ（TEST-STAFF-xxxxxx）
  - テスト権限グループ（TEST-GROUP-xxxxxx）
- **生存期間**: **デプロイ直後のみ。スモークテスト完了直後に削除必須**
- **管理**: [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) に従う
- **リスク**: 残置すると UI に表示（ユーザー混乱）、FK 制約違反の原因に
詳細は [Delete / Soft Delete 設計パターン](../../DELETE_SOFT_DELETE_PATTERNS.md) で hard delete と soft delete の使い分け、FK 制約の注意点、STG-001 事例を参照してください。

### 2.4 Investigation Data
- **目的**: バグ再現、仕様検証、パフォーマンス調査
- **具体例**: 
  - 特定の condition を満たす飼い主の作成（LTV 計算検証等）
  - エラーログ再現用の特殊レコード
  - 監査ログ動作確認用の一時データ
- **生存期間**: 調査完了まで。終了後は翌 DB_RESET で廃棄
- **管理**: 調査用途、作成者、削除予定日を docs/tasks に記録
- **重要**: 調査完了後は即削除推奨（DB_RESET 待ちは放置リスク高）

---

## 3. 作成元

STG データは以下の 4 つの方法で作成されます。それぞれの作成元ごとに、削除責任者が異なります。

### 3.1 Migration（自動）
- **対象**: seed data のみ
- **ファイル**: `backend/migrations/` の .sql ファイル
- **削除方法**: 不可。migration は immutable。削除は新規 migration で実装
- **例**: `migration_001_initial_setup.sql` で system_admin group 作成

### 3.2 API 経由（推奨）
- **対象**: demo accounts, smoke test data, investigation data
- **メソッド**: POST / PATCH / DELETE の REST エンドポイント
- **削除方法**: API DELETE エンドポイント を呼び出し（audit_log 自動記録）
- **例**: `curl -X DELETE /api/v1/masters/staffs/{id}` で テストスタッフ削除

### 3.3 Frontend 経由
- **対象**: smoke test data, investigation data（ユーザー操作経由）
- **削除方法**: 
  1. Frontend UI で削除操作、または
  2. API DELETE を curl で実行（§4 に従う）
- **例**: ブラウザで clinic を作成 → smoke test 後に API DELETE

### 3.4 Database 直接操作（例外のみ）
- **対象**: API が 5xx エラーで応答不可な緊急時のみ
- **実行要件**:
  1. Team lead の書面承認を取得
  2. 削除内容・理由・実行時刻を記録
  3. トランザクション内で実行（ロールバック可能）
  4. 削除後、API 層で 404 を確認
- **例**: `DELETE FROM staffs WHERE id = 'test-staff-xxx';` （禁止パターン）
- **重要**: この方法は監査ログを自動記録しないため、手動で audit_logs 行を追記

---

## 4. Cleanup 方針

### 4.1 デプロイ直後の Smoke Test Data（必須削除）

**タイミング**: CRUD-SMOKE-TEST.md 完全実行直後

**方法**（優先順）:

| # | 方法 | 条件 | audit_log |
|---|------|------|-----------|
| **1** | API DELETE エンドポイント | API 正常応答（200/201/204） | ✅ 自動記録 |
| **2** | curl DELETE（jq で 404 確認） | API 部分障害（一部 5xx） | ✅ 手動 curl で記録 |
| **3** | DB 直接削除 | API 完全障害（全 5xx） | ❌ 手動記録必須 |

**実行フロー**:

```bash
# Step 1: 作成したテストレコード一覧を記録
TEST_CLINIC_ID="<clinic_id>"
TEST_STAFF_ID="<staff_id>"
TEST_GROUP_ID="<permission_group_id>"

# Step 2: 関連レコード依存度を確認
curl -s "${API_V1}/clinics/${TEST_CLINIC_ID}" -b "${TOKEN}" | jq '.related_count'

# Step 3: スタッフ削除（関連レコード無し確認後）
curl -X DELETE "${API_V1}/masters/staffs/${TEST_STAFF_ID}" \
  -b "access_token=${TOKEN}"
# 期待: 204 No Content

# Step 4: 医院削除（関連スタッフ削除後）
curl -X DELETE "${API_V1}/clinics/${TEST_CLINIC_ID}" \
  -b "access_token=${TOKEN}"
# 期待: 204 No Content

# Step 5: 権限グループ削除（削除不可グループ以外）
curl -X DELETE "${API_V1}/masters/permission-groups/${TEST_GROUP_ID}" \
  -b "access_token=${TOKEN}"
# 期待: 204 No Content

# Step 6: 削除完了確認
curl -s "${API_V1}/clinics/${TEST_CLINIC_ID}" \
  -b "access_token=${TOKEN}" | jq '.error'
# 期待: 404 Not Found (または error フィールド)
```

**削除順序の重要性**:
1. 関連レコード（スタッフ）を先に削除
2. FK を持つ親レコード（医院）を後に削除
3. 権限グループは最後（FK 制約少ない）

### 4.2 削除完了記録

削除後、以下を CloudWatch ログまたは audit_logs テーブルで確認：

```bash
# 監査ログ確認
aws logs tail /ecs/animalekarte-stg --follow --region us-east-1 \
  | grep -E "DELETE|SOFT_DELETE" | tail -20

# 例: 期待ログ
# 2026-05-26T16:30:45.123Z staff_id=admin resource=staff resource_id=test-staff-123 action=DELETE
```

記録内容（docs/tasks/closed/{ISSUE}.md または CloudWatch):
- 削除日時（JST）
- 削除した レコード種別 + ID + 件数
- 削除操作者（staff_id または CI/CD job）
- 削除方法（API / curl / direct DB）

---

## 5. 残置許容条件

以下の **4 つすべて** に該当する場合のみ、STG データの残置が許容されます。

| # | 条件 | 確認方法 | 必須 |
|---|------|--------|------|
| **1** | UI に表示されない | ブラウザで確認、list API で 404 | ✅ |
| **2** | マルチテナント隔離に影響しない | clinic_id scope 確認 | ✅ |
| **3** | 次の DB_RESET で自動消滅 | DB_RESET=true 環境変数未設定なら DB_RESET 発動日を確認 | ✅ |
| **4** | 残置理由が文書化されている | docs/tasks/{ISSUE}.md に記載 | ✅ |

**許容例**:
- 調査結果を docs/tasks に記録し、「2026-05-31 DB_RESET で廃棄」と明記
- 次の本番反映時に DB_RESET 発動予定の場合

**許容されない例**:
- 理由なく放置
- UI 上で visible なまま
- 複数の clinic に関連するデータが混在

---

## 6. 残置不可条件

以下の **いずれか 1 つ** に該当する場合、**即座に削除必須**。残置は運用リスク。

| # | 条件 | リスク | 対応 |
|---|------|--------|------|
| **1** | UI の list/detail に表示される | ユーザー（営業・QA）の混乱、誤操作 | 即削除（API DELETE） |
| **2** | デモ/スモークテスト動作を妨害 | 次デプロイの検証失敗、lead time 増加 | 即削除（API DELETE） |
| **3** | マルチテナント隔離を破綻 | 他 clinic のデータが混在、セキュリティリスク | 即削除（DB 確認後 DELETE） |
| **4** | スモークテスト再実行を阻止 | 同名 resouce が存在して POST 409 / 重複 error | 即削除（API DELETE） |

---

## 7. DB_RESET=true の扱い

### 7.1 DB_RESET とは

`DB_RESET=true` 環境変数を設定してアプリケーションを起動すると、**ステージング環境のデータベースが完全にリセット** されます。

**実行内容**:
1. 全テーブルのレコードを削除（seed data 含む）
2. 全 migration を再実行（001 → 最新）
3. seed data を migration から再生成
4. demo accounts を初期化スクリプトから再生成
5. アプリケーション起動完了

### 7.2 使用シーン

| シーン | 実行者 | 条件 |
|--------|--------|------|
| **開発環境ローカル化** | エンジニア | 個人的に schema をリセット、ローカル検証したい時 |
| **CI/CD パイプライン** | GitHub Actions | PR merge → staging deploy 時、自動実行（オプション） |
| **本番リリース前検証** | DevOps | staging → production 反映前、クリーンな状態を保証したい時 |
| **定期保守（計画停止）** | DevOps | 毎月末日 or 四半期ごと、garbage collection（データベース最適化） |

### 7.3 DB_RESET 発動の設定方法

#### ローカル開発（Docker）
```bash
# docker compose で環境変数を設定
export DB_RESET=true
docker compose up backend

# または docker run で直接
docker run -e DB_RESET=true backend:latest
```

#### GitHub Actions CI/CD
`.github/workflows/backend-deploy.yml` の `workflow_dispatch` 入力 `db_reset` で明示指定します
（push トリガー時は `.env.staging` の `DB_RESET` 値がそのまま使用されます）。

```bash
gh workflow run backend-deploy.yml --ref staging -f db_reset=true
```

### 7.4 DB_RESET 実行前のチェックリスト

**実行前**:
- [ ] 現在のデータベースに「保存すべきデータ」が存在しないことを確認
  - 進行中の バグ調査 investigation data を別環境に backup したか
  - demo account 設定変更（パスワード等）を記録したか
- [ ] 他のエンジニアが staging で作業中でないことを確認（Slack 確認）
- [ ] 本番への反映予定日時を確認（リセット予定時刻と重複しないか）
- [ ] CloudWatch ログを取得（変更前の state 記録）

**実行後**:
- [ ] API ヘルスチェック `/health` で 200 OK 確認
- [ ] migration logs を CloudWatch で確認（004 までの full execution）
- [ ] seed data（system_admin group, clinic id=1）が再生成されたことを確認
- [ ] demo accounts でログイン可能か確認
- [ ] スモークテストを再実行（一度のみ）

### 7.5 DB_RESET が実行されないケース

- `DB_RESET` 環境変数が設定されていない（デフォルト: false）
- `DB_RESET=false` を明示設定
- アプリケーション起動時に `schema_migrations` テーブルが既に存在

---

## 8. STG-001 からの教訓

### 8.1 インシデント概要

**日時**: 2026-05-15（本番反映前日）
**発見**: 本番反映前チェックリスト実行中に clinics/5, 6, 7 が UI に表示されたまま
**原因**: スモークテスト直後に削除スクリプトが実行されず、データが STG 環境に残置
**影響**: 本番 deploy 予定が 1 日遅延、QA チームによる最終検証が中断

### 8.2 根本原因分析

#### 原因 A: Smoke Test Data 削除スクリプトなし
- CRUD-SMOKE-TEST.md に削除手順は記載
- しかし **実際の削除スクリプト/自動化がない**
- デプロイ直後の manual 作業フロー未確立
- → **結果**: デプロイチェックリスト作成者が削除を「実施し忘れた」

#### 原因 B: Soft-delete による FK 制約
- `permission_groups` は soft delete（deleted_at カラム使用）
- スモークテスト用テスト権限グループを作成後、delete API で削除
- しかし soft delete のため deleted_at に timestamp → レコード DB 上に残存
- → 後続の hard delete（clinics, staffs）時に FK constraint で DELETE 拒否 (409)
- → **結果**: 依存データが削除できず、残置リスク増加

#### 原因 C: 削除順序の誤解
- 医院（clinics）を先に削除しようとした
- しかし医院に FK 依存するスタッフ・権限グループが存在
- FK 制約で DELETE 409 → cleanup 中断
- → **結果**: 医院・スタッフ・権限グループが混在して残置

### 8.3 PR #64 での修正

**修正内容**:
1. hard-delete cleanup 前に **soft-deleted permission_groups の hard delete** を先行実行
2. soft-deleted rows を Unscoped() で明示的に SELECT して DELETE
3. 削除順序を明記（権限 → スタッフ → 医院）
4. cleanup 完了後に API list で 0 件確認を追加

**対応コミット**:
- `1755c193` fix: hard-delete soft-deleted permission_groups before clinic hard-delete
- deletion order 明記、test coverage 追加

### 8.4 再発防止策

#### 戦略 1: Cleanup を CI/CD 自動化（推奨）
```yaml
# .github/workflows/stg-smoke-cleanup.yml
name: STG Smoke Test Cleanup
on:
  workflow_dispatch:  # Manual trigger
  schedule:
    - cron: '0 16 * * *'  # Daily 16:00 JST
jobs:
  cleanup:
    steps:
      - name: Execute Cleanup Script
        run: |
          docker compose exec backend \
            bash /scripts/stg-cleanup-smoke-test-data.sh
      - name: Verify
        run: |
          curl -s https://api.stg.noah-karte.com/health | jq '.status'
```

#### 戦略 2: Manual Checklist の強制
- CRUD-SMOKE-TEST.md の cleanup section を独立した checklist として独立
- デプロイ完了 → スモークテスト実行 → **cleanup 実行 → verification** の明示的 4 ステップ化
- cleanup skip は team lead approval 必須

#### 戦略 3: Data Classification と Retention Policy
- 本ドキュメント（§1〜6）で smoke test data = 削除必須 を明記
- デプロイ README.md に「cleanup 完了」を success criterion に追加
- 残置リスク 4 条件（§6）を monitoring dashboard に可視化

#### 戦略 4: Soft-delete 設計再検討
- `permission_groups` の soft delete は FK blocker を生む
- 代替案: permission_groups も hard delete に統一、audit_log で DELETE 操作記録
- または: soft-delete rows に TTL（Time-To-Live）を設定、自動 hard-delete

### 8.5 運用上の指針

**デプロイ直後のチェックリスト（§4.1）**:
1. ✅ API ヘルスチェック通過
2. ✅ CRUD スモークテスト実行完了
3. ✅ **Smoke test data 全削除完了** ← 重要
4. ✅ UI 再訪問、test レコード 0 件確認
5. ✅ 本番反映前チェックリスト開始

**Smoke test data 削除の必須化**:
- 削除は **「推奨」から「必須」** へ昇格
- 削除スキップは team lead 承認必須（書面記録）
- CI/CD で自動化されるまでは manual checklist 遵守

**残置データの audit**:
- 本番反映前日に STG UI で「demo/test レコード非表示」を確認
- SQL query で test-prefixed records を検索
  ```sql
  SELECT COUNT(*) FROM clinics WHERE name LIKE 'TEST-%';
  SELECT COUNT(*) FROM staffs WHERE name LIKE 'TEST-%';
  SELECT COUNT(*) FROM permission_groups WHERE name LIKE 'TEST-%' AND deleted_at IS NULL;
  ```
- 0 件以外は cleanup 不完了 → deploy 延期

---

## 参考資料

- **[デプロイメント運用](./README.md)**: ロールバック判定、ヘルスチェック、成功基準
- **[CRUD スモークテスト](./CRUD-SMOKE-TEST.md)**: テスト実行手順、ステータスコード期待値、cleanup 手順
- **[CI/CD パイプライン](./CI-CD-PIPELINE.md)**: デプロイ自動化フロー
- **[本番反映前チェック](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md)**: リリース前最終検証
- AWS RDS: `animalekarte-stg` PostgreSQL cluster（AWS console）
- CloudWatch Logs: `/ecs/animalekarte-stg` ロググループ
