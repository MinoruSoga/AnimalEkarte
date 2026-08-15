# STG AWS 変更前事前準備ドキュメント

> **時点スナップショット**: 2026-06-17 時点の調査・承認材料。改変不可（P1/P2実施完了後はarchiveへ移動想定）。

> **Animal Ekarte**: STG 環境 AWS 変更の事前チェックリスト・承認材料・手順・検証・ロールバック整理
> **作成日**: 2026-06-17 | **対象環境**: Staging (us-east-1 / ap-northeast-1)
> **参照元**: [STG_AWS_COST_REDUCTION.md](./stg-aws-cost-reduction.md)
> **ステータス**: 事前準備フェーズ（AWS リソース変更は未実施）

---

## 0. リスク分類と安全境界

| 分類 | 内容 | 判断 |
|------|------|------|
| Read-only | 既存ドキュメント確認・AWS CLI describe 系 | 自律実行可 |
| Local write | docs/ops/ 配下 Markdown 作成・更新 | 自律実行可 |
| External write | AWS 変更・PR 作成・デプロイ・外部投稿 | **明示承認必須** |
| Destructive / credential-impacting | EIP release・ALB 置換・DNS 変更・認証情報変更 | **明示承認必須** |

**このドキュメントに記載の「承認後手順」は一切実行しない。承認後に別セッションで実施する。**

---

## 1. 変更候補サマリー

| 案 | 推定削減額 | 優先度 | 不可逆性 | 承認状況 |
|---|-----------|--------|---------|---------|
| P1: ap-northeast-1 EIP release | ~$1.65/月 | P1 | **高（IP 復元不可）** | 未承認 |
| P2: internal ALB + CloudFront VPC Origin | ~$7.20/月 | P2 | 中（Terraform で再作成可） | 未承認 |
| P4: fck-nat 代替検討 | ~$2.93/月 | P4 | 高（単純削除は不可） | 未着手 |

> P3（ap-northeast-1 t3.micro 遅延請求）は自然消滅待ち。追加アクション不要。

---

## 2. 変更前 Read-only 情報取得項目

以下のコマンドはすべて **read-only**。承認前・確認フェーズで実行可能。
**必ず `--profile AnimalEkarte` を付与すること。**

### 2.1 EIP 関連（P1 対象）

```bash
# ap-northeast-1 の全 EIP 確認（AllocationId・AssociationId・PublicIp）
aws --profile AnimalEkarte --region ap-northeast-1 ec2 describe-addresses \
  --output table

# us-east-1 の EIP 一覧（比較用）
aws --profile AnimalEkarte --region us-east-1 ec2 describe-addresses \
  --output table
```

**確認すべき内容**:
- `AssociationId` が空欄 = 未関連（release 対象候補）
- `PublicIp` の値を控え、DNS・allowlist・CI/CD に参照がないか検索する

### 2.2 ALB・SG・CloudFront 関連（P2 対象）

```bash
# ALB の scheme・AZ・SG 確認
aws --profile AnimalEkarte --region us-east-1 elbv2 describe-load-balancers \
  --output table

# ALB の listener 確認（HTTP ポート・TLS 設定）
aws --profile AnimalEkarte --region us-east-1 elbv2 describe-listeners \
  --load-balancer-arn <ALB_ARN> --output table

# ECS Service の loadBalancers・networkConfiguration 確認
aws --profile AnimalEkarte --region us-east-1 ecs describe-services \
  --cluster animalekarte-stg --services animalekarte-stg-service --output json

# CloudFront distribution のオリジン設定確認
aws --profile AnimalEkarte --region us-east-1 cloudfront list-distributions \
  --output table

# ALB の Security Group inbound ルール確認
aws --profile AnimalEkarte --region us-east-1 ec2 describe-security-groups \
  --filters "Name=description,Values=*animalekarte*stg*" --output table
```

**確認すべき内容**:
- ALB scheme が `internet-facing` であることを確認（変更後は `internal`）
- CloudFront origin domain が ALB の DNS を参照していることを確認
- ALB SG inbound ルールの現状（現行は `0.0.0.0/0`。internal 化後は VPC CIDR に絞る）

### 2.3 fck-nat・ルートテーブル（P4 対象）

