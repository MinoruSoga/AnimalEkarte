# 020: 治療明細（treatment items）UI 実装

## 概要
カルテに紐づく治療明細（診療費明細）を追加・編集・削除・並び替えできるUIを実装する。
会計フローのコアとなる機能で、会計画面への連携に使用される。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/treatments`
- POST `/v1/medical-records/{id}/treatments`
- PATCH `/v1/medical-records/{id}/treatments/{treatmentId}`
- DELETE `/v1/medical-records/{id}/treatments/{treatmentId}`
- PUT `/v1/medical-records/{id}/treatments`（sort_order 一括更新）

## 利用可能なAPI hooks（実装済み）
`features/medical-records/api/treatments.ts` に以下が実装済み:
- `useTreatments(medicalRecordId)` — 一覧取得（sort_order 昇順）
- `useCreateTreatment(medicalRecordId)` — 新規作成
- `useUpdateTreatment(medicalRecordId)` — 部分更新
- `useDeleteTreatment(medicalRecordId)` — 削除
- `useReorderTreatments(medicalRecordId)` — 並び替え

`features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx` にテーブルUIが実装済み。

## 実装内容

### 機能要件
- 治療明細を `sort_order` 昇順で表示する
- 明細を追加・編集・削除できる
- 並び替え（PUT bulk update）が動作する
- 合計金額（`unit_price × quantity - discount_amount`）をクライアント側で自動計算・表示する
- `insurance` フラグ（保険対象）を各行に表示する
- `selected` フラグの切り替えができる
- `item_type`（consultation / procedure / medicine / other）別の表示区分けを行う

### 治療明細フィールド

| フィールド | 型 | 備考 |
|----------|-----|------|
| item_type | enum | consultation / procedure / medicine / other |
| content | string | 治療内容 |
| unit_price | number | 単価（税込）|
| quantity | int | 数量 |
| discount_rate | number | 割引率(%) |
| discount_amount | number | 値引額(￥) |
| insurance | bool | 保険対象フラグ |
| selected | bool | 選択フラグ |
| memo | string | メモ |

## 完了条件
- [ ] 治療明細一覧が `sort_order` 昇順で表示される
- [ ] 明細を追加・編集・削除できる
- [ ] 合計金額がリアルタイムで自動計算される
- [ ] `item_type` 別の表示区分けが実装されている
- [ ] 新規カルテ（未保存）時は利用不可の旨を表示する
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- `TreatmentsTab.tsx` のテーブル実装を参照・再利用すること
- 現在の `MedicalRecordDiagnosisPlan` 内のモック `TreatmentTable` との役割分担を整理すること
