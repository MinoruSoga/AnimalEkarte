# 認証・認可 設計仕様書 (Authentication & Authorization)

> **目的**: RBAC権限モデルとマルチテナント認証・認可設計を定義する。
> **読者**: 権限まわりの実装者・セキュリティレビュアー。
> **タイミング**: 認可ロジックの実装時・レビュー時。

> **Animal Ekarte**: マルチクリニック対応の堅牢なセキュリティ基盤
> **バージョン**: v9.2 | **最新更新**: 2026-09-09

---

## 1. ユーザー・権限モデル

本システムの実装上の authority mode は 2 つです。

### 1.1 Authority modes
- **システム管理者 (System Admin)**: `Account.is_system_admin` で表す。全クリニックを無条件に信用せず、現在存在する active clinic に限定する。
- **Clinic-scoped staff**: staff と clinic assignment を request-time に再解決し、権限グループの grant で認可する。

「クリニック管理者」は独立した user type / flag ではなく、院内で full access を与えるよう設定した permission-group profile である。

### 1.2 リソースベース認可 (RBAC)
システム内の **37 種類のリソース** に対し、`View (閲覧)`, `Create (作成)`, `Edit (編集)`, `Delete (削除)` の 4 アクション単位でアクセスを制御します。

---

## 2. 実効権限の計算ロジック

スタッフの実権限は、所属する複数の権限グループの「和集合 (OR)」として動的に計算されます。

1.  **グループ所属**: スタッフは 0 個以上の `permission_groups` に紐付けられます。該当 grant が 1 つもない場合は deny-by-default です。
2.  **ルール統合**: 各グループが持つ `permission_group_rules` を収集。
3.  **パーミッション・マップ**: 同一リソースに対して複数のルールがある場合、いずれかのグループで許可されていれば「許可」と判定します。実効権限の集計は `backend/internal/auth/permission_group_repository.go` の `FindAllEffectivePermissionsByStaffID`、HTTP 境界での強制は `backend/internal/auth/http_permission.go` の `RequirePermission` / `RequirePermissionAny` が担当します。
4.  **Admin 特例**: `is_system_admin` フラグが true の場合、リソース・アクションの計算をバイパスします。ただし、アクセス対象の clinic scope はバイパスせず、現在も存在する有効なクリニックに限定します。

### 2.1 権限グループ変更の自己ロックアウト (D2)

非システム管理者が自分の所属グループを変更・無効化・自己割当解除するとき、**変更後の選択医院における実効権限**で `master-permission:view` と `master-permission:edit` の両方を維持できなければ 403 で拒否する。別の有効グループからの OR 付与が残る場合は許可する。他の管理担当者が残ることだけでは例外にしない。

- 最終判定は mutation 後の `guardActorKeepsPermissionAdministration`（`permission_group_service_self_lockout.go`）。`UpdateRules` と `Update`（Rules 付き / Update-with-Rules）の両方は、所属グループ単体の事前拒否（旧 `validateNotSelfReference`）に依存しない。別の有効グループからの OR 付与が残れば許可し、最終の view/edit 喪失・lookup/audit 失敗は拒否して rollback する。
- lookup 失敗・監査書き込み失敗は成功扱いせず、ルール変更と成功監査を rollback する。service 単体では staged/committed TX double（`permission_group_service_rules_d2_test.go` / `permission_group_service_mutate_d2_test.go`）で同一 `WithTx` 参加と失敗時の committed 不変を観測する（実 DB 原子性の証明ではない）。
- システム管理者はこの自己喪失判定を免除する。システム管理者アカウント自体の削除・無効化保護は別契約。
- 医院単位の permission policy lock（`pg_advisory_xact_lock`）で並行変更を直列化する。並行 DB テストコードは holder/contender の `pg_locks` 待機関係と worker 終了を観測するが、実行は承認済み隔離 Postgres が必要。

---

## 3. 全リソース・キー一覧 (Verified)

実装コード (`backend/internal/model/permission.go` の `AllResources`) に定義されている全 37 リソースキーです。

