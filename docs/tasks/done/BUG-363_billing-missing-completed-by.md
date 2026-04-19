# BUG-363: billings テーブルにステータス変更者（completed_by）が記録されない

## 優先度: MEDIUM → CLOSED (対応不要)

## Close 理由

以下の2つのメカニズムで監査証跡が確保されており、billings.completed_by の追加は不要。

### 1. payments.paid_by (BUG-362 で実装済み)
会計完了（waiting → completed）時に Payment が upsert され、`paid_by` に処理スタッフIDが記録される。
「誰が会計を完了させたか」はここで追跡可能。

### 2. billing_confirmations.confirmed_by / returned_by (既存)
カルテ→会計の承認フローで「誰が医師確認/差戻したか」が記録されている。

### 結論
billings.completed_by を追加すると payments.paid_by と意味が重複し、二重管理になる。
