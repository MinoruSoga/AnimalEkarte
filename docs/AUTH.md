# 認証・認可 設計仕様書 (Authentication & Authorization)

> **Animal Ekarte**: マルチクリニック対応の堅牢なセキュリティ基盤
> **バージョン**: v8.6 | **最新更新**: 2026-06-12

---

## 1. ユーザー・権限モデル

本システムは、医療データの機密性と拠点運営の柔軟性を両立するため、3 層の認可構造を採用しています。

### 1.1 ユーザー種別 (User Types)
- **システム管理者 (System Admin)**: 全クリニックの横断管理、インフラ設定権限。
- **クリニック管理者 (Clinic Admin)**: 所属院内の全リソースに対するフルアクセス。
- **一般スタッフ (Staff)**: 業務ロール（獣医師、看護師、受付等）に応じた権限グループに所属。

### 1.2 リソースベース認可 (RBAC)
システム内の **31 種類のリソース** に対し、`View (閲覧)`, `Create (作成)`, `Edit (編集)`, `Delete (削除)` の 4 アクション単位でアクセスを制御します。

---

## 2. 実効権限の計算ロジック

スタッフの実権限は、所属する複数の権限グループの「和集合 (OR)」として動的に計算されます。

1.  **グループ所属**: スタッフは 1 つ以上の `permission_groups` に紐付けられます。
2.  **ルール統合**: 各グループが持つ `permission_group_rules` を収集。
3.  **パーミッション・マップ**: 同一リソースに対して複数のルールがある場合、いずれかのグループで許可されていれば「許可」と判定（`middleware/auth.go`）。
4.  **Admin 特例**: `is_system_admin` フラグが true の場合、全ての計算をバイパスし、全リソースに対して全アクションが許可されます。

---

## 3. 全リソース・キー一覧 (Verified)

実装コード (`backend/internal/model/permission.go`) に定義されている全 31 リソースキーです。

| カテゴリ | リソースキー | 管理対象 |
|:---|:---|:---|
| **臨床コア** | `reception`, `owners`, `reservations`, `medical-records`, `hospitalization`, `trimming`, `examinations`, `vaccinations`, `checkups` | 受付、飼主、予約、カルテ、入院、トリミング、検査、ワクチン、健診。 |
| **会計・経営** | `accounting`, `cash-register-close`, `accounting-reports`, `discount`, `closing-settings`, `master-payment-method` | 会計、レジ締め、売上レポート、値引操作、締め時間設定、支払方法。 |
| **物流・管理** | `inventory`, `estimates`, `shifts`, `hospital-settings` | 在庫、見積書、シフト、医院基本設定。 |
| **マスタ設定** | `master-animal-species`, `master-medical`, `master-reservation-type`, `master-hospitalization`, `master-trimming`, `master-permission`, `master-staff`, `master-insurance`, `master-merchandise` | 各種定義データの管理。 |
| **外部連携** | `lstep-analytics`, `lstep-csv-import` | CRM 分析、CSV インポート履歴。 |
| **その他** | `manual-edit` | 取扱説明書の編集権限。 |

---

## 4. セッションとセキュリティ

### 4.1 dual-token 方式
- **Access Token (JWT)**: 15 分有効。`httpOnly` Cookie に格納。全 API リクエストの認可に使用。
- **Refresh Token (JWT)**: 7 日間有効。トークンローテーション方式を採用。

### 4.2 マルチテナント分離 (X-Clinic-ID)
- ログイン時に許可された `clinic_ids` のリストをトークンに封入。
- リクエストヘッダー `X-Clinic-ID` による医院切り替え時、トークン内の許可リストと照合し、他院データへの越境を物理的に遮断します。

---
