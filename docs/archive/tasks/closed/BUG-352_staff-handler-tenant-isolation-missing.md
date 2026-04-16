# BUG-352: staff_handler.go の 7 エンドポイントでテナント分離が欠落（TODO 放置）

## 概要

`staff_handler.go` の 7 つのエンドポイントで `_ = clinicID`（`TODO: pass clinicID to repo for tenant isolation`）が放置されており、
認証済みの任意クリニックのユーザーが**他クリニックの全スタッフ**の情報取得・権限変更・クリニック割当変更・除外設定変更を行える。

## 影響範囲

### 読み取り系（3件）— 他クリニックのスタッフ情報が閲覧可能
| 行 | メソッド | エンドポイント |
|----|---------|---------------|
| L204 | `GetStaff` | `GET /v1/masters/staffs/:id` |
| L244 | `GetStaffPermissionGroups` | `GET /v1/masters/staffs/:id/permission-groups` |
| L356 | `GetStaffExcludedReservationTypes` | `GET /v1/masters/staffs/:id/excluded-reservation-types` |

### 書き込み系（4件）— 他クリニックのスタッフ設定を改ざん可能
| 行 | メソッド | エンドポイント |
|----|---------|---------------|
| L266 | `SetStaffPermissionGroups` | `PUT /v1/masters/staffs/:id/permission-groups` |
| L323 | `SetStaffClinicAssignments` | `PUT /v1/masters/staffs/:id/clinics` |
| L378 | `SetStaffExcludedReservationTypes` | `PUT /v1/masters/staffs/:id/excluded-reservation-types` |

※ `ListStaffs`, `CreateStaff`, `UpdateStaff`, `DeleteStaff`, `ReorderStaffs` は `clinicID` を正しくサービスに渡しており問題なし。

## 根本原因

`staffs` テーブルが `clinic_id` カラムを持たず `staff_clinic_assignments` で多対多関係のため、
直接 `clinicScope` が適用できない。そのため `TODO` として後回しにされ、全く検証されていない。

## 修正方針

サービス層で `staff_clinic_assignments` を確認し、対象スタッフが呼び出し元クリニックに所属しているか検証する。

```go
// service/staff_service.go に追加
func (s *staffService) verifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error {
    exists, err := s.assignmentRepo.ExistsByStaffAndClinic(ctx, staffID, clinicID)
    if err != nil {
        return apperrors.Wrap(err, "failed to verify clinic membership")
    }
    if !exists {
        return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
    }
    return nil
}
```

各サービスメソッドにも `clinicID` を追加:
```go
// Before
GetByID(ctx context.Context, id uint64) (*model.Staff, error)

// After
GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
```

## 修正が必要なファイル

1. `backend/internal/handler/staff_handler.go` — 7箇所の `_ = clinicID` を `clinicID` として渡す
2. `backend/internal/service/staff_service.go` — 7メソッドのシグネチャに `clinicID` 追加 + `verifyClinicMembership` 呼び出し
3. `backend/internal/repository/staff_clinic_assignment_repository.go` — `ExistsByStaffAndClinic` メソッド追加（なければ）

## 優先度

**CRITICAL（セキュリティ）** — 認証済みユーザーによるクロステナント読み書きが可能。
書き込み系 4 エンドポイントで他クリニックのスタッフの権限グループ・クリニック割当・予約除外設定を改ざんできる。
