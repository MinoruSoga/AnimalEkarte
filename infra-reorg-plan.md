# インフラ再編計画（コード・ドキュメントの整理と STG/PROD 分離）

> **起案**: 2026-07-20（責任者: PO）／**目的**: ① AWS 廃止後の死骸を除去 ② Cloudflare のコード/ドキュメントを STG・PROD で明確に分離 ③ IaC 運用のベストプラクティス（リモート state・ドリフト検知・トークン分割）を導入する。
> **前提**: STG は Cloudflare へ完全移行済み・AWS STG は 2026-07-20 に destroy 済み（`migration-cloudflare.md` Phase 8）。本番（#253）は未構築。

## 0. 現状の問題（実測 2026-07-20）

| 領域 | 問題 |
|---|---|
| **AWS コードの死骸** | `infra/terraform/`(25 tf) / `infra/terraform-bootstrap/` / `infra/ecs/` は**環境破棄済みで適用不能な死にコード**。保守・混乱コスト（②削除対象そのもの） |
| **AWS ドキュメントの陳腐化** | `docs/ops/infra-architecture.md`(AWS ECS/RDS/ALB 構成のまま) / `stg-aws-change-readiness.md` / `stg-aws-cost-reduction.md` / `p2-terraform-plan-runbook.md` は**存在しない環境を説明している** |
| **env 分離が不完全** | `infra/cloudflare/`(フラット=暗黙 STG) と `infra/cloudflare/production/`(同一ファイルのコピペ) → **DRY 違反・ドリフトの温床**。どちらが STG か命名で分からない |
| **tfstate がローカル backend** | `infra/cloudflare/backend.tf` が「当面 local・R2 backend は TODO」。チーム/CI 運用で破綻（ロックなし・マシン依存・状態散逸） |
| **移行記録が root 直置き** | `migration-cloudflare.md`（完了した移行の実施記録）が root にあり現行ドキュメントと混在 |
| **docs 集約先の二重化** | `docs/infra/`（空）と `docs/ops/`（実体）が併存。SSOT が曖昧 |

## 1. 目標構造（To-Be）

### 1.1 コード（env-first・module 共有）

```
infra/
  cloudflare/
    modules/                 # DRY: STG/PROD 共有の再利用モジュール
      zone/                  #   zone + DNS レコード
      r2/                    #   R2 バケット
      hyperdrive/            #   Hyperdrive config
      notifications/         #   通知ポリシー
    envs/
      staging/               # 薄いルート: backend.tf(R2 state key=stg) + main.tf(module呼出) + staging.auto.tfvars
      production/            # 薄いルート: backend.tf(key=prod) + production.auto.tfvars
    README.md                # 何がどこにあるか・apply 手順
  scripts/                   # CF/pscale 運用スクリプトのみ残す（AWS 専用スクリプトは削除）
```

- **AWS 系（`infra/terraform` / `terraform-bootstrap` / `infra/ecs`）は削除**（git 履歴が保存する。適用不能な IaC を残さない）。tfstate S3 バケット + DynamoDB ロックも空になったので合わせて撤去
- **Wrangler**: `backend/wrangler.jsonc`（STG）→ `wrangler.staging.jsonc` にリネームし `wrangler.production.jsonc` と対称化（どちらの env か一目で分かる）。Wrangler ネイティブ環境（1ファイル `env.*` ブロック）も選択肢だが、**別ファイルの方が env 分離が明示的**なので採用

### 1.2 ドキュメント（env-first・SSOT を docs/ops/infra へ）

```
docs/ops/infra/
  README.md                  # インデックス（現行構成の入口）
  architecture.md            # 現行 Cloudflare 構成の全体像（env 共通・AWS 記述を完全除去）
  iac-guidelines.md          # IaC 運用規約（Terraform/Wrangler 境界・state・token・ドリフト）
  staging/
    runbook.md               # STG 運用手順（デプロイ・migrate・seed・障害対応）
  production/
    setup.md                 # 本番構築手順（既存 docs/ops/deploy/PRODUCTION_CF_SETUP.md を移設）
    runbook.md               # 本番運用手順（構築後に埋める）
  _archive/
    migration-cloudflare.md  # 完了した STG 移行の実施記録（root から移設・凍結）
    aws-legacy/              # stg-aws-*.md・infra-architecture(AWS版)・p2-runbook 等（廃止環境の記録）
```

- 原則: **「現行を説明する doc」と「完了した作業の記録」を物理的に分ける**。後者は `_archive/` で凍結し、現行 doc から参照しない
- `docs/infra/`（空）は廃止し `docs/ops/infra/` に一本化

## 2. 実施フェーズ（順序厳守・各フェーズ独立コミット）

