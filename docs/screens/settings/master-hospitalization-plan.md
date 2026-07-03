# 入院・宿泊プランマスタ 仕様書 (Hospitalization Plans)

## 概要
- **画面 of Purpose**: 入院治療やペットホテルの基本料金体系、および初期ケアセットの管理。
- **URLパターン**: `/settings/hospitalization`
- **アクセス権限**: 設備マスタ管理権限が必要（`ResourceMasterHospitalization`）

---

## 1. 画面構成

### 1.1 プラン一覧
- **表示項目**: プラン名、1 日単価、税率、デフォルトの給餌回数。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **プラン名称**: 例：「ICU 管理（24h）」「一般入院」「猫専用スイート」。
- **標準価格**: 1 日（1 泊）あたりの税込単価。
- **初期ケアセット**:
    - 入院登録時に自動でケアプラン（Tab 2）に展開されるタスク（例：朝夕のバイタル測定、食事提供等）を定義。

---

## 主要な機能

### 1. 会計への自動集計
退院処理実行時、入院日数（または泊数）とこのマスタの単価を掛け合わせ、自動的に会計明細が作成されます。

### 2. スタッフ業務の標準化
プランごとにデフォルトのケアタスクを設定しておくことで、入院登録時の指示漏れを防止し、看護業務の質を均一化します。

---

## 技術仕様

### 3.1 構成コンポーネント
- **`HospitalizationSettings`**: メインページ。
- **`TaskTemplateList`**: プランに紐づく初期タスクの動的編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/hospitalization-plans` | 有効なプラン一覧の取得 | `master-hospitalization` | `view` |
| GET | `/api/v1/masters/hospitalization-plans/:id` | 特定の入院プラン詳細の取得 | `master-hospitalization` | `view` |
| POST | `/api/v1/masters/hospitalization-plans` | 新規入院プランの登録 | `master-hospitalization` | `create` |
| PATCH | `/api/v1/masters/hospitalization-plans/:id` | 価格や初期タスクの更新 | `master-hospitalization` | `edit` |
| DELETE | `/api/v1/masters/hospitalization-plans/:id` | 入院プランの削除 | `master-hospitalization` | `delete` |
| PATCH | `/api/v1/masters/hospitalization-plans/reorder` | 表示順序の一括保存 | `master-hospitalization` | `edit` |

---

