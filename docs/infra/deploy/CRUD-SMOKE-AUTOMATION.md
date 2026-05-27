# STG CRUD スモーク自動化戦略 (Automation Strategy)

> **Animal Ekarte**: デプロイ後の CRUD スモークテスト自動化スコープ・実装方針・GitHub Actions ロードマップ
> **最新更新**: 2026-05-27 | **目的**: 本番リリース候補判定の全自動化ビジョン

---

## 1. 目的と現状

本ドキュメントは、[CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) で定義された 11 項目のスモークテスト（A-1 ～ C-4）を段階的に自動化するための戦略を記載します。

**現状**:
- CRUD テスト: 手動実行（curl コマンド）
- デモログイン: 手動検証（UI）
- テストデータ削除: 手動実行（cleanup curl）
- CI/CD: 自動デプロイのみ実装済み

**目標**:
- デプロイ直後の自動スモークテスト実行
- 失敗時の自動ロールバック判定
- 削除/復旧ログの自動監査

---

## 2. 自動化対象（IN SCOPE）

### 2.1 Health Check Polling

```bash
# エンドポイント疎通確認（1 回または loop）
curl -s https://api.stg.noah-karte.com/health | jq '.status'
# 期待: "ok"
```

**自動化方針**:
- デプロイ完了直後に即座に実行（最大 30 秒）
- 失敗時はロールバック判定フロー（RFC-001 § Rollback Decision Tree へ）
- 成功時はフェーズ 2 へ進行

**GitHub Actions Phase 1 実装済み**:
- [x] `.github/workflows/stg-health-check.yml` 作成（bash script via curl）
- [x] workflow_dispatch 手動トリガー実装完了
- [x] jq による JSON 安全パース + HTTP 200 & `.status=="ok"` 二段階検証

**実装詳細**:
- **ワークフロー**: `.github/workflows/stg-health-check.yml` 作成済み
- **トリガー**: `workflow_dispatch` により手動実行のみ（自動実行なし）
- **検証ロジック**: `curl -sS -o "$body_file" -w "%{http_code}"` で HTTP ステータスと応答本体を分離；`jq -r '.status // empty'` で JSON を安全にパース；HTTP 200 かつ `.status=="ok"` を確認
- **実行コマンド**: `gh workflow run stg-health-check.yml`（ローカル実行）又は GitHub Web UI より手動トリガー
- **Phase 1 スコープ**: エンドポイント疎通確認のみ。ログイン・CRUD・DB・テストデータ削除は Phase 2 以降

