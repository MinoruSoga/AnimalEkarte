# 運用・外部実行 TODO

統合日: 2026-09-11。実行 SoT は Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)**。

本書は、開発以外の STG 実データ、秘密・本番環境、納品、研修に関する実行作業を扱う。検証・受入・go-live 判定は [todo-verification.md](todo-verification.md) を正本とする。秘密値・個人情報は記録しない。

## 対応順

| 順 | ID | 実行者 | 状態 | 前提 |
|---|---|---|---|---|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | **BLOCKED**（HAC-INPUT-2。完全 KNJO 未受領） | CHECKDB clean な新規 BAK、または全32列完全 KNJO。同一・既知破損 BAK の再復元と城東 live の上書き load は禁止 |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | UNKNOWN・投入判断待ち | 現行状態、H0-2、H0-3b |
| 4 | **H3-9 staff attach apply** | USER | 実施有無 UNKNOWN | 現行 attach と STG 実行ゲート |
| 5 | **H3-11 画面確認** | USER | UNKNOWN・証跡未取得 | H3-9 と自医院ログイン |
| 6 | **Lane 4** | 医院スタッフ / USER | 完了未証明 | 両院 Lane 3 verify、H3-11 |
| 7 | **P1 → P2 → P3 → P5 → P6 → P7** | USER / 開発 | 待ち | 各作業の外部承認と前段の非機密 receipt |

## 外部環境・本番の実行作業

| ID | 実行者 | 状態 | 完了条件 |
|---|---|---|---|
| **P1: SEC-SECRETS-5 / #89 / #97** | USER | 4系統 rotation receipt 未記入 | 新発行、投入、再 deploy、health、旧値 revoke、旧値拒否 |
| **P2: #253 / U12 PROD-SETUP** | USER / 開発 | Production 未構築 | Cloudflare本番、Required reviewers、workflow、rollback、backup rehearsal、URL/CI receipt |
| **P3: #250 PROD-DATA-MIGRATION** | USER / 開発 | 事前準備待ち | rehearsal、最終import、入力停止、backup/rollback、件数・clinic_id・金額突合 |
| **P5: #255 STAFF-PROVISION** | USER | 入力未記入 | roster、email方針、clinic、role、actor、環境承認の PII-free receipt |
| **P6: #258 / U1〜U12 DELIVERY** | USER | 最終承認待ち | P1・P2と契約責任者の非機密事実を `DELIVERY_PACKAGE.md` に反映 |
| **P7: #256 / U13 TRAINING** | USER | 操作説明会未完 | 日程、形式、範囲、結果、opaque receipt |

P4、P8、E1、E2を含む受入・go-live判定は [統合検証TODO](todo-verification.md) で管理する。

## STG 実データの現状

対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。証跡が無いだけなら **UNKNOWN**。再 apply 前に USER が現行状態を確認する。

2026-09-06 の観測では、城東・敷島・箱は AE handoff に bundle があり、八王子ディレクトリは空だった。old_db export にも八王子 run はなく、producer は未実行。Lane 4 は5営業日の証跡がない。

実行ゲートは [STG 手順の停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)、実行後の検証は [統合検証TODO](todo-verification.md#stg-データレーン) を参照する。
