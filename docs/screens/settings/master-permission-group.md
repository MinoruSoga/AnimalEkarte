# 権限グループマスタ設定 仕様書

## 概要
- **画面の目的**: 職種や役職に応じたシステム操作権限（閲覧・作成・編集・削除）をグループ単位で定義する。マルチテナント（`clinic_id`）ごとに独立した権限セットを持つ。
- **URLパターン**: `/settings/permission-groups`
- **コンポーネント**: `PermissionGroupSettings`
- **アクセス権限**: `ResourceMasterStaff` (edit以上) 権限が必要

## 画面構成
- **メインエリア**: `DataTable` によるグループ一覧。
  - カラム: 名称（カラードット付）、説明、ステータス、操作。
- **サイドパネル**: `PermissionGroupSidePanel` による詳細編集。
  - 基本情報（名称、色、説明）
  - **権限マトリクス**: 各リソース（飼主、カルテ、入院、会計、マスタ等）に対するアクション（view, create, edit, delete）の許可/拒否をチェックボックスで設定。

## 権限設定対象リソース
- **Owners**: 飼主・ペット情報
- **MedicalRecords**: 電子カルテ
- **Hospitalization**: 入院管理
- **Accounting**: 会計・精算
- **Estimates**: 見積書
- **Inventory**: 在庫管理
- **Shifts**: シフト管理
- **MasterMedical**: 診療項目・診断等の医療系マスタ
- **MasterTrimming**: トリミングマスタ
- **MasterHospitalization**: 入院・ケージマスタ
- **MasterMerchandise**: 商品・保険マスタ
- **MasterStaff**: スタッフ・職種・権限マスタ
- **HospitalSettings**: 医院基本情報
- **CashRegisterClose** (`cash-register-close`): レジ締め実行・締め履歴閲覧（FEAT-368）
- **AccountingReports** (`accounting-reports`): 月次売上集計・CSV エクスポート（FEAT-368）
- **ClosingSettings** (`closing-settings`): 締め時間・特別期間の設定変更（FEAT-368）

## 主要機能
- **カラーコーディング**: グループごとに色を設定でき、スタッフ一覧等での識別性を高める。
- **きめ細やかなアクセス制御**: リソースごとに「閲覧はできるが編集はできない」といった詳細な設定が可能。
- **即時反映**: 権限設定を変更して保存すると、該当グループに所属するスタッフの次回操作から新しい権限が適用される（フロントエンドでは `usePermission` を通じて UI を動的に制御）。

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/permission-groups` | グループ一覧取得 | 実装済 |
| POST | `/api/v1/permission-groups` | 新規作成 | 実装済 |
| PATCH | `/api/v1/permission-groups/:id` | 更新 | 実装済 |
| DELETE | `/api/v1/permission-groups/:id` | 削除 | 実装済 |