```bash
# fck-nat EC2 インスタンス確認
aws --profile AnimalEkarte --region us-east-1 ec2 describe-instances \
  --filters "Name=tag:Name,Values=*fck-nat*" --output table

# Private Subnet のルートテーブル確認（0.0.0.0/0 の向き先）
aws --profile AnimalEkarte --region us-east-1 ec2 describe-route-tables \
  --output table

# ECS タスクが使う VPC エンドポイント一覧（あれば）
aws --profile AnimalEkarte --region us-east-1 ec2 describe-vpc-endpoints \
  --output table
```

**確認すべき内容**:
- Private Subnet RT の `0.0.0.0/0` が fck-nat EC2 の ENI を向いていることを確認
- ECS が ECR・CloudWatch・SSM に接続する経路（NAT か VPC エンドポイントか）

### 2.4 コスト確認（全体）

```bash
# 直近30日のサービス別コスト（Cost Explorer）
aws --profile AnimalEkarte --region us-east-1 ce get-cost-and-usage \
  --time-period Start=2026-05-18,End=2026-06-17 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=SERVICE \
  --output table
```

---

## 3. 変更候補別：事前チェックリスト・実施条件・ロールバック

---

### 3.1 [P1] ap-northeast-1 EIP release

#### 目的

ap-northeast-1 に関連リソースなしで保持されている EIP の解放。
推定 ~$1.65/月 の削減。

#### 対象

- ap-northeast-1 未関連 EIP（AllocationId を事前 describe で確認）

#### 前提条件

| 確認項目 | 確認方法 | 結果 |
|---------|---------|------|
| EIP が実際に未関連（AssociationId が空）| `describe-addresses` | 要確認 |
| DNS レコードがこの IP を参照していない | Route 53 / 外部 DNS の A レコード検索 | 要確認 |
| CI/CD（GitHub Actions）にこの IP がハードコードされていない | `.github/workflows/` 全ファイル grep | 要確認 |
| allowlist（Webhook・外部サービス）にこの IP がない | Slack/LINE/外部 SaaS 設定確認 | 要確認 |
| 監視設定にこの IP が含まれていない | CloudWatch アラーム・外部監視確認 | 要確認 |
| 承認者の明示承認を得ている | 承認欄（§4）に記入 | 要承認 |

#### 依存関係

- STG リソースは us-east-1 に集約。ap-northeast-1 の EIP は STG 構成と無関係の可能性が高い。
- ただし過去の作業で使用された IP が外部サービスの allowlist に登録されている可能性を排除するまで実施不可。

#### 不可逆性の明示

> ⛔ **EIP を解放すると同一 IP アドレスの復元は保証されない。**
> AWS が IP をプールに返却するため、解放後に同一 IP を取得しようとしても別の IP が割り当てられる。
> 外部サービスの allowlist に登録済みの IP を解放すると、その外部サービスとの接続が断絶する。

#### 実施条件（すべて満たした場合のみ実施）

- [ ] `describe-addresses` で AssociationId が空であることを最終確認
- [ ] DNS・allowlist・CI/CD・監視の参照なしを確認
- [ ] 承認者が § 4 の承認欄に署名（名前・日付）

#### 承認後実施手順（実行しない。承認後に別セッションで実施）

```bash
# Step 1: 解放直前に再確認
aws --profile AnimalEkarte --region ap-northeast-1 ec2 describe-addresses \
  --output table

# Step 2: 承認済みの AllocationId を指定して解放
aws --profile AnimalEkarte --region ap-northeast-1 ec2 release-address \
  --allocation-id <AllocationId>

# Step 3: 解放確認
aws --profile AnimalEkarte --region ap-northeast-1 ec2 describe-addresses \
  --output table
```

#### ロールバック方針

| 状況 | 対応 |
|------|------|
| 解放後に外部接続断が発生した場合 | 新 EIP 取得（別 IP になる）+ allowlist 更新が必要 |
| 解放前に問題が発覚した場合 | 実施を中止（EIP を保持） |

> 解放後の完全ロールバック（同一 IP での復元）は **不可能**。

#### 検証方法

- 解放後に `describe-addresses` で ap-northeast-1 に EIP が存在しないことを確認
- 翌月の Cost Explorer で EC2 Other（EIP 保持費用）が ap-northeast-1 から消えることを確認

---

### 3.2 [P2] internal ALB + CloudFront VPC Origin 移行

#### 目的

ALB を `internet-facing` から `internal` に変更し、CloudFront VPC Origin 経由で接続。
ALB のパブリック IP 2 個（推定 ~$7.20/月）を削減。ALB hourly コストは変化しない。

#### 対象

- ALB scheme 変更（internet-facing → internal）
- CloudFront VPC Origin 設定追加
- ALB Security Group inbound ルール更新

