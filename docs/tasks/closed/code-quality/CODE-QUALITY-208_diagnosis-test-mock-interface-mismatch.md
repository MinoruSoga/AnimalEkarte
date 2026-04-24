# CODE-QUALITY-208: diagnosis_handler_test.go の mock インターフェース不一致（go vet 失敗）

## 概要

`diagnosis_service.go` の `DiagnosisNameService.List` シグネチャに `typeID *uint64` 引数が追加されたが、
`diagnosis_handler_test.go` のモックが旧シグネチャのまま残っており、`go vet` が失敗してテストビルドが壊れている。

## 優先度

**CRITICAL（マージブロック）**

## 影響ファイル

| ファイル | 問題 |
|---------|-----|
| `backend/internal/handler/diagnosis_handler_test.go` | L53-74: mock シグネチャが旧 API のまま |

---

## 問題

### go vet エラー内容

```
cannot use &mockDiagnosisNameService{} as service.DiagnosisNameService value:
  have List(context.Context, uint64, int, int)
  want List(context.Context, uint64, *uint64, int, int)
```

### 現状コード（diagnosis_handler_test.go）

```go
// 旧シグネチャ（間違い）
func (m *mockDiagnosisNameService) List(
    ctx context.Context,
    clinicID uint64,
    page, limit int,  // ← typeID *uint64 が欠落
) ([]model.DiagnosisName, int64, error) {
    return m.listFn(ctx, clinicID, page, limit)
}

// 旧メソッド（インターフェースから削除済みのため不要）
func (m *mockDiagnosisNameService) ListByCategoryID(
    ctx context.Context, clinicID, categoryID uint64, page, limit int,
) ([]model.DiagnosisName, int64, error) {
    return m.listByCategoryIDFn(ctx, clinicID, categoryID, page, limit)
}
```

### 現在のインターフェース（diagnosis_service.go）

```go
// 新シグネチャ（正）
List(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error)
// ※ ListByCategoryID は削除済み
```

---

## 修正方針

### Step 1: mock の struct と List シグネチャを更新

```go
type mockDiagnosisNameService struct {
    listFn func(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error)
    // listByCategoryIDFn フィールドは削除
    listNamesFn func(...)
    // ... 他フィールドは維持
}

func (m *mockDiagnosisNameService) List(
    ctx context.Context,
    clinicID uint64,
    typeID *uint64,   // ← 追加
    page, limit int,
) ([]model.DiagnosisName, int64, error) {
    return m.listFn(ctx, clinicID, typeID, page, limit)
}

// ListByCategoryID メソッドを完全に削除
```

### Step 2: テストケースの listFn を新シグネチャに合わせて更新

各テストケースで `listFn` を設定している箇所を `typeID *uint64` 引数を受け取るように修正する。

---

## 規約参照

- `.claude/rules/go-language.md`: インターフェース設計
- `go vet` 失敗はマージブロック要件

## テスト

修正後 `docker compose exec backend go vet ./...` が成功することを確認。
