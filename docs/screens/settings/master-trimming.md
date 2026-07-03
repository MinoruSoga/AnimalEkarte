# トリミングマスタ 仕様書 (Trimming Master)

## 概要
- **画面の目的**: トリミングにおける「基本コース」と「追加オプション」の定義、および犬種別の料金設定。
- **URLパターン**: `/settings/trimming`
- **アクセス権限**: トリミング管理権限が必要（`ResourceMasterTrimming`）

---

## 画面構成

### 1. サービス区分切り替え
- **コース (Courses)**: シャンプー、シャンプー＆カット等のベースメニュー。
- **オプション (Options)**: 薬用シャンプー、泥パック、歯磨き等の追加メニュー。

### 2. コース別・犬種別料金表
コースを選択すると、動物種（犬/猫）および品種（トイプードル/チワワ等）ごとの価格一覧が表示されます。
- **一括編集**: 複数の品種に対して共通の価格を適用。
- **個別設定**: 特定の品種のみ「毛量が多い」等の理由で価格を調整。

---

## 主要な機能

### 1. 予約システムとの連動
ここで有効化されたコースとオプションは、LINE 予約 LIFF アプリおよび院内予約画面の選択肢として表示されます。

### 2. 施術記録・会計連携
トリミング記録画面 (`/trimming/:id`) で選択されたサービスは、その犬種に応じた価格で自動的に会計明細（レジ）へ引き継がれます。

---

## 技術仕様

### 使用コンポーネント
- **`TrimmingPriceMatrix`**: コース × 品種の価格マトリックス編集部品（※コースマスタ更新による価格反映を想定）。
- **`MasterSelectModal`**: 対象となる品種の検索・選択。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/trimming-courses` | 基本コース一覧の取得 | `master-trimming` | `view` |
| GET | `/api/v1/masters/trimming-courses/:id` | 特定の基本コース詳細の取得 | `master-trimming` | `view` |
| POST | `/api/v1/masters/trimming-courses` | 基本コースの作成 | `master-trimming` | `create` |
| PATCH | `/api/v1/masters/trimming-courses/:id` | コース情報の更新 | `master-trimming` | `edit` |
| DELETE | `/api/v1/masters/trimming-courses/:id` | コースの削除 | `master-trimming` | `delete` |
| PATCH | `/api/v1/masters/trimming-courses/reorder` | コース表示順の一括保存 | `master-trimming` | `edit` |
| GET | `/api/v1/masters/trimming-options` | オプション項目の一覧取得 | `master-trimming` | `view` |
| GET | `/api/v1/masters/trimming-options/:id` | 特定のオプション項目詳細の取得 | `master-trimming` | `view` |
| POST | `/api/v1/masters/trimming-options` | オプション項目の作成 | `master-trimming` | `create` |
| PATCH | `/api/v1/masters/trimming-options/:id` | オプション項目の更新 | `master-trimming` | `edit` |
| DELETE | `/api/v1/masters/trimming-options/:id` | オプション項目の削除 | `master-trimming` | `delete` |
| PATCH | `/api/v1/masters/trimming-options/reorder` | オプション表示順の一括保存 | `master-trimming` | `edit` |

---

