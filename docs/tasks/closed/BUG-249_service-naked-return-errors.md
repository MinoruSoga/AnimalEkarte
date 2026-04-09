# BUG-249: Service 層で `apperrors.Wrap` なし naked return（12+サービス）

## 概要

プロジェクト規約では Service 層で内部エラーは `apperrors.Wrap(err, "message")` でラッピングすることが必須。
しかし多数のサービスが `return s.repo.FindAll(ctx, ...)` のように Repository エラーをそのまま返しており、
エラーにサービス層のコンテキスト情報が付与されない。

## 影響範囲

| サービスファイル | 行 | メソッド |
|-----------------|-----|---------|
| `insurance_service.go` | 33,36,39,62 | List, GetByID, Create, Delete |
| `inventory_service.go` | 65 | Delete（FK チェック後） |
| `accounting_service.go` | 108,112 | List, GetByID |
| `estimate_service.go` | 63,67 | List, GetByID |
| `billing_item_service.go` | 88 | GetByID |
| `consultation_service.go` | 39 | Create |
| `examination_service.go` | 38 | Create |
| `medical_record_service.go` | 168 | Delete |
| `diagnosis_service.go` | 135,250 | DiagnosisCategory.Delete, DiagnosisName.Delete |
| `vital_service.go` | 108 | Update 後の FindByID |
| `hospitalization_plan_service.go` | 32-65 | List, GetByID, Create, Reorder |
| `cage_service.go` | 34,37,40,62,69 | List, GetByID, Create, Update, Reorder |
| `daily_record_service.go` | 61,65 | 複数メソッド |
| `reservation_course_service.go` | 62,66 | List, GetByID |
| `reservation_staff_service.go` | 119 | Delete |

## 修正方針

全箇所で naked return をラップ形式に変更する。

### 修正パターン

```go
// 修正前（passthrough）
func (s *insuranceService) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
    return s.repo.FindAll(ctx, clinicID)
}

// 修正後
func (s *insuranceService) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
    result, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list insurances")
    }
    return result, nil
}
```

```go
// 修正前（void 系）
func (s *inventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
    // ... FK チェック ...
    return s.repo.Delete(ctx, clinicID, id)
}

// 修正後
func (s *inventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
    // ... FK チェック ...
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete inventory item")
    }
    return nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> **Service**: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。

### `.claude/rules/error-handling.md` — Service 層
> `apperrors.Wrap(err, "failed to get owner")` でコンテキストを付与する。

## 優先度
**High** — エラー発生時のデバッグ・監査が困難になる。エラーメッセージからどのサービスで問題が発生したか特定できない。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）
