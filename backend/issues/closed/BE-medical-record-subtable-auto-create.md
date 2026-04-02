# BE: カルテ作成時にサブテーブルを空レコードで自動作成する

## 背景
フロントエンドでは、カルテ新規作成ページ表示時に即座に空のカルテレコードを作成し、
その後各タブで保存する度にそのタブのサブテーブルを upsert する設計になった。

現状の create medical record エンドポイント（POST /v1/medical-records）は
メインの `medical_records` テーブルのみ作成する。
各タブの更新エンドポイントは UPDATE 前提のため、サブテーブルが存在しない場合 404 になる。

## 必要な変更

### 1. POST /v1/medical-records
カルテ作成時に以下のサブテーブルも空レコードで自動作成する:
- `inquiries`（問診）: medical_record_id を持つ空レコード
- `treatment_plans`（治療プラン）: medical_record_id を持つ空レコード
- `clinical_plans`（診察所見）: medical_record_id を持つ空レコード

### 2. 各サブテーブルの UPDATE エンドポイントを UPSERT に変更
対象エンドポイント:
- PUT/PATCH /v1/medical-records/:id/inquiries
- PUT/PATCH /v1/medical-records/:id/treatment-plans
- PUT/PATCH /v1/medical-records/:id/clinical-plans
- PUT/PATCH /v1/medical-records/:id/estimates

レコードが存在しない場合は INSERT、存在する場合は UPDATE する（ON CONFLICT DO UPDATE パターン）。

## 優先度
High（フロントエンドのタブ別保存実装に必須）
