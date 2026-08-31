# 権限グループ設定 仕様書 (Permission Groups)

## 概要
- **画面の目的**: スタッフの役割（院長、獣医師、看護師、事務等）に応じたシステム操作権限の体系的定義。
- **URLパターン**: `/settings/permission-groups`
- **アクセス権限**: 権限グループ管理権限が必要（`ResourceMasterPermission`）

---

## 画面構成

### 1. グループ一覧
「獣医師（フルアクセス）」「受付（会計のみ）」「研修生（閲覧のみ）」などの定義済みリスト。

### 2. 権限マトリックス編集
各リソースに対し、以下の 4 段階のアクセスレベルをチェックボックス形式で設定します。

| アクション | 説明 |
|:---|:---|
| **View (表示)** | データの参照。リスト表示。 |
| **Create (作成)** | 新規レコードの追加。 |
| **Edit (編集)** | 既存データの修正、ステータスの更新。 |
| **Delete (削除)** | 論理削除、取り消し操作。 |

---

## 管理リソース（主要カテゴリ）

- **医療**: `medical-records`, `examinations`, `vaccinations`, `hospitalization`
- **フロント**: `reception`, `owners`, `reservations`
- **経営**: `accounting`, `cash-register-close`, `accounting-reports`
- **マスタ**: `master-staff`, `master-medical`, `hospital-settings`

---

## 主要な機能

### 1. 即時反映とキャッシュ
権限グループの設定変更は、該当グループに所属する全スタッフに即時反映されます。ログイン中のユーザーに対しては、次回のページ遷移または API リクエスト時に権限の再評価が行われます。

### 2. デフォルトグループの提供
新規院開設時に「執行」と「一般」の 2 グループが、**デフォルトの権限ルール付きで**自動作成されます（`clinic_service.go` の `defaultPermissionRuleTable`。SD-9 対応・2026-07-17）。ルールの内容は`clinic_service.go`の現行default tableを正本とし、おおむね「執行 = 全業務リソースのフル権限（削除含む）／一般 = 閲覧＋主要業務の作成・編集（削除・会計確定・設定系は不可）」。開設直後から system_admin 以外もログイン運用でき、細部は本画面で調整します。

---

## 技術仕様

### 使用コンポーネント
- **`PermissionGroupSettings`**: メインページ（`MasterCRUDPage` ベース、ドラッグによる並び替え対応）。
- **`PermissionGroupSidePanel`**: グループ名・カラー・権限ルールの編集パネル（`MasterSidePanel` ベース）。
- **`PermissionRuleTable`**: リソース × アクションのチェックボックス格子入力コンポーネント。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/permission-groups` | 権限定義の一覧取得 | `master-permission` | `view` |
| GET | `/api/v1/masters/permission-groups/:id` | 特定の権限グループ詳細の取得 | `master-permission` | `view` |
| POST | `/api/v1/masters/permission-groups` | 新規権限グループの作成 | `master-permission` | `create` |
| PATCH | `/api/v1/masters/permission-groups/:id` | グループ名・説明・カラー等メタデータの更新 | `master-permission` | `edit` |
| DELETE | `/api/v1/masters/permission-groups/:id` | 権限グループの削除 | `master-permission` | `delete` |
| PATCH | `/api/v1/masters/permission-groups/reorder` | 表示順序の一括保存 | `master-permission` | `edit` |
| PUT | `/api/v1/masters/permission-groups/:id/rules` | 権限マトリックス（ルール）の更新 | `master-permission` | `edit` |

---

