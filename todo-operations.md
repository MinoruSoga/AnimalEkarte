# 運用・外部実行の受入記録（現在のタスク状態はPlane）

> Current task state is in Plane. Local task details were migrated 2026-09-23; see [crosswalk](docs/work/plane-md-migration-20260923-receipt.md).

最終タスク移行: 2026-09-23。未完了の環境・データ・本番・納品作業と現在状態はPlaneを正本とする。下記の実行条件・過去receiptは運用根拠として保持する。

remote・CI・配備・Linear・STG/PROD の DB/秘密/投入 receipt は今回再照会していない。以前の「未構築」「未受領」を現在の事実として断定せず、再実行前に実施有無を確認する。ローカル準備票の作成は実行完了ではない。

運用全体の着手プランは 2026-09-21 に再点検し、統合済みの LINMIG / remaining 準備票と現行コードを照合した。既存票を作り直す段階ではなく、不足入力・対象環境の適用状態・実行証拠を揃える段階。各表の ID から下の個別手順を参照する。承認前にも、既存資料の照合・不足入力表・実行案の作成は進められる。外部の実操作を、この計画の記載だけで開始しない。

<a id="readiness-preparation"></a>

## 今着手する運用準備

次の担当は運用/producer/USER。個人の割当は未確定であり、実行票で確定する。実行票は `ID / 対象環境・医院 / revision・manifest等の入力識別 / 読取・変更する範囲 / 前段証拠 / 操作者・承認者 / 実行枠 / 中止条件 / 復旧 / 証拠保存先` を持つ。秘密・実 roster・患者情報は repo 外の承認済み保管先に置き、台帳には非機密参照だけを残す。既存 receipt が見つかれば同一対象・入力との一致を先に検証し、再実行の要否を決める。

| ID | 準備可: ローカルで作る成果物 | 外部実行を開始する条件 |
|---|---|---|
| `UAT-Q3-GENDER-MAP` → `MIG-16` | Planeへ移行済み |  |
| `UAT-Q2-VACCINE-SPECIES` → `EMR-106` | Planeへ移行済み |  |
| `UAT-Q4-UNPAID-TRIAGE` → `EMR-107` | Planeへ移行済み |  |
| `PO-PET-DECEASED-DATA-BACKFILL` → `EMR-108` | Planeへ移行済み |  |
| `BUG-LOCAL-HANDOFF-CSV-CONTRACT` → `MIG-19` | Planeへ移行済み |  |
| `SLACK-HAC-IMPORT` → `MIG-17`, `H0-2 / HAC-CSV-1` → `MIG-20` | Planeへ移行済み |  |
| `H0-3b / H1-2` → `MIG-21` | Planeへ移行済み |  |
| `AE-STG-UAT-LANE3-HAC` → `MIG-22` | Planeへ移行済み |  |
| `SLACK-ACCESS` → `EMR-111`, `H3-9` → `EMR-144` | Planeへ移行済み |  |
| `H3-9` → `EMR-144`, `H3-11` → `EMR-145` | Planeへ移行済み |  |
| `H3-11` → `EMR-145`, `Lane 4` → `EMR-146` | Planeへ移行済み |  |
| `P1 / SEC-SECRETS-5 / #89 / #97` → `EMR-147` | Planeへ移行済み |  |
| `P2 / #253 / PROD-SETUP` → `EMR-148` | Planeへ移行済み |  |
| `P3 / #250 PROD-DATA-MIGRATION` → `MIG-23` | Planeへ移行済み |  |
| `P5 / #255 STAFF-PROVISION` → `EMR-149` | Planeへ移行済み |  |
| `P1 / SEC-SECRETS-5 / #89 / #97` → `EMR-147`, `P2 / #253 / PROD-SETUP` → `EMR-148`, `P6 / #258 / U1-U12 DELIVERY` → `EMR-150` | Planeへ移行済み |  |
| `P7 / #256 / TRAINING` → `EMR-151` | Planeへ移行済み |  |

