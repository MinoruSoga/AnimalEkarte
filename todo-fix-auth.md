# 認証・認可レビューの修正TODO

> 作成日: 2026-09-08  
> 実施日: 2026-09-08〜2026-09-09  
> 対象: [todo-check-auth.md](todo-check-auth.md) とレビュー時の作業ツリー。  
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)。  
> 最終状態確認: 2026-09-09、ローカル `main` = `48e89dbe4`。初回実装（[PR #390](https://github.com/MinoruSoga/AnimalEkarte/pull/390)、`2dc3a2d51`）に加え、再レビュー修正 `950404408` と残件テスト・検証契約 `9be825a66` も `main` に統合済み。
>
> 本書はレビュー指摘のローカル整理と対応記録である。Linear は更新していない。

> 現在の結論: 実装・限定テスト（UpdateRules / Update-with-Rules 含む）のローカル修正は完了。実DB並行処理、全経路の実DB返却データ分離、初回管理者SQL実行、独立2プロセスの失効検証・負荷測定、本番付与・対象環境メール・Linear反映は未完了。ローカル修正完了を実環境・運用の完了と扱わない。
>
> Git整理: 旧認証作業のブランチ・claimは整理済み。今回の文書更新開始前はローカルブランチが `main` / `staging` のみであることを確認した。この文書更新用に `claim/TODO-FIX-AUTH` を再取得している。旧作業のロック継続を意味しない。
>
> 履歴の読み方: 下記の「未コミット」「claim保持」「commit/push/mergeなし」「独立レビュー未実施」は各検証時点の記録。後続の「main統合前の追加確認」で追加検証・独立レビューを実施し、上記コミットへ統合した。今回の最新化ではテスト再実行・外部状態の再照会・更新をしていない。

## 対応サマリー

| ID | 項目 | 判定 | 証拠 |
| --- | --- | --- | --- |
| 1 | 非管理者によるシステム管理者パスワード変更の拒否 | 対応 | `staff_service_update.go` の TX 内 `FindByIDForUpdate`。`staff_admin_password_guard_test.go` |
| 2 | GET/HEAD 他院 grant fallback を閉じる | 部分対応 | 選択医院 grant の実装と203経路の完全一致台帳。clinic-fixed は middleware deny と許可（Bに grant）を全件実行。横断 class 21件（うち `GET /clinics` は所属一覧）は allowing middleware と handler を実行。実DB臨床返却分離の対象は横断20+clinic-fixed158=178（台帳D3）。元バグ面は handler の拒否と許可も実行。実DB返却データの分離は未検証 |
| 3 | login/refresh で無効医院を選ばない | 対応（限定テスト） | active 医院解決共有。全医院失効時の403をFEで判別。自動復旧は一度、以後は手動再試行。失敗書き込みの再送なし。ログアウト開始時に復旧を中止 |
| 4 | 権限グループ無効化の自己ロックアウト防止 | 対応（実装・単体） / 実DB並行はBLOCKED | 2026-09-09: `UpdateRules` と Update-with-Rules の旧自己グループ事前拒否をやめ、post-mutation 実効 view+edit を最終判定に統一。証拠修復で staged/committed TX double・view/edit喪失命名対応・pg_locks holder/contender 待機・worker Cleanup join を追加。実DB実行は隔離Postgres未承認でBLOCKED |
| 5 | 旧権限・旧セッションの即時失効 | 部分対応 | production から current-access キャッシュ除去。**SKIP**: 同一 DB・独立 2 サーバープロセスのライブ検証（`make up` 禁止） |
| 6 | `/me.main_clinic_id` と `is_main` の文書訂正 | 対応 | API 改名なし。`docs/architecture/auth.md` §4.8、OpenAPI 説明 |
| 7 | `staleTime=5分` の文書訂正 | 対応 | 自動ポーリング追加なし。`get-me.ts` / AuthProvider / auth.md §4.8 |
| 8 | 既存スタッフへのアカウント付与と初回管理者手順 | 部分対応 | 画面APIと初回管理者の人間用SQL手順を整備。SQLを `001_init.sql` と手順書へ静的照合。実DB実行・本番実付与・メール基盤の本番検証は未実施 |
| 9 | 非システム管理者の医院一覧を所属に限定 | 対応 | `scope=all` でも所属のみ。フラグ欠落は 401 |

### 全体スキップ（原因）

| 対象 | 原因 |
| --- | --- |
| Linear 更新 | 外部チケット未承認 |
| `make migrate` / migration apply | エージェント禁止。今回 migration 追加なし |
| `make codegen` | エージェント禁止。OpenAPI は手書き同期 |
| `make up` / STG/PROD / デプロイ | 起動・本番操作は範囲外 |
| 資格情報の発行・変更 | 秘密を git / 応答に載せない |
| D5 の独立 2 プロセス検証 | 実行中スタック起動禁止。compose が共有 main をマウントし得る |
| D1 の対象環境メール実送信 | メール自動送信を仕様で行わない。本番メール基盤の現地確認なし |
| `main_clinic_id` API 改名 | 文書訂正のみ。互換破壊をしない |
| `/me` 自動ポーリング追加 | 指摘は誤認訂正であり追加しない |

## 対応範囲と進め方


- 本書の番号は文書内の参照番号であり、正式なタスクIDではない。
- 下記仕様はユーザーが決定。実装担当者・本番初回作成の運用担当者の個人名は未確定であり、各着手前に確定する。
- 最優先は 1・2 の認可問題。次に 3・4 の機能不整合を修正する。
- 4・5・8・9 の質問事項は下記回答で確定。追加レビューの5点も末尾「詳細仕様の確定」で決定済み。実装設計・実装・検証と、実運用の担当者・環境指定は残る。
- 6・7 は現行動作を説明する文書訂正。API変更や定期ポーリング追加を自動的に意味しない。
- 各実装前に現行コード・関連チケット・claim・他者WIPを再確認し、必要な作業領域を隔離する。
- migration、資格情報変更、STG/PROD操作、デプロイはこのTODO作成依頼の範囲外。

## 仕様確認の回答（2026-09-08）

ユーザー回答: **①-1、②-1、③-1、④-2、⑤-1**。仕様の選択であり、本番操作・資格情報発行の実行承認ではない。

| 質問 | 確定した仕様 | 対応項目 |
| --- | --- | --- |
| ① 失効タイミング | スタッフ無効化・所属解除・パスワード変更の完了後、次のリクエストから旧権限・旧セッションを拒否する。完了前に受け付けた処理の取消しは含めない。 | 5 |
| ② 自己ロックアウト | 非システム管理者が自分の最後の権限管理権限をなくす操作は禁止する。別の有効グループに管理権限が残る場合は許可する。システム管理者による他スタッフの権限解除は許可する。 | 4 |
| ③ 既存スタッフへのアカウント付与 | システム管理者がスタッフ画面から追加する。既存staff IDと診療履歴を維持し、別staffを作らない。 | 8 |
| ④ 本番初回管理者 | ユーザー申告では未作成。指定された運用担当者が初回専用手順で作成する。誰でも利用できる管理者登録画面は作らない。実環境での存在確認は未実施。 | 8 |
| ⑤ 所属外医院の表示 | 非システム管理者には所属医院だけ表示する。システム管理者の全院管理は維持する。active医院の認証スコープを無条件に拡張しない。 | 9 |

②は医院内の権限管理についての決定であり、最後のシステム管理者の削除・利用停止を許可するものではない。

## A. 修正すべき実装問題

### 1. 非システム管理者によるシステム管理者のパスワード変更（高・最優先）

- 対応状況: [x] 対応（2026-09-08）
- 該当: 元文書 §0・§7.2・§8.1。
- 問題: 対象者の全所属医院で `master-staff:edit`、選択医院で `master-permission:edit` を持つ非システム管理者が、対象の管理者フラグを検査されずにパスワードを変更できる。`is_system_admin` をAPIで直接変更できなくても、管理者の資格情報の乗っ取りにより権限境界を越えられる。
- 成立条件: 対象がログイン可能なシステム管理者で、操作者が上記の医院・権限条件を満たす。全医院の執行スタッフと、同じ医院集合に所属するオペレータも該当し得る。
- 根拠: `backend/internal/staff/staff_handler.go:92`、`staff_service_update.go:18-87`、`staff_service_core.go:266-315`（いずれも `backend/internal/staff/`）。
- 修正方針: パスワード更新と同じトランザクションで対象アカウントをロック・確認し、対象がシステム管理者なら操作者にもシステム管理者を要求する。アカウント確認失敗は拒否し、既存の監査・reset token失効・原子性を維持する。
- [x] 非管理者が全所属医院の編集権限を持っていても、管理者のパスワード変更を403で拒否する。
- [x] 拒否時にpassword hash・reset token・成功監査を変更しない。
- [x] 許可された管理者による変更と、通常スタッフへの正当な再設定は維持する。
- [x] 対象アカウント取得失敗・資格情報監査失敗・競合するアカウント変更を含む回帰テストを追加する。
- [x] 元文書に「フラグ昇格防止」と「既存管理者の資格情報保護」を分けて記載する。

### 2. GET/HEADの他院権限fallbackによる選択医院の閲覧認可漏れ（高・最優先）

- 対応状況: [x] 認可実装修正・静的台帳整備（2026-09-09）。D3 の全経路の実データ分離検証は未完了。
- 該当: 元文書 §4.5・§6.2・§8.4。
- 問題: `RequirePermission` は選択医院でdenyでも、他の所属医院に同じgrantがあればGET/HEADを通す。スタッフ一覧・詳細などでは後段に選択医院の権限再確認がなく、選択医院のデータが返る。権限グループの一覧・詳細にも同型の経路がある。
- 再現条件: A・B両医院に所属し、対象リソースのviewはAだけに付与した利用者が、Bを選択してBの一覧・詳細を要求する。所属内のリソース認可漏れであり、所属外の任意医院へアクセスできると主張しているわけではない。
- 根拠: `backend/internal/auth/http_permission.go:83-99`、同ファイルの権限グループGET、`backend/internal/staff/staff_handler.go:23-36,151-170`、`backend/internal/staff/staff_repository.go:38-56`。
- 修正方針: 医院固定APIは選択医院のgrantを必須にする。横断APIだけ明示的に対象医院ごとのFilter/Authorizeを適用する。共通fallback変更時は、その利用箇所を点検して正当な横断閲覧を壊さない。
- [x] 上記条件でBのスタッフ一覧・詳細、権限グループ一覧・詳細を拒否する。
- [x] A選択時の許可、およびBにもviewを付けた場合の許可を確認する。
- [x] 正当な横断APIでは、許可された医院だけが返り、結果0件・権限取得障害は契約どおり拒否される。
- [x] 共通middlewareと実handlerを組み合わせたHTTP回帰テストを追加する。
- [x] 元文書の「一覧・詳細は絞る」を実装済み経路に限定し、§8.4の現象を正常仕様として残さない。
- [x] pet / reservation の医院固定 GET は選択医院 grant を handler で再確認する。LINE 予約基本設定はパス医院の所属と `hospital_settings` grant。横断の List/Get（pets / owners / reservations）は Allowing のまま。

### 3. 一般スタッフのlogin/refreshが無効化済み医院を選ぶ（中）

- 対応状況: [x] 対応（2026-09-08。失敗 login の医院復旧除外は 2026-09-09）
- 該当: 元文書 §2.2・§6.1。
- 問題: 通常リクエストはactiveな医院に限定するが、一般スタッフのlogin/refreshは所属から無効医院を除外しない。主医院が無効だと、ログイン成功後に通常APIが403になり得る。他の有効な所属医院が残る場合も不整合が生じる。
- 根拠: `backend/internal/auth/http_session_login.go:95-110`、`http_session_refresh.go:127-143`、`current_access_service.go`（後二者も同ディレクトリ）。所属repositoryは所属のsoft-deleteを除外するが、医院のactive判定は行わない。
- 修正方針: login・refresh・通常リクエストで有効医院の解決規則を共有する。主医院が無効なら有効な所属へ切り替え、有効な所属が0件ならcookie発行前に拒否する。保存済みの旧選択医院からの画面復旧も確認する。
- [x] 主医院が無効・別の所属医院が有効なら、有効医院を既定として返す。
- [x] 全所属医院が無効なら、login/refreshで利用可能なセッションを発行しない。
- [x] 医院一覧の取得障害を、医院情報なしのログイン成功へ読み替えない。
- [x] login/refreshのHTTPテストと、旧医院をlocalStorageに保持した画面の復旧テストを追加する。
- [x] `POST /v1/login` の `clinic_selection_unavailable` では医院復旧に入らず、書き込み停止もしない。停止中でも login / logout は通す。

### 4. 権限グループ無効化による自己ロックアウト（中）

- 対応状況: [x] 対応（2026-09-08）
- 該当: 元文書 §4.2・§7.1。
- 問題: 自己所属グループのルールから `master-permission:edit` を外す操作は保護される一方、metadata更新の `is_active=false` は同じ保護を通らない。唯一の管理グループを無効化すると、操作者自身では復旧できない。
- 根拠: `backend/internal/auth/permission_group_service_rules.go:79-96,140-166`、`permission_group_service_mutate.go:12-29,171-194`、`permission_group_repository.go` のactive group限定集計（後二者も同ディレクトリ）。
- 修正方針（②-1で確定）: ルール変更・無効化・割当解除を、変更後の本人の実効権限で一貫して判定する。非システム管理者による最後の権限管理権限の自己喪失は拒否する。システム管理者による他スタッフの権限解除は許可する。
- [x] 非管理者が唯一の管理権限を失う無効化を拒否する。
- [x] 別の有効グループに本人の管理権限が残る場合は許可する。他の管理担当者が存在するだけでは自己喪失を許可しない。
- [x] システム管理者による他スタッフの権限解除を許可し、最後のシステム管理者の削除・利用停止の既存保護は維持する。
- [x] 自分のグループ全解除やルール変更から同じ保護を迂回できないことを確認する。
- [x] 判定用lookup失敗は拒否する。グループ更新・ルール変更・削除・スタッフ割当置換は同じ医院policy lockをrow lockより前に取得し、実効権限確認・監査・commitまで保持する。
- [ ] 異なる2グループの同時無効化を実DBで確認する。`TestPermissionPolicyDB_ConcurrentGroupDeactivation` / `ConcurrentRuleReplacement` / `TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation` を追加。`-short` では SKIP。実行は承認済み隔離Postgresが必要で BLOCKED。割り当て済みグループの同時削除は usage チェックで先に拒否されるため、自己ロックアウトの並行証明には使わない。

## B. 文書訂正・失効仕様の明確化

### 5. 旧権限・旧セッションの即時失効（中・仕様確定）

- 対応状況: [x] 部分対応（キャッシュ除去と文書は完了。独立2プロセスライブ検証は SKIP: `make up` 禁止）
- 問題: runtime compositionではcurrent accessに2秒キャッシュを挟む。cache hitではaccount/staff/所属/医院を再取得せず、無効化・所属解除・パスワード変更前の状態を使い得る。「毎回最新を再解決」「即時失効」と解釈できる説明は正確でない。キャッシュ自体を一律禁止する指摘ではない。
- 根拠: `backend/cmd/api/composition_auth.go:156-164`、`backend/internal/auth/current_access_cache.go:40-68`。
- 決定（①-1）: 無効化・所属解除・パスワード変更の完了後に受け付ける次のリクエストから拒否する。変更前に受け付けた処理の取消しは対象外。
- 着手プラン: 該当readのキャッシュ回避を含む最小の方式を比較し、複数実行インスタンスでも即時失効を守れる方式を採用する。1プロセス内だけのキャッシュ削除で完了としない。
- [x] TTL満了を待たずに上記3操作の失効を反映する。所属解除は解除した医院へのアクセスを拒否し、残る医院への正当なアクセスは維持する。
- [x] 変更完了後のリクエストと実行中処理の境界、取得障害時の拒否を両文書へ記載する。
- [ ] cache hit中の無効化、所属解除、password epoch更新とTTL満了後をテストする。新しい資格情報でのログイン直後に古いepochキャッシュで拒否されないことも確認する。 **SKIP**: production からキャッシュを除去したため cache-hit ケースは成立しない。独立2プロセスのライブ検証は `make up` 禁止。

### 6. `/me.main_clinic_id` と `clinics[].is_main` の意味を訂正する（文書）

- 対応状況: [x] 対応（2026-09-08）
- 該当: 元文書 §2.3。
- 問題: `/me` は選択中の `clinic_id` を `main_clinic_id` に渡し、`clinics[].is_main` もその値との一致で生成する。永続化された主所属を返す説明とは一致しない。login応答は既定医院から作るため、同じ形のレスポンスでも区別が必要。
- 根拠: `backend/internal/auth/http_session_me.go:31-32`、`backend/internal/auth/http_response.go:116-124,155-162`。
- [x] 元文書の「主医院」を現行の選択医院として訂正し、永続化された主所属と区別する。
- [x] login応答・`/me`・DB所属のそれぞれの意味を正規設計書と照合する。
- [x] API名の変更や主所属フィールド追加が必要かは利用箇所を確認して別途判断する。文書訂正だけのために互換性を壊さない。

### 7. `staleTime=5分` を自動更新間隔と誤認させない（文書）

- 対応状況: [x] 対応（2026-09-08）
- 該当: 元文書 §2.3。
- 問題: `staleTime` はキャッシュをfreshとみなす期間であり、5分後の自動再取得を保証しない。現行は定期取得・window focusでの再取得を無効にしている。
- 根拠: `frontend/src/features/auth/api/get-me.ts:15`、`frontend/src/features/auth/components/AuthProvider.tsx` の `refreshPermissions`。
- [x] 再取得の契機（明示refresh、query無効化、再マウント等）と、5分経過だけでは更新されないことを記載する。
- [x] UIの古い権限表示とBEの最終認可を分けて説明する。
- [x] 自動ポーリングは本指摘だけを理由に追加しない。必要なら権限変更の反映要件と負荷を別途確認する。

## C. 確定した運用・公開範囲の実装と手順整備

### 8. 既存スタッフへの画面からのアカウント付与と本番初回管理者（仕様確定）

- 対応状況: [x] 部分対応（画面 API・手順文書は実装。本番実付与・メール現地確認は SKIP）
- 該当: 元文書 §5.7。
- 問題: account無しの既存スタッフへのpassword PATCHは `staff does not have an account` で拒否される。一方、画面の新規作成は別staff行を作る。本番の付与手順が不明なままでは、ログイン用スタッフの重複作成や過去の診療担当との分断を招く。
- 根拠: `backend/internal/staff/staff_service_update.go:45-46`、`backend/internal/staff/staff_handler.go:53-63`、[スタッフアカウント発行手順](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md)。STG/UAT専用経路は本番の承認済み経路とみなさない。
- 決定（③-1）: システム管理者がスタッフ画面から、既存staffにaccountを追加する。通常のアカウント付与用の専用コマンドを追加する選択はしていない。
- 決定（④-2）: 最初の本番システム管理者は、指定運用担当者が初回専用手順で作成する。公開の管理者登録画面は作らない。
- 着手プラン: 画面・API契約と対象認可を設計 → staff/accountの原子的な紐付け → 画面実装 → 合成データの検証。初回管理者は別の準備単位として、担当者・対象環境・安全な資格情報受領・既存管理者確認・重複/競合防止・監査を定義した手順を用意する。
- [x] 非システム管理者からの付与をAPIで拒否する。既存staff ID・診療履歴を維持し、新しいstaff行を作らない。
- [x] 既存手順で満たせる範囲と、実装が不足する範囲を分ける。
- [x] 対象staff/clinicの認可・既存accountとの衝突・競合・再実行・監査・失敗時の原子性をテストする。
- [x] 初回専用手順に、実行するSQL・接続先確認・既存管理者/メール重複拒否・既存staffへの付与・テーブルロック・同一transactionの監査・commit後receiptを記載した。公開の管理者登録経路は追加しない。
- [x] 初回手順にstaffを複製しないことと、有効な既存staff/主所属を前提とすることを明記した。
- [ ] 初回SQLを合成データの実DBで検証する。構文・競合・監査rollback・本人ログインは未実測。`testdata/first_system_admin.sql` と `001_init.sql` / `FIRST_SYSTEM_ADMIN.md` の静的照合は実施済み。手順の静的照合を実行成功と扱わない。本番への実付与と対象環境メール実送信は別承認。

### 9. 非システム管理者の医院一覧を所属医院に限定する（仕様確定）

- 対応状況: [x] 対応（2026-09-08）・仕様選択済み
- 該当: 元文書 §3.1・§9.2。
- 問題: 他院の詳細GETは拒否する一方、一覧の `scope=all` は `hospital-settings:view` で所属外医院も返す。全院ディレクトリとして意図された実装だが、公開目的・返却項目と詳細APIの制限の一貫性を確認する必要がある。2の認可漏れとは分けて判断する。
- 根拠: `backend/internal/clinic/clinic_handler.go:14-29,59-83`、`backend/internal/clinic/clinic_repository.go:25-33`。
- 決定（⑤-1）: 非システム管理者には所属医院だけを表示する。所属外医院は名称・IDを含め一覧に出さない。システム管理者の全院管理は維持する。
- 着手プラン: 医院一覧APIと画面の `scope=all` 利用箇所を確認 → 非システム管理者の返却範囲をBEで所属に限定 → 画面の説明を同期 → HTTP/画面テスト。UIのフィルタだけで制限しない。
- [x] 非システム管理者が `scope=all` を直接指定しても所属外医院の情報を取得できない。
- [x] 所属判定にクライアント申告の医院集合を使わず、対象リソースの閲覧権限も維持する。
- [x] 採用した仕様について、単一所属・複数所属・管理者・inactive医院・権限なしのケースをテストする。

## 検証状況と共通完了条件

候補コードの修正証明は worktree をマウントした scoped Docker テストである。共有 `animalekarte-backend-1`（main マウント）では検証しない。

- [x] 指摘に対応する失敗テストを追加し、scoped `go test` / vitest で確認する（結果は PR に記載）。
- [ ] DB競合・ロックの隔離DB検証 — **BLOCKED**: 承認済みの捨てPostgresがない。`old-db-postgres`（`127.0.0.1:15432`）と共有 `ekarte_db` は使わない。`-short` の auth/staff は PASS（DBテストは明示SKIP）。共有DB・migration applyは実施しない。
- [x] 認可漏れは UI 非表示のみにしない（BE 403 / selected-clinic grant）。
- [x] `todo-check-auth.md` と `docs/architecture/auth.md` を同期。候補コードと実環境を区別する。
- [x] 独立レビューの HIGH（復旧 GET が stale JWT で詰まる）を反映。未実施を PASS にしない。
- [ ] Linear 反映 — **SKIP**: 外部チケット未承認。

### 2026-09-09 再レビュー修正の検証

当時の対象は `main` の基点 `db7b6fa24` に対する未コミット差分（後に `950404408` へコミット・main統合済み）。全runnerに当時の作業ツリーをread-only mountし、`--network none`、entrypoint上書きで実行した。

| 検査 | 結果 |
| --- | --- |
| FE 回帰のRED | 全医院失効403、自動復旧再発、ログアウト後応答の3件が修正前にFAIL。追加の「手動再試行中のログアウト」「AuthProviderのlogout待機前キャンセル」も各修正前にFAIL |
| FE GREEN | `clinic-selection-recovery.test.ts`、`clinic-selection-axios.test.ts`、`ClinicSelectionBlockedScreen.test.tsx`、`use-auth-initial-session.test.tsx` の25件PASS |
| FE 静的検査 | 変更6ファイルのESLint（`--max-warnings 0`）とPrettier整形を確認 |
| BE 回帰のRED | policy lock前のstaff row取得・lock取得失敗の無視を検出。既知owners prefix内の未確認経路も旧分類でFAIL |
| BE GREEN | `go test -short ./internal/auth ./internal/staff -count=1` PASS。`go test ./cmd/api -run '^TestGETHEAD' -count=1` PASS。前者はDBケースSKIP |
| BE 静的検査 | `go vet ./internal/auth ./internal/staff ./cmd/api` PASS。実DB並行テストはコンパイル済み・実行BLOCKED |
| 初回管理者手順 | schemaと監査フィールドを静的照合。実SQL・本番付与・本人ログインは未実施 |
| 独立レビュー | logout開始時の復旧中止と、初回SQLの必要権限明示を反映後、今回のロック・復旧・手順書の範囲でBlockなし。DB/runtimeの成立証明ではない |
| 作業差分 | `git diff --check` PASS。commit/push/merge/Linear更新なし |

runner image: FE `animalekarte-frontend:latest`（`sha256:532501622cd024ab786a32eb9798db1cd1a0e4d47cddb3dbd56ae107f95d9cb4`）、BE `animalekarte-backend:latest`（`sha256:6c5b455bf14e3f0ec7bece9ae9a4be2e8ad1441adfb33fdf2ce53d9eaaa22ed4`）。BEは`ekarte-go-mod-cache` / `ekarte-go-build-cache`と`GOPROXY=off`、FEテストは`vitest run <上記4ファイル> --configLoader native`を使用した。過去の PASS 基点 `db7b6fa24` は現 HEAD `950404408` の証拠に流用しない。

### 2026-09-09 残件再開の検証（worktree `AnimalEkarte-todo-fix-auth-remaining`、HEAD `950404408`）

共有 compose は使わず、worktree を bind mount した ephemeral `docker run`（`--network none`、`GOPROXY=off`）で実施した。当時はclaim `claim/TODO-FIX-AUTH` を保持し、commit/push/merge なし。後続の追加確認を経て `9be825a66` へコミット・main統合済み。

| 検査 | 結果 |
| --- | --- |
| BE `-short` | `go test -short -count=1 ./internal/auth ./internal/staff ./internal/clinic ./internal/identitylink ./internal/owner ./internal/pet ./internal/reservation ./internal/billing` PASS。DB並行テストは SKIP |
| medicalrecord スコープ | `go test -short -count=1 ./internal/medicalrecord -run MembershipABGrantASelectedB` PASS。パッケージ全体は docs 未マウントの既存契約テストで FAIL し得るため未実施 |
| GET/HEAD 台帳 | `go test -count=1 ./cmd/api -run '^TestGETHEAD'` PASS（clinic-fixed deny/許可 middleware 全件、横断 allowing 21件） |
| clinic / identitylink / staff / auth grant 分離 | `go test -short -count=1 ./internal/clinic ./internal/identitylink ./internal/staff ./internal/auth -run 'MembershipABGrantASelectedB|MembershipABGrantBSelectedB'` PASS |
| 初回管理者SQL | docs マウントありで `TestFirstSystemAdminProcedureMatchesInitSchema` PASS（testdata = 手順書 SQL、列は `001_init.sql`） |
| `go vet` | `./internal/auth ./internal/staff ./internal/clinic ./internal/identitylink ./cmd/api` PASS（残件再開で再実行）。他パッケージは前回 PASS |
| `git diff --check` | PASS |
| FE 回帰 | `clinic-selection-recovery` related 37 tests PASS（Vitest 6 files） |
| 実DB並行 / 2プロセス失効 / 初回SQL実行 | BLOCKED / SKIP。隔離Postgres・独立APIプロセス未承認 |
| 独立レビュー | 未実施。本セッションの自己確認であり、別 worktree の独立レビューではない |

### 2026-09-09 main統合前の追加確認

上表は終了したCursorセッションの記録。後続の統合作業ではユーザー承認に基づきclaimを引き継ぎ、同じ `950404408` 基点の残件差分を検証した。今回のclaim解除とmain統合は明示承認された例外であり、一般規則は変更しない。

- `cmd/api`、auth、staff、clinic、identitylink、owner、pet、reservation、billing、medicalrecord の10パッケージで `go test -short -count=1` / `go build` / `go vet` PASS。対象worktreeをread-only mountし、networkなしの既存BE imageで実施。実DBテストはSKIP。
- medicalrecordの既存文書契約も必要な2文書をread-only mountし、今回はパッケージ全体の `-short` を実行した。初回管理者SQLの文書も対象worktreeから個別mountし、欠落時は意図どおりFAIL、存在時は静的照合PASSを確認した。
- SQLテストデータと初回管理者手順書を専用の検証契約へ登録。必要な文書だけを読み取り専用で参照し、未登録SQL・文書欠落・symlink・未stageの依存変更を拒否する。文書だけの変更でもbackend依存を検査し、検証中の文書変更をfingerprintへ反映する。検証スクリプトの回帰テスト40件PASS。
- 別worktreeでGoテスト・台帳と検証スクリプトを独立レビューした。横断経路の実DB証明を過大に示す記述、および文書だけの変更時のbackend依存検査を修正し、両レビューApprove。
- 実DB並行、初回SQLの実行、2プロセス即時失効、全経路の実DB返却データ分離、本番操作・メール・Linearは未実施のまま。テスト追加の統合をこれらの完了と扱わない。

### D1–D5 の実施メモ

- D1: 画面・API は実装。メール自動送信なし。本人設定は既存 forgot-password。対象環境のメール実送信は SKIP。初回SQLは testdata と schema/手順書へ静的照合。実DB実行は未実施。
- D2: 医院policy lockで権限変更を直列化し、view+editの自己喪失を拒否。admin免除。2026-09-09: `UpdateRules` と Update-with-Rules が別有効グループの OR 付与を許すよう修正。証拠修復で service TX 状態観測・pg_locks 待機・worker 終了保証をコード化。実DB実行は上記BLOCKED。
- D3: `get_head_permissions.json` の完全一致台帳と登録集合を照合（local 203件、S3 201件）。clinic-fixed は deny（Aのみ grant）と許可（Bに grant）の middleware を全件実行。横断 class allowing/handler は JSON上21件（`GET /clinics` 含む）を実行。実DB臨床返却分離の INCLUDE は 178（clinic-fixed158 + 横断20；`GET /clinics` は所属一覧として除外）。handlerテストはservice stubへの引数と応答の検査であり、clinic-fixed・横断経路とも実DB上の返却データ分離は未検証。未検証0件とはしない。
- D4: `clinic_selection_unavailable` のみ自動復旧を一度実行。ヘッダーなし`/me`も同コードの403なら全医院失効として案内。障害後は手動再試行、再試行中もログアウト可能。ログアウト開始時に復旧を中止。失敗mutationの自動再送なし。セッション未発行loginは復旧対象外。related Vitest 37件 PASS。
- D5: production は uncached resolver。2 インスタンス検証は SKIP。

## 詳細仕様の確定（2026-09-08・ユーザーによる判断委任）

以下は「それらも確定させてください。あなたが判断して」という依頼に基づく採用仕様である。残件には本番操作・メール現地・2プロセス検証・Linearに加え、D2の実DB並行検証、D3の全経路の実データ分離検証、初回管理者SQLの実DB検証がある。ローカル修正・限定テストの完了と区別する。

### D1. アカウント追加画面と初回パスワード（項目8）

- システム管理者が既存スタッフの詳細画面で「ログインアカウントを追加」を操作し、本人専用のメールアドレスを入力する。氏名と所属を確認できるようにする。共用メール・代理受信を標準運用にしない。
- 管理者は初期パスワードを入力・閲覧・配布しない。サーバーが暗号学的乱数から推測不能な初期秘密値を生成してhashのみ保存し、平文を破棄する。本人は既存の「パスワードを忘れた方」からメールを受け、自分でパスワードを設定する。初回設定後は通常ログインへ戻す。
- アカウント追加時にメールの自動送信は行わない。画面には「アカウントを追加しました。本人がログイン画面のパスワード再設定から設定してください」と表示する。「メール送信済み」「ログイン準備完了」とは表示しない。送信失敗や期限切れは既存の再設定経路で再試行する。
- 対象は未削除・有効なstaffで、選択中のactive医院に所属し、account未紐付けのもの。accountは `is_active=true`・`is_system_admin=false` で作る。既存staff ID、氏名、職種、所属、主医院、権限グループ、診療履歴は一切変更しない。権限なしでも勝手にgrantを付けない。
- 既存staffへの専用の追加操作として扱い、プロフィールPATCHや新規staff作成へ混ぜない。staff/account/監査の原子的処理を必須とする。既にaccountがある対象やメール衝突は409、非管理者は403、対象外staffは404。検証エラーは既存APIの入力エラー形式に合わせる。
- [x] 管理者にはパスワード・設定token・設定URLを返さず、ログにも出さない。非本番デモの共有パスワード対象メールは追加不可とする。
- [x] account未設定状態では既知の資格情報でログインできず、本人の有効な一回限りの設定tokenでパスワード設定後にログインできる。期限切れ・再利用は拒否する。初回秘密は返さず、本人は既存 forgot-password を使う。
- [x] 通信失敗後の再操作・同時追加でもaccount/staffが重複せず、既存権限・履歴が不変である。監査失敗時は全体rollbackする。
- [ ] メール送信基盤と本人による設定が対象環境で利用可能なことを運用開始条件とする。未準備なら付与運用を開始せず、共有パスワードで代替しない。 **SKIP**: 対象環境メール実送信は別承認。
- 本番初回管理者はこれとは別の初回専用手順とする。担当者が安全な経路で資格情報を扱い、作成結果の確認と監査を残す。実際の担当者・アドレス・対象環境は本書へ推測記入しない。

### D2. 自己ロックアウト判定の権限定義（項目4）

- 「権限管理を続けられる」は、選択医院で変更後も `master-permission:view` と `master-permission:edit` の両方を実効権限として持つこととする。別の有効グループからのOR付与も含める。
- 非システム管理者による、自分が所属するグループの変更・無効化・自己割当解除では、両方を満たさなくなる操作を403で拒否する。他の担当者が残ることは例外理由にしない。create/deleteまで追加で要求しない。
- システム管理者はグループに依存しないためこの自己喪失判定を免除する。システム管理者アカウントの削除・無効化に関する既存保護は別途維持する。
- [x] viewだけ喪失、editだけ喪失、両方喪失の拒否を単体テストで確認する。グループ間の並行変更も共通policy lockへ参加させる実装に修正した。
- [x] 2026-09-09: `UpdateRules` の旧 `validateNotSelfReference` 事前拒否を外し、別有効グループに view+edit が残る自己ルール変更を成功させる。`permission_group_service_rules_d2_test.go` で RED→GREEN。lookup/audit 失敗の fail-closed と system admin 免除を確認。
- [x] 2026-09-09 evidence-repair: staged/committed TX double で成功時 rules+監査 commit、失敗時 committed 不変を service 経由で観測（実DB原子性ではない）。viewだけ喪失=`CanView=false/CanEdit=true`、editだけ喪失=`CanView=true/CanEdit=false` に対応。並行DBテストは holder/contender `pg_locks` advisory 待機・Cleanup release/cancel/bounded join をコード化（コンパイル確認。実行は未実施）。
- [x] 2026-09-09 cleanup-waiter: auth/staff Cleanup の `go workers.Wait` 追加 waiter を除去。起動済み worker ごとの done channel を親 Cleanup が共有 deadline で直接 select。未起動は待たず、期限超過は worker 名付き FAIL。pg_locks・監査・実効権限 assert は保持。
- [x] 2026-09-09 coordinator: `permission_group_service_mutate.go` Update-with-Rules の旧 `validateNotSelfReference` 事前拒否を外し、`UpdateRules` と同じ post-mutation OR 判定へ揃えた。`permission_group_service_mutate_d2_test.go` で RED→GREEN（修正前の OR 付与自己 strip は `validateNotSelfReference` の InvalidInput；修正後は他グループ OR 付与成功、真の view/edit/both 喪失のみ Forbidden、lookup/audit 失敗 rollback、admin 免除）。旧 helper と単体表テストを除去。
- [ ] 並行変更後も同じ結果となることを実DBテストで確認する（項目4のBLOCKED参照。人間実行・承認済み使い捨てPostgres・共有DB禁止）。入口:
  - 管理接続と `TEST_DATABASE_URL` の両方を確認し、共有 `ekarte_db` / `old_db` / `127.0.0.1:15432` を使わない。AutoMigrate は `001_init.sql` 全制約の成立証拠ではない。
  - `go test -count=1 -timeout=60s -v ./internal/auth -run '^TestPermissionPolicyDB_ConcurrentGroupDeactivation$'`
  - `go test -count=1 -timeout=60s -v ./internal/auth -run '^TestPermissionPolicyDB_ConcurrentRuleReplacement$'`
  - `go test -count=1 -timeout=60s -v ./internal/staff -run '^TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation$'`
  - 停止条件: 共有DB検出、接続先不明、migration apply 要求、`-short` だけの SKIP を PASS 扱いしない。

### D3. GET認可の確認対象を閉じる（項目2）

- 実装の最初の成果物として、`RequirePermission` と `RequirePermissionAny` が適用される登録済みGET/HEADを全件抽出する。列は「method/path、resource/action、handler、医院固定/横断/共有マスタ、認可箇所、返却範囲、回帰テスト」とする。
- 医院固定は選択医院のgrantを必須、横断は返却対象医院ごとのgrantを必須にする。共有マスタは既存の明示的契約だけを採用し、分類不能を許可扱いにしない。HEAD未登録はその旨記録し、新たに追加する必要はない。
- [x] 登録ルートと完全一致台帳の件数・集合・handlerが一致する。既知prefix配下の未登録経路も拒否する。
- [ ] 全経路について認可実行・返却対象医院の実データ分離を検証し、未検証0件とする。clinic-fixed の middleware deny/許可と横断 handler は実行済み。clinic-fixed・横断経路とも実DB上の返却データ分離は残件。service stubの応答、前方一致やテスト名存在だけでは完了にしない。
- [x] 表とテストを既存の認可設計・テスト領域へ集約し、別の実行台帳を増やさない。正常に0行の一覧と、認可可能な医院が0件の拒否を区別する。

### D4. 旧医院選択からの画面復旧（項目3）

- BEは認証失敗、医院選択失効、通常のリソース権限不足を機械判定できる別エラーコードにする。医院選択失効だけを復旧対象とし、任意の403や通信障害を医院切替の根拠にしない。
- 旧医院の失効が判明したら画面の書き込みを停止する。選択医院ヘッダーを省いた `/me` を1回だけ取得し、BEが現在の所属から解決したactiveな既定医院へ戻す。この復旧GETでは旧JWTの既定医院を最終根拠にしない。
- 有効医院があれば選択を保存し、医院依存キャッシュを破棄して再読込する。「選択していた医院は利用できなくなったため、○○医院に切り替えました」と通知する。保存に失敗したら切替成功とせず、書き込み停止のままエラーを表示する。
- 有効医院が0件なら利用不可画面にし、管理者への確認案内とログアウト操作を提供する。復旧GETの障害時も書き込みを再開せず、手動再試行を提供する。自動ループさせない。
- [x] 失敗したPOST/PUT/PATCH/DELETEを、新しい医院へ自動再送しない。旧医院の未保存入力を新医院へ引き継いで保存しない。
- [x] 旧医院だけ失効、全所属失効、アカウント失効、一般の権限不足、通信障害、ストレージ失敗を個別にテストする。

### D5. 即時失効の方式と検証範囲（項目5）

- 認証上の最終判断に2秒current-accessキャッシュを使わず、通常リクエストとrefreshで最新のaccount/staff/医院所属を信頼できるDBから確認する方式を採用する。古い状態を返し得るread replicaや別インスタンスのメモリキャッシュを最終認可にしない。
- 失効の境界は対象変更のDB commitとする。commit後に受け付けるリクエストは変更後の状態で判定し、commit前に受け付けた処理を取消す保証は追加しない。DB確認障害は503で拒否する。
- [ ] 同じDBを使う独立した2つのサーバー実行インスタンスと、別端末相当の2セッションで検証する。両方で旧状態を利用した後に変更し、TTL待ちのsleepなしで拒否を確認する。 **SKIP**: `make up` 禁止。共有 compose は main をマウントし得る。
- [x] スタッフ無効化は両セッションのアクセスを拒否する。所属解除は解除医院だけを拒否し、他の有効所属へのアクセスを維持する。実装は uncached current-access。ライブ2インスタンスは上記 SKIP。
- [x] パスワード変更は本人変更・管理者再設定・再設定tokenの各経路で、旧access tokenと旧refresh tokenを拒否する。新しいパスワードで発行したセッションは利用できる。既存 epoch 契約を維持し、認可は毎回 DB を見る。
- [ ] 即時失効の成立を限定テストで確認したうえで、認証に追加されるDB処理を測定する。性能対策で旧権限を再び許容する仕様へ戻さない。 **SKIP**: 本番相当の負荷測定は起動禁止のため未実施。

参考: [OWASP Authorization Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html)（デフォルト拒否、各リクエストの対象に対する認可確認）。

## Coordinator 実行準備（2026-09-09・計画のみ。実行成功ではない）

計画完成と実行成功を分ける。秘密値は収集・表示しない。本節は準備AC修復（2026-09-09 prep-repair）。**実行していない。**

### 前回ローカル検証証拠（再実行なし・artifact参照）

| 検査 | artifact | 結果 |
| --- | --- | --- |
| verify-agent-task auth short + gofmt | `~/.grok/sessions/.../01a08537-317d-7c03-9483-3dc423aa7a3a/terminal/call-40b27744-d2fd-4279-bce2-4734bd871758-86.log` | `completed_tests: 476`, checks PASS, `"reason": "Selected checks passed"`, exit 0 |
| docker build+vet auth/staff | `.../terminal/call-4013bd17-e95e-44d3-b07d-3d8f374e5563-82.log` | `build_vet_exit=0` |
| santa TX/auth review | `.../subagents/01a0855a-b5d2-7951-960d-67b2cbac04e6/output.json` | `frozen_hashes_checked: true`, verdict APPROVE |
| santa test/evidence review | `.../subagents/01a0855a-b5d2-7951-960d-67cd35ef2644/output.json` | `frozen_hashes_checked: true`, verdict APPROVE |
| RED InvalidInput (Update-with-Rules) | `.../terminal/call-bf8e512a-61a3-4a77-a693-c01e3e0fa13b-54.log` | Allows/SystemAdmin FAIL with `validateNotSelfReference` InvalidInput |
| GREEN after fix | `.../terminal/call-e9d061f8-fa8f-4f9c-a6ee-daaf3bfa606b-68.log`（`UpdateWithRules_D2_|UpdateRules_D2_|...`） | exit 0 / `ok` |

準備単位の完了 ≠ 台帳全体完了。

### D1 初回SQL・メール・本番付与（実行準備）

| 項目 | 内容 |
| --- | --- |
| 入口 | `backend/internal/auth/testdata/first_system_admin.sql`、`docs/ops/deploy/FIRST_SYSTEM_ADMIN.md`、静的 `TestFirstSystemAdminProcedureMatchesInitSchema`（実行済み静的のみ） |
| 必要入力 (UNKNOWN可) | 対象 env 名、`PGSERVICEFILE` 実体 path、承認記録の provider/host/port/db/role、合成 staff_id/clinic_id/email、bcrypt hash、approval_ref、operator_ref、実行オペレータ |
| 接続確認コマンド（文書） | `PGSERVICEFILE=/secure/first-system-admin/pg_service.conf PSQL_HISTORY=/dev/null psql -X -w --dbname='service=first-system-admin' --command='\\conninfo'` → 承認記録と照合。不一致なら停止 |
| 合成 input.csv | ヘッダなし1行・8列（database_name, database_role, staff_id, clinic_id, email, password_hash, approval_ref, operator_ref）。repo外 `0600`、symlink不可、read-only mount `/secure/first-system-admin/input.csv` |
| 正常系手順 | schema=`001_init` 適用済み使い捨てDB → conninfo照合 → SQL実行 → receipt（account_id/staff_id/clinic） |
| 競合試験 | (a) 既存 `is_system_admin` あり (b) 同一 email あり → `bootstrap account conflict` EXCEPTION。部分行なし |
| 監査rollback試験 | 同一TX内で audit INSERT を失敗させる（権限欠落または意図的制約）→ accounts/staffs 不変、admin数0のまま |
| 本人ログイン | 作成後、通常loginで対象email+本人設定パスワード。失敗なら付与未完了 |
| assert | admin有効1、`staffs.account_id` 紐付け、audit 1、失敗時不変、平文秘密がログに出ない |
| 停止 | 接続先/role UNKNOWN、共有パスワード代替、本番付与未承認、メール実送信未承認、staff/主所属未整備 |
| 承認境界 | 合成DB実行 / 本番付与 / 対象環境メール現地確認は別承認。担当者・環境未確定は UNKNOWN |

### D2 実DB並行3テスト（実行準備）

| 項目 | 内容 |
| --- | --- |
| テスト | `TestPermissionPolicyDB_ConcurrentGroupDeactivation` / `ConcurrentRuleReplacement`（auth）、`TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation`（staff） |
| 管理接続 (testdb) | `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME`（default host `db`, db `ekarte_db`）で main DSN に接続し `{DB_NAME}_test` を CREATE。共有運用DB自体でテストしない |
| テスト接続 | 既定は `postgres://…/{DB_NAME}_test`。**`TEST_DATABASE_URL` があればそれを優先**（`backend/internal/testdb/testdb.go` `connectTestDatabase`）。両系統を実行前に確認し、共有運用DB・`old_db`・`127.0.0.1:15432` なら停止 |
| AutoMigrate | `SetupTestDB` は共有ベース AutoMigrate + 毎テスト TRUNCATE。**AutoMigrate は 001_init 全制約の成立証拠ではない**（複合FK/EXCLUDE/triggerは未再現） |
| TRUNCATE | コア表 CASCADE TRUNCATE でテスト間分離。共有プール MaxOpenConns=10 |
| 隔離候補 | 使い捨て Postgres container/volume、backend image digest `sha256:6c5b455bf14e3f0ec7bece9ae9a4be2e8ad1441adfb33fdf2ce53d9eaaa22ed4`、`ekarte-go-mod-cache`。共有 compose main mount 禁止 |
| コマンド（Docker内・直列） | `go test -count=1 -timeout=60s -v ./internal/auth -run '^TestPermissionPolicyDB_ConcurrentGroupDeactivation$'` → 同様に `ConcurrentRuleReplacement` → `./internal/staff -run '^TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation$'` |
| assert | 一方 Forbidden、もう一方成功、actor が view+edit 維持、audit件数、pg_locks holder/contender 待機（テスト内） |
| 終了 | テストプロセス終了、`CloseSharedTestDB`、使い捨てDB drop、接続漏洩確認。`-short` SKIP を PASS にしない |
| 停止 | 共有DB検出、管理接続と TEST_DATABASE_URL の片側のみ確認、migration apply要求、接続先不明 |

### D3-0 集合証明

| 指標 | 値 |
| --- | --- |
| JSON 総数 | 203 |
| 対象 (clinic-fixed + cross-clinic 臨床/横断返却) | 178 (= 158 + 20) |
| 除外 | 25 (= public5 + liff9 + self1 + shared-master9 + `GET /api/v1/clinics` 所属一覧1) |
| 対象欠落 / 未知 class | 0 / 0 |
| method/path 集合 fingerprint (sha256-16) | `30d13703b3cad42b` |

### D3-1 除外一覧（25・全件）

| method/path | class | 除外理由 |
| --- | --- | --- |
| `GET /api/liff/:clinicId/available-dates` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/available-times` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/courses` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/health-card` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/my-reservations` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/profile` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/staffs` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/trimming-courses` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/trimming-options` | liff | LIFF顧客スコープ。スタッフ選択医院 grant の臨床分離対象外 |
| `GET /api/liff/:clinicId/settings` | public | 公開/ヘルス/アップロード。選択医院の臨床返却分離対象外 |
| `GET /api/v1/health` | public | 公開/ヘルス/アップロード。選択医院の臨床返却分離対象外 |
| `GET /health` | public | 公開/ヘルス/アップロード。選択医院の臨床返却分離対象外 |
| `GET /uploads/*filepath` | public | 公開/ヘルス/アップロード。選択医院の臨床返却分離対象外 |
| `HEAD /uploads/*filepath` | public | 公開/ヘルス/アップロード。選択医院の臨床返却分離対象外 |
| `GET /api/v1/me` | self | 認証本人 `/me`。医院返却データの横断分離対象外 |
| `GET /api/v1/clinics` | cross-clinic | 所属医院ディレクトリ一覧。臨床データの返却医院分離対象外（所属集合の別契約） |
| `GET /api/v1/company` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/lstep-tag-config/auto-managed-prefixes` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/lstep-tag-config/condition-tag-mappings` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/lstep-tag-config/send-purpose-tag-prefixes` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/manual/articles` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/manual/articles/:category/:slug` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/manual/articles/:category/:slug/versions` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/masters/animal-species` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |
| `GET /api/v1/masters/animal-species/:id` | shared-master | 共有マスタ/マニュアル。clinic query 次元なし |

### D3-2 共通 fixture と期待応答契約（対象178）

共通 fixture:
- Clinics A=1 / B=2 active。Staff S が A+B 所属（`clinic_ids=[1,2]`）。非 admin。
- Grant セット例: **Aのみ**（負の既定） / **A+B**（正）。Selected clinic header 既定は **B=2**（特記なき限り）。
- 正系では観測可能な B seed を置く（空 list の自明 PASS 禁止）。実DB返却分離は **NOT_RUN**。

期待応答は経路種別で分岐する（全横断を一律403/filterにしない）:

| 区分 | 代表ヘルパ | 負 (selected B・grant Aのみ) の代表 | 正 (Bにも grant + B seed) | 観測 |
| --- | --- | --- | --- | --- |
| **clinic-fixed** (158) | `RequirePermission` selected | **HTTP 403**（middleware deny。handler 未到達） | **HTTP 200** + B seed。A seed の clinic-owned 行を含まない | JSON `clinic_id` / entity ID |
| **cross-clinic list** (`ResolveListClinicIDsForPermission`) | `httpapi/clinic_permission.go` Filter 空→403 | **既定 selected B**: query `clinic_ids` なし → 対象=[B] → filter 空 → **403**、service 未呼出。**混合 `clinic_ids=1,2`**: filter→[A] → **200**（A空可。`"total":0`）。**明示 `clinic_ids=2`**: **403** | 200。応答に B seed≥1。要素 clinic ⊆ 許可集合 | list 要素の clinic 所属 |
| **cross-clinic detail** (`ResolveAllClinicIDsForPermission` + `*ForClinics`) | 同上 + `FromGORM` NotFound | grant 集合で service 呼出。対象が許可 clinic 外/不存在なら **404**（Bデータを返さない）。grant 0件なら **403**（service 未呼出） | 200 で B seed detail | ID と clinic 所属 |
| **cross-clinic 集計** (`daily-summary` 等 list ヘルパ) | list と同型 | selected B 既定は **403**（service 未呼出）。混合 query は A のみ集計して **200** | 200。B seed が集計に反映（ゼロ自明PASS禁止） | 集計値 |
| **cross-clinic path-clinic authorize** (`AuthorizeClinicIDsForPermission` on `:clinic_id`) | `line_reservation_setting_handler.go` | **path=B** かつ grant なし → **403**、service 未呼出（list/filter ではない） | **path=B** かつ B grant → 200/204 | path clinic の設定 body |
| **cross-clinic identity-link** (`FilterClinicIDsForPermission`→`VerifiedClinics`) | `identitylink/handler.go` + service | grant 0 → **403**。path clinic が VerifiedClinics 外 → service 呼出後 **403**。可視 member 0 の group → **404** | 200。member/履歴は許可 clinic のみ | members[].clinic_id / items |

clinic-fixed で「403 または空」の曖昧 OR は採用しない。middleware deny が成立する経路は **403 のみ**。
横断でも **403/404/200 を OR で書かず**、条件行を分割する（D3-3）。

### D3-3 cross-clinic 対象（20）— 条件別契約

凡例: selected/path/query/target は clinic ID。grant は resource:action を持つ clinic 集合。`service_called` は handler が domain service に到達したか。evidence は TEST（既存 stub テスト名）または STATIC_DERIVED。realDB=**NOT_RUN**。

| method/path | scenario_id | selected | path | query clinic_ids | target | grant | service_result | status | body/allowed | service_called | source | evidence | realDB |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/accountings` | ACC-LIST-DEF-B-403 | 2 | — | (none→B) | B | accounting:view@{1} | n/a | **403** | forbidden。B一覧なし | no | `accounting_handler.go:37-44` + `clinic_permission.go:141-144` | TEST `TestListAccountings_MembershipABGrantASelectedB` defaults L27-40 | NOT_RUN |
| `GET /api/v1/accountings` | ACC-LIST-MIX-200 | 2 | — | 1,2 | A(filtered) | accounting:view@{1} | empty A list | **200** | total=0。allowed={1} | yes(clinic=1) | `accounting_handler.go:37-82` | TEST same filters mixed L43-58 | NOT_RUN |
| `GET /api/v1/accountings` | ACC-LIST-B-ONLY-403 | 2 | — | 2 | B | accounting:view@{1} | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` + `context.go:175-193` | STATIC_DERIVED（owners 明示Bと同型 helper） | NOT_RUN |
| `GET /api/v1/accountings` | ACC-LIST-AB-DEF-B-200 | 2 | — | (none→B) | B | accounting:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={2} | yes(clinic=2) | `accounting_handler.go:37-82` + `context.go:175-183` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings` | ACC-LIST-AB-MIX-200 | 2 | — | 1,2 | A+B | accounting:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={1,2} | yes(ListForClinics) | `accounting_handler.go:37-82` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings/:id` | ACC-GET-A-200 | 2 | id∈A | — | A | accounting:view@{1} | billing clinic=1 | **200** | clinic_id≠2 | yes(clinicIDs=[1]) | `accounting_handler.go:88-106` | TEST `TestGetAccounting_MembershipABGrantASelectedB` L60-76 | NOT_RUN |
| `GET /api/v1/accountings/:id` | ACC-GET-B-404 | 2 | id∈Bのみ | — | B | accounting:view@{1} | scoped miss | **404** | B bodyなし | yes(clinicIDs=[1]) | `accounting_handler.go:101-104` + `accounting_repository.go:379-401` + `apperrors.FromGORM` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings/:id` | ACC-GET-B-200 | 2 | id∈B | — | B | accounting:view@{1,2} | billing clinic=2 | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `accounting_handler.go:88-106` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings/:id` | ACC-GET-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings/daily-summary` | ACC-SUM-DEF-B-403 | 2 | — | (none→B) | B | accounting:view@{1} | n/a | **403** | — | no | `accounting_handler.go:432-439` | TEST `TestGetDailySummary_MembershipABGrantASelectedB` defaults L81-94 | NOT_RUN |
| `GET /api/v1/accountings/daily-summary` | ACC-SUM-MIX-200 | 2 | — | 1,2 | A | accounting:view@{1} | A summary | **200** | billing_count for A | yes(clinic=1) | `accounting_handler.go:432-450` | TEST same filters mixed L97-110 | NOT_RUN |
| `GET /api/v1/accountings/daily-summary` | ACC-SUM-AB-200 | 2 | — | 含B | B | accounting:view@{1,2} | B seed反映 | **200** | 非自明B値 | yes | `accounting_handler.go:432-459` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/accountings/daily-summary` | ACC-SUM-B-ONLY-403 | 2 | — | 2 | B | accounting:view@{1} | n/a | **403** | — | no | `accounting_handler.go:432-439` + `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/clinics/:clinic_id/line-reservation-settings` | LRS-PATH-B-403 | 2 | 2 | — | B | hospital-settings:view@{1} | n/a | **403** | — | **no** | `line_reservation_setting_handler.go:25-40` | TEST `TestGetLineReservationSetting_MembershipABGrantASelectedB` rejects path B L19-38 | NOT_RUN |
| `GET /api/v1/clinics/:clinic_id/line-reservation-settings` | LRS-PATH-A-200 | 2 | 1 | — | A | hospital-settings:view@{1} | status=running | **200** | clinic1 setting | yes(clinic=1) | `line_reservation_setting_handler.go:37-51` | TEST same allows path A L41-62 | NOT_RUN |
| `GET /api/v1/clinics/:clinic_id/line-reservation-settings` | LRS-PATH-A-204 | 2 | 1 | — | A | hospital-settings:view@{1} | nil setting | **204** | empty | yes | `line_reservation_setting_handler.go:47-49` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/clinics/:clinic_id/line-reservation-settings` | LRS-PATH-B-200 | 2 | 2 | — | B | hospital-settings:view@{1,2} | setting exists | **200** | B setting JSON | yes | `line_reservation_setting_handler.go:37-51` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/clinics/:clinic_id/line-reservation-settings` | LRS-PATH-B-204 | 2 | 2 | — | B | hospital-settings:view@{1,2} | nil setting | **204** | empty | yes | `line_reservation_setting_handler.go:47-49` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records` | MR-LIST-DEF-B-403 | 2 | — | (none→B) | B | medical-records:view@{1} | n/a | **403** | — | no | `medical_record_handler.go:28-35` | TEST `TestListMedicalRecords_MembershipABGrantASelectedB` L27-40 | NOT_RUN |
| `GET /api/v1/medical-records` | MR-LIST-MIX-200 | 2 | — | 1,2 | A | medical-records:view@{1} | empty | **200** | total=0 allowed={1} | yes | same L28-62 | TEST same L43-58 | NOT_RUN |
| `GET /api/v1/medical-records` | MR-LIST-AB-DEF-B-200 | 2 | — | (none→B) | B | medical-records:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={2} | yes | `medical_record_handler.go:28-62` + `context.go:175-183` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records` | MR-LIST-AB-MIX-200 | 2 | — | 1,2 | A+B | medical-records:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={1,2} | yes | `medical_record_handler.go:28-62` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records` | MR-LIST-B-ONLY-403 | 2 | — | 2 | B | medical-records:view@{1} | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` + `context.go:175-193` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records/:id` | MR-GET-A-200 | 2 | id∈A | — | A | medical-records:view@{1} | record A | **200** | clinic_id≠2 | yes([1]) | `medical_record_handler.go:68-84` | TEST `TestGetMedicalRecord_MembershipABGrantASelectedB` L60-76 | NOT_RUN |
| `GET /api/v1/medical-records/:id` | MR-GET-B-404 | 2 | id∈B | — | B | medical-records:view@{1} | miss | **404** | Bなし | yes([1]) | GetByIDForClinics+FromGORM | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records/:id` | MR-GET-B-200 | 2 | id∈B | — | B | medical-records:view@{1,2} | record B | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `medical_record_handler.go:68-84` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/medical-records/:id` | MR-GET-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners` | OWN-LIST-DEF-B-403 | 2 | — | (none→B) | B | owners:view@{1} | n/a | **403** | — | no | `http_owner.go:17-24` | TEST `TestListOwners_MembershipABGrantASelectedB` L27-40 | NOT_RUN |
| `GET /api/v1/owners` | OWN-LIST-MIX-200 | 2 | — | 1,2 | A | owners:view@{1} | empty | **200** | total=0 | yes | same | TEST L43-57 | NOT_RUN |
| `GET /api/v1/owners` | OWN-LIST-B-ONLY-403 | 2 | — | 2 | B | owners:view@{1} | n/a | **403** | — | no | same | TEST L59-72 | NOT_RUN |
| `GET /api/v1/owners` | OWN-LIST-AB-DEF-B-200 | 2 | — | (none→B) | B | owners:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={2} | yes | `http_owner.go:17-42` + `context.go:175-183` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners` | OWN-LIST-AB-MIX-200 | 2 | — | 1,2 | A+B | owners:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={1,2} | yes | `http_owner.go:17-42` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners/:id` | OWN-GET-A-200 | 2 | id∈A | — | A | owners:view@{1} | owner A | **200** | clinic_id≠2 | yes([1]) | `http_owner.go:48-65` | TEST `TestGetOwner_MembershipABGrantASelectedB` L75-93 | NOT_RUN |
| `GET /api/v1/owners/:id` | OWN-GET-B-404 | 2 | id∈B | — | B | owners:view@{1} | miss | **404** | Bなし | yes([1]) | GetByIDForClinics+FromGORM | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners/:id` | OWN-GET-B-200 | 2 | id∈B | — | B | owners:view@{1,2} | owner B | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `http_owner.go:48-65` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners/:id` | OWN-GET-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners/:id/report/pets` | OWN-RPT-A-200 | 2 | owner∈scope | — | A-owned pets | owners:view@{1} | A-scope pets (empty allowed) | **200** | clinic_id≠2 | yes([1]) | `pet_handler.go:122-141` | TEST `TestListOwnerReportPets_MembershipABGrantASelectedB` L79-94 | NOT_RUN |
| `GET /api/v1/owners/:id/report/pets` | OWN-RPT-B-200 | 2 | owner∈B-scope | — | B-owned pets | owners:view@{1,2} | B pets≥1 | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `pet_handler.go:122-141` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/owners/:id/report/pets` | OWN-RPT-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | — | no | `ResolveAllClinicIDsForPermission` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets` | PET-LIST-DEF-B-403 | 2 | — | (none→B) | B | owners:view@{1} | n/a | **403** | — | no | `pet_handler.go:92-98` | TEST `TestListPets_MembershipABGrantASelectedB` L27-40 | NOT_RUN |
| `GET /api/v1/pets` | PET-LIST-MIX-200 | 2 | — | 1,2 | A | owners:view@{1} | empty | **200** | total=0 | yes | same | TEST L43-57 | NOT_RUN |
| `GET /api/v1/pets` | PET-LIST-AB-DEF-B-200 | 2 | — | (none→B) | B | owners:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={2} | yes | `pet_handler.go:92-118` + `context.go:175-183` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets` | PET-LIST-AB-MIX-200 | 2 | — | 1,2 | A+B | owners:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={1,2} | yes | `pet_handler.go:92-118` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets` | PET-LIST-B-ONLY-403 | 2 | — | 2 | B | owners:view@{1} | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` + `context.go:175-193` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets/:id` | PET-GET-A-200 | 2 | id∈A | — | A | owners:view@{1} | pet A | **200** | name A | yes([1]) | `pet_handler.go:147-164` | TEST `TestGetPet_MembershipABGrantASelectedB` L60-76 | NOT_RUN |
| `GET /api/v1/pets/:id` | PET-GET-B-404 | 2 | id∈B | — | B | owners:view@{1} | miss | **404** | Bなし | yes([1]) | GetByIDForClinics+FromGORM | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets/:id` | PET-GET-B-200 | 2 | id∈B | — | B | owners:view@{1,2} | pet B | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `pet_handler.go:147-164` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/pets/:id` | PET-GET-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations` | RSV-LIST-DEF-B-403 | 2 | — | (none→B) | B | reservations:view@{1} | n/a | **403** | — | no | `reservation_handler.go:57-64` | TEST `TestListReservations_MembershipABGrantASelectedB` L28-41 | NOT_RUN |
| `GET /api/v1/reservations` | RSV-LIST-MIX-200 | 2 | — | 1,2 | A | reservations:view@{1} | empty | **200** | total=0 | yes | same | TEST L44-58 | NOT_RUN |
| `GET /api/v1/reservations` | RSV-LIST-AB-DEF-B-200 | 2 | — | (none→B) | B | reservations:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={2} | yes | `reservation_handler.go:57-83` + `context.go:175-183` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations` | RSV-LIST-AB-MIX-200 | 2 | — | 1,2 | A+B | reservations:view@{1,2} | B seed≥1 | **200** | B seed 含む。allowed={1,2} | yes | `reservation_handler.go:57-83` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations` | RSV-LIST-B-ONLY-403 | 2 | — | 2 | B | reservations:view@{1} | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` + `context.go:175-193` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations/:id` | RSV-GET-A-200 | 2 | id∈A | — | A | reservations:view@{1} | reservation A | **200** | notes A | yes([1]) | `reservation_handler.go:89-106` | TEST `TestGetReservation_MembershipABGrantASelectedB` L61-77 | NOT_RUN |
| `GET /api/v1/reservations/:id` | RSV-GET-B-404 | 2 | id∈B | — | B | reservations:view@{1} | miss | **404** | Bなし | yes([1]) | GetByIDForClinics+FromGORM | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations/:id` | RSV-GET-B-200 | 2 | id∈B | — | B | reservations:view@{1,2} | reservation B | **200** | clinic_id=2 B seed | yes(clinicIDs⊇{2}) | `reservation_handler.go:89-106` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/reservations/:id` | RSV-GET-ZERO-403 | 2 | any | — | — | (none) | n/a | **403** | forbidden | no | `clinic_permission.go:141-144` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owners/search` | IL-OWN-SEARCH-200 | 2 | — | — | Verified={1} | identity-links:view@{1} | items=[] | **200** | items; allowed={1} | yes | `handler.go:66-82` + `51-62` | TEST `TestSearchOwners_MembershipABGrantASelectedB` L99-113 | NOT_RUN |
| `GET /api/v1/identity-links/owners/search` | IL-OWN-SEARCH-AB-200 | 2 | — | — | Verified={1,2} | identity-links:view@{1,2} | items include B | **200** | items[].clinic_id に2; allowed={1,2} | yes | `handler.go:66-82` + `51-62` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owners/search` | IL-OWN-SEARCH-0-403 | 2 | — | — | — | (none) | n/a | **403** | — | no | `handler.go:51-62` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/search` | IL-PET-SEARCH-200 | 2 | — | — | Verified={1} | identity-links:view@{1} | items=[] | **200** | items | yes | `handler.go:86-102` | TEST `TestSearchPets_MembershipABGrantASelectedB` L137-151 | NOT_RUN |
| `GET /api/v1/identity-links/pets/search` | IL-PET-SEARCH-AB-200 | 2 | — | — | Verified={1,2} | identity-links:view@{1,2} | items include B | **200** | items[].clinic_id に2; allowed={1,2} | yes | `handler.go:86-102` + `51-62` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/search` | IL-PET-SEARCH-0-403 | 2 | — | — | — | (none) | n/a | **403** | — | no | `handler.go:51-62` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owner-groups/:id` | IL-OG-A-200 | 2 | group | — | visible∈A | identity-links:view@{1} | group+A members | **200** | members clinic⊆{1} | yes | `handler.go:106-121` + `service.go:114-135` | TEST `TestGetOwnerGroup_MembershipABGrantASelectedB` L154-169 | NOT_RUN |
| `GET /api/v1/identity-links/owner-groups/:id` | IL-OG-AB-200 | 2 | group | — | visible includes B | identity-links:view@{1,2} | group+B members | **200** | members clinic_id に2; allowed={1,2} | yes | `handler.go:106-121` + `service.go:114-135` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owner-groups/:id` | IL-OG-BONLY-404 | 2 | group | — | members⊆B | identity-links:view@{1} | visible=0 | **404** | — | yes | `service.go:122-129` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pet-groups/:id` | IL-PG-A-200 | 2 | group | — | visible∈A | identity-links:view@{1} | group+A members | **200** | members⊆{1} | yes | `handler.go:215-230` + `service.go:138-157` | TEST `TestGetPetGroup_MembershipABGrantASelectedB` L172-188 | NOT_RUN |
| `GET /api/v1/identity-links/pet-groups/:id` | IL-PG-AB-200 | 2 | group | — | visible includes B | identity-links:view@{1,2} | group+B members | **200** | members clinic_id に2; allowed={1,2} | yes | `handler.go:215-230` + `service.go:138-157` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pet-groups/:id` | IL-PG-BONLY-404 | 2 | group | — | members⊆B | identity-links:view@{1} | visible=0 | **404** | — | yes | `service.go:150-151` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owners/:clinic_id/:owner_id/group` | IL-OWN-PATH-B-403 | 2 | clinic=2 | — | B | identity-links:view@{1} | outside scope | **403** | — | **yes** (Verified={1}) | `handler.go:125-145` + `service.go:169` | TEST `TestFindOwnerGroupByMember_MembershipABGrantASelectedB` L116-134 | NOT_RUN |
| `GET /api/v1/identity-links/owners/:clinic_id/:owner_id/group` | IL-OWN-PATH-B-200 | 2 | clinic=2 | — | B member | identity-links:view@{1,2} | group visible | **200** | members⊆{1,2} にB | yes | `handler.go:125-145` + `service.go:160-178` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owners/:clinic_id/:owner_id/group` | IL-OWN-PATH-A-404 | 2 | clinic=1 | — | A miss | identity-links:view@{1} | no membership | **404** | — | yes | `service.go:176` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/owners/:clinic_id/:owner_id/group` | IL-OWN-PATH-A-200 | 2 | clinic=1 | — | A member | identity-links:view@{1} | group visible | **200** | members⊆{1} | yes | `handler.go:125-145` + `service.go:160-178` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/group` | IL-PET-PATH-B-403 | 2 | clinic=2 | — | B | identity-links:view@{1} | outside scope | **403** | — | yes | `handler.go:234-254` + `service.go:190` | TEST `TestFindPetGroupByMember_MembershipABGrantASelectedB` L190-209 | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/group` | IL-PET-PATH-B-200 | 2 | clinic=2 | — | B member | identity-links:view@{1,2} | group visible | **200** | members⊆{1,2} にB | yes | `handler.go:234-254` + `service.go:181-199` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/group` | IL-PET-PATH-A-404 | 2 | clinic=1 | — | A miss | identity-links:view@{1} | no membership | **404** | — | yes | `service.go:197` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/group` | IL-PET-PATH-A-200 | 2 | clinic=1 | — | A member | identity-links:view@{1} | group visible | **200** | members⊆{1} | yes | `handler.go:234-254` + `service.go:181-199` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history` | IL-HIST-SEED-B-403 | 2 | clinic=2 | — | B seed | identity-links:view@{1} | outside scope | **403** | — | yes | `service_history.go:21-22` | TEST `TestListLinkedTreatmentHistory_MembershipABGrantASelectedB` L211-229 | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history` | IL-HIST-SEED-B-200 | 2 | clinic=2 | — | B seed | identity-links:view@{1,2} | history includes B | **200** | clinic_id=2 rows; allowed⊇{2} | yes | `service_history.go:11-44` | STATIC_DERIVED | NOT_RUN |
| `GET /api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history` | IL-HIST-SEED-A-200 | 2 | clinic=1 | — | A seed | identity-links:view@{1} | history⊆allowed | **200** | clinic⊆{1} | yes | `service_history.go:11-44` | STATIC_DERIVED | NOT_RUN |

**旧台帳との矛盾（修正点）**: `GET /api/v1/clinics/:clinic_id/line-reservation-settings` を「selected B 既定一覧 / 混合 clinic_ids filter」と書いていた記述は誤り。本経路は path `:clinic_id` の `AuthorizeClinicIDsForPermission` であり、path=B×grantAは **403・service未呼出**、path=A×grantAは **200**（A path 成功は B grant 成功の証拠ではない）。

handlerSource 参照（inventory）: accountings*=`internal/billing/accounting_handler.go`; line-reservation-settings=`internal/reservation/line_reservation_setting_handler.go`; identity-links*=`internal/identitylink/handler.go`; medical-records*=`internal/medicalrecord/medical_record_handler.go`; owners*=`internal/owner/http_owner.go`; owners/:id/report/pets・pets*=`internal/pet/pet_handler.go`; reservations*=`internal/reservation/reservation_handler.go`。

集合照合: inventory cross-clinic 21 − `GET /api/v1/clinics` = **20**。上表の unique method/path = 20。missing=0 / extra=0。

### D3-3 単位結果（契約表・docs-only）

- 状態: **単位 COMPLETE（条件別契約表・condition-gap repair）** / 台帳 overall **未完了**（実DB返却分離・D3テスト実装は残件）
- 変更ファイル: `todo-fix-auth.md` のみ（D3-3表 + 本結果節）
- 本単位で閉じた gap: **18**（final-byte audit の 8 HIGH + 10 MEDIUM; LRS LOW は非欠陥のため非対象）
  - detail B-grant 正系 5: ACC/MR/OWN/PET/RSV `*-GET-B-200`
  - identity-link path/history B-grant 正系 3: `IL-OWN-PATH-B-200` / `IL-PET-PATH-B-200` / `IL-HIST-SEED-B-200`
  - list AB request shape 分割 5→10: `*-LIST-AB-DEF-B-200` + `*-LIST-AB-MIX-200`（`(none→B) or 含B` 除去）
  - observable B 正系 5: `OWN-RPT-B-200` / `IL-OWN-SEARCH-AB-200` / `IL-PET-SEARCH-AB-200` / `IL-OG-AB-200` / `IL-PG-AB-200`
- inventory: unique=20 / missing=0 / extra=0（`GET /api/v1/clinics` 除外維持）
- 証拠区分: 追加行はすべて STATIC_DERIVED。既存 TEST 負系/コントロール維持。realDB=**NOT_RUN**。新規テスト **unimplemented**。
- LRS 行は内容変更なし（行位置移動のみあり得る）。
- set/format/scope: table-bounded 20/0/0、scenario_id unique、mandatory columns、combined-request-condition=0、`git diff --check -- todo-fix-auth.md` exit 0、foreign WIP preserved。
- D3-3表セクション hash（単位結果節を含まない）: sha1=`9bd3b02d87b27042566e0191619dfb3298438777` / sha256=`32b71083d0b7040d8078836538b464caf3eb9be4502de6a0983285c1d3a1d8ad`
- file-level freeze: parent and shell-capable reviewers recompute live `git hash-object todo-fix-auth.md` and `shasum -a 256 todo-fix-auth.md` and must match each other (self-describing file hash is not embedded).
- Independent Review: pending shell-capable A/B
- ledger overall: **INCOMPLETE**（realDB isolation / D3 test implementation 残件）

### D3-4 clinic-fixed 対象（158）— handlerSource package 集約

URL prefix から package を推測しない。配置は JSON `handlerSource` の実 package。親 directory は全候補で実在確認済み。

| handlerSource package | 件数 | 代表 path 例 | 負 | 正/観測 | 後続テスト候補（未作成） |
| --- | --- | --- | --- | --- | --- |
| `backend/internal/medicalrecord/` (61) | 61 | `/api/v1/checkups`, `/api/v1/checkups/field-results`, `/api/v1/examinations`； cages=`cage_handler.go`; | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/medicalrecord/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/lstep/` (28) | 28 | `/api/v1/clinics/:clinic_id/line-customers`, `/api/v1/clinics/:clinic_id/lstep-settings`, `/api/v1/clinics/:clinic_id/lstep-tag-code-mappings`； shared-files=`shared_file_handler.go`; | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/lstep/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/billing/` (22) | 22 | `/api/v1/accountings/:id/refunds`, `/api/v1/accountings/unpaid`, `/api/v1/accountings/unpaid-balance`； reports=`accounting_report_handler.go`; | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/billing/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/reservation/` (12) | 12 | `/api/v1/clinics/:clinic_id/reservation-staffs`, `/api/v1/clinics/:clinic_id/reservation-staffs/:staffId/schedules`, `/api/v1/clinics/:clinic_id/reservation-types` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/reservation/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/staff/` (12) | 12 | `/api/v1/masters/occupations`, `/api/v1/masters/occupations/:id`, `/api/v1/masters/staffs` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/staff/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/trimming/` (8) | 8 | `/api/v1/masters/trimming-course-types`, `/api/v1/masters/trimming-course-types/:id`, `/api/v1/masters/trimming-courses` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/trimming/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/clinic/` (5) | 5 | `/api/v1/clinic-holidays`, `/api/v1/clinics/:clinic_id`, `/api/v1/closing-settings` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/clinic/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/inventory/` (4) | 4 | `/api/v1/inventory`, `/api/v1/inventory/:id`, `/api/v1/masters/merchandise-items` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/inventory/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/pet/` (4) | 4 | `/api/v1/owners/:id/shared-pets`, `/api/v1/pets/:id/chronic-conditions`, `/api/v1/pets/:id/first-visit` | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/pet/realdb_selected_clinic_b_grant_a_isolation_test.go` |
| `backend/internal/auth/` (2) | 2 | `/api/v1/masters/permission-groups`, `/api/v1/masters/permission-groups/:id`； permission-groups=`http_permission.go`; | **403** | **200** + B seed 存在、A seed 非混入。CSV/署名URL等は B 対象 ID で観測 | `backend/internal/auth/realdb_selected_clinic_b_grant_a_isolation_test.go` — **test implemented / compile checked / realDB NOT_RUN** |

clinic-fixed 件数合計 = 158（158）。全候補親 directory `Path.is_dir` = true。`backend/internal/masters` / `report` / `file` / `hospitalization` / `lab` は **不存在**のため候補に使わない。

経路→候補の辿り方: `get_head_permissions.json` の method/path → `handlerSource` → 上表 package の単一テストファイル候補。

### D3-4 単位結果（auth package 2経路）

- 状態: **auth 2経路 test implemented / offline compile checked** / realDB **NOT_RUN** / 台帳 overall **INCOMPLETE**
- 変更ファイル: `backend/internal/auth/realdb_selected_clinic_b_grant_a_isolation_test.go`（新規）+ 本節/auth行のみ
- 対象: `GET /api/v1/masters/permission-groups` と `GET /api/v1/masters/permission-groups/:id`
- 配線: `testdb.SetupTestDB` + `NewPermissionGroupRepository` + `NewPermissionGroupService` + `NewHTTPHandler`（list/detail payload は実DB。checker は deterministic callback）
- ケース: grantA+selectedB list/detail **403**（B query guard）; grantB+selectedB list nonempty B / detail B **200**（A id/name/clinic_id 非混入）; selectedB+A-id detail **404**（A body なし）
- offline: image `sha256:6c5b455bf14e3f0ec7bece9ae9a4be2e8ad1441adfb33fdf2ce53d9eaaa22ed4` + volume `ekarte-go-mod-cache` + `--network none` + `go test -short ./internal/auth` exit 0; 新規5件は `database tests are skipped in -short / offline verification` で **SKIP**; gofmt clean
- realDB runtime: **NOT_RUN / BLOCKED** — 承認済み捨てPostgreSQLの明示認可が必要。将来 regex: `TestRealDB_PermissionGroups_`
- 残件: clinic-fixed 残り **156** unimplemented/NOT_RUN; cross-clinic **20** unimplemented/NOT_RUN; D3 overall **INCOMPLETE**
- Independent Review: R1=`01a086e5-c305-7af1-9155-8fdc11d760d1` + R2=`01a086e5-c305-7af1-9155-8fe6c6f0e3f7` APPROVE (0 CRITICAL/HIGH/MEDIUM) on identical bytes test=`193be05ffc12aa867d722cc43316e14ee8b066fa5b0d320857b74356a48267d9` / pre-stamp ledger=`e3430643b828e9d512e37b0e3607de4d8451d3fb834894d558597b58912291c9` / patch=`7f41ea12e65d22a8187add72daad8321c288543e7cacf0a2b48a2404fe51b41c` (round-1 workflow reviewers BLOCKED on missing shell; round-2 spawn_subagent cleared)

### D5 独立2プロセス・性能（設計条件・実測は未測定）

| 設計項目 | 採用値（今確定） | 根拠 | 実測 |
| --- | --- | --- | --- |
| トポロジ | 同一隔離DB + **独立2 APIプロセス** + **2セッション**（別cookie jar） | ①-1 即時失効・多インスタンス | 未測定 |
| 変更3種 | (1) staff `is_active=false` (2) staff_clinic_assignment 解除 (3) password/epoch 更新 | 項目5確定範囲 | 未測定 |
| 正アサート | commit後の**次リクエスト**で旧session拒否（401/403）。TTL sleep **禁止** | production uncached resolver | 未測定 |
| 負アサート | 未変更医院への正当アクセス維持（所属解除時）；新passwordで再login成功 | 仕様 | 未測定 |
| 測定エンドポイント | `GET /api/v1/me` と clinic-fixed 1本（例: staffs list） | 認可DB再読の代表 | 未測定 |
| メトリクス | サーバ処理時間 p50/p95、DB statements/request（`pg_stat_statements` またはSQL log count）、error率 | 追加DBコスト可視化 | 未測定 |
| 負荷 | セッションあたり **1並行**、ウォームアップ5 + 本測定 **30** 反復、インターバル50ms | 起動禁止下の最小再現；soakしない | 未測定 |
| **安全性判定（独立）** | 旧権限・旧sessionの許容（失効後も通る）= **即 FAIL / 即停止**。性能閾値は参照しない | 認可回帰は性能と無関係 | 未測定 |
| **性能判定（独立）** | 同fixture 1プロセス p95 をベースライン。2プロセス p95 が **+100%超**のみ → **要調査**（FAILにしない）。+200%超または測定不能 → 性能ゲート停止 | 性能悪化だけで認可FAILにしない | 未測定 |
| 証拠 | 生latency一覧、要約表、プロセスPID、DB名、コミット前後の応答status | 再現 | 未測定 |
| 停止（環境） | 共有 compose/main mount、`make up` 必須化、TTL sleep、cache-hit試験復活、本番負荷 | 安全境界 | — |

真理値表:

| 旧権限許容（認可回帰） | 性能 +100%超 | 結果 |
| --- | --- | --- |
| あり | 任意 | **即 FAIL/停止**（性能条件不要） |
| なし | なし | PASS（測定としては完了） |
| なし | あり（+100%〜+200%） | 認可 PASS、性能 **要調査** |
| なし | あり（+200%超/測定不能） | 認可 PASS、性能ゲート停止 |

実装済み参照: `backend/cmd/api/composition_auth.go` uncached、`composition_auth_cache_contract_test.go`。

### Linear ローカル投稿本文ドラフト（投稿しない）

- Ticket ID: **UNKNOWN**（本セッションに Linear issue MCP なし）
- 貼付本文:

```text
タイトル: 認証・認可レビュー修正 — ローカル完了と実環境残件

完了（ローカル）:
- 項目1/3/6/7/9 および D4 復旧フロー
- 項目4 自己ロックアウト: UpdateRules + Update-with-Rules を post-mutation OR(view+edit) に統一。mutate_d2_test RED(InvalidInput)→GREEN。scoped verify auth 476 PASS、build/vet 0、独立2review APPROVE
- 項目5: production current-access キャッシュ除去（uncached）

未完了 / BLOCKED / SKIP:
- D2 実DB並行3テスト（隔離Postgres未承認）
- D1 初回SQL実DB・本番付与・対象環境メール
- D3 GET/HEAD 全対象経路の実DB返却医院分離（middleware/stubのみ。対応表は台帳に記載済み・テスト未実装）
- D5 独立2プロセス失効ライブ + 性能測定（設計条件は台帳に固定・実測未）
- Linear 反映（本投稿の承認待ち）

証拠（ローカル）: todo-fix-auth.md / docs/architecture/auth.md、session verify/review artifacts（再実行せず参照）

次の操作（承認後）:
1) 使い捨てPostgresで D2 3コマンド直列
2) 合成DBで first_system_admin.sql（競合/監査rollback含む）
3) D3 real-DB isolation テスト実装単位
4) D5 2プロセス検証+測定
5) 本チケット更新 / 必要なら commit
```

### 承認対象サマリ（操作ごと）

| 操作 | 対象 | 副作用 | 必要権限 |
| --- | --- | --- | --- |
| D2 3テスト実行 | 使い捨てDB | テストデータ書込/TRUNCATE | DB作成・接続、テスト実行承認 |
| D1 合成SQL | 合成DB | admin作成/監査 | psql運用ロール、合成入力 |
| D1 本番付与 | 本番DB | 初回admin永続化 | 本番承認・秘密管理 |
| D1 メール現地 | 対象env | 外部メール | メール基盤確認承認 |
| D3 テスト実装/実行 | repo+隔離DB | 新規テスト追加・DB読 | 実装単位claim、DB |
| D5 2プロセス+測定 | 隔離スタック | プロセス起動・測定負荷 | 起動手段承認（make up禁止の代替） |
| Linear更新 | チケット UNKNOWN | 外部投稿 | 投稿承認 |
| git commit/push/merge | remote | 履歴公開 | 明示承認 |

### 台帳全体ステータス

**INCOMPLETE / BLOCKED（実環境・運用残件あり）**。prep-repair 準備ACの完了と混同しない。
