# 認証・認可レビューの未完了TODO

> 作成日: 2026-09-08  
> 最終整理: 2026-09-10
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)
> 元レビュー: [todo-check-auth.md](todo-check-auth.md)

本書には未完了作業だけを記載する。ローカル実装と限定検証が完了した項目、過去の実装計画、PASS表、レビュー履歴はGit履歴へ退避した。

実装基準は `a4292a241`。認証・認可のローカル修正と、D2/D3 auth用テストは `main` に統合済み。以下の実DB・実環境・外部状態は未完了であり、ローカル実装完了から推定しない。

## 未完了サマリー

| 優先 | ID | 状態 | 残作業 | 完了条件 |
| --- | --- | --- | --- | --- |
| P0 | D3 | INCOMPLETE | GET/HEAD 178経路の実DB返却データ分離。auth 2経路はテスト実装済み・実DB NOT_RUN。clinic-fixed 156、cross-clinic 20はテスト未実装・NOT_RUN | 全対象で認可と返却clinicを実DB検証し、未検証0件 |
| P0 | D2 | BLOCKED（環境なし） | 自己ロックアウト防止の実DB並行3テスト | 既存環境の利用は承認済み。使い捨てPostgresが利用可能になったら3テストPASS。`-short` SKIPをPASSにしない |
| P1 | D1 | BLOCKED / UNKNOWN | 初回管理者SQLの実DB検証、本番付与、対象環境メール確認 | 合成DB検証後、承認済み本番手順と非機密receiptを完了 |
| P1 | D5 | BLOCKED（環境なし） | 同一DBを使う独立2 APIプロセスで即時失効と追加DBコストを測定 | 旧session拒否と未変更医院アクセスを確認し、測定結果を保存 |
| P2 | LINEAR | UNKNOWN | 対応チケットの特定と状態更新 | 対象issueをライブ確認し、承認後に未完了境界を投稿 |

## 共通の安全条件

- 共有 `ekarte_db`、`old-db-postgres`、STG、PRODをテストDBに使わない。
- agentはmigrationをapplyしない。`make up`、共有環境起動、デプロイ、本番操作を自動実行しない。
- 資格情報、接続文字列、メールアドレス、cookie、token、患者情報を本書・ログ・Gitへ記録しない。
- 実DBテストは使い捨てDBだけで行い、接続先を非機密メタデータで照合してから開始する。
- Linear投稿、本番付与、対象環境メール、STG/PROD変更は個別の明示承認を得る。

## 既存環境の確認結果（2026-09-10）

ユーザーから既存ローカル環境をD2/D3/D5の検証に使う許可を受領した。OrbStackは起動できたが、`docker ps -a` は0件で、既存の使い捨てPostgres・backend・frontend containerは存在しなかった。

この確認では新規container作成、`make up`、migration apply、DB接続、テスト実行をしていない。D2、D3の実DB部分、D5は **NOT_RUN**。既存containerが利用可能になった時点で、接続先が共有DB・old_db・STG・PRODでないことを確認して再開する。

## D3. GET/HEADの実DB返却データ分離

### 対象集合

正本は [get_head_permissions.json](backend/cmd/api/testdata/get_head_permissions.json)。登録203経路のうち、実DB返却分離の対象は178経路。

| class | 対象 | 現在 |
| --- | ---: | --- |
| clinic-fixed | 158 | auth 2経路はテスト実装済み・実DB NOT_RUN。残り156は未実装・NOT_RUN |
| cross-clinic | 20 | 条件別契約は整理済み。実DBテストは未実装・NOT_RUN |
| 除外 | 25 | write、セッション、直接ファイル、`GET /clinics`など。除外理由の変更時は集合を再計算 |

### 検証契約

共通fixtureは医院A/B、A/B両所属staff、選択医院Bを使う。

- grant Aのみ + selected B: clinic-fixedは403でservice未到達。cross-clinicは契約に従い403、またはAだけを返す。
- grant Bあり + selected B: Bの観測可能なseedを返す。空一覧による自明PASSを禁止する。
- AのIDをselected Bで指定: 404または契約上の403。Aのbody、名称、件数、集計、署名URLを漏らさない。
- list/detail/search/report/CSV/集計は、返却要素、件数、合計、関連行まで許可clinic集合内であることを確認する。
- 403/404/200を一つのOR期待値にしない。request条件ごとにケースを分ける。
- handler stubやmiddleware単体、テスト名の存在、`-short` compileだけを実DB分離PASSにしない。

