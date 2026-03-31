# INFRA-001: ステージング環境構築 — stg.noah-karte.com + AWSリソース名 stg 化

**起票日:** 2026-03-31
**優先度:** High
**カテゴリ:** インフラ移行
**担当:** インフラ担当

---

## 背景・目的

Vercel にカスタムドメイン `noah-karte.com` を取得した。
現在の `animalekarte-test-*` 環境をステージング環境として正式に位置づけ、
ドメイン `stg.noah-karte.com` を割り当てる。
合わせて AWS リソース名のプレフィックスを `animalekarte-test` → `animalekarte-stg` に統一する。

**本番環境はこのチケットでは用意しない。**

---

## 完了後の状態

| サービス | URL |
|---------|-----|
| フロントエンド (Vercel) | `https://stg.noah-karte.com` |
| バックエンド API (CloudFront) | `https://api.stg.noah-karte.com/api` |
| ALB（直接・内部用） | `http://animalekarte-stg-alb-*.us-east-1.elb.amazonaws.com` |
| RDS | `animalekarte-stg-db.*` |

---

## 作業フェーズ（クリティカルパス順）

```
Phase 1: DNS + ACM (リードタイム: 数分〜数時間)
    ↓
Phase 2: CloudFront + Vercel カスタムドメイン
    ↓
Phase 3: Terraform name_prefix 変更 + RDS in-place リネーム
    ↓
Phase 4: ECS/ALB/IAM/SSM 切り替え + GitHub Actions 更新
    ↓
Phase 5: 旧リソース削除 + ドキュメント更新
```

---

## Phase 1: DNS 設定

### 1-1. `stg.noah-karte.com` → Vercel (CNAME)

| 項目 | 値 |
|------|-----|
| レコード種別 | CNAME |
| ホスト名 | `stg` |
| 値 | `cname.vercel-dns.com` |

> Vercel ダッシュボード → Project Settings → Domains で確認した値と一致させること。

### 1-2. `api.stg.noah-karte.com` → CloudFront (CNAME)

| 項目 | 値 |
|------|-----|
| レコード種別 | CNAME |
| ホスト名 | `api.stg` |
| 値 | `dcqico6azu5w2.cloudfront.net` (現 CloudFront Distribution) |

> ⚠️ CloudFront に ACM 証明書をアタッチするまでは CNAME 追加のみ。実際の切り替えは Phase 2 で行う。

---

## Phase 2: ACM 証明書 + CloudFront / Vercel カスタムドメイン

### 2-1. ACM 証明書リクエスト (us-east-1 必須)

CloudFront は `us-east-1` の ACM 証明書しか使用できない。

```bash
export AWS_PROFILE=AnimalEkarte
aws acm request-certificate \
  --domain-name "stg.noah-karte.com" \
  --subject-alternative-names "*.stg.noah-karte.com" \
  --validation-method DNS \
  --region us-east-1
```

- DNS バリデーションレコード（CNAME）を `noah-karte.com` の DNS に追加
- 証明書ステータスが `ISSUED` になるまで待機

### 2-2. CloudFront Distribution に `api.stg.noah-karte.com` を追加

```bash
# Distribution の設定を取得・更新
aws cloudfront get-distribution-config \
  --id ERCVR5P0IAJKS \
  --region us-east-1 > cf-config.json

# Aliases に api.stg.noah-karte.com を追加
# ViewerCertificate を ACM 証明書 ARN に変更 (MinimumProtocolVersion: TLSv1.2_2021)
# jq で編集後に update-distribution を実行
```

**変更内容:**
- `Aliases.Items`: `["api.stg.noah-karte.com"]` を追加
- `ViewerCertificate.ACMCertificateArn`: Phase 2-1 で発行した証明書 ARN
- `ViewerCertificate.SSLSupportMethod`: `sni-only`
- `ViewerCertificate.MinimumProtocolVersion`: `TLSv1.2_2021`

### 2-3. Vercel カスタムドメイン追加

Vercel ダッシュボード → Project (animalekarte-frontend) → Settings → Domains:
- `stg.noah-karte.com` を追加
- DNS が正しく伝播されるまで待機（1-3 DNS 設定の CNAME が有効化されること）

---

## AWS リソース リネーム可否一覧

`name_prefix` 変更に伴う各リソースの対応方法を整理する。

