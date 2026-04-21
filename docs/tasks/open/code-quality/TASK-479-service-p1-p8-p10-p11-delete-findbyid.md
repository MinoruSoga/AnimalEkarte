---
title: Service P1/P8/P10 Delete/Update FindByID + Error Wrapping
issue: '#479'
priority: HIGH
status: open
area: service
pattern: P1/P8/P10
---

## 概要

Service 層の Delete/Update メソッドで、以下の複数パターン違反が 7 ファイルで検出されました：

1. **P1 違反**：FindByID を削除前に呼び出していない、または呼び出すが条件チェックが不充分
2. **P8 違反**：エラーが apperrors.Wrap で包まれていない
3. **P10 違反**：Delete メソッドで CountUsageBy{Id} による依存チェックがない、または存在確認が不足

### パターン
- **P1**: Delete/Update メソッドの先頭で必ず `FindByID(ctx, clinicID, id)` を呼ぶ（バリデーション前）
- **P8**: すべてのエラーリターンは `apperrors.Wrap(err, "message")` で包む
- **P10**: Delete メソッドで `CountUsageBy{Id}` → count > 0 なら `WrapConflict` で 409 返却

### 違反ファイル一覧

| ファイル | 違反パターン | 行番号 | メソッド | 問題 |
|---------|-----------|--------|---------|------|
| procedure_service.go | P1/P10 | 209 | Delete | FindByID がない、CountChildren チェック前に存在確認が不足 |
| vaccine_service.go | P1/P10 | 177 | Delete | CountChildren/CountUsage チェックあるが FindByID 存在確認がない |
| diagnosis_service.go | P10 | 349 | DeleteDiagnosisType | CountUsageByDiagnosisTypeID チェック前に FindByID 存在確認が不足 |
| procedure_service.go | P8 | 115 | Create | エラーが apperrors.Wrap なしで返却 |
| vaccine_service.go | P8 | 101 | Create | apperrors.Wrap なしで返却 |
| staff_service.go | P1 | 417 | Update | FindByID 呼び出すが結果の不在チェックが不明確 |
| checkup_type_service.go | P11 | 102 | Create | slog.ErrorContext のあとに apperrors.Wrap なし |

## 修正方法

### P1 Fix: Delete メソッドに FindByID を追加
```go
// procedure_service.go (L209 例)
func (s *ProcedureService) Delete(ctx context.Context, clinicID, id uint64) error {
    // Step 1: FindByID で存在確認（バリデーション前）
    procedure, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return err  // NotFound は service が伝播
    }
    
    // Step 2: FK 依存チェック（CountUsageBy）
    count, err := s.repo.CountChildrenByParentID(ctx, id)
    if err != nil {
        slog.ErrorContext(ctx, "failed to count children", "error", err)
        return apperrors.Wrap(err, "check children failed")
    }
    if count > 0 {
        return apperrors.WrapConflict("procedure", fmt.Sprintf("has %d children", count))
    }
    
    // Step 3: 削除実行
    return s.repo.Delete(ctx, clinicID, id)
}
```

### P8 Fix: apperrors.Wrap でエラー包装
```go
// procedure_service.go (L115 例)
func (s *ProcedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
    // バリデーション
    if input.Name == "" {
        return nil, apperrors.WrapInvalidInput("name is required")
    }
    
    procedure := &model.Procedure{
        ClinicID: clinicID,
        Name:     input.Name,
    }
    
    if err := s.repo.Create(ctx, procedure); err != nil {
        slog.ErrorContext(ctx, "failed to create procedure", "error", err)
        return nil, apperrors.Wrap(err, "create procedure failed")  // ← 必須
    }
    
    return procedure, nil
}
```

### P10 Fix: Delete で依存チェック + 存在確認
```go
// vaccine_service.go (L177 例)
func (s *VaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
    // Step 1: 存在確認
    _, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return err
    }
    
    // Step 2: 親の依存チェック
    parentCount, err := s.repo.CountParentUsage(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "check parent usage failed")
    }
    if parentCount > 0 {
        return apperrors.WrapConflict("vaccine", "has parent usages")
    }
    
    // Step 3: 子の依存チェック
    childCount, err := s.repo.CountChildrenByParentID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "check children failed")
    }
    if childCount > 0 {
        return apperrors.WrapConflict("vaccine", "has children")
    }
    
    // Step 4: 削除
    return s.repo.Delete(ctx, clinicID, id)
}
```

## テスト

修正後、以下の確認を実施：
- [ ] Delete 時に存在しないリソースで NotFound を返すこと
- [ ] 子リソースがある場合 Conflict (409) を返すこと
- [ ] エラー返却が apperrors.Wrap で包装されていること
- [ ] Create 時のエラーが適切にログ・包装されること
- [ ] service テストが全件パス

## 参考

- Pattern: P1 (FindByID before Delete/Update)
- Pattern: P8 (apperrors.Wrap for all errors)
- Pattern: P10 (FK dependency check before Delete)
