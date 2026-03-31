# 006: ExaminationSection 検査項目リスト実装

**ステータス:** open
**優先度:** high
**関連API:** `GET /v1/masters/examination-types`（レスポンスの `items[]`）

## 背景

バックエンドのハンドラは `items[]` を既にレスポンスに含めて返している（Preload済み）。
カルテ画面（Tab6: 検査）の `ExaminationForm` は「マスタ連動で自動生成」される検査項目テーブルを必要とする。
マスタ設定画面（`/settings/treatment-items` > 検査タブ）の `ExaminationSection` にitemsの追加・編集UIが未実装。

## 実装が必要なUI

### マスタ設定（ExaminationSection）

| フィールド | UIコンポーネント |
|---|---|
| 検査項目リスト | 動的リスト（追加/削除） |
| └ 項目名 | `Input`（placeholder: 例: RBC） |
| └ 単位 | `Input`（placeholder: 例: mg/dL） |
| └ 基準値 | `Input`（placeholder: 例: 550-850） |
| └ 削除 | Trash2 ボタン |
| 項目追加 | `Button`（Plus アイコン） |

### カルテ画面（ExaminationForm）

検査種別選択後、その `items[]` を自動展開して測定値入力テーブルを表示する。

## 実装場所

- `ExaminationSection` コンポーネントに items 動的リストUI追加
- `GET /v1/masters/examination-types` レスポンスの `items[]` を活用（APIは既実装）
- `api.yaml` の `ExaminationType` スキーマに `items[]` が追加済み

## 型定義

`models.ts` に `ExamTypeItem` は既に定義済み。
`ExaminationType.items?: ExamTypeItem[]` も定義済み。

## 注意

items の作成・更新・削除エンドポイント（`/v1/masters/examination-types/:id/items`）が存在しない場合は、
バックエンドへの追加も同時に検討する。