| リソース | リネーム方法 | 備考 |
|---------|------------|------|
| **RDS インスタンス** | **in-place リネーム可** (`modify-db-instance`) | データ保持したままリネーム |
| VPC / Subnet / SG / NAT | Name タグ更新のみ | ID は変わらない |
| SSM パラメータ | 新パス作成 → 旧削除 | コピーで移行 |
| ECS Cluster | **destroy + create 必須** | 不変 ID リソース |
| ECS Service | **destroy + create 必須** | 不変 ID リソース |
| ALB / Target Group | **destroy + create 必須** | 不変 ID リソース |
| IAM Role | **destroy + create 必須** | 不変 ID リソース |
| ECR リポジトリ | **destroy + create 必須** (または継続使用) | 後述 Option A/B |
| CloudWatch Log Group | **destroy + create 必須** | 不変 ID リソース |
| ECS Task Definition | **destroy + create 必須** | Family 名が固定 |

---

## Phase 3: Terraform name_prefix 変更 + RDS in-place リネーム

### 3-1. Terraform state key 変更

**⚠️ 既存 state を消す前に必ずバックアップを取ること。**

```bash
export AWS_PROFILE=AnimalEkarte

# 現 state をコピー（バックアップ）
aws s3 cp \
  s3://animalekarte-tfstate-698109622668/env/test/terraform.tfstate \
  s3://animalekarte-tfstate-698109622668/env/test/terraform.tfstate.backup-$(date +%Y%m%d)

# stg 用にコピー
aws s3 cp \
  s3://animalekarte-tfstate-698109622668/env/test/terraform.tfstate \
  s3://animalekarte-tfstate-698109622668/env/stg/terraform.tfstate
```

**`infra/terraform/backend.tf` 変更:**
```hcl
# before
key = "env/test/terraform.tfstate"

# after
key = "env/stg/terraform.tfstate"
```

### 3-2. `infra/terraform/terraform.tfvars` 変更

```hcl
# before
name_prefix = "animalekarte-test"

# after
name_prefix             = "animalekarte-stg"
cors_allowed_origin     = "https://stg.noah-karte.com,https://api.stg.noah-karte.com"
```

### 3-3. RDS in-place リネーム

RDS は `modify-db-instance --new-db-instance-identifier` で**データを保持したままリネーム**できる。スナップショット移行は不要。

```bash
export AWS_PROFILE=AnimalEkarte

# Step 1: in-place リネーム（ダウンタイムなし）
aws rds modify-db-instance \
  --db-instance-identifier animalekarte-test-db \
  --new-db-instance-identifier animalekarte-stg-db \
  --apply-immediately \
  --region us-east-1

# Step 2: リネーム完了まで待機（数分）
aws rds wait db-instance-available \
  --db-instance-identifier animalekarte-stg-db \
  --region us-east-1

# Step 3: 新エンドポイント確認
aws rds describe-db-instances \
  --db-instance-identifier animalekarte-stg-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --region us-east-1
```

**Step 4: Terraform state を新しい識別子に対応させる**

リネーム後、Terraform は state の `animalekarte-test-db` と実態の `animalekarte-stg-db` が乖離するため、`moved` ブロックか `terraform state mv` で対応する。

```bash
# terraform state mv で対応（推奨）
terraform state mv \
  module.rds.aws_db_instance.main \
  module.rds.aws_db_instance.main
# ※ ただし identifier が変わるため、state 内の db_instance_identifier を手動パッチするか
# 一度 state から remove → import し直す方が確実

# あるいは state から一時 remove → 新 identifier で import
terraform state rm module.rds.aws_db_instance.main
terraform import module.rds.aws_db_instance.main animalekarte-stg-db
```

> **注意:** `terraform apply` 前に RDS の state 整合を取ること。整合が取れていないと Terraform が destroy + create しようとする。

### 3-4. terraform plan 実施・変更内容確認

```bash
cd infra/terraform
terraform init -reconfigure  # backend key 変更後は -reconfigure が必要
terraform plan -out=stg-migration.tfplan
```

**plan で確認すべき変更一覧:**

| リソース | 変更種別 | 備考 |
|---------|---------|------|
| `aws_ecs_cluster` | destroy + create | 不変 ID のため |
| `aws_ecs_service` | destroy + create | 不変 ID のため |
| `aws_lb` (ALB) | destroy + create | 不変 ID のため |
| `aws_lb_target_group` | destroy + create | 不変 ID のため |
| `aws_db_instance` | **update のみ** (3-3 で事前リネーム済み) | データ保持 |
| `aws_iam_role` (4個) | destroy + create | 不変 ID のため |
| `aws_cloudwatch_log_group` | destroy + create | 不変 ID のため |
| `aws_ssm_parameter` (3個) | destroy + create | 4-2 で事前コピー済みなら安全 |
| `aws_security_group` (3個) | Name タグ更新のみ | ID 変更なし |
| `aws_vpc` | Name タグ更新のみ | ID 変更なし |
| `aws_subnet` (4個) | Name タグ更新のみ | ID 変更なし |
| `aws_nat_gateway` | Name タグ更新のみ | ID 変更なし |
| `aws_ecr_repository` | ⚠️ リネーム不可（後述 Option A/B） | |

