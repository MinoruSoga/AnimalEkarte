# BE-052: 会計APIレスポンスにOwner/Pet情報が含まれていない

**Status**: Open
**Priority**: High
**Affects**: 会計一覧 (`/accounting`)、会計詳細
**Date Created**: 2026-03-21
**Related**: BUG-004, FE-089 (FE側変更不要・BEのみ対応)

## Summary

`accounting_response.go` の `toAccountingResponse()` が `model.Billing` の `Owner` / `Pet` リレーションをJSONに含めていないため、会計一覧で飼主名・ペット名が表示されない。リポジトリ層では `Preload("Owner").Preload("Pet")` が実行されているためデータは存在するが、レスポンス構造体に対応フィールドがない。

## 現状のコード

```go
// backend/internal/handler/accounting_response.go:40-60
type accountingResponse struct {
    ID                uint64                `json:"id"`
    ClinicID          uint64                `json:"clinic_id"`
    MedicalRecordID   *uint64               `json:"medical_record_id,omitempty"`
    HospitalizationID *uint64               `json:"hospitalization_id,omitempty"`
    OwnerID           *uint64               `json:"owner_id,omitempty"`
    PetID             *uint64               `json:"pet_id,omitempty"`
    // ... Owner/Pet ネストフィールドが存在しない
}

// backend/internal/handler/accounting_response.go:98-127
func toAccountingResponse(b *model.Billing) accountingResponse {
    return accountingResponse{
        OwnerID:  b.OwnerID,
        PetID:    b.PetID,
        // Owner/Pet の owner_name, name を設定していない
    }
}
```

```go
// backend/internal/repository/accounting_repository.go:47-60
// Preload は実行されているのでデータは b.Owner, b.Pet に存在する
if err := q.Preload("Owner").Preload("Pet").Preload("Payments").Preload("Items").
    Find(&billings).Error; err != nil { ... }
```

## 必要な変更

### 1. レスポンス構造体にOwner/Petサマリーを追加

```go
// backend/internal/handler/accounting_response.go

// 追加: オーナーサマリー
type accountingOwnerSummary struct {
    ID        uint64 `json:"id"`
    OwnerName string `json:"owner_name"`
}

// 追加: ペットサマリー
type accountingPetSummary struct {
    ID   uint64 `json:"id"`
    Name string `json:"name"`
}

// 既存 accountingResponse 構造体に追加
type accountingResponse struct {
    ID                uint64                   `json:"id"`
    ClinicID          uint64                   `json:"clinic_id"`
    MedicalRecordID   *uint64                  `json:"medical_record_id,omitempty"`
    HospitalizationID *uint64                  `json:"hospitalization_id,omitempty"`
    OwnerID           *uint64                  `json:"owner_id,omitempty"`
    PetID             *uint64                  `json:"pet_id,omitempty"`
    Owner             *accountingOwnerSummary  `json:"owner,omitempty"`  // ← 追加
    Pet               *accountingPetSummary    `json:"pet,omitempty"`    // ← 追加
    // ... 既存フィールド
}
```

### 2. toAccountingResponse() でOwner/Petを設定

```go
func toAccountingResponse(b *model.Billing) accountingResponse {
    // ... 既存処理 ...

    var owner *accountingOwnerSummary
    if b.Owner != nil {
        owner = &accountingOwnerSummary{
            ID:        b.Owner.ID,
            OwnerName: b.Owner.OwnerName,
        }
    }

    var pet *accountingPetSummary
    if b.Pet != nil {
        pet = &accountingPetSummary{
            ID:   b.Pet.ID,
            Name: b.Pet.Name,
        }
    }

    return accountingResponse{
        // ... 既存フィールド ...
        Owner: owner,
        Pet:   pet,
    }
}
```

## APIレスポンス形式（修正後）

```json
{
  "id": 1,
  "owner_id": 1,
  "pet_id": 1,
  "owner": {
    "id": 1,
    "owner_name": "佐藤太郎"
  },
  "pet": {
    "id": 1,
    "name": "ポチ"
  },
  "total_amount": 5000,
  "status": "pending"
}
```

## フロントエンド影響

`frontend/src/features/accounting/api/transforms.ts:43-45` はすでに `data.owner?.owner_name` / `data.pet?.name` にアクセスしており、BE修正後は即座に動作する。FEの追加変更は不要。

## 完了条件

- [ ] `accountingOwnerSummary` / `accountingPetSummary` 構造体を追加
- [ ] `accountingResponse` に `Owner *accountingOwnerSummary` / `Pet *accountingPetSummary` フィールドを追加
- [ ] `toAccountingResponse()` で `b.Owner` / `b.Pet` が nil でない場合にサマリーを設定
- [ ] `GET /api/v1/billings` レスポンスの各アイテムに `owner.owner_name` と `pet.name` が含まれる
- [ ] 会計一覧画面で飼主名・ペット名が表示される
