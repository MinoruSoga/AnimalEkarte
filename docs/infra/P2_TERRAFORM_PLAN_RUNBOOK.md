# P2 Terraform Full Plan Runbook（STG）

> **目的**: P2 Terraform変更の承認・実行手順を定義する。
> **読者**: 承認者・実行者。
> **タイミング**: P2 full plan 承認プロセス実行時。

STG の **P2（internal ALB + CloudFront VPC Origin）** を承認するための full plan 手順。
秘密値を露出させずに前提条件を検証し、承認判断を `terraform plan -out=tfplan` の結果のみで行うためのランブック。

> 真実の源泉: `infra/terraform/variables.tf` / `infra/terraform/.gitignore` / `infra/CLAUDE.md` Terraform 安全ルール。
> 本ドキュメントは秘密値・完全 ARN・完全アカウント ID・tfvars の中身を含まない。

---

## 1. 現在のブロッカー

- `infra/terraform/terraform.tfvars` が**不在**のため full plan が実行できない。
- `db_password` は `variables.tf` で唯一 `default` を持たない `sensitive` 変数であり、値の供給元が無いと plan が成立しない。
- P2 差分（ALB replace / VPC Origin create）は `alb_internal = true` が実 tfvars で設定されて初めて plan に現れる。default は `false`（internet-facing = 安全側）。

**結論**: 実 tfvars を用意するまで full plan は BLOCKED。preflight script は秘密値を読まずにこの状態を判定する。

---

## 2. Preflight script

`scripts/infra-terraform-plan-preflight.sh` は前提条件を検証し、tfvars が揃っている場合のみ validate -> full plan を実行する。

```sh
sh scripts/infra-terraform-plan-preflight.sh
```

挙動:

- `terraform.tfvars` は**存在のみ**を確認し、内容は読まない/表示しない。
- tfvars 不在時は plan を実行せず BLOCKED で終了する。
- `AWS_PROFILE=AnimalEkarte` を明示する（安全ルール #1）。
- validate を plan より前に必ず実行する。
- 対象を絞らない full plan（`-out=tfplan`）のみを生成する。適用は行わない。

終了コード:

| code | 意味 | 次のアクション |
|------|------|---------------|
| 0 | plan 生成成功 | §5 の差分レビューへ |
| 1 | validate / plan 失敗 | エラー内容を修正して再実行 |
| 2 | terraform ディレクトリ不在 | リポジトリ構成を確認 |
| 3 | `terraform.tfvars` 不在 | §3 で tfvars を用意 |
| 4 | terraform 未初期化（`.terraform` 不在） | `(cd infra/terraform && AWS_PROFILE=AnimalEkarte terraform init)` を手動実行 |
| 5 | terraform コマンド不在 | terraform をインストール |

> init は backend（S3 state）と network を触るため、script は自動実行しない。code 4 のときのみ手動で init する。

---

## 3. 必須 tfvars フィールド

`infra/terraform/terraform.tfvars` を用意する。**このファイルは gitignore 済み（commit 禁止）**。
`variables.tf` の `default` を持たない変数は `db_password` のみ。P2 承認では `alb_internal = true` を追加する。

```hcl
# infra/terraform/terraform.tfvars  (git ignored — commit しない)
#
# db_password: variables.tf で唯一 default を持たない sensitive 変数（8文字以上）。
#   実値はローカル tfvars にのみ記述し、本ドキュメントには記載しない（プレースホルダも置かない）。
alb_internal = true   # P2 承認 plan に必須（default false では差分が出ない）
```

- 雛形は `infra/terraform/terraform.tfvars.example` を参照（`variables.tf` を正とし、example の余剰キーは無視してよい）。
- 秘密値は本ドキュメントや example ではなくローカル tfvars にのみ置く。SecureString の実値は SSM Parameter Store を参照。
- dummy 値や `TF_VAR_*` 注入で plan を強制しない（承認材料にならない）。

---

## 4. なぜ `alb_internal = true` が必須か

`alb_internal` の default は `false`（internet-facing）。**false のままでは P2 の差分が plan に一切現れない**:

- `module.ecs.aws_lb.main` の `internal` は false のまま、subnets も public のまま -> ALB の replace が出ない。
- `module.ecs.aws_cloudfront_vpc_origin.alb` は `count = var.alb_internal ? 1 : 0` のため 0 のまま -> VPC Origin の create が出ない。

承認は P2 差分を含む full plan で行う必要があるため、実 tfvars で `alb_internal = true` を設定する。

---

## 5. 差分レビュー（expected / unexpected）

### 5.1 期待される P2 差分（`alb_internal` false -> true）

| リソース | 期待アクション | 理由 |
|----------|---------------|------|
| `module.ecs.aws_lb.main` | **replace**（destroy -> create） | `internal=true` + subnets を public -> private に変更（いずれも ForceNew） |
| `module.ecs.aws_lb_listener.http` | **replace** | 親 ALB の arn が変わるため再作成 |
| `module.ecs.aws_cloudfront_vpc_origin.alb[0]` | **create** | `count` 0 -> 1 |
| `module.security.aws_security_group.alb` の ingress | **in-place update** | `alb_internal=true` で ①VPC CIDR 80/443（intra-VPC 用）+ ②**service SG 参照の port 80 ingress（CloudFront 疎通に必須・§6.1）**。`false` では `0.0.0.0/0` 80/443 |
| outputs（alb dns / vpc_origin_id 等） | 値の変化 | ALB 再作成・VPC Origin 新設に伴う |

