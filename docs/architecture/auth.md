# 認証・認可 設計仕様書 (Authentication & Authorization)

> **目的**: RBAC権限モデルとマルチテナント認証・認可設計を定義する。
> **読者**: 権限まわりの実装者・セキュリティレビュアー。
> **タイミング**: 認可ロジックの実装時・レビュー時。

> **Animal Ekarte**: マルチクリニック対応の堅牢なセキュリティ基盤
> **バージョン**: v9.1 | **最新更新**: 2026-08-14

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
- **Access Token (JWT)**: 15 分有効。`httpOnly` Cookie に格納。全 API リクエストの認可に使用。
- **Refresh Token (JWT)**: 最大 7 日間有効。ログイン単位の family ID と一意な JTI を持ち、ローテーション後も family の初回失効時刻を延長しません。
- **再利用検知**: ローテーション済み refresh token の再利用を検知した場合は、その token family 全体を失効させます。並行 refresh でも family の有効期間を延長しません。
- **Cookie 境界**: 現行・旧 path の refresh cookie を明示的に扱い、重複値は集約します。不正な複数 cookie やサイズ上限超過は fail closed とし、logout は検証できた family を失効させたうえで現行・旧 cookie を消去します。

### 4.2 マルチテナント分離 (X-Clinic-ID)
- ログイン時に許可された `clinic_ids` のスナップショットをトークンに封入しますが、通常リクエストの最終 authority としては使用しません。
- 原則としてリクエストごとに account、staff、clinic assignment、対象 clinic の現在状態を `backend/internal/auth/current_access_service.go` で再解決します。正常に再解決できた場合、無効化・削除・所属解除後の stale token は fail closed です。
- request-time authority lookup の一時的な取得障害も fail closed とし、middleware は 503 を返します。JWT の clinic snapshot を continuity authority に昇格しません。failure notifier は運用通知専用であり、認可結果を変更しません。
- 一般スタッフの `X-Clinic-ID` は、現在有効な所属クリニックとの一致を必須とします。
- システム管理者も任意の正数 clinic ID を選択できるわけではなく、現在存在する `is_active=true` のクリニックだけを選択できます。stale な main clinic は有効な集合から再選択し、有効な clinic がなければ拒否します。
- query/body で別の `clinic_id` / `clinic_ids` を指定できるAPIも、system adminを含めてrequest-timeのtrusted `clinic_ids` の部分集合だけを許可します。inactive clinicや集合外IDはHTTP境界で拒否します。
- clinic 切り替えは監査ログへ記録します。初回アクセスでも、指定 clinic が現在の既定 clinic と異なる場合は切り替えとして記録します。

### 4.3 資格情報変更と監査ログ

- 本人によるパスワード変更、パスワードリセット、管理者によるスタッフのパスワード再設定は、パスワード更新・reset token 失効・成功監査ログを同一 DB transaction で実行します。
- 監査ログの永続化に失敗した場合は資格情報変更も rollback し、成功レスポンスを返しません。
- 監査入力は actor、clinic、対象 staff、IP address、User-Agent のみに限定し、平文パスワード、hash、reset token、JWT、メールアドレスを記録しません。
- transaction の所有権は `backend/internal/auth/account_service.go`、`backend/internal/auth/password_reset_service.go`、`backend/internal/staff/staff_service_core.go` に置き、監査 writer は ambient transaction を必須とします。

### 4.4 実装サーフェス

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

---