準備は不足票・実行案を同じ ID に残した時点で一区切り。2026-09-21 時点で上表からリンクした準備票は main に統合済み。ただし、旧IDの担当分界や実行枠などの未確定欄まで埋まったことを意味しない。実行条件を満たさない行は BLOCKED/UNKNOWN を維持する。必要な入力を揃える作業と、実環境を変更する作業を同じ「着手済み」にまとめない。

<a id="uat-data-operations"></a>

## 医院 UAT に伴うデータ操作

| ID | 残作業・担当 | 開始条件と完了条件 |
|---|---|---|
| `UAT-Q3-GENDER-MAP` → `MIG-16` | Planeへ移行済み |  |
| `UAT-Q2-VACCINE-SPECIES` → `EMR-106` | Planeへ移行済み |  |
| `UAT-Q4-UNPAID-TRIAGE` → `EMR-107` | Planeへ移行済み |  |
| `PO-PET-DECEASED-DATA-BACKFILL` → `EMR-108` | Planeへ移行済み |  |
| `BUG-LOCAL-HANDOFF-CSV-CONTRACT` → `MIG-19` | Planeへ移行済み |  |

実行手順は [ローカル handoff](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)、[STG 停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。共有環境への書込み・再取込・DB 作成/破棄・migration は今回実施しない。必要な `make migrate` は、対象環境の適用状態を確認したうえでユーザーが実行する。

<a id="billing-schema-readiness"></a>

### UAT-R2-EXCLUSIVE-LOCK: 会計競合防御のDB適用確認

> Task detail migrated to Plane `EMR-85` and verified by readback. Historical/evidence material remains in linked source records.

## STG データレーンの残り

| ID / 段階 | 現在の判定 | 次の一手 |
|---|---|---|
| `SLACK-HAC-IMPORT` → `MIG-17`, `H0-2 / HAC-CSV-1` → `MIG-20` | Planeへ移行済み |  |
| `H0-3b / H1-2` → `MIG-21` | Planeへ移行済み |  |
| `AE-STG-UAT-LANE3-HAC` → `MIG-22` | Planeへ移行済み |  |
| `H3-9` → `EMR-144` | Planeへ移行済み |  |
| `H3-11` → `EMR-145` | Planeへ移行済み |  |
| `H3-11` → `EMR-145`, `Lane 4` → `EMR-146` | Planeへ移行済み |  |

同一の既知破損 BAK の再復元や、既存 STG の無条件上書き load は行わない。9月13日の CRUD 記録と9月15日の配備成功は、これらの全データ受入の代替にならない。

## 本番・納品の残り

| ID | 状態 | 残作業・完了条件 |
|---|---|---|
| `P1 / SEC-SECRETS-5 / #89 / #97` → `EMR-147` | Planeへ移行済み |  |
| `P2 / #253 / PROD-SETUP` → `EMR-148` | Planeへ移行済み |  |
| `P3 / #250 PROD-DATA-MIGRATION` → `MIG-23` | Planeへ移行済み |  |
| `P5 / #255 STAFF-PROVISION` → `EMR-149` | Planeへ移行済み |  |
| `P1 / SEC-SECRETS-5 / #89 / #97` → `EMR-147`, `P2 / #253 / PROD-SETUP` → `EMR-148`, `P6 / #258 / U1-U12 DELIVERY` → `EMR-150` | Planeへ移行済み |  |
| `P7 / #256 / TRAINING` → `EMR-151` | Planeへ移行済み |  |

Q1検索・Q4保険・Q2履歴、[追加実装5件と会計等の部分対応](todo-verification.md#code-followup-20260921)、今回 `cd2feaa14` の主訴null hydrate修正は、現行productionへの反映証拠を今回未照合（UNKNOWN）。本番releaseの対象revisionと既存receiptを照合し、反映済みなら重複配備しない。ブラウザ受入と上記前提を先に確認し、STG 配備だけで本番反映済みにしない。P4 / P8 / E1 / E2、go-live は [検証 TODO](todo-verification.md) を正本とする。

## データ操作の着手プラン

### UAT-Q3-GENDER-MAP

1. [性別修正](todo-issue.md#uat-q3-gender-map) は old_db `5fbc3b2` → merge `a2cea37` でmain統合済み（9月19日読取）。producer担当は修正を含むrevisionと現行bundle/検証receiptを照合し、未生成なら正規経路で再生成する。38 tests / SQLite CASEの既存成功をPostgreSQL・export・STGの証拠にしない。再統合タスクは起こさない。
2. 運用担当が対象医院、旧コードを追跡できる根拠、訂正対象件数、除外条件、backup・監査・復旧方法を事前照合する。producer の修正と既存 STG の修正は別々に判定する。
3. 承認された限定訂正後に雄/雌の件数・表示を照合する。手術日は変更対象と混ぜず、推測日付を入れない。成果物は producer revision、承認参照、前後集計、画面確認の非機密 receipt。

### UAT-Q2-VACCINE-SPECIES

1. [調査案](todo-issue.md#uat-q2-vaccine-species) を医院・時刻窓・出力列（件数/参照関係だけ）に限定し、読取担当と対象を固定する。
2. 読取承認後に集計を取得し、マスタ種欠損・参照違い・旧記録由来の分類へ渡す。
3. 成果物は機密除去済み分類表と訂正要否。原因未確定や参照不整合が残る間は更新案を適用しない。履歴修正は対象行と復旧方法を確定した別承認とする。

### UAT-Q4-UNPAID-TRIAGE

1. [集計案](todo-issue.md#uat-q4-unpaid-triage) の対象医院・期間・status・payment の結合条件を確認し、読取担当と保存先を固定する。
2. 承認後に status・支払有無・金額帯を集計し、旧未精算と突合漏れを分ける。重複結合で件数/金額が増えていないか元集計と突合する。
3. 成果物は総件数・金額と原因別集計、追加照合または限定訂正の提案。一括完了・請求削除は実行しない。

### PO-PET-DECEASED-DATA-BACKFILL

1. [元の修復計画](bug.md#plan-po-pet-deceased-data-backfill) に沿って不整合の種類、日付根拠、対象医院と件数を整理する。日付根拠がある行だけを訂正対象とし、根拠がない行は補完対象外として残す。
2. 根拠に基づく対象・訂正内容を確定した後、操作者、backup、監査、失敗/通信断時の照合・復旧を含む限定実行案を作る。
3. 承認後の訂正と前後件数・画面・監査の一致で完了とする。死亡 write ガードを再実装せず、監査や既存履歴を削除しない。

### BUG-LOCAL-HANDOFF-CSV-CONTRACT

> 移行済み: Plane `MIG-19`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## STG データレーンの着手プラン

### H0-2 / HAC-CSV-1

> 移行済み: Plane `MIG-20`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### H0-3b / H1-2

> 移行済み: Plane `MIG-21`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### AE-STG-UAT-LANE3-HAC

> 移行済み: Plane `MIG-22`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### H3-9

> 移行済み: Plane `EMR-144`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### H3-11

> 移行済み: Plane `EMR-145`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### Lane 4

> 移行済み: Plane `EMR-146`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## 本番・納品の着手プラン

### P1 / SEC-SECRETS-5

> 移行済み: Plane `EMR-147`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### P2 / PROD-SETUP

> 移行済み: Plane `EMR-148`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### P3 / PROD-DATA-MIGRATION

> 移行済み: Plane `MIG-23`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### P5 / STAFF-PROVISION

> 移行済み: Plane `EMR-149`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### P6 / DELIVERY

> 移行済み: Plane `EMR-150`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### P7 / TRAINING

> 移行済み: Plane `EMR-151`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
