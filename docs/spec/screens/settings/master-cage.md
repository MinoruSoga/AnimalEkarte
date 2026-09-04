# ケージマスタ 仕様書 (Cage Management)

## 概要
- **画面の目的**: 院内のケージ、入院室、ペットホテルの宿泊スペースの定義。
- **URLパターン**: `/settings/cage`
- **アクセス権限**: 設備マスタ管理権限が必要（`ResourceMasterHospitalization`）

---

## 画面構成

### 1. ケージ一覧
- **表示**: ケージ名、エリア、サイズ、単価(税込)、有効/無効ステータス。
- **並び順**: 一覧内の表示順序をドラッグ&ドロップ（`dnd-kit`）で管理。ステータスは有効/無効のマスタフラグであり、入院ボードの空き/使用中といったリアルタイム稼働状況は表示しない。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **ケージ名**: 院内での識別名。
- **エリア（`cageType`）**: ICU / 犬舎(`dog`) / 猫舎(`cat`) / 汎用(`general`)。
- **サイズ（`cageSize`）**: 小型、中型、大型。
- **単価(税込)**: 入院・宿泊料金の基礎となる金額。
- **ステータス**: 有効/無効の切り替え（`StatusToggleButton`。特定の「一時利用停止」概念ではなく汎用の有効/無効フラグ）。
- **備考**: 自由記述。

---

## 主要な機能

### 1. 入院ボードとの連動
ここで定義されたケージは、入院一覧画面（`/hospitalization`）のボードビューで利用される。

### 2. 重複割り当てチェック（未実装）
`backend/internal/medicalrecord/hospitalization_service.go` の Create/Update には、同一ケージへの重複割り当てを防止するバリデーションは実装されていない。複数の入院記録に同じ `cage_id` を設定することは現状ブロックされない。

---

## 技術仕様

### 使用コンポーネント
- **`CageSettings`**: メインページ。
- **`CageSortableTable`**: `dnd-kit` によるドラッグ並び替え一覧。
- **`CageSidePanel`**: 基本属性の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/cages` | 定義済みケージの一覧取得 | `master-hospitalization` | `view` |
| GET | `/api/v1/masters/cages/:id` | 特定のケージ詳細情報の取得 | `master-hospitalization` | `view` |
| POST | `/api/v1/masters/cages` | 新規ケージの登録 | `master-hospitalization` | `create` |
| PATCH | `/api/v1/masters/cages/:id` | ケージ情報の更新 | `master-hospitalization` | `edit` |
| DELETE | `/api/v1/masters/cages/:id` | ケージの削除 | `master-hospitalization` | `delete` |
| PATCH | `/api/v1/masters/cages/reorder` | 表示レイアウトの保存 | `master-hospitalization` | `edit` |

---

