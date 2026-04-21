---
title: Repository P4 Upsert clinicScope 欠落
issue: '#473'
priority: HIGH
status: open
area: repository
pattern: P4
---

## 概要

リポジトリの Upsert（CreateOrUpdate）パターンで、マルチテナント隔離を担保する `clinicScope` が欠落しているケースが 1 件検出されました。

### パターン
- **P4 違反**：UPDATE/UPSERT 時に `Scopes(clinicScope(clinicID))` がない、またはネストされた関連エンティティの scope が不足

### 違反ファイル一覧

| ファイル | 行番号 | 現在の実装 | 問題 |
|---------|--------|----------|------|
| reservation_staff_repository.go | 82 | `db.Model(&reservationStaff).Updates(fields).Error` | Scopes(clinicScope) がない → 他テナントのレコードも更新可能 |

## 修正方法

Upsert 前に必ず `Scopes(clinicScope(clinicID))` を適用：

```go
// 修正例
func (r *ReservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationStaff, error) {
    var reservationStaff model.ReservationStaff
    if err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Model(&reservationStaff).
        Where("id = ?", id).
        Updates(fields).Error; err != nil {
        return nil, apperrors.FromGORM(err, "reservation_staff", id)
    }
    return &reservationStaff, nil
}
```

## テスト

修正後、以下の確認を実施：
- [ ] 他テナントのレコードが更新されないこと
- [ ] 同一テナント内のレコードは正常に更新されること
- [ ] リポジトリテストが全件パス

## 参考

- Pattern: P4 (Upsert clinicScope)
- 関連タスク: TASK-469 (P4 clinicScope 類似違反)
