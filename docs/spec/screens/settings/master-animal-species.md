# 動物種類マスタ 仕様書 (Animal Species Management)

## 概要
- **画面の目的**: システムで取り扱う動物種（犬、猫、うさぎ等）の定義。
- **URLパターン**: `/settings/animal-species`
- **アクセス権限**:
  - **一覧・参照**: `master-animal-species` の `view`（`ResourceMasterAnimalSpecies`）。
  - **作成・更新・削除・並び替え**: システム管理者のみ（`is_system_admin` / backend `requireSystemAdminForGlobalMaster`）。全クリニック共有マスタのため、clinic-scoped の resource create/edit/delete では mutation 不可。FE も `canMutate = isSystemAdmin`。

---

## 1. 画面構成

### 1.1 種別一覧テーブル
- **項目**: 種別名、有効/無効ステータス。
- **ドラッグ操作**: リスト左端のハンドルで表示順序を直感的に変更（`sort_order` を更新）。実行はシステム管理者のみ。

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

### 3. 監査（fail-closed）
Create / Update / Delete / Reorder は業務 write と `audit_logs` を同一 transaction で commit する。監査失敗時は業務変更も rollback する（`TestAnimalSpeciesService_*_AuditFailureRollsBack`）。

---

## 技術仕様

### 使用コンポーネント
- **`AnimalSpeciesSettings`**: メインページ。
- **`dnd-kit`**: 並び替えエンジン。
- **`AnimalSpeciesSidePanel`**: `MasterSidePanel` によるボーダーレスな名称編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/animal-species` | 動物種一覧の取得（無効を含む全件） | `master-animal-species` | `view` |
| GET | `/api/v1/masters/animal-species/:id` | 特定の動物種情報の取得 | `master-animal-species` | `view` |
| POST | `/api/v1/masters/animal-species` | 新規動物種の登録 | システム管理者 (`is_system_admin`) | — |
| PATCH | `/api/v1/masters/animal-species/:id` | 属性の更新 | システム管理者 (`is_system_admin`) | — |
| DELETE | `/api/v1/masters/animal-species/:id` | 動物種の削除（紐付けなしの場合のみ） | システム管理者 (`is_system_admin`) | — |
| PATCH | `/api/v1/masters/animal-species/reorder` | 表示順序の一括保存 | システム管理者 (`is_system_admin`) | — |

---