#### SG 設計（Phase 1 / Phase 2）

| フェーズ | SG inbound 許可元 | 実施タイミング |
|---------|-----------------|--------------|
| **Phase 1**（Terraform apply 時 / `alb_internal=true`） | **VPC CIDR**（`var.vpc_cidr`）。VPC Origin の CloudFront トラフィックは VPC Origins サービスが VPC subnet に配置する ENI 経由で到達するため source は VPC 内 | `terraform apply` と同時に適用される |
| **Phase 2**（apply 後・任意の追加絞り込み） | VPC Origin 作成後に AWS が自動生成する VPC Origins サービス管理 SG（名称は apply 後に確認） | VPC Origin 作成確認後に実名を確認してから ALB SG を更新 |

> **注意**: ALB SG inbound の許可元は Terraform で `alb_internal` に連動する（`modules/security/main.tf` の `local.alb_ingress_cidrs`）。
> `alb_internal=false`（現行 internet-facing）では `0.0.0.0/0` を維持し、`alb_internal=true`（P2）では VPC CIDR に絞る。
> したがって `0.0.0.0/0` → VPC CIDR の SG 差分は **ALB の internal 化（`alb_internal=true`）と同一 apply で同時に適用される**。
>
> ⚠️ **public CloudFront managed prefix list は使わない**: internal ALB に VPC Origin 経由で到達する CloudFront
> トラフィックの source は VPC CIDR 内であり、public な `com.amazonaws.global.cloudfront.origin-facing` prefix list は
> internet-facing オリジン用で internal ALB には不適。加えて managed prefix list は SG ルール上限に対し重み ~45/参照 を
> 消費し、80/443 の 2 参照で既定 quota 60 を超過する（VPC CIDR ベースは重み 2 で quota 内）。
>
> ⚠️ **Phase 2 は機能要件ではなく追加の最小権限化**: Phase 1 の VPC CIDR ベースだけでも
> CloudFront → VPC Origin → internal ALB の inbound 許可として成立する。Phase 2 は AWS が自動生成する
> `CloudFront-VPCOrigins-Service-SG` を許可元にして、対象 CloudFront distribution からの通信へさらに絞り込む場合に実施する。
> Phase 2 を実施する場合も、まず Phase 1 で疎通確認し、その後に SG 差し替えと再 smoke test を行う。
>
> ⚠️ **Phase 2 SG 実名確認手順**: VPC Origin apply 後に AWS が `CloudFront-VPCOrigins-Service-SG` を含む名称の
> サービス管理 SG を自動生成する。正確な SG 名は AWS が決定するため事前に確定できない。
> apply 後に以下の `describe-security-groups` で実名を確認し、その実名で data source 化すること。
>
> ```bash
> # apply 後に実行（apply 前は SG が存在しないため実行不可）
> aws --profile AnimalEkarte --region us-east-1 ec2 describe-security-groups \
>   --filters "Name=description,Values=*VPCOrigins*" \
>   --query 'SecurityGroups[*].{Name:GroupName,ID:GroupId}' \
>   --output table
> ```

#### CloudFront distribution 更新（Terraform 管理外・手動）

> ⛔ **CloudFront distribution は現在 Terraform 管理外（手動作成済み）**。
> `terraform plan` / `apply` の差分には CloudFront distribution の変更は**含まれない**。
> VPC Origin apply 後に AWS コンソールまたは AWS CLI で手動でオリジンを切り替える。

手動手順（apply 後）:
```bash
# 1. VPC Origin ID を確認（terraform output から取得）
terraform output vpc_origin_id

# 2. CloudFront distribution のオリジン設定を VPC Origin に手動更新
#    AWS コンソール → CloudFront → Distribution → Origins → Edit → VPC Origin ID を設定
```

#### スモークテスト条件（ECS 夜間停止に注意）

> ⚠️ **夜間メンテナンス窓（22:00〜08:00 JST）中は ECS desiredCount=0 のため `/health` へのアクセス不可**。
> `curl -I https://api.stg.noah-karte.com/health` は ALB HealthCheck が全タスク UNHEALTHY 状態のため 503 を返す。
>
> **smoke test を実施するには以下のいずれかが必要**:
> 1. ECS を一時起動（`aws --profile AnimalEkarte ecs update-service ... --desired-count 1`）⛔ **AWS write・明示承認必須**
> 2. 翌朝 08:00 JST 以降に確認（EventBridge スケジューラが自動起動）

