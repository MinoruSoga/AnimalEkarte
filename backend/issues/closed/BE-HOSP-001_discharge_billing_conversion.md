# BE-HOSP-001: 退院時の入院プラン → 会計明細への自動変換 API

## 概要
退院処理時に「そのまま会計画面へ進む」オプションが選択された場合、
入院中に登録したケアプラン（`care_plan_items`）を会計明細（`accounting_details`）へ
自動変換して会計レコードを生成する API エンドポイントを追加する。

## 背景
フロントエンドの `DischargeAlertDialog` に「退院後、そのまま会計画面へ進む」
チェックボックスを追加済み（2026-03-31）。
現状は単に `/accounting/new?petId=XXX` へ遷移するだけで、
プラン明細は手動入力になっている。
バックエンドで退院→会計の自動連携を実現することで二重入力を防ぐ。

## 実装内容

### エンドポイント
```
POST /api/v1/hospitalizations/:id/discharge-with-billing
```

### リクエスト
```json
{
  "discharge_date": "2026-03-31",
  "create_accounting": true
}
```

### レスポンス
```json
{
  "hospitalization_id": 1,
  "accounting_id": 42,   // 生成された会計レコードID（create_accounting=true の場合）
  "status": "discharged"
}
```

### 処理フロー（service 層）
1. `hospitalizations.status = 'discharged'`, `end_date = discharge_date` に更新
2. `create_accounting = true` の場合:
   a. `care_plan_items` を取得
   b. `accounting` レコードを作成（`status = 'waiting'`, `pet_id`, `scheduled_date = discharge_date`）
   c. `care_plan_items` を `accounting_details` に変換して一括挿入
   d. 合計金額を計算して `accounting.total_amount` を更新
3. トランザクション内で原子的に実行

### model 変更
なし（既存テーブルのみ使用）

## 優先度
Medium

## 関連
- フロントエンド: `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx`
- フロントエンド: `frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx`
- 仕様書: `docs/screens/08-hospitalization-detail.md`
