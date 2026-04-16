# BUG-350: ReservationType / ChiefComplaintType モデルのフィールドに json タグが欠落

## 概要

`reservation_type.go` の LINE 予約用フィールド 5件に `json` タグが付いていない。
Go のデフォルトシリアライズ（PascalCase）でレスポンスが返され、フロントエンドが snake_case で期待するフィールドを受け取れない。

## 影響範囲

- **機能**: LINE 予約画面の予約種別表示
- **症状**: 予約種別の説明文・表示名・アイコン画像が空になる

## 欠落フィールド

```go
// backend/internal/model/reservation_type.go（現在・バグあり）
Description             string `gorm:"default:''"`                 // json タグなし → "Description" でシリアライズ
ReservationDisplayName  string `gorm:"not null;default:''"`        // json タグなし → "ReservationDisplayName"
ShortName               string `gorm:"not null;default:''"`        // json タグなし → "ShortName"
ReservationComment      string `gorm:"not null;default:''"`        // json タグなし → "ReservationComment"
ReservationImageURL     string `gorm:"not null;default:''"`        // json タグなし → "ReservationImageURL"
```

## 修正内容

```go
// 修正後
Description             string `gorm:"default:''"                              json:"description"`
ReservationDisplayName  string `gorm:"not null;default:''"                    json:"reservation_display_name"`
ShortName               string `gorm:"not null;default:''"                    json:"short_name"`
ReservationComment      string `gorm:"not null;default:''"                    json:"reservation_comment"`
ReservationImageURL     string `gorm:"not null;default:''"                    json:"reservation_image_url"`
```

## 確認方法

1. `GET /api/clinics/:id/reservation-types` を実行
2. 修正前: レスポンスに `"ReservationDisplayName"` (PascalCase) が返る、または `"reservation_display_name"` が含まれない
3. 修正後: `"reservation_display_name"`, `"description"` 等 snake_case で返る

## 追加: ChiefComplaintType も同様

```go
// backend/internal/model/chief_complaint_type.go:12（バグ）
Description string `gorm:"default:''"` // json タグなし → "Description" でシリアライズ
```

修正:
```go
Description string `gorm:"default:''"                              json:"description"`
```

## 関連ファイル

- `backend/internal/model/reservation_type.go:23,30,32,35,36`
- `backend/internal/model/chief_complaint_type.go:12`