#### 前提条件

| 確認項目 | 確認方法 | 結果 |
|---------|---------|------|
| CloudFront VPC Origin が us-east-1 で利用可能 | AWS コンソール or Terraform plan | 要確認 |
| Terraform state が最新 (`terraform plan` で差分なし) | `terraform plan` (ドライラン) | 要確認 |
| ALB replace が Terraform plan で明示されている | `terraform plan` 出力確認 | 要確認 |
| ALB replace 中のダウンタイムを関係者に周知済み | メンバー確認 | 要承認 |
| Phase 1 SG（`0.0.0.0/0` → VPC CIDR の絞り込み）が ALB internal 化と同一 apply で適用されることを確認した | `terraform plan` で SG 差分確認 | 要確認 |
| CloudFront distribution のオリジン切り替えが手動作業であることを理解した | 設計確認（本ドキュメント SG 設計参照） | 要確認 |
| 移行後のスモークテスト手順と ECS 夜間停止の制約を確認した | § 3.2 スモークテスト条件参照 | 要確認 |
| 承認者の明示承認を得ている | 承認欄（§ 4）に記入 | 要承認 |

#### 依存関係

| 依存元 | 変更への影響 |
|--------|-----------|
| ECS service `loadBalancers` 設定 | ALB replace で Target Group も再作成される可能性。ECS サービス更新が必要 |
| ECS タスク SG（ALB SG からの 8080 のみ許可） | 新 ALB SG が変わる場合は ECS SG の inbound 更新も必要 |
| CloudFront オリジン（ALB DNS を参照） | VPC Origin 設定後に CloudFront origin を更新 |
| `backend-deploy.yml` wait-for-service-stability | 新 ALB・新 Target Group のヘルスチェック確認が必要 |
| `stg-smoke.yml`（`/health` エンドポイント） | 移行後のスモークテスト確認 |
| Terraform state | ALB replace は Terraform が `destroy` + `create` を実行する（plan で確認必須） |

#### ダウンタイム見積もり

| フェーズ | 操作 | 推定ダウンタイム |
|---------|------|----------------|
| `terraform apply` (ALB replace) | plan 結果に従う（replace 順序は Terraform 設定次第） | 5〜15 分（推定） |
| CloudFront → 新 ALB オリジン切り替え | CloudFront 設定更新 | 伝播に最大 数分 |
| ECS サービス更新（Target Group 変更時） | ECS ローリング更新 | 変更内容による |

> STG 環境のため、**夜間メンテナンス窓（22:00〜08:00 JST / ECS desiredCount=0 の時間帯）**に実施を推奨。

#### Terraform 操作計画（実行しない。承認後に別セッションで実施）

```bash
# Step 1: Terraform plan で影響確認（read-only）
cd infra/terraform
terraform plan -out=plan_p2_internal_alb.tfplan

# plan 出力で以下を確認:
#   - aws_lb が "must be replaced" になっているか（internal=true に変更するため必須）
#   - aws_lb_listener が再作成対象か
#   - aws_ecs_service が更新対象か
#   - aws_security_group.alb の ingress 80/443 が `0.0.0.0/0` -> VPC CIDR に絞り込み更新されるか
#   - aws_cloudfront_vpc_origin.alb が "will be created" になっているか
#
# ⚠️ CloudFront distribution の変更は Terraform 管理外のため plan に含まれない。
#    VPC Origin apply 後に AWS コンソールで手動オリジン切り替えが別途必要。

# Step 2: plan 内容を承認者に共有・承認後に apply
terraform apply plan_p2_internal_alb.tfplan

# Step 3: VPC Origin ID 確認（distribution 手動切り替えに使用）
terraform output vpc_origin_id

# Step 4: CloudFront distribution のオリジンを VPC Origin に手動切り替え（AWS コンソール）

# Step 5: smoke test（ECS が稼働中であることを確認してから実行）
# 注意: 夜間（22:00-08:00 JST）は ECS desiredCount=0 のため実行不可
curl -I https://api.stg.noah-karte.com/health
```

#### ロールバック方針

| 状況 | ロールバック手順 |
|------|---------------|
| `terraform apply` 失敗（ALB 作成失敗） | Terraform state と実リソースの状態を確認し、旧構成へ戻す差分を `terraform plan` で作成して承認後に apply |
| ALB 作成成功・ECS 疎通不可 | ALB SG の inbound ルール確認、VPC Origin SG の設定確認 |
| CloudFront → 新 ALB 疎通不可 | CloudFront origin を旧 ALB DNS に一時戻す（旧 ALB が削除済みの場合は不可）|
| 全断・即時復旧必要 | 旧構成（internet-facing ALB）に戻す Terraform 変更を plan → apply |

