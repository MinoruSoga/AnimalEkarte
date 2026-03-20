# BUG-006: 予約新規作成が400エラー — リクエストの型不一致

## 種類
バグ

## 発見日
2026-03-20

## 再現手順
1. `/reservations` で「新規予約」ボタンをクリック
2. 患者検索でペットを選択
3. 予約区分、担当者、メモを入力
4. 「予約を確定」ボタンをクリック

## 期待動作
予約がDBに保存される。

## 実際の動作
`POST /api/v1/reservations` → 400 Bad Request

## エラー詳細
```
json: cannot unmarshal string into Go struct field createReservationRequest.pet_id of type uint64
```

## リクエストボディ（実際）
```json
{
  "pet_id": "1",           // ← 文字列（バックエンドは uint64 を期待）
  "owner_id": "1",         // ← 文字列
  "start_time": "2026-03-20T01:00:00.000Z",
  "end_time": "2026-03-20T02:00:00.000Z",
  "visit_type": "first",
  "service_type": "一般診療",  // ← 名前文字列（IDを送るべき）
  "doctor_id": "山田 太郎",   // ← 名前文字列（IDを送るべき）
  "is_designated": false,
  "notes": "テスト予約メモ"
}
```

## 修正方針
フロントエンドの予約作成API呼び出しで：
1. `pet_id`, `owner_id` を数値型で送信
2. `service_type_id` にサービス種別のIDを送信（名前ではなく）
3. `doctor_id` に担当者のIDを送信（名前ではなく）
