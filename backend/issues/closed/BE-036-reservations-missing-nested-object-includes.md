# BE-036: GET /v1/reservations レスポンスに animal_species オブジェクトが含まれない

## 概要

Dashboard ページの受付状況表示で、予約データ取得時に `animal_species` オブジェクトがネストされていない問題。

## 問題の詳細

フロントエンドでは Pet 型に `animal_species?: AnimalSpecies` が期待されているが、バックエンドの `/v1/reservations` API レスポンスで include されていない。

### 実際のレスポンス構造
```json
{
  "pet": {
    "id": 13,
    "animal_species_id": 2,
    "name": "ソラ"
    // animal_species オブジェクトなし
  }
}
```

### 期待値
```json
{
  "pet": {
    "id": 13,
    "animal_species_id": 2,
    "animal_species": {
      "id": 2,
      "name": "猫",
      "is_active": true,
      "sort_order": 2
    },
    "name": "ソラ"
  }
}
```

## 影響範囲

- Dashboard（当日の受付）ページ
- その他 Reservation を使用する全 API

## 修正方法

`internal/repository/reservation_repository.go` の FindAll / FindByID メソッドで、Pet の Preload に `AnimalSpecies` を追加

```go
db.Preload("Pet").Preload("Pet.AnimalSpecies")  // ← AnimalSpecies を追加
```

## 関連する他の API も確認

- POST /v1/reservations
- PATCH /v1/reservations/{id}
- DELETE /v1/reservations/{id}
- status/update エンドポイント

同じ問題を持つ可能性がある

## 優先度

- 中（Dashboard は正常に動作するが、フロントでワークアラウンド対応が必要）

## テスト方法

1. GET /v1/reservations?date=2026-03-16 を呼び出し
2. レスポンスの pet.animal_species が存在し、name が含まれているか確認
