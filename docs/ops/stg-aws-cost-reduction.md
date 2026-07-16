# STG AWS コスト削減・リソース改善レポート

> **時点スナップショット**: 2026-06-17 時点の調査レポート。改変不可（P1/P2実施完了後はarchiveへ移動想定）。

> **Animal Ekarte**: STG 環境の AWS コスト最適化調査と改善候補の整理
> **最終更新**: 2026-06-17 | **対象環境**: Staging (us-east-1)
> **調査方式**: read-only (AWS CLI describe/get/list + Cost Explorer のみ。AWS リソース変更・削除なし)

---

## 1. 目的と前提

### 目的

STG 環境の AWS コストを read-only で調査し、不要・過剰・代替可能なリソースを特定する。
ALB を中心に、現行アーキテクチャの必要性と代替案を評価する。

### 前提と制約

| 項目 | 内容 |
|------|------|
| 調査期間 | 2026-06-01〜2026-06-17（16日間実績 + 月額推定） |
| 調査対象 | STG 環境のみ (us-east-1) |
| 操作範囲 | **read-only のみ**。削除・変更は別途承認後に実施 |
| 除外 | Production 環境、PR 作成、Terraform 適用、AWS Support 問い合わせ |

---

## 2. 現行アーキテクチャ（確認済み）

### 2.1 構成概要

```
インターネット (HTTPS)
    ↓
CloudFront  [api.stg.noah-karte.com]  ← ACM 証明書付き
    ↓ HTTP:80（CloudFront がオリジンに HTTP で接続）
ALB  [animalekarte-stg-alb]  internet-facing, 2 AZ (us-east-1a / us-east-1b)
    ↓ HTTP:8080（ALB SG からの接続のみ許可）
ECS Fargate  [animalekarte-stg-service]  Private Subnet, assign_public_ip=DISABLED
    ↓ TCP:5432
RDS PostgreSQL 18  [animalekarte-stg-db]  db.t4g.micro, 20GB gp3

fck-nat EC2  [t4g.nano, running]  ← Private Subnet の 0.0.0.0/0 ルート先
```

### 2.2 重要な構成事実

- ECS タスクは **Private Subnet** に配置。インターネットから直接到達不可。
- ECS タスクの Security Group は **ALB SG からの 8080 ポートのみ**を受信許可。
- CloudFront → ALB → ECS が **唯一のインターネット経路**。
- `backend-deploy.yml` の `wait-for-service-stability` は ALB Target Group のヘルスチェックに依存。
- ECS Cluster: Fargate Spot weight=4 / Fargate on-demand weight=1（コスト最適化済み）。
- Container Insights: `disabled`（コスト最適化済み）。
- 夜間スケジューラ: 22:00〜08:00 JST に ECS desiredCount=0 / RDS 停止（**ALB は停止しない**）。
- インフラは `./infra/terraform/` で Terraform 管理（ALB は ECS module 内に定義）。

---

## 3. STG 月額費用（推定）

> 調査期間 16 日間の実績から推定。月額 = 実績 × 30/16。

| サービス | 16日実績 | 推定月額 | 割合 | 分類 |
|---------|---------|---------|-----|------|
| **ALB (ELB)** | $8.44 | **$15.83** | **36%** | 必須（後述） |
| VPC Public IPv4 | $5.04 | $9.46 | 22% | 現行構成では必須（P2で削減候補） |
| RDS (db.t4g.micro) | $4.92 | $9.23 | 21% | 必須 |
| EC2 (fck-nat t4g.nano) | $1.56 | $2.93 | 7% | 必須 |
| ECS Fargate | $1.02 | $1.92 | 4% | 必須 |
| EC2 Other (EIP関連) | $0.18 | $0.33 | 1% | 必須 |
| ECR | $0.01 | $0.03 | <1% | 必須 |
| CloudFront | $0 | $0（※） | — | 現時点で $0（使用量次第で変動） |
| WAF | $0 | $0 | — | 未使用 |
| **合計（税前）** | **$21.17** | **$39.73** | — | — |

**コスト構造の特徴**:

