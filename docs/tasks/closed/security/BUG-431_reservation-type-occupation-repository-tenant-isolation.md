# BUG-431: reservation_type_occupation_repository の Preload テナント分離欠落と Delete の clinicScope 不統一

## 概要

`reservation_type_occupation_repository.go` に2つの問題がある。

1. `FindAll` の `Preload("Occupation")` が clinic_id によるテナント分離を行っていない
2. `Delete` メソッドが `clinicScope` ヘルパーを使わず直接 WHERE 句を記述している（BUG-416 の reservation_type_unavailable_time_repository と同種）

## 問題 1: Preload のテナント分離欠落

```go
// reservation_type_occupation_repository.go:45-58（FindAll）
func (r *reservationTypeOccupationRepository) FindAll(
    ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeOccupation, error) {
    var results []model.ReservationTypeOccupation
    err := r.db.WithContext(ctx).
        Preload("Occupation").                              // ← clinicID フィルタなし
        Scopes(clinicScope(clinicID)).
        Where("reservation_type_id = ?", reservationTypeID).
        Order("id ASC").
        Find(&results).Error
```

**リスク**: `Occupation`（スタッフ職種）モデルが clinic_id を持ち、かつ複数クリニックで同じ occupation_id が存在する場合、Preload が他クリニックの Occupation を返す可能性がある。

**Occupation モデルの確認が必要**: 
- もし Occupation がグローバルマスタ（clinic_id なし）であれば問題なし
- clinic_id を持つ場合は Preload に条件追加が必要

```go
// 修正案（Occupation が clinic_id を持つ場合）
Preload("Occupation", "clinic_id = ?", clinicID).
```

## 問題 2: Delete が clinicScope ヘルパーを使わず直接 WHERE

```go
// reservation_type_occupation_repository.go:70-82（Delete）
func (r *reservationTypeOccupationRepository) Delete(
    ctx context.Context, clinicID, reservationTypeID, occupationID uint64,
) error {
    result := r.db.WithContext(ctx).
        Where("clinic_id = ? AND reservation_type_id = ? AND occupation_id = ?",
              clinicID, reservationTypeID, occupationID).  // ← clinicScope 未使用
        Delete(&model.ReservationTypeOccupation{})
```

**標準パターン**との比較:

```go
// 他リポジトリの Delete（clinicScope 使用）
result := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).
    Delete(&model.Xxx{})
```

BUG-416 で `reservation_type_unavailable_time_repository` の同種問題を指摘済みだが、
本ファイルは別ファイルのため改めて起票する。

## 修正方針

### 問題 1: Occupation モデルの clinic_id 有無を確認し対処

```bash
# 確認コマンド
grep -n "ClinicID\|clinic_id" backend/internal/model/account.go
```

- clinic_id あり → `Preload("Occupation", "clinic_id = ?", clinicID)` に変更
- clinic_id なし（グローバル）→ 変更不要

### 問題 2: Delete を clinicScope パターンに統一

```go
// 修正後
func (r *reservationTypeOccupationRepository) Delete(
    ctx context.Context, clinicID, reservationTypeID, occupationID uint64,
) error {
    result := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("reservation_type_id = ? AND occupation_id = ?", reservationTypeID, occupationID).
        Delete(&model.ReservationTypeOccupation{})
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "reservation_type_occupation", "")
    }
    return nil
}
```

## 影響ファイル

- `backend/internal/repository/reservation_type_occupation_repository.go` — 行 45-58（FindAll）、行 70-82（Delete）

## 優先度

- 問題 1（Preload テナント分離）: **Medium** — Occupation のモデル設計確認次第
- 問題 2（clinicScope 不統一）: **Low** — 機能的には同等だが一貫性のため修正

## 関連チケット

- BUG-416（reservation_type_unavailable_time_repository の clinicScope 不統一）
- BUG-427（occupation_repository の StaffClinicAssignment soft delete 問題）