- `aws_lb_listener.https[0]` は `alb_certificate_arn` が空のとき `count=0` で plan に出ない。証明書を設定済みの場合のみ replace。
- ALB SG ingress の許可元は `alb_internal` に連動する。`false -> true` で `0.0.0.0/0` が VPC CIDR に絞られ、さらに **service-managed SG 参照の port 80 ingress が追加される**（CloudFront → ALB 疎通に必須。VPC CIDR だけでは疎通しない＝§6.1）。いずれも絞り込み方向で広域開放ではない。`alb_internal=false` のままなら SG 差分は出ない。

### 5.2 変化しないことを確認するリソース

以下が差分に出たら **要調査**（この flip では変わらないはず）:

- RDS（`aws_db_instance`）、ECR、ECS cluster / service / task
- VPC / subnet / route table、fck-nat インスタンス、IAM role / policy
- ECS SG / RDS SG（ALB SG の ingress 80/443 のみ §5.1 の期待差分。ECS SG・RDS SG は変化しないはず）

### 5.3 unexpected / destructive drift — STOP 条件

| 兆候 | 対応 |
|------|------|
| RDS の replace / destroy | **STOP**。承認しない（データ損失リスク） |
| VPC / subnet / route table の destroy / recreate | **STOP** |
| ECS cluster / ECR repository の destroy | **STOP** |
| IAM role / policy の delete | **STOP** |
| SG ingress に `0.0.0.0/0` 等の**広域開放**（絞り込みではなく開放方向） | **STOP**。※ ALB SG の `0.0.0.0/0` -> VPC CIDR は絞り込み方向で期待差分（§5.1）。逆向き（VPC CIDR -> `0.0.0.0/0` 等）が出たら STOP |
| §5.1 の期待集合（ALB / http listener / VPC Origin）以外の destroy | **STOP**・原因調査 |

期待集合のみが diff に出ていれば承認可。1 つでも STOP 条件に該当したら承認せず原因を切り分ける。

---

## 6. CloudFront distribution は Terraform 管理外

- `aws_cloudfront_distribution` リソースは Terraform に存在しない（手動作成・管理）。よって **distribution 自体の変更は plan に現れない**。
- ALB が replace されると ALB の DNS が変わる。apply 後、CloudFront distribution のオリジンを VPC Origin に**手動で切り替える**必要がある（`module.ecs` の `vpc_origin_id` output を使用）。切替前に現 distribution config + ETag を控える（rollback 用）。
- apply 後、AWS が `CloudFront-VPCOrigins-Service-SG` を自動生成する。**この service-managed SG を source 参照する ALB SG の port 80 ingress は CloudFront → internal ALB 疎通に【必須】**（任意の強化ではない）。

> ### ⚠️ 6.1 VPC Origin 疎通の要点（2026-06-18 live 検証で確定）
>
> **VPC CIDR(`10.0.0.0/16`)許可だけでは CloudFront → internal ALB は疎通しない。**
> VPC Origins の ENI は VPC CIDR 内に居る（例 10.0.11.x / 10.0.12.x）が、ALB SG を CIDR で許可しても
> CloudFront からのトラフィックは **ALB に全く到達せず**（ALB `RequestCount=0` / `/health` が 504 Gateway Timeout 30s）。
> AWS 自動生成の `CloudFront-VPCOrigins-Service-SG` を **source security group として参照**する port 80 ingress を
> ALB SG に追加して初めて疎通する。
>
> - Terraform 管理: `modules/security/main.tf` の `data.aws_security_group.vpc_origins_service`
>   （`name = "CloudFront-VPCOrigins-Service-SG"` を `alb_internal` gated で参照）+ dynamic inline ingress。
>   data source は VPC Origin リソースへの依存 edge を持たないため循環依存にならない。
> - **fresh 構築時の制約**: service SG は VPC Origin 作成後に AWS が生成するため、ゼロからの新規構築では
>   VPC Origin 作成後の 2 段階 apply が必要（VPC Origins の chicken-and-egg）。
> - 切り分け: `/health` 504 + ALB `RequestCount=0` + target healthy なら「CloudFront → ALB 未到達」、
>   ALB `RequestCount>0` + 5xx なら「ALB → ECS 側」。
> - **最終結果（2026-06-18）**: 上記 service-SG ingress 追加で `https://api.stg.noah-karte.com/health` = **200**
>   （`{"status":"ok"}`）。Terraform 恒久反映 apply 済で `terraform plan` = **No changes**（state ↔ live SG 一致）。

---

## 7. 安全ルール（承認・適用）

- 承認は対象を絞らない full plan（`terraform plan -out=tfplan`）の結果のみで行う。scoped target plan はデバッグ専用で承認根拠にしない。
- 適用ステップは承認者が plan を確認し明示承認した後に手動実行する。preflight script や Claude Code は適用を自動実行しない。
- `terraform.tfvars` に秘密値を置いたまま commit しない（gitignore 済み）。`tfplan` も git 追跡しない（root `.gitignore`）。

---

## 8. 次のアクション

1. §3 に従い `infra/terraform/terraform.tfvars` を用意（`db_password` + `alb_internal = true`）。
2. 未初期化なら `(cd infra/terraform && AWS_PROFILE=AnimalEkarte terraform init)`。
3. `sh scripts/infra-terraform-plan-preflight.sh` を実行し `tfplan` を生成。
4. §5 の表で差分を分類し、期待集合のみであることを確認。
5. 承認後、承認者が手動で適用 -> CloudFront distribution のオリジンを手動切替（§6）。