- ALB の `LoadBalancerUsage` が全体の 36%。夜間 ECS 停止中も 24 時間課金が継続する。
- VPC Public IPv4 費用（推定 $9.46/月）の内訳: ALB 2IP + fck-nat 1IP = 計 3 パブリック IP。
- ECS コストが低いのは Fargate Spot が大半（weight=4）のため（既最適化済み）。
- RDS は夜間停止により実際の課金は約 14 時間/日に抑制されている。
- ※ CloudFront は現時点の Cost Explorer では $0。ただし STG のリクエスト量が増えると課金が発生する場合がある。なお CloudFront VPC Origin **機能自体**の追加費用はない（後述 §6.3）。

---

## 4. ALB の評価

### 4.1 現行構成での ALB 依存関係

| 依存元 | 種別 | ALB 削除時の影響 |
|--------|------|----------------|
| ECS service `loadBalancers` 設定 | ハード | ECS へのトラフィック不可 |
| ECS タスク SG（ALB SG 経由のみ 8080 許可） | ハード | ECS タスクへの到達不可 |
| CloudFront オリジン（ALB の DNS を参照） | ハード | API 外部アクセス全断 |
| `backend-deploy.yml` wait-for-service-stability | ハード | デプロイ確認がタイムアウト |
| Terraform: `depends_on = [aws_lb_listener.http]` | Terraform | ALB なしで ECS 作成不可 |
| `stg-smoke.yml` (`api.stg.noah-karte.com` 経由) | ソフト | スモークテスト失敗 |

### 4.2 ALB の評価結論

| 観点 | 結論 |
|------|------|
| **現行構成での必要性** | **必須**。ECS Private Subnet + SG 制約により代替手段なし |
| **STG 要件としての必要性** | **代替可能**。構成変更（後述 §6.3）により ALB を内部化できる |
| **単純削除の可否** | **不可**。ECS 疎通・デプロイ・スモークテストが全断する |
| **コスト停止の可否** | **不可**。ALB に「停止」状態はなく、削除しか選択肢がない |

---

## 5. リソース分類表

| リソース | 推定月額 | 分類 | 根拠 |
|---------|---------|------|------|
| ALB (internet-facing) | $15.83 | **必須**（代替可能） | CloudFront→ECS の唯一の経路（確認済み） |
| VPC Public IPv4 × 3 | $9.46 | **現行構成では必須**（P2で削減候補） | ALB 2IP + fck-nat 1IP、全て使用中（確認済み）。P2 で ALB internal 化すると ALB 分 2IP を削減可能 |
| RDS db.t4g.micro | $9.23 | **必須** | STG DB として稼働中 |
| fck-nat EC2 t4g.nano | $2.93 | **必須**（代替検討可） | Private Subnet 唯一の NAT 出口（RT 確認済み） |
| ECS Fargate (Spot) | $1.92 | **必須** | 最小構成・Spot 最適化済み |
| EC2 Other | $0.33 | **必須** | EIP 保持コスト |
| ECR | $0.03 | **必須** | コンテナイメージ格納 |
| ap-northeast-1 EIP（未関連） | 推定 ~$1.65 | **不要候補** | 関連リソースなし（確認済み）→ §6.2 |

---

## 6. 削減候補と実施計画

### 6.1 分類サマリー

| 案 | 推定削減額 | 分類 | 優先度 |
|---|-----------|------|--------|
| ap-northeast-1 EIP 解放 | ~$1.65/月 | ✅ 実行可能（承認後） | P1 |
| internal ALB + CloudFront VPC Origin | ~$7.20/月 | 🟡 要追加確認 | P2 |
| ap-northeast-1 t3.micro 遅延請求の終了 | ~$4.49/月 | ⏳ 自然消滅待ち | P3 |
| fck-nat の代替検討 | ~$2.93/月 | 🟡 要追加確認（長期） | P4 |

---

### 6.2 [P1] ap-northeast-1 EIP 解放（実行可能）

**根拠**: ap-northeast-1 に未関連 EIP が存在（インスタンス・ENI ともに紐付けなし確認済み）。
STG リソースは us-east-1 に集約されており、ap-northeast-1 に保持する理由がない。

> ⚠️ **重要な注意事項**: EIP を解放すると**同一 IP アドレスの復元は保証されない**（AWS プールに返却される）。
> 解放前に必ず以下を確認すること:
>
> - この IP を参照している DNS レコードや外部 allowlist がないか
> - CI/CD・監視設定・Webhook 設定に IP がハードコードされていないか

**実施方法**（承認後に実施）:

