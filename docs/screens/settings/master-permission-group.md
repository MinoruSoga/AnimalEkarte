# 権限グループ設定 仕様書 (Permission Groups)

## 概要
- **画面の目的**: スタッフの役割（院長、獣医師、看護師、事務等）に応じたシステム操作権限の体系的定義。
- **URLパターン**: `/settings/permission-groups`
- **アクセス権限**: 最高管理者権限のみ（`ResourceMasterPermission`）

---

## 画面構成

### 1. グループ一覧
「獣医師（フルアクセス）」「受付（会計のみ）」「研修生（閲覧のみ）」などの定義済みリスト。

### 2. 権限マトリックス編集
各リソースに対し、以下の 4 段階のアクセスレベルをチェックボックス形式で設定します。

| アクション | 説明 |
|:---|:---|
| **View (閲覧)** | データの参照。リスト表示。 |
| **Create (作成)** | 新規レコードの追加。 |
| **Edit (編集)** | 既存データの修正、ステータスの更新。 |
| **Delete (削除)** | 論理削除、取り消し操作。 |

---

## 管理リソース（主要カテゴリ）

- **医療**: `medical-records`, `exams`, `vaccinations`, `hospitalization`
- **フロント**: `reception`, `owners`, `reservations`
- **経営**: `accounting`, `cash-register-close`, `accounting-reports`
- **マーケ**: `lstep-analytics`
- **マスタ**: `master-staff`, `master-medical`, `hospital-settings`

---

## 主要な機能

### 1. 即時反映とキャッシュ
権限グループの設定変更は、該当グループに所属する全スタッフに即時反映されます。ログイン中のユーザーに対しては、次回のページ遷移または API リクエスト時に権限の再評価が行われます。

### 2. デフォルトグループの提供
新規院開設時など、標準的な「獣医師」「看護師」用の権限セットがテンプレートとして提供されます。

---

## 技術仕様

### 使用コンポーネント
- **`PermissionMatrixTable`**: リソース × アクションの格子状入力コンポーネント。
- **`GroupAssignmentInfo`**: 当該グループに現在所属しているスタッフの一覧表示。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/masters/permission-groups` | 権限定義の一覧取得。 |
| PATCH | `/api/v1/masters/permission-groups/:id` | 権限マトリックスの更新。 |

---
