# BE-TEST-001: treatment_service_test.go の NewTreatmentService 呼び出しを修正

## 概要
`NewTreatmentService` のシグネチャが `(repos *repository.Repositories)` に変更されたが、
テストファイルが旧シグネチャ `(repo, inventoryRepo)` のまま。CI の `go vet` が失敗している。

## エラー内容
```
internal/service/treatment_service_test.go:96:37: too many arguments in call to NewTreatmentService
  have (*mockTreatmentRepository, *mockInventoryRepository)
  want (*repository.Repositories)
```

## 対象ファイル
`backend/internal/service/treatment_service_test.go`

## 影響行
- L96, L192, L343, L424, L487（計5箇所）

## 修正方針

### 現状
```go
svc := NewTreatmentService(repo, &mockInventoryRepository{})
```

### 修正後
`*repository.Repositories` を渡す必要があるが、テストでは mock を使いたい。

以下のいずれかを選択：

**案A: Repositories struct に mock を詰めて渡す**
```go
repos := &repository.Repositories{
    Treatment: repo,
    Inventory: &mockInventoryRepository{},
    // 他フィールドは nil でも今のテストでは問題なし
}
svc := NewTreatmentService(repos)
```

**案B: TreatmentService のコンストラクタに interface を直接受け取る形に戻す**
```go
func NewTreatmentService(treatmentRepo TreatmentRepository, inventoryRepo InventoryRepository) TreatmentService
```
→ ただし他サービスとの一貫性を崩すため非推奨。

**推奨: 案A** — `Repositories` struct の `Treatment` / `Inventory` フィールドに mock を設定して渡す。

## 優先度
High（CI が全ブランチで失敗している）

## 関連
- `backend/internal/service/treatment_service.go:82`
- `backend/internal/repository/repositories.go`（Repositories struct の確認が必要）