### Phase A — 死骸の除去（②削除・最優先・低リスク）
1. `infra/terraform/` `infra/terraform-bootstrap/` `infra/ecs/` を削除（AWS 環境は破棄済み・適用不能）
2. `infra/scripts/` の AWS 専用スクリプト（`stg-db-tunnel.sh` 等）を削除・CF/pscale 系は残す
3. AWS-era docs を `docs/ops/infra/_archive/aws-legacy/` へ移動（削除ではなく凍結 — 過去の意思決定記録の価値があるため）
4. tfstate S3 バケット（`animalekarte-tfstate-698109622668`）+ DynamoDB ロックの撤去（**AWS リソース削除 = 要 PO 承認**。Phase A の最後に単独で）
- **検証**: `grep -rn "infra/terraform\|infra/ecs" .github/ Makefile docs/` で参照残がないこと（死んだ参照を残さない）

### Phase B — Cloudflare コードの env 分離 + module 化 + リモート state
1. `infra/cloudflare/*.tf` の共通部分を `modules/{zone,r2,hyperdrive,notifications}/` へ抽出
2. `envs/staging/` を新設し現行 STG を module 呼出の薄いルートへ移行（**state を壊さないため `terraform state mv` で移設・作り直さない**）
3. `envs/production/` を同 module から再構成（既存 `production/` のコピペを置換）
4. **R2 リモート backend 化**: `backend.tf` を R2（S3 互換）backend に。state key を env で分離（`stg/terraform.tfstate` / `prod/terraform.tfstate`）
- **重要**: この Phase は**本番構築（#253）と統合して実施する**のが効率的。単独リファクタ pass を切らず、prod を建てるついでに module 化する（③: 既にやる作業に畳み込む）
- **リスク**: state 移設ミスで既存 STG リソースを terraform が「再作成」しようとする → 必ず `terraform plan` が「0 to add/change/destroy」になることを確認してから apply

### Phase C — ドキュメントの env 分離・SSOT 化
1. `docs/ops/infra/` を新設し上記構造へ再配置
2. `migration-cloudflare.md`（root・完了記録）→ `_archive/` へ移動
3. `architecture.md` を現行 CF 構成で新規作成（AWS 記述を持ち込まない）
4. `iac-guidelines.md` に運用規約を明文化（下記 §3）
5. root の infra 系 md（`migration-cloudflare.md`）を撤去し、残す root doc を最小化

### Phase D — 運用ガードレール（本番構築前に必須）
1. **ドリフト検知 CI**: 日次 `terraform plan` を GitHub Actions で回し、drift（ダッシュボード手動変更）を検知して通知。**今日の CloudFront 手動リソースの依存地獄は、これがあれば事前に見えていた**
2. **API トークン分割**: Terraform 用（zone/account edit）と CI deploy 用（Workers Scripts edit のみ）を別発行・最小権限。本番トークンは STG と別
3. **provider/wrangler バージョン pin** の明文化（Cloudflare provider v4→v5 破壊的変更対策）
4. **例外操作ログの徹底**: 手動ダッシュボード操作は必ず記録し `cf-terraforming import` で state に取り込む運用を規約化

## 3. IaC 運用規約（iac-guidelines.md の骨子）

- **2 層の境界**: Terraform = ゾーン/DNS/R2/Hyperdrive/WAF/通知/トークン（土台）。Wrangler = Worker/Container/ルート/secrets（デプロイ）。**Workers を Terraform で管理しない**
- **state**: リモート（R2）・env 別 key・ロック有効・git 管理禁止・平文 secret を state に載せない
- **secrets**: `wrangler secret put` or GitHub Secrets 経由。`wrangler.jsonc` の `vars` は非機密のみ。Terraform リソースに secret を置かない
- **token**: 用途別・最小権限・env 別・ローテーション前提
- **drift**: 定期 plan で検知。手動操作は即 import か例外ログ
- **apply**: 共有 env はローカル apply 禁止・CI で plan(PR)→apply(merge・承認ゲート)

## 4. 実施順の推奨

**Phase A（死骸除去）→ Phase C（docs 整理）→ Phase D（ガードレール）→ Phase B（module 化・本番構築 #253 と統合）**

理由: A と C は**低リスクの純削除・整理で即やる価値がある**（今日の破棄直後が最適）。D は本番を建てる前の必須整備。B（module 化）は本番構築と一体でやるのが最も効率的で、単独の大リファクタ pass を避けられる（③）。

## 5. リスク・非対象

- **リスク**: Phase B の `terraform state mv` は最も慎重を要する（plan=0 変更の確認が絶対条件）。tfstate バケット撤去（Phase A-4）は AWS リソース削除のため PO 承認必須
- **非対象**: `infra/scripts` の CF/pscale 運用スクリプトの中身の書き換え（配置整理のみ）／本番の実構築そのもの（#253 で別途）／PlanetScale の所有権 REASSIGN（AWS とは別件のサポートチケット）

## 6. 実践ゲート（PRODUCT_PHILOSOPHY）

- **① 責任者**: PO（2026-07-20・業務目的 = インフラの保守性向上と env 事故の防止）
- **② 削除**: AWS 死にコード 25+tf・陳腐化 docs・コピペ tf・空 docs/infra — 純削除が本計画の中心
- **④ メトリクス**: 「どの env のどのリソースか」を特定するまでの手数、ドリフト検知までの時間（手動発覚 → 日次自動検知）