---

## Phase 4: 切り替え作業

### 4-1. ECR リポジトリ対応

ECR はリポジトリ名の変更が不可。以下のいずれかを選択:

**Option A (推奨・簡易):** 既存 `animalekarte-api` を継続使用し、タグで環境を区別する
- `stg-latest`, `stg-<sha>` タグをつける運用に変更
- Terraform の ECR リソースはリネームせず変数で分岐

**Option B:** 新リポジトリ `animalekarte-stg-api` を作成
- 既存イメージを新リポジトリに push
- Terraform・GitHub Actions の ECR_REPOSITORY を更新
- 旧リポジトリを削除

### 4-2. SSM パラメータ移行

```bash
export AWS_PROFILE=AnimalEkarte

# 値を取得して新パスに作成
for param in user password name; do
  VALUE=$(aws ssm get-parameter \
    --name "/animalekarte/test/db/${param}" \
    --with-decryption \
    --query 'Parameter.Value' \
    --output text \
    --region us-east-1)

  TYPE="String"
  if [ "$param" = "password" ]; then TYPE="SecureString"; fi

  aws ssm put-parameter \
    --name "/animalekarte/stg/db/${param}" \
    --value "$VALUE" \
    --type "$TYPE" \
    --region us-east-1
done

# 確認後に旧パラメータを削除
for param in user password name; do
  aws ssm delete-parameter \
    --name "/animalekarte/test/db/${param}" \
    --region us-east-1
done
```

### 4-3. IAM Role 更新（GitHub OIDC trust policy）

GitHub Actions の OIDC Role は ARN が変わるため、`backend-deploy.yml` の更新が必要。
また、新 IAM Role の trust policy のリポジトリ名は `MinoruSoga/AnimalEkarte` のまま維持する。

### 4-4. `backend/.env.production` 更新

```bash
# 変更箇所
CORS_ALLOWED_ORIGIN=https://stg.noah-karte.com,https://api.stg.noah-karte.com

# CloudWatch ログ参照（workflow 内で使用）
# /ecs/animalekarte-test → /ecs/animalekarte-stg
```

### 4-5. `.github/workflows/backend-deploy.yml` 更新

```yaml
env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: animalekarte-api          # Option A なら変更なし
  ECS_CLUSTER: animalekarte-stg-cluster     # ← 変更
  ECS_SERVICE: animalekarte-stg-service     # ← 変更
  ECS_TASK_DEFINITION_FAMILY: animalekarte-stg-api    # ← 変更
  ECS_MIGRATE_TASK_FAMILY: animalekarte-stg-migrate   # ← 変更
  ECS_SUBNETS: <新 SG 再作成の場合は更新>
  ECS_SECURITY_GROUPS: <新 SG ID に更新>

# L47: role-to-assume
role-to-assume: arn:aws:iam::698109622668:role/animalekarte-stg-github-ecs-deploy-role  # ← 変更

# L229: ログ参照
--log-group-name /ecs/animalekarte-stg   # ← 変更
```

### 4-6. Vercel 環境変数 `VITE_API_URL` 更新

Vercel ダッシュボード → Project Settings → Environment Variables (Production):

```
VITE_API_URL=https://api.stg.noah-karte.com/api
```

---

## Phase 5: 旧リソース削除 + ドキュメント更新

### 5-1. 動作確認完了後に旧リソース削除

以下を確認してから削除:
- [ ] `https://stg.noah-karte.com/login` でログイン → セッション維持確認
- [ ] `https://api.stg.noah-karte.com/health` が 200 を返すこと
- [ ] ECS タスクが `animalekarte-stg-*` で稼働中であること
- [ ] CloudWatch Logs `/ecs/animalekarte-stg` にログが流れていること

削除対象（Terraform が管理していないもの / 手動削除）:
- 旧 ECS Cluster: `animalekarte-test-cluster`（サービスを削除後）
- 旧 CloudWatch Log Group: `/ecs/animalekarte-test`
- 旧 Vercel URL `frontend-eta-six-20.vercel.app` のプロジェクト設定 (オプション)

### 5-2. ドキュメント更新対象ファイル