| カテゴリ | リソースキー | 管理対象 |
|:---|:---|:---|
| **臨床コア** | `reception`, `owners`, `reservations`, `medical-records`, `hospitalization`, `trimming`, `examinations`, `examination-unconfirm`, `vaccinations`, `checkups`, `checkup-package-import`, `lab-import` | 受付、飼主、予約、カルテ、入院、トリミング、検査、検査確定解除、ワクチン、健診、健診パッケージ取込、外部検査結果インポート。 |
| **会計・経営** | `accounting`, `accounting-cancel`, `accounting-post-close-edit`, `cash-register-close`, `accounting-reports`, `discount`, `closing-settings`, `master-payment-method` | 会計、会計キャンセル、締め後編集、レジ締め、売上レポート、値引操作、締め時間設定、支払方法。 |
| **物流・管理** | `inventory`, `estimates`, `shifts`, `hospital-settings` | 在庫、見積書、シフト、医院基本設定。 |
| **マスタ設定** | `master-animal-species`, `master-medical`, `master-reservation-type`, `master-hospitalization`, `master-trimming`, `master-permission`, `master-staff`, `master-insurance`, `master-merchandise` | 各種定義データの管理。 |
| **外部連携** | `lstep-analytics`, `lstep-csv-import` | CRM 分析、CSV インポート履歴。 |
| **横断（医院間リンク）** | `identity-links` | 医院別 owner/pet の明示リンク（view/edit 分離・fail-closed 既定）。 |
| **その他** | `manual-edit` | 取扱説明書の編集権限。 |

---

## 4. セッションとセキュリティ

### 4.1 dual-token 方式
- **Access Token (JWT)**: 15分有効。`httpOnly` Cookieに格納し、保護されたstaff向け`/api/v1` routeの認可に使用する。login/refresh/password-reset等のpublic auth routeと、LIFF専用routeは各route固有の認証・rate limitを使う。
- **Refresh Token (JWT)**: 最大 7 日間有効。ログイン単位の family ID と一意な JTI を持ち、ローテーション後も family の初回失効時刻を延長しません。
- **再利用検知**: ローテーション済み refresh token の再利用を検知した場合は、その token family 全体を失効させます。並行 refresh でも family の有効期間を延長しません。
- **Cookie 境界**: 現行・旧 path の refresh cookie を明示的に扱い、重複値は集約します。不正な複数 cookie やサイズ上限超過は fail closed とし、logout は検証できた family を失効させたうえで現行・旧 cookie を消去します。

### 4.2 マルチテナント分離 (X-Clinic-ID)
- ログイン時に許可された `clinic_ids` のスナップショットをトークンに封入しますが、通常リクエストの最終 authority としては使用しません。
- 原則としてリクエストごとに account、staff、clinic assignment、対象 clinic の現在状態を `backend/internal/auth/current_access_service.go` で再解決します。production composition は `NewCachedCurrentAccessResolver` を挟みません。スタッフ無効化・所属解除・パスワード変更の **DB commit 後に受け付ける次のリクエスト** から拒否します。commit 前に受け付けた処理の取消しは保証しません。lookup 障害は 503 です。
- 同じ DB を使う独立 2 プロセスでのライブ検証は、この変更の範囲では実施していません（`make up` 禁止）。`current_access_cache.go` は残置しますが、最終認可の入力には使いません。
- request-time authority lookup の一時的な取得障害も fail closed とし、middleware は 503 を返します。JWT の clinic snapshot を continuity authority に昇格しません。failure notifier は運用通知専用であり、認可結果を変更しません。
- 一般スタッフの `X-Clinic-ID` は、現在有効な所属クリニックとの一致を必須とします。
- システム管理者も任意の正数 clinic ID を選択できるわけではなく、現在存在する `is_active=true` のクリニックだけを選択できます。stale な main clinic は有効な集合から再選択し、有効な clinic がなければ拒否します。
- query/body で別の `clinic_id` / `clinic_ids` を指定できるAPIも、system adminを含めてrequest-timeのtrusted `clinic_ids` の部分集合だけを許可します。inactive clinicや集合外IDはHTTP境界で拒否します。
- clinic 切り替えは監査ログへ記録します。初回アクセスでも、指定 clinic が現在の既定 clinic と異なる場合は切り替えとして記録します。

### 4.3 資格情報変更と監査ログ

