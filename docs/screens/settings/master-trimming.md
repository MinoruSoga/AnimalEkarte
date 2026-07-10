# トリミングマスタ 仕様書 (Trimming Master)

## 概要
- **画面の目的**: トリミングにおける「基本コース」と「追加オプション」の定義。
- **URLパターン**: `/settings/trimming`
- **アクセス権限**: トリミング管理権限が必要（`ResourceMasterTrimming`）
- **注意**: 品種（犬種）別の価格マトリックスは存在しない。価格の変動要因は「対象サイズ」区分（下記）のみで、犬種そのものに紐づく価格テーブルはバックエンドモデルにもない（backend/internal/model/trimming_master.go:9-51）。

---

## 画面構成

### 1. サービス区分切り替え（タブ）
- **コース (Courses)**: ベースメニュー。コース種別（[master-trimming-course-type.md](./master-trimming-course-type.md)）に紐付け可能。
- **オプション (Options)**: 追加メニュー（`is_combinable` で併用可否を設定）。

### 2. コース詳細サイドパネル
- **ステータス**、**コース種別**（指定なし可）、**対象サイズ**（小型/中型/大型/猫の4区分。品種単位ではない）、**所要時間(分)**、**単価(税込)**、**備考**。

### 3. オプション詳細サイドパネル
- **ステータス**、**所要時間(分)**、**組合せ可否**（`is_combinable`）、**単価(税込)**、**備考**。

---

## 主要な機能

### 1. 予約システムとの連動
ここで有効化されたコースとオプションは、LINE 予約 LIFF アプリおよび院内予約画面の選択肢として表示されます。

### 2. 施術記録・会計連携
トリミング記録画面で選択されたコース・オプションは、ここで設定された単価で会計明細へ引き継がれる（価格は対象サイズ区分のみに基づき、犬種別の価格分岐はない）。

---

## 技術仕様

### 使用コンポーネント
- **`TrimmingSettings`**: メインページ（コース/オプションタブ切り替え）。
- **`TrimmingCourseSidePanel` / `TrimmingOptionSidePanel`**: 詳細編集パネル。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/trimming-courses` | 基本コース一覧の取得 | `master-trimming` | `view` |
| GET | `/api/v1/masters/trimming-courses/:id` | 特定の基本コース詳細の取得 | `master-trimming` | `view` |
| POST | `/api/v1/masters/trimming-courses` | 基本コースの作成 | `master-trimming` | `create` |
| PATCH | `/api/v1/masters/trimming-courses/:id` | コース情報の更新 | `master-trimming` | `edit` |
| DELETE | `/api/v1/masters/trimming-courses/:id` | コースの削除 | `master-trimming` | `delete` |
| PATCH | `/api/v1/masters/trimming-courses/reorder` | コース表示順の一括保存（BE実装済みだが本画面からは未呼出） | `master-trimming` | `edit` |
| GET | `/api/v1/masters/trimming-options` | オプション項目の一覧取得 | `master-trimming` | `view` |
| GET | `/api/v1/masters/trimming-options/:id` | 特定のオプション項目詳細の取得 | `master-trimming` | `view` |
| POST | `/api/v1/masters/trimming-options` | オプション項目の作成 | `master-trimming` | `create` |
| PATCH | `/api/v1/masters/trimming-options/:id` | オプション項目の更新 | `master-trimming` | `edit` |
| DELETE | `/api/v1/masters/trimming-options/:id` | オプション項目の削除 | `master-trimming` | `delete` |
| PATCH | `/api/v1/masters/trimming-options/reorder` | オプション表示順の一括保存（BE実装済みだが本画面からは未呼出） | `master-trimming` | `edit` |

---

