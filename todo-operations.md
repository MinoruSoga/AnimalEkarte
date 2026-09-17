# 運用・外部実行 TODO

最終照合: 2026-09-17。未完了の環境・データ・本番・納品作業を管理する。実装は [todo-issue.md](todo-issue.md)、受入は [todo-verification.md](todo-verification.md)。既存チケットは [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下。完了した PR の CI 修復・マージ待ちは削除した。

9月17日の `git ls-remote` では remote main は `f5699c2a`、local main `72807128` は1コミット先。新しい4単位の push・GitHub CI・STG/production は未確認。Linear は `USER_NOT_LOGGED_IN` で現在 UNKNOWN。9月15日の GitHub / Linear 読取は履歴。STG / PROD の DB、秘密、最新データ投入 receipt は確認していない。以前の「未構築」「未受領」を現在の事実として断定せず、再実行前に実施有無を確認する。

着手プランを 2026-09-15 に補完した。各表の ID から下の個別手順を参照する。承認前にも、既存資料の照合・不足入力表・実行案の作成は進められる。外部の実操作を、この計画の記載だけで開始しない。

<a id="uat-data-operations"></a>

## 医院 UAT に伴うデータ操作

| ID | 残作業・担当 | 開始条件と完了条件 |
|---|---|---|
| [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) | old_db / USER: 検証済み隔離候補の main 統合・bundle、既存 STG の性別訂正 | [課題本文](todo-issue.md#uat-q3-gender-map) のデコード・対象・バックアップ・監査・復旧を固定。承認後に適用し、性別と手術日を別々に照合 |
| [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) | USER / 調査担当: STG 集計の読取範囲を確保 | 医院と時刻窓を固定し、件数・参照関係で原因分類。個体情報を共有せず、履歴修正は別承認 |
| [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) | USER / 調査担当: STG 未納の集計 | status、支払有無、金額で切り分け。既存請求を一括完了にしない |
| [PO-PET-DECEASED-DATA-BACKFILL](#po-pet-deceased-data-backfill) | USER: 日付根拠がある行だけの限定訂正 | [修復方針](todo-issue.md#po-pet-deceased-data-backfill) の対象と死亡日根拠を承認後に適用。監査・前後件数・復旧を確認 |
| [BUG-LOCAL-HANDOFF-CSV-CONTRACT](#bug-local-handoff-csv-contract) | producer / USER: 現行契約の bundle 再生成・受領 | [元の障害](bug.md#plan-bug-local-handoff-csv-contract) は古い rehearsal bundle の不一致。現在の producer / manifest を照合し、未解消なら一体で再生成。digest の手修正や eligibility 昇格で通さない |

実行手順は [ローカル handoff](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)、[STG 停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。共有環境への書込み・再取込・DB 作成/破棄・migration は今回実施しない。必要な `make migrate` は、対象環境の適用状態を確認したうえでユーザーが実行する。

<a id="stg-data-lanes"></a>

## STG データレーンの残り

| ID / 段階 | 現在の判定 | 次の一手 |
|---|---|---|
| [H0-2 / HAC-CSV-1](#h0-2--hac-csv-1) | UNKNOWN（producer の追加状況未照合） | 前回の HAC-INPUT-2 は完全 KNJO 未受領。現在の受領・producer 結果を確認し、未解消なら完全入力を待つ |
| [H0-3b / H1-2](#h0-3b--h1-2) | 前段と実施証拠待ち | 入力・対象・承認が揃っているか確認。過去 receipt があれば重複実行を避ける |
| [AE-STG-UAT-LANE3-HAC](#ae-stg-uat-lane3-hac) | UNKNOWN | 現行投入状態と件数・金額・医院境界を確認して投入要否を判断 |
| [H3-9 staff attach apply](#h3-9) | UNKNOWN | 現在の attach と既存実行証拠を確認後、必要分だけ承認を取得 |
| [H3-11 画面確認](#h3-11) | UNKNOWN | 自医院 login と担当・所属の画面証拠を確認 |
| [Lane 4](#lane-4) | 完了未証明 | 両院 Lane 3 verify と H3-11、所定の運用日数の証拠を [受入側](todo-verification.md#stg-データレーン) で照合 |

同一の既知破損 BAK の再復元や、既存 STG の無条件上書き load は行わない。9月13日の CRUD 記録と9月15日の配備成功は、これらの全データ受入の代替にならない。

## 本番・納品の残り

| ID | 状態 | 残作業・完了条件 |
|---|---|---|
| [P1 / SEC-SECRETS-5 / #89 / #97](#p1--sec-secrets-5) | receipt 未照合 | 4系統の発行・投入・deploy・health・旧値 revoke / 拒否の非機密 receipt を確認。不足時は明示承認後に実施 |
| [P2 / #253 / U12 PROD-SETUP](#p2--prod-setup) | 現行環境 UNKNOWN | Cloudflare 本番、reviewer / workflow、rollback、backup rehearsal、URL/CI receipt を確認 |
| [P3 / #250 PROD-DATA-MIGRATION](#p3--prod-data-migration) | 前提・実施証拠待ち | rehearsal、入力停止、backup/rollback、最終 import、件数・医院・金額突合 |
| [P5 / #255 STAFF-PROVISION](#p5--staff-provision) | 入力・実施証拠待ち | roster・email 方針・clinic・role・操作者・環境承認を非機密 receipt で確認 |
| [P6 / #258 / U1–U12 DELIVERY](#p6--delivery) | 最終承認待ち | P1 / P2 と契約責任者の事実を [納品パッケージ](docs/delivery/DELIVERY_PACKAGE.md) に反映 |
| [P7 / #256 / U13 TRAINING](#p7--training) | 完了未証明 | 説明会の日程・形式・範囲・実施結果の receipt を確認 |

Q1検索・Q4保険・Q2履歴を含む STG 修正の production 反映は未完了として本番 release に含める。ブラウザ受入と上記前提を先に確認し、STG 配備だけで本番反映済みにしない。P4 / P8 / E1 / E2、go-live は [検証 TODO](todo-verification.md) を正本とする。

## データ操作の着手プラン

### UAT-Q3-GENDER-MAP

1. [検証済み隔離候補](todo-issue.md#uat-q3-gender-map) は未コミット・old_db main 未統合。producer 担当が対象差分を統合し、現行契約の bundle を再生成・照合する。38 tests / SQLite CASE の成功を PostgreSQL・export・STG の証拠にしない。
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

1. [配置手順](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) と [受入契約](docs/ops/deploy/CLINIC_CSV_IMPORT.md) を開き、現行 producer/consumer revision、21 CSV の header・digest、manifest の run/医院/eligibility を比較する。過去 bundle だけで現在の失敗を断定しない。
2. 差異が残れば producer 担当に CSV と manifest を一体で再生成してもらい、承認されたローカル配置先で `make old-db-handoff-check` を確認する。この check だけでは DB 受入完了にならない。
3. 成果物は契約差分、再生成 run/digest、配置 check の結果、次の preflight への引継ぎ。手作業で digest や eligibility を書き換えず、formal と rehearsal を分離する。

## STG データレーンの着手プラン

### H0-2 / HAC-CSV-1

1. [BRT-42](https://linear.app/baritechllc/issue/BRT-42) の既存本文・最新 receipt と producer の現在の受領記録を照合する（9月15日の本文読取は Needs Human。入力そのものは未確認）。過去の「完全 KNJO 未受領」だけで現在も未受領と決めない。
2. 未受領なら、data owner に CHECKDB clean な新規 BAK または全32列完全 KNJO、取得日時、source provenance の不足表を渡す準備をする。送信・抽出・再復元は担当者の承認範囲で行う。
3. 入力が揃った後に producer 担当が正規 pipeline と検証を行い、完全性・21表・eligibility を備えた bundle を引き渡す。成果物は受領参照、producer revision/run、manifest digest、検証結果。既知破損 BAK の再復元や PARTIAL の昇格で通さない。

### H0-3b / H1-2

1. H0-2 の最新 bundle と [handoff 手順](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) を照合する。旧 ID の担当分界は現台帳だけでは詳細不明のため、既存 Issue/producer 引継ぎで「受領・配置・preflight」の担当と対象を確定するところから始める。
2. manifest の医院、source run、consumer 契約、全 CSV digest、formal/rehearsal の用途を固定する。ローカル配置確認後、承認された対象で用途に合う read-only preflight を行う。
3. 成果物は ID 別の担当分界、同一 manifest に結び付いた preflight と対象環境の receipt。前段不足・契約不一致・既存 band 占有・担当分界不明なら apply に進まない。過去 receipt が現在の入力と一致すれば重複実行しない。

### AE-STG-UAT-LANE3-HAC

1. H0-2 と H0-3b の結果、現在の八王子データ、対象医院・band・seed binding を照合し、投入要否を決める。
2. [現行 handoff](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) の一括 wrapper は八王子を対象外としているため、そのまま流用しない。[CSV import 契約](docs/ops/deploy/CLINIC_CSV_IMPORT.md) と [Makefile](Makefile) の formal / STG UAT 個別経路から、eligibility と対象に合う経路を運用担当が確定する。
3. 承認された backup/rollback・時間枠・operator・同一 manifest で preflight → apply → verify。医院別の件数・参照・金額・監査を成果物にする。既存非空 band、不明な commit 結果、入力不一致なら停止し、無条件上書きしない。

### H3-9

1. [Makefile の staff attach](Makefile) と [実行入口](backend/cmd/stg-uat-staff-attach/main.go) を開き、移行済み staff と account の現在の結合・既存 receipt を確認する。新しい staff を作る `staff-provision` とは分ける。
2. repo 外の roster/secrets、明示された staff/clinic/role 対応、target host/database を承認記録と固定する。`make stg-uat-staff-attach-preflight` の指摘を解消してから、必要な対象だけ `make stg-uat-staff-attach` の実行案を承認に出す。
3. 人間による実行後に既存 staff の保持、所属と権限、監査、通常 login を確認し H3-11 へ渡す。成果物は非機密件数・digest・承認参照。氏名一致での自動統合や既存 attach の上書きをしない。

### H3-11

1. H3-9 の対象と receipt を受領し、承認された本人・医院・ブラウザ・画面操作範囲を固定する。
2. 自医院 login → 医院選択 → 担当スタッフ表示 → 再読込を確認し、権限外医院や未所属スタッフを選択できないことを検証する。
3. 成果物は対象 build、匿名化した操作結果、所属/権限の一致、不一致時の引戻し先。同一名が表示されただけで一致とせず、H3-9 と同じ対象であることを運用担当が照合する。

### Lane 4

1. 両院の Lane 3 verify と H3-11 をそろえ、対象 revision・運用開始日・5営業日の観測枠・医院側担当・中止条件を固定する。
2. [継続運用チェックリスト](docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md) に沿って日次の login・主要操作・締め・監査失敗/障害を記録する。必要な操作は承認範囲内のデータで行い、途中の配備変更は影響と再確認範囲を残す。
3. 成果物は両院の日別 receipt、障害/残件の扱い、受入担当の判断。[TODO-V-STG-DATA](todo-verification.md#stg-データレーン) がこれを照合する。営業日不足・未確認操作・未解消の重大 FAIL があれば終了にしない。

## 本番・納品の着手プラン

### P1 / SEC-SECRETS-5

1. [資格情報手順の不足表](docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) を使い、4系統の既存 receipt と対象 config/環境を names-only で照合する。
2. 不足分について人間が発行 → 投入 → deploy → health と対象機能確認 → 旧値 revoke/拒否の実行案を作る。暗号鍵は既存暗号化データの再暗号化/復元手順が別途検証されるまで変更しない。
3. 明示承認後に担当者が実施し、系統別の非機密 receipt を保存する。#97 本文のマスクはローテーション完了後の別外部操作。秘密値を台帳へ載せず、未確認の系統があれば P1 を閉じない。

### P2 / PROD-SETUP

1. [production setup](docs/ops/infra/production/setup.md) §1–6 と `.github/workflows/backend-deploy.yml`、`backend/wrangler.production.jsonc` を比較し、実装差分と provider 設定の不足を分ける。checked-in の draft を構築済みの根拠にしない。
2. 人間が Environment protection/reviewer・課金・DB/R2/DNS・秘密 scope を確認。必要な workflow 変更は対象と検証を別実装単位にして、保護条件が整った後に有効化する。
3. [production runbook](docs/ops/infra/production/runbook.md) の backup/isolated restore と rollback を確認した後、承認された配備を行う。成果物は ref/run、環境承認、配信先、health、DB を使う操作、frontend API target の receipt。未確認条件があれば構築/配備を停止する。

### P3 / PROD-DATA-MIGRATION

1. P2、正式 bundle、[CSV import](docs/ops/deploy/CLINIC_CSV_IMPORT.md)、[当日手順](docs/delivery/GOLIVE_RUNBOOK.md) を照合し、対象・操作者・入力停止・backup/rollback・突合方法を固定する。
2. 承認済み disposable で preflight/import/verify と復旧 rehearsal を先に完了させる。本番作業は別の承認済み window で入力停止 → 最終抽出/差分 → backup → 最終 import → verify の順に行う。
3. 成果物は同一 manifest/revision に対応する件数・clinic・参照・金額の突合と復旧判断。rehearsal PASS を本番実行済みに読み替えず、不明な commit 結果や突合 FAIL では P4/P8 へ進まない。

### P5 / STAFF-PROVISION

1. [スタッフ発行の不足入力表](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md#不足入力チェックリスト2026-08-20) の I-ROSTER〜I-RECEIPT を現在の入力と照合する。名簿受領履歴と現在版の適用可否を分ける。
2. 個人 email 方針、clinic、明示した permission group、休退職者、actor、対象環境を確定し、repo 外に manifest/secrets を準備する。初回管理者が未整備なら [D1](todo-verification.md#認証認可の外部境界) を先に満たす。
3. 承認済み preflight → 人間による apply → login・所属・最小権限・audit の確認を行う。成果物は batch/digest/count と結果の receipt。実 roster や秘密を Git に保存せず、架空スタッフで未入力を埋めない。

### P6 / DELIVERY

1. [納品パッケージ](docs/delivery/DELIVERY_PACKAGE.md) の U1–U12 を契約名義・本番環境・通知先・運用担当の現在の事実と照合し、不足欄を供給者別にまとめる。
2. P1/P2 と関連する実行 receipt を取得して、手順・構成・管理者設定・backup/障害連絡の記載を実際の納品対象へ合わせる。秘密の投入と本文更新を同じ作業にしない。
3. 成果物は未確定欄の解消と契約責任者の承認がある納品版。U13 の研修は P7 に残し、文書整備だけで go-live としない。

### P7 / TRAINING

1. [操作マニュアル](docs/delivery/OPERATION_MANUAL.md) の U13 を基に、日程・形式・対象職種・担当・合成データと説明範囲を確定する。案内の送信は担当者の承認後とする。
2. 受付→診療→会計→締め、例外処理、障害連絡を対象版で予行し、説明会でスタッフが操作できるか確認する。
3. 成果物は実施日時・範囲・質疑/未解決項目・担当者の非機密 receipt。開催予定だけで完了にせず、未解決の操作課題は該当タスクへ戻す。
