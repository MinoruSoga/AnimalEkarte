# 動物種類マスタ 仕様書 (Animal Species Management)

## 概要
- **画面 of Purpose**: システムで取り扱う動物種（犬、猫、うさぎ等）の定義。
- **URLパターン**: `/settings/animal-species`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterAnimalSpecies`）

---

## 1. 画面構成

### 1.1 種別一覧テーブル
- **項目**: 種別名、アイコン、表示順、有効/無効ステータス。
- **ドラッグ操作**: リスト左端のハンドルで表示順序を直感的に変更。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **種別名**: 臨床および統計で使用される正式名称。
- **ステータス**: 使用しなくなった動物種を「無効」に設定し、新規登録時の選択肢から除外。

---

## 主要な機能

### 1. システム全体への波及
ここで定義された動物種は、以下の機能のトリガーとなります。
- **品種サジェスト**: 犬を選択した場合は犬の品種リスト、猫の場合は猫のリストを提案。
- **医療プロトコル**: ワクチン接種周期や駆虫薬の推奨（種別フィルタ）に使用。

### 2. 削除の安全性
すでにペット情報が紐付いている動物種を削除しようとした場合、データ不整合を防ぐため、エラー（409 Conflict）として物理削除がブロックされます。

---

## 技術仕様

### 使用コンポーネント
- **`AnimalSpeciesSettings`**: メインページ。
- **`dnd-kit`**: 並び替えエンジン。
- **`PropInput`**: ボーダーレスな名称編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/animal-species` | 有効な動物種一覧の取得 | `master-animal-species` | `view` |
| GET | `/api/v1/masters/animal-species/:id` | 特定の動物種情報の取得 | `master-animal-species` | `view` |
| POST | `/api/v1/masters/animal-species` | 新規動物種の登録 | `master-animal-species` | `create` |
| PATCH | `/api/v1/masters/animal-species/:id` | 属性の更新 | `master-animal-species` | `edit` |
| DELETE | `/api/v1/masters/animal-species/:id` | 動物種の削除（紐付けなしの場合のみ） | `master-animal-species` | `delete` |
| PATCH | `/api/v1/masters/animal-species/reorder` | 表示順序の一括保存 | `master-animal-species` | `edit` |

---

