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
API 側は次のリクエストで実効権限を再評価します。画面のメニュー・ボタンは `/v1/me` のキャッシュに従います。`useUpdatePermissionGroup` は更新した端末の `ME_QUERY_KEY` を無効化しますが、別端末へはプッシュしません。`get-me.ts` は定期ポーリング・フォーカス時再取得を無効にしており、起動・ログイン・トークン更新、または明示的な権限再取得で反映します。

### 2. デフォルトグループの提供
新規院開設時に「執行」と「一般」の 2 グループが、**デフォルトの権限ルール付きで**自動作成されます（`backend/internal/clinic/clinic_service.go` の `defaultPermissionRuleTable`）。執行でも全リソース・全操作が許可されるわけではありません。例えば医院設定の作成・削除、動物種マスタの変更、検査確定解除は既定で許可されず、identity-links も既定付与しません。一般は主に閲覧と一部臨床業務の作成・編集を持ちます。正確な resource/action は同テーブルを参照し、運用上必要な差分を本画面で設定します。

### 3. 複数グループの権限合算

スタッフに複数の権限グループを割り当てると、いずれかのグループが許可する操作を実行できる（OR）。一つのグループのチェックを外しても、別グループの許可は取り消されない。権限を縮小する場合は割当グループ全体を確認する。どのグループにも許可がなければ拒否され、職種名だけでは権限を付与しない。詳細は[認証・認可仕様](../../../architecture/auth.md#2-実効権限の計算ロジック)。

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