> Terraform の `terraform.tfstate` が正しければロールバックは可能。state 破損時は手動対応が必要。

#### 検証方法（承認後実施後）

```bash
# 1. 新 ALB の scheme が internal であることを確認
aws --profile AnimalEkarte --region us-east-1 elbv2 describe-load-balancers \
  --output table

# 2. VPC Origin が作成されていることを確認
terraform output vpc_origin_id
aws --profile AnimalEkarte --region us-east-1 cloudfront list-vpc-origins --output table

# 3. CloudFront → 新 ALB のヘルスチェック確認（スモークテスト）
# ⚠️ 夜間（22:00-08:00 JST）は ECS desiredCount=0 のため /health は 503 を返す。
#    一時的に ECS を起動するか、翌朝 08:00 JST 以降に実行すること。
curl -I https://api.stg.noah-karte.com/health

# 4. ECS タスクが Running かつ HEALTHY であることを確認（ECS 稼働中のみ有効）
aws --profile AnimalEkarte --region us-east-1 ecs describe-services \
  --cluster animalekarte-stg --services animalekarte-stg-service --output json

# 5. 翌月の Cost Explorer で VPC Public IPv4 コストが減少していることを確認
# （ALB 分 2IP が internal 化で課金なし）
```

---

### 3.3 [P4] fck-nat 代替検討（長期・現フェーズでは実施しない）

#### 目的

Private Subnet の NAT 経路を fck-nat EC2 ($2.93/月) から代替構成に変更。
**現フェーズでは実施しない。P2 完了後に優先度を再評価する。**

#### 現状の制約

- Private Subnet の `0.0.0.0/0` ルートが fck-nat EC2 ENI を参照。
- ECS タスク（ECR イメージ pull・CloudWatch Logs push・SSM Parameter Store 参照）に NAT が必須。
- **fck-nat 単純削除は不可**。削除すると ECS の ECR/CW/SSM 接続が全断する。

#### 代替候補と評価（現時点の調査）

| 代替案 | 推定月額 | メリット | デメリット | 実施判断 |
|--------|---------|---------|----------|---------|
| fck-nat 現状維持 | $2.93 | 安定稼働中・変更不要 | 停止時 ECS 通信断リスク | **現状維持（P2 後再評価）** |
| VPC Endpoint（ECR・CW・SSM 個別） | $7〜15 | NAT 不要 | エンドポイント数増でコスト増 | 要詳細調査 |
| ECS Public Subnet 移行 | $0（NAT 不要） | NAT 費用ゼロ | SG 設計変更・セキュリティ再評価必要 | 要設計検討 |

#### 事前に必要な追加調査（実施前に完了すること）

- [ ] ECS タスクが実際に接続している外部エンドポイントの一覧化
- [ ] VPC Endpoint で代替可能なエンドポイントの特定
- [ ] ECS Public Subnet 移行時の SG 設計案・セキュリティレビュー
- [ ] P2（internal ALB）完了後のコスト再計算

---

## 4. 承認チェックリスト

> 変更実施前に承認者が記入・署名すること。未記入の項目がある変更は実施禁止。

### P1: ap-northeast-1 EIP release

| 確認項目 | 確認者 | 確認日 | 結果 |
|---------|--------|--------|------|
| EIP が未関連であることを describe-addresses で確認した | | | |
| DNS・外部 allowlist に参照がないことを確認した | | | |
| CI/CD・監視設定に IP がないことを確認した | | | |
| EIP 解放後に同一 IP 復元不可であることを理解した | | | |
| **P1 実施を承認する** | | | |

### P2: internal ALB + CloudFront VPC Origin

| 確認項目 | 確認者 | 確認日 | 結果 |
|---------|--------|--------|------|
| Terraform plan を確認し、ALB replace を含む影響範囲を把握した | | | |
| ダウンタイム（5〜15 分推定）を関係者に周知した | | | |
| 夜間メンテナンス窓での実施に合意した | | | |
| ロールバック手順を確認した | | | |
| スモークテスト手順（§ 3.2 検証方法）を準備した | | | |
| **P2 実施を承認する** | | | |

---

## 5. 実施判断表

