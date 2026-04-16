# BUG-347: staffService にトランザクションなしの複数テーブル操作が 2 箇所

## 概要

`staff_service.go` の以下 2 メソッドで、複数テーブルを跨ぐ操作がトランザクション保護なしで実行されている。

### 1. SetClinicAssignments (L256-272)
コメントに「トランザクション内で差し替える」と明記しているにもかかわらず、トランザクション未使用。
`DeleteByStaffID` 成功後に `Create` が失敗すると、スタッフのクリニック割当が空になるデータ破損が発生する。

### 2. CreateWithAccount (L186-240)
Account 作成（L211）→ Staff 作成（L235）が別々の DB 操作。
Staff 作成が失敗すると、孤立した Account レコードが残る。

## 影響範囲

- **呼び出し元**: スタッフのクリニック割当更新 API
- **症状**: 割当更新中にDBエラーが発生した場合、スタッフが全クリニックから除外された状態になる
- **頻度**: 低頻度だが発生時の影響は重大（スタッフがシステムにアクセスできなくなる）

## 根本原因

コメントと実装が乖離している。`staffService` 構造体に `Transactor` が注入されておらず、
Delete と Create が独立した DB オペレーションとして実行される。

```go
// backend/internal/service/staff_service.go:255-272（現在・バグあり）
// SetClinicAssignments はスタッフのクリニック割当をトランザクション内で差し替える。
func (s *staffService) SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error {
    if err := s.assignmentRepo.DeleteByStaffID(ctx, staffID); err != nil {
        return apperrors.Wrap(err, "failed to delete existing clinic assignments")
    }
    // ↑ Delete 成功後にここで DB 障害が起きると割当が空になる
    for i, clinicID := range clinicIDs {
        assignment := &model.StaffClinicAssignment{
            StaffID:  staffID,
            ClinicID: clinicID,
            IsMain:   i == 0,
        }
        if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
            return apperrors.Wrap(err, "failed to create clinic assignment")
        }
    }
    slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", staffID), slog.Int("count", len(clinicIDs)))
    return nil
}
```

## 修正方針

`staffService` 構造体に `Transactor` を注入し、`WithTx` でラップする。

```go
// staffService 構造体に追加
type staffService struct {
    repo           repository.StaffRepository
    accountRepo    repository.AccountRepository
    assignmentRepo repository.StaffClinicAssignmentRepository
    tx             repository.Transactor // 追加
}

// 修正後の SetClinicAssignments
func (s *staffService) SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error {
    return s.tx.WithTx(ctx, func(ctx context.Context) error {
        if err := s.assignmentRepo.DeleteByStaffID(ctx, staffID); err != nil {
            return apperrors.Wrap(err, "failed to delete existing clinic assignments")
        }
        for i, clinicID := range clinicIDs {
            assignment := &model.StaffClinicAssignment{
                StaffID:  staffID,
                ClinicID: clinicID,
                IsMain:   i == 0,
            }
            if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
                return apperrors.Wrap(err, "failed to create clinic assignment")
            }
        }
        return nil
    })
    slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", staffID), slog.Int("count", len(clinicIDs)))
    return nil
}
```

また、DI 配線（`cmd/api/main.go` 等）で `Transactor` をコンストラクタ引数として渡す必要がある。

## 優先度

**HIGH** — データ整合性バグ。DB 障害時にスタッフのクリニック割当が消失する。

## 確認方法

1. スタッフのクリニック割当を更新するAPIを実行し、Create の途中で DB 接続を切断する
2. 修正前: DeleteByStaffID が適用されてスタッフの割当が空になる
3. 修正後: トランザクションがロールバックされて元の割当が保持される

## 関連ファイル

- `backend/internal/service/staff_service.go:256-272`
- `backend/internal/repository/transactor.go`（Transactor インターフェース）
- `backend/cmd/api/main.go`（DI 配線の更新が必要）