```bash
# 1. まず ap-northeast-1 の EIP を確認
aws --profile AnimalEkarte --region ap-northeast-1 ec2 describe-addresses

# 2. 問題なければ解放（AllocationId は上記コマンドで確認）
aws --profile AnimalEkarte --region ap-northeast-1 ec2 release-address \
  --allocation-id <AllocationId>
```

**ロールバック**: 解放後の同一 IP への復元は不可。新 EIP 取得は可能だが別 IP になる。

**参考**: [EIP の動作と解放](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/elastic-ip-addresses-eip.html)

---

### 6.3 [P2] internal ALB + CloudFront VPC Origin（要追加確認）

**概要**:

現在の ALB は `internet-facing`（パブリック IP 2 個）。
CloudFront VPC Origin を使うと ALB を `internal`（パブリック IP 不要）に変更できる。

```
変更後アーキテクチャ:
CloudFront (api.stg.noah-karte.com)
  ↓  CloudFront VPC Origin（AWS バックボーン経由、追加費用なし）
internal ALB  ← パブリック IP 不要
  ↓
ECS Fargate (Private Subnet) ← 変更なし
```

**期待効果**:

| 項目 | 現状 | 変更後 |
|------|------|--------|
| ALB パブリック IP | 2 個 | 0 個 |
| VPC Public IPv4 コスト（ALB 分） | 推定 $7.20/月 | $0 |
| ALB hourly コスト | $15.83/月 | $15.83/月（変化なし） |
| **月額削減見込み** | — | **〜$7.20/月** |
| CloudFront VPC Origin **機能**の追加費用 | — | $0（機能自体は無料） |

**必要な事前確認**:

- [ ] CloudFront VPC Origin が us-east-1 で正常利用可能であること（GA 済みだが動作検証要）
- [ ] `scheme = internal` への変更は ALB **再作成（replace）**が必要（Terraform `plan` で事前確認）
- [ ] ALB 再作成中は CloudFront からの到達不可 → ダウンタイムあり（夜間実施を推奨）
- [ ] SG: Phase 1 として ALB Security Group の inbound 許可元を **VPC CIDR**（`var.vpc_cidr`）に絞る。VPC Origin の CloudFront トラフィックは VPC Origins ENI 経由で VPC 内から到達するため。許可元は Terraform で `alb_internal` に連動し、ALB の internal 化と同一 apply で `0.0.0.0/0` → VPC CIDR に切り替わる（public CloudFront managed prefix list は internal ALB には使わない）。さらに絞り込む場合は AWS が自動生成する `CloudFront-VPCOrigins-Service-SG` を許可元にする Phase 2 を実施する（詳細は STG_AWS_CHANGE_READINESS.md §3.2）
- [ ] 移行後に stg-smoke.yml（`/health`）で疎通確認

**移行手順（承認後・概要）**:

1. Terraform で internal ALB + CloudFront VPC Origin 設定を追加
2. `terraform plan` で影響確認（ALB replace の確認）
3. 夜間にメンテナンス窓を設けて `terraform apply`
4. stg-smoke.yml でヘルスチェック確認
5. 移行後に Cost Explorer で VPC Public IPv4（ALB 分 2IP 相当）の課金が消えることを確認（internal ALB には public IP が割り当てられないため自動的に解消される）

