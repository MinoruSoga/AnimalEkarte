# 019: バイタル記録モーダル実装

## 概要
カルテ画面でバイタル（体温・心拍数・呼吸数・体重）を記録・参照できるモーダルを実装する。
記録の追加・編集・削除と、過去の記録一覧の参照を1つのモーダル内で行う。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/vitals`
- POST `/v1/medical-records/{id}/vitals`
- PATCH `/v1/medical-records/{id}/vitals/{vitalId}`
- DELETE `/v1/medical-records/{id}/vitals/{vitalId}`

## 利用可能なAPI hooks（実装済み）
`features/medical-records/api/vitals.ts` に以下が実装済み:
- `useVitals(medicalRecordId)` — 一覧取得
- `useCreateVital(medicalRecordId)` — 新規作成
- `useUpdateVital(medicalRecordId)` — 更新
- `useDeleteVital(medicalRecordId)` — 削除

`features/medical-records/components/VitalsTab/VitalsTab.tsx` にテーブルUIが実装済み。

## 実装内容

### 機能要件
- バイタル記録の一覧を `recorded_at` 昇順で表示する
- 新規バイタルを追加できる
- バイタル記録を編集・削除できる
- 体温・心拍数・呼吸数・体重はすべて任意入力（空欄可）
- モーダルを閉じると入力内容はリセットされる

### 表示フィールド

| フィールド | 型 | 単位 |
|----------|-----|------|
| 記録日時 | datetime | - |
| 体温 | number | ℃ |
| 心拍数 | number | bpm |
| 呼吸数 | number | /min |
| 体重 | number | kg |
| メモ | string | - |

## 完了条件
- [ ] バイタル一覧が `recorded_at` 昇順で表示される
- [ ] 新規バイタルを追加できる（追加後にキャッシュ invalidate）
- [ ] バイタルを編集・削除できる
- [ ] 数値フィールドは空送信時に `null` で送る
- [ ] 新規カルテ（未保存）時は利用不可の旨を表示する
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- `VitalsTab.tsx` のテーブル実装を参照・再利用すること
- `VitalInputDialog.tsx`（現在ローカルstateのみ）はこのモーダルに置き換えるか、拡張する
