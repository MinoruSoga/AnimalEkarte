# BUG-397: マスタサービスの Update で len(fields)==0 時の動作が不統一

## 概要
マスタサービスの `Update` メソッドで、更新フィールドが空（すべて nil）の場合の動作が2つのパターンに分裂している。
- **パターンA（11サービス）**: `apperrors.WrapInvalidInput("at least one field must be provided")` を返す（400 Bad Request）
- **パターンB（4サービス）**: 既存レコードを `FindByID` で取得して返す（200 OK）

パターンBは余計なDBクエリを発行し、クライアントから見てPATCHセマンティクスが曖昧になる。

## 再現手順
1. `PATCH /masters/medicines/:id` に空のボディ `{}` を送信
2. **結果（medicine）**: 200 OK + 既存レコードを返す
3. `PATCH /masters/cages/:id` に空のボディ `{}` を送信
4. **結果（cage）**: 400 Bad Request + `"at least one field must be provided"`
5. **同じ PATCH セマンティクスで挙動が異なる**

## 現状コード

### パターンA（400 を返す — 多数派、11サービス）
```go
// backend/internal/service/cage_service.go:101
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput("at least one field must be provided")
}
```

### パターンB（200 OK で既存レコードを返す — 少数派、4サービス）
```go
// backend/internal/service/medicine_service.go:225
if len(fields) == 0 {
    result, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get medicine")
    }
    return result, nil  // ← 余計な SELECT が発生する
}

// 同様のパターン:
// merchandise_item_service.go:161
// reservation_type_group_service.go:101
// reservation_type_service.go:268
```

## 影響範囲

| サービス | 動作 |
|---------|------|
| animal_species_service | パターンA（400） |
| cage_service | パターンA（400） |
| checkup_type_service | パターンA（400） |
| chief_complaint_service | パターンA（400） |
| diagnosis_service | パターンA（400） |
| exam_type_service | パターンA（400） |
| insurance_service | パターンA（400） |
| occupation_service | パターンA（400） |
| procedure_service | パターンA（400） |
| trimming_master_service | パターンA（400） |
| vaccine_service | パターンA（400） |
| **medicine_service** | **パターンB（200）** |
| **merchandise_item_service** | **パターンB（200）** |
| **reservation_type_group_service** | **パターンB（200）** |
| **reservation_type_service** | **パターンB（200）** |

## 修正方針

### 推奨: パターンA に統一（400 を返す）

PATCH の意図は「指定したフィールドのみ更新する」であり、何も指定しない場合はクライアントのミスとして扱うべき。パターンAが正しい。

```go
// medicine_service.go:225（修正後）
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
    // ← BUG-385 の日本語化対応と合わせて実施
}

// merchandise_item_service.go:161, reservation_type_group_service.go:101,
// reservation_type_service.go:268 も同様に修正
```

**注意**: medicine_service と merchandise_item_service の Update テストがパターンBの動作を前提にしている可能性があるため、テスト修正も必要。

## 準拠すべきプロジェクト規約・ベストプラクティス

### プロジェクト内参照実装
`backend/internal/service/cage_service.go:101` — パターンA（400）の正しい実装

## 優先度
**Medium** — API 動作の一貫性問題。クライアント側でサービスごとに異なる動作を前提にしたコードを書く必要が生じる。パターンBは余計な SELECT クエリも発生させる。

## 関連チケット
- **BUG-385**: WrapInvalidInput メッセージの日本語化（修正時に合わせて実施）

## 関連ファイル
- `backend/internal/service/medicine_service.go:225` — 修正対象
- `backend/internal/service/merchandise_item_service.go:161` — 修正対象
- `backend/internal/service/reservation_type_group_service.go:101` — 修正対象
- `backend/internal/service/reservation_type_service.go:268` — 修正対象
- `backend/internal/service/cage_service.go:101` — 参照実装（正）
