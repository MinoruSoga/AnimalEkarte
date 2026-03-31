---
status: open
---

# [reservation] GET /v1/reservations の FindAll に Pet/ServiceType/Doctor の Preload 欠落

## 背景

予約管理ページ（ダッシュボード・予約一覧）はペット名、サービス種別、担当医を
表示するが、`reservation_repository.FindAll` に Preload が一切ないため
これらのフィールドがすべて空で返る。

## 問題

```go
// reservation_repository.go FindAll: Preload なし
q.Offset((page-1)*limit).Limit(limit).Order("start_time ASC").Find(&reservations)

// FindByID: Preload あり
r.db.Preload("Pet").Preload("ServiceType").Preload("Doctor").First(...)
```

`GET /v1/reservations` は `pet_id`, `service_type_id`, `doctor_id` の ID しか返さない。
フロントエンドでペット名・サービス名・担当医名を表示できない。

## 修正方針

`reservation_repository.FindAll` に Preload を追加:

```go
if err := q.
    Preload("Pet").
    Preload("Pet.Owner").
    Preload("ServiceType").
    Preload("Doctor").
    Offset((page - 1) * limit).Limit(limit).Order("start_time ASC").
    Find(&reservations).Error; err != nil {
```

`reservationResponse` にネスト情報が欠落していれば合わせて追加する。

## 完了条件

- [ ] `reservation_repository.FindAll` に `Preload("Pet").Preload("Pet.Owner").Preload("ServiceType").Preload("Doctor")` 追加
- [ ] `GET /v1/reservations` レスポンスにペット名・サービス名・担当医名が含まれる
- [ ] ダッシュボードのカンバン表示でペット名・担当医が正しく表示される
- [ ] `docker compose exec backend go test ./... -v` がパス