| ファイル | 更新内容 |
|---------|---------|
| `infra/CLAUDE.md` | 全エンドポイント URL・リソース名 |
| `infra/terraform/terraform.tfvars` | name_prefix, cors_allowed_origin |
| `infra/terraform/backend.tf` | state key |
| `infra/docs/deployment-guide.md` | エンドポイント・リソース名 |
| `infra/docs/architecture.md` | 構成図・リソース名 |
| `docs/infra/deploy/deployment-status.md` | 全リソース名・URL |
| `docs/infra/deploy/CI-CD-PIPELINE.md` | ECS クラスター名・URL |
| `docs/infra/deploy/README.md` | アクセス URL |
| `docs/infra/deploy/DEPLOYMENT-CHECKLIST.md` | 全コマンドのリソース名 |
| `docs/infra/docs_infra_architecture.md` | 構成図・リソース名 |
| `docs/infra/SECURITY-CHECKLIST.md` | リソース名 |

---

## リスク・注意事項

| リスク | 影響 | 対策 |
|--------|------|------|
| RDS の state 乖離 → terraform apply が destroy しようとする | **最大** | 3-3 で in-place リネーム後、`terraform state rm` → `terraform import` で state を同期してから apply |
| IAM Role ARN 変更 → GitHub Actions 認証エラー | High | role-to-assume を backend-deploy.yml に更新してから旧 Role を削除 |
| ECS Security Group ID 変更 → backend-deploy.yml の ECS_SECURITY_GROUPS が古い | High | terraform apply 後に新 SG ID を workflow に反映 |
| CloudFront CNAME 追加前に DNS を向けると SSL エラー | Medium | Phase 2 の順序を厳守（ACM 発行 → CF 設定 → DNS 向け） |
| Terraform state key 変更後に init 忘れ | Medium | `terraform init -reconfigure` を必ず実行 |
| 旧 Vercel URL が残ったままで CORS_ALLOWED_ORIGIN から外れる | Low | `.env.production` 更新後に ECR push → ECS deploy を実行 |

---

## 実施順序チェックリスト

### Phase 1（即時）
- [ ] `stg.noah-karte.com` CNAME → `cname.vercel-dns.com`
- [ ] `api.stg.noah-karte.com` CNAME → `dcqico6azu5w2.cloudfront.net`

### Phase 2（ACM 証明書発行後）
- [ ] ACM 証明書リクエスト (`us-east-1`)
- [ ] DNS バリデーション CNAME 追加 → `ISSUED` 確認
- [ ] CloudFront に `api.stg.noah-karte.com` CNAME + ACM 証明書設定
- [ ] Vercel に `stg.noah-karte.com` カスタムドメイン追加・確認

### Phase 3
- [ ] Terraform state のバックアップ (S3 コピー)
- [ ] `backend.tf` の state key 変更 → `terraform init -reconfigure`
- [ ] `terraform.tfvars` の `name_prefix` 変更
- [ ] RDS in-place リネーム: `modify-db-instance --new-db-instance-identifier animalekarte-stg-db`
- [ ] RDS リネーム完了待機: `aws rds wait db-instance-available`
- [ ] Terraform state 同期: `state rm` → `terraform import` で `animalekarte-stg-db` を登録
- [ ] `terraform plan` で変更内容確認（RDS が update のみになっていること）

### Phase 4
- [ ] ECR 対応方針決定 (Option A / B)
- [ ] `terraform apply`（RDS は事前リネーム + state import 済みのため安全）
- [ ] SSM パラメータ移行 (test → stg パス)
- [ ] `backend/.env.production` 更新 (CORS_ALLOWED_ORIGIN)
- [ ] `backend-deploy.yml` 更新 (ECS リソース名・IAM Role ARN・SG ID・ログ名)
- [ ] Vercel `VITE_API_URL` を `https://api.stg.noah-karte.com/api` に更新
- [ ] `git push origin main` → GitHub Actions デプロイ確認

### Phase 5
- [ ] `https://stg.noah-karte.com/login` でログイン動作確認
- [ ] `https://api.stg.noah-karte.com/health` が 200 確認
- [ ] CloudWatch Logs `/ecs/animalekarte-stg` にログ確認
- [ ] 旧リソース削除 (ECS cluster, Log Group 等)
- [ ] ドキュメント 12ファイル更新

---

## 参照ドキュメント

- `infra/CLAUDE.md` — 現環境の全リソース一覧
- `infra/terraform/terraform.tfvars` — Terraform 変数
- `infra/terraform/backend.tf` — state 設定
- `.github/workflows/backend-deploy.yml` — CI/CD パイプライン
- `docs/infra/deploy/deployment-status.md` — AWS リソース詳細