**参考リンク**:
- [CloudFront VPC Origins 公式ドキュメント](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-vpc-origins.html)
- [CloudFront VPC Origins 発表（追加費用なし）](https://aws.amazon.com/blogs/aws/introducing-amazon-cloudfront-vpc-origins-enhanced-security-and-streamlined-operations-for-your-applications/)
- [ALB scheme の変更仕様（replace 必須）](https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-elasticloadbalancingv2-loadbalancer.html)

---

### 6.4 [P3] ap-northeast-1 t3.micro 遅延請求の終了（自然消滅待ち）

インスタンス実体は存在しない。過去に削除済みのリソースに対する遅延請求（推定 ~$4.49/月）が発生中。
月末には自然に終了する見込み。追加アクション不要。

---

### 6.5 [P4] fck-nat の代替検討（長期・要追加確認）

**現状**: Private Subnet のルートテーブルが `0.0.0.0/0 → fck-nat EC2` を参照している。
ECS タスクが ECR/CloudWatch/SSM Parameter Store に接続するために必要。単純削除は不可。

**将来の代替候補**（いずれも調査フェーズ）:

| 代替案 | 推定月額 | メリット | デメリット |
|--------|---------|---------|----------|
| fck-nat 現状維持 | $2.93 | 安定稼働中 | 停止時 ECS 通信断 |
| VPC Endpoint（ECR・CW・SSM 個別）| $7〜15 | NAT 不要 | エンドポイント数に応じコスト増 |
| ECS Public Subnet 移行 | $0（NAT不要） | NAT 費用ゼロ | SG 設計変更・セキュリティ再評価 |

> P2 の internal ALB 検討後に優先度を評価する。短期の削減優先度は低い。

---

## 7. 不可案（仕様・アーキテクチャ制約により対象外）

以下はコスト削減のために提案されることがあるが、**技術的に実施不可または実施後に問題が発生する**ため推奨しない。

| 案 | 判定 | 理由 |
|----|------|------|
| ECS Fargate 128 CPU | **不可** | Fargate の有効 CPU 値は 256/512/1024/2048/4096 のみ。128 は非存在 |
| RDS db.t4g.nano への変更 | **不可** | ① AWS API で当該 engine/region における有効クラスとして確認できない限り推奨不可 ② 実測 FreeableMemory 0.13〜0.87 GB に対し nano 上限 0.5 GB では余裕がない |
| RDS ストレージ 10GB への縮小 | **不可** | gp3 最小は 20 GB。RDS はストレージ縮小非対応 |
| ALB 単純削除 | **不可** | ECS SG・CloudFront 経路・デプロイ依存チェーンが全断。代替構成が先行必要 |
| fck-nat 単純削除 | **不可** | Private Subnet RT が fck-nat 依存。削除で ECS の ECR/CW/SSM 接続が全断 |

---

## 8. 実施順序と承認ポイント

```
[今すぐ確認] ap-northeast-1 EIP の DNS/allowlist 参照有無を確認
    ↓ 参照なし確認後・承認後
[P1] ap-northeast-1 EIP 解放  → ~$1.65/月 削減
    ↓
[P2 事前調査] CloudFront VPC Origin 利用可否確認
              Terraform plan で ALB replace 影響確認
    ↓ 影響確認 + 承認後 + ダウンタイム合意後
[P2] internal ALB + CloudFront VPC Origin 移行  → ~$7.20/月 追加削減
    ↓
[P4] fck-nat 代替案の調査（VPC Endpoint vs Public Subnet 移行）
```

---

## 9. リスクまとめ

| リスク | 対象施策 | 影響 | 緩和策 |
|--------|---------|------|--------|
| EIP 解放後に同一 IP を要求された場合 | P1 | 外部サービスの接続断 | 解放前に DNS/allowlist 確認 |
| ALB scheme 変更（replace）中のダウンタイム | P2 | STG API 一時断 | 夜間実施 + スモークテスト確認 |
| CloudFront VPC Origin 設定誤り | P2 | CF → ALB 経路断 | staging distribution で検証後に適用 |
| Cost Explorer 数値の推定誤差 | 全体 | 月額見込みとの乖離 | 実施後に翌月コストを再確認 |

---

## 10. 参考リンク

| 内容 | URL |
|------|-----|
| CloudFront VPC Origins 公式ドキュメント | https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-vpc-origins.html |
| CloudFront VPC Origins 発表（追加費用なし）| https://aws.amazon.com/blogs/aws/introducing-amazon-cloudfront-vpc-origins-enhanced-security-and-streamlined-operations-for-your-applications/ |
| ALB scheme 変更仕様（replace 必須）| https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-elasticloadbalancingv2-loadbalancer.html |
| EIP の動作と解放 | https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/elastic-ip-addresses-eip.html |
| Fargate タスクサイズ有効値 | https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html |
| RDS gp3 ストレージ仕様 | https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html |

---

## 付録: 今回の調査で実行した操作（read-only のみ）

変更・作成・削除系コマンド（`modify-*` / `update-*` / `delete-*` / `create-*` / `release-*` / `terminate-*`）は**一切実行していない**。

実行した read-only コマンド（カテゴリ）:

```
describe-load-balancers / describe-listeners / describe-target-groups / describe-target-health
describe-instances / describe-addresses / describe-network-interfaces
describe-security-groups / describe-subnets / describe-route-tables
describe-db-instances / ecs describe-services / ecs list-services
cloudwatch get-metric-statistics / acm describe-certificate / acm list-certificates
cloudfront get-distribution / cloudfront list-distributions
wafv2 list-web-acls
ce get-cost-and-usage
```
