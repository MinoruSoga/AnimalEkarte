# BUG-153: スタッフ作成時にクリニック割り当てが自動で行われない

## 概要
`POST /api/v1/masters/staffs` でスタッフを作成すると、`staff_clinic_assignments` テーブルに
レコードが作成されない。作成したスタッフが一覧に表示されず、PATCH/DELETE も 404 になる。

スタッフ一覧はクリニックスコープ（`staff_clinic_assignments` JOIN）で取得されるため、
クリニック未割り当てのスタッフは「見えない」状態になる。

## 脆弱性分類
- **機能バグ**
- **影響**: 新規作成したスタッフが一覧に表示されず、編集・削除もできない。孤立レコードが発生。

## 再現手順
```bash
# スタッフ作成
curl -X POST /api/v1/masters/staffs \
  -H 'Content-Type: application/json' \
  -d '{"name": "テスト太郎", "occupation_id": 1}'
# → 201 Created, id=43

# 一覧取得
curl /api/v1/masters/staffs
# → 15件（id=43 が含まれない）

# 直接取得
curl /api/v1/masters/staffs/43
# → 200 OK（存在する）

# 更新
curl -X PATCH /api/v1/masters/staffs/43 -d '{"name":"edited"}'
# → 404 Not Found（クリニックスコープで見つからない）
```

## ブラウザテスト結果
| 操作 | 期待 | 実際 |
|------|------|------|
| Create | 201 | 201 ✅ |
| Read (list) | リストに含まれる | **含まれない** ❌ |
| Read (direct) | 200 | 200 ✅ |
| Update | 200 | **404** ❌ |
| Delete | 204 | **404** ❌ |

## 期待する動作
スタッフ作成時に、現在のクリニック ID で `staff_clinic_assignments` に自動レコード作成。

## 修正方針

### Handler または Service 層で自動割り当て

```go
func (s *StaffService) Create(ctx context.Context, clinicID uint64, input CreateInput) (*model.Staff, error) {
    staff, err := s.repo.Create(ctx, &model.Staff{...})
    if err != nil { return nil, err }
    
    // 現在のクリニックに自動割り当て
    if err := s.clinicAssignmentRepo.Create(ctx, staff.ID, clinicID); err != nil {
        return nil, apperrors.Wrap(err, "failed to assign clinic")
    }
    
    return staff, nil
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md`
> 全テーブルに `clinic_id`（マルチテナント）

スタッフはクリニックスコープで管理される。作成時にスコープに含まれるようにすべき。

## 優先度
**High** — スタッフ作成機能が実質的に壊れている。作成しても一覧に表示されず、編集・削除もできない。

## 関連ファイル
- `backend/internal/handler/staff_handler.go` — CreateStaff
- `backend/internal/service/staff_service.go` — Create
- `backend/internal/repository/staff_clinic_assignment_repository.go`
