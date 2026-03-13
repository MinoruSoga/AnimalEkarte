# [IMPROVEMENT] 飼主新規登録時のペット一括作成: アトミックなバッチエンドポイントの検討

## 優先度: Medium

---

## 背景

飼主新規登録画面では、飼主登録と同時に複数のペットを追加できる。

現在のフロントエンド実装:
1. `POST /v1/owners` → 飼主を作成（owner_id を取得）
2. `Promise.all([POST /v1/pets, POST /v1/pets, ...])` → ペットをN件並列作成

---

## 問題点

### 1. アトミック性の欠如（最重要）

飼主作成後にペット作成が途中で失敗した場合、**飼主だけが作成されペットが存在しない中途半端な状態**になる。

```
POST /v1/owners  → 201 OK  (owner_id=42 が作成される)
POST /v1/pets    → 201 OK  (1匹目 成功)
POST /v1/pets    → 500 Error (2匹目 失敗)
→ 飼主は作成済みだが、2匹目のペットは存在しない
→ ユーザーが再試行しても重複飼主が作成されるリスク
```

### 2. N+1 リクエスト

ペット3匹登録 = 1(飼主) + 3(ペット) = 4リクエスト。ペット数に比例してリクエストが増える。

---

## 提案: バッチ作成エンドポイント or トランザクション統合

### 案A: `POST /v1/owners`のレスポンスにペット作成を統合

リクエストボディにペット配列を含めて1リクエストで完結させる。

```json
POST /v1/owners
{
  "owner_name": "林 文昭",
  "phone": "090-1234-5678",
  ...
  "pets": [
    {
      "name": "ポチ",
      "animal_species_id": 1,
      "gender": "male"
    }
  ]
}
```

サービス層でトランザクション内に飼主 + ペット作成をまとめる:
```go
func (s *ownerService) Create(ctx context.Context, owner *model.Owner) error {
    return s.repo.CreateWithPets(ctx, owner, owner.Pets)
}
```

```go
// owner_repository.go
func (r *ownerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(owner).Error; err != nil {
            return apperrors.Wrap(err, "create owner")
        }
        for i := range pets {
            pets[i].OwnerID = owner.ID
            pets[i].ClinicID = owner.ClinicID
            if err := tx.Create(&pets[i]).Error; err != nil {
                return apperrors.Wrap(err, "create pet")
            }
        }
        return nil
    })
}
```

### 案B: `POST /v1/owners/:id/pets/batch`

```json
POST /v1/owners/42/pets/batch
{
  "pets": [
    { "name": "ポチ", "animal_species_id": 1, "gender": "male" },
    { "name": "タマ", "animal_species_id": 2, "gender": "female" }
  ]
}
```

---

## 推奨

**案A**が望ましい。理由:
- 1リクエストで完結（レイテンシ最小）
- DBトランザクションでアトミック性が保証される
- フロントエンドのロジックがシンプルになる（Promise.all 不要）
- 後方互換性: `pets` フィールドは optional なので既存の呼び出し元に影響なし

---

## ステータス: 🔲 未対応

現在は暫定としてフロント側 `Promise.all` で対応中。
