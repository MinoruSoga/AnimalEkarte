# 動物種類マスタ 仕様書 (Animal Species Management)

## 概要
- **画面の目的**: システムで取り扱う動物種（犬、猫、うさぎ等）の定義。
- **URLパターン**: `/settings/animal-species`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterAnimalSpecies`）

---

## 1. 画面構成

### 1.1 種別一覧テーブル
- **項目**: 種別名、有効/無効ステータス。
- **ドラッグ操作**: リスト左端のハンドルで表示順序を直感的に変更（`sort_order` を更新）。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **種別名**: 臨床および統計で使用される正式名称。
- **ステータス**: 使用しなくなった動物種を「無効」に設定し、新規登録時の選択肢から除外。

---

## 主要な機能

### 1. システム全体への波及
- **品種サジェスト**: ペット編集画面で動物種名（マスタの `name`）が「犬」「猫」の場合のみ品種候補リストを提案します（ハードコードされたリテラル一致。それ以外の登録種は自由入力）。
- **医療プロトコル非連動**: ワクチンの対象種フィルタ（`Vaccine.Species`: dog/cat/both）と薬量自動計算の対象種（`MedicineDoseSpecies`: dog/cat）は、いずれも本マスタとは独立したハードコード enum です。本マスタに動物種を追加・編集しても、ワクチン接種周期や投薬量計算のロジックには反映されません。

### 2. 削除の安全性
すでにペット情報が紐付いている動物種を削除しようとした場合、データ不整合を防ぐため、エラー（409 Conflict）として物理削除がブロックされます。

---

## 技術仕様

### 使用コンポーネント
- **`AnimalSpeciesSettings`**: メインページ。
- **`dnd-kit`**: 並び替えエンジン。
- **`AnimalSpeciesSidePanel`**: `MasterSidePanel` によるボーダーレスな名称編集。

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