- 管理者によるスタッフのパスワード再設定は、対象アカウントを同一 transaction で `FOR UPDATE` し、対象がシステム管理者なら操作者にもシステム管理者を要求します。拒否時は password hash・reset token・成功監査を変更しません。
- 既存スタッフへのログインアカウント追加は `POST /api/v1/masters/staffs/{id}/account`（システム管理者のみ）。staff ID は維持し、初期秘密値の平文は返しません。本人はパスワード再設定から設定します。メール自動送信はしません。
- 監査ログの永続化に失敗した場合は資格情報変更も rollback し、成功レスポンスを返しません。
- 監査入力は actor、clinic、対象 staff、IP address、User-Agent のみに限定し、平文パスワード、hash、reset token、JWT、メールアドレスを記録しません。
- transaction の所有権は `backend/internal/auth/account_service.go`、`backend/internal/auth/password_reset_service.go`、`backend/internal/staff/staff_service_core.go` に置き、監査 writer は ambient transaction を必須とします。

### 4.4 非本番デモログインの例外

`auth_service.go` は通常の password hash 照合に加え、`seedlogin.AcceptSharedPassword` による合成デモ catalog 限定の認証補助を持つ。`APP_ENV` の許可集合は `development` / `local` / `dev` / `test` / `staging`（trim/lowercase 後）で、production・未設定・未知値および catalog 外アカウントは対象外である。認証補助が成立しても account の active/deleted 判定と後続の staff/clinic authority 解決を通る。資格情報の値は本書へ複製しない。

