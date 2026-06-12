# トリミングコース種別マスタ 仕様書 (Trimming Course Type)

## 概要
- **画面の目的**: トリミングコースの大分類（シャンプー、シャンプー+カットなど）を管理し、トリミングコース管理画面の分類基盤を整える。
- **URLパターン**: `/settings/trimming-course-type`
- **アクセス権限**: スタッフ管理者以上（`ResourceMasterTrimming`）

---

## 画面構成

### 1. 一覧
- **表示項目**: 種別名、有効/無効ステータス、操作ボタン。
- **編集方式**: SidePeek（右側パネル）で名称と有効状態を更新。

### 2. 新規登録／更新
- **名称**: 必須。空文字は保存不可。
- **有効状態**: トグルで `isActive` を変更。

---

## 技術仕様

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/masters/trimming-course-types` | 定義済みのコース種別を取得。 |
| POST | `/api/v1/masters/trimming-course-types` | コース種別を追加。 |
| PATCH | `/api/v1/masters/trimming-course-types/:id` | コース種別情報を更新。 |
| DELETE | `/api/v1/masters/trimming-course-types/:id` | コース種別を削除。 |
| PATCH | `/api/v1/masters/trimming-course-types/reorder` | 種別の表示順を一括保存。 |

### 関連画面
- トリミングマスタ（`/settings/trimming`）で表示するコース群の上位分類として参照されます。

