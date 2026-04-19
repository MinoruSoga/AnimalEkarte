# TASK-029: 検査・診断系マスタ MEDIUM 問題 3件

## 優先度

MEDIUM

---

## 問題 1: Repository の Update メソッド命名・シグネチャが 4 ドメイン間で不統一

### ファイル

| ドメイン | メソッド | 戻り値 |
|---------|---------|--------|
| `exam_type_repository.go` | `UpdateFields` | `*model.ExaminationType, error` |
| `checkup_type_repository.go` | `UpdateFields` | `*model.CheckupType, error` |
| `diagnosis_repository.go` | `Update` | `error` のみ（service が別途 FindByID を呼ぶ） |
| `chief_complaint_type_repository.go` | `Update` | `error` のみ（同上） |

### 問題
`UpdateFields`（entity 返却）と `Update`（error のみ）が混在。`Update` 系は更新後に `FindByID` を追加で呼ぶため DB クエリが1本多くなる。

### 修正案
`UpdateFields` に統一し、全ドメインで更新済みエンティティを直接返す。diagnosis / chief_complaint の service の `FindByID` 呼び出しを削除してクエリを削減する。

---

## 問題 2: 全ドメインの Reorder に slog.InfoContext なし

### ファイル
- `backend/internal/service/exam_type_service.go:84-92`
- `backend/internal/service/checkup_type_service.go:86-94`
- `backend/internal/service/diagnosis_service.go:159-167`（相当箇所）
- `backend/internal/service/chief_complaint_type_service.go`（Reorder 未実装：TASK-026 で追加予定）

### 問題
Create / Update / Delete には `slog.InfoContext` があるが、Reorder だけ4ドメイン全てでログなし。

### 修正案
```go
func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder exam types")
    }
    slog.InfoContext(ctx, "exam_types reordered",
        slog.Uint64("clinic_id", clinicID),
        slog.Int("count", len(ids)))
    return nil
}
```
checkup_type / diagnosis_type も同パターンで追加。

---

## 問題 3: diagnosis ハンドラで List の total をレスポンスに含めていない（計算コストの無駄）

### ファイル
`backend/internal/handler/diagnosis_handler.go:47, 186, 193`

### 問題
`DiagnosisType.List` と `DiagnosisName.List` が service/repository レベルで `total int64` を計算しているが、handler 側で `_` で捨ててレスポンスに含めていない。ページネーション計算だけが走り結果を使わない。他3ドメインの List は total なしで実装されており実装が非対称。

### 修正案（2択、プロジェクト方針に応じて選択）

**選択肢 A**: total をレスポンスに含める
```go
c.JSON(http.StatusOK, gin.H{
    "items": mapSlice(items, toDiagnosisTypeResponse),
    "total": total,
})
```

**選択肢 B**: List シグネチャから total を除去して他3ドメインと統一（推奨）
```go
// service / repository から total 戻り値を削除
// handler 側の _ 変数を削除
```