適用責務と境界は [exception-package-discipline.md](exception-package-discipline.md#a8-7--seedlogin-is-an-explicit-non-production-exception) を参照する。環境設定の実測や本番受入の記録ではない。

### 4.5 Cookie認証とCSRF対策

Cookie認証を使う保護routeとlogin/refresh/logoutには `RequireXRequestedWith` を適用する。適用routeのGET/HEAD/OPTIONSを除くリクエストは、非空の `X-Requested-With` ヘッダーがなければ403となる。publicな `/auth/forgot-password`・`/auth/reset-password` はこのmiddlewareの対象外で、専用のrate limitを使う。Frontendの共通axiosは通常 `XMLHttpRequest` を付けるが、testモードでは省略する。手動HTTP検証でも必要なヘッダーを付け、403をすべて権限不足と判断しない。

別originのブラウザアクセスはCORSによるorigin・credentials・送信ヘッダーの許可も必要。OPTIONSの成功ステータスだけでは確認できない。会計確定の `Idempotency-Key` を含む確認手順は [Vercel STG検証](../ops/deploy/VERCEL-FRONTEND-STAGING-TEST.md) を参照する。

### 4.6 実装サーフェス

| 責務 | 実装 |
|:---|:---|
| JWT 発行・検証・refresh family | `backend/internal/auth/token_service.go` |
| login / refresh / logout HTTP 境界 | `backend/internal/auth/http_session.go` |
| password HTTP 境界 | `backend/internal/auth/http_password.go` |
| request-time authority | `backend/internal/auth/current_access_service.go` |
| 認証 middleware と clinic 切り替え | `backend/internal/middleware/auth.go` |
| RBAC repository / use case | `backend/internal/auth/permission_group_repository.go`, `permission_group_service.go` |
| 資格情報 transaction | `backend/internal/auth/account_service.go`, `password_reset_service.go`, `backend/internal/staff/staff_service_core.go` |
| production composition | `backend/cmd/api/composition_auth.go`, `composition_staff_account.go` |

### 4.7 GET/HEAD の選択医院 grant

`RequirePermission` / `RequirePermissionAny` の既定は **選択医院の grant のみ** です。GET/HEAD でも所属する他院の grant では通りません。横断一覧・詳細は `RequirePermissionAllowingAssignedClinicGrant`（または Any 版）を composition で明示し、handler が宛先医院ごとに Filter/Authorize します。同じ composition の医院固定 handler は `RequireSelectedClinicGrant` または `extractSelectedClinicGrant` で選択医院を再確認します。

全件の静的対応表は [`get_head_permissions.json`](../../backend/cmd/api/testdata/get_head_permissions.json) です。各行に method/path、handler、resource/action、登録式、認可位置、返却範囲、回帰テスト参照、検証の限界を記録します。[分類テスト](../../backend/cmd/api/get_head_permission_classification_test.go) は登録集合と台帳を双方向照合し、既知 prefix 内の追加も含め、追加・削除・method/handler 変更を拒否します。登録式の変更と参照先の欠落も検出します。local storage は 203 件、S3 設定では uploads の GET/HEAD を除く 201 件です。

| 分類 | 認可 | 代表経路 |
|:---|:---|:---|
| public | 認証なし | health、LIFF settings、local uploads |
| liff | LIFF 認証 + path clinic/customer | LIFF profile、予約一覧、予約可能枠など（settings を除く） |
| self | 認証済み本人 | `/me` の本人情報・所属医院 |
| cross-clinic | 宛先医院ごとの grant または明示した所属/admin 判定 | owners / pets / reservations / medical-records / accountings の横断一覧・詳細、identity-links、医院一覧、path clinic の LINE 予約設定 |
| clinic-fixed | 選択医院 grant + 選択医院の query scope | 診療子リソース、検査、入院、見積、未収、医院別マスタ、staffs、shifts、inventory、trimming、LSTEP 医院データ |
| shared-master | 明示した grant + 医院非依存データ | animal-species、manual articles、company、LSTEP tag config |

login、password reset、LINE webhook、scheduled jobs は GET/HEAD ではないため、この表の対象外です。HEAD 未登録は追加しません。正常に 0 行の一覧と、認可可能な医院が 0 件の 403 は別です。

この gate が証明するのは **登録済み経路の静的対応表が欠落・陳腐化していないこと** です。参照テストの存在は実行 PASS を意味せず、登録式の照合だけでは middleware の実行、間接 helper の変更、実 DB の返却範囲は証明できません。実 DB の返却分離は台帳 `verification` の `realdb-return-data:`（[get_head_permission_realdb_coverage_test.go](../../backend/cmd/api/get_head_permission_realdb_coverage_test.go)）と disposable 実行証跡で別途証明します。静的分類だけを D3 完了の根拠にはしません。進捗の正本は [todo-fix-auth.md](../../todo-fix-auth.md) です。

### 4.8 `/me` と login の医院フィールド

`main_clinic_id` と `clinics[].is_main` は **永続化された主所属ではなく、この応答時点の選択医院（login/refresh では解決した既定医院）** です。DB の `staff_clinic_assignments.is_main` とは別です。API 名の変更はこの文書訂正だけでは行いません。

Frontend の `/me` は `staleTime` 5 分・window focus 再取得なし・定期ポーリングなしです。5 分経過だけでは自動更新しません。再取得はログイン、token refresh、`refreshPermissions`、`ME_QUERY_KEY` の無効化、再マウントです。UI の古い権限表示と BE の最終認可は別です。自動ポーリングは追加しません。

### 4.9 医院選択失効

選択医院または既定医院が利用できないとき、BE は通常の 403 と区別して `error_code: clinic_selection_unavailable` を返します。Frontend は書き込みを停止し、`X-Clinic-ID` なしの `/me` を 1 回取得して active な既定医院へ戻します。失敗した POST/PUT/PATCH/DELETE は新しい医院へ自動再送しません。

ヘッダーなしの復旧 `/me` も同コードの403なら有効医院が0件のため、管理者への確認とログアウトを案内します。401・通常403・通信障害は「医院なし」と混同しません。自動復旧終了後は後続GETやポーリングで再開せず、復旧障害時の再試行ボタンからだけ実行します。ログアウト開始時に進行中の復旧を中止し、遅延応答が医院保存やreloadを起こすことを防ぎます。

### 4.10 医院一覧 `scope=all`

非システム管理者が `scope=all` を指定しても、所属医院だけを返します。所属判定はサーバ側の staff assignment です。システム管理者の全院一覧は維持します。

---

## 5. Issue 仕様との読み分け

[GitHub #1](https://github.com/MinoruSoga/AnimalEkarte/issues/1) は CLOSED だが、本文の `UserRole` / `users` / `user_clinics` は当初案である。現行の認可契約は本書の `Account`・`Staff`・`staff_clinic_assignments` と permission-group grant、`backend/internal/model/permission.go` / `auth` / `middleware` を参照する。Issue の固定職種ロールを現在の authority mode として再導入しない。