| 案 | 実施条件（すべて必要） | 現時点の状態 |
|---|-------------------|-----------| 
| P1: EIP release | DNS/allowlist 確認済み + 承認署名済み | ⬜ 未完了 |
| P2: ALB 移行 | Terraform plan 確認済み + ダウンタイム同意 + 承認署名済み | ⬜ 未完了 |
| P4: fck-nat 代替 | P2 完了 + 代替設計レビュー完了 + 承認署名済み | ⬜ 未着手（P2 後） |

---

## 6. 事後確認計画（CloudWatch / Cost Explorer）

変更実施後、以下を確認する。

### P1 完了後

| 確認項目 | 確認タイミング | 確認コマンド |
|---------|-------------|------------|
| ap-northeast-1 EIP が消えていること | 実施直後 | `describe-addresses --region ap-northeast-1` |
| EC2 Other（EIP 保持費用）が翌月請求から消えること | 翌月 Cost Explorer | `ce get-cost-and-usage` |
| 外部サービスの接続断がないこと | 実施後 24 時間以内 | 監視アラート確認 |

### P2 完了後

| 確認項目 | 確認タイミング | 確認コマンド |
|---------|-------------|------------|
| ALB scheme が internal になっていること | apply 直後 | `elbv2 describe-load-balancers` |
| `/health` エンドポイントへの疎通 | apply 直後 | `curl -I https://api.stg.noah-karte.com/health` |
| ECS サービスが RUNNING / HEALTHY | apply 直後 | `ecs describe-services` |
| VPC Public IPv4 コスト（ALB 分）が翌月から減少 | 翌月 Cost Explorer | `ce get-cost-and-usage` |
| デプロイパイプライン（`backend-deploy.yml`）が正常動作 | 次回デプロイ時 | GitHub Actions ジョブ結果 |

---

## 7. 不可案の再確認（変更禁止）

以下は既存レポートで不可と確認済み。このドキュメントでも変更対象外とする。

| 案 | 判定 | 理由 |
|----|------|------|
| ALB 単純削除 | **禁止** | ECS SG・CloudFront 経路・デプロイ依存が全断 |
| fck-nat 単純削除 | **禁止** | ECS の ECR/CW/SSM 接続が全断 |
| ECS Fargate 128 CPU | **禁止** | Fargate に 128 CPU は存在しない |
| RDS db.t4g.nano | **禁止** | メモリ不足リスクあり・有効クラス未確認 |
| RDS ストレージ 10GB 縮小 | **禁止** | gp3 最小 20GB・RDS はストレージ縮小非対応 |

---

## 8. リスクまとめ

| リスク | 対象施策 | 影響 | 緩和策 |
|--------|---------|------|--------|
| EIP 解放後に外部サービスの接続断 | P1 | 外部 Webhook・監視断絶 | 解放前に DNS/allowlist を網羅確認 |
| ALB replace 中のダウンタイム | P2 | STG API 一時断（5〜15 分推定） | 夜間メンテナンス窓実施 + 事前周知 |
| CloudFront VPC Origin 設定誤りによる経路断 | P2 | STG API 全断 | plan 確認 + スモークテスト（ECS 稼働中に限る） |
| 夜間停止（22:00-08:00 JST）中の smoke test 失敗 | P2 | /health が 503（ECS desiredCount=0 のため） | 翌朝 08:00 JST 以降または一時 ECS 起動で確認 |
| Terraform state 不整合によるロールバック困難 | P2 | 手動復旧が必要 | apply 前に state バックアップを確保 |
| Cost Explorer 推定値との乖離 | 全体 | 月額削減効果のズレ | 実施後翌月に Cost Explorer 再確認 |

---

## 9. 機密情報取り扱い注意

このドキュメントには以下を**記載しない**:

- AWS アカウント ID（完全形）
- 完全 ARN
- パブリック IP アドレスの完全な値
- 認証情報・API キー・シークレット

環境固有の値（AllocationId・ARN 等）は実施時に `describe-*` コマンドで取得すること。

---

## 10. 参照

| 内容 | リンク |
|------|--------|
| STG コスト削減レポート（調査元） | [STG_AWS_COST_REDUCTION.md](./stg-aws-cost-reduction.md) |
| CloudFront VPC Origins 公式ドキュメント | https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-vpc-origins.html |
| ALB scheme 変更仕様（replace 必須） | https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-elasticloadbalancingv2-loadbalancer.html |
| EIP の動作と解放 | https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/elastic-ip-addresses-eip.html |
