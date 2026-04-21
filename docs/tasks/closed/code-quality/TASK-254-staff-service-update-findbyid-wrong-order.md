# TASK-254: staff_service.go — Update が FindByID なしで repo.Update を実行し、その後で FindByID を呼ぶ（順序バグ）

## 優先度
High

## 対象ファイル
- `backend/internal/service/staff_service.go`

## 問題概要
`Update`（行337〜）が存在確認なしに `s.repo.Update(ctx, clinicID, id, fields)` を実行し、
その後（行365）で `s.repo.FindByID(ctx, id)` を呼ぶ設計になっている。

存在しない `id` に対してリクエストが来ると:
1. `repo.Update` が 0行更新で正常終了（GORM は 0件でもエラーなし）
2. `repo.FindByID` が 404 を返す
3. 呼び出し元は「更新処理の後で 404」という矛盾したレスポンスを受け取る

規約: **FindByID → バリデーション → buildUpdateFields → repo.Update の順にする。**

## 現状コード（行337〜369）

```go
func (s *staffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error) {
    if input.Name != nil {
        // ← ❌ FindByID なしでバリデーション開始
        if err := validateRequiredName(*input.Name); err != nil {
            return nil, err
        }
    }
    // ...
    if hasProfileUpdate {
        fields := buildStaffUpdateFields(input)
        if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {  // ❌ 存在確認なしで Update
            return nil, apperrors.Wrap(err, "failed to update staff")
        }
    }

    updated, err := s.repo.FindByID(ctx, id)  // ❌ Update の後で FindByID（順序逆）
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get updated staff")
    }
```

## あるべき姿

```go
func (s *staffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error) {
    // ✅ 最初に存在確認
    if _, err := s.repo.FindByID(ctx, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get staff")
    }

    if input.Name != nil {
        // バリデーションはその後
        if err := validateRequiredName(*input.Name); err != nil {
            return nil, err
        }
    }
    // ...
    if hasProfileUpdate {
        fields := buildStaffUpdateFields(input)
        if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
            return nil, apperrors.Wrap(err, "failed to update staff")
        }
        slog.InfoContext(ctx, "staff updated", ...)
    }

    // ✅ 更新後の再取得は FindByID を再利用
    updated, err := s.repo.FindByID(ctx, id)
    // ...
```

## 完了条件
- [ ] `Update` の先頭（バリデーション前）に `FindByID` による存在確認を追加
- [ ] `go test ./backend/internal/...` がパス
