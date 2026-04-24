# BUG-420: trimming_service の Update でマルチテナント検証（FindByID）が欠落

## 概要

`trimming_service.go` の `Update` メソッドが、更新前に `FindByID(ctx, clinicID, id)` による
存在確認・テナント検証を行わずに直接 `UpdateFields` を呼び出している。
他のマスタサービス（cage, insurance, medicine 等）では更新前の存在確認が実装済みであり、
一貫性が欠如している。悪意あるリクエストが別クリニックのトリミングデータを改ざんできる可能性がある。

## 問題箇所

```go
// trimming_service.go:174-200（Update メソッド）
func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
    apptFields := map[string]any{}
    if input.StartTime != nil {
        apptFields["start_time"] = *input.StartTime
    }
    // ... フィールド構築 ...

    if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
        if len(apptFields) > 0 {
            if _, err := s.reservation.UpdateFields(txCtx, clinicID, id, apptFields); err != nil {
                return apperrors.Wrap(err, "failed to update trimming appointment")
            }
        }
        // ...
    }); err != nil { ... }
    // ← FindByID による存在確認・clinicID 検証なし
}
```

## 他サービスとの比較（標準パターン）

```go
// cage_service.go（標準パターン）
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    // 存在確認（テナント検証含む）
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // ...
}

// insurance_service.go（同様）
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get insurance")
    }
    // ...
}
```

## リスク

`UpdateFields` が内部で `clinic_id = ?` を条件に持っていれば実害はないが、
Repository 層の実装変更時に clinicID の条件が外れた場合、別クリニックのデータを上書きできる。

## 修正方針

Update メソッドの先頭に存在確認を追加する。

```go
func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
    // ← 追加: 存在確認・テナント検証
    if _, err := s.reservation.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get trimming appointment")
    }

    apptFields := map[string]any{}
    // ... 以降は変更なし
}
```

## 影響ファイル

- `backend/internal/service/trimming_service.go` — 行 174-200（Update メソッド全体）

## 優先度

**High** — マルチテナント境界の保護が不完全。他サービスとの一貫性違反。

## 関連チケット

- BUG-413（chief_complaint_service の clinicID 欠落）