**参考**:
- [README.md § 4.1](./README.md#41-ヘルスチェック手順) に自動化コマンド例記載
- Phase 2（Demo Account Login API）実装予定は CRUD-SMOKE-AUTOMATION.md § 2.2 参照

---

### 2.2 Demo Account Login (API)

**認証フロー** (API 経由):
```bash
# 手順 1: SSM Parameter Store からcredentials 取得
export DEMO_EMAIL=$(aws ssm get-parameter --name /stg/demo/email --query 'Parameter.Value' --output text)
export DEMO_PASSWORD=$(aws ssm get-parameter --name /stg/demo/password --with-decryption --query 'Parameter.Value' --output text)

# 手順 2: ログイン API 実行
RESPONSE=$(curl -s -X POST "https://api.stg.noah-karte.com/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${DEMO_EMAIL}\",\"password\":\"${DEMO_PASSWORD}\"}")

# 手順 3: Token 抽出
TOKEN=$(echo ${RESPONSE} | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo ${RESPONSE} | jq -r '.data.refresh_token')
```

**期待結果**:
- HTTP 200 OK
- `.data.access_token` 存在
- `.data.refresh_token` 存在

**自動化方針**:
- Health check PASS 後、直ちに実行
- SSM Parameter Store から credentials 取得（ハードコード禁止）
- Token 抽出に成功したら環境変数へ設定（後続 CRUD テストで使用）
- 失敗時はロールバック判定

**GitHub Actions TODO**:
- [ ] `.github/workflows/stg-login-test.yml` 作成
- [ ] `aws ssm get-parameter` コマンド統合
- [ ] output: `access_token`, `refresh_token` （secrets として設定）
- [ ] error handling: `jq -e` で null チェック

---

### 2.3 CRUD フル実行 (Clinics / Permission Groups / Staffs)

**テスト項目**: [CRUD-SMOKE-TEST.md § 3](./CRUD-SMOKE-TEST.md#3-テスト項目と実行例) 参照

- **A-1**: Clinics 一覧取得
- **A-2**: 権限制限チェック（403）
- **A-3**: 医院編集・保存
- **A-4**: マルチテナント検証
- **B-1**: 新規グループ作成
- **B-2**: グループ削除保護（409）
- **B-3**: テスト用グループ削除
- **C-1**: スタッフ新規登録
- **C-2**: ログイン検証
- **C-3**: 削除保護（FK 検証）
- **C-4**: 削除可能スタッフ削除

**自動化方針**:
- 全 11 項目を curl スクリプトで自動実行
- 各項目の HTTP ステータスコードを期待値と比較
- 不一致時は即座にロールバック判定
- テストデータ生成・削除も含める（cleanup 段階で一括削除）

**GitHub Actions TODO**:
- [ ] `.github/workflows/stg-crud-smoke-test.yml` 作成
- [ ] bash 配列で 11 テストケースを定義
- [ ] 各テストの status code 検証ロジック実装
- [ ] 失敗サマリー出力（どのテストが失敗したか明記）
- [ ] timeout: 120s（各テスト）

---

### 2.4 テストデータ削除（Cleanup）

**削除対象**: CRUD テストで生成したテンポラリデータ

**削除順序** (FK 順):
1. Staffs （最後に削除）
2. Permission Groups
3. Clinics （最後に作成したものから逆順）

**自動化方針**:
- CRUD テスト成功後、即座に DELETE API を実行
- 404 or 204 を期待値として検証
- 削除ログを CloudWatch に出力
- 削除失敗時は**手動ロールバック対象** (自動化は削除成功時のみ)

**GitHub Actions TODO**:
- [ ] cleanup job を CRUD テスト job の直後に配置
- [ ] dependsOn: crud-test （失敗時は cleanup 実行しない）
- [ ] DELETE 実行 + 監査ログ確認（CloudWatch Logs Insights query）

---

## 3. 非自動化項目（OUT OF SCOPE）

以下は手動検証が必要であり、自動化対象外です。

### 3.1 フロントエンド UI 表示確認

**理由**: Visual regression detection は複雑で、E2E フレームワーク（Playwright/Cypress）が必要。

**確認内容**:
- ページ読み込み（5 秒以内）
- ログイン画面表示、CSS 適用
- Settings 画面アクセス・フォーム操作

**実装予定**: 将来的に Playwright E2E スイート導入を検討（FUTURE-001）

---

### 3.2 パスワード入力フィールド（セキュリティ）

**理由**: テスト時のパスワード暗号化・安全な削除が複雑。

**実装方針**:
- パスワードは SSM Parameter Store に格納（ハードコード禁止）
- curl 実行時のみメモリに展開（スクリプト終了時にクリア）
- ログ出力にパスワードが含まれないこと確認

---

### 3.3 直接 DB 操作（テストデータ削除）

**理由**: API 経由の削除が標準。DB 直接操作は例外的で、チーム lead 承認必須。

**実装方針**:
- API DELETE コマンドで削除（FK 制約検証も含む）
- API 失敗時のみ、手動で DB 操作を検討（MANUAL-ROLLBACK-001）

---

### 3.4 CloudWatch ログ監査

**理由**: ログ分析は複雑なクエリが必要。AWS CloudWatch Logs Insights は非決定的。

**実装方針**:
- ERROR/FATAL ログの個数カウントのみ自動化（閾値: 5 分間 3 件以下）
- 詳細な error pattern は手動で確認

**GitHub Actions TODO**:
- [ ] `aws logs start-query` で過去 5 分の ERROR/FATAL をカウント
- [ ] 結果が 3 以下なら PASS、以上なら FAIL

---

## 4. 実行タイミング

### 4.1 Option A: デプロイ直後・自動実行（推奨）

**トリガー**: GitHub Actions workflow による自動呼び出し

```yaml
# .github/workflows/stg-deploy.yml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to STG
        run: make deploy-stg
      - name: Wait for Vercel
        run: sleep 30
      - name: Trigger smoke test
        run: gh workflow run stg-smoke-test.yml --ref main
```

**メリット**:
- デプロイ直後に即座に検証（時間差リスク最小化）
- 失敗時は自動でロールバック判定

**デメリット**:
- CI ジョブ数増加（コスト）
- 依存関係が増えると複雑化

---

### 4.2 Option B: 手動トリガー（現在）

**トリガー**: `workflow_dispatch` で明示的に実行

```bash
gh workflow run stg-smoke-test.yml
```

**メリット**:
- 運用時に柔軟（デバッグ・再実行が容易）
- 不要な自動実行を避けられる

**デメリット**:
- 手動実行の取り忘れ
- タイミング遅延（デプロイ直後でない可能性）

**推奨**: 当面は Option B（手動） → 運用が安定したら Option A（自動）へ移行

---

### 4.3 Option C: 定期実行（将来検討）

**トリガー**: schedule で毎日実行（例: 日次スモークテスト）

```yaml
on:
  schedule:
    - cron: '0 9 * * *'  # 毎日 09:00 JST
```

**用途**: STG 環境の定期ヘルスチェック（デプロイがない日も）

**実装予定**: FUTURE-002

---

## 5. Secrets 管理

### 5.1 必須 Secrets

GitHub Actions Secrets に以下を登録：

| Secret 名 | 値 | 管理元 |
|-----------|-----|-------|
| `STG_DEMO_EMAIL` | demo account email | SSM Parameter Store |
| `STG_DEMO_PASSWORD` | demo account password | SSM Parameter Store (encrypted) |
| `STG_API_BASE_URL` | https://api.stg.noah-karte.com | Repository environment (public) |
| `AWS_ROLE_ARN` | OIDC role for GitHub Actions | AWS IAM (Team Lead 設定) |

### 5.2 実装方針

**Option A: GitHub Secrets（推奨）**
- SSM Parameter Store から CI 起動時に取得
- GitHub Secrets に cache（セッション内のみ）
- ログ出力時にマスク（自動）

```yaml
env:
  DEMO_EMAIL: ${{ secrets.STG_DEMO_EMAIL }}
  DEMO_PASSWORD: ${{ secrets.STG_DEMO_PASSWORD }}
```

**Option B: SSM Parameter Store 直接呼び出し**
- `aws ssm get-parameter` で実行時取得
- ネットワーク遅延あり（最大 5s）

```bash
export DEMO_EMAIL=$(aws ssm get-parameter --name /stg/demo/email --query 'Parameter.Value' --output text)
```

**推奨**: Option A（GitHub Secrets + SSM 初期同期）

---

## 6. 失敗判定基準

自動テストが FAIL と判定される条件：

| # | テスト | 失敗条件 | ロールバック判定 |
|---|--------|----------|-----------------|
| 1 | Health Check | HTTP 非 200 | **即ロールバック** |
| 2 | Demo Login | 401 / Token 未取得 | **即ロールバック** |
| 3 | CRUD A-1 | HTTP 非 200 | **即ロールバック** |
| 4 | CRUD A-2 | HTTP 非 403 | **即ロールバック** |
| 5 | CRUD A-3 | HTTP 非 200 | **即ロールバック** |
| 6 | CRUD A-4 | 権限隔離失敗 | **即ロールバック** |
| 7 | CRUD B-1 | HTTP 非 201 | **即ロールバック** |
| 8 | CRUD B-2 | HTTP 非 409 | **即ロールバック** |
| 9 | CRUD B-3 | HTTP 非 204 | **即ロールバック** |
| 10 | CRUD C-1 | HTTP 非 201 | **即ロールバック** |
| 11 | CRUD C-2 | HTTP 非 200 | **即ロールバック** |
| 12 | CRUD C-3 | HTTP 非 409 | **即ロールバック** |
| 13 | CRUD C-4 | HTTP 非 204 | **即ロールバック** |
| 14 | Cleanup | DELETE 失敗 | 警告（手動削除対象） |

**判定ロジック**:
- すべてのテストが期待値を返す → PASS
- 1 つでも失敗 → FAIL → ロールバック判定へ

```bash
# 例: bash での判定
if [ ${STATUS} -ne 200 ]; then
  echo "FAIL: Expected 200, got ${STATUS}"
  exit 1
fi
```

---

## 7. GitHub Actions ロードマップ

### Phase 1: ファイル整備（2 週間）

- [ ] `.github/workflows/stg-health-check.yml` 作成
- [ ] `.github/workflows/stg-login-test.yml` 作成
- [ ] `.github/workflows/stg-crud-smoke-test.yml` 作成（11 テスト項目実装）
- [ ] `.github/workflows/stg-cleanup.yml` 作成
- [ ] GitHub Secrets に `STG_DEMO_EMAIL`, `STG_DEMO_PASSWORD` 登録

**実装チェックリスト**:
- [ ] Health Check: curl + jq で status 検証
- [ ] Login: credentials 取得 + token 抽出
- [ ] CRUD: 11 テストのループ + status code 検証
- [ ] Cleanup: DELETE + 監査ログ確認
- [ ] Error handling: `set -e` + trap で cleanup 保証
- [ ] timeout: 各 job に max-wait 設定

### Phase 2: 統合（1 週間）

- [ ] 既存 `stg-deploy.yml` に `stg-health-check.yml` トリガー追加
- [ ] Health check PASS 後に CRUD test を呼び出し
- [ ] CRUD test PASS 後に Cleanup を呼び出し
- [ ] 全フェーズ失敗時のロールバック判定フロー作成

**依存グラフ**:
```
stg-deploy
    ↓
stg-health-check (fail → rollback)
    ↓
stg-login-test (fail → rollback)
    ↓
stg-crud-smoke-test (fail → rollback)
    ↓
stg-cleanup (fail → manual cleanup)
```

### Phase 3: 本番化（当面は Option B 推奨）

- [ ] workflow_dispatch で手動トリガー確認（1 週間検証）
- [ ] daily smoke test の schedule (cron) 設定（オプション）
- [ ] ロールバック自動化の実装（CI failure → git revert → re-push）

---

## 8. 参考資料

- [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) - テスト項目・期待値・curl コマンド詳細
- [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) - GitHub Actions ワークフロー全体
- [README.md § 4](./README.md#4-デプロイ後のロールバック判定フレームワーク) - ロールバック判定基準
- AWS SSM Parameter Store ドキュメント
- GitHub Actions 公式ドキュメント

---

## 9. 認証情報保護ポリシー

本自動化で以下の情報は **絶対に記載しないこと**：

- ❌ demo account email（実メール）
- ❌ demo account password
- ❌ access_token / refresh_token（実トークン値）
- ❌ AWS credentials / Role ARN（直接記載）

**実装方法**:
- SSM Parameter Store または GitHub Secrets に格納
- スクリプト内では環境変数 (`${DEMO_EMAIL}` 等) で参照
- ログ出力時はマスク (`***` など)

