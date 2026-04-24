# BUG-407: diagnosis_repository.FindAll がページネーション対応だが total count を返さない

## 概要
`diagnosis_repository.go` の `DiagnosisTypeRepository.FindAll` および `DiagnosisNameRepository.FindAll` は `page, limit int` パラメータを受け取るページネーション対応メソッドだが、戻り値に `total int64`（全件数）が含まれていない。`medicine_repository` や `merchandise_item_repository` など他のページネーション対応リポジトリは `([]T, int64, error)` を返す標準パターンを採用しており、diagnosis だけが不統一。フロントエンドは全件数なしではページネーション UI（総ページ数・次ページの有無）を正しく実装できない。

## 再現手順
```bash
GET /v1/masters/diagnoses?page=1&limit=20
# レスポンスに total が含まれない → フロントエンドで全件数不明
```

## 現状コード

### `backend/internal/repository/diagnosis_repository.go:17`（インターフェース）
```go
type DiagnosisTypeRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, error)
    //                                                              ↑ total int64 が欠落
    ...
}

type DiagnosisNameRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, error)
    //                                                              ↑ 同様に欠落
    ...
}
```

### 比較: 正しい実装（medicine_repository）
```go
type MedicineRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
    //                                                              ↑ total int64 を返す ✅
}
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/repository/diagnosis_repository.go:17` | `FindAll` の戻り値に `int64` 追加 |
| `backend/internal/repository/diagnosis_repository.go:114` | `DiagnosisNameRepository.FindAll` も同様 |
| `backend/internal/repository/diagnosis_repository.go:32-43` | 実装で `COUNT` クエリを追加 |
| `backend/internal/service/diagnosis_service.go` | `FindAll` 呼び出し側で total を受け取り上位に返す |
| `backend/internal/handler/diagnosis_handler.go` | `newPaginatedResponse(items, total, page, limit)` でラップして返す |
| `backend/internal/service/diagnosis_service_test.go` | テスト修正 |

## 修正方針

### 1. `diagnosis_repository.go` — インターフェースと実装を修正
```go
type DiagnosisTypeRepository interface {
    FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
    ...
}

func (r *diagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
    var total int64
    base := r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(clinicScope(clinicID))

    if err := base.Count(&total).Error; err != nil {
        return nil, 0, apperrors.FromGORM(err, "diagnosis_type", "")
    }

    categories := make([]model.DiagnosisType, 0)
    if err := base.
        Preload("Names").
        Offset((page - 1) * limit).Limit(limit).
        Order("sort_order ASC, name ASC").
        Find(&categories).Error; err != nil {
        return nil, 0, apperrors.FromGORM(err, "diagnosis_type", "")
    }
    return categories, total, nil
}
```

### 2. `diagnosis_service.go` — total を伝搬させる

### 3. `diagnosis_handler.go` — ページネーションレスポンスに total を含める
```go
items, total, err := h.service.Diagnosis.ListTypes(ctx, clinicID, page, limit)
// ...
c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(items, toDiagnosisTypeResponse), total, page, limit))
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### プロジェクト内参照実装
`backend/internal/repository/medicine_repository.go` — `([]model.Medicine, int64, error)` の正しいページネーション実装

## 優先度
**Medium** — フロントエンドが診断名・診断タイプ一覧の総件数を取得できないため、ページネーション UI が正しく実装できない。現状は全件を 1 ページで取得している可能性があり、データ増加時に問題が顕在化する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/diagnosis_repository.go:17,32-43,114` — 修正対象
- `backend/internal/service/diagnosis_service.go` — 呼び出し側修正が必要
- `backend/internal/handler/diagnosis_handler.go` — レスポンス修正が必要
- `backend/internal/repository/medicine_repository.go` — 参照実装
