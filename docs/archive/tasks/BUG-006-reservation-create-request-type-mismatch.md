# BUG-006: 予約新規作成で400エラー（リクエスト型不一致）

## 種類
バグ

## 発見日
2026-03-20

## 再現手順
1. `/reservations` で「新規予約」ボタンをクリック
2. 患者検索 → Iris選択
3. 予約区分「一般診療」、担当者「山田太郎」、初診、メモ入力
4. 「予約を確定」クリック

## 期待動作
予約がDBに保存され、カレンダーに表示される。

## 実際の動作
400 Bad Request:
```
json: cannot unmarshal string into Go struct field createReservationRequest.pet_id of type uint64
```

## リクエストボディ（実際に送信された値）
```json
{
  "pet_id": "1",           // ← 文字列（数値であるべき）
  "owner_id": "1",         // ← 文字列（数値であるべき）
  "doctor_id": "山田 太郎", // ← 名前（IDであるべき）
  "service_type": "一般診療" // ← 名前（IDであるべき）
}
```

## 根本原因
フロントエンドの予約作成APIで：
1. `pet_id`, `owner_id` が文字列型で送信されている（`Number()` 変換が必要）
2. `doctor_id` にスタッフ名が送信されている（スタッフIDを送るべき）
3. `service_type_id` にサービス種別名が送信されている（サービス種別IDを送るべき）

## 修正方針
`frontend/src/features/reservations/` の予約作成API呼び出し箇所で：
- ID系フィールドを `Number()` で数値変換
- `doctor_id` はセレクトの値をIDに変更
- `service_type` → `service_type_id` に変更し、IDを送信