### 次の実装単位

1. `backend/internal/auth/realdb_selected_clinic_b_grant_a_isolation_test.go` の5ケースを使い捨てPostgresで実行する。
2. clinic-fixed残り156経路をhandlerSource package単位で実装する。対象packageは medicalrecord 61、lstep 28、billing 22、reservation 12、staff 12、trimming 8、clinic 5、inventory 4、pet 4。
3. cross-clinic 20経路を条件別に実装する。
4. JSON集合と実装済みテスト集合を機械照合し、missing / extraを0にする。
5. 実DB実行後、全返却値のclinic境界と非空B seedを確認する。

実装ごとに個別claimを取得し、対象worktreeに結び付いたDockerで検証する。実DB環境がなくてもテスト実装とoffline compileは進められるが、実DB結果はNOT_RUNのまま保持する。

## D2. 自己ロックアウト防止の実DB並行検証

対象テスト:

- `backend/internal/auth/permission_group_policy_concurrency_db_test.go`
  - `TestPermissionPolicyDB_ConcurrentGroupDeactivation`
  - `TestPermissionPolicyDB_ConcurrentRuleReplacement`
- `backend/internal/staff/staff_service_permission_policy_concurrency_db_test.go`
  - `TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation`

使い捨てPostgresへ接続し、各テストを直列に実行する。`TEST_DATABASE_URL` と管理接続の双方を確認し、共有DB・old_db・`127.0.0.1:15432`を検出したら停止する。

確認事項:

- 競合する2操作の一方がForbidden、他方が成功する。
- actorに `master-permission:view` と `edit` が残る。
- business writeと監査が同一transactionでcommitまたはrollbackする。
- `pg_locks` のholder/contender待機が成立し、goroutine・connectionが終了する。
- 終了後に使い捨てDBを破棄し、接続漏れがない。

## D1. 初回管理者の実DB・実環境確認

入口:

- [first_system_admin.sql](backend/internal/auth/testdata/first_system_admin.sql)
- [FIRST_SYSTEM_ADMIN.md](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)
- `backend/internal/auth/first_system_admin_procedure_test.go`

残作業:

1. `001_init` 適用済みの使い捨てDBで正常系、既存admin競合、既存email競合、監査失敗rollbackを実行する。
2. account、staff紐付け、clinic、監査行、失敗時不変を確認する。
3. 対象環境、operator、承認記録、既存staff/主所属を確定する。
4. 本番付与を別承認で実行し、本人が通常loginできることを確認する。
5. 対象環境のメール経路を別承認で確認する。

接続先・role・対象環境・担当者のいずれかがUNKNOWNなら実行しない。合成DB検証、本番付与、メール確認はそれぞれ別の操作として扱う。

## D5. 独立2プロセスの即時失効と性能測定

同一の隔離DBに接続する独立2 APIプロセスと、別cookie jarの2セッションを使う。共有compose/main mountと本番負荷は使わない。

| 項目 | 条件 |
| --- | --- |
| 変更 | staff無効化、clinic所属解除、password/epoch更新 |
| 安全性 | commit後の次リクエストで旧sessionを401/403。TTL待ち禁止 |
| 負アサート | 所属解除時も未変更医院への正当アクセスを維持。新passwordで再login可能 |
| endpoint | `GET /api/v1/me` とclinic-fixed 1本 |
| 測定 | 1プロセス基準と2プロセスのserver p50/p95、DB statements/request、error率 |
| 負荷 | 1セッション1並行、warm-up 5、本測定30、50ms間隔 |
| 即時停止 | 旧権限・旧sessionを許可した場合。性能結果を待たない |
| 要調査 | 2プロセスp95が基準比+100%超。+200%超または測定不能なら性能ゲート停止 |

生latency、要約、プロセス識別子、DBの非機密識別子、変更commit前後のHTTP statusを証拠として残す。

## Linear反映

Ticket IDと現在状態はUNKNOWN。ライブ照会で対象issueを特定し、外部投稿の明示承認後に次だけを反映する。

- ローカル実装は `a4292a241` まで統合済み。
- D2実DB並行、D1実環境、D3実DB分離、D5独立2プロセスは未完了。
- static、offline、`-short`、stubのPASSを実DB・本番・受入完了へ変換しない。

## 台帳全体ステータス

**INCOMPLETE / BLOCKED**。D3のテスト実装は独立して継続可能。実DB、実環境、Linearはそれぞれの開始条件と承認が揃うまでNOT_RUN / UNKNOWNを維持する。
