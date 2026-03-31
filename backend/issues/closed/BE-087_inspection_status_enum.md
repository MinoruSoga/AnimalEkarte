# BE-087: 検査ステータス enum に「結果入力済み」「確定」を追加

## 概要
検査（examinations）テーブルの `status` enum に2値を追加し、
確定後は更新を拒否するバリデーションをサービス層に実装する。

## 現状
`ExaminationStatus` enum: `pending` / `in_progress` / `completed`

## 変更内容

### 1. migration: enum 値追加

```sql
-- backend/migrations/001_init.sql を直接編集（リリース前運用のため）
ALTER TYPE examination_status ADD VALUE IF NOT EXISTS 'result_entered';
ALTER TYPE examination_status ADD VALUE IF NOT EXISTS 'confirmed';
```

001_init.sql の `CREATE TYPE examination_status AS ENUM` 行に以下を追加:
```sql
'result_entered',  -- 結果入力済み
'confirmed'        -- 確定（編集禁止）
```

### 2. model 更新

```go
// backend/internal/model/examination.go
const (
  ExaminationStatusPending       ExaminationStatus = "pending"
  ExaminationStatusInProgress    ExaminationStatus = "in_progress"
  ExaminationStatusResultEntered ExaminationStatus = "result_entered"
  ExaminationStatusCompleted     ExaminationStatus = "completed"
  ExaminationStatusConfirmed     ExaminationStatus = "confirmed"
)
```

### 3. service: 確定後更新禁止

```go
// backend/internal/service/examination_service.go の Update メソッド
func (s *ExaminationService) Update(ctx context.Context, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
  exam, err := s.repo.GetByID(ctx, id)
  if err != nil {
    return nil, apperrors.Wrap(err, "failed to get examination")
  }
  if exam.Status == model.ExaminationStatusConfirmed {
    return nil, apperrors.NewInvalidInput("確定済みの検査は編集できません")
  }
  // ... 以降の更新処理
}
```

### 4. make codegen 実行

モデル変更後に `make codegen` を実行し `frontend/src/types/generated/models.ts` を更新する。

## 影響範囲
- `backend/migrations/001_init.sql`
- `backend/internal/model/examination.go`
- `backend/internal/service/examination_service.go`
- `frontend/src/types/generated/models.ts`（codegen 後に自動更新）

## 優先度
Low

## 関連
- docs/tasks/open/feature/BUG-087_inspection_status_missing.md
