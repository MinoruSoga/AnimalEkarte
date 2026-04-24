# BE-medical-record-accounting-id: カルテ一覧レスポンスに accounting_id を含める

## 背景

カルテ一覧画面の「関連」カラムに会計レコードへのリンク（`会計`バッジ）を表示するため、
カルテ一覧 API のレスポンスに `accounting_id` を含める必要がある。

現在 `transforms.ts` の `accountingId: undefined` が固定であり、会計が存在するカルテでも
バッジが表示されない。

## 現状

- `Accounting` モデルは `medical_record_id` FK で `MedicalRecord` を参照している
- `MedicalRecord` モデルには逆方向の `Accounting` リレーションが存在しない
- 一覧取得クエリでは `Accounting` の Preload が行われていない

## 対応内容

### 1. `MedicalRecord` モデルに逆リレーション追加

```go
// backend/internal/model/medical_record.go
type MedicalRecord struct {
    // ... 既存フィールド ...

    // Relations
    // ... 既存 ...
    Accounting *Accounting `gorm:"foreignKey:MedicalRecordID" json:"accounting,omitempty"`
}
```

### 2. リポジトリの一覧クエリに Preload 追加

```go
// backend/internal/repository/medical_record_repository.go
db.Preload("Owner").
    Preload("Pet.AnimalSpecies").
    Preload("Doctor").
    Preload("Inquiry").
    Preload("Accounting").  // ← 追加
    Find(&records)
```

### 3. make codegen で models.ts を再生成

```bash
make codegen
```

生成後、`frontend/src/types/generated/models.ts` の `MedicalRecord` に
`accounting?: Accounting` フィールドが追加される。

## フロントエンド対応（BE修正後）

`frontend/src/features/medical-records/api/transforms.ts` を以下に修正：

```typescript
accountingId: record.accounting?.id ? String(record.accounting.id) : undefined,
```

## 対応フロントエンド

- `frontend/src/features/medical-records/api/transforms.ts` — `accountingId` フィールド
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx` — `関連` カラムの `会計` バッジ表示

## 優先度

Medium（UI上は `—` が表示されるため機能的に壊れていないが、Figma仕様との差分）
